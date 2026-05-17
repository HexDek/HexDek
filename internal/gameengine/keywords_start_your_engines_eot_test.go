package gameengine

import (
	"math/rand"
	"testing"
)

// Round 23 cleanup: EndStepClearStartYourEngines reverses
// ApplyStartYourEngines at the cleanup step (CR §514, §702.179).

func newSYEGame(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(2, rand.New(rand.NewSource(7)), nil)
}

func newVehiclePerm(seat int, name string) *Permanent {
	c := &Card{
		Name:          name,
		Owner:         seat,
		Types:         []string{"artifact", "vehicle"},
		BasePower:     3,
		BaseToughness: 3,
	}
	return &Permanent{
		Card:       c,
		Controller: seat,
		Owner:      seat,
		Flags:      map[string]int{},
	}
}

func hasTypeFold(p *Permanent, want string) bool {
	if p == nil || p.Card == nil {
		return false
	}
	for _, t := range p.Card.Types {
		if t == want {
			return true
		}
	}
	return false
}

// (a) Animated vehicle returns to artifact-only at end of turn.
func TestEndStepClearStartYourEngines_RestoresType(t *testing.T) {
	gs := newSYEGame(t)
	v := newVehiclePerm(0, "Greasewrench Goblin")
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, v)

	ApplyStartYourEngines(gs, 0)
	if !hasTypeFold(v, "creature") {
		t.Fatalf("vehicle should be animated as creature, got %v", v.Card.Types)
	}

	cleared := EndStepClearStartYourEngines(gs)
	if cleared != 1 {
		t.Fatalf("expected 1 perm cleared, got %d", cleared)
	}
	if hasTypeFold(v, "creature") {
		t.Fatalf("creature type should be removed after EOT; types=%v", v.Card.Types)
	}
	if !hasTypeFold(v, "vehicle") || !hasTypeFold(v, "artifact") {
		t.Fatalf("vehicle should still be artifact+vehicle; got %v", v.Card.Types)
	}
}

// (b) Flag cleared after cleanup.
func TestEndStepClearStartYourEngines_FlagsCleared(t *testing.T) {
	gs := newSYEGame(t)
	v := newVehiclePerm(0, "Bonepile Buggy")
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, v)

	ApplyStartYourEngines(gs, 0)
	if v.Flags["start_your_engines"] != 1 {
		t.Fatal("start_your_engines flag should be set after animation")
	}
	if v.Flags["start_your_engines_added_creature"] != 1 {
		t.Fatal("added-creature marker should be set when we appended the type")
	}

	EndStepClearStartYourEngines(gs)

	if _, exists := v.Flags["start_your_engines"]; exists {
		t.Fatal("start_your_engines flag should be removed after cleanup")
	}
	if _, exists := v.Flags["start_your_engines_added_creature"]; exists {
		t.Fatal("added-creature marker should be removed after cleanup")
	}
}

// (c) Idempotent — second call is a no-op.
func TestEndStepClearStartYourEngines_Idempotent(t *testing.T) {
	gs := newSYEGame(t)
	v := newVehiclePerm(0, "Hijack Hotwirer")
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, v)

	ApplyStartYourEngines(gs, 0)

	if got := EndStepClearStartYourEngines(gs); got != 1 {
		t.Fatalf("first cleanup should clear 1, got %d", got)
	}
	typesAfter := append([]string(nil), v.Card.Types...)

	if got := EndStepClearStartYourEngines(gs); got != 0 {
		t.Fatalf("second cleanup should be a no-op (0), got %d", got)
	}
	if len(v.Card.Types) != len(typesAfter) {
		t.Fatalf("types changed across idempotent calls: before=%v after=%v", typesAfter, v.Card.Types)
	}
	for i := range typesAfter {
		if v.Card.Types[i] != typesAfter[i] {
			t.Fatalf("types diverged across idempotent calls: before=%v after=%v", typesAfter, v.Card.Types)
		}
	}
}

// Bonus: a vehicle that was already a creature (Living Metal style) at
// animation time must NOT have its creature type stripped on cleanup.
func TestEndStepClearStartYourEngines_PreservesPreExistingCreatureType(t *testing.T) {
	gs := newSYEGame(t)
	v := newVehiclePerm(0, "Living Vehicle")
	v.Card.Types = append(v.Card.Types, "creature") // already a creature pre-SYE
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, v)

	ApplyStartYourEngines(gs, 0)
	if v.Flags["start_your_engines_added_creature"] == 1 {
		t.Fatal("added-creature marker must NOT be set when creature type already present")
	}

	EndStepClearStartYourEngines(gs)
	if !hasTypeFold(v, "creature") {
		t.Fatalf("creature type must be preserved (was native); got %v", v.Card.Types)
	}
}

// Bonus: no-op safe across empty / nil battlefields.
func TestEndStepClearStartYourEngines_NilSafe(t *testing.T) {
	if got := EndStepClearStartYourEngines(nil); got != 0 {
		t.Fatalf("nil game should return 0, got %d", got)
	}
	gs := newSYEGame(t)
	if got := EndStepClearStartYourEngines(gs); got != 0 {
		t.Fatalf("empty battlefield should return 0, got %d", got)
	}
}
