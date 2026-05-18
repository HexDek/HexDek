package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Prowl tests — CR §702.74
// ---------------------------------------------------------------------------

func newProwlGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(29))
	gs := NewGameState(2, rng, nil)
	gs.Active = 0
	gs.Step = "precombat_main"
	return gs
}

// newProwlSpell makes a Rogue-typed instant with prowl {B}.
func newProwlSpell(name string, owner int, subtype string, prowlArg string) *Card {
	return &Card{
		Name:  name,
		Owner: owner,
		Types: []string{"creature", subtype},
		CMC:   3,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "prowl", Args: []any{prowlArg}},
			},
		},
	}
}

// newCreatureWithSubtype creates a creature card with a single
// creature subtype (Rogue, Goblin, etc.) for damage-source modeling.
func newCreatureWithSubtype(name string, owner int, subtype string, power, toughness int) *Card {
	return &Card{
		Name:          name,
		Owner:         owner,
		Types:         []string{"creature", subtype},
		BasePower:     power,
		BaseToughness: toughness,
		AST:           &gameast.CardAST{Name: name},
	}
}

func putProwlBattlefield(gs *GameState, seat int, card *Card) *Permanent {
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
// HasProwl / ProwlCost
// ---------------------------------------------------------------------------

func TestHasProwl_Detects(t *testing.T) {
	spell := newProwlSpell("Stinkdrinker Bandit", 0, "rogue", "{B}")
	if !HasProwl(spell) {
		t.Fatal("HasProwl should be true for a prowl spell")
	}
}

func TestHasProwl_Negative(t *testing.T) {
	if HasProwl(nil) {
		t.Fatal("HasProwl(nil) should be false")
	}
	plain := &Card{
		Name:  "Lightning Bolt",
		Owner: 0,
		Types: []string{"instant"},
		AST:   &gameast.CardAST{Name: "Lightning Bolt"},
	}
	if HasProwl(plain) {
		t.Fatal("HasProwl should be false for a card without the keyword")
	}
}

func TestProwlCost_ParsesManaString(t *testing.T) {
	spell := newProwlSpell("Stinkdrinker Bandit", 0, "rogue", "{B}")
	if got := ProwlCost(spell); got != 1 {
		t.Fatalf("ProwlCost = %d, want 1 (for {B})", got)
	}
}

// ---------------------------------------------------------------------------
// (a) Rogue dealt combat damage + Rogue-typed spell = prowl-castable
// ---------------------------------------------------------------------------

func TestCanPayProwl_SharedSubtype_Active(t *testing.T) {
	gs := newProwlGame(t)
	// Seat 0 has a Rogue attacker that just hit seat 1 in combat.
	rogue := newCreatureWithSubtype("Nightshade Stinger", 0, "rogue", 1, 1)
	attacker := putProwlBattlefield(gs, 0, rogue)

	// Drive the combat-damage-to-player path so the tracker actually
	// records "this Rogue card dealt combat damage to a player."
	applyCombatDamageToPlayer(gs, attacker, 1, 1)

	// Cast a Rogue-typed prowl spell — precondition satisfied.
	spell := newProwlSpell("Stinkdrinker Bandit", 0, "rogue", "{B}")
	if !CanPayProwl(gs, 0, spell) {
		t.Fatal("CanPayProwl should be true when a Rogue creature dealt combat damage and the spell is a Rogue")
	}
}

func TestCastWithProwl_SucceedsWithSharedSubtype(t *testing.T) {
	gs := newProwlGame(t)
	gs.Seats[0].ManaPool = 5
	rogue := newCreatureWithSubtype("Nightshade Stinger", 0, "rogue", 1, 1)
	attacker := putProwlBattlefield(gs, 0, rogue)
	applyCombatDamageToPlayer(gs, attacker, 1, 1)

	spell := newProwlSpell("Stinkdrinker Bandit", 0, "rogue", "{B}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, spell)

	if _, err := CastWithProwl(gs, 0, spell, 1); err != nil {
		t.Fatalf("CastWithProwl failed: %v", err)
	}
	if len(gs.Stack) != 1 {
		t.Fatalf("expected 1 stack item, got %d", len(gs.Stack))
	}
	// Card off hand, per-turn flag set.
	for _, c := range gs.Seats[0].Hand {
		if c == spell {
			t.Fatal("prowl-cast spell should be removed from hand")
		}
	}
	if !SpellProwledThisTurn(gs, 0) {
		t.Fatal("SpellProwledThisTurn should be true after a prowl cast")
	}
}

// ---------------------------------------------------------------------------
// (b) Goblin damage + Rogue spell = NOT prowl-castable
// ---------------------------------------------------------------------------

func TestCanPayProwl_MismatchedSubtype_Inactive(t *testing.T) {
	gs := newProwlGame(t)
	goblin := newCreatureWithSubtype("Goblin Piker", 0, "goblin", 2, 1)
	attacker := putProwlBattlefield(gs, 0, goblin)
	applyCombatDamageToPlayer(gs, attacker, 2, 1)

	spell := newProwlSpell("Stinkdrinker Bandit", 0, "rogue", "{B}")
	if CanPayProwl(gs, 0, spell) {
		t.Fatal("CanPayProwl should be false when the dealer's subtype (Goblin) doesn't match the spell's (Rogue)")
	}
}

func TestCastWithProwl_RejectedWhenInactive(t *testing.T) {
	gs := newProwlGame(t)
	gs.Seats[0].ManaPool = 5
	goblin := newCreatureWithSubtype("Goblin Piker", 0, "goblin", 2, 1)
	attacker := putProwlBattlefield(gs, 0, goblin)
	applyCombatDamageToPlayer(gs, attacker, 2, 1)

	spell := newProwlSpell("Stinkdrinker Bandit", 0, "rogue", "{B}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, spell)

	_, err := CastWithProwl(gs, 0, spell, 1)
	if err == nil {
		t.Fatal("CastWithProwl should fail when the prowl precondition isn't met")
	}
	ce, ok := err.(*CastError)
	if !ok || ce.Reason != "prowl_not_active" {
		t.Fatalf("expected CastError prowl_not_active, got %v", err)
	}
	// State preserved.
	if gs.Seats[0].ManaPool != 5 {
		t.Fatalf("mana should be untouched, got %d", gs.Seats[0].ManaPool)
	}
	if len(gs.Stack) != 0 {
		t.Fatalf("stack should be empty, got %d items", len(gs.Stack))
	}
	found := false
	for _, c := range gs.Seats[0].Hand {
		if c == spell {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("spell should remain in hand after rejected prowl cast")
	}
}

func TestCanPayProwl_NoDamageDealtThisTurn(t *testing.T) {
	gs := newProwlGame(t)
	// No combat damage at all this turn.
	spell := newProwlSpell("Stinkdrinker Bandit", 0, "rogue", "{B}")
	if CanPayProwl(gs, 0, spell) {
		t.Fatal("CanPayProwl should be false when nobody dealt combat damage this turn")
	}
}

// ---------------------------------------------------------------------------
// (c) Damage to creature (not player) does NOT enable prowl
// ---------------------------------------------------------------------------

func TestCanPayProwl_DamageToCreatureDoesNotEnable(t *testing.T) {
	gs := newProwlGame(t)
	rogue := newCreatureWithSubtype("Nightshade Stinger", 0, "rogue", 1, 1)
	attacker := putProwlBattlefield(gs, 0, rogue)
	// Opposing creature for combat-to-creature.
	bear := newCreatureWithSubtype("Grizzly Bears", 1, "bear", 2, 2)
	defender := putProwlBattlefield(gs, 1, bear)

	// Damage to a creature, NOT a player. CombatDamageBy tracker
	// should remain empty.
	applyCombatDamageToCreature(gs, attacker, 1, defender)

	spell := newProwlSpell("Stinkdrinker Bandit", 0, "rogue", "{B}")
	if CanPayProwl(gs, 0, spell) {
		t.Fatal("CanPayProwl should be false when the Rogue dealt combat damage to a creature, not a player")
	}
}

// ---------------------------------------------------------------------------
// (d) Blocked attacker that dealt 0 damage to player does not enable
// ---------------------------------------------------------------------------

func TestCanPayProwl_BlockedZeroDamageDoesNotEnable(t *testing.T) {
	gs := newProwlGame(t)
	rogue := newCreatureWithSubtype("Nightshade Stinger", 0, "rogue", 1, 1)
	putProwlBattlefield(gs, 0, rogue)

	// Blocked-to-zero attacker: the live combat path short-circuits
	// applyCombatDamageToPlayer when amount <= 0, so the tracker
	// never sees this card.
	if amount := 0; amount > 0 {
		applyCombatDamageToPlayer(gs, gs.Seats[0].Battlefield[0], amount, 1)
	}

	spell := newProwlSpell("Stinkdrinker Bandit", 0, "rogue", "{B}")
	if CanPayProwl(gs, 0, spell) {
		t.Fatal("CanPayProwl should be false when the only Rogue attacker dealt zero damage to the player")
	}
}

// ---------------------------------------------------------------------------
// (e) CostMeta stamped correctly on the stack item
// ---------------------------------------------------------------------------

func TestCastWithProwl_StampsCostMeta(t *testing.T) {
	gs := newProwlGame(t)
	gs.Seats[0].ManaPool = 5
	rogue := newCreatureWithSubtype("Nightshade Stinger", 0, "rogue", 1, 1)
	attacker := putProwlBattlefield(gs, 0, rogue)
	applyCombatDamageToPlayer(gs, attacker, 1, 1)

	spell := newProwlSpell("Stinkdrinker Bandit", 0, "rogue", "{B}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, spell)

	if _, err := CastWithProwl(gs, 0, spell, 1); err != nil {
		t.Fatalf("CastWithProwl failed: %v", err)
	}
	item := gs.Stack[0]
	if item.Card != spell {
		t.Fatal("stack item should reference the prowl-cast spell")
	}
	if item.Controller != 0 {
		t.Fatalf("Controller = %d, want 0", item.Controller)
	}
	if v, ok := item.CostMeta["prowl_cast"]; !ok || v != true {
		t.Fatalf("CostMeta[\"prowl_cast\"] = %v, want true", item.CostMeta["prowl_cast"])
	}
	if v, ok := item.CostMeta["prowl_cost"]; !ok || v != 1 {
		t.Fatalf("CostMeta[\"prowl_cost\"] = %v, want 1", item.CostMeta["prowl_cost"])
	}
	if !IsProwlCast(item) {
		t.Fatal("IsProwlCast should return true for a prowl-cast stack item")
	}
	// Prowl does not alter resolution destination.
	if ShouldExileOnResolve(item) {
		t.Fatal("prowl cast must not stamp exile_on_resolve")
	}
	if ShouldReturnToHandOnResolve(item) {
		t.Fatal("prowl cast must not stamp bought_back / return_to_hand")
	}
}

// ---------------------------------------------------------------------------
// Tracker semantics: dedupe + survives creature death
// ---------------------------------------------------------------------------

func TestCombatDamageBy_DedupesAcrossMultipleHits(t *testing.T) {
	gs := newProwlGame(t)
	rogue := newCreatureWithSubtype("Nightshade Stinger", 0, "rogue", 1, 1)
	attacker := putProwlBattlefield(gs, 0, rogue)
	applyCombatDamageToPlayer(gs, attacker, 1, 1)
	applyCombatDamageToPlayer(gs, attacker, 1, 1)
	applyCombatDamageToPlayer(gs, attacker, 1, 1)

	if got := len(gs.Seats[0].Turn.CombatDamageBy); got != 1 {
		t.Fatalf("CombatDamageBy length = %d, want 1 (dedupe per card)", got)
	}
	if gs.Seats[0].Turn.CombatDamageBy[0] != rogue {
		t.Fatal("CombatDamageBy[0] should be the rogue card")
	}
}

func TestCanPayProwl_AfterAttackerDies(t *testing.T) {
	gs := newProwlGame(t)
	rogue := newCreatureWithSubtype("Nightshade Stinger", 0, "rogue", 1, 1)
	attacker := putProwlBattlefield(gs, 0, rogue)
	applyCombatDamageToPlayer(gs, attacker, 1, 1)

	// Simulate the attacker dying after combat (it's no longer on
	// the battlefield, but the historical fact survives in
	// CombatDamageBy). The old battlefield-scanning CanCastForProwl
	// would miss this; the new CanPayProwl should not.
	gs.Seats[0].Battlefield = nil

	spell := newProwlSpell("Stinkdrinker Bandit", 0, "rogue", "{B}")
	if !CanPayProwl(gs, 0, spell) {
		t.Fatal("CanPayProwl should still be true after the damaging Rogue left the battlefield (history fact)")
	}
}

func TestTurnReset_ClearsCombatDamageBy(t *testing.T) {
	gs := newProwlGame(t)
	rogue := newCreatureWithSubtype("Nightshade Stinger", 0, "rogue", 1, 1)
	attacker := putProwlBattlefield(gs, 0, rogue)
	applyCombatDamageToPlayer(gs, attacker, 1, 1)

	if got := len(gs.Seats[0].Turn.CombatDamageBy); got != 1 {
		t.Fatalf("setup: expected 1 dealer, got %d", got)
	}
	gs.Seats[0].Turn.Reset()
	if got := len(gs.Seats[0].Turn.CombatDamageBy); got != 0 {
		t.Fatalf("after Reset, CombatDamageBy length = %d, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// creatureSubtypesOf — supertype filtering
// ---------------------------------------------------------------------------

func TestCreatureSubtypesOf_FiltersSupertypes(t *testing.T) {
	card := &Card{
		Name:  "Legendary Snow Faerie Rogue",
		Owner: 0,
		Types: []string{"legendary", "snow", "creature", "faerie", "rogue"},
		AST:   &gameast.CardAST{Name: "x"},
	}
	got := creatureSubtypesOf(card)
	if _, ok := got["faerie"]; !ok {
		t.Fatal("expected faerie subtype")
	}
	if _, ok := got["rogue"]; !ok {
		t.Fatal("expected rogue subtype")
	}
	for _, banned := range []string{"legendary", "snow", "creature"} {
		if _, ok := got[banned]; ok {
			t.Fatalf("supertype/base-type %q should not appear in subtype set", banned)
		}
	}
}
