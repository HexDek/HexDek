package hat

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// Multi-blocker gang-block tests for AssignBlockers.

// Gang-block at lethal: 6/6 attacker, three 2/3 blockers. None survive
// the 1v1 (3 < 6 toughness math), but together their combined 6 power
// kills the 6-toughness attacker. The hat should assign all three.
func TestGangBlock_ThreeSmallBlockersKillBigAttacker(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 5 // we'd die to the 6/6 swing
	h := NewYggdrasilHat(nil, 0)

	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("Big", []string{"creature"}, 5, nil), 6, 6)
	b1 := newTestPermanent(gs.Seats[1], newTestCardMinimal("B1", []string{"creature"}, 2, nil), 2, 3)
	b2 := newTestPermanent(gs.Seats[1], newTestCardMinimal("B2", []string{"creature"}, 2, nil), 2, 3)
	b3 := newTestPermanent(gs.Seats[1], newTestCardMinimal("B3", []string{"creature"}, 2, nil), 2, 3)

	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	if len(out[atk]) < 2 {
		t.Fatalf("hat should gang-block 6/6 with multiple 2/3s; got %d blockers", len(out[atk]))
	}
	totalPow := 0
	for _, b := range out[atk] {
		totalPow += gs.PowerOf(b)
	}
	if totalPow < 6 {
		t.Fatalf("gang-block should sum to ≥6 power to kill 6/6; got %d", totalPow)
	}
	// Sanity: at least two of the three 2/3s should be in the assignment.
	count := 0
	for _, b := range out[atk] {
		if b == b1 || b == b2 || b == b3 {
			count++
		}
	}
	if count < 2 {
		t.Fatalf("gang should pull from the 2/3 pool; got %d/3 of them", count)
	}
}

// At plenty of life, the hat should NOT gang-block — the 6/6 swing
// is annoying but not lethal, and 2-3 creatures lost for one is a
// bad trade outside lethal.
func TestGangBlock_NoGangAtFullLife(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 30 // safe
	h := NewYggdrasilHat(nil, 0)

	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("Big", []string{"creature"}, 5, nil), 6, 6)
	_ = newTestPermanent(gs.Seats[1], newTestCardMinimal("B1", []string{"creature"}, 2, nil), 2, 3)
	_ = newTestPermanent(gs.Seats[1], newTestCardMinimal("B2", []string{"creature"}, 2, nil), 2, 3)
	_ = newTestPermanent(gs.Seats[1], newTestCardMinimal("B3", []string{"creature"}, 2, nil), 2, 3)

	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	// At 30 life vs a 6/6, the favorable-trade or chump branch may
	// pick a single blocker, but multi-blocker gang should NOT fire.
	if len(out[atk]) > 1 {
		t.Fatalf("hat should NOT gang-block at safe life; got %d blockers", len(out[atk]))
	}
}

// Indestructible attacker: gang-block can't kill it, so don't burn
// the bodies.
func TestGangBlock_IndestructibleAttackerSkipsGang(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 5
	h := NewYggdrasilHat(nil, 0)

	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("God", []string{"creature"}, 5, nil), 6, 6)
	addKeyword(atk, "indestructible")
	b1 := newTestPermanent(gs.Seats[1], newTestCardMinimal("B1", []string{"creature"}, 2, nil), 2, 3)
	b2 := newTestPermanent(gs.Seats[1], newTestCardMinimal("B2", []string{"creature"}, 2, nil), 2, 3)
	b3 := newTestPermanent(gs.Seats[1], newTestCardMinimal("B3", []string{"creature"}, 2, nil), 2, 3)
	_, _, _ = b1, b2, b3

	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	// At lethal we still must block (chump branch fires for damage
	// prevention), but we shouldn't burn 2-3 creatures gang-blocking
	// when the attacker can't be killed.
	if len(out[atk]) > 1 {
		t.Fatalf("hat should not gang-block an indestructible attacker; got %d blockers", len(out[atk]))
	}
}

// Lifelink trigger: a sub-4-power attacker still counts as "dangerous"
// for the gang-block gate when it has lifelink. Two 2/2 blockers gang
// up to kill the 3/3 lifelinker.
func TestGangBlock_LifelinkTriggersGang(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 4 // 3 lifelink dmg = 3 lost + 3 gained → ~lethal
	h := NewYggdrasilHat(nil, 0)

	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("Vampire", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(atk, "lifelink")
	b1 := newTestPermanent(gs.Seats[1], newTestCardMinimal("B1", []string{"creature"}, 2, nil), 2, 2)
	b2 := newTestPermanent(gs.Seats[1], newTestCardMinimal("B2", []string{"creature"}, 2, nil), 2, 2)
	_, _ = b1, b2

	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	if len(out[atk]) < 2 {
		t.Fatalf("hat should gang-block lifelink threat at lethal; got %d blockers", len(out[atk]))
	}
	totalPow := 0
	for _, b := range out[atk] {
		totalPow += gs.PowerOf(b)
	}
	if totalPow < 3 {
		t.Fatalf("gang should sum to ≥3 to kill the 3/3 lifelink; got %d", totalPow)
	}
}

// Gang doesn't replace a single blocker that ALREADY kills the
// attacker (e.g. a deathtouch trader handles it solo).
func TestGangBlock_DeathtouchSoloBlockerNotReplacedByGang(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 5
	h := NewYggdrasilHat(nil, 0)

	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("Big", []string{"creature"}, 5, nil), 6, 6)
	dt := newTestPermanent(gs.Seats[1], newTestCardMinimal("Snake", []string{"creature"}, 2, nil), 1, 1)
	addKeyword(dt, "deathtouch")
	b1 := newTestPermanent(gs.Seats[1], newTestCardMinimal("B1", []string{"creature"}, 2, nil), 2, 3)
	b2 := newTestPermanent(gs.Seats[1], newTestCardMinimal("B2", []string{"creature"}, 2, nil), 2, 3)
	_, _ = b1, b2

	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	// Single DT blocker handles a non-FS 6/6 just fine — the gang
	// branch should leave it alone.
	if len(out[atk]) != 1 || out[atk][0] != dt {
		t.Fatalf("hat should solo-block with DT (kills attacker on its own); got %v", out[atk])
	}
}
