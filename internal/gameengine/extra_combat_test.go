package gameengine

import "testing"

// TestAddExtraCombat_FIFOOrdering — the producer-order invariant that
// the whole queue model depends on. If trigger resolution adds A first
// then B, the queue MUST be [A, B] (not [B, A]) so the active player's
// chosen stack order propagates correctly through to the order of
// extra combat phases.
func TestAddExtraCombat_FIFOOrdering(t *testing.T) {
	gs := &GameState{}
	gs.AddExtraCombat(PendingExtraCombat{SourceCard: "First"})
	gs.AddExtraCombat(PendingExtraCombat{SourceCard: "Second"})
	gs.AddExtraCombat(PendingExtraCombat{SourceCard: "Third"})

	if got, want := len(gs.PendingExtraCombats), 3; got != want {
		t.Fatalf("queue length: got %d, want %d", got, want)
	}
	if gs.PendingExtraCombats[0].SourceCard != "First" {
		t.Errorf("front of queue should be First, got %q", gs.PendingExtraCombats[0].SourceCard)
	}
	if gs.PendingExtraCombats[1].SourceCard != "Second" {
		t.Errorf("middle of queue should be Second, got %q", gs.PendingExtraCombats[1].SourceCard)
	}
	if gs.PendingExtraCombats[2].SourceCard != "Third" {
		t.Errorf("back of queue should be Third, got %q", gs.PendingExtraCombats[2].SourceCard)
	}
}

// TestAddExtraCombat_NilGameState — defensive: ensure the helper
// doesn't panic on nil. Per_card handlers occasionally receive nil
// game state in test scaffolds; the helper must no-op rather than crash.
func TestAddExtraCombat_NilGameState(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AddExtraCombat on nil GameState panicked: %v", r)
		}
	}()
	var gs *GameState
	gs.AddExtraCombat(PendingExtraCombat{})
}

// TestPassesCombatRestriction_NoRestriction — empty CurrentCombatRestriction
// means we're in a vanilla extra combat (or in the normal turn's combat).
// Every permanent that canAttack passes also passes the restriction
// gate.
func TestPassesCombatRestriction_NoRestriction(t *testing.T) {
	gs := &GameState{CurrentCombatRestriction: ""}
	// nil permanent should pass when no restriction is in effect (the
	// canAttack check at the call site handles nil safety).
	if !passesCombatRestriction(gs, nil) {
		t.Error("nil permanent should pass when no restriction is in effect")
	}
}

// TestPassesCombatRestriction_UnknownTag — defensive: an unrecognized
// restriction tag should fail closed (return false). This forces
// explicit case-statement additions when introducing new tags instead
// of silently letting unknown tags pass through.
func TestPassesCombatRestriction_UnknownTag(t *testing.T) {
	gs := &GameState{CurrentCombatRestriction: "non_existent_restriction"}
	// We need a non-nil Permanent with a non-nil Card to reach the
	// switch statement — passesCombatRestriction returns false early
	// when p or p.Card is nil. Build a minimal permanent for this.
	card := &Card{}
	p := &Permanent{Card: card}
	if passesCombatRestriction(gs, p) {
		t.Error("unknown restriction tag should fail closed (return false), got true")
	}
}
