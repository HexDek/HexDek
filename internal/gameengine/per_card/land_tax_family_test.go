package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// Tests for the land-tax family generic handler covering Loyal Warhound,
// Sand Scout, and Aerial Surveyor. The gate (opponent controls more lands)
// + the fetch (first matching land in library, onto battlefield tapped)
// + the shuffle are all exercised here.

func TestLoyalWarhound_ETBFetchesPlainsWhenBehindOnLands(t *testing.T) {
	gs := newGame(t, 2)
	// Opponent has 2 lands, I have 0 — gate opens.
	addPerm(gs, 1, "Plains", "basic", "land", "plains")
	addPerm(gs, 1, "Forest", "basic", "land", "forest")
	// Library: a basic Plains as the second card so we know the filter
	// skipped a non-matching card.
	addLibraryWithTypes(gs, 0, "Mountain", []string{"basic", "land", "mountain"})
	addLibraryWithTypes(gs, 0, "Plains", []string{"basic", "land", "plains"})

	hound := addPerm(gs, 0, "Loyal Warhound", "creature")
	gameengine.InvokeETBHook(gs, hound)

	// Plains should now be on seat 0's battlefield tapped.
	foundPlains := false
	for _, p := range gs.Seats[0].Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if p.Card.DisplayName() == "Plains" {
			if !p.Tapped {
				t.Errorf("Plains should ETB tapped from Loyal Warhound; got untapped")
			}
			foundPlains = true
		}
	}
	if !foundPlains {
		t.Errorf("expected Plains on battlefield after Loyal Warhound ETB")
	}
}

func TestLoyalWarhound_NoFetchWhenAheadOnLands(t *testing.T) {
	gs := newGame(t, 2)
	// I have 2 lands, opponent has 0 — gate closed.
	addPerm(gs, 0, "Plains", "basic", "land", "plains")
	addPerm(gs, 0, "Plains", "basic", "land", "plains")
	addLibraryWithTypes(gs, 0, "Plains", []string{"basic", "land", "plains"})

	hound := addPerm(gs, 0, "Loyal Warhound", "creature")
	libBefore := len(gs.Seats[0].Library)
	gameengine.InvokeETBHook(gs, hound)

	if len(gs.Seats[0].Library) != libBefore {
		t.Errorf("library should be untouched when gate is closed; before=%d after=%d",
			libBefore, len(gs.Seats[0].Library))
	}
}

func TestSandScout_FetchesDesertNotPlains(t *testing.T) {
	gs := newGame(t, 2)
	addPerm(gs, 1, "Plains", "basic", "land", "plains")
	addPerm(gs, 1, "Plains", "basic", "land", "plains")
	// Library has both a Plains and a Desert — Sand Scout must pick Desert.
	addLibraryWithTypes(gs, 0, "Plains", []string{"basic", "land", "plains"})
	addLibraryWithTypes(gs, 0, "Desert", []string{"land", "desert"})

	scout := addPerm(gs, 0, "Sand Scout", "creature")
	gameengine.InvokeETBHook(gs, scout)

	gotDesert := false
	gotPlains := false
	for _, p := range gs.Seats[0].Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		switch p.Card.DisplayName() {
		case "Desert":
			gotDesert = true
		case "Plains":
			gotPlains = true
		}
	}
	if !gotDesert {
		t.Errorf("Sand Scout should fetch a Desert")
	}
	if gotPlains {
		t.Errorf("Sand Scout should NOT pick the Plains over the Desert")
	}
}

func TestAerialSurveyor_AttackTriggerFetchesPlains(t *testing.T) {
	gs := newGame(t, 2)
	// Defender (seat 1) has 3 lands; I have 1 → gate open.
	addPerm(gs, 0, "Plains", "basic", "land", "plains")
	addPerm(gs, 1, "Plains", "basic", "land", "plains")
	addPerm(gs, 1, "Plains", "basic", "land", "plains")
	addPerm(gs, 1, "Plains", "basic", "land", "plains")
	addLibraryWithTypes(gs, 0, "Plains", []string{"basic", "land", "plains"})

	surveyor := addPerm(gs, 0, "Aerial Surveyor", "artifact", "creature")
	gameengine.FireCardTrigger(gs, "attacks", map[string]interface{}{
		"attacker_perm": surveyor,
		"defender_seat": 1,
	})

	gotPlains := 0
	for _, p := range gs.Seats[0].Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if p.Card.DisplayName() == "Plains" {
			gotPlains++
		}
	}
	if gotPlains < 2 {
		t.Errorf("Aerial Surveyor should fetch a second Plains via attack trigger; total Plains on bf=%d", gotPlains)
	}
}

// addLibraryWithTypes is the typed analogue of addLibrary — needed so
// the land-fetch helper's cardHasType / cardHasSubtype checks match.
func addLibraryWithTypes(gs *gameengine.GameState, seat int, name string, types []string) {
	c := &gameengine.Card{
		Name:  name,
		Owner: seat,
		Types: append([]string{}, types...),
	}
	gs.Seats[seat].Library = append(gs.Seats[seat].Library, c)
}
