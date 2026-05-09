package hat

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// dev/commander-damage-tracking — verifies YggdrasilHat.AssignBlockers
// blocks an incoming commander when the commander-damage clock would
// reach 21 (CR §704.6c) even when life total looks comfortable.
//
// Setup: 2-seat commander game, defender at 35 life with 16 commander
// damage already on the board from seat 0's "Krenko". Seat 0 swings
// with Krenko (5/5). Without the commander-damage gate the hat would
// see "5 damage vs 35 life" and decline to block. With the gate, 16+5
// = 21 → must-block fires and a blocker is assigned.
func TestYggdrasil_BlocksCommanderAtClock16(t *testing.T) {
	gs := gameengine.NewGameState(2, nil, nil)
	gs.CommanderFormat = true

	// Seat 0 is the attacker — register Krenko as their commander.
	krenkoCard := newTestCardMinimal("Krenko, Mob Boss", []string{"creature", "legendary"}, 4, nil)
	gs.Seats[0].CommanderNames = []string{"Krenko, Mob Boss"}
	krenko := newTestPermanent(gs.Seats[0], krenkoCard, 5, 5)

	// Seat 1 is the defender — high life, but already at 16 cmdr damage
	// from Krenko. A 5-power swing would clock to 21 → kill.
	gs.Seats[1].Life = 35
	gs.Seats[1].CommanderDamage = map[int]map[string]int{
		0: {"Krenko, Mob Boss": 16},
	}
	// Defender has a 1/1 chump available.
	chumpCard := newTestCardMinimal("Squire", []string{"creature"}, 1, nil)
	chump := newTestPermanent(gs.Seats[1], chumpCard, 1, 1)
	chump.Tapped = false

	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{krenko})

	if len(out[krenko]) == 0 {
		t.Fatalf("hat must block Krenko (clock 16 + 5 power = 21); got no blockers. Out=%v", out)
	}
	if out[krenko][0] != chump {
		t.Errorf("expected the chump assigned as blocker; got %+v", out[krenko])
	}
}

// TestAttackerRank_CommanderBonus — the +10 commander bonus in
// poker.attackerRank ensures a commander attacker outranks a vanilla
// attacker of identical stats so Yggdrasil's ranking loop processes it
// first when limited blockers are available.
func TestAttackerRank_CommanderBonus(t *testing.T) {
	gs := gameengine.NewGameState(2, nil, nil)
	gs.CommanderFormat = true
	gs.Seats[0].CommanderNames = []string{"Cmdr"}

	cmdr := newTestPermanent(gs.Seats[0],
		newTestCardMinimal("Cmdr", []string{"creature", "legendary"}, 3, nil), 3, 3)
	vanilla := newTestPermanent(gs.Seats[0],
		newTestCardMinimal("Beast", []string{"creature"}, 3, nil), 3, 3)

	cmdrRank := attackerRank(gs, cmdr)
	vanillaRank := attackerRank(gs, vanilla)

	if cmdrRank-vanillaRank != 10 {
		t.Errorf("commander attacker should rank exactly +10 over identical vanilla; cmdr=%d vanilla=%d (delta %d)",
			cmdrRank, vanillaRank, cmdrRank-vanillaRank)
	}
}
