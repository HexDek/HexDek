package heimdall

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	seedBufSize = 1000
)

// Observer is the singleton that receives game results from ALL game paths
// (grinder, fishtank, gauntlet). It buffers seeds in memory and flushes to
// disk in batches, then routes observations to downstream sinks (Huginn,
// Muninn, telemetry).
type Observer struct {
	seedBuf   []GameSeed
	obsBuf    []Observation
	huginn    HuginnSink
	muninn    MuninnSink
	telemetry TelemetrySink
	dataDir   string
	mu        sync.Mutex
}

// New creates an Observer. Sinks can be nil (observations are silently dropped).
func New(dataDir string, huginn HuginnSink, muninn MuninnSink, telemetry TelemetrySink) *Observer {
	o := &Observer{
		seedBuf:   make([]GameSeed, 0, seedBufSize),
		huginn:    huginn,
		muninn:    muninn,
		telemetry: telemetry,
		dataDir:   dataDir,
	}
	os.MkdirAll(filepath.Join(dataDir, "heimdall"), 0755)
	return o
}

// RecordSeed is called after EVERY game. Zero-allocation fast path -- appends
// to ring buffer, flushes to disk when full.
func (o *Observer) RecordSeed(seed GameSeed) {
	o.mu.Lock()
	o.seedBuf = append(o.seedBuf, seed)
	if len(o.seedBuf) >= seedBufSize {
		seeds := make([]GameSeed, len(o.seedBuf))
		copy(seeds, o.seedBuf)
		o.seedBuf = o.seedBuf[:0]
		o.mu.Unlock()
		o.flushSeeds(seeds)
		return
	}
	o.mu.Unlock()
}

// RecordObservation is called during batch replay or live observation mode.
// Routes co-triggers to Huginn, parser gaps and dead triggers to Muninn.
func (o *Observer) RecordObservation(obs Observation) {
	if o.huginn != nil && len(obs.CoTriggers) > 0 {
		names := make([]string, len(obs.Seed.DeckKeys))
		copy(names, obs.Seed.DeckKeys[:])
		o.huginn.IngestCoTriggers(obs.CoTriggers, names)
	}
	if o.muninn != nil {
		gameID := fmt.Sprintf("%d", obs.Seed.RNGSeed)
		if len(obs.ParserGaps) > 0 {
			o.muninn.RecordParserGaps(obs.ParserGaps, gameID)
		}
		if len(obs.DeadTriggers) > 0 {
			o.muninn.RecordDeadTriggers(obs.DeadTriggers, gameID)
		}
	}
	o.mu.Lock()
	o.obsBuf = append(o.obsBuf, obs)
	o.mu.Unlock()
}

// RecordCrash is called from panic recovery. Routes to Muninn immediately.
func (o *Observer) RecordCrash(panicMsg string, stack []byte, deckKeys []string) {
	if o.muninn != nil {
		o.muninn.RecordCrash(panicMsg, string(stack), deckKeys)
	}
}

// Pulse forwards a health pulse to the telemetry sink (GA4).
func (o *Observer) Pulse(stats HealthPulse) {
	if o.telemetry != nil {
		o.telemetry.Pulse(stats)
	}
}

// Flush writes any buffered seeds to disk. Call on graceful shutdown.
func (o *Observer) Flush() {
	o.mu.Lock()
	if len(o.seedBuf) > 0 {
		seeds := make([]GameSeed, len(o.seedBuf))
		copy(seeds, o.seedBuf)
		o.seedBuf = o.seedBuf[:0]
		o.mu.Unlock()
		o.flushSeeds(seeds)
		return
	}
	o.mu.Unlock()
}

// Observations returns a snapshot of the buffered observations. Safe for
// concurrent reads -- the caller gets an independent copy.
func (o *Observer) Observations() []Observation {
	o.mu.Lock()
	defer o.mu.Unlock()
	out := make([]Observation, len(o.obsBuf))
	copy(out, o.obsBuf)
	return out
}

func (o *Observer) flushSeeds(seeds []GameSeed) {
	// Append to seed file (JSON lines for now, binary later).
	fname := filepath.Join(o.dataDir, "heimdall", "seeds.jsonl")
	f, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, s := range seeds {
		enc.Encode(s)
	}
}
