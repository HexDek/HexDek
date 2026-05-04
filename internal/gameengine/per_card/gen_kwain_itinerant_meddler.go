package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKwainItinerantMeddler wires Kwain, Itinerant Meddler.
//
// Oracle text:
//
//   {T}: Each player may draw a card, then each player who drew a card this way gains 1 life.
//
// Implementation: each non-eliminated player draws (heuristic: always
// take the optional draw) and gains 1 life if they drew.
func registerKwainItinerantMeddler(r *Registry) {
	r.OnActivated("Kwain, Itinerant Meddler", kwainItinerantMeddlerActivate)
}

func kwainItinerantMeddlerActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "kwain_itinerant_meddler_activate"
	if gs == nil || src == nil {
		return
	}
	drew := 0
	for i := range gs.Seats {
		s := gs.Seats[i]
		if s == nil || s.Lost {
			continue
		}
		if drawOne(gs, i, src.Card.DisplayName()) != nil {
			gameengine.GainLife(gs, i, 1, src.Card.DisplayName())
			drew++
		}
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":     src.Controller,
		"drew":     drew,
	})
}
