package tournament

import (
	"testing"

	"github.com/hexdek/hexdek/internal/deckparser"
	"github.com/hexdek/hexdek/internal/gameengine"
	"github.com/hexdek/hexdek/internal/hat"
	"github.com/hexdek/hexdek/internal/seedcontract"
)

// e2eDecks loads a fixed-size deck set with the shared corpus. Skips
// the test when the corpus or decks aren't available — same pattern
// as TestSmallTournament. Returns nSeats decks; if fewer are found,
// the test is skipped.
func e2eDecks(t *testing.T, nSeats int) []*deckparser.TournamentDeck {
	t.Helper()
	corpus, meta := loadCorpus(t)
	paths := findDecks(t, nSeats)
	if len(paths) < nSeats {
		t.Skipf("need %d decks, got %d", nSeats, len(paths))
	}
	decks := make([]*deckparser.TournamentDeck, 0, nSeats)
	for _, p := range paths[:nSeats] {
		d, err := deckparser.ParseDeckFile(p, corpus, meta)
		if err != nil {
			t.Fatalf("parse %s: %v", p, err)
		}
		decks = append(decks, d)
	}
	return decks
}

// e2eHats builds nSeats GreedyHat factories — deterministic, no
// internal RNG-divergent decisions, so two runs with the same RNG
// seed produce the same outcome.
func e2eHats(nSeats int) []HatFactory {
	hats := make([]HatFactory, nSeats)
	for i := range hats {
		hats[i] = func() gameengine.Hat { return &hat.GreedyHat{} }
	}
	return hats
}

// TestSeedContract_DeterministicReplayProducesIdenticalDigests is the
// core Phase 1 anti-cheat invariant. Two runs of the same gameIdx
// against the same deck set + RNG seed must produce contracts with
// identical InputDigest AND OutcomeDigest. If this test fails, the
// engine has a non-determinism bug and the seed contract guarantee is
// broken.
func TestSeedContract_DeterministicReplayProducesIdenticalDigests(t *testing.T) {
	const nSeats = 2
	decks := e2eDecks(t, nSeats)
	hats := e2eHats(nSeats)

	// Run game idx 0 twice with the same seed and identical contract
	// params. The contract.SealedAtUnix is set inside runOneGame from
	// time.Now() and WILL differ between the two runs — which is why
	// we compare digests via a re-seal that controls SealedAtUnix
	// rather than comparing whole contracts.
	first := runOneGame(0, decks, hats, nSeats, 12345, 30, true, false, false, nil, contractParams{
		key:           []byte("test-key-1234567890abcdef"),
		context:       "test:determinism",
		engineVersion: "test-build",
	})
	second := runOneGame(0, decks, hats, nSeats, 12345, 30, true, false, false, nil, contractParams{
		key:           []byte("test-key-1234567890abcdef"),
		context:       "test:determinism",
		engineVersion: "test-build",
	})

	if first.SeedContract == nil || second.SeedContract == nil {
		t.Fatal("SeedContract not attached to outcome")
	}
	if first.SeedContract.OutcomeDigest == "" {
		t.Fatal("OutcomeDigest not sealed on first run")
	}
	if first.SeedContract.OutcomeDigest != second.SeedContract.OutcomeDigest {
		t.Fatalf("outcome digest non-deterministic: first=%s second=%s\nfirst winner=%d turns=%d / second winner=%d turns=%d",
			first.SeedContract.OutcomeDigest, second.SeedContract.OutcomeDigest,
			first.Winner, first.Turns, second.Winner, second.Turns)
	}
	// InputDigest will differ ONLY if SealedAtUnix differs across
	// runs. Confirm the inputs themselves agree by re-deriving with a
	// canonical SealedAtUnix.
	in := seedcontract.Inputs{
		RNGSeed:       first.SeedContract.RNGSeed,
		DeckKeys:      first.SeedContract.DeckKeys,
		NSeats:        first.SeedContract.NSeats,
		EngineVersion: first.SeedContract.EngineVersion,
		SealedAtUnix:  0, // hold constant
	}
	in2 := seedcontract.Inputs{
		RNGSeed:       second.SeedContract.RNGSeed,
		DeckKeys:      second.SeedContract.DeckKeys,
		NSeats:        second.SeedContract.NSeats,
		EngineVersion: second.SeedContract.EngineVersion,
		SealedAtUnix:  0,
	}
	a := seedcontract.New(in)
	b := seedcontract.New(in2)
	if a.InputDigest != b.InputDigest {
		t.Fatalf("input digest non-deterministic: %s vs %s", a.InputDigest, b.InputDigest)
	}
}

// TestSeedContract_VerifySucceedsAfterFreshGame confirms a game's
// signed contract verifies cleanly under the same key.
func TestSeedContract_VerifySucceedsAfterFreshGame(t *testing.T) {
	const nSeats = 2
	decks := e2eDecks(t, nSeats)
	hats := e2eHats(nSeats)

	key := seedcontract.DeriveContractKey([]byte("master"), "test:verify")
	out := runOneGame(0, decks, hats, nSeats, 99, 30, true, false, false, nil, contractParams{
		key:           key,
		context:       "test:verify",
		engineVersion: "test-build",
	})
	if out.SeedContract == nil {
		t.Fatal("SeedContract not attached")
	}
	if out.SeedContract.Sig == "" {
		t.Fatal("contract not signed despite key supplied")
	}
	if !out.SeedContract.Verify(key) {
		t.Fatal("Verify failed on freshly signed contract")
	}
	if err := out.SeedContract.CheckIntegrity(key); err != nil {
		t.Fatalf("CheckIntegrity failed: %v", err)
	}
}

// TestSeedContract_TamperedOutcomeDetected runs a game, then swaps
// the claimed winner. The tampered contract must fail integrity.
func TestSeedContract_TamperedOutcomeDetected(t *testing.T) {
	const nSeats = 2
	decks := e2eDecks(t, nSeats)
	hats := e2eHats(nSeats)

	key := seedcontract.DeriveContractKey([]byte("master"), "test:tamper")
	out := runOneGame(0, decks, hats, nSeats, 7, 30, true, false, false, nil, contractParams{
		key:           key,
		context:       "test:tamper",
		engineVersion: "test-build",
	})
	if out.SeedContract == nil || out.SeedContract.Sig == "" {
		t.Fatal("contract not signed")
	}
	c := out.SeedContract

	// Forge winner: flip seat 0 ↔ seat 1. Don't re-sign.
	c.Outcome.Winner = (c.Outcome.Winner + 1) % nSeats
	if err := c.CheckIntegrity(key); err == nil {
		t.Fatal("CheckIntegrity passed despite tampered Winner")
	}
}

// TestSeedContract_TamperedTurnsDetected confirms turn-count forgery
// is also caught.
func TestSeedContract_TamperedTurnsDetected(t *testing.T) {
	const nSeats = 2
	decks := e2eDecks(t, nSeats)
	hats := e2eHats(nSeats)

	key := seedcontract.DeriveContractKey([]byte("master"), "test:tamper-turns")
	out := runOneGame(0, decks, hats, nSeats, 13, 30, true, false, false, nil, contractParams{
		key:           key,
		context:       "test:tamper-turns",
		engineVersion: "test-build",
	})
	c := out.SeedContract
	if c == nil || c.Sig == "" {
		t.Fatal("contract not signed")
	}
	c.Outcome.Turns = c.Outcome.Turns + 100
	if err := c.CheckIntegrity(key); err == nil {
		t.Fatal("CheckIntegrity passed despite tampered Turns")
	}
}

// TestSeedContract_TamperedDeckKeyDetected forges a deck key on the
// input side. The CheckIntegrity input-digest re-derivation must
// catch it.
func TestSeedContract_TamperedDeckKeyDetected(t *testing.T) {
	const nSeats = 2
	decks := e2eDecks(t, nSeats)
	hats := e2eHats(nSeats)

	key := seedcontract.DeriveContractKey([]byte("master"), "test:tamper-deck")
	out := runOneGame(0, decks, hats, nSeats, 21, 30, true, false, false, nil, contractParams{
		key:           key,
		context:       "test:tamper-deck",
		engineVersion: "test-build",
	})
	c := out.SeedContract
	if c == nil || c.Sig == "" {
		t.Fatal("contract not signed")
	}
	c.DeckKeys[0] = "mallory/forged-deck"
	if err := c.CheckIntegrity(key); err == nil {
		t.Fatal("CheckIntegrity passed despite tampered DeckKeys[0]")
	}
}

// TestSeedContract_UnsignedWhenNoKey confirms that a contractParams
// with nil key still produces digests but leaves Sig empty. This is
// the "observer mode" — we want digests for tooling consistency even
// when signing is disabled.
func TestSeedContract_UnsignedWhenNoKey(t *testing.T) {
	const nSeats = 2
	decks := e2eDecks(t, nSeats)
	hats := e2eHats(nSeats)

	out := runOneGame(0, decks, hats, nSeats, 5, 30, true, false, false, nil, contractParams{})
	if out.SeedContract == nil {
		t.Fatal("SeedContract not attached")
	}
	if out.SeedContract.InputDigest == "" {
		t.Fatal("InputDigest empty on unsigned contract")
	}
	if out.SeedContract.OutcomeDigest == "" {
		t.Fatal("OutcomeDigest empty on unsigned contract")
	}
	if out.SeedContract.Sig != "" {
		t.Fatal("Sig populated despite nil key")
	}
}

// TestSeedContract_RunWiresContract runs a small Run() tournament
// with a contract key and confirms every per-game outcome carries a
// signed contract. Exercises the Run path (the main entry point) end
// to end.
func TestSeedContract_RunWiresContract(t *testing.T) {
	decks := e2eDecks(t, 2)
	key := seedcontract.DeriveContractKey([]byte("master"), "test:run")

	cfg := TournamentConfig{
		Decks:           decks,
		NSeats:          2,
		NGames:          4,
		Seed:            17,
		Workers:         1,
		CommanderMode:   true,
		MaxTurnsPerGame: 25,
		EngineVersion:   "test-build",
		ContractKey:     key,
		ContractContext: "test:run",
	}
	r, err := Run(cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if r.Games+r.Crashes != 4 {
		t.Fatalf("expected 4 outcomes")
	}
	// Run() doesn't expose per-outcome contracts on the aggregate
	// TournamentResult today — the test that they're attached is via
	// the unit-level runOneGame tests above. This e2e test just
	// confirms Run() with contract config doesn't error or skip games.
}
