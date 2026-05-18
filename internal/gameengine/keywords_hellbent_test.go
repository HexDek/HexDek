package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// Round-29 tests for Hellbent (CR §702.45). Mirrors the Threshold /
// Metalcraft predicate pattern.

func hb_makeGame(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(2, rand.New(rand.NewSource(45)), nil)
}

func hb_makeCard(name string) *Card {
	return &Card{
		Name:  name,
		Owner: 0,
		Types: []string{"creature"},
	}
}

// hb_makeHellbentCard builds a card with a Static ability whose Raw
// text carries the printed Hellbent rider — enough for the oracle-
// text detector path to fire (OracleTextLower reconstructs from
// AST.Abilities raw strings).
func hb_makeHellbentCard(name, riderText string) *Card {
	return &Card{
		Name:          name,
		Owner:         0,
		Types:         []string{"creature"},
		BasePower:     2,
		BaseToughness: 2,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: riderText},
			},
		},
	}
}

// hb_makeKeywordHellbent uses the AST keyword tagging path.
func hb_makeKeywordHellbent(name string) *Card {
	return &Card{
		Name:  name,
		Owner: 0,
		Types: []string{"creature"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "hellbent", Raw: "hellbent"},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// (a) HellbentActive true at 0 cards in hand.
// ---------------------------------------------------------------------------

func TestHellbentActive_TrueAtZeroCards(t *testing.T) {
	gs := hb_makeGame(t)
	gs.Seats[0].Hand = nil
	if !HellbentActive(gs, 0) {
		t.Fatal("HellbentActive should be true with 0 cards in hand")
	}
}

// ---------------------------------------------------------------------------
// (b) False at 1+ cards.
// ---------------------------------------------------------------------------

func TestHellbentActive_FalseAtOnePlusCards(t *testing.T) {
	gs := hb_makeGame(t)
	gs.Seats[0].Hand = []*Card{hb_makeCard("Anything")}
	if HellbentActive(gs, 0) {
		t.Fatal("HellbentActive should be false with 1 card in hand")
	}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand,
		hb_makeCard("Second"), hb_makeCard("Third"))
	if HellbentActive(gs, 0) {
		t.Fatal("HellbentActive should be false with 3 cards in hand")
	}
}

// ---------------------------------------------------------------------------
// (c) HasHellbent detects 'hellbent' in oracle.
// ---------------------------------------------------------------------------

func TestHasHellbent_DetectsOracleEmDash(t *testing.T) {
	c := hb_makeHellbentCard("Demonfire",
		"Demonfire deals X damage to any target. Hellbent — If you have no cards in hand, Demonfire can't be countered.")
	if !HasHellbent(c) {
		t.Fatal("HasHellbent should detect em-dash rider in oracle text")
	}
}

func TestHasHellbent_DetectsOracleAsciiHyphen(t *testing.T) {
	c := hb_makeHellbentCard("Anger of the Gods",
		"Hellbent - Creatures you control get +1/+0.")
	if !HasHellbent(c) {
		t.Fatal("HasHellbent should detect ASCII-hyphen rider")
	}
}

func TestHasHellbent_DetectsKeywordTagged(t *testing.T) {
	c := hb_makeKeywordHellbent("Tagged Hellbent")
	if !HasHellbent(c) {
		t.Fatal("HasHellbent should detect AST keyword tag")
	}
}

func TestHasHellbent_NegativeOnPlainCard(t *testing.T) {
	c := hb_makeHellbentCard("Plain", "Plain has flying.")
	if HasHellbent(c) {
		t.Fatal("HasHellbent must be false on a card without the rider")
	}
	if HasHellbent(nil) {
		t.Fatal("HasHellbent(nil) should be false")
	}
}

// ---------------------------------------------------------------------------
// (d) Flips dynamically as cards enter/leave hand.
// ---------------------------------------------------------------------------

func TestHellbentActive_FlipsDynamically(t *testing.T) {
	gs := hb_makeGame(t)

	// Start empty → true.
	gs.Seats[0].Hand = nil
	if !HellbentActive(gs, 0) {
		t.Fatal("step 1: should be true with empty hand")
	}

	// Add card → false.
	c := hb_makeCard("Lightning Bolt")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, c)
	if HellbentActive(gs, 0) {
		t.Fatal("step 2: should flip to false when card added")
	}

	// Remove card → true again.
	gs.Seats[0].Hand = gs.Seats[0].Hand[:0]
	if !HellbentActive(gs, 0) {
		t.Fatal("step 3: should flip back to true when hand emptied")
	}

	// Add multiple, then remove all but one.
	gs.Seats[0].Hand = []*Card{hb_makeCard("A"), hb_makeCard("B"), hb_makeCard("C")}
	if HellbentActive(gs, 0) {
		t.Fatal("step 4: should be false with 3 cards")
	}
	gs.Seats[0].Hand = gs.Seats[0].Hand[:1]
	if HellbentActive(gs, 0) {
		t.Fatal("step 5: should still be false with 1 card")
	}
	gs.Seats[0].Hand = nil
	if !HellbentActive(gs, 0) {
		t.Fatal("step 6: should flip back to true when all cards removed")
	}
}

// ---------------------------------------------------------------------------
// (e) Per-seat independence.
// ---------------------------------------------------------------------------

func TestHellbentActive_PerSeatIndependent(t *testing.T) {
	gs := hb_makeGame(t)
	gs.Seats[0].Hand = nil
	gs.Seats[1].Hand = []*Card{hb_makeCard("X"), hb_makeCard("Y")}

	if !HellbentActive(gs, 0) {
		t.Fatal("seat 0 with 0 cards should be hellbent")
	}
	if HellbentActive(gs, 1) {
		t.Fatal("seat 1 with 2 cards should NOT be hellbent")
	}

	// Swap.
	gs.Seats[0].Hand = []*Card{hb_makeCard("Z")}
	gs.Seats[1].Hand = nil
	if HellbentActive(gs, 0) {
		t.Fatal("seat 0 with 1 card should NOT be hellbent")
	}
	if !HellbentActive(gs, 1) {
		t.Fatal("seat 1 with 0 cards should be hellbent")
	}
}

// ---------------------------------------------------------------------------
// Bonus: IsHellbent (Permanent-facing) routes through controller speed.
// ---------------------------------------------------------------------------

func TestIsHellbent_RoutesThroughController(t *testing.T) {
	gs := hb_makeGame(t)
	gs.Seats[0].Hand = nil
	gs.Seats[1].Hand = []*Card{hb_makeCard("Block")}

	pSeat0 := &Permanent{Controller: 0, Owner: 0, Card: hb_makeCard("Source")}
	pSeat1 := &Permanent{Controller: 1, Owner: 1, Card: hb_makeCard("Source")}

	if !IsHellbent(gs, pSeat0) {
		t.Fatal("IsHellbent for seat 0 controller (empty hand) should be true")
	}
	if IsHellbent(gs, pSeat1) {
		t.Fatal("IsHellbent for seat 1 controller (1 card) should be false")
	}
	if IsHellbent(gs, nil) {
		t.Fatal("IsHellbent(nil) should be false")
	}
}

// Bonus: nil safety.
func TestHellbentActive_NilSafe(t *testing.T) {
	if HellbentActive(nil, 0) {
		t.Fatal("HellbentActive(nil) should be false")
	}
	gs := hb_makeGame(t)
	if HellbentActive(gs, -1) {
		t.Fatal("HellbentActive with negative seat should be false")
	}
	if HellbentActive(gs, 99) {
		t.Fatal("HellbentActive with out-of-range seat should be false")
	}
}

// Bonus: evalCondition routes "hellbent" through the predicate.
func TestEvalCondition_HellbentRouting(t *testing.T) {
	gs := hb_makeGame(t)
	src := &Permanent{Controller: 0, Owner: 0, Card: hb_makeCard("Src")}

	gs.Seats[0].Hand = nil
	if !evalCondition(gs, src, &gameast.Condition{Kind: "hellbent"}) {
		t.Fatal("evalCondition(hellbent) should be true with empty hand")
	}
	gs.Seats[0].Hand = []*Card{hb_makeCard("Anything")}
	if evalCondition(gs, src, &gameast.Condition{Kind: "hellbent"}) {
		t.Fatal("evalCondition(hellbent) should be false with 1 card in hand")
	}
}
