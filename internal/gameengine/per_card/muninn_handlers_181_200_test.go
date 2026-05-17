package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// Smoke tests for the dev/muninn-handlers-181-200 wave.

func TestBurnishedHart_ActivateSacsAndFetchesTwoBasics(t *testing.T) {
	gs := newGame(t, 2)
	hart := addPerm(gs, 0, "Burnished Hart", "creature", "artifact")
	addLibraryWithTypes(gs, 0, "Plains", []string{"basic", "land", "plains"})
	addLibraryWithTypes(gs, 0, "Forest", []string{"basic", "land", "forest"})
	addLibrary(gs, 0, "Lightning Bolt")
	addLibraryWithTypes(gs, 0, "Mountain", []string{"basic", "land", "mountain"})

	gameengine.InvokeActivatedHook(gs, hart, 0, nil)

	// Hart is sacrificed → graveyard.
	for _, p := range gs.Seats[0].Battlefield {
		if p == hart {
			t.Errorf("expected Burnished Hart to be off the battlefield after activation")
		}
	}
	// Two basics on the battlefield, both tapped.
	basics := 0
	for _, p := range gs.Seats[0].Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if cardHasType(p.Card, "basic") && cardHasType(p.Card, "land") {
			basics++
			if !p.Tapped {
				t.Errorf("expected fetched basic %q to enter tapped", p.Card.DisplayName())
			}
		}
	}
	if basics != 2 {
		t.Errorf("expected 2 basics on battlefield, got %d", basics)
	}
}

func TestBurnishedHart_NoBasicsInLibraryNoOpFetch(t *testing.T) {
	gs := newGame(t, 2)
	hart := addPerm(gs, 0, "Burnished Hart", "creature", "artifact")
	// Library has only non-basics.
	addLibrary(gs, 0, "Lightning Bolt", "Counterspell")

	gameengine.InvokeActivatedHook(gs, hart, 0, nil)

	// Hart still sacrificed even when fetch finds nothing.
	for _, p := range gs.Seats[0].Battlefield {
		if p == hart {
			t.Errorf("expected Burnished Hart to be sacrificed even without basics")
		}
	}
	for _, p := range gs.Seats[0].Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if cardHasType(p.Card, "basic") {
			t.Errorf("did not expect a basic on battlefield, got %q", p.Card.DisplayName())
		}
	}
}

func TestTrostaniSelesnyasVoice_OtherCreatureETBGainsLifeEqualToToughness(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].Life = 40
	trostani := addPerm(gs, 0, "Trostani, Selesnya's Voice", "creature", "legendary")
	trostani.Card.BasePower = 1
	trostani.Card.BaseToughness = 4

	other := addPerm(gs, 0, "Centaur Healer", "creature")
	other.Card.BasePower = 3
	other.Card.BaseToughness = 3

	gameengine.FireCardTrigger(gs, "permanent_etb", map[string]interface{}{
		"controller_seat": 0,
		"perm":            other,
	})

	if gs.Seats[0].Life != 43 {
		t.Errorf("expected life 43 after Centaur Healer enters (gain 3), got %d", gs.Seats[0].Life)
	}
}

func TestTrostaniSelesnyasVoice_SelfETBDoesNotTrigger(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].Life = 40
	trostani := addPerm(gs, 0, "Trostani, Selesnya's Voice", "creature", "legendary")
	trostani.Card.BasePower = 1
	trostani.Card.BaseToughness = 4

	gameengine.FireCardTrigger(gs, "permanent_etb", map[string]interface{}{
		"controller_seat": 0,
		"perm":            trostani,
	})

	if gs.Seats[0].Life != 40 {
		t.Errorf("self ETB should not trigger; life=%d (want 40)", gs.Seats[0].Life)
	}
}

func TestTrostaniSelesnyasVoice_OpponentCreatureETBDoesNotTrigger(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].Life = 40
	trostani := addPerm(gs, 0, "Trostani, Selesnya's Voice", "creature", "legendary")
	trostani.Card.BasePower = 1
	trostani.Card.BaseToughness = 4

	other := addPerm(gs, 1, "Centaur Healer", "creature")
	other.Card.BasePower = 3
	other.Card.BaseToughness = 3

	gameengine.FireCardTrigger(gs, "permanent_etb", map[string]interface{}{
		"controller_seat": 1,
		"perm":            other,
	})

	if gs.Seats[0].Life != 40 {
		t.Errorf("opponent ETB should not trigger Trostani; life=%d (want 40)", gs.Seats[0].Life)
	}
}

func TestTrostaniSelesnyasVoice_PopulateCopiesBestCreatureToken(t *testing.T) {
	gs := newGame(t, 2)
	trostani := addPerm(gs, 0, "Trostani, Selesnya's Voice", "creature", "legendary")
	// Survive SBAs through the ETB-relifegain trigger the populated
	// token will queue.
	trostani.Card.BasePower = 2
	trostani.Card.BaseToughness = 5

	// Two token creatures: a 1/1 Soldier and a 4/4 Centaur. Populate
	// should pick the Centaur (higher power*2 + toughness score).
	weak := addPerm(gs, 0, "Soldier Token", "creature", "token", "soldier")
	weak.Card.BasePower = 1
	weak.Card.BaseToughness = 1
	strong := addPerm(gs, 0, "Centaur Token", "creature", "token", "centaur")
	strong.Card.BasePower = 4
	strong.Card.BaseToughness = 4

	gameengine.InvokeActivatedHook(gs, trostani, 0, nil)

	centaurs := 0
	for _, p := range gs.Seats[0].Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if p.Card.BasePower == 4 && p.Card.BaseToughness == 4 && cardHasType(p.Card, "token") && cardHasType(p.Card, "centaur") {
			centaurs++
		}
	}
	if centaurs != 2 {
		t.Errorf("expected 2 Centaur 4/4 tokens after populate (original + copy), got %d", centaurs)
	}
}

func TestTrostaniSelesnyasVoice_PopulateNoCreatureTokenNoOp(t *testing.T) {
	gs := newGame(t, 2)
	trostani := addPerm(gs, 0, "Trostani, Selesnya's Voice", "creature", "legendary")
	trostani.Card.BasePower = 2
	trostani.Card.BaseToughness = 5

	// Only a non-token creature.
	addPerm(gs, 0, "Llanowar Elves", "creature")

	before := len(gs.Seats[0].Battlefield)
	gameengine.InvokeActivatedHook(gs, trostani, 0, nil)
	after := len(gs.Seats[0].Battlefield)

	if after != before {
		t.Errorf("populate with no creature token should be a no-op, before=%d after=%d", before, after)
	}
}
