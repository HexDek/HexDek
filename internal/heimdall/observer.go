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
// Routes co-triggers + zone-cast + exile-link events to Huginn, and parser
// gaps + dead triggers + (optionally) exile-link audit data to Muninn.
func (o *Observer) RecordObservation(obs Observation) {
	gameID := fmt.Sprintf("%d", obs.Seed.RNGSeed)
	names := make([]string, len(obs.Seed.DeckKeys))
	copy(names, obs.Seed.DeckKeys[:])

	if o.huginn != nil {
		if len(obs.CoTriggers) > 0 {
			o.huginn.IngestCoTriggers(obs.CoTriggers, names)
		}
		if len(obs.ZoneCastEvents) > 0 {
			if zc, ok := o.huginn.(HuginnZoneCastSink); ok {
				zc.IngestZoneCastEvents(obs.ZoneCastEvents, names, gameID)
			}
		}
		if len(obs.ExileLinkEvents) > 0 {
			if el, ok := o.huginn.(HuginnExileLinkSink); ok {
				el.IngestExileLinkEvents(obs.ExileLinkEvents, names, gameID)
			}
		}
	}
	if o.muninn != nil {
		if len(obs.ParserGaps) > 0 {
			o.muninn.RecordParserGaps(obs.ParserGaps, gameID)
		}
		if len(obs.DeadTriggers) > 0 {
			o.muninn.RecordDeadTriggers(obs.DeadTriggers, gameID)
		}
		if len(obs.ExileLinkEvents) > 0 {
			if el, ok := o.muninn.(MuninnExileLinkSink); ok {
				el.RecordExileLinkEvents(obs.ExileLinkEvents, gameID)
			}
		}
	}
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

func (o *Observer) flushSeeds(seeds []GameSeed) {
	fname := filepath.Join(o.dataDir, "heimdall", "seeds.jsonl")
	if info, err := os.Stat(fname); err == nil && info.Size() > maxSeedFileSize {
		os.Rename(fname, fname+".prev")
	}
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
