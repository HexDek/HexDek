package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Replicate — CR §702.56 / §706.10 integration tests
// ---------------------------------------------------------------------------

// newReplicateGame builds a deterministic 2-seat game suitable for spell-cast
// stack manipulation. The library is pre-populated with vanilla cards so any
// resolution-side card-draw can succeed without empty-library effects.
func newReplicateGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(99))
	gs := NewGameState(2, rng, nil)
	for s := 0; s < 2; s++ {
		seat := gs.Seats[s]
		for i := 0; i < 20; i++ {
			seat.Library = append(seat.Library, &Card{Name: "Forest", Owner: s, Types: []string{"land"}})
		}
	}
	gs.Active = 0
	return gs
}

// newCacklingCounterpart builds the engine's test stand-in for Cackling
// Counterpart — Sorin's blue instant that creates a token copy of a creature.
// For replicate validation we override the printed text and attach a
// `replicate {3}` keyword plus a Draw-1 effect, so each resolution is
// individually observable (one draw per copy that resolves).
func newCacklingCounterpart(owner int) *Card {
	return &Card{
		Name:  "Cackling Counterpart",
		Owner: owner,
		Types: []string{"instant"},
		CMC:   3,
		Colors: []string{"U"},
		AST: &gameast.CardAST{
			Name: "Cackling Counterpart",
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "replicate", Args: []any{3}},
			},
		},
	}
}

// pushReplicateSpell shoves a fully-formed spell stack item for `card` with
// the given effect onto the stack so the replicate path can be exercised in
// isolation (without depending on the full CastSpellWithCosts pipeline).
func pushReplicateSpell(gs *GameState, card *Card, controller int, eff gameast.Effect) *StackItem {
	item := &StackItem{
		Controller: controller,
		Card:       card,
		Effect:     eff,
		Kind:       "spell",
	}
	item.ID = nextStackID(gs)
	gs.Stack = append(gs.Stack, item)
	return item
}

// ---------------------------------------------------------------------------
// HasReplicate / ReplicateCost
// ---------------------------------------------------------------------------

func TestHasReplicate_DetectsKeyword(t *testing.T) {
	c := newCacklingCounterpart(0)
	if !HasReplicate(c) {
		t.Fatal("HasReplicate returned false for a card with replicate keyword")
	}
	if got := ReplicateCost(c); got != 3 {
		t.Fatalf("ReplicateCost = %d, want 3", got)
	}
}

func TestHasReplicate_NegativeAndNil(t *testing.T) {
	if HasReplicate(nil) {
		t.Fatal("HasReplicate(nil) should be false")
	}
	plain := &Card{
		Name: "Lightning Bolt",
		AST:  &gameast.CardAST{Name: "Lightning Bolt"},
	}
	if HasReplicate(plain) {
		t.Fatal("HasReplicate returned true for a non-replicate card")
	}
	if ReplicateCost(plain) != 0 {
		t.Fatal("ReplicateCost returned non-zero for a non-replicate card")
	}
}

// ---------------------------------------------------------------------------
// ApplyReplicate — cost payment + stack shape
// ---------------------------------------------------------------------------

func TestApplyReplicate_PaysCostNTimesAndCopiesShareCharacteristics(t *testing.T) {
	gs := newReplicateGame(t)
	seat := gs.Seats[0]
	seat.ManaPool = 12 // 3 (cast) + 6 (replicate ×2) + slack

	card := newCacklingCounterpart(0)
	// Drop the printed cost off the pool so the pool entering ApplyReplicate
	// represents "post-cast" mana — the replicate cost is the only thing left
	// to pay here. CR §702.56b says replicate is an additional cost paid as
	// the spell is being cast; we model it as a discrete step that ApplyReplicate
	// handles after the initial cast push.
	seat.ManaPool -= card.CMC // simulated cast cost
	manaBefore := seat.ManaPool

	target := Target{Kind: TargetKindSeat, Seat: 1}
	item := pushReplicateSpell(gs, card, 0, &gameast.Draw{Count: gameast.NumberOrRef{IsInt: true, Int: 1}})
	item.Targets = []Target{target}

	placed := ApplyReplicate(gs, item, 2)
	if placed != 2 {
		t.Fatalf("ApplyReplicate returned %d, want 2", placed)
	}

	// (1) Cost paid 2× the replicate cost (3 each → 6 total).
	wantPaid := 6
	if got := manaBefore - seat.ManaPool; got != wantPaid {
		t.Fatalf("mana paid = %d, want %d", got, wantPaid)
	}

	// (2) Stack: original (index 0) + 2 copies (indices 1, 2).
	if len(gs.Stack) != 3 {
		t.Fatalf("stack size = %d, want 3 (original + 2 copies)", len(gs.Stack))
	}
	if gs.Stack[0] != item {
		t.Fatal("original stack item should remain at index 0; copies append above")
	}
	if gs.Stack[0].IsCopy {
		t.Fatal("original spell must not be flagged IsCopy")
	}

	// (3) Each copy: IsCopy=true, same controller, same name + CMC + types
	//     + AST + targets per CR §706.10c. Distinct stack IDs.
	seenIDs := map[int]bool{item.ID: true}
	for i := 1; i <= 2; i++ {
		c := gs.Stack[i]
		if c == nil || c.Card == nil {
			t.Fatalf("copy %d: nil stack item or nil Card", i)
		}
		if !c.IsCopy {
			t.Fatalf("copy %d: IsCopy=false, want true (CR §706.10)", i)
		}
		if c.Controller != 0 {
			t.Fatalf("copy %d: Controller=%d, want 0 (CR §706.10b)", i, c.Controller)
		}
		if c.Card.Name != "Cackling Counterpart" {
			t.Fatalf("copy %d: Name=%q, want %q (CR §706.10c — same characteristics)", i, c.Card.Name, "Cackling Counterpart")
		}
		if c.Card.CMC != card.CMC {
			t.Fatalf("copy %d: CMC=%d, want %d (CR §706.10c)", i, c.Card.CMC, card.CMC)
		}
		if c.Card.AST != card.AST {
			t.Fatalf("copy %d: AST not shared with original", i)
		}
		if len(c.Targets) != 1 || c.Targets[0] != target {
			t.Fatalf("copy %d: Targets=%v, want %v (CR §706.10f)", i, c.Targets, []Target{target})
		}
		if seenIDs[c.ID] {
			t.Fatalf("copy %d: duplicate stack ID %d", i, c.ID)
		}
		seenIDs[c.ID] = true
	}
}

// ---------------------------------------------------------------------------
// ApplyReplicate — insufficient mana is a no-op (no partial copies)
// ---------------------------------------------------------------------------

func TestApplyReplicate_InsufficientManaPlacesNoCopies(t *testing.T) {
	gs := newReplicateGame(t)
	seat := gs.Seats[0]
	seat.ManaPool = 4 // only enough for one extra copy (3), not two (6)

	card := newCacklingCounterpart(0)
	item := pushReplicateSpell(gs, card, 0, &gameast.Draw{Count: gameast.NumberOrRef{IsInt: true, Int: 1}})

	manaBefore := seat.ManaPool
	stackBefore := len(gs.Stack)

	placed := ApplyReplicate(gs, item, 2)
	if placed != 0 {
		t.Fatalf("ApplyReplicate with insufficient mana returned %d, want 0", placed)
	}
	if seat.ManaPool != manaBefore {
		t.Fatalf("mana pool changed on failure: %d → %d", manaBefore, seat.ManaPool)
	}
	if len(gs.Stack) != stackBefore {
		t.Fatalf("stack changed on failure: %d → %d items", stackBefore, len(gs.Stack))
	}
}

// ---------------------------------------------------------------------------
// Full integration — Cackling Counterpart with replicate=2 resolves 3×
// ---------------------------------------------------------------------------

// TestApplyReplicate_CacklingCounterpartTwoCopiesResolve is the canonical
// end-to-end check: the original spell plus two replicate copies must each
// resolve their effect, and the two copies must "cease to exist" rather
// than route to graveyard (CR §706.10 — they are not cards in any deck).
//
// We attach a Draw-1 effect so resolutions are observable: 1 cast + 2
// replicate copies = 3 cards drawn from the controller's library.
func TestApplyReplicate_CacklingCounterpartTwoCopiesResolve(t *testing.T) {
	gs := newReplicateGame(t)
	seat := gs.Seats[0]
	seat.ManaPool = 9 // 3 cast + 6 replicate
	// Clear hand so the post-resolution count is purely from the spell.
	seat.Hand = nil

	card := newCacklingCounterpart(0)
	// Simulated cast: pay base cost, push onto stack.
	seat.ManaPool -= card.CMC
	item := pushReplicateSpell(gs, card, 0, &gameast.Draw{Count: gameast.NumberOrRef{IsInt: true, Int: 1}})

	if placed := ApplyReplicate(gs, item, 2); placed != 2 {
		t.Fatalf("ApplyReplicate placed %d copies, want 2", placed)
	}

	// Mana fully spent: 9 - 3 (cast) - 6 (replicate ×2) = 0.
	if seat.ManaPool != 0 {
		t.Fatalf("mana pool = %d after cast + replicate, want 0", seat.ManaPool)
	}

	// Resolve the stack top-down. Three pops total: copy, copy, original.
	if len(gs.Stack) != 3 {
		t.Fatalf("pre-resolve stack size = %d, want 3", len(gs.Stack))
	}
	for i := 0; i < 3; i++ {
		ResolveStackTop(gs)
		StateBasedActions(gs)
	}
	if len(gs.Stack) != 0 {
		t.Fatalf("post-resolve stack size = %d, want 0", len(gs.Stack))
	}

	// Three resolutions = three cards drawn into hand.
	if got := len(seat.Hand); got != 3 {
		t.Fatalf("hand size after resolution = %d, want 3 (1 cast + 2 replicate copies)", got)
	}

	// The original spell card goes to the graveyard (CR §608.2g).
	foundInGY := false
	for _, c := range seat.Graveyard {
		if c == card {
			foundInGY = true
			break
		}
	}
	if !foundInGY {
		t.Fatalf("original Cackling Counterpart not found in graveyard after resolution")
	}

	// The copies must NOT appear in graveyard, library, hand, or exile —
	// they cease to exist per CR §706.10. Scan all zones for any card whose
	// Name matches the original but is not the original instance.
	check := func(zone string, cards []*Card) {
		for _, c := range cards {
			if c == nil || c == card {
				continue
			}
			if c.Name == "Cackling Counterpart" {
				t.Fatalf("copy of Cackling Counterpart leaked into %s (zone conservation violation)", zone)
			}
		}
	}
	check("hand", seat.Hand)
	check("graveyard", seat.Graveyard)
	check("library", seat.Library)
	check("exile", seat.Exile)
}
