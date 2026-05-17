package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Aetherdrift speed mechanics — CR §702.178 Max Speed + §702.179 Start Your
// Engines!
// ---------------------------------------------------------------------------

func newSpeedGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(1))
	return NewGameState(2, rng, nil)
}

func newSpeedCard(name string, owner int, keyword string) *Card {
	return &Card{
		Name:  name,
		Owner: owner,
		Types: []string{"artifact"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: keyword},
			},
		},
	}
}

func putSYEPerm(gs *GameState, seatIdx int) *Permanent {
	card := newSpeedCard("Howling Engine", seatIdx, "start your engines!")
	perm := &Permanent{
		Card:       card,
		Controller: seatIdx,
		Owner:      seatIdx,
		Timestamp:  gs.NextTimestamp(),
	}
	gs.Seats[seatIdx].Battlefield = append(gs.Seats[seatIdx].Battlefield, perm)
	return perm
}

// ---------------------------------------------------------------------------
// HasStartYourEngines
// ---------------------------------------------------------------------------

func TestHasStartYourEngines_Detects(t *testing.T) {
	card := newSpeedCard("Howling Engine", 0, "start your engines!")
	if !HasStartYourEngines(card) {
		t.Fatal("HasStartYourEngines should be true for an SYE card")
	}
	// Also accepts the bang-stripped form parsers may emit.
	stripped := newSpeedCard("Slipstream", 0, "start your engines")
	if !HasStartYourEngines(stripped) {
		t.Fatal("HasStartYourEngines should accept bang-stripped keyword form")
	}
}

func TestHasStartYourEngines_Negative(t *testing.T) {
	card := &Card{
		Name:  "Plain Bear",
		Owner: 0,
		Types: []string{"creature"},
		AST: &gameast.CardAST{
			Name:      "Plain Bear",
			Abilities: []gameast.Ability{},
		},
	}
	if HasStartYourEngines(card) {
		t.Fatal("HasStartYourEngines should be false for a card without the keyword")
	}
}

func TestHasStartYourEngines_Nil(t *testing.T) {
	if HasStartYourEngines(nil) {
		t.Fatal("HasStartYourEngines(nil) must return false")
	}
}

// ---------------------------------------------------------------------------
// RunStartYourEnginesUpkeep
// ---------------------------------------------------------------------------

func TestRunStartYourEnginesUpkeep_InitializesSpeedToOne(t *testing.T) {
	gs := newSpeedGame(t)
	putSYEPerm(gs, 0)

	if !RunStartYourEnginesUpkeep(gs, 0) {
		t.Fatal("RunStartYourEnginesUpkeep should report it set speed")
	}
	if got := gs.Seats[0].Flags["speed"]; got != 1 {
		t.Fatalf("seat 0 speed = %d, want 1", got)
	}
}

func TestRunStartYourEnginesUpkeep_NoSYE_NoOp(t *testing.T) {
	gs := newSpeedGame(t)

	if RunStartYourEnginesUpkeep(gs, 0) {
		t.Fatal("RunStartYourEnginesUpkeep should be a no-op without an SYE permanent")
	}
	if got := gs.Seats[0].Flags["speed"]; got != 0 {
		t.Fatalf("seat 0 speed = %d, want 0 (no SYE on battlefield)", got)
	}
}

func TestRunStartYourEnginesUpkeep_DoesNotOverwriteExistingSpeed(t *testing.T) {
	gs := newSpeedGame(t)
	putSYEPerm(gs, 0)
	gs.Seats[0].Flags = map[string]int{"speed": 3}

	if RunStartYourEnginesUpkeep(gs, 0) {
		t.Fatal("RunStartYourEnginesUpkeep must not advance speed when already > 0")
	}
	if got := gs.Seats[0].Flags["speed"]; got != 3 {
		t.Fatalf("seat 0 speed = %d, want 3 (untouched)", got)
	}
}

// ---------------------------------------------------------------------------
// AdvanceSpeedOnCombatDamage
// ---------------------------------------------------------------------------

func TestAdvanceSpeedOnCombatDamage_InitializesAndAdvances(t *testing.T) {
	gs := newSpeedGame(t)
	putSYEPerm(gs, 0)

	got := AdvanceSpeedOnCombatDamage(gs, 0, 1)
	if got != 1 {
		t.Fatalf("first advance returned %d, want 1 (initial bump)", got)
	}
	got = AdvanceSpeedOnCombatDamage(gs, 0, 1)
	if got != 2 {
		t.Fatalf("second advance returned %d, want 2", got)
	}
}

func TestAdvanceSpeedOnCombatDamage_CapsAtMaximum(t *testing.T) {
	gs := newSpeedGame(t)
	putSYEPerm(gs, 0)

	for i := 0; i < 10; i++ {
		AdvanceSpeedOnCombatDamage(gs, 0, 1)
	}
	if got := gs.Seats[0].Flags["speed"]; got != SpeedMaximum {
		t.Fatalf("speed = %d after 10 advances, want %d", got, SpeedMaximum)
	}
}

func TestAdvanceSpeedOnCombatDamage_NoSYE_NoOp(t *testing.T) {
	gs := newSpeedGame(t)

	got := AdvanceSpeedOnCombatDamage(gs, 0, 1)
	if got != 0 {
		t.Fatalf("returned %d without an SYE permanent, want 0", got)
	}
	if gs.Seats[0].Flags["speed"] != 0 {
		t.Fatalf("speed bumped to %d without SYE; should stay 0", gs.Seats[0].Flags["speed"])
	}
}

func TestAdvanceSpeedOnCombatDamage_IgnoresSelfDamage(t *testing.T) {
	gs := newSpeedGame(t)
	putSYEPerm(gs, 0)

	got := AdvanceSpeedOnCombatDamage(gs, 0, 0)
	if got != 0 {
		t.Fatalf("self-damage advance returned %d, want 0", got)
	}
	if gs.Seats[0].Flags["speed"] != 0 {
		t.Fatalf("self-damage bumped speed to %d, want 0", gs.Seats[0].Flags["speed"])
	}
}

// ---------------------------------------------------------------------------
// MaxSpeed
// ---------------------------------------------------------------------------

func TestHasMaxSpeed_PermAndCard(t *testing.T) {
	card := newSpeedCard("Turbo Compressor", 0, "max speed")
	if !HasMaxSpeedCard(card) {
		t.Fatal("HasMaxSpeedCard should detect the keyword on the card")
	}
	perm := &Permanent{Card: card, Controller: 0, Owner: 0}
	if !HasMaxSpeed(perm) {
		t.Fatal("HasMaxSpeed should detect the keyword via the underlying card")
	}
}

func TestHasMaxSpeed_Nil(t *testing.T) {
	if HasMaxSpeed(nil) {
		t.Fatal("HasMaxSpeed(nil) must return false")
	}
}

func TestMaxSpeedActive_Gate(t *testing.T) {
	gs := newSpeedGame(t)

	if MaxSpeedActive(gs, 0) {
		t.Fatal("MaxSpeedActive should be false with no speed")
	}

	gs.Seats[0].Flags = map[string]int{"speed": 3}
	if MaxSpeedActive(gs, 0) {
		t.Fatal("MaxSpeedActive should be false at speed 3")
	}

	gs.Seats[0].Flags["speed"] = SpeedMaximum
	if !MaxSpeedActive(gs, 0) {
		t.Fatalf("MaxSpeedActive should be true at speed %d", SpeedMaximum)
	}
}

// ---------------------------------------------------------------------------
// Integration: combat damage path increments speed via the combat hook.
// ---------------------------------------------------------------------------

func TestSpeed_CombatDamageHook_BumpsSpeed(t *testing.T) {
	gs := newSpeedGame(t)
	putSYEPerm(gs, 0)

	// Build an attacker controlled by seat 0.
	attackerCard := &Card{
		Name:          "Speed Demon",
		Owner:         0,
		Types:         []string{"creature"},
		BasePower:     3,
		BaseToughness: 3,
		AST:           &gameast.CardAST{Name: "Speed Demon"},
	}
	attacker := &Permanent{
		Card:       attackerCard,
		Controller: 0,
		Owner:      0,
		Timestamp:  gs.NextTimestamp(),
	}
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, attacker)

	applyCombatDamageToPlayer(gs, attacker, 3, 1)

	if got := gs.Seats[0].Flags["speed"]; got != 1 {
		t.Fatalf("speed after first hit = %d, want 1", got)
	}
	applyCombatDamageToPlayer(gs, attacker, 3, 1)
	applyCombatDamageToPlayer(gs, attacker, 3, 1)
	applyCombatDamageToPlayer(gs, attacker, 3, 1)
	applyCombatDamageToPlayer(gs, attacker, 3, 1)
	if got := gs.Seats[0].Flags["speed"]; got != SpeedMaximum {
		t.Fatalf("speed after many hits = %d, want %d", got, SpeedMaximum)
	}
}
