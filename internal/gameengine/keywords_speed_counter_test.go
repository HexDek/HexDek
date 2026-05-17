package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// Tests for the Aetherdrift Speed counter system (CR §702.178 / §702.179).
// Required coverage from the round-24 spec:
//   (a) start = 0
//   (b) combat damage to player advances by 1
//   (c) cap at 4 (no overflow)
//   (d) once-per-turn limit (multiple damage events same turn = +1 max)
//   (e) MaxSpeedActive at speed=4

func newSpeedGame(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(2, rand.New(rand.NewSource(11)), nil)
}

func newSpeedAttacker(seat int, name string, power int) *Permanent {
	c := &Card{
		Name:          name,
		Owner:         seat,
		Types:         []string{"creature"},
		BasePower:     power,
		BaseToughness: 2,
	}
	return &Permanent{
		Card:       c,
		Controller: seat,
		Owner:      seat,
		Flags:      map[string]int{},
	}
}

// (a) Start = 0.
func TestSpeed_StartsAtZero(t *testing.T) {
	gs := newSpeedGame(t)
	for i := range gs.Seats {
		if got := SpeedOf(gs, i); got != 0 {
			t.Fatalf("seat %d starting speed = %d, want 0", i, got)
		}
		if MaxSpeedActive(gs, i) {
			t.Fatalf("seat %d should not be at max speed on game start", i)
		}
	}
}

// (b) Combat damage to player advances by 1.
func TestSpeed_CombatDamageToPlayerAdvances(t *testing.T) {
	gs := newSpeedGame(t)
	attacker := newSpeedAttacker(0, "Ghalta's Stampede", 3)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, attacker)

	applyCombatDamageToPlayer(gs, attacker, 3, 1)

	if got := SpeedOf(gs, 0); got != 1 {
		t.Fatalf("dealer speed after combat damage = %d, want 1", got)
	}
	if got := SpeedOf(gs, 1); got != 0 {
		t.Fatalf("victim speed = %d, want 0 (only dealer advances)", got)
	}
	if !gs.Seats[0].Turn.SpeedAdvancedThisTurn {
		t.Fatal("SpeedAdvancedThisTurn flag should be set on dealer seat")
	}
}

// (c) Cap at 4 (no overflow).
func TestSpeed_CapsAtFour(t *testing.T) {
	gs := newSpeedGame(t)
	gs.Seats[0].Speed = 3
	// Tick once → 4.
	if !AdvanceSpeed(gs, 0) {
		t.Fatal("AdvanceSpeed from 3 should return true")
	}
	if got := SpeedOf(gs, 0); got != MaxSpeedCap {
		t.Fatalf("speed after advance from 3 = %d, want %d", got, MaxSpeedCap)
	}
	// New turn — reset gate, attempt another advance.
	ResetSpeedAdvancedFlag(gs, 0)
	if AdvanceSpeed(gs, 0) {
		t.Fatal("AdvanceSpeed at cap should return false")
	}
	if got := SpeedOf(gs, 0); got != MaxSpeedCap {
		t.Fatalf("speed must stay at cap, got %d", got)
	}
	// SetSpeed cannot exceed cap either.
	SetSpeed(gs, 0, 99)
	if got := SpeedOf(gs, 0); got != MaxSpeedCap {
		t.Fatalf("SetSpeed(99) should clamp to %d, got %d", MaxSpeedCap, got)
	}
	SetSpeed(gs, 0, -5)
	if got := SpeedOf(gs, 0); got != 0 {
		t.Fatalf("SetSpeed(-5) should clamp to 0, got %d", got)
	}
}

// (d) Once-per-turn limit — multiple damage events same turn = +1 max.
func TestSpeed_OncePerTurn(t *testing.T) {
	gs := newSpeedGame(t)
	a1 := newSpeedAttacker(0, "First Striker", 2)
	a2 := newSpeedAttacker(0, "Trampler", 4)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, a1, a2)

	// Two combat hits in the same turn.
	applyCombatDamageToPlayer(gs, a1, 2, 1)
	applyCombatDamageToPlayer(gs, a2, 4, 1)

	if got := SpeedOf(gs, 0); got != 1 {
		t.Fatalf("speed after 2 combat hits same turn = %d, want 1 (once-per-turn)", got)
	}

	// Simulate a new turn → flag clears → another advance is allowed.
	ResetSpeedAdvancedFlag(gs, 0)
	applyCombatDamageToPlayer(gs, a1, 2, 1)
	if got := SpeedOf(gs, 0); got != 2 {
		t.Fatalf("speed after new-turn advance = %d, want 2", got)
	}
}

// (e) MaxSpeedActive at speed=4.
func TestSpeed_MaxSpeedActivePredicate(t *testing.T) {
	gs := newSpeedGame(t)
	for s := 0; s < MaxSpeedCap; s++ {
		gs.Seats[0].Speed = s
		if MaxSpeedActive(gs, 0) {
			t.Fatalf("MaxSpeedActive at speed=%d should be false", s)
		}
	}
	gs.Seats[0].Speed = MaxSpeedCap
	if !MaxSpeedActive(gs, 0) {
		t.Fatalf("MaxSpeedActive at speed=%d should be true", MaxSpeedCap)
	}

	// HasMaxSpeed wiring: rider only fires when controller is at max speed.
	perm := newSpeedAttacker(0, "Speed Demon", 3)
	// No keyword → never has max-speed rider active.
	if HasMaxSpeedKeyword(perm) {
		t.Fatal("perm without keyword should not report HasMaxSpeedKeyword")
	}
	if HasMaxSpeed(gs, perm) {
		t.Fatal("perm without keyword should not have an active max-speed rider")
	}
	// Add keyword to the AST and re-check.
	perm.Card.AST = makeMaxSpeedAST("Speed Demon")
	if !HasMaxSpeedKeyword(perm) {
		t.Fatal("perm with max-speed keyword should report HasMaxSpeedKeyword")
	}
	if !HasMaxSpeed(gs, perm) {
		t.Fatal("perm with max-speed keyword + controller at MaxSpeedCap should report HasMaxSpeed=true")
	}
	// Drop controller below max → rider goes inactive.
	gs.Seats[0].Speed = MaxSpeedCap - 1
	if HasMaxSpeed(gs, perm) {
		t.Fatal("HasMaxSpeed should be false when controller drops below MaxSpeedCap")
	}
}

// Bonus: spell/ability damage path also advances speed.
func TestSpeed_NoncombatDamageAdvances(t *testing.T) {
	gs := newSpeedGame(t)
	src := newSpeedAttacker(0, "Lightning Source", 0)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, src)

	applyDamage(gs, src, Target{Kind: TargetKindSeat, Seat: 1}, 2)

	if got := SpeedOf(gs, 0); got != 1 {
		t.Fatalf("speed after spell-style damage = %d, want 1", got)
	}
}

// Bonus: AdvanceSpeed is safe with invalid seat / nil game.
func TestSpeed_AdvanceNilSafe(t *testing.T) {
	if AdvanceSpeed(nil, 0) {
		t.Fatal("AdvanceSpeed(nil) should return false")
	}
	gs := newSpeedGame(t)
	if AdvanceSpeed(gs, -1) {
		t.Fatal("AdvanceSpeed seat=-1 should return false")
	}
	if AdvanceSpeed(gs, 999) {
		t.Fatal("AdvanceSpeed seat=999 should return false")
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// makeMaxSpeedAST returns a CardAST carrying just the "max speed"
// keyword, enough for perm.HasKeyword("max speed") to return true.
func makeMaxSpeedAST(name string) *gameast.CardAST {
	return &gameast.CardAST{
		Name: name,
		Abilities: []gameast.Ability{
			&gameast.Keyword{Name: "max speed"},
		},
	}
}
