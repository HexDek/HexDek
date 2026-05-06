package gameengine

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// TestResolvePermanentSpell_StampsWasCast confirms that a permanent spell
// resolved through the stack pipeline carries was_cast=1 and
// cast_from_hand=1 on its Flags map. This is the foundation that the
// "if you cast it" intervening-if condition (CR §603.6c) reads off.
func TestResolvePermanentSpell_StampsWasCast(t *testing.T) {
	gs := newFixtureGame(t)
	gs.Seats[0].Hat = &GreedyHatStub{}
	gs.Seats[0].ManaPool = 4
	card := &Card{
		Name:  "Test Artifact",
		Owner: 0,
		Types: []string{"artifact"},
	}
	gs.Seats[0].Hand = []*Card{card}

	if err := CastSpell(gs, 0, card, nil); err != nil {
		t.Fatalf("CastSpell failed: %v", err)
	}

	if len(gs.Seats[0].Battlefield) != 1 {
		t.Fatalf("expected 1 permanent on battlefield, got %d", len(gs.Seats[0].Battlefield))
	}
	perm := gs.Seats[0].Battlefield[0]
	if perm.Flags["was_cast"] != 1 {
		t.Errorf("expected was_cast=1, got %d", perm.Flags["was_cast"])
	}
	if perm.Flags["cast_from_hand"] != 1 {
		t.Errorf("expected cast_from_hand=1, got %d", perm.Flags["cast_from_hand"])
	}
}

// TestEvalCondition_SelfWasCast verifies the new condition kinds.
func TestEvalCondition_SelfWasCast(t *testing.T) {
	gs := newFixtureGame(t)
	cast := &Permanent{Flags: map[string]int{"was_cast": 1, "cast_from_hand": 1}}
	notCast := &Permanent{Flags: map[string]int{}}

	condCast := &gameast.Condition{Kind: "self_was_cast"}
	condCastFromHand := &gameast.Condition{Kind: "self_was_cast", Args: []interface{}{true}}
	condNotCast := &gameast.Condition{Kind: "self_was_not_cast"}

	if !evalCondition(gs, cast, condCast) {
		t.Error("self_was_cast should be true for cast permanent")
	}
	if evalCondition(gs, notCast, condCast) {
		t.Error("self_was_cast should be false for non-cast permanent")
	}
	if !evalCondition(gs, cast, condCastFromHand) {
		t.Error("self_was_cast(fromHand=true) should be true when cast_from_hand is set")
	}
	if evalCondition(gs, notCast, condCastFromHand) {
		t.Error("self_was_cast(fromHand=true) should be false for non-cast permanent")
	}
	if evalCondition(gs, cast, condNotCast) {
		t.Error("self_was_not_cast should be false for cast permanent")
	}
	if !evalCondition(gs, notCast, condNotCast) {
		t.Error("self_was_not_cast should be true for non-cast permanent")
	}
}

// TestParseConditionText_CastVariants verifies the parser recognises the
// known oracle phrasings.
func TestParseConditionText_CastVariants(t *testing.T) {
	cases := []struct {
		text string
		kind string
	}{
		{"you cast it", "self_was_cast"},
		{"you cast it from your hand", "self_was_cast"},
		{"you cast this", "self_was_cast"},
		{"it was cast", "self_was_cast"},
		{"it was cast from your hand", "self_was_cast"},
		{"it wasn't cast", "self_was_not_cast"},
		{"it was not cast", "self_was_not_cast"},
		{"you didn't cast it", "self_was_not_cast"},
	}
	for _, tc := range cases {
		t.Run(tc.text, func(t *testing.T) {
			c := parseConditionText(tc.text)
			if c == nil {
				t.Fatalf("parseConditionText(%q) returned nil", tc.text)
			}
			if c.Kind != tc.kind {
				t.Errorf("parseConditionText(%q).Kind = %q, want %q", tc.text, c.Kind, tc.kind)
			}
		})
	}
}

// TestResolveConditionalEffect_OnCast end-to-end: a permanent cast through
// the stack pipeline → ResolveEffect on its conditional_effect → the "if
// you cast it" branch fires (logs conditional_effect_fires) and the inner
// Draw effect resolves.
func TestResolveConditionalEffect_OnCast(t *testing.T) {
	gs := newFixtureGame(t)
	gs.Seats[0].Hat = &GreedyHatStub{}
	addLibrary(gs, 0, "Card1", "Card2")
	card := &Card{
		Name:  "Test Permanent",
		Owner: 0,
		Types: []string{"artifact"},
	}
	gs.Seats[0].ManaPool = 4
	gs.Seats[0].Hand = []*Card{card}
	if err := CastSpell(gs, 0, card, nil); err != nil {
		t.Fatalf("CastSpell failed: %v", err)
	}
	perm := gs.Seats[0].Battlefield[0]

	// Now apply a conditional_effect ModificationEffect simulating
	// "if you cast it, draw a card" on the permanent.
	eff := &gameast.ModificationEffect{
		ModKind: "conditional_effect",
		Args:    []interface{}{"if you cast it, draw a card"},
	}
	before := len(gs.Seats[0].Hand)
	ResolveEffect(gs, perm, eff)
	after := len(gs.Seats[0].Hand)
	if after != before+1 {
		t.Errorf("expected hand size to grow by 1, before=%d after=%d", before, after)
	}
}
