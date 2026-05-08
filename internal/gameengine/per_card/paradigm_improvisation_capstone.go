package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerImprovisationCapstone wires up Improvisation Capstone.
//
// Oracle text:
//
//	Exile cards from the top of your library until you exile cards
//	with total mana value 4 or greater. You may cast any number of
//	spells from among them without paying their mana costs. Paradigm
//
// {3}{R} Sorcery — Secrets of Strixhaven.
//
// Simplified: exile top cards until total CMC >= 4, then for each
// exiled nonland spell card, register a ZoneCastGrant for free
// casting from exile until end of turn.
func registerImprovisationCapstone(r *Registry) {
	r.OnResolve("Improvisation Capstone", improvisationCapstoneResolve)
}

func improvisationCapstoneResolve(gs *gameengine.GameState, item *gameengine.StackItem) {
	const slug = "improvisation_capstone"
	const cardName = "Improvisation Capstone"
	if gs == nil || item == nil {
		return
	}
	seat := item.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil {
		return
	}

	// --- Effect: exile from top of library until total CMC >= 4. ---
	totalCMC := 0
	exiledCards := []*gameengine.Card{}
	const maxExile = 20 // safety cap

	for totalCMC < 4 && len(s.Library) > 0 && len(exiledCards) < maxExile {
		card := s.Library[0]
		gameengine.MoveCard(gs, card, seat, "library", "exile", "effect")
		cmc := cardCMC(card)
		totalCMC += cmc
		exiledCards = append(exiledCards, card)
		gs.LogEvent(gameengine.Event{
			Kind:   "exile_from_library",
			Seat:   seat,
			Source: cardName,
			Details: map[string]interface{}{
				"card":      card.DisplayName(),
				"cmc":       cmc,
				"total_cmc": totalCMC,
			},
		})
	}

	// Register free cast grants for nonland spell cards among the exiled.
	castable := 0
	for _, c := range exiledCards {
		if c == nil {
			continue
		}
		// Skip land cards — "spells" only.
		if cardHasType(c, "land") {
			continue
		}
		perm := &gameengine.ZoneCastPermission{
			Zone:              "exile",
			Keyword:           "improvisation_capstone",
			ManaCost:          0,
			ExileOnResolve:    false,
			RequireController: seat,
			SourceName:        cardName,
			Duration:          "until_end_of_turn",
			GrantTurn:         gs.Turn,
		}
		gameengine.RegisterZoneCastGrant(gs, c, perm)
		castable++
	}

	// --- Paradigm: exile instead of graveyard, register for auto-copy. ---
	paradigmExileItem(gs, item, seat, slug, cardName)

	emit(gs, slug, cardName, map[string]interface{}{
		"seat":          seat,
		"exiled_count":  len(exiledCards),
		"total_cmc":     totalCMC,
		"castable":      castable,
	})
}
