package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Class tests — CR §716
// ---------------------------------------------------------------------------

func newClassGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(47))
	gs := NewGameState(2, rng, nil)
	gs.Active = 0
	gs.Step = "precombat_main"
	return gs
}

// newClassCard builds an AFR-shape Class enchantment with 3 levels.
// The AST encodes the band metadata using the same shape the parser
// emits (Static + Modification with ModKind="class_level_band" and
// Args=[lo, hi]) so parseLevelBracketsFromAST reads them correctly.
//
// Class abilities are CUMULATIVE — once unlocked, a lower-level
// ability remains active at higher levels (Cleric Class's "spells
// that target a creature cost {1} less" stays active at lvl 2/3).
// To model that, each Class band is open-ended on the high side
// (args = [N, nil] → MaxLevel = -1). This differs from Level Up
// creatures where each bracket is exclusive — same bracket encoding,
// different semantics decided by the consuming code.
//
// Keywords listed between bands are attached to the most recent
// open band per parseLevelBracketsFromAST's accumulator.
func newClassCard(name string, owner int) *Card {
	return &Card{
		Name:  name,
		Owner: owner,
		Types: []string{"enchantment", "class"},
		CMC:   2,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				// Level 1 (base) band — open-ended cumulative.
				&gameast.Static{
					Modification: &gameast.Modification{
						ModKind: "class_level_band",
						Args:    []interface{}{1, nil},
					},
				},
				// Level 2 band — open-ended cumulative.
				&gameast.Static{
					Modification: &gameast.Modification{
						ModKind: "class_level_band",
						Args:    []interface{}{2, nil},
					},
				},
				// Keyword for level-2 band: "vigilance" (granted while level >= 2).
				&gameast.Keyword{Name: "vigilance"},
				// Level 3 band — open-ended cumulative.
				&gameast.Static{
					Modification: &gameast.Modification{
						ModKind: "class_level_band",
						Args:    []interface{}{3, nil},
					},
				},
				// Keyword for level-3 band: "trample".
				&gameast.Keyword{Name: "trample"},
			},
		},
	}
}

func newPlainEnchantmentForClass(name string, owner int) *Card {
	return &Card{
		Name:  name,
		Owner: owner,
		Types: []string{"enchantment"},
		AST:   &gameast.CardAST{Name: name},
	}
}

func putClassBattlefield(gs *GameState, seat int, card *Card) *Permanent {
	perm := &Permanent{
		Card:       card,
		Controller: seat,
		Owner:      seat,
		Timestamp:  gs.NextTimestamp(),
	}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, perm)
	return perm
}

// ---------------------------------------------------------------------------
// HasClass detector
// ---------------------------------------------------------------------------

func TestHasClass_Detects(t *testing.T) {
	if !HasClass(newClassCard("Cleric Class", 0)) {
		t.Fatal("HasClass should be true for an Enchantment — Class card")
	}
}

func TestHasClass_Negative(t *testing.T) {
	if HasClass(nil) {
		t.Fatal("HasClass(nil) should be false")
	}
	if HasClass(newPlainEnchantmentForClass("Plain Enchantment", 0)) {
		t.Fatal("HasClass should be false for a non-class enchantment")
	}
}

// ---------------------------------------------------------------------------
// (a) Class ETBs at level 1
// ---------------------------------------------------------------------------

func TestClassLevel_ETBsAtLevelOne(t *testing.T) {
	gs := newClassGame(t)
	perm := putClassBattlefield(gs, 0, newClassCard("Cleric Class", 0))
	if got := ClassLevel(perm); got != 1 {
		t.Fatalf("ClassLevel at ETB = %d, want 1 (§716.1)", got)
	}
}

func TestClassLevel_NonClassReadsZero(t *testing.T) {
	gs := newClassGame(t)
	perm := putClassBattlefield(gs, 0, newPlainEnchantmentForClass("Plain", 0))
	if got := ClassLevel(perm); got != 0 {
		t.Fatalf("ClassLevel on non-class permanent = %d, want 0", got)
	}
}

func TestMaxClassLevel_DerivesFromBands(t *testing.T) {
	if got := MaxClassLevel(newClassCard("Cleric Class", 0)); got != 3 {
		t.Fatalf("MaxClassLevel = %d, want 3 (highest band MaxLevel)", got)
	}
}

func TestMaxClassLevel_DefaultsToThreeForBandlessClass(t *testing.T) {
	bandless := &Card{
		Name:  "Bandless Class",
		Owner: 0,
		Types: []string{"enchantment", "class"},
		AST:   &gameast.CardAST{Name: "Bandless Class"},
	}
	if got := MaxClassLevel(bandless); got != 3 {
		t.Fatalf("MaxClassLevel default = %d, want 3", got)
	}
}

// ---------------------------------------------------------------------------
// (b) LevelUpClass costs + advances + fires class_level_up trigger
// ---------------------------------------------------------------------------

func TestLevelUpClass_AdvancesAndPays(t *testing.T) {
	gs := newClassGame(t)
	gs.Seats[0].ManaPool = 5
	perm := putClassBattlefield(gs, 0, newClassCard("Cleric Class", 0))

	captured, restore := installCapturingTriggerHook(t)
	defer restore()

	if !LevelUpClass(gs, perm, 2) {
		t.Fatal("LevelUpClass should report true on a successful advance")
	}
	if got := ClassLevel(perm); got != 2 {
		t.Fatalf("ClassLevel after first level-up = %d, want 2", got)
	}
	if gs.Seats[0].ManaPool != 3 {
		t.Fatalf("mana left = %d, want 3 (paid 2 of 5)", gs.Seats[0].ManaPool)
	}
	// class_level_up trigger fired with the right ctx.
	fires := 0
	for _, c := range *captured {
		if c.event != "class_level_up" {
			continue
		}
		fires++
		if c.ctx["new_level"] != 2 {
			t.Fatalf("trigger ctx[\"new_level\"] = %v, want 2", c.ctx["new_level"])
		}
		if c.ctx["previous_level"] != 1 {
			t.Fatalf("trigger ctx[\"previous_level\"] = %v, want 1", c.ctx["previous_level"])
		}
		if c.ctx["source"] != perm {
			t.Fatal("trigger ctx[\"source\"] should reference the class perm")
		}
		if c.ctx["controller_seat"] != 0 {
			t.Fatalf("trigger ctx[\"controller_seat\"] = %v, want 0", c.ctx["controller_seat"])
		}
	}
	if fires != 1 {
		t.Fatalf("class_level_up fired %d times, want 1", fires)
	}
}

func TestLevelUpClass_AdvancesTwoToThree(t *testing.T) {
	gs := newClassGame(t)
	gs.Seats[0].ManaPool = 10
	perm := putClassBattlefield(gs, 0, newClassCard("Cleric Class", 0))

	// Two consecutive level-ups: 1→2, then 2→3.
	if !LevelUpClass(gs, perm, 2) {
		t.Fatal("first level-up should succeed")
	}
	if !LevelUpClass(gs, perm, 4) {
		t.Fatal("second level-up should succeed")
	}
	if got := ClassLevel(perm); got != 3 {
		t.Fatalf("ClassLevel after two level-ups = %d, want 3", got)
	}
	if gs.Seats[0].ManaPool != 4 {
		t.Fatalf("mana left = %d, want 4 (paid 2 + 4 of 10)", gs.Seats[0].ManaPool)
	}
}

// ---------------------------------------------------------------------------
// (c) Cannot level past max
// ---------------------------------------------------------------------------

func TestLevelUpClass_CannotLevelPastMax(t *testing.T) {
	gs := newClassGame(t)
	gs.Seats[0].ManaPool = 20
	perm := putClassBattlefield(gs, 0, newClassCard("Cleric Class", 0))
	// Walk up to max.
	LevelUpClass(gs, perm, 2)
	LevelUpClass(gs, perm, 4)
	if got := ClassLevel(perm); got != 3 {
		t.Fatalf("setup: expected level 3, got %d", got)
	}
	manaBefore := gs.Seats[0].ManaPool

	// Attempting another level-up at max should fail and NOT consume mana.
	if LevelUpClass(gs, perm, 6) {
		t.Fatal("LevelUpClass at max level should fail")
	}
	if got := ClassLevel(perm); got != 3 {
		t.Fatalf("ClassLevel after rejected level-up = %d, want 3", got)
	}
	if gs.Seats[0].ManaPool != manaBefore {
		t.Fatalf("mana changed on rejected level-up: %d → %d", manaBefore, gs.Seats[0].ManaPool)
	}
}

// ---------------------------------------------------------------------------
// (d) Sorcery speed only
// ---------------------------------------------------------------------------

func TestLevelUpClass_RejectedAtInstantSpeed(t *testing.T) {
	gs := newClassGame(t)
	gs.Seats[0].ManaPool = 5
	gs.Step = "combat_declare_attackers" // non-main phase
	perm := putClassBattlefield(gs, 0, newClassCard("Cleric Class", 0))

	if LevelUpClass(gs, perm, 2) {
		t.Fatal("LevelUpClass off-main-phase should be rejected (§716.3 sorcery-speed)")
	}
	if got := ClassLevel(perm); got != 1 {
		t.Fatalf("ClassLevel after rejected level-up = %d, want 1", got)
	}
	if gs.Seats[0].ManaPool != 5 {
		t.Fatalf("mana should be untouched on rejected level-up, got %d", gs.Seats[0].ManaPool)
	}
}

func TestLevelUpClass_RejectedOnOpponentsTurn(t *testing.T) {
	gs := newClassGame(t)
	gs.Active = 1 // opponent's turn
	gs.Seats[0].ManaPool = 5
	perm := putClassBattlefield(gs, 0, newClassCard("Cleric Class", 0))

	if LevelUpClass(gs, perm, 2) {
		t.Fatal("LevelUpClass on opponent's turn should be rejected (sorcery-speed)")
	}
}

func TestLevelUpClass_RejectedWithStackNonEmpty(t *testing.T) {
	gs := newClassGame(t)
	gs.Seats[0].ManaPool = 5
	// Drop a stack item so isSorceryTiming returns false.
	gs.Stack = append(gs.Stack, &StackItem{Card: &Card{Name: "Junk"}, Controller: 0})
	perm := putClassBattlefield(gs, 0, newClassCard("Cleric Class", 0))

	if LevelUpClass(gs, perm, 2) {
		t.Fatal("LevelUpClass with non-empty stack should be rejected")
	}
}

// ---------------------------------------------------------------------------
// (e) Static abilities from levels 2-3 active only at appropriate level
// ---------------------------------------------------------------------------

func TestClassLevelStaticActive_GatesByLevel(t *testing.T) {
	gs := newClassGame(t)
	perm := putClassBattlefield(gs, 0, newClassCard("Cleric Class", 0))

	// Level 1: only level-1 band active.
	if !ClassLevelStaticActive(perm, 1) {
		t.Fatal("level-1 band should be active at ClassLevel 1")
	}
	if ClassLevelStaticActive(perm, 2) {
		t.Fatal("level-2 band must NOT be active at ClassLevel 1")
	}
	if ClassLevelStaticActive(perm, 3) {
		t.Fatal("level-3 band must NOT be active at ClassLevel 1")
	}

	// Level up to 2.
	gs.Seats[0].ManaPool = 100
	if !LevelUpClass(gs, perm, 2) {
		t.Fatal("level-up to 2 should succeed")
	}
	if !ClassLevelStaticActive(perm, 1) {
		t.Fatal("level-1 band should remain active at ClassLevel 2")
	}
	if !ClassLevelStaticActive(perm, 2) {
		t.Fatal("level-2 band should be active at ClassLevel 2")
	}
	if ClassLevelStaticActive(perm, 3) {
		t.Fatal("level-3 band must NOT be active at ClassLevel 2")
	}

	// Level up to 3.
	if !LevelUpClass(gs, perm, 4) {
		t.Fatal("level-up to 3 should succeed")
	}
	if !ClassLevelStaticActive(perm, 3) {
		t.Fatal("level-3 band should be active at ClassLevel 3")
	}
}

func TestActiveClassBrackets_ReturnsLevelBands(t *testing.T) {
	gs := newClassGame(t)
	gs.Seats[0].ManaPool = 100
	perm := putClassBattlefield(gs, 0, newClassCard("Cleric Class", 0))

	// At ClassLevel 1: only the [1,-1] band active.
	active := ActiveClassBrackets(perm)
	if len(active) != 1 || active[0].MinLevel != 1 {
		t.Fatalf("at level 1: active brackets = %v, want 1 bracket starting at 1", active)
	}

	// Level up to 2; brackets [1,-1] and [2,-1] active (cumulative).
	LevelUpClass(gs, perm, 2)
	active = ActiveClassBrackets(perm)
	if len(active) != 2 {
		t.Fatalf("at level 2: %d brackets active, want 2", len(active))
	}
	// Verify the level-2 band carries the vigilance keyword.
	foundVigilance := false
	for _, b := range active {
		for _, kw := range b.Keywords {
			if kw == "vigilance" {
				foundVigilance = true
			}
		}
	}
	if !foundVigilance {
		t.Fatal("active level-2 band should carry the vigilance keyword")
	}

	// Level up to 3; brackets [1,-1], [2,-1], [3,-1] all active (cumulative).
	LevelUpClass(gs, perm, 4)
	active = ActiveClassBrackets(perm)
	if len(active) != 3 {
		t.Fatalf("at level 3: %d brackets active, want 3", len(active))
	}
	foundTrample := false
	for _, b := range active {
		for _, kw := range b.Keywords {
			if kw == "trample" {
				foundTrample = true
			}
		}
	}
	if !foundTrample {
		t.Fatal("active level-3 band should carry the trample keyword")
	}
}

// ---------------------------------------------------------------------------
// Failure paths
// ---------------------------------------------------------------------------

func TestLevelUpClass_NonClassPermRejected(t *testing.T) {
	gs := newClassGame(t)
	gs.Seats[0].ManaPool = 5
	perm := putClassBattlefield(gs, 0, newPlainEnchantmentForClass("Plain", 0))
	if LevelUpClass(gs, perm, 2) {
		t.Fatal("LevelUpClass should reject a non-class permanent")
	}
}

func TestLevelUpClass_InsufficientMana(t *testing.T) {
	gs := newClassGame(t)
	gs.Seats[0].ManaPool = 1
	perm := putClassBattlefield(gs, 0, newClassCard("Cleric Class", 0))
	if LevelUpClass(gs, perm, 2) {
		t.Fatal("LevelUpClass should reject with insufficient mana")
	}
	if got := ClassLevel(perm); got != 1 {
		t.Fatalf("ClassLevel after rejected level-up = %d, want 1", got)
	}
	if gs.Seats[0].ManaPool != 1 {
		t.Fatalf("mana should be untouched, got %d", gs.Seats[0].ManaPool)
	}
}

func TestLevelUpClass_NegativeCostRejected(t *testing.T) {
	gs := newClassGame(t)
	gs.Seats[0].ManaPool = 5
	perm := putClassBattlefield(gs, 0, newClassCard("Cleric Class", 0))
	if LevelUpClass(gs, perm, -1) {
		t.Fatal("LevelUpClass with negative cost should be rejected")
	}
}
