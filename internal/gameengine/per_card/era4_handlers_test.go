package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// dev/era4-unification (PR#34) added 10 hand-written commander handlers
// without dedicated test coverage (kalamax/tiamat/veyran touched in the
// same PR are exercised under handler_coverage_2_test.go and
// era5_unification_test.go). This file backfills one focused test per
// untested handler.

// ---------------------------------------------------------------------------
// Galazeth Prismari — ETB creates a Treasure
// ---------------------------------------------------------------------------

func TestGalazethEra4_ETBCreatesTreasureToken(t *testing.T) {
	gs := newGame(t, 2)
	galazeth := addPerm(gs, 0, "Galazeth Prismari", "creature", "legendary")

	bfBefore := len(gs.Seats[0].Battlefield)
	galazethPrismariCustomETB(gs, galazeth)

	if len(gs.Seats[0].Battlefield) != bfBefore+1 {
		t.Fatalf("expected one Treasure token created; before=%d after=%d",
			bfBefore, len(gs.Seats[0].Battlefield))
	}
	tok := gs.Seats[0].Battlefield[len(gs.Seats[0].Battlefield)-1]
	if !cardHasType(tok.Card, "treasure") {
		t.Errorf("expected treasure token; types=%v", tok.Card.Types)
	}
}

// ---------------------------------------------------------------------------
// Silverquill, the Disputant — ETB drain 3 / gain 3
// ---------------------------------------------------------------------------

func TestSilverquillEra4_ETBDrainsHighestLifeOpponent(t *testing.T) {
	gs := newGame(t, 3)
	silverquill := addPerm(gs, 0, "Silverquill, the Disputant", "creature", "legendary")
	gs.Seats[0].Life = 30
	gs.Seats[1].Life = 35
	gs.Seats[2].Life = 40 // highest — should be the target

	silverquillETBDrain(gs, silverquill)

	if gs.Seats[0].Life != 33 {
		t.Errorf("expected controller to gain 3 (30 → 33); got %d", gs.Seats[0].Life)
	}
	if gs.Seats[2].Life != 37 {
		t.Errorf("expected highest-life opp to lose 3 (40 → 37); got %d", gs.Seats[2].Life)
	}
	if gs.Seats[1].Life != 35 {
		t.Errorf("non-targeted opp should be untouched; got %d", gs.Seats[1].Life)
	}
}

// ---------------------------------------------------------------------------
// Quandrix, the Proof — ETB doubles +1/+1 counters
// ---------------------------------------------------------------------------

func TestQuandrixEra4_ETBDoublesCounters(t *testing.T) {
	gs := newGame(t, 2)
	quandrix := addPerm(gs, 0, "Quandrix, the Proof", "creature", "legendary")
	a := addPerm(gs, 0, "Hydra A", "creature")
	a.AddCounter("+1/+1", 4)
	b := addPerm(gs, 0, "Hydra B", "creature")
	b.AddCounter("+1/+1", 1)
	noCounter := addPerm(gs, 0, "Llanowar Elves", "creature")

	quandrixTheProofETB(gs, quandrix)

	if a.Counters["+1/+1"] != 8 {
		t.Errorf("expected Hydra A 4→8; got %d", a.Counters["+1/+1"])
	}
	if b.Counters["+1/+1"] != 2 {
		t.Errorf("expected Hydra B 1→2; got %d", b.Counters["+1/+1"])
	}
	if noCounter.Counters["+1/+1"] != 0 {
		t.Errorf("creature with no counters should remain at 0; got %d", noCounter.Counters["+1/+1"])
	}
}

// ---------------------------------------------------------------------------
// Asmoranomardicadaistinaculdacar — ETB Cookbook tutor + sac-foods ping
// ---------------------------------------------------------------------------

func TestAsmoranEra4_ETBTutorsCookbook(t *testing.T) {
	gs := newGame(t, 2)
	asmoran := addPerm(gs, 0, "Asmoranomardicadaistinaculdacar", "creature", "legendary")
	addLibrary(gs, 0, "Tarmogoyf", "The Underworld Cookbook", "Lightning Bolt")

	asmoranETBCookbookTutor(gs, asmoran)

	found := false
	for _, c := range gs.Seats[0].Hand {
		if c.DisplayName() == "The Underworld Cookbook" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected The Underworld Cookbook in hand; hand=%v", handNames(gs.Seats[0].Hand))
	}
}

func TestAsmoranEra4_SacrificesTwoFoodsAndPingsBiggestThreat(t *testing.T) {
	gs := newGame(t, 2)
	asmoran := addPerm(gs, 0, "Asmoranomardicadaistinaculdacar", "creature", "legendary")
	food1 := addPerm(gs, 0, "Food Token", "artifact", "food", "token")
	food2 := addPerm(gs, 0, "Food Token", "artifact", "food", "token")
	threat := addPerm(gs, 1, "Hydra Beast", "creature")
	threat.Card.BasePower = 7
	threat.Card.BaseToughness = 7

	asmoranSacFoodsPing(gs, asmoran, 0, nil)

	if threat.MarkedDamage != 6 {
		t.Errorf("expected 6 marked damage on biggest threat; got %d", threat.MarkedDamage)
	}
	// Both foods should be sacrificed (no longer on battlefield).
	for _, p := range gs.Seats[0].Battlefield {
		if p == food1 || p == food2 {
			t.Errorf("food token still on battlefield after sacrifice")
		}
	}
}

func TestAsmoranEra4_FailsWithoutTwoFoods(t *testing.T) {
	gs := newGame(t, 2)
	asmoran := addPerm(gs, 0, "Asmoranomardicadaistinaculdacar", "creature", "legendary")
	addPerm(gs, 0, "Food Token", "artifact", "food", "token") // only one
	addPerm(gs, 1, "Bear", "creature")

	asmoranSacFoodsPing(gs, asmoran, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when fewer than 2 foods")
	}
}

// ---------------------------------------------------------------------------
// Lier, Disciple of the Drowned — uncounterable mark
// ---------------------------------------------------------------------------

func TestLierEra4_MarksOwnInstantUncounterable(t *testing.T) {
	gs := newGame(t, 2)
	lier := addPerm(gs, 0, "Lier, Disciple of the Drowned", "creature", "legendary")
	item := &gameengine.StackItem{
		Controller: 0,
		Card:       &gameengine.Card{Name: "Lightning Bolt", Owner: 0, Types: []string{"instant"}},
	}

	lierUncounterableMark(gs, lier, map[string]interface{}{
		"caster_seat": 0,
		"stack_item":  item,
	})

	if item.CostMeta == nil || item.CostMeta["cannot_be_countered"] != true {
		t.Fatalf("expected stack item to be marked uncounterable; CostMeta=%+v", item.CostMeta)
	}
}

func TestLierEra4_DoesNotMarkOpponentSpells(t *testing.T) {
	gs := newGame(t, 2)
	lier := addPerm(gs, 0, "Lier, Disciple of the Drowned", "creature", "legendary")
	item := &gameengine.StackItem{
		Controller: 1,
		Card:       &gameengine.Card{Name: "Counterspell", Owner: 1, Types: []string{"instant"}},
	}

	lierUncounterableMark(gs, lier, map[string]interface{}{
		"caster_seat": 1,
		"stack_item":  item,
	})

	if item.CostMeta != nil && item.CostMeta["cannot_be_countered"] == true {
		t.Errorf("opponent's spell should not be marked uncounterable")
	}
}

// ---------------------------------------------------------------------------
// Jadzi, Oracle of Arcavios — magecraft reveal + discard-bounce
// ---------------------------------------------------------------------------

func TestJadziEra4_MagecraftLandPlayedFreely(t *testing.T) {
	gs := newGame(t, 2)
	jadzi := addPerm(gs, 0, "Jadzi, Oracle of Arcavios", "creature", "legendary")
	land := &gameengine.Card{Name: "Forest", Owner: 0, Types: []string{"land", "basic"}}
	gs.Seats[0].Library = append(gs.Seats[0].Library, land)
	bfBefore := len(gs.Seats[0].Battlefield)

	jadziMagecraftReveal(gs, jadzi, map[string]interface{}{
		"caster_seat": 0,
	})

	if len(gs.Seats[0].Battlefield) != bfBefore+1 {
		t.Fatalf("expected revealed land to enter battlefield; before=%d after=%d",
			bfBefore, len(gs.Seats[0].Battlefield))
	}
	last := gs.Seats[0].Battlefield[len(gs.Seats[0].Battlefield)-1]
	if last.Card != land {
		t.Errorf("expected revealed land on battlefield; got %s", last.Card.DisplayName())
	}
}

func TestJadziEra4_MagecraftNonlandRoutedToHand(t *testing.T) {
	gs := newGame(t, 2)
	jadzi := addPerm(gs, 0, "Jadzi, Oracle of Arcavios", "creature", "legendary")
	bolt := &gameengine.Card{Name: "Lightning Bolt", Owner: 0, Types: []string{"instant"}}
	gs.Seats[0].Library = append(gs.Seats[0].Library, bolt)

	jadziMagecraftReveal(gs, jadzi, map[string]interface{}{
		"caster_seat": 0,
	})

	found := false
	for _, c := range gs.Seats[0].Hand {
		if c == bolt {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected Lightning Bolt to land in hand (alt-cast partial); hand=%v",
			handNames(gs.Seats[0].Hand))
	}
}

func TestJadziEra4_DiscardBouncesJadziToHand(t *testing.T) {
	gs := newGame(t, 2)
	jadzi := addPerm(gs, 0, "Jadzi, Oracle of Arcavios", "creature", "legendary")
	jadziCard := jadzi.Card
	junk := &gameengine.Card{Name: "Brainstorm", Owner: 0, Types: []string{"instant", "cmc:1"}}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, junk)

	jadziDiscardBounce(gs, jadzi, 0, nil)

	// Jadzi should be off the battlefield…
	for _, p := range gs.Seats[0].Battlefield {
		if p == jadzi {
			t.Errorf("Jadzi should be off battlefield after activation")
		}
	}
	// …and back in hand by name (the wrapping permanent is gone).
	foundInHand := false
	for _, c := range gs.Seats[0].Hand {
		if c == jadziCard {
			foundInHand = true
		}
	}
	if !foundInHand {
		t.Errorf("expected Jadzi card returned to hand")
	}
	// Junk should be discarded.
	if len(gs.Seats[0].Graveyard) != 1 || gs.Seats[0].Graveyard[0] != junk {
		t.Errorf("expected Brainstorm discarded to graveyard; gy=%v", gs.Seats[0].Graveyard)
	}
}

// ---------------------------------------------------------------------------
// Toxrill, the Corrosive — slime end step + remove-slime draw
// ---------------------------------------------------------------------------

func TestToxrillEra4_OpponentEndStepSlimesAndDestroys(t *testing.T) {
	gs := newGame(t, 2)
	toxrill := addPerm(gs, 0, "Toxrill, the Corrosive", "creature", "legendary")
	bear := addPerm(gs, 1, "Grizzly Bears", "creature")
	bear.Card.BasePower = 2
	bear.Card.BaseToughness = 2
	wolf := addPerm(gs, 1, "Werewolf", "creature")
	wolf.Card.BasePower = 3
	wolf.Card.BaseToughness = 3

	toxrillEndStepSlime(gs, toxrill, map[string]interface{}{
		"active_seat": 1,
	})

	// Both opponent creatures should be off the battlefield (destroyed).
	for _, p := range gs.Seats[1].Battlefield {
		if p == bear || p == wolf {
			t.Errorf("expected slime-counter destruction; %s still on battlefield",
				p.Card.DisplayName())
		}
	}
}

func TestToxrillEra4_OwnEndStepIsNoop(t *testing.T) {
	gs := newGame(t, 2)
	toxrill := addPerm(gs, 0, "Toxrill, the Corrosive", "creature", "legendary")
	bear := addPerm(gs, 1, "Grizzly Bears", "creature")

	// Active seat = Toxrill's controller — trigger should NOT fire.
	toxrillEndStepSlime(gs, toxrill, map[string]interface{}{
		"active_seat": 0,
	})

	if bear.Counters["slime"] > 0 {
		t.Errorf("opponent creature should not be slimed on Toxrill controller's end step")
	}
}

func TestToxrillEra4_RemoveSlimeAtSorcerySpeedDraws(t *testing.T) {
	gs := newGame(t, 2)
	toxrill := addPerm(gs, 0, "Toxrill, the Corrosive", "creature", "legendary")
	bear := addPerm(gs, 1, "Grizzly Bears", "creature")
	bear.AddCounter("slime", 2)
	addLibrary(gs, 0, "TopCard")

	gs.Active = 0
	gs.Phase = "precombat_main"

	toxrillRemoveSlimeDraw(gs, toxrill, 0, nil)

	if bear.Counters["slime"] != 1 {
		t.Errorf("expected one slime counter removed (2→1); got %d", bear.Counters["slime"])
	}
	if len(gs.Seats[0].Hand) != 1 {
		t.Errorf("expected 1 card drawn; hand=%d", len(gs.Seats[0].Hand))
	}
}

func TestToxrillEra4_RemoveSlimeFailsAtInstantSpeed(t *testing.T) {
	gs := newGame(t, 2)
	toxrill := addPerm(gs, 0, "Toxrill, the Corrosive", "creature", "legendary")
	bear := addPerm(gs, 1, "Grizzly Bears", "creature")
	bear.AddCounter("slime", 1)

	// Opponent's turn — sorcery-speed gate should reject.
	gs.Active = 1
	gs.Phase = "precombat_main"

	toxrillRemoveSlimeDraw(gs, toxrill, 0, nil)

	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed when activated outside sorcery speed")
	}
	if bear.Counters["slime"] != 1 {
		t.Errorf("slime counter should be untouched on failed activation; got %d",
			bear.Counters["slime"])
	}
}
