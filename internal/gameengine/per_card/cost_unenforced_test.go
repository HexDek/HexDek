package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// dev/cost-unenforced-era1 — verifies the new defensive cost gates added
// to era1 + era2 commander handlers. Each test pairs a "cost paid" path
// (the original behavior is unchanged once costs are met) with a "cost
// missing" negative path (the handler now fails cleanly via emitFail
// instead of resolving the effect for free).

// ---------------------------------------------------------------------------
// Bristly Bill — {3}{G}{G} double-counters
// ---------------------------------------------------------------------------

func TestBristlyBillCost_InsufficientManaBlocksDouble(t *testing.T) {
	gs := newGame(t, 2)
	bill := addPerm(gs, 0, "Bristly Bill, Spine Sower", "creature", "legendary")
	bill.AddCounter("+1/+1", 4)
	gs.Seats[0].ManaPool = 4 // {3}{G}{G} = 5

	bristlyBillSpineSowerDouble(gs, bill, 0, nil)

	if bill.Counters["+1/+1"] != 4 {
		t.Fatalf("counters should not double when mana is short; got %d", bill.Counters["+1/+1"])
	}
	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when mana is insufficient")
	}
}

// ---------------------------------------------------------------------------
// Kardur — goad flag expires at controller's next upkeep
// ---------------------------------------------------------------------------

func TestKardurCost_GoadExpiryDelayedTriggerRegistered(t *testing.T) {
	gs := newGame(t, 3)
	kardur := addPerm(gs, 0, "Kardur, Doomscourge", "creature", "legendary")
	delayedBefore := len(gs.DelayedTriggers)

	kardurETBGoadFlag(gs, kardur)

	if gs.Flags["kardur_goad_seat"] != 1 {
		t.Fatalf("expected goad flag set to seat+1=1, got %d", gs.Flags["kardur_goad_seat"])
	}
	if len(gs.DelayedTriggers) != delayedBefore+1 {
		t.Fatalf("expected one delayed-trigger registered for goad expiry; before=%d after=%d",
			delayedBefore, len(gs.DelayedTriggers))
	}
	dt := gs.DelayedTriggers[len(gs.DelayedTriggers)-1]
	if dt.TriggerAt != "next_upkeep" {
		t.Errorf("expected TriggerAt=next_upkeep; got %q", dt.TriggerAt)
	}
	if dt.ControllerSeat != 0 {
		t.Errorf("expected ControllerSeat=0; got %d", dt.ControllerSeat)
	}
	// Run the cleanup function — flag should clear.
	dt.EffectFn(gs)
	if _, present := gs.Flags["kardur_goad_seat"]; present {
		t.Errorf("goad flag should be cleared after delayed trigger fires")
	}
}

// ---------------------------------------------------------------------------
// Commander Mustard — {2}{R}{W} soldier-attack-ping flag
// ---------------------------------------------------------------------------

func TestCommanderMustardCost_PaysAndSetsFlag(t *testing.T) {
	gs := newGame(t, 2)
	mustard := addPerm(gs, 0, "Commander Mustard", "creature", "legendary")
	gs.Seats[0].ManaPool = 4

	commanderMustardActivate(gs, mustard, 0, nil)

	if gs.Seats[0].ManaPool != 0 {
		t.Errorf("expected mana pool drained to 0 after paying 4; got %d", gs.Seats[0].ManaPool)
	}
	if gs.Seats[0].Flags["mustard_soldier_attack_ping"] != 1 {
		t.Errorf("expected soldier-ping flag set to 1; got %d", gs.Seats[0].Flags["mustard_soldier_attack_ping"])
	}
}

func TestCommanderMustardCost_InsufficientManaBlocks(t *testing.T) {
	gs := newGame(t, 2)
	mustard := addPerm(gs, 0, "Commander Mustard", "creature", "legendary")
	gs.Seats[0].ManaPool = 3 // need 4

	commanderMustardActivate(gs, mustard, 0, nil)

	if gs.Seats[0].Flags["mustard_soldier_attack_ping"] != 0 {
		t.Errorf("flag should NOT be set when mana is short; got %d",
			gs.Seats[0].Flags["mustard_soldier_attack_ping"])
	}
	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when mana insufficient")
	}
}

// ---------------------------------------------------------------------------
// Jolly Balloon Man — {1},{T}, sorcery-speed copy
// ---------------------------------------------------------------------------

func TestJollyBalloonManCost_CopiesWhenAllCostsPaid(t *testing.T) {
	gs := newGame(t, 2)
	jbm := addPerm(gs, 0, "The Jolly Balloon Man", "creature", "legendary")
	target := addPerm(gs, 0, "Avenger of Zendikar", "creature")
	target.Card.BasePower = 5
	target.Card.BaseToughness = 5
	gs.Seats[0].ManaPool = 1
	gs.Active = 0
	gs.Phase = "precombat_main"
	bfBefore := len(gs.Seats[0].Battlefield)

	theJollyBalloonManCopy(gs, jbm, 0, nil)

	if len(gs.Seats[0].Battlefield) != bfBefore+1 {
		t.Fatalf("expected one Balloon token; before=%d after=%d", bfBefore, len(gs.Seats[0].Battlefield))
	}
	if !jbm.Tapped {
		t.Errorf("Jolly Balloon Man should be tapped after activation")
	}
	if gs.Seats[0].ManaPool != 0 {
		t.Errorf("mana should be drained; got %d", gs.Seats[0].ManaPool)
	}
	tok := gs.Seats[0].Battlefield[len(gs.Seats[0].Battlefield)-1]
	if tok.Card.BasePower != 1 || tok.Card.BaseToughness != 1 {
		t.Errorf("expected 1/1 token; got %d/%d", tok.Card.BasePower, tok.Card.BaseToughness)
	}
}

func TestJollyBalloonManCost_FailsAtInstantSpeed(t *testing.T) {
	gs := newGame(t, 2)
	jbm := addPerm(gs, 0, "The Jolly Balloon Man", "creature", "legendary")
	addPerm(gs, 0, "Bear", "creature")
	gs.Seats[0].ManaPool = 1
	gs.Active = 1 // not the controller's turn
	gs.Phase = "precombat_main"

	theJollyBalloonManCopy(gs, jbm, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when activated outside sorcery speed")
	}
	if jbm.Tapped {
		t.Errorf("source should not be tapped on failed activation")
	}
	if gs.Seats[0].ManaPool != 1 {
		t.Errorf("mana should not be paid on failed activation; got %d", gs.Seats[0].ManaPool)
	}
}

func TestJollyBalloonManCost_FailsWhenAlreadyTapped(t *testing.T) {
	gs := newGame(t, 2)
	jbm := addPerm(gs, 0, "The Jolly Balloon Man", "creature", "legendary")
	jbm.Tapped = true
	addPerm(gs, 0, "Bear", "creature")
	gs.Seats[0].ManaPool = 1
	gs.Active = 0
	gs.Phase = "precombat_main"

	theJollyBalloonManCopy(gs, jbm, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when source already tapped")
	}
	if gs.Seats[0].ManaPool != 1 {
		t.Errorf("mana should not be paid on failed activation")
	}
}

// ---------------------------------------------------------------------------
// Master of Keys — X-counter ETB reads cost-paid X
// ---------------------------------------------------------------------------

func TestMasterOfKeysCost_AppliesXCountersAndMills2X(t *testing.T) {
	gs := newGame(t, 2)
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	gs.Flags["_master_of_keys_x_0"] = 3
	addLibrary(gs, 0, "A", "B", "C", "D", "E", "F", "G")
	master := addPerm(gs, 0, "The Master of Keys", "creature", "legendary")

	theMasterOfKeysETB(gs, master)

	if master.Counters["+1/+1"] != 3 {
		t.Errorf("expected 3 +1/+1 counters from X=3; got %d", master.Counters["+1/+1"])
	}
	if len(gs.Seats[0].Graveyard) != 6 {
		t.Errorf("expected 2X=6 cards milled; gy=%d", len(gs.Seats[0].Graveyard))
	}
}

func TestMasterOfKeysCost_NoXIsNoop(t *testing.T) {
	gs := newGame(t, 2)
	addLibrary(gs, 0, "A", "B", "C")
	master := addPerm(gs, 0, "The Master of Keys", "creature", "legendary")

	theMasterOfKeysETB(gs, master)

	if master.Counters["+1/+1"] != 0 {
		t.Errorf("expected 0 counters when X flag absent; got %d", master.Counters["+1/+1"])
	}
	if len(gs.Seats[0].Graveyard) != 0 {
		t.Errorf("expected no mill when X=0; gy=%d", len(gs.Seats[0].Graveyard))
	}
}

// ---------------------------------------------------------------------------
// Ezrim — {1}, Sacrifice an artifact, grant lifelink
// ---------------------------------------------------------------------------

func TestEzrimCost_SacrificesAndGrantsLifelink(t *testing.T) {
	gs := newGame(t, 2)
	ezrim := addPerm(gs, 0, "Ezrim, Agency Chief", "creature", "legendary")
	clue := addPerm(gs, 0, "Clue Token", "artifact", "token")
	gs.Seats[0].ManaPool = 1

	ezrimSacGrantKeyword(gs, ezrim, 0, nil)

	if ezrim.Flags["kw:lifelink"] != 1 {
		t.Errorf("expected kw:lifelink flag set on Ezrim; flags=%+v", ezrim.Flags)
	}
	for _, p := range gs.Seats[0].Battlefield {
		if p == clue {
			t.Errorf("clue should have been sacrificed")
		}
	}
	if gs.Seats[0].ManaPool != 0 {
		t.Errorf("expected mana drained; pool=%d", gs.Seats[0].ManaPool)
	}
}

func TestEzrimCost_FailsWithNoArtifactToSacrifice(t *testing.T) {
	gs := newGame(t, 2)
	ezrim := addPerm(gs, 0, "Ezrim, Agency Chief", "creature", "legendary")
	gs.Seats[0].ManaPool = 1

	ezrimSacGrantKeyword(gs, ezrim, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when no artifact to sacrifice")
	}
	if ezrim.Flags["kw:lifelink"] == 1 {
		t.Errorf("lifelink should not be granted on failed activation")
	}
	if gs.Seats[0].ManaPool != 1 {
		t.Errorf("mana should not be paid on failed activation; pool=%d", gs.Seats[0].ManaPool)
	}
}

func TestEzrimCost_FailsWithNoMana(t *testing.T) {
	gs := newGame(t, 2)
	ezrim := addPerm(gs, 0, "Ezrim, Agency Chief", "creature", "legendary")
	clue := addPerm(gs, 0, "Clue Token", "artifact", "token")
	gs.Seats[0].ManaPool = 0

	ezrimSacGrantKeyword(gs, ezrim, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when mana insufficient")
	}
	// Clue should still be on battlefield — sacrifice never happened.
	stillThere := false
	for _, p := range gs.Seats[0].Battlefield {
		if p == clue {
			stillThere = true
		}
	}
	if !stillThere {
		t.Errorf("clue must NOT be sacrificed when mana cost cannot be paid")
	}
}

// ---------------------------------------------------------------------------
// Sliver Gravemother — encore sorcery-speed gate
// ---------------------------------------------------------------------------

func TestSliverGravemotherCost_FailsAtInstantSpeed(t *testing.T) {
	gs := newGame(t, 4)
	mother := addPerm(gs, 0, "Sliver Gravemother", "creature", "sliver", "legendary")
	gs.Seats[0].ManaPool = 5
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, &gameengine.Card{
		Name:          "Two-Headed Sliver",
		Owner:         0,
		Types:         []string{"creature", "sliver", "cmc:3"},
		BasePower:     2,
		BaseToughness: 2,
	})
	// Opponent's turn — sorcery-speed gate must reject.
	gs.Active = 1
	gs.Phase = "precombat_main"

	bfBefore := len(gs.Seats[0].Battlefield)
	sliverGravemotherEncore(gs, mother, 0, nil)

	if len(gs.Seats[0].Battlefield) != bfBefore {
		t.Errorf("no tokens should spawn at instant speed; before=%d after=%d",
			bfBefore, len(gs.Seats[0].Battlefield))
	}
	if gs.Seats[0].ManaPool != 5 {
		t.Errorf("mana should not be paid on rejected activation; pool=%d", gs.Seats[0].ManaPool)
	}
	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed event")
	}
}

// ---------------------------------------------------------------------------
// Yenna — {2},{T} sorcery-only copy enchantment
// ---------------------------------------------------------------------------

func TestYennaCost_FailsAtInstantSpeed(t *testing.T) {
	gs := newGame(t, 2)
	yenna := addPerm(gs, 0, "Yenna, Redtooth Regent", "creature")
	addPerm(gs, 0, "Rhystic Study", "enchantment")
	gs.Seats[0].ManaPool = 2
	gs.Active = 1 // opponent's turn
	gs.Phase = "precombat_main"

	bfBefore := len(gs.Seats[0].Battlefield)
	yennaCopyEnchantment(gs, yenna, 0, nil)

	if len(gs.Seats[0].Battlefield) != bfBefore {
		t.Errorf("no copy at instant speed; before=%d after=%d", bfBefore, len(gs.Seats[0].Battlefield))
	}
	if gs.Seats[0].ManaPool != 2 {
		t.Errorf("mana should not be paid on rejected activation; pool=%d", gs.Seats[0].ManaPool)
	}
}

func TestYennaCost_FailsWhenAlreadyTapped(t *testing.T) {
	gs := newGame(t, 2)
	yenna := addPerm(gs, 0, "Yenna, Redtooth Regent", "creature")
	yenna.Tapped = true
	addPerm(gs, 0, "Rhystic Study", "enchantment")
	gs.Seats[0].ManaPool = 2
	gs.Active = 0
	gs.Phase = "precombat_main"

	yennaCopyEnchantment(gs, yenna, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when source already tapped")
	}
	if gs.Seats[0].ManaPool != 2 {
		t.Errorf("mana should not be paid on rejected activation")
	}
}

func TestYennaCost_FailsWithInsufficientMana(t *testing.T) {
	gs := newGame(t, 2)
	yenna := addPerm(gs, 0, "Yenna, Redtooth Regent", "creature")
	addPerm(gs, 0, "Rhystic Study", "enchantment")
	gs.Seats[0].ManaPool = 1
	gs.Active = 0
	gs.Phase = "precombat_main"

	yennaCopyEnchantment(gs, yenna, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed for mana shortfall")
	}
}

// ---------------------------------------------------------------------------
// Felothar — {3},{T},Sac creature; reject when source already tapped
// ---------------------------------------------------------------------------

func TestFelotharCost_FailsWhenAlreadyTapped(t *testing.T) {
	gs := newGame(t, 2)
	felothar := addPerm(gs, 0, "Felothar the Steadfast", "creature")
	felothar.Tapped = true
	gs.Seats[0].ManaPool = 3
	sac := addPerm(gs, 0, "Wall of Mulch", "creature")
	sac.Card.BasePower = 2
	sac.Card.BaseToughness = 4

	felotharSacDrawDiscard(gs, felothar, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when source already tapped")
	}
	if gs.Seats[0].ManaPool != 3 {
		t.Errorf("mana should not be paid on rejected activation")
	}
}

// ---------------------------------------------------------------------------
// Mayael — {3}{R}{G}{W},{T} mana + tap gates
// ---------------------------------------------------------------------------

func TestMayaelCost_FailsWithoutMana(t *testing.T) {
	gs := newGame(t, 2)
	mayael := addPerm(gs, 0, "Mayael the Anima", "creature")
	gs.Seats[0].Library = []*gameengine.Card{
		{Name: "Avenger of Zendikar", Owner: 0, Types: []string{"creature"}, BasePower: 5},
	}
	gs.Seats[0].ManaPool = 5 // need 6

	mayaelLookFive(gs, mayael, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when mana short")
	}
	if mayael.Tapped {
		t.Errorf("Mayael should not be tapped on failed activation")
	}
	if len(gs.Seats[0].Library) != 1 {
		t.Errorf("library should be untouched on failed activation")
	}
}

func TestMayaelCost_FailsWhenAlreadyTapped(t *testing.T) {
	gs := newGame(t, 2)
	mayael := addPerm(gs, 0, "Mayael the Anima", "creature")
	mayael.Tapped = true
	gs.Seats[0].ManaPool = 6
	gs.Seats[0].Library = []*gameengine.Card{
		{Name: "Avenger of Zendikar", Owner: 0, Types: []string{"creature"}, BasePower: 5},
	}

	mayaelLookFive(gs, mayael, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when source already tapped")
	}
	if gs.Seats[0].ManaPool != 6 {
		t.Errorf("mana should not be paid on rejected activation")
	}
}

func TestMayaelCost_TapsAndPaysOnSuccess(t *testing.T) {
	gs := newGame(t, 2)
	mayael := addPerm(gs, 0, "Mayael the Anima", "creature")
	gs.Seats[0].ManaPool = 6
	gs.Seats[0].Library = []*gameengine.Card{
		{Name: "Avenger of Zendikar", Owner: 0, Types: []string{"creature"}, BasePower: 5},
	}

	mayaelLookFive(gs, mayael, 0, nil)

	if !mayael.Tapped {
		t.Errorf("Mayael should be tapped after successful activation")
	}
	if gs.Seats[0].ManaPool != 0 {
		t.Errorf("expected 6 mana drained; pool=%d", gs.Seats[0].ManaPool)
	}
}
