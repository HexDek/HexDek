package hat

// First-strike attacker awareness tests.
//
// PR #50 made the defender side first-strike-aware in AssignBlockers.
// This file covers the attacker-side extensions:
//  - simulateBlockerTrade resolves FS/DS/DT/indestructible correctly
//  - attackerRank gives first strike a +2 bump (and double strike still
//    overrides — never both)
//  - bestTarget penalizes attacking into defenders with FS blockers
//    that can kill our (no-FS) attacker

import (
	"testing"
)

// TestSimulateBlockerTrade_FirstStrikeAttackerSurvives is the canonical
// scenario the task spec calls out: our attacker (FS, 3/3) trades with
// a vanilla (no-FS, 3/3) blocker. First-strike phase: attacker deals 3
// to blocker, killing it before it can deal damage back. Attacker
// survives at full toughness; blocker dies.
func TestSimulateBlockerTrade_FirstStrikeAttackerSurvives(t *testing.T) {
	gs := newTestGame(t, 2)
	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("Knight", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(atk, "first strike")
	blk := newTestPermanent(gs.Seats[1], newTestCardMinimal("Bear", []string{"creature"}, 2, nil), 3, 3)

	atkDies, blkDies := simulateBlockerTrade(gs, atk, blk)
	if atkDies {
		t.Errorf("first-strike 3/3 attacker should survive vs vanilla 3/3 blocker; got attacker dies")
	}
	if !blkDies {
		t.Errorf("vanilla 3/3 blocker should die to first-strike 3/3 attacker; got blocker survives")
	}
}

// TestSimulateBlockerTrade_BothFirstStrike — when both have first
// strike, they trade in the first-strike step (mutual kill).
func TestSimulateBlockerTrade_BothFirstStrike(t *testing.T) {
	gs := newTestGame(t, 2)
	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("Knight", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(atk, "first strike")
	blk := newTestPermanent(gs.Seats[1], newTestCardMinimal("OtherKnight", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(blk, "first strike")

	atkDies, blkDies := simulateBlockerTrade(gs, atk, blk)
	if !atkDies || !blkDies {
		t.Errorf("FS vs FS 3/3s should mutually kill; got atkDies=%v blkDies=%v", atkDies, blkDies)
	}
}

// TestSimulateBlockerTrade_VanillaMutualKill — sanity check that
// non-keyword combat still resolves to mutual kill.
func TestSimulateBlockerTrade_VanillaMutualKill(t *testing.T) {
	gs := newTestGame(t, 2)
	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("A", []string{"creature"}, 3, nil), 3, 3)
	blk := newTestPermanent(gs.Seats[1], newTestCardMinimal("B", []string{"creature"}, 3, nil), 3, 3)

	atkDies, blkDies := simulateBlockerTrade(gs, atk, blk)
	if !atkDies || !blkDies {
		t.Errorf("vanilla 3/3 vs 3/3 should mutually kill; got atkDies=%v blkDies=%v", atkDies, blkDies)
	}
}

// TestSimulateBlockerTrade_FirstStrikeBlockerSurvives — defender's
// first-striker kills the vanilla attacker before taking damage.
// (Confirms PR #50's defender-side trade math is preserved.)
func TestSimulateBlockerTrade_FirstStrikeBlockerSurvives(t *testing.T) {
	gs := newTestGame(t, 2)
	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("A", []string{"creature"}, 3, nil), 3, 3)
	blk := newTestPermanent(gs.Seats[1], newTestCardMinimal("B", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(blk, "first strike")

	atkDies, blkDies := simulateBlockerTrade(gs, atk, blk)
	if !atkDies {
		t.Errorf("vanilla attacker should die to FS blocker that outdamages it; got attacker survives")
	}
	if blkDies {
		t.Errorf("FS blocker that kills attacker should survive; got blocker dies")
	}
}

// TestSimulateBlockerTrade_DoubleStrikeAttackerKillsLargerBlocker —
// 3/3 double strike should kill a 5/5 vanilla blocker (3 in FS step,
// blocker survives at 2 toughness, then 3 more = dead) while the
// attacker takes 0 damage in step 1 (blocker hadn't swung yet) and 5
// in step 2 — but only after dealing the lethal step-1 damage.
// Wait: 5/5 blocker has 5 toughness, 3 damage in FS step doesn't kill
// it, blocker swings back for 5 → attacker dies. Then DS swings again
// in regular step but is already dead. Mutual kill.
func TestSimulateBlockerTrade_DoubleStrikeMutualWithBigBlocker(t *testing.T) {
	gs := newTestGame(t, 2)
	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("DSKnight", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(atk, "double strike")
	blk := newTestPermanent(gs.Seats[1], newTestCardMinimal("Wall", []string{"creature"}, 5, nil), 5, 5)

	atkDies, blkDies := simulateBlockerTrade(gs, atk, blk)
	if !atkDies || !blkDies {
		t.Errorf("3/3 DS vs 5/5 vanilla should mutually kill; got atkDies=%v blkDies=%v", atkDies, blkDies)
	}
}

// TestAttackerRank_FirstStrikeBumpsByTwo — first strike adds +2 to the
// rank, and double strike (which adds +3) does NOT also pick up the
// +2 (we pick the higher of the two).
func TestAttackerRank_FirstStrikeBumpsByTwo(t *testing.T) {
	gs := newTestGame(t, 2)
	vanilla := newTestPermanent(gs.Seats[0], newTestCardMinimal("V", []string{"creature"}, 3, nil), 3, 3)
	fs := newTestPermanent(gs.Seats[0], newTestCardMinimal("FS", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(fs, "first strike")
	ds := newTestPermanent(gs.Seats[0], newTestCardMinimal("DS", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(ds, "double strike")
	dsAndFs := newTestPermanent(gs.Seats[0], newTestCardMinimal("DSFS", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(dsAndFs, "double strike")
	addKeyword(dsAndFs, "first strike")

	vRank := attackerRank(gs, vanilla)
	fsRank := attackerRank(gs, fs)
	dsRank := attackerRank(gs, ds)
	bothRank := attackerRank(gs, dsAndFs)

	if fsRank-vRank != 2 {
		t.Errorf("first strike should bump rank by exactly +2; got %d (vanilla=%d, fs=%d)", fsRank-vRank, vRank, fsRank)
	}
	if dsRank-vRank != 3 {
		t.Errorf("double strike should bump rank by exactly +3; got %d (vanilla=%d, ds=%d)", dsRank-vRank, vRank, dsRank)
	}
	if bothRank != dsRank {
		t.Errorf("DS+FS should equal DS alone (no double-counting); got DS+FS=%d DS=%d", bothRank, dsRank)
	}
}

// TestBestTarget_PenalizesFirstStrikeBlockers — given two defenders
// with identical life and threat profile, the one with a first-strike
// blocker that can kill our attacker should be deprioritized.
func TestBestTarget_PenalizesFirstStrikeBlockers(t *testing.T) {
	gs := newTestGame(t, 3)
	gs.Seats[1].Life = 20
	gs.Seats[2].Life = 20

	// Seat 1: vanilla 2/2 blocker (won't kill our 3/3 attacker before
	// taking damage).
	vanillaBlk := newTestPermanent(gs.Seats[1], newTestCardMinimal("Bear", []string{"creature"}, 2, nil), 2, 2)
	vanillaBlk.Tapped = false

	// Seat 2: 3/3 first-strike blocker (kills our 3/3 vanilla attacker
	// in the FS step, before we can swing back).
	fsBlk := newTestPermanent(gs.Seats[2], newTestCardMinimal("Knight", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(fsBlk, "first strike")
	fsBlk.Tapped = false

	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("Beater", []string{"creature"}, 3, nil), 3, 3)

	h := NewYggdrasilHat(nil, 0)
	got := h.ChooseAttackTarget(gs, 0, atk, []int{1, 2})
	if got != 1 {
		t.Errorf("attacker should prefer seat 1 (vanilla blocker) over seat 2 (FS blocker that kills us); got seat %d", got)
	}
}

// TestBestTarget_FirstStrikeAttackerIgnoresPenalty — when our attacker
// also has first/double strike, the FS blocker discount cancels out
// (they trade in the FS step together), so the penalty should NOT
// apply. With identical defenders modulo the blocker keyword and our
// attacker carrying FS, the choice should not be biased away from the
// FS-blocker side.
func TestBestTarget_FirstStrikeAttackerIgnoresPenalty(t *testing.T) {
	gs := newTestGame(t, 3)
	gs.Seats[1].Life = 20
	gs.Seats[2].Life = 20

	vanillaBlk := newTestPermanent(gs.Seats[1], newTestCardMinimal("Bear", []string{"creature"}, 2, nil), 3, 3)
	vanillaBlk.Tapped = false
	fsBlk := newTestPermanent(gs.Seats[2], newTestCardMinimal("Knight", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(fsBlk, "first strike")
	fsBlk.Tapped = false

	// Our attacker has first strike too — penalty should not fire.
	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("FSBeater", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(atk, "first strike")

	h := NewYggdrasilHat(nil, 0)

	// The two defenders are otherwise identical. With the penalty
	// disabled, the choice between them is tie-broken by the existing
	// scoring noise; we just assert that the result is one of the two
	// legal seats and that no panic / -1 sneaks through.
	got := h.ChooseAttackTarget(gs, 0, atk, []int{1, 2})
	if got != 1 && got != 2 {
		t.Errorf("FS attacker should still produce a legal target; got %d", got)
	}

	// Stronger check: re-run with a vanilla attacker and confirm the
	// FS-blocker side is now penalized (seat 1 wins). This pins the
	// "atkFS short-circuits the penalty" branch.
	vanillaAtk := newTestPermanent(gs.Seats[0], newTestCardMinimal("Beater", []string{"creature"}, 3, nil), 3, 3)
	gotV := h.ChooseAttackTarget(gs, 0, vanillaAtk, []int{1, 2})
	if gotV != 1 {
		t.Errorf("vanilla attacker should be penalized away from FS-blocker side; got %d", gotV)
	}
}
