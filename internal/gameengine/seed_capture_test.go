package gameengine

import (
	"math/rand"
	"testing"
)

// TestSeedCaptureExplicit confirms callers that set gs.Seed after the
// existing NewGameState constructor see the value preserved on the
// state. Anti-cheat Phase 1 relies on this for storage; Phase 2+ will
// compare it against a re-derived seed at replay time.
func TestSeedCaptureExplicit(t *testing.T) {
	const want int64 = 0x1234567890abcdef
	gs := NewGameState(2, rand.New(rand.NewSource(want)), nil)
	gs.Seed = want
	if gs.Seed != want {
		t.Fatalf("Seed not preserved: got %d, want %d", gs.Seed, want)
	}
	// Default for unset seed must be 0 so consumers can treat 0 as
	// "unknown" rather than "actually seeded with 0".
	gs2 := NewGameState(2, rand.New(rand.NewSource(1)), nil)
	if gs2.Seed != 0 {
		t.Fatalf("expected zero default Seed, got %d", gs2.Seed)
	}
}

// TestNewGameStateSeeded confirms the helper constructor binds the
// seed and the RNG together — same seed produces the same first
// random draw, and gs.Seed returns the input verbatim.
func TestNewGameStateSeeded(t *testing.T) {
	const seed int64 = 42
	gs := NewGameStateSeeded(2, seed, nil)
	if gs.Seed != seed {
		t.Fatalf("Seed not captured: got %d, want %d", gs.Seed, seed)
	}
	if gs.Rng == nil {
		t.Fatalf("expected RNG to be constructed")
	}

	// Determinism: two seeded states with the same seed produce the same
	// stream — necessary precondition for replay.
	a := NewGameStateSeeded(2, 999, nil)
	b := NewGameStateSeeded(2, 999, nil)
	for i := 0; i < 8; i++ {
		av, bv := a.Rng.Int63(), b.Rng.Int63()
		if av != bv {
			t.Fatalf("seeded RNGs diverged at draw %d: a=%d b=%d", i, av, bv)
		}
	}
}
