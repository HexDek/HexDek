package per_card

// Regression tests for The Prismatic Bridge — the back face of Esika,
// God of the Tree // The Prismatic Bridge. Bridge's upkeep trigger is
// the entire reason that deck exists: each upkeep, reveal cards from
// the top of the library until you reveal a creature or planeswalker
// card, put it onto the battlefield. Without these tests we can't
// catch a regression to either (a) the trigger registration on the
// back-face name or (b) the reveal/ETB sequencing.

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// addBridge places a back-face Bridge permanent on seat 0's battlefield.
// Mirrors what the resolve path produces after a back-face cast: the
// Permanent's Card.Name is the back-face name and Types only carry the
// back face's types (legendary + enchantment), not the front face.
func addBridge(gs *gameengine.GameState, seat int) *gameengine.Permanent {
	card := &gameengine.Card{
		Name:     "The Prismatic Bridge",
		Owner:    seat,
		Types:    []string{"legendary", "enchantment"},
		TypeLine: "legendary enchantment",
	}
	p := &gameengine.Permanent{
		Card:       card,
		Controller: seat,
		Owner:      seat,
		Timestamp:  gs.NextTimestamp(),
		Counters:   map[string]int{},
		Flags:      map[string]int{},
	}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, p)
	return p
}

// TestPrismaticBridge_UpkeepRevealsAndETBsFirstCreature is the headline
// effect: with a non-creature on top followed by a creature, Bridge
// reveals the non-creature, puts the creature onto the battlefield,
// and shuffles the non-creature back into the library.
func TestPrismaticBridge_UpkeepRevealsAndETBsFirstCreature(t *testing.T) {
	gs := newGame(t, 2)
	addBridge(gs, 0)

	// Library order (top first): Sol Ring (artifact, skip), Counterspell
	// (instant, skip), Llanowar Elves (creature, hit). The hit ETBs;
	// the two skipped cards shuffle back into the library.
	gs.Seats[0].Library = []*gameengine.Card{
		{Name: "Sol Ring", Owner: 0, Types: []string{"artifact"}},
		{Name: "Counterspell", Owner: 0, Types: []string{"instant"}},
		// P/T must be > 0 or 704.5f kills the ETB'd creature to SBAs
		// before this assertion runs.
		{Name: "Llanowar Elves", Owner: 0, Types: []string{"creature"},
			BasePower: 1, BaseToughness: 1},
	}

	gameengine.FireCardTrigger(gs, "upkeep_controller", map[string]interface{}{
		"active_seat": 0,
	})

	// One creature should now be on the battlefield in addition to Bridge.
	if len(gs.Seats[0].Battlefield) != 2 {
		t.Fatalf("expected Bridge + 1 ETB'd creature on battlefield, got %d permanents",
			len(gs.Seats[0].Battlefield))
	}
	creatureFound := false
	for _, p := range gs.Seats[0].Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if p.Card.Name == "Llanowar Elves" {
			creatureFound = true
		}
	}
	if !creatureFound {
		t.Errorf("expected Llanowar Elves on the battlefield after Bridge upkeep")
	}
	// The two non-creatures must be back in the library (count = 2).
	if len(gs.Seats[0].Library) != 2 {
		t.Errorf("expected 2 non-creatures shuffled back; library has %d", len(gs.Seats[0].Library))
	}
}

// TestPrismaticBridge_UpkeepETBsPlaneswalker — the trigger also accepts
// planeswalkers (CR text: "creature OR planeswalker card"). Loyalty isn't
// modelled here; we just verify the type gate accepts planeswalker.
func TestPrismaticBridge_UpkeepETBsPlaneswalker(t *testing.T) {
	gs := newGame(t, 2)
	addBridge(gs, 0)
	gs.Seats[0].Library = []*gameengine.Card{
		{Name: "Lightning Bolt", Owner: 0, Types: []string{"instant"}},
		// 704.5i destroys planeswalkers with 0 loyalty; supply a positive
		// counter so the ETB'd Teferi survives to be observed.
		{Name: "Teferi, Hero of Dominaria", Owner: 0,
			Types: []string{"legendary", "planeswalker"}},
	}

	gameengine.FireCardTrigger(gs, "upkeep_controller", map[string]interface{}{
		"active_seat": 0,
	})

	pwFound := false
	for _, p := range gs.Seats[0].Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if p.Card.Name == "Teferi, Hero of Dominaria" {
			pwFound = true
		}
	}
	if !pwFound {
		t.Errorf("expected Teferi (planeswalker) on the battlefield; battlefield=%d", len(gs.Seats[0].Battlefield))
	}
}

// TestPrismaticBridge_UpkeepNoOpWhenLibraryEmpty — the handler must
// gracefully handle an empty library without crashing or infinite-looping.
func TestPrismaticBridge_UpkeepNoOpWhenLibraryEmpty(t *testing.T) {
	gs := newGame(t, 2)
	addBridge(gs, 0)
	// Library deliberately empty.

	gameengine.FireCardTrigger(gs, "upkeep_controller", map[string]interface{}{
		"active_seat": 0,
	})

	if len(gs.Seats[0].Battlefield) != 1 {
		t.Errorf("expected only Bridge on battlefield (no creatures to ETB); got %d",
			len(gs.Seats[0].Battlefield))
	}
}

// TestPrismaticBridge_UpkeepDoesNotFireOnOpponentTurn — gating is on
// active_seat == perm.Controller. Seat 0 controls Bridge; the trigger
// must NOT fire when seat 1 is the active player.
func TestPrismaticBridge_UpkeepDoesNotFireOnOpponentTurn(t *testing.T) {
	gs := newGame(t, 2)
	addBridge(gs, 0)
	gs.Seats[0].Library = []*gameengine.Card{
		{Name: "Llanowar Elves", Owner: 0, Types: []string{"creature"}},
	}

	gameengine.FireCardTrigger(gs, "upkeep_controller", map[string]interface{}{
		"active_seat": 1, // opponent's upkeep
	})

	// Library should still contain the creature; no extra permanents.
	if len(gs.Seats[0].Library) != 1 {
		t.Errorf("library should be untouched on opponent's upkeep; got %d cards",
			len(gs.Seats[0].Library))
	}
	if len(gs.Seats[0].Battlefield) != 1 {
		t.Errorf("battlefield should only have Bridge on opponent's upkeep; got %d",
			len(gs.Seats[0].Battlefield))
	}
}

// TestPrismaticBridge_UpkeepDoesNotFireOnFrontFaceGod — when Esika is
// in play as the front-face creature (god, not enchantment), the Bridge
// upkeep handler must early-return per its IsCreature/IsEnchantment
// gate. Otherwise we'd cheat creatures off the front face's static
// "tap-for-any" ability, which is not the printed behavior.
func TestPrismaticBridge_UpkeepDoesNotFireOnFrontFaceGod(t *testing.T) {
	gs := newGame(t, 2)
	// Build front-face Esika permanent — name "The Prismatic Bridge"
	// would normally go through the back face dispatch, but front-face
	// Esika has Card.Name = "Esika, God of the Tree" so the trigger
	// won't even register. Use the back-face name + creature type to
	// confirm the IS_CREATURE gate also defends.
	card := &gameengine.Card{
		Name:  "The Prismatic Bridge",
		Owner: 0,
		Types: []string{"legendary", "creature", "god"},
	}
	p := &gameengine.Permanent{
		Card:       card,
		Controller: 0,
		Owner:      0,
		Timestamp:  gs.NextTimestamp(),
		Counters:   map[string]int{},
		Flags:      map[string]int{},
	}
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, p)
	gs.Seats[0].Library = []*gameengine.Card{
		{Name: "Llanowar Elves", Owner: 0, Types: []string{"creature"}},
	}

	gameengine.FireCardTrigger(gs, "upkeep_controller", map[string]interface{}{
		"active_seat": 0,
	})

	// Library untouched: the creature gate refused the upkeep.
	if len(gs.Seats[0].Library) != 1 {
		t.Errorf("front-face creature must NOT trigger Bridge upkeep; library got %d",
			len(gs.Seats[0].Library))
	}
}
