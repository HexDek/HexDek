package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// Smoke tests for the two bulk-pattern families added in
// dev/muninn-bulk-patterns-5: etb_basic_land_ramp_family and
// etb_drain_target_opponent_family. Focus is on the algorithm contract;
// full-engine integration is covered by goldilocks / tournament runs.

// ---------------------------------------------------------------------
// etb_basic_land_ramp_family — Farhaven Elf, Civic Wayfinder, Pilgrim's
// Eye, Borderland Ranger, Sylvan Ranger.
// ---------------------------------------------------------------------

func addBasicLandBP5(gs *gameengine.GameState, seat int, name string) *gameengine.Card {
	c := &gameengine.Card{
		Name:  name,
		Owner: seat,
		Types: []string{"basic", "land"},
	}
	gs.Seats[seat].Library = append(gs.Seats[seat].Library, c)
	return c
}

func addNonBasicBP5(gs *gameengine.GameState, seat int, name string) *gameengine.Card {
	c := &gameengine.Card{
		Name:  name,
		Owner: seat,
		Types: []string{"creature"},
	}
	gs.Seats[seat].Library = append(gs.Seats[seat].Library, c)
	return c
}

func bp5LibraryContains(gs *gameengine.GameState, seat int, want *gameengine.Card) bool {
	for _, c := range gs.Seats[seat].Library {
		if c == want {
			return true
		}
	}
	return false
}

func bp5HandContains(gs *gameengine.GameState, seat int, want *gameengine.Card) bool {
	for _, c := range gs.Seats[seat].Hand {
		if c == want {
			return true
		}
	}
	return false
}

func bp5BattlefieldContainsCard(gs *gameengine.GameState, seat int, want *gameengine.Card) *gameengine.Permanent {
	for _, p := range gs.Seats[seat].Battlefield {
		if p != nil && p.Card == want {
			return p
		}
	}
	return nil
}

func TestEtbBasicLandRamp_FarhavenElfPutsBasicTappedOntoBattlefield(t *testing.T) {
	gs := newGame(t, 2)
	addNonBasicBP5(gs, 0, "Wheel of Fortune")
	want := addBasicLandBP5(gs, 0, "Forest")
	addNonBasicBP5(gs, 0, "Brainstorm")

	farhaven := addPerm(gs, 0, "Farhaven Elf", "creature")
	gameengine.InvokeETBHook(gs, farhaven)

	perm := bp5BattlefieldContainsCard(gs, 0, want)
	if perm == nil {
		t.Fatalf("Forest should be on seat 0 battlefield; bf=%d", len(gs.Seats[0].Battlefield))
	}
	if !perm.Tapped {
		t.Errorf("Forest should enter tapped from Farhaven Elf")
	}
	if bp5LibraryContains(gs, 0, want) {
		t.Errorf("Forest should have been removed from library")
	}
}

func TestEtbBasicLandRamp_PilgrimsEyePutsBasicIntoHand(t *testing.T) {
	gs := newGame(t, 2)
	addNonBasicBP5(gs, 0, "Wheel of Fortune")
	want := addBasicLandBP5(gs, 0, "Island")

	pilgrim := addPerm(gs, 0, "Pilgrim's Eye", "artifact", "creature")
	gameengine.InvokeETBHook(gs, pilgrim)

	if !bp5HandContains(gs, 0, want) {
		t.Errorf("Pilgrim's Eye should put Island into hand; hand=%v",
			cardNamesBP4(gs.Seats[0].Hand))
	}
	if bp5LibraryContains(gs, 0, want) {
		t.Errorf("Island should have been removed from library")
	}
}

func TestEtbBasicLandRamp_CivicWayfinderTakesFirstBasic(t *testing.T) {
	gs := newGame(t, 2)
	first := addBasicLandBP5(gs, 0, "Plains")
	addBasicLandBP5(gs, 0, "Mountain") // not chosen (first-match)

	wayfinder := addPerm(gs, 0, "Civic Wayfinder", "creature")
	gameengine.InvokeETBHook(gs, wayfinder)

	if !bp5HandContains(gs, 0, first) {
		t.Errorf("Civic Wayfinder should take the first basic (Plains); hand=%v",
			cardNamesBP4(gs.Seats[0].Hand))
	}
}

func TestEtbBasicLandRamp_NoBasicEmitsFailEvent(t *testing.T) {
	gs := newGame(t, 2)
	addNonBasicBP5(gs, 0, "Lightning Bolt")
	addNonBasicBP5(gs, 0, "Sol Ring")

	ranger := addPerm(gs, 0, "Sylvan Ranger", "creature")
	handBefore := len(gs.Seats[0].Hand)
	gameengine.InvokeETBHook(gs, ranger)

	if len(gs.Seats[0].Hand) != handBefore {
		t.Errorf("Sylvan Ranger whiff should not add to hand")
	}
	if hasEvent(gs, "per_card_failed") < 1 {
		t.Errorf("expected per_card_failed event on whiff")
	}
}

func TestEtbBasicLandRamp_BorderlandRangerSiblingShape(t *testing.T) {
	gs := newGame(t, 2)
	want := addBasicLandBP5(gs, 0, "Forest")

	br := addPerm(gs, 0, "Borderland Ranger", "creature")
	gameengine.InvokeETBHook(gs, br)

	if !bp5HandContains(gs, 0, want) {
		t.Errorf("Borderland Ranger should put Forest into hand")
	}
}

// ---------------------------------------------------------------------
// etb_drain_target_opponent_family — Skymarch Bloodletter, Vampire
// Sovereign, Highway Robber, Dakmor Ghoul, Bloodborn Scoundrels.
// ---------------------------------------------------------------------

func TestEtbDrain_SkymarchBloodletterDrainsOneFromLowestLife(t *testing.T) {
	gs := newGame(t, 3)
	gs.Seats[0].Life = 20
	gs.Seats[1].Life = 30
	gs.Seats[2].Life = 10 // lowest-life opponent — gets targeted

	skymarch := addPerm(gs, 0, "Skymarch Bloodletter", "creature")
	gameengine.InvokeETBHook(gs, skymarch)

	if gs.Seats[2].Life != 9 {
		t.Errorf("seat 2 (lowest life) should lose 1; life=%d", gs.Seats[2].Life)
	}
	if gs.Seats[1].Life != 30 {
		t.Errorf("seat 1 (highest life) should be untouched; life=%d", gs.Seats[1].Life)
	}
	if gs.Seats[0].Life != 21 {
		t.Errorf("seat 0 (controller) should gain 1; life=%d", gs.Seats[0].Life)
	}
}

func TestEtbDrain_VampireSovereignDrainsThree(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].Life = 20
	gs.Seats[1].Life = 20

	sovereign := addPerm(gs, 0, "Vampire Sovereign", "creature")
	gameengine.InvokeETBHook(gs, sovereign)

	if gs.Seats[1].Life != 17 {
		t.Errorf("seat 1 should lose 3; life=%d", gs.Seats[1].Life)
	}
	if gs.Seats[0].Life != 23 {
		t.Errorf("seat 0 should gain 3; life=%d", gs.Seats[0].Life)
	}
}

func TestEtbDrain_HighwayRobberAndDakmorGhoulShareShape(t *testing.T) {
	for _, name := range []string{"Highway Robber", "Dakmor Ghoul", "Bloodborn Scoundrels"} {
		gs := newGame(t, 2)
		gs.Seats[0].Life = 20
		gs.Seats[1].Life = 20

		p := addPerm(gs, 0, name, "creature")
		gameengine.InvokeETBHook(gs, p)

		if gs.Seats[1].Life != 18 {
			t.Errorf("%s: seat 1 should lose 2; life=%d", name, gs.Seats[1].Life)
		}
		if gs.Seats[0].Life != 22 {
			t.Errorf("%s: seat 0 should gain 2; life=%d", name, gs.Seats[0].Life)
		}
	}
}

func TestEtbDrain_SkipsDeadOpponent(t *testing.T) {
	gs := newGame(t, 3)
	gs.Seats[0].Life = 20
	gs.Seats[1].Life = 1 // alive but lowest
	gs.Seats[2].Life = 5

	// Mark seat 1 as Lost — drain should skip them and hit seat 2.
	gs.Seats[1].Lost = true

	highway := addPerm(gs, 0, "Highway Robber", "creature")
	gameengine.InvokeETBHook(gs, highway)

	if gs.Seats[1].Life != 1 {
		t.Errorf("dead seat 1 should be untouched; life=%d", gs.Seats[1].Life)
	}
	if gs.Seats[2].Life != 3 {
		t.Errorf("seat 2 should lose 2; life=%d", gs.Seats[2].Life)
	}
}
