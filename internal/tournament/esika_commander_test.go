package tournament

// Regression tests for the MDFC commander back-face cast bug.
//
// Background: tryCastCommander historically read the front-face mana
// cost via gameengine.ManaCostOf and never set CastingBackFace, so an
// MDFC commander like Esika, God of the Tree // The Prismatic Bridge
// always entered as the 4/4 god creature. Bridge — the entire reason
// the deck exists — never deployed. Esika decks measured at 9.2% WR.
//
// Fix: tryCastCommander now consults mdfcPreferBackFace for MDFC
// commanders whose back face is a non-creature spell, sets
// CastingBackFace=true when the back face is affordable, and pays the
// back-face cost. The resolve path in stack.go swaps the runtime
// Name/Types/CMC to the back face before constructing the Permanent,
// so the on-battlefield card is the Bridge (legendary enchantment),
// not the god (legendary creature).

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
	"github.com/hexdek/hexdek/internal/hat"
)

// esikaCommanderCard builds a synthetic Esika, God of the Tree //
// The Prismatic Bridge MDFC card matching what the deckparser
// produces from oracle-cards.json (full DFC name, both faces' types,
// BackFaceCMC populated).
func esikaCommanderCard() *gameengine.Card {
	return &gameengine.Card{
		Name:             "Esika, God of the Tree // The Prismatic Bridge",
		Types:            []string{"legendary", "creature", "god"},
		TypeLine:         "legendary creature — god // legendary enchantment",
		CMC:              4, // {1}{G}{G}{G}
		Colors:           []string{"G", "B", "R", "U", "W"},
		BasePower:        4,
		BaseToughness:    4,
		BackFaceName:     "The Prismatic Bridge",
		BackFaceCMC:      6, // {1}{W}{U}{B}{R}{G}
		BackFaceTypes:    []string{"legendary", "enchantment"},
		BackFaceTypeLine: "legendary enchantment",
	}
}

// setupCommanderTestGame stands up a minimal 2-seat commander game with
// Esika in seat 0's command zone, a GreedyHat, and the requested amount
// of available mana. Seat 0 is the active player.
func setupCommanderTestGame(t *testing.T, mana int) (*gameengine.GameState, *gameengine.Card) {
	t.Helper()
	gs := gameengine.NewGameState(2, rand.New(rand.NewSource(1)), nil)
	cmdr := esikaCommanderCard()
	cmdr.Owner = 0
	deck := &gameengine.CommanderDeck{CommanderCards: []*gameengine.Card{cmdr}}
	emptyDeck := &gameengine.CommanderDeck{}
	gameengine.SetupCommanderGame(gs, []*gameengine.CommanderDeck{deck, emptyDeck})
	for i := range gs.Seats {
		gs.Seats[i].Hat = &hat.GreedyHat{}
	}
	gs.Active = 0
	gs.Seats[0].ManaPool = mana
	gameengine.EnsureTypedPool(gs.Seats[0])
	return gs, cmdr
}

// drainAllStack resolves every item on the stack — needed because
// tryCastCommander only pushes the spell; resolution happens via the
// turn-loop's drainStack. Tests bypass the turn loop, so we drive
// resolution manually.
func drainAllStack(gs *gameengine.GameState) {
	for len(gs.Stack) > 0 {
		gameengine.ResolveStackTop(gs)
		gameengine.StateBasedActions(gs)
	}
}

// TestEsikaCommander_PrefersBackFaceWhenAffordable is the headline
// regression: with 6 mana available (back-face cost), the AI must
// cast Esika as the Bridge enchantment, not as the front-face god.
func TestEsikaCommander_PrefersBackFaceWhenAffordable(t *testing.T) {
	gs, cmdr := setupCommanderTestGame(t, 6)

	tryCastCommander(gs, 0)
	drainAllStack(gs)

	if len(gs.Seats[0].CommandZone) != 0 {
		t.Fatalf("expected commander to leave command zone; %d remain", len(gs.Seats[0].CommandZone))
	}
	if len(gs.Seats[0].Battlefield) != 1 {
		t.Fatalf("expected 1 permanent on battlefield, got %d", len(gs.Seats[0].Battlefield))
	}
	perm := gs.Seats[0].Battlefield[0]
	if perm.IsCreature() {
		t.Errorf("expected back face (Bridge enchantment), but permanent is a creature; types=%v",
			perm.Card.Types)
	}
	if !perm.IsEnchantment() {
		t.Errorf("expected back face (Bridge enchantment); types=%v", perm.Card.Types)
	}
	if perm.Card.Name != "The Prismatic Bridge" {
		t.Errorf("expected Card.Name swapped to 'The Prismatic Bridge'; got %q", perm.Card.Name)
	}
	// CastingBackFace is a transient flag — the resolve path must clear
	// it so a stale flip doesn't re-fire on a subsequent cast.
	if cmdr.CastingBackFace {
		t.Errorf("CastingBackFace should be cleared after resolve; still true")
	}
}

// TestEsikaCommander_FallsBackToFrontFaceWhenBackUnaffordable verifies
// that the back-face preference doesn't refuse to cast Esika just
// because Bridge's 6-mana cost isn't met. With only 4 mana, the AI
// should cast the front-face god (cost 4) rather than waiting.
func TestEsikaCommander_FallsBackToFrontFaceWhenBackUnaffordable(t *testing.T) {
	gs, _ := setupCommanderTestGame(t, 4)

	tryCastCommander(gs, 0)
	drainAllStack(gs)

	if len(gs.Seats[0].Battlefield) != 1 {
		t.Fatalf("expected 1 permanent on battlefield, got %d", len(gs.Seats[0].Battlefield))
	}
	perm := gs.Seats[0].Battlefield[0]
	if !perm.IsCreature() {
		t.Errorf("expected front-face creature when back face unaffordable; types=%v",
			perm.Card.Types)
	}
	if perm.Card.Name != "Esika, God of the Tree // The Prismatic Bridge" {
		t.Errorf("expected card name unchanged for front-face cast; got %q", perm.Card.Name)
	}
}

// TestEsikaCommander_NoCastWhenInsufficientMana verifies the AI
// doesn't double-bill: with 3 mana (less than even front-face cost),
// neither face should be cast and the commander stays in the zone.
func TestEsikaCommander_NoCastWhenInsufficientMana(t *testing.T) {
	gs, _ := setupCommanderTestGame(t, 3)

	tryCastCommander(gs, 0)
	drainAllStack(gs)

	if len(gs.Seats[0].CommandZone) != 1 {
		t.Errorf("expected commander to stay in command zone; %d cards there",
			len(gs.Seats[0].CommandZone))
	}
	if len(gs.Seats[0].Battlefield) != 0 {
		t.Errorf("expected empty battlefield; got %d permanents",
			len(gs.Seats[0].Battlefield))
	}
}

// TestEsikaCommander_DFCAwareLookupAfterBackFaceCast verifies that
// CastCommanderFromCommandZone's lookup is DFC-aware. After a back-face
// cast mutates Card.Name to "The Prismatic Bridge", the card may
// later return to the command zone (e.g. via §903.9b commander-zone
// redirect). A subsequent cast attempt must still find the card by
// the canonical full-DFC commander name. We simulate that by manually
// returning the (mutated) card to the command zone.
func TestEsikaCommander_DFCAwareLookupAfterBackFaceCast(t *testing.T) {
	gs, cmdr := setupCommanderTestGame(t, 6)

	// First cast: back face.
	tryCastCommander(gs, 0)
	drainAllStack(gs)
	if len(gs.Seats[0].Battlefield) != 1 {
		t.Fatalf("setup: first cast must place Bridge on battlefield")
	}

	// Simulate Bridge dying and §903.9b redirecting to command zone.
	// At that point the card's Name is still mutated to "The Prismatic Bridge".
	gs.Seats[0].Battlefield = nil
	gs.Seats[0].CommandZone = append(gs.Seats[0].CommandZone, cmdr)

	// Give the seat enough mana for the recast (6 base + 2 tax = 8).
	gs.Seats[0].ManaPool = 8
	gameengine.EnsureTypedPool(gs.Seats[0])

	// Calling CastCommanderFromCommandZone with the canonical full-DFC
	// commanderName must still find the (mutated) card.
	err := gameengine.CastCommanderFromCommandZone(gs, 0,
		"Esika, God of the Tree // The Prismatic Bridge", 6)
	if err != nil {
		t.Fatalf("DFC-aware lookup must find the card despite name mutation; got %v", err)
	}
}
