package hexapi

import (
	"strings"
	"sync"

	"github.com/hexdek/hexdek/internal/astload"
	"github.com/hexdek/hexdek/internal/deckparser"
	"github.com/hexdek/hexdek/internal/tournament"
)

// SpotCheckRunner re-runs an assignment locally and returns the same
// (winners, turns) the client should have produced for the given seed.
// Wired into ContribDispatcher.SpotCheck — the dispatcher samples a
// fraction of returned chunks, calls this function with the same
// assignment, and compares outcome hashes for parity.
//
// The runner caches the corpus + meta once; subsequent calls are cheap.
type SpotCheckRunner struct {
	corpusPath string
	metaPath   string
	oraclePath string

	once   sync.Once
	corpus *astload.Corpus
	meta   *deckparser.MetaDB
	loadOK bool
}

// NewSpotCheckRunner builds a runner with the given paths. Loading is
// deferred until first use so server startup isn't blocked when no
// contributor traffic ever shows up.
func NewSpotCheckRunner(corpusPath, metaPath, oraclePath string) *SpotCheckRunner {
	return &SpotCheckRunner{
		corpusPath: corpusPath,
		metaPath:   metaPath,
		oraclePath: oraclePath,
	}
}

// Run satisfies the ContribDispatcher.SpotCheck signature.
func (s *SpotCheckRunner) Run(a *ContribAssignment) (winners []int, turns []int, ok bool) {
	s.once.Do(s.load)
	if !s.loadOK || a == nil {
		return nil, nil, false
	}
	if len(a.Decks) != a.NSeats {
		return nil, nil, false
	}
	decks := make([]*deckparser.TournamentDeck, a.NSeats)
	for i, raw := range a.Decks {
		d, err := deckparser.ParseDeckReader(strings.NewReader(raw), s.corpus, s.meta)
		if err != nil {
			return nil, nil, false
		}
		decks[i] = d
	}
	cfg := tournament.ChunkConfig{
		Decks:           decks,
		NSeats:          a.NSeats,
		NGames:          a.GamesCount,
		Seed:            a.Seed,
		MaxTurnsPerGame: a.MaxTurns,
		CommanderMode:   true,
		Workers:         1, // serial in spot-check; latency over throughput
	}
	out, err := tournament.RunChunk(cfg)
	if err != nil {
		return nil, nil, false
	}
	winners = make([]int, len(out))
	turns = make([]int, len(out))
	for i, o := range out {
		winners[i] = o.Winner
		turns[i] = o.Turns
	}
	return winners, turns, true
}

func (s *SpotCheckRunner) load() {
	corpus, err := astload.Load(s.corpusPath)
	if err != nil {
		s.loadOK = false
		return
	}
	meta, err := deckparser.LoadMetaFromJSONL(s.metaPath)
	if err != nil {
		s.loadOK = false
		return
	}
	if s.oraclePath != "" {
		_ = meta.SupplementWithOracleJSON(s.oraclePath)
	}
	s.corpus = corpus
	s.meta = meta
	s.loadOK = true
}
