package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRestorationSeminar wires up Restoration Seminar.
//
// Oracle text:
//
//	Return target nonland permanent card from your graveyard to the
//	battlefield. Paradigm
//
// {3}{W} Sorcery — Secrets of Strixhaven.
//
// Simplified: find the best nonland permanent card in controller's
// graveyard (highest CMC), return it to the battlefield. Then exile
// this spell and register for paradigm auto-copy.
func registerRestorationSeminar(r *Registry) {
	r.OnResolve("Restoration Seminar", restorationSeminarResolve)
}

func restorationSeminarResolve(gs *gameengine.GameState, item *gameengine.StackItem) {
	const slug = "restoration_seminar"
	const cardName = "Restoration Seminar"
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

	// --- Effect: return best nonland permanent card from graveyard. ---
	var bestCard *gameengine.Card
	bestCMC := -1
	for _, c := range s.Graveyard {
		if c == nil {
			continue
		}
		// Must be a permanent card (not instant/sorcery) and not a land.
		if cardHasType(c, "land") {
			continue
		}
		if cardHasType(c, "instant") || cardHasType(c, "sorcery") {
			continue
		}
		// Must be a permanent type.
		isPerm := cardHasType(c, "creature") || cardHasType(c, "artifact") ||
			cardHasType(c, "enchantment") || cardHasType(c, "planeswalker") ||
			cardHasType(c, "battle")
		if !isPerm {
			continue
		}
		cmc := cardCMC(c)
		if cmc > bestCMC {
			bestCMC = cmc
			bestCard = c
		}
	}

	if bestCard == nil {
		emitFail(gs, slug, cardName, "no_nonland_permanent_in_graveyard", nil)
		// Still do paradigm exile even if no valid target.
		paradigmExileItem(gs, item, seat, slug, cardName)
		return
	}

	// Move from graveyard to battlefield. MoveCard handles ETB triggers.
	gameengine.MoveCard(gs, bestCard, seat, "graveyard", "battlefield", "restoration_seminar")
	gs.LogEvent(gameengine.Event{
		Kind:   "return_to_battlefield",
		Seat:   seat,
		Source: cardName,
		Details: map[string]interface{}{
			"card":   bestCard.DisplayName(),
			"reason": slug,
		},
	})

	// --- Paradigm: exile instead of graveyard, register for auto-copy. ---
	paradigmExileItem(gs, item, seat, slug, cardName)

	emit(gs, slug, cardName, map[string]interface{}{
		"seat":     seat,
		"returned": bestCard.DisplayName(),
	})
}
