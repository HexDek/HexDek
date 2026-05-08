package hexapi

import (
	"math"
	"testing"
)

func TestHexELOGravity_AboveFloor(t *testing.T) {
	// B2 center=1100, floor=600. Rating 1000 is above floor — mild pull.
	g := hexELOGravity(1000, 2)
	if g < 0 || g > 5.0 {
		t.Fatalf("above-floor gravity should be in [0, 5], got %.2f", g)
	}
}

func TestHexELOGravity_BelowFloor_RampsUp(t *testing.T) {
	// B2 floor=600, basement=500. At 550 (halfway), gravity should be ~22.5.
	g := hexELOGravity(550, 2)
	if g < 15 || g > 30 {
		t.Fatalf("mid-ramp gravity should be ~20, got %.2f", g)
	}
	// At basement (500), gravity should be maxed at 40.
	g2 := hexELOGravity(500, 2)
	if math.Abs(g2-40.0) > 0.1 {
		t.Fatalf("basement gravity should be 40, got %.2f", g2)
	}
	// Below basement, still capped at 40.
	g3 := hexELOGravity(300, 2)
	if math.Abs(g3-40.0) > 0.1 {
		t.Fatalf("below-basement gravity should be 40, got %.2f", g3)
	}
}

func TestHexELOGravity_AboveCenter_PullsDown(t *testing.T) {
	// B2 center=1100. Rating 1400 is above center — should pull down.
	g := hexELOGravity(1400, 2)
	if g > 0 || g < -5.0 {
		t.Fatalf("above-center gravity should be in [-5, 0], got %.2f", g)
	}
}

func TestHexELOLossDamper_AboveFloor(t *testing.T) {
	// Above floor, no streak: damper should be 1.0.
	d := hexELOLossDamper(1000, 2, 0)
	if d != 1.0 {
		t.Fatalf("above-floor no-streak damper should be 1.0, got %.2f", d)
	}
}

func TestHexELOLossDamper_AtBasement(t *testing.T) {
	// B2 basement=500. At 500 with long streak: damper should be very low.
	d := hexELOLossDamper(500, 2, 10)
	if d > 0.15 {
		t.Fatalf("basement + long streak damper should be near minimum, got %.2f", d)
	}
}

func TestHexELOLossDamper_StreakOnly(t *testing.T) {
	// Above floor, streak of 8 (5 past threshold): 1.0 * max(0.2, 1-0.75) = 0.25.
	d := hexELOLossDamper(1000, 2, 8)
	if math.Abs(d-0.25) > 0.01 {
		t.Fatalf("streak-8 damper should be ~0.25, got %.2f", d)
	}
}

func TestHexELOStreakBreakBonus(t *testing.T) {
	if b := hexELOStreakBreakBonus(0); b != 0 {
		t.Fatalf("no streak = no bonus, got %.2f", b)
	}
	if b := hexELOStreakBreakBonus(2); b != 0 {
		t.Fatalf("short streak = no bonus, got %.2f", b)
	}
	if b := hexELOStreakBreakBonus(5); math.Abs(b-15.0) > 0.01 {
		t.Fatalf("streak-5 bonus should be 15, got %.2f", b)
	}
	if b := hexELOStreakBreakBonus(20); math.Abs(b-30.0) > 0.01 {
		t.Fatalf("long streak bonus should cap at 30, got %.2f", b)
	}
}

func TestHexELOBracketFloor_AllBrackets(t *testing.T) {
	cases := []struct {
		bracket          int
		wantFloor, wantBasement float64
	}{
		{1, 100, 0},
		{2, 600, 500},
		{3, 1300, 1200},
		{4, 2000, 1900},
		{5, 2800, 2700},
		{0, 600, 500}, // default
	}
	for _, tc := range cases {
		f, b := hexELOBracketFloor(tc.bracket)
		if f != tc.wantFloor || b != tc.wantBasement {
			t.Errorf("B%d: floor=%.0f basement=%.0f, want %.0f/%.0f",
				tc.bracket, f, b, tc.wantFloor, tc.wantBasement)
		}
	}
}
