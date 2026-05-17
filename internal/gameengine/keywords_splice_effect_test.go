package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Splice — CR §702.47 effect-merging tests
//
// These tests are the regression coverage for the round-23 fix: ApplySplice
// must actually MERGE the spliced card's effect into the resolving spell,
// not just pay the cost. Per §702.47b, "add the spliced card's text" and
// "the card stays in your hand."
// ---------------------------------------------------------------------------

func newSpliceGame(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(2, rand.New(rand.NewSource(47)), nil)
}

// arcaneSpellCard builds an Arcane sorcery whose body draws N cards. The
// resolver's resolveDraw defaults to the source's controller when no
// target is set, so we can read seat.Hand size to verify resolution.
func arcaneSpellCard(name string, owner, drawN int) *Card {
	return &Card{
		Name:  name,
		Owner: owner,
		Types: []string{"sorcery", "arcane"},
		CMC:   1,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Activated{Effect: &gameast.Draw{Count: gameast.NumberOrRef{IsInt: true, Int: drawN}}},
			},
		},
	}
}

// spliceRiderCard builds a card in hand that carries splice {cost} AND a
// real spell effect (draw N cards) so the merge actually contributes
// observable behavior on resolution.
func spliceRiderCard(name string, owner, cost, drawN int) *Card {
	return &Card{
		Name:  name,
		Owner: owner,
		Types: []string{"instant", "arcane"},
		CMC:   cost,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "splice", Args: []interface{}{float64(cost)}},
				&gameast.Activated{Effect: &gameast.Draw{Count: gameast.NumberOrRef{IsInt: true, Int: drawN}}},
			},
		},
	}
}

// libraryFiller stuffs the seat's library with enough dummy cards that
// draw effects in tests don't fizzle from an empty deck.
func libraryFiller(gs *GameState, seat, n int) {
	for i := 0; i < n; i++ {
		gs.Seats[seat].Library = append(gs.Seats[seat].Library, &Card{
			Name:  "Library Filler",
			Owner: seat,
			Types: []string{"creature"},
		})
	}
}

// pushSpellItem mimics what CastSpell would do: builds a StackItem with
// the card's resolvable effect attached. Tests call this so ApplySplice
// has something concrete to mutate.
func pushSpellItem(gs *GameState, seat int, card *Card) *StackItem {
	item := &StackItem{
		Card:       card,
		Controller: seat,
		Effect:     collectSpellEffect(card),
	}
	PushStackItem(gs, item)
	return item
}

// ---------------------------------------------------------------------------
// (a) Base spell resolves AND spliced rider resolves
// ---------------------------------------------------------------------------

func TestSplice_EffectMerge_BaseAndSpliceBothResolve(t *testing.T) {
	gs := newSpliceGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	libraryFiller(gs, 0, 10)

	base := arcaneSpellCard("Lava Spike", 0, 1)              // draws 1
	rider := spliceRiderCard("Glacial Ray", 0, 2, 2)          // draws 2
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, rider)

	item := pushSpellItem(gs, 0, base)
	if n := ApplySplice(gs, 0, item); n != 1 {
		t.Fatalf("expected 1 splice applied, got %d", n)
	}

	// item.Effect should now be a Sequence{base-draw, splice-draw}.
	seq, ok := item.Effect.(*gameast.Sequence)
	if !ok {
		t.Fatalf("expected item.Effect to be a Sequence after splice, got %T", item.Effect)
	}
	if len(seq.Items) != 2 {
		t.Fatalf("expected Sequence with 2 items (base + 1 splice), got %d", len(seq.Items))
	}

	handBefore := len(gs.Seats[0].Hand)
	ResolveStackTop(gs)

	// Base (draw 1) + splice (draw 2) = 3 cards drawn. The rider card is
	// still in hand on top of the 3 new draws.
	if got := len(gs.Seats[0].Hand) - handBefore; got != 3 {
		t.Fatalf("expected 3 cards drawn (1 base + 2 splice), got %d", got)
	}
}

// ---------------------------------------------------------------------------
// (b) Spliced card remains in hand after resolution
// ---------------------------------------------------------------------------

func TestSplice_RiderRemainsInHandAfterResolution(t *testing.T) {
	gs := newSpliceGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	libraryFiller(gs, 0, 10)

	base := arcaneSpellCard("Lava Spike", 0, 1)
	rider := spliceRiderCard("Glacial Ray", 0, 2, 1)
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, rider)

	item := pushSpellItem(gs, 0, base)
	ApplySplice(gs, 0, item)
	ResolveStackTop(gs)

	// §702.47b — spliced card must remain in hand.
	foundInHand := false
	for _, c := range gs.Seats[0].Hand {
		if c == rider {
			foundInHand = true
			break
		}
	}
	if !foundInHand {
		t.Fatal("§702.47b: spliced rider should remain in the caster's hand after resolution")
	}
	// And not in graveyard or exile.
	for _, c := range gs.Seats[0].Graveyard {
		if c == rider {
			t.Fatal("spliced rider must not enter graveyard")
		}
	}
	for _, c := range gs.Seats[0].Exile {
		if c == rider {
			t.Fatal("spliced rider must not enter exile")
		}
	}
}

// ---------------------------------------------------------------------------
// (c) Splicing onto a copy doesn't double-apply
// ---------------------------------------------------------------------------

func TestSplice_OnCopyIsRefused(t *testing.T) {
	gs := newSpliceGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	libraryFiller(gs, 0, 10)

	base := arcaneSpellCard("Lava Spike", 0, 1)
	rider := spliceRiderCard("Glacial Ray", 0, 2, 1)
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, rider)

	// Storm-style copy on the stack: same card but flagged IsCopy.
	copyItem := &StackItem{
		Card:       base,
		Controller: 0,
		IsCopy:     true,
		Effect:     collectSpellEffect(base),
	}
	PushStackItem(gs, copyItem)
	manaBefore := gs.Seats[0].ManaPool
	effBefore := copyItem.Effect

	if n := ApplySplice(gs, 0, copyItem); n != 0 {
		t.Fatalf("ApplySplice on a copy must return 0, got %d", n)
	}
	if gs.Seats[0].ManaPool != manaBefore {
		t.Fatalf("splice on a copy must not pay cost; mana before=%d after=%d",
			manaBefore, gs.Seats[0].ManaPool)
	}
	if copyItem.Effect != effBefore {
		t.Fatal("splice on a copy must not mutate item.Effect")
	}

	// Now verify the "doesn't double-apply" half: original cast splices
	// once; the copy that inherits the merged Effect resolves it ONCE
	// (not twice). The merged effect is a Sequence — pushing a copy
	// with that same effect resolves draw 1 + draw 1 = 2, not 4.
	origItem := pushSpellItem(gs, 0, base)
	if n := ApplySplice(gs, 0, origItem); n != 1 {
		t.Fatalf("ApplySplice on the original cast should apply 1 splice, got %d", n)
	}
	mergedEffect := origItem.Effect

	// Push a non-cast copy of the original spell with the SAME merged
	// effect — storm/Twinflame inherit the merged effect, but a second
	// splice should NOT apply.
	pseudoCopy := &StackItem{
		Card:       base,
		Controller: 0,
		IsCopy:     true,
		Effect:     mergedEffect,
	}
	PushStackItem(gs, pseudoCopy)
	manaBefore = gs.Seats[0].ManaPool

	if n := ApplySplice(gs, 0, pseudoCopy); n != 0 {
		t.Fatalf("ApplySplice on an inherited-effect copy must return 0, got %d", n)
	}
	if gs.Seats[0].ManaPool != manaBefore {
		t.Fatal("splice on inherited-effect copy must not pay cost")
	}
	// pseudoCopy.Effect should still point at exactly mergedEffect (no
	// second wrap, no rider appended).
	if pseudoCopy.Effect != mergedEffect {
		t.Fatal("splice on a copy must not mutate item.Effect pointer")
	}
}

// ---------------------------------------------------------------------------
// (d) Multiple splices stack properly
// ---------------------------------------------------------------------------

func TestSplice_MultipleSplicesStackInOrder(t *testing.T) {
	gs := newSpliceGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 10
	libraryFiller(gs, 0, 20)

	base := arcaneSpellCard("Lava Spike", 0, 1)             // draws 1
	rider1 := spliceRiderCard("Glacial Ray", 0, 2, 2)        // splice {2}, draws 2
	rider2 := spliceRiderCard("Goryo's Vengeance", 0, 3, 3)  // splice {3}, draws 3
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, rider1, rider2)

	item := pushSpellItem(gs, 0, base)
	if n := ApplySplice(gs, 0, item); n != 2 {
		t.Fatalf("expected 2 splices applied, got %d", n)
	}

	// item.Effect should be a flat Sequence{base, rider1, rider2} — NOT
	// nested Sequence{Sequence{base, rider1}, rider2}. The flat shape
	// keeps resolveSequence linear and matches §702.47c's "splice
	// multiple cards onto one spell" wording.
	seq, ok := item.Effect.(*gameast.Sequence)
	if !ok {
		t.Fatalf("expected flat Sequence after two splices, got %T", item.Effect)
	}
	if len(seq.Items) != 3 {
		t.Fatalf("expected Sequence with 3 items (base + 2 splices), got %d", len(seq.Items))
	}
	for i, it := range seq.Items {
		if _, isSeq := it.(*gameast.Sequence); isSeq {
			t.Fatalf("Sequence.Items[%d] should not itself be a Sequence (no nesting)", i)
		}
	}

	// Splice costs paid: 2 + 3 = 5 from a 10 pool.
	if gs.Seats[0].ManaPool != 5 {
		t.Fatalf("splice costs {2}+{3} should bring mana 10→5, got %d", gs.Seats[0].ManaPool)
	}

	// Both rider cards remain in hand.
	for _, want := range []*Card{rider1, rider2} {
		found := false
		for _, c := range gs.Seats[0].Hand {
			if c == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s must remain in hand after splicing", want.Name)
		}
	}

	handBefore := len(gs.Seats[0].Hand)
	ResolveStackTop(gs)
	// Base (1) + rider1 (2) + rider2 (3) = 6 cards drawn.
	if got := len(gs.Seats[0].Hand) - handBefore; got != 6 {
		t.Fatalf("expected 6 cards drawn (1+2+3), got %d", got)
	}
}

// ---------------------------------------------------------------------------
// Negative paths — make sure splice is gated
// ---------------------------------------------------------------------------

func TestSplice_NonArcaneTargetRefused(t *testing.T) {
	gs := newSpliceGame(t)
	gs.Seats[0].ManaPool = 10

	// Non-arcane spell.
	nonArcane := &Card{
		Name:  "Counterspell",
		Owner: 0,
		Types: []string{"instant"},
		AST: &gameast.CardAST{
			Name: "Counterspell",
			Abilities: []gameast.Ability{
				&gameast.Activated{Effect: &gameast.Draw{Count: gameast.NumberOrRef{IsInt: true, Int: 1}}},
			},
		},
	}
	rider := spliceRiderCard("Glacial Ray", 0, 2, 1)
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, rider)

	item := pushSpellItem(gs, 0, nonArcane)
	if n := ApplySplice(gs, 0, item); n != 0 {
		t.Fatalf("§702.47b: splice only attaches to Arcane spells; got %d splices on a non-arcane", n)
	}
	if gs.Seats[0].ManaPool != 10 {
		t.Fatal("non-arcane target must not pay splice cost")
	}
}

func TestSplice_InsufficientManaRefused(t *testing.T) {
	gs := newSpliceGame(t)
	gs.Seats[0].ManaPool = 1 // less than the rider's cost of 2

	base := arcaneSpellCard("Lava Spike", 0, 1)
	rider := spliceRiderCard("Glacial Ray", 0, 2, 1)
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, rider)

	item := pushSpellItem(gs, 0, base)
	if n := ApplySplice(gs, 0, item); n != 0 {
		t.Fatalf("expected 0 splices when caster can't afford the cost, got %d", n)
	}
	if gs.Seats[0].ManaPool != 1 {
		t.Fatal("ApplySplice must not partially pay an unaffordable splice")
	}
}

func TestSplice_RiderWithoutSpellEffectRefused(t *testing.T) {
	gs := newSpliceGame(t)
	gs.Seats[0].ManaPool = 5

	base := arcaneSpellCard("Lava Spike", 0, 1)
	// Keyword present, but no Activated effect body — collectSpellEffect
	// will return nil. Splicing such a card would burn mana for no rider.
	emptyRider := &Card{
		Name:  "Empty Splice",
		Owner: 0,
		Types: []string{"instant", "arcane"},
		AST: &gameast.CardAST{
			Name: "Empty Splice",
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "splice", Args: []interface{}{float64(2)}},
			},
		},
	}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, emptyRider)

	item := pushSpellItem(gs, 0, base)
	if n := ApplySplice(gs, 0, item); n != 0 {
		t.Fatalf("rider with no spell body should not splice; got %d", n)
	}
	if gs.Seats[0].ManaPool != 5 {
		t.Fatal("rider with no spell body must not pay splice cost")
	}
}

func TestSplice_NilSafe(t *testing.T) {
	gs := newSpliceGame(t)
	if ApplySplice(nil, 0, &StackItem{}) != 0 {
		t.Fatal("ApplySplice(nil game) should return 0")
	}
	if ApplySplice(gs, 0, nil) != 0 {
		t.Fatal("ApplySplice(nil item) should return 0")
	}
	if ApplySplice(gs, -1, &StackItem{Card: &Card{}}) != 0 {
		t.Fatal("ApplySplice with invalid seat should return 0")
	}
}
