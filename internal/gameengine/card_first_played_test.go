package gameengine

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// First-play instrumentation: the engine should record the turn on
// which a card first resolves as a spell. Recasts of the same card
// (storm copies, recursion) must not overwrite the first-play turn.
// Countered spells are not recorded — they didn't resolve.

func TestCardFirstPlayed_RecordedOnSpellResolve(t *testing.T) {
	gs := newFixtureGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 4
	gs.Turn = 3

	bolt := addHandCardWithEffect(gs, 0, "Lightning Bolt", 1,
		&gameast.Damage{
			Amount: *gameast.NumInt(3),
			Target: gameast.TargetOpponent(),
		}, "instant")

	if err := CastSpell(gs, 0, bolt, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	got, ok := gs.CardFirstPlayed["Lightning Bolt"]
	if !ok {
		t.Fatalf("expected Lightning Bolt in CardFirstPlayed map")
	}
	if got != 3 {
		t.Errorf("expected first-played turn 3, got %d", got)
	}
}

func TestCardFirstPlayed_RecastDoesNotOverwrite(t *testing.T) {
	gs := newFixtureGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 10
	gs.Turn = 2

	bolt1 := addHandCardWithEffect(gs, 0, "Lightning Bolt", 1,
		&gameast.Damage{Amount: *gameast.NumInt(3), Target: gameast.TargetOpponent()},
		"instant")
	if err := CastSpell(gs, 0, bolt1, nil); err != nil {
		t.Fatalf("first cast: %v", err)
	}

	gs.Turn = 5
	bolt2 := addHandCardWithEffect(gs, 0, "Lightning Bolt", 1,
		&gameast.Damage{Amount: *gameast.NumInt(3), Target: gameast.TargetOpponent()},
		"instant")
	if err := CastSpell(gs, 0, bolt2, nil); err != nil {
		t.Fatalf("second cast: %v", err)
	}

	if got := gs.CardFirstPlayed["Lightning Bolt"]; got != 2 {
		t.Errorf("expected first-played turn 2 to survive recast, got %d", got)
	}
}

func TestCardFirstPlayed_CounteredSpellNotRecorded(t *testing.T) {
	gs := newFixtureGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 1
	gs.Turn = 4

	bolt := addHandCardWithEffect(gs, 0, "Lightning Bolt", 1,
		&gameast.Damage{Amount: *gameast.NumInt(3), Target: gameast.TargetOpponent()},
		"instant")
	if err := CastSpell(gs, 0, bolt, nil); err == nil {
		// CastSpell drives stack push + resolution — to test counter
		// behavior we instead push manually and flip Countered.
	}

	// Reset state and push a stack item we can mark Countered before
	// resolving — this hits the counter branch in ResolveStackTop.
	gs2 := newFixtureGame(t)
	gs2.Active = 0
	gs2.Turn = 4
	bolt2 := addHandCardWithEffect(gs2, 0, "Counterbait", 1,
		&gameast.Damage{Amount: *gameast.NumInt(1), Target: gameast.TargetOpponent()},
		"instant")
	gs2.Stack = append(gs2.Stack, &StackItem{
		Card:       bolt2,
		Controller: 0,
		Countered:  true,
	})
	ResolveStackTop(gs2)
	if _, ok := gs2.CardFirstPlayed["Counterbait"]; ok {
		t.Errorf("countered spell should not be recorded as first-played")
	}
}
