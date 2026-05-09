package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// Tests for the activated-stub batch (Phenax, Obeka, Jhoira,
// Shadowheart, Splinter, Ghen). Each test calls the activated handler
// directly rather than going through the engine's activation
// dispatcher — the dispatcher's mana / tap gates are exercised by
// other test suites; what we cover here is the resolver shape.

// ---------------------------------------------------------------------
// Phenax — granted mill
// ---------------------------------------------------------------------

func TestPhenax_TapsCreatureAndMillsToughnessFromOpponent(t *testing.T) {
	gs := newGame(t, 2)
	phenax := addPerm(gs, 0, "Phenax, God of Deception", "creature")
	wall := addPerm(gs, 0, "Wall of Souls", "creature")
	wall.Card.BaseToughness = 5
	addLibrary(gs, 1, "C1", "C2", "C3", "C4", "C5", "C6", "C7")

	phenaxGrantedMill(gs, phenax, 0, nil)

	if !wall.Tapped {
		t.Errorf("highest-toughness creature should be tapped, got Tapped=false")
	}
	if got := len(gs.Seats[1].Graveyard); got != 5 {
		t.Errorf("expected 5 mills (toughness=5); got graveyard size=%d", got)
	}
	if got := len(gs.Seats[1].Library); got != 2 {
		t.Errorf("expected library to drop from 7 to 2; got %d", got)
	}
	if hasEvent(gs, "per_card_handler") < 1 {
		t.Errorf("expected per_card_handler event")
	}
}

func TestPhenax_NoUntappedCreatureFails(t *testing.T) {
	gs := newGame(t, 2)
	phenax := addPerm(gs, 0, "Phenax, God of Deception", "creature")
	phenax.Tapped = true
	addLibrary(gs, 1, "X")

	phenaxGrantedMill(gs, phenax, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when no untapped creature available")
	}
	if len(gs.Seats[1].Graveyard) != 0 {
		t.Errorf("no mills expected, got %d", len(gs.Seats[1].Graveyard))
	}
}

// ---------------------------------------------------------------------
// Obeka — end the turn
// ---------------------------------------------------------------------

func TestObeka_TapsAndCountersStack(t *testing.T) {
	gs := newGame(t, 2)
	obeka := addPerm(gs, 0, "Obeka, Brute Chronologist", "creature")
	gs.Stack = []*gameengine.StackItem{
		{Card: &gameengine.Card{Name: "Spell A"}, Controller: 1},
		{Card: &gameengine.Card{Name: "Spell B"}, Controller: 0},
		{Card: &gameengine.Card{Name: "Already Countered"}, Controller: 1, Countered: true},
	}

	obekaEndTheTurn(gs, obeka, 0, nil)

	if !obeka.Tapped {
		t.Errorf("Obeka should be tapped")
	}
	if !gs.Stack[0].Countered || !gs.Stack[1].Countered {
		t.Errorf("expected uncountered stack items to be countered, got %+v", gs.Stack)
	}
	if gs.Flags["end_turn_requested"] != 1 {
		t.Errorf("end_turn_requested flag not set")
	}
	if hasEvent(gs, "end_turn") < 1 {
		t.Errorf("expected end_turn event")
	}
	if hasEvent(gs, "per_card_partial") < 1 {
		t.Errorf("expected per_card_partial flagging missing phase machinery")
	}
}

func TestObeka_AlreadyTappedFails(t *testing.T) {
	gs := newGame(t, 2)
	obeka := addPerm(gs, 0, "Obeka, Brute Chronologist", "creature")
	obeka.Tapped = true

	obekaEndTheTurn(gs, obeka, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when already tapped")
	}
}

// ---------------------------------------------------------------------
// Jhoira — ingenuity counters + cheat artifact
// ---------------------------------------------------------------------

func TestJhoira_AddsCountersAndCheatsArtifact(t *testing.T) {
	gs := newGame(t, 2)
	jhoira := addPerm(gs, 0, "Jhoira, Ageless Innovator", "creature")
	// Hand has a CMC-2 artifact and a CMC-5 artifact. After +2 counters
	// X=2 so only the CMC-2 should land (not the CMC-5).
	cheap := &gameengine.Card{Name: "Sol Ring", Owner: 0, Types: []string{"artifact", "cmc:2"}}
	expensive := &gameengine.Card{Name: "Mox Opal", Owner: 0, Types: []string{"artifact", "cmc:5"}}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, cheap, expensive)

	jhoiraIngenuityActivate(gs, jhoira, 0, nil)

	if !jhoira.Tapped {
		t.Errorf("Jhoira should be tapped")
	}
	if jhoira.Counters["ingenuity"] != 2 {
		t.Errorf("ingenuity counters=%d, want 2", jhoira.Counters["ingenuity"])
	}
	// Cheap artifact should be on battlefield, expensive should still be in hand.
	cheapOnBoard := false
	for _, p := range gs.Seats[0].Battlefield {
		if p != nil && p.Card == cheap {
			cheapOnBoard = true
		}
		if p != nil && p.Card == expensive {
			t.Errorf("CMC-5 artifact should NOT be on battlefield with X=2")
		}
	}
	if !cheapOnBoard {
		t.Errorf("CMC-2 artifact should be on battlefield")
	}
	expensiveStillInHand := false
	for _, c := range gs.Seats[0].Hand {
		if c == expensive {
			expensiveStillInHand = true
		}
	}
	if !expensiveStillInHand {
		t.Errorf("CMC-5 artifact should still be in hand")
	}
}

func TestJhoira_NoEligibleArtifactStillAddsCounters(t *testing.T) {
	gs := newGame(t, 2)
	jhoira := addPerm(gs, 0, "Jhoira, Ageless Innovator", "creature")
	// No artifacts in hand.
	gs.Seats[0].Hand = append(gs.Seats[0].Hand,
		&gameengine.Card{Name: "Lightning Bolt", Owner: 0, Types: []string{"instant", "cmc:1"}})

	jhoiraIngenuityActivate(gs, jhoira, 0, nil)

	if jhoira.Counters["ingenuity"] != 2 {
		t.Errorf("counters should still be added even when nothing to cheat: got %d",
			jhoira.Counters["ingenuity"])
	}
}

// ---------------------------------------------------------------------
// Shadowheart — sac creature, draw X = power
// ---------------------------------------------------------------------

func TestShadowheart_SacsBestCreatureAndDrawsX(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].ManaPool = 4
	sh := addPerm(gs, 0, "Shadowheart, Dark Justiciar", "creature")
	// Two non-commander creatures: power 4 and power 2. We expect the
	// 4-power one to be sacrificed (maximizes X).
	big := addPerm(gs, 0, "Hill Giant", "creature")
	big.Card.BasePower = 4
	big.Card.BaseToughness = 3
	small := addPerm(gs, 0, "Goblin", "creature")
	small.Card.BasePower = 2
	small.Card.BaseToughness = 2
	addLibrary(gs, 0, "C1", "C2", "C3", "C4", "C5", "C6")

	shadowheartSacDraw(gs, sh, 0, nil)

	if !sh.Tapped {
		t.Errorf("Shadowheart should be tapped")
	}
	if gs.Seats[0].ManaPool != 2 {
		t.Errorf("mana not paid: pool=%d, want 2", gs.Seats[0].ManaPool)
	}
	// Big creature should be in graveyard, small still on battlefield.
	bigInGY := false
	for _, c := range gs.Seats[0].Graveyard {
		if c == big.Card {
			bigInGY = true
		}
	}
	if !bigInGY {
		t.Errorf("4-power creature should have been sacrificed (highest power)")
	}
	stillBoard := false
	for _, p := range gs.Seats[0].Battlefield {
		if p == small {
			stillBoard = true
		}
	}
	if !stillBoard {
		t.Errorf("2-power creature should still be on battlefield")
	}
	// Drew 4 cards.
	if got := len(gs.Seats[0].Hand); got != 4 {
		t.Errorf("expected 4 cards drawn (X=4); got hand size=%d", got)
	}
}

func TestShadowheart_InsufficientManaFails(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].ManaPool = 1
	sh := addPerm(gs, 0, "Shadowheart, Dark Justiciar", "creature")
	addPerm(gs, 0, "Goblin", "creature")

	shadowheartSacDraw(gs, sh, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when mana < 2")
	}
}

func TestShadowheart_NoSacrificeTargetFails(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].ManaPool = 4
	sh := addPerm(gs, 0, "Shadowheart, Dark Justiciar", "creature")
	// No other creatures.

	shadowheartSacDraw(gs, sh, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when nothing to sacrifice")
	}
}

// ---------------------------------------------------------------------
// Splinter — Ninja can't be blocked this turn
// ---------------------------------------------------------------------

func TestSplinter_PicksHighestPowerNinjaAndFlagsUnblockable(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].ManaPool = 2
	splinter := addPerm(gs, 0, "Splinter, Radical Rat", "creature")
	weak := addPerm(gs, 0, "Throat Slitter", "creature", "ninja")
	weak.Card.BasePower = 2
	strong := addPerm(gs, 0, "Higure, the Still Wind", "creature", "ninja")
	strong.Card.BasePower = 5
	notNinja := addPerm(gs, 0, "Goblin", "creature")
	notNinja.Card.BasePower = 6

	splinterNinjaUnblockable(gs, splinter, 0, nil)

	if gs.Seats[0].ManaPool != 0 {
		t.Errorf("mana should be 0 after paying {1}{U}; got %d", gs.Seats[0].ManaPool)
	}
	if strong.Flags["unblockable"] != 1 {
		t.Errorf("highest-power Ninja should have unblockable=1; got %v", strong.Flags)
	}
	if weak.Flags["unblockable"] == 1 {
		t.Errorf("weaker Ninja should NOT have been picked")
	}
	if notNinja.Flags["unblockable"] == 1 {
		t.Errorf("non-Ninja should NEVER be picked")
	}
}

func TestSplinter_NoNinjaFails(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].ManaPool = 2
	splinter := addPerm(gs, 0, "Splinter, Radical Rat", "creature")
	addPerm(gs, 0, "Goblin", "creature")

	splinterNinjaUnblockable(gs, splinter, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when no Ninja controlled")
	}
}

// ---------------------------------------------------------------------
// Ghen — return enchantment from graveyard
// ---------------------------------------------------------------------

func TestGhen_SacsEnchantmentAndReturnsBest(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].ManaPool = 3
	ghen := addPerm(gs, 0, "Ghen, Arcanum Weaver", "creature")
	// Sacrifice fodder (low CMC).
	fodder := addPerm(gs, 0, "Pacifism", "enchantment", "cmc:2")
	// Two enchantments in graveyard — should pick highest CMC.
	cheapEnch := &gameengine.Card{Name: "Spreading Seas", Owner: 0, Types: []string{"enchantment", "cmc:2"}}
	bigEnch := &gameengine.Card{Name: "Smothering Tithe", Owner: 0, Types: []string{"enchantment", "cmc:4"}}
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, cheapEnch, bigEnch)

	ghenEnchantmentRecursion(gs, ghen, 0, nil)

	if !ghen.Tapped {
		t.Errorf("Ghen should be tapped")
	}
	if gs.Seats[0].ManaPool != 0 {
		t.Errorf("mana not paid: pool=%d, want 0", gs.Seats[0].ManaPool)
	}
	// Fodder should be sacrificed.
	fodderStillBoard := false
	for _, p := range gs.Seats[0].Battlefield {
		if p == fodder {
			fodderStillBoard = true
		}
	}
	if fodderStillBoard {
		t.Errorf("sac fodder enchantment should be off the battlefield")
	}
	// Big enchantment should be on battlefield, no longer in graveyard.
	bigOnBoard := false
	for _, p := range gs.Seats[0].Battlefield {
		if p != nil && p.Card == bigEnch {
			bigOnBoard = true
		}
	}
	if !bigOnBoard {
		t.Errorf("highest-CMC enchantment should have been returned to battlefield")
	}
	for _, c := range gs.Seats[0].Graveyard {
		if c == bigEnch {
			t.Errorf("returned enchantment should not still be in graveyard")
		}
	}
}

func TestGhen_NoEnchantmentToSacFails(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].ManaPool = 3
	ghen := addPerm(gs, 0, "Ghen, Arcanum Weaver", "creature")
	// Graveyard has an enchantment but no other enchantment to sac.
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard,
		&gameengine.Card{Name: "Pacifism", Owner: 0, Types: []string{"enchantment", "cmc:2"}})

	ghenEnchantmentRecursion(gs, ghen, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when no other enchantment to sac")
	}
}
