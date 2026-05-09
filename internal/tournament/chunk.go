package tournament

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hexdek/hexdek/internal/deckparser"
)

// ChunkConfig is a minimal config for the BOINC-style distributed-
// compute path. Unlike Run, RunChunk has no persistence side effects:
// no Muninn batcher, no Huginn writes, no markdown report. It exists so
// the contributor client (cmd/hexdek-contrib) and the server-side
// validator can both run a small batch of games and compare winner
// sequences byte-for-byte.
type ChunkConfig struct {
	// Decks is the per-seat deck list. len(Decks) MUST equal NSeats.
	// Unlike TournamentConfig.Decks the runner does NOT rotate; each
	// seat plays the deck at its own index for game 0, then rotates
	// per-game so every deck plays every seat across the chunk.
	Decks []*deckparser.TournamentDeck

	// NSeats — usually 2 (head-to-head) or 4 (commander pod).
	NSeats int

	// NGames — number of games to simulate.
	NGames int

	// Seed — master RNG seed. Per-game seed = Seed + idx*1000 + 1.
	// MUST match the seed in the signed assignment so the server can
	// re-run a spot-check and get bit-identical winners.
	Seed int64

	// HatFactories — same semantics as TournamentConfig.HatFactories.
	// Empty defaults to GreedyHat for all seats. Determinism note:
	// stochastic hats (anything that uses runtime.UnixNano or wallclock
	// noise) will break spot-check parity. Stick to deterministic hats
	// keyed off the per-game seed.
	HatFactories []HatFactory

	// MaxTurnsPerGame — per-game turn cap. 0 = default.
	MaxTurnsPerGame int

	// CommanderMode — true for §903 commander rules. Default: true.
	CommanderMode bool

	// Workers — goroutine pool size. 0 = runtime.NumCPU().
	Workers int

	// PerGameTimeout — kills a runaway game and records a draw. 0 = default.
	PerGameTimeout time.Duration
}

// ChunkOutcome is the per-game subset returned to the dispatcher.
// Stripped of analytics/audit fields so the JSON payload stays small.
type ChunkOutcome struct {
	GameIdx int
	Winner  int // commander (deck) index that won — already un-rotated
	Turns   int
}

// RunChunk executes a small batch of games in process and returns the
// per-game outcomes. Deterministic across runs that share Seed, NSeats,
// Decks, and HatFactories.
//
// This is the entry point used by:
//   - cmd/hexdek-contrib (contributor client, runs the full chunk)
//   - internal/hexapi (server-side spot-check validator, re-runs a
//     subset for hash comparison)
func RunChunk(cfg ChunkConfig) ([]ChunkOutcome, error) {
	if cfg.NSeats <= 0 {
		return nil, fmt.Errorf("chunk: NSeats must be > 0")
	}
	if cfg.NGames <= 0 {
		return nil, fmt.Errorf("chunk: NGames must be > 0")
	}
	if len(cfg.Decks) != cfg.NSeats {
		return nil, fmt.Errorf("chunk: len(Decks)=%d != NSeats=%d", len(cfg.Decks), cfg.NSeats)
	}

	maxTurns := cfg.MaxTurnsPerGame
	if maxTurns <= 0 {
		maxTurns = defaultMaxTurns
	}
	timeout := cfg.PerGameTimeout
	if timeout <= 0 {
		timeout = defaultPerGameTimeout
	}
	workers := cfg.Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers > cfg.NGames {
		workers = cfg.NGames
	}

	// Normalize hat factories (same rules as Run).
	hats := make([]HatFactory, cfg.NSeats)
	switch len(cfg.HatFactories) {
	case 0:
		for i := range hats {
			hats[i] = defaultHatFactory
		}
	case 1:
		for i := range hats {
			hats[i] = cfg.HatFactories[0]
		}
	default:
		if len(cfg.HatFactories) < cfg.NSeats {
			return nil, fmt.Errorf("chunk: HatFactories must be 0, 1, or NSeats entries")
		}
		copy(hats, cfg.HatFactories[:cfg.NSeats])
	}

	commanderMode := cfg.CommanderMode || true // default to commander; explicitly mutable if a future caller wants vanilla

	out := make([]ChunkOutcome, cfg.NGames)
	var idx int64 = -1
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				gi := int(atomic.AddInt64(&idx, 1))
				if gi >= cfg.NGames {
					return
				}
				// runOneGameSafe wraps runOneGame in a recover() and a
				// per-game timeout, returning a draw on either failure.
				go0 := runOneGameSafe(
					gi, cfg.Decks, hats, cfg.NSeats,
					cfg.Seed, maxTurns, timeout,
					commanderMode,
					false, // auditEnabled — no event-log retention for chunks
					false, // analyticsEnabled — no per-game analytics
				)
				out[gi] = ChunkOutcome{
					GameIdx: gi,
					Winner:  go0.WinnerCommanderIdx,
					Turns:   go0.Turns,
				}
			}
		}()
	}
	wg.Wait()
	return out, nil
}
