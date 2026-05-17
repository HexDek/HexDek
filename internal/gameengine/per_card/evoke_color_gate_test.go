package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// Tests for the evoke-hybrid color-gate family covering Wistfulness and
// Deceit. Both halves fire when the spell was cast (Vibrance convention,
// since hybrid-pip tracking isn't yet modeled). Both halves are no-ops
// when the permanent entered via a non-cast path.

func TestWistfulness_CastFiresBothColorModes(t *testing.T) {
	gs := newGame(t, 2)

	// Opponent battlefield: an artifact + an enchantment for the G mode
	// to exile. Higher-CMC target wins the picker.
	artifact := &gameengine.Card{
		Name: "Sol Ring", Owner: 1,
		Types: []string{"artifact", "cmc:1"},
	}
	artifactPerm := &gameengine.Permanent{
		Card: artifact, Controller: 1, Owner: 1,
		Timestamp: gs.NextTimestamp(),
		Counters:  map[string]int{}, Flags: map[string]int{},
	}
	gs.Seats[1].Battlefield = append(gs.Seats[1].Battlefield, artifactPerm)

	enchantment := &gameengine.Card{
		Name: "Rhystic Study", Owner: 1,
		Types: []string{"enchantment", "cmc:3"},
	}
	enchantPerm := &gameengine.Permanent{
		Card: enchantment, Controller: 1, Owner: 1,
		Timestamp: gs.NextTimestamp(),
		Counters:  map[string]int{}, Flags: map[string]int{},
	}
	gs.Seats[1].Battlefield = append(gs.Seats[1].Battlefield, enchantPerm)

	// Library has cards for the U mode (draw 2, discard 1).
	addLibraryWithTypes(gs, 0, "A", []string{"sorcery", "cmc:2"})
	addLibraryWithTypes(gs, 0, "B", []string{"sorcery", "cmc:3"})

	wistful := addPerm(gs, 0, "Wistfulness", "creature")
	wistful.Flags["was_cast"] = 1
	gameengine.InvokeETBHook(gs, wistful)

	// G mode: the higher-CMC opponent permanent (Rhystic Study, cmc 3)
	// should have been exiled.
	for _, p := range gs.Seats[1].Battlefield {
		if p == enchantPerm {
			t.Errorf("Wistfulness G mode should have exiled the highest-CMC artifact/enchantment (Rhystic Study)")
		}
	}

	// U mode: net hand should be 2 (drew 2, discarded 1). Hand starts at
	// 0 in the test fixture.
	if len(gs.Seats[0].Hand) != 1 {
		t.Errorf("Wistfulness U mode: expected hand size 1 after draw 2 + discard 1, got %d",
			len(gs.Seats[0].Hand))
	}
	if len(gs.Seats[0].Graveyard) < 1 {
		t.Errorf("Wistfulness U mode: expected discarded card in graveyard, got %d",
			len(gs.Seats[0].Graveyard))
	}
}

func TestWistfulness_NonCastEntryFiresNeitherMode(t *testing.T) {
	gs := newGame(t, 2)

	artifact := &gameengine.Card{
		Name: "Sol Ring", Owner: 1,
		Types: []string{"artifact", "cmc:1"},
	}
	artifactPerm := &gameengine.Permanent{
		Card: artifact, Controller: 1, Owner: 1,
		Timestamp: gs.NextTimestamp(),
		Counters:  map[string]int{}, Flags: map[string]int{},
	}
	gs.Seats[1].Battlefield = append(gs.Seats[1].Battlefield, artifactPerm)

	addLibraryWithTypes(gs, 0, "A", []string{"sorcery", "cmc:2"})

	wistful := addPerm(gs, 0, "Wistfulness", "creature")
	// No was_cast flag — reanimated / blinked / Sneak Attack path.
	gameengine.InvokeETBHook(gs, wistful)

	// Sol Ring should still be on the battlefield.
	stillThere := false
	for _, p := range gs.Seats[1].Battlefield {
		if p == artifactPerm {
			stillThere = true
		}
	}
	if !stillThere {
		t.Errorf("non-cast Wistfulness should not exile anything")
	}
	if len(gs.Seats[0].Hand) != 0 {
		t.Errorf("non-cast Wistfulness should not draw cards; hand=%d", len(gs.Seats[0].Hand))
	}
}

func TestDeceit_CastBouncesAndDiscards(t *testing.T) {
	gs := newGame(t, 2)

	// Opponent has a nonland permanent (will be bounced) and a hand with
	// a nonland card (will be discarded).
	threat := &gameengine.Card{
		Name: "Birds of Paradise", Owner: 1,
		Types: []string{"creature", "cmc:1"},
	}
	threatPerm := &gameengine.Permanent{
		Card: threat, Controller: 1, Owner: 1,
		Timestamp: gs.NextTimestamp(),
		Counters:  map[string]int{}, Flags: map[string]int{},
	}
	gs.Seats[1].Battlefield = append(gs.Seats[1].Battlefield, threatPerm)

	handCard := &gameengine.Card{
		Name: "Counterspell", Owner: 1,
		Types: []string{"instant", "cmc:2"},
	}
	gs.Seats[1].Hand = append(gs.Seats[1].Hand, handCard)

	deceit := addPerm(gs, 0, "Deceit", "creature")
	deceit.Flags["was_cast"] = 1
	gameengine.InvokeETBHook(gs, deceit)

	// Bounce: threatPerm should be gone from opponent's battlefield, in
	// their hand instead.
	for _, p := range gs.Seats[1].Battlefield {
		if p == threatPerm {
			t.Errorf("Deceit U mode should have bounced Birds of Paradise")
		}
	}

	// Discard: Counterspell should be gone from opponent's hand into their
	// graveyard. (Bounced Birds of Paradise now also occupies hand; check
	// the specific card name went to graveyard.)
	stillHasCounterspell := false
	for _, c := range gs.Seats[1].Hand {
		if c == handCard {
			stillHasCounterspell = true
		}
	}
	if stillHasCounterspell {
		t.Errorf("Deceit B mode should have discarded the opponent's Counterspell")
	}
	foundInGY := false
	for _, c := range gs.Seats[1].Graveyard {
		if c == handCard {
			foundInGY = true
		}
	}
	if !foundInGY {
		t.Errorf("Counterspell should have ended up in opponent's graveyard")
	}
}
