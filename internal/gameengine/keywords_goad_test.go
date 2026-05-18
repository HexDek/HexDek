package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Goad tests — CR §701.39
// ---------------------------------------------------------------------------

func newGoadGame(t *testing.T, seats int) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(37))
	gs := NewGameState(seats, rng, nil)
	gs.Active = 0
	gs.Step = "precombat_main"
	gs.Turn = 10
	return gs
}

func newGoadCreature(name string, owner int) *Card {
	return &Card{
		Name:          name,
		Owner:         owner,
		Types:         []string{"creature"},
		BasePower:     3,
		BaseToughness: 3,
		AST:           &gameast.CardAST{Name: name},
	}
}

func putGoadBattlefield(gs *GameState, seat int, card *Card) *Permanent {
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
// (a) GoadCreature stamps flags
// ---------------------------------------------------------------------------

func TestGoadCreature_StampsFlags(t *testing.T) {
	gs := newGoadGame(t, 4)
	// Goader = seat 0 (currently active). Target on seat 2.
	target := putGoadBattlefield(gs, 2, newGoadCreature("Wurm", 2))

	GoadCreature(gs, 0, target)

	if target.Flags == nil {
		t.Fatal("GoadCreature should initialise Flags")
	}
	if got := target.Flags["goaded_by_seat"]; got != 0 {
		t.Fatalf("goaded_by_seat = %d, want 0", got)
	}
	// In a 4-seat game with goader=active, seatsBefore(0)=4, so
	// goaded_until_turn = 10 + 4 = 14.
	if got := target.Flags["goaded_until_turn"]; got != 14 {
		t.Fatalf("goaded_until_turn = %d, want 14 (Turn 10 + seatsBefore 4)", got)
	}
	if got := target.Flags["goaded"]; got != 1 {
		t.Fatalf("legacy goaded flag = %d, want 1", got)
	}
}

func TestGoadCreature_NilSafe(t *testing.T) {
	gs := newGoadGame(t, 4)
	GoadCreature(nil, 0, nil)                                                 // doesn't panic
	GoadCreature(gs, 0, nil)                                                  // doesn't panic
	GoadCreature(gs, 99, putGoadBattlefield(gs, 0, newGoadCreature("X", 0))) // bad seat → no-op
}

func TestGoadCreature_Regoad_ReplacesGoader(t *testing.T) {
	gs := newGoadGame(t, 4)
	target := putGoadBattlefield(gs, 2, newGoadCreature("Wurm", 2))
	GoadCreature(gs, 0, target)
	GoadCreature(gs, 1, target) // re-goad by a different seat
	if target.Flags["goaded_by_seat"] != 1 {
		t.Fatalf("re-goad should replace goader to seat 1, got %d", target.Flags["goaded_by_seat"])
	}
}

// ---------------------------------------------------------------------------
// seatsBefore semantics
// ---------------------------------------------------------------------------

func TestSeatsBefore_4SeatRotation(t *testing.T) {
	gs := newGoadGame(t, 4)
	gs.Active = 1
	cases := []struct {
		seat int
		want int
	}{
		{2, 1}, // next turn
		{3, 2},
		{0, 3},
		{1, 4}, // current active — full lap until next turn
	}
	for _, c := range cases {
		if got := seatsBefore(gs, c.seat); got != c.want {
			t.Fatalf("seatsBefore(Active=1, seat=%d) = %d, want %d", c.seat, got, c.want)
		}
	}
}

func TestSeatsBefore_NilSafe(t *testing.T) {
	if got := seatsBefore(nil, 0); got != 0 {
		t.Fatalf("seatsBefore(nil) = %d, want 0", got)
	}
	gs := newGoadGame(t, 4)
	if got := seatsBefore(gs, 99); got != 0 {
		t.Fatalf("seatsBefore(out-of-range) = %d, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// (b) IsGoaded true until source's next turn
// ---------------------------------------------------------------------------

func TestIsGoaded_TrueUntilGoadersNextTurn(t *testing.T) {
	gs := newGoadGame(t, 4)
	// Goader = seat 0, active = 0, turn 10. seatsBefore(0) = 4.
	// goaded_until_turn = 14.
	target := putGoadBattlefield(gs, 2, newGoadCreature("Wurm", 2))
	GoadCreature(gs, 0, target)

	// Walk the table: turns 10..13 should all read goaded; turn 14
	// (goader's next turn) reads NOT goaded.
	for _, turn := range []int{10, 11, 12, 13} {
		if !IsGoaded(target, turn) {
			t.Fatalf("IsGoaded should be true on turn %d", turn)
		}
	}
	if IsGoaded(target, 14) {
		t.Fatal("IsGoaded should be false on turn 14 (goader's next turn — goad expires)")
	}
	if IsGoaded(target, 15) {
		t.Fatal("IsGoaded should remain false beyond expiry")
	}
}

func TestIsGoadedBy_KeysOnGoaderSeat(t *testing.T) {
	gs := newGoadGame(t, 4)
	target := putGoadBattlefield(gs, 2, newGoadCreature("Wurm", 2))
	GoadCreature(gs, 0, target)
	if !IsGoadedBy(target, 0, gs.Turn) {
		t.Fatal("IsGoadedBy(0) should be true after goader=0 stamps")
	}
	if IsGoadedBy(target, 1, gs.Turn) {
		t.Fatal("IsGoadedBy(1) should be false — goader was seat 0")
	}
}

func TestGoadedBySeat_ReturnsGoaderAndStillActive(t *testing.T) {
	gs := newGoadGame(t, 4)
	target := putGoadBattlefield(gs, 2, newGoadCreature("Wurm", 2))
	GoadCreature(gs, 3, target)
	goader, active := GoadedBySeat(target, gs.Turn)
	if !active {
		t.Fatal("GoadedBySeat should report active goad")
	}
	if goader != 3 {
		t.Fatalf("goader = %d, want 3", goader)
	}
	// After expiry: returns -1, false.
	if goader, active := GoadedBySeat(target, 14); active || goader != -1 {
		t.Fatalf("after expiry: got (%d, %v), want (-1, false)", goader, active)
	}
}

// ---------------------------------------------------------------------------
// (c) MustAttackIfAble forces attack declaration (predicate truth)
// ---------------------------------------------------------------------------

func TestMustAttackIfAble_TrueOnGoadedCreaturesTurn(t *testing.T) {
	gs := newGoadGame(t, 4)
	target := putGoadBattlefield(gs, 2, newGoadCreature("Wurm", 2))
	GoadCreature(gs, 0, target)

	// On seat 2's turn, the goaded creature MUST attack.
	gs.Active = 2
	gs.Turn = 12
	if !MustAttackIfAble(gs, target) {
		t.Fatal("MustAttackIfAble should be true on the goaded creature's controller's turn")
	}
}

func TestMustAttackIfAble_FalseOnOtherSeatsTurn(t *testing.T) {
	gs := newGoadGame(t, 4)
	target := putGoadBattlefield(gs, 2, newGoadCreature("Wurm", 2))
	GoadCreature(gs, 0, target)

	// On seat 1's turn, seat 2's goaded creature has no must-attack
	// obligation (it isn't seat 2's combat).
	gs.Active = 1
	gs.Turn = 11
	if MustAttackIfAble(gs, target) {
		t.Fatal("MustAttackIfAble should be false outside the goaded creature's controller's turn")
	}
}

func TestMustAttackIfAble_FalseAfterExpiry(t *testing.T) {
	gs := newGoadGame(t, 4)
	target := putGoadBattlefield(gs, 2, newGoadCreature("Wurm", 2))
	GoadCreature(gs, 0, target)
	// Fast-forward past the expiry; seat 2's next turn after seat 0's
	// is turn 16 in this game's clock, but seat 2's next attack
	// opportunity after goad expires is turn 14+2=16 — anything past
	// 13 has IsGoaded false.
	gs.Active = 2
	gs.Turn = 16
	if MustAttackIfAble(gs, target) {
		t.Fatal("MustAttackIfAble should be false once goad has expired")
	}
}

func TestMustAttackIfAble_FalseForNonGoaded(t *testing.T) {
	gs := newGoadGame(t, 4)
	plain := putGoadBattlefield(gs, 0, newGoadCreature("Plain Wurm", 0))
	if MustAttackIfAble(gs, plain) {
		t.Fatal("MustAttackIfAble should be false for a non-goaded creature")
	}
}

// ---------------------------------------------------------------------------
// (d) Can't attack the goader (must attack someone else)
// ---------------------------------------------------------------------------

func TestCannotAttackGoader_BlocksTheGoader(t *testing.T) {
	gs := newGoadGame(t, 4)
	target := putGoadBattlefield(gs, 2, newGoadCreature("Wurm", 2))
	GoadCreature(gs, 0, target) // goader = 0

	if !CannotAttackGoader(gs, target, 0) {
		t.Fatal("CannotAttackGoader should be true for the goader (seat 0)")
	}
	for _, otherDefender := range []int{1, 3} {
		if CannotAttackGoader(gs, target, otherDefender) {
			t.Fatalf("CannotAttackGoader should be false for non-goader defender %d", otherDefender)
		}
	}
}

func TestCannotAttackGoader_FalseAfterExpiry(t *testing.T) {
	gs := newGoadGame(t, 4)
	target := putGoadBattlefield(gs, 2, newGoadCreature("Wurm", 2))
	GoadCreature(gs, 0, target)
	gs.Turn = 14 // goad expired
	if CannotAttackGoader(gs, target, 0) {
		t.Fatal("CannotAttackGoader should be false once goad has expired")
	}
}

func TestCannotAttackGoader_FalseForNonGoaded(t *testing.T) {
	gs := newGoadGame(t, 4)
	plain := putGoadBattlefield(gs, 2, newGoadCreature("Plain Wurm", 2))
	if CannotAttackGoader(gs, plain, 0) {
		t.Fatal("CannotAttackGoader should be false for a non-goaded creature")
	}
}

// ---------------------------------------------------------------------------
// (e) Goad expires correctly at goader's next turn start
// ---------------------------------------------------------------------------

func TestExpireGoadAtCleanup_ClearsAtGoadersNextTurn(t *testing.T) {
	gs := newGoadGame(t, 4)
	target := putGoadBattlefield(gs, 2, newGoadCreature("Wurm", 2))
	GoadCreature(gs, 0, target) // until_turn = 14

	// Mid-window: should NOT clear.
	gs.Turn = 12
	if ExpireGoadAtCleanup(gs, target) {
		t.Fatal("ExpireGoadAtCleanup mid-window should not clear")
	}
	if target.Flags["goaded_by_seat"] != 0 || target.Flags["goaded_until_turn"] != 14 {
		t.Fatal("mid-window cleanup should preserve flags")
	}

	// At the goader's next turn start (turn 14): should clear.
	gs.Turn = 14
	if !ExpireGoadAtCleanup(gs, target) {
		t.Fatal("ExpireGoadAtCleanup should clear at goader's next turn start")
	}
	if _, ok := target.Flags["goaded_by_seat"]; ok {
		t.Fatal("goaded_by_seat should be deleted after expire")
	}
	if _, ok := target.Flags["goaded_until_turn"]; ok {
		t.Fatal("goaded_until_turn should be deleted after expire")
	}
	if _, ok := target.Flags["goaded"]; ok {
		t.Fatal("legacy goaded flag should be deleted after expire")
	}
	// IsGoaded reads false too.
	if IsGoaded(target, gs.Turn) {
		t.Fatal("IsGoaded should be false after explicit cleanup")
	}
}

func TestExpireGoadAtCleanup_NoOpForNonGoaded(t *testing.T) {
	gs := newGoadGame(t, 4)
	plain := putGoadBattlefield(gs, 0, newGoadCreature("Plain", 0))
	if ExpireGoadAtCleanup(gs, plain) {
		t.Fatal("ExpireGoadAtCleanup should return false for a non-goaded perm")
	}
}

func TestExpireAllGoadsAtCleanup_SweepsEveryBattlefield(t *testing.T) {
	gs := newGoadGame(t, 4)
	a := putGoadBattlefield(gs, 2, newGoadCreature("A", 2))
	b := putGoadBattlefield(gs, 3, newGoadCreature("B", 3))
	c := putGoadBattlefield(gs, 0, newGoadCreature("C", 0))
	GoadCreature(gs, 0, a) // expiry 14
	GoadCreature(gs, 0, b) // expiry 14
	// c is unfiltered (no goad)

	gs.Turn = 14
	cleared := ExpireAllGoadsAtCleanup(gs)
	if cleared != 2 {
		t.Fatalf("ExpireAllGoadsAtCleanup cleared %d, want 2", cleared)
	}
	if IsGoaded(a, gs.Turn) || IsGoaded(b, gs.Turn) {
		t.Fatal("both goaded creatures should be cleared after sweep")
	}
	if _, ok := c.Flags["goaded"]; ok {
		t.Fatal("non-goaded perm should be untouched")
	}
}

// ---------------------------------------------------------------------------
// Combined "must attack someone other than goader" worked example
// ---------------------------------------------------------------------------

func TestGoad_AttackTargetDecision_4SeatGame(t *testing.T) {
	// Worked example: seat 0 goads seat 2's creature. On seat 2's
	// turn, the goaded creature must attack (MustAttackIfAble=true)
	// and must attack a seat other than 0 (CannotAttackGoader(0)).
	// Seats 1 and 3 are the only legal defenders.
	gs := newGoadGame(t, 4)
	target := putGoadBattlefield(gs, 2, newGoadCreature("Wurm", 2))
	GoadCreature(gs, 0, target)

	gs.Active = 2
	gs.Turn = 12

	if !MustAttackIfAble(gs, target) {
		t.Fatal("expected MustAttackIfAble=true on goaded creature's turn")
	}
	defenderPool := []int{0, 1, 3} // opponents of seat 2
	var legal []int
	for _, def := range defenderPool {
		if CannotAttackGoader(gs, target, def) {
			continue
		}
		legal = append(legal, def)
	}
	want := []int{1, 3}
	if len(legal) != len(want) {
		t.Fatalf("legal defenders = %v, want %v", legal, want)
	}
	for i, w := range want {
		if legal[i] != w {
			t.Fatalf("legal[%d] = %d, want %d", i, legal[i], w)
		}
	}
}
