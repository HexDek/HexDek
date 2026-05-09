package gameengine

import (
	"testing"
)

// dev/commander-damage-tracking — verify that combat damage from a
// commander accumulates into Seat.CommanderDamage[dealer][name] and
// trips the §704.6c SBA at 21.

// TestCombatCommanderDamage_AccumulatesAndLoses puts a 21/21 commander
// in front of a defender, swings unblocked, and asserts both the
// per-source bucket and the SBA loss fire.
func TestCombatCommanderDamage_AccumulatesAndLoses(t *testing.T) {
	gs := newCommanderGame(t, 2, "A", "B")
	atkSeat, defSeat := 0, 1
	cmdrCard := gs.Seats[atkSeat].CommandZone[0]
	cmdrCard.BasePower = 21
	cmdrCard.BaseToughness = 21
	gs.Seats[atkSeat].CommandZone = nil
	atk := &Permanent{
		Card:       cmdrCard,
		Controller: atkSeat,
		Owner:      atkSeat,
		Counters:   map[string]int{},
		Flags:      map[string]int{"attacking": 1},
	}
	SetAttackerDefender(atk, defSeat)
	gs.Seats[atkSeat].Battlefield = append(gs.Seats[atkSeat].Battlefield, atk)

	DealCombatDamageStep(gs, []*Permanent{atk}, map[*Permanent][]*Permanent{atk: nil}, false)
	StateBasedActions(gs)

	if got := CommanderDamageFrom(gs.Seats[defSeat], atkSeat, "A"); got != 21 {
		t.Fatalf("expected 21 commander damage from A; got %d", got)
	}
	if !gs.Seats[defSeat].Lost {
		t.Fatalf("defender should lose at 21 commander damage (CR §704.6c)")
	}
}

// TestCombatCommanderDamage_NonCommanderDoesNotAccumulate verifies that
// a non-commander hitting for 19 doesn't pollute the commander-damage
// tracker (and 19 from a non-commander leaves a 40-life defender alive).
func TestCombatCommanderDamage_NonCommanderDoesNotAccumulate(t *testing.T) {
	gs := newCommanderGame(t, 2, "A", "B")
	defSeat := 1
	beast := &Card{
		Name:          "Vanilla 19/19",
		Owner:         0,
		BasePower:     19,
		BaseToughness: 19,
		Types:         []string{"creature"},
	}
	atk := &Permanent{
		Card:       beast,
		Controller: 0,
		Owner:      0,
		Counters:   map[string]int{},
		Flags:      map[string]int{"attacking": 1},
	}
	SetAttackerDefender(atk, defSeat)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, atk)

	DealCombatDamageStep(gs, []*Permanent{atk}, map[*Permanent][]*Permanent{atk: nil}, false)
	StateBasedActions(gs)

	if got := CommanderDamageFrom(gs.Seats[defSeat], 0, "Vanilla 19/19"); got != 0 {
		t.Errorf("non-commander damage must not accumulate as commander damage; got %d", got)
	}
	if gs.Seats[defSeat].Lost {
		t.Fatalf("defender at 40 - 19 = 21 life should still be alive")
	}
}

// TestCombatCommanderDamage_PartnerKeyedSeparately confirms that two
// partner-style commanders dealing 15 each don't combine — neither
// bucket reaches 21 — but a single commander reaching 21 does.
func TestCombatCommanderDamage_PartnerKeyedSeparately(t *testing.T) {
	gs := newCommanderGame(t, 2, "A", "B")
	defSeat := 1
	AccumulateCommanderDamage(gs, defSeat, 0, "Kraum", 15)
	AccumulateCommanderDamage(gs, defSeat, 0, "Tymna", 15)
	StateBasedActions(gs)
	if gs.Seats[defSeat].Lost {
		t.Fatalf("defender should survive 15 each from two distinct commanders")
	}
	AccumulateCommanderDamage(gs, defSeat, 0, "Kraum", 6)
	StateBasedActions(gs)
	if !gs.Seats[defSeat].Lost {
		t.Fatalf("defender should lose at 21 from Kraum")
	}
	if gs.Seats[defSeat].LossReason == "" {
		t.Errorf("loss reason should be populated; got empty")
	}
}
