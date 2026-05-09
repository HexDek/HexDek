package hat

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// First-strike-aware blocking + attack tests.
//
// Convention: seat 0 is the attacker, seat 1 is the blocker. We invoke
// AssignBlockers from the blocker's perspective (seatIdx=1) and inspect
// the assignment for the supplied attacker.

// ---------------------------------------------------------------------
// Blocking decisions
// ---------------------------------------------------------------------

// A first-strike attacker faces a "would-survive-vs-vanilla" blocker
// at full life. The hat's existing survivor math (incomingToBlocker,
// killsInFirstStrike) already handles this — we lock the behavior in.
func TestFirstStrike_NoFalseSurvivorAtFullLife(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 30
	h := NewYggdrasilHat(nil, 0)

	// 4/3 first-striker; 1/4 blocker. Blocker dies in FS step (4 ≥ 4).
	// At 30 life with only 4 incoming, the hat shouldn't burn the
	// blocker on a no-trade swap.
	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("FS", []string{"creature"}, 3, nil), 4, 3)
	addKeyword(atk, "first strike")
	_ = newTestPermanent(gs.Seats[1], newTestCardMinimal("Toad", []string{"creature"}, 2, nil), 1, 4)

	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	if len(out[atk]) > 0 {
		t.Fatalf("hat shouldn't block 4/3 FS attacker with a 1/4 at safe life (no trade); got %d blockers", len(out[atk]))
	}
}

// At lethal we still chump even against an FS attacker — losing a
// creature beats losing the game.
func TestFirstStrike_ChumpAtLethal(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 4
	h := NewYggdrasilHat(nil, 0)

	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("FS", []string{"creature"}, 3, nil), 4, 3)
	addKeyword(atk, "first strike")
	_ = newTestPermanent(gs.Seats[1], newTestCardMinimal("Toad", []string{"creature"}, 2, nil), 1, 4)

	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	if len(out[atk]) == 0 {
		t.Fatalf("at lethal life, hat must chump even vs FS; got 0 blockers")
	}
}

// Double-strike attacker vs vanilla "would-survive" blocker: the FS+regular
// damage stack kills the blocker (3+3=6 dmg through 4 toughness). The
// existing DS math already covers this — verify it.
func TestFirstStrike_DoubleStrikeRejectsVanillaSurvivor(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 25
	h := NewYggdrasilHat(nil, 0)

	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("DS", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(atk, "double strike")
	_ = newTestPermanent(gs.Seats[1], newTestCardMinimal("Wall", []string{"creature"}, 3, nil), 2, 4)

	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	if len(out[atk]) > 0 {
		t.Fatalf("hat shouldn't block 3/3 DS attacker with a 2/4 (DS kills blocker); got %d blockers", len(out[atk]))
	}
}

// Deathtouch + first-strike blocker is a real trade — kills the
// attacker before taking damage. Block at lethal even though the
// blocker dies (it can't survive a 5-power FS counter-hit, but it
// trades up).
func TestFirstStrike_DTPlusFirstStrikeBlockerTradesAtLethal(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 4
	h := NewYggdrasilHat(nil, 0)

	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("Big Beater", []string{"creature"}, 5, nil), 5, 5)
	dt := newTestPermanent(gs.Seats[1], newTestCardMinimal("Assassin", []string{"creature"}, 2, nil), 1, 1)
	addKeyword(dt, "first strike")
	addKeyword(dt, "deathtouch")

	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	if len(out[atk]) != 1 || out[atk][0] != dt {
		t.Fatalf("hat should pick the DT+FS assassin to trade up; got %v", out[atk])
	}
}

// Vanilla-deathtouch blocker (no FS) vs first-strike attacker is the
// "no upside" trap — the DT blocker dies in FS step before its DT
// fires. The hat must NOT pick it as a trade-up. At lethal we'd still
// chump-block, but with a different blocker if available — this test
// uses ONLY the DT blocker as legal, so the chump branch fires it
// anyway. Use a non-lethal scenario to verify the DT-trader logic
// rejects it without falling back to chump.
func TestFirstStrike_VanillaDTNotPickedAsTradeUpVsFirstStrike(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 30
	h := NewYggdrasilHat(nil, 0)

	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("FS Beater", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(atk, "first strike")
	dt := newTestPermanent(gs.Seats[1], newTestCardMinimal("Snake", []string{"creature"}, 2, nil), 1, 1)
	addKeyword(dt, "deathtouch")

	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	if len(out[atk]) > 0 {
		t.Fatalf("vanilla DT blocker should NOT be picked as a trade-up vs FS attacker (dies in FS step before DT fires); got %v", out[atk])
	}
}

// Indestructible + deathtouch blocker DOES trade up vs FS attacker —
// indestructible saves it from the FS hit, then DT fires in the
// regular step.
func TestFirstStrike_IndestructibleDTBlocksFirstStrikeAttacker(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 4
	h := NewYggdrasilHat(nil, 0)

	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("FS", []string{"creature"}, 3, nil), 5, 3)
	addKeyword(atk, "first strike")
	dt := newTestPermanent(gs.Seats[1], newTestCardMinimal("DT God", []string{"creature"}, 4, nil), 1, 1)
	addKeyword(dt, "deathtouch")
	addKeyword(dt, "indestructible")

	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	if len(out[atk]) != 1 || out[atk][0] != dt {
		t.Fatalf("indestructible+DT blocker should trade up against FS attacker; got %v", out[atk])
	}
}

// Trample + first-strike attacker at lethal where the chump can't
// absorb enough to save us — preserve the chump.
func TestFirstStrike_TrampleChumpSkippedWhenItDoesntSave(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 3
	h := NewYggdrasilHat(nil, 0)

	// 5/5 FS+trample; 1/2 chump. Chump absorbs 2, 3 leaks, still kills
	// us. Don't burn the chump.
	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("FS Trampler", []string{"creature"}, 5, nil), 5, 5)
	addKeyword(atk, "first strike")
	addKeyword(atk, "trample")
	chump := newTestPermanent(gs.Seats[1], newTestCardMinimal("Squire", []string{"creature"}, 1, nil), 1, 2)

	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	for _, b := range out[atk] {
		if b == chump {
			t.Fatalf("hat shouldn't burn a chump that can't save us from FS+trample lethal; got %v", out[atk])
		}
	}
}

// Trample + first-strike attacker at lethal where the chump's
// toughness fully absorbs the swing — block it.
func TestFirstStrike_TrampleChumpFullAbsorbBlocks(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 3
	h := NewYggdrasilHat(nil, 0)

	// 4/5 FS+trample; 1/4 chump. Chump absorbs 4, 0 leak — saves us.
	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("FS Trampler", []string{"creature"}, 5, nil), 4, 5)
	addKeyword(atk, "first strike")
	addKeyword(atk, "trample")
	chump := newTestPermanent(gs.Seats[1], newTestCardMinimal("Big Wall", []string{"creature"}, 4, nil), 1, 4)
	_ = chump

	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	if len(out[atk]) == 0 {
		t.Fatalf("hat should block when chump fully absorbs FS+trample damage; got 0 blockers")
	}
}

// ---------------------------------------------------------------------
// Attack-side: first-strike attackers get a value bonus
// ---------------------------------------------------------------------

// First-strike attackers should see a positive value bump in the
// attacker-selection weight scheme (sandwiched between vanilla and
// double strike).
func TestFirstStrike_AttackValueBonus(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)

	vanilla := newTestPermanent(gs.Seats[0], newTestCardMinimal("Bear", []string{"creature"}, 2, nil), 2, 2)
	fs := newTestPermanent(gs.Seats[0], newTestCardMinimal("Knight", []string{"creature"}, 2, nil), 2, 2)
	addKeyword(fs, "first strike")
	ds := newTestPermanent(gs.Seats[0], newTestCardMinimal("Champion", []string{"creature"}, 2, nil), 2, 2)
	addKeyword(ds, "double strike")

	// Smoke-test that ChooseAttackers runs cleanly with FS attackers.
	// (We can't directly inspect val without exposing internals; the
	// behavior change is demonstrated by the lethal-detection +
	// stance-threshold downstream choices, which already have separate
	// coverage. This test exists primarily to lock in that the FS
	// branch doesn't regress build/runtime.)
	out := h.ChooseAttackers(gs, 0, []*gameengine.Permanent{vanilla, fs, ds})
	_ = out
}
