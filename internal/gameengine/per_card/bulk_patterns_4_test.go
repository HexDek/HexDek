package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// Smoke tests for the two bulk-pattern families added in
// dev/muninn-bulk-patterns-4: shuffle_self_from_grave_family and
// etb_library_tutor_family. Focus is on the algorithm contract, not
// the engine integration paths (those are exercised by goldilocks).

// ---------------------------------------------------------------------
// shuffle_self_from_grave_family — Purity / Vigor.
// ---------------------------------------------------------------------

func TestShuffleSelfFromGrave_PurityDyingMovesToLibrary(t *testing.T) {
	gs := newGame(t, 2)
	purity := addPerm(gs, 0, "Purity", "creature")
	// creature_dies fires post-zone-move, so the card is in graveyard
	// when the trigger lands.
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, purity.Card)

	gameengine.FireCardTrigger(gs, "creature_dies", map[string]interface{}{
		"perm":            purity,
		"card":            purity.Card,
		"controller_seat": 0,
	})

	inLibrary := false
	for _, c := range gs.Seats[0].Library {
		if c == purity.Card {
			inLibrary = true
			break
		}
	}
	if !inLibrary {
		t.Errorf("Purity should be in seat 0's library after creature_dies; library=%v",
			cardNamesBP4(gs.Seats[0].Library))
	}
}

func TestShuffleSelfFromGrave_VigorDyingMovesToLibrary(t *testing.T) {
	gs := newGame(t, 2)
	vigor := addPerm(gs, 0, "Vigor", "creature")
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, vigor.Card)

	gameengine.FireCardTrigger(gs, "creature_dies", map[string]interface{}{
		"perm":            vigor,
		"card":            vigor.Card,
		"controller_seat": 0,
	})

	inLibrary := false
	for _, c := range gs.Seats[0].Library {
		if c == vigor.Card {
			inLibrary = true
			break
		}
	}
	if !inLibrary {
		t.Errorf("Vigor should be in seat 0's library after creature_dies")
	}
}

// ---------------------------------------------------------------------
// etb_library_tutor_family — Stoneforge / Trophy / Treasure / Recruiter.
// ---------------------------------------------------------------------

// addLibraryCMCbp4 adds a card with explicit types and CMC.
func addLibraryCMCbp4(gs *gameengine.GameState, seat int, name string, cmc int, types ...string) *gameengine.Card {
	c := &gameengine.Card{
		Name:  name,
		Owner: seat,
		Types: append([]string{}, types...),
		CMC:   cmc,
	}
	gs.Seats[seat].Library = append(gs.Seats[seat].Library, c)
	return c
}

// addLibraryPowerBP4 adds a creature card with explicit BasePower.
func addLibraryPowerBP4(gs *gameengine.GameState, seat int, name string, power int, types ...string) *gameengine.Card {
	c := &gameengine.Card{
		Name:      name,
		Owner:     seat,
		Types:     append([]string{}, types...),
		BasePower: power,
	}
	gs.Seats[seat].Library = append(gs.Seats[seat].Library, c)
	return c
}

func TestEtbLibraryTutor_StoneforgeFetchesEquipment(t *testing.T) {
	gs := newGame(t, 2)
	addLibraryCMCbp4(gs, 0, "Sol Ring", 1, "artifact")
	wantEq := addLibraryCMCbp4(gs, 0, "Skullclamp", 1, "artifact", "equipment")

	stoneforge := addPerm(gs, 0, "Stoneforge Mystic", "creature")
	gameengine.InvokeETBHook(gs, stoneforge)

	inHand := false
	for _, c := range gs.Seats[0].Hand {
		if c == wantEq {
			inHand = true
			break
		}
	}
	if !inHand {
		t.Errorf("Stoneforge Mystic should have tutored Skullclamp; hand=%v",
			cardNamesBP4(gs.Seats[0].Hand))
	}
}

func TestEtbLibraryTutor_TrophyMageRequiresCMC3(t *testing.T) {
	gs := newGame(t, 2)
	addLibraryCMCbp4(gs, 0, "Sol Ring", 1, "artifact")
	wantCmc3 := addLibraryCMCbp4(gs, 0, "Basalt Monolith", 3, "artifact")

	trophy := addPerm(gs, 0, "Trophy Mage", "creature")
	gameengine.InvokeETBHook(gs, trophy)

	if !handContainsBP4(gs, 0, wantCmc3) {
		t.Errorf("Trophy Mage should tutor the CMC-3 artifact; hand=%v",
			cardNamesBP4(gs.Seats[0].Hand))
	}
}

func TestEtbLibraryTutor_TreasureMageRequiresCMC6Plus(t *testing.T) {
	gs := newGame(t, 2)
	addLibraryCMCbp4(gs, 0, "Sol Ring", 1, "artifact")
	addLibraryCMCbp4(gs, 0, "Basalt Monolith", 3, "artifact")
	wantCmc6 := addLibraryCMCbp4(gs, 0, "Wurmcoil Engine", 6, "artifact")

	treasure := addPerm(gs, 0, "Treasure Mage", "creature")
	gameengine.InvokeETBHook(gs, treasure)

	if !handContainsBP4(gs, 0, wantCmc6) {
		t.Errorf("Treasure Mage should tutor the CMC-6 artifact; hand=%v",
			cardNamesBP4(gs.Seats[0].Hand))
	}
}

func TestEtbLibraryTutor_ImperialRecruiterFiltersByPower(t *testing.T) {
	gs := newGame(t, 2)
	wantP2 := addLibraryPowerBP4(gs, 0, "Goblin Welder", 1, "creature")
	addLibraryPowerBP4(gs, 0, "Serra Angel", 4, "creature")

	recruiter := addPerm(gs, 0, "Imperial Recruiter", "creature")
	gameengine.InvokeETBHook(gs, recruiter)

	if !handContainsBP4(gs, 0, wantP2) {
		t.Errorf("Imperial Recruiter should tutor the power-1 creature; hand=%v",
			cardNamesBP4(gs.Seats[0].Hand))
	}
}

func TestEtbLibraryTutor_NoMatchEmitsFailEvent(t *testing.T) {
	gs := newGame(t, 2)
	// Spellseeker needs an instant or sorcery with CMC <= 2. Library
	// has only a creature — no match.
	addLibraryCMCbp4(gs, 0, "Grizzly Bears", 2, "creature")

	spellseeker := addPerm(gs, 0, "Spellseeker", "creature")
	handBefore := len(gs.Seats[0].Hand)

	gameengine.InvokeETBHook(gs, spellseeker)

	if len(gs.Seats[0].Hand) != handBefore {
		t.Errorf("Spellseeker whiff should not put a card in hand; before=%d after=%d",
			handBefore, len(gs.Seats[0].Hand))
	}
	if hasEvent(gs, "per_card_failed") == 0 {
		t.Errorf("Spellseeker whiff should log per_card_failed event")
	}
}

func handContainsBP4(gs *gameengine.GameState, seat int, want *gameengine.Card) bool {
	for _, c := range gs.Seats[seat].Hand {
		if c == want {
			return true
		}
	}
	return false
}

func cardNamesBP4(cards []*gameengine.Card) []string {
	out := make([]string, 0, len(cards))
	for _, c := range cards {
		if c != nil {
			out = append(out, c.Name)
		}
	}
	return out
}
