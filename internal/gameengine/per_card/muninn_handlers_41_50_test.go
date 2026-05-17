package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// Smoke tests for the dev/muninn-handlers-41-50 wave. Each test
// exercises one handler's primary clause and verifies the canonical
// per_card_handler event fires with the expected effect.

func TestCrownOfGondor_LegendaryETBMakesMonarch(t *testing.T) {
	gs := newGame(t, 2)
	crown := addPerm(gs, 0, "Crown of Gondor", "artifact", "equipment", "legendary")

	leg := addPerm(gs, 0, "Aragorn, the Uniter", "creature", "legendary")
	gameengine.FirePermanentETBTriggers(gs, leg)

	if gs.Flags["has_monarch"] != 1 {
		t.Fatalf("expected has_monarch=1 after legendary ETB under Crown, got %d",
			gs.Flags["has_monarch"])
	}
	if gs.Flags["monarch_seat"] != 0 {
		t.Errorf("expected monarch_seat=0, got %d", gs.Flags["monarch_seat"])
	}
	_ = crown
}

func TestCrownOfGondor_NoTriggerWhenMonarchExists(t *testing.T) {
	gs := newGame(t, 2)
	_ = addPerm(gs, 0, "Crown of Gondor", "artifact", "equipment", "legendary")
	gs.Flags = map[string]int{"has_monarch": 1, "monarch_seat": 1}

	leg := addPerm(gs, 0, "Aragorn, the Uniter", "creature", "legendary")
	gameengine.FirePermanentETBTriggers(gs, leg)

	if gs.Flags["monarch_seat"] != 1 {
		t.Errorf("monarch should remain seat 1 (already exists), got %d",
			gs.Flags["monarch_seat"])
	}
}

func TestSepticRats_PumpsWhenDefenderPoisoned(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[1].PoisonCounters = 3
	rats := addPerm(gs, 0, "Septic Rats", "creature")
	rats.Card.BasePower = 2
	rats.Card.BaseToughness = 2

	gameengine.FireCardTrigger(gs, "creature_attacks", map[string]interface{}{
		"attacker_perm": rats,
		"defender_seat": 1,
	})
	if rats.Flags["temp_power"] != 1 || rats.Flags["temp_toughness"] != 1 {
		t.Errorf("expected +1/+1 temp pump, got power=%d toughness=%d",
			rats.Flags["temp_power"], rats.Flags["temp_toughness"])
	}
}

func TestSepticRats_NoPumpWhenDefenderClean(t *testing.T) {
	gs := newGame(t, 2)
	rats := addPerm(gs, 0, "Septic Rats", "creature")
	gameengine.FireCardTrigger(gs, "creature_attacks", map[string]interface{}{
		"attacker_perm": rats,
		"defender_seat": 1,
	})
	if rats.Flags["temp_power"] != 0 {
		t.Errorf("expected no pump (clean defender), got temp_power=%d",
			rats.Flags["temp_power"])
	}
}

func TestCourierBat_ReturnsCreatureWhenLifeGained(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].Turn.LifeGained = 3
	gravecreature := &gameengine.Card{
		Name: "Birds of Paradise", Owner: 0,
		Types: []string{"creature", "cmc:1"},
	}
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, gravecreature)

	bat := addPerm(gs, 0, "Courier Bat", "creature")
	gameengine.InvokeETBHook(gs, bat)

	if len(gs.Seats[0].Graveyard) != 0 {
		t.Errorf("expected creature to leave graveyard, got %d remaining",
			len(gs.Seats[0].Graveyard))
	}
	inHand := false
	for _, c := range gs.Seats[0].Hand {
		if c == gravecreature {
			inHand = true
		}
	}
	if !inHand {
		t.Errorf("expected returned creature in hand")
	}
}

func TestCourierBat_NoReturnWithoutLifeGain(t *testing.T) {
	gs := newGame(t, 2)
	gravecreature := &gameengine.Card{
		Name: "Birds of Paradise", Owner: 0,
		Types: []string{"creature", "cmc:1"},
	}
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, gravecreature)

	bat := addPerm(gs, 0, "Courier Bat", "creature")
	gameengine.InvokeETBHook(gs, bat)

	if len(gs.Seats[0].Graveyard) != 1 {
		t.Errorf("expected creature to remain in graveyard (no life gain)")
	}
}

func TestElanorGardner_ETBFoodToken(t *testing.T) {
	gs := newGame(t, 2)
	elanor := addPerm(gs, 0, "Elanor Gardner", "creature", "legendary")
	gameengine.InvokeETBHook(gs, elanor)

	foundFood := false
	for _, p := range gs.Seats[0].Battlefield {
		if p != nil && p.Card != nil && cardHasType(p.Card, "food") {
			foundFood = true
		}
	}
	if !foundFood {
		t.Errorf("expected a Food token on battlefield after Elanor ETB")
	}
}

func TestElanorGardner_EndStepLandTutorAfterFoodSac(t *testing.T) {
	gs := newGame(t, 2)
	elanor := addPerm(gs, 0, "Elanor Gardner", "creature", "legendary")
	// Stats so SBAs don't destroy Elanor as a 0/0 mid-trigger.
	elanor.Card.BasePower = 2
	elanor.Card.BaseToughness = 1
	addLibraryWithTypes(gs, 0, "Plains", []string{"basic", "land", "plains"})
	addLibraryWithTypes(gs, 0, "Mountain", []string{"basic", "land", "mountain"})

	// Stamp the food-sac flag for this turn.
	gameengine.FireCardTrigger(gs, "food_sacrificed", map[string]interface{}{
		"controller_seat": 0,
	})
	if elanor.Flags["elanor_food_sacced_turn"] != gs.Turn+1 {
		t.Fatalf("food_sacrificed should have stamped elanor's flag")
	}

	gameengine.FireCardTrigger(gs, "end_step", map[string]interface{}{
		"active_seat": 0,
	})

	landOnBF := false
	for _, p := range gs.Seats[0].Battlefield {
		if p != nil && p.Card != nil && cardHasType(p.Card, "basic") {
			landOnBF = true
		}
	}
	if !landOnBF {
		t.Errorf("expected a basic land tutored to battlefield")
	}
}

func TestStarlitSoothsayer_SurveilsOnLifeChange(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].Turn.LifeGained = 1
	addLibrary(gs, 0, "Top1")
	sooth := addPerm(gs, 0, "Starlit Soothsayer", "creature")

	gameengine.FireCardTrigger(gs, "end_step", map[string]interface{}{
		"active_seat": 0,
	})
	_ = sooth
	// Surveil disposes the top card either to graveyard or keeps it on
	// top; either way, the per_card_handler event for surveil should
	// have logged.
	want := "starlit_soothsayer_end_step_surveil"
	found := false
	for _, ev := range gs.EventLog {
		if ev.Kind == "per_card_handler" {
			if d, ok := ev.Details["slug"].(string); ok && d == want {
				if triggered, _ := ev.Details["triggered"].(bool); triggered {
					found = true
				}
			}
		}
	}
	if !found {
		t.Errorf("expected starlit_soothsayer triggered event")
	}
}

func TestWildPair_FindsMatchingPT(t *testing.T) {
	gs := newGame(t, 2)
	_ = addPerm(gs, 0, "Wild Pair", "enchantment")

	// Library has a 2/2 creature.
	twoTwo := &gameengine.Card{
		Name: "Grizzly Bears", Owner: 0,
		Types:         []string{"creature", "cmc:2"},
		BasePower:     2,
		BaseToughness: 2,
	}
	gs.Seats[0].Library = append(gs.Seats[0].Library, twoTwo)

	// A 1/3 enters (total 4) — should NOT match (Bears total = 4 actually).
	// Use a 3/1 enters (total 4) to match the 2/2 (also total 4).
	entering := addPerm(gs, 0, "Falkenrath Reaver", "creature")
	entering.Card.BasePower = 3
	entering.Card.BaseToughness = 1
	entering.Flags["was_cast"] = 1
	entering.Flags["cast_from_hand"] = 1

	gameengine.FirePermanentETBTriggers(gs, entering)

	onBF := false
	for _, p := range gs.Seats[0].Battlefield {
		if p != nil && p.Card == twoTwo {
			onBF = true
		}
	}
	if !onBF {
		t.Errorf("expected Grizzly Bears (P+T=4) to be tutored by Wild Pair for 3/1 ETB (P+T=4)")
	}
}

func TestWildPair_NoTriggerWhenNotCastFromHand(t *testing.T) {
	gs := newGame(t, 2)
	_ = addPerm(gs, 0, "Wild Pair", "enchantment")
	twoTwo := &gameengine.Card{
		Name: "Grizzly Bears", Owner: 0,
		Types:         []string{"creature", "cmc:2"},
		BasePower:     2,
		BaseToughness: 2,
	}
	gs.Seats[0].Library = append(gs.Seats[0].Library, twoTwo)

	// Token / reanimated / blink — no cast_from_hand flag.
	entering := addPerm(gs, 0, "Falkenrath Reaver", "creature")
	entering.Card.BasePower = 3
	entering.Card.BaseToughness = 1

	gameengine.FirePermanentETBTriggers(gs, entering)

	for _, p := range gs.Seats[0].Battlefield {
		if p != nil && p.Card == twoTwo {
			t.Errorf("Wild Pair should NOT fire when entering wasn't cast from hand")
		}
	}
}

func TestIngeniousProdigy_EntersWithXCounters(t *testing.T) {
	gs := newGame(t, 2)
	prod := addPerm(gs, 0, "Ingenious Prodigy", "creature")
	prod.Flags["x_paid"] = 4
	gameengine.InvokeETBHook(gs, prod)
	if prod.Counters["+1/+1"] != 4 {
		t.Errorf("expected 4 +1/+1 counters, got %d", prod.Counters["+1/+1"])
	}
}

func TestIngeniousProdigy_UpkeepRemovesCounterDrawsCard(t *testing.T) {
	gs := newGame(t, 2)
	addLibrary(gs, 0, "TopOfDeck")
	prod := addPerm(gs, 0, "Ingenious Prodigy", "creature")
	prod.AddCounter("+1/+1", 3)

	gameengine.FireCardTrigger(gs, "upkeep", map[string]interface{}{
		"active_seat": 0,
	})

	if prod.Counters["+1/+1"] != 2 {
		t.Errorf("expected 2 counters after removal, got %d", prod.Counters["+1/+1"])
	}
	if len(gs.Seats[0].Hand) != 1 {
		t.Errorf("expected 1 card in hand after draw, got %d", len(gs.Seats[0].Hand))
	}
}
