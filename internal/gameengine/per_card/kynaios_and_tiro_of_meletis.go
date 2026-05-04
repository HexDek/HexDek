package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKynaiosAndTiroOfMeletis wires Kynaios and Tiro of Meletis.
//
// Oracle text:
//
//	At the beginning of your end step, draw a card. Each player may put
//	a land card from their hand onto the battlefield, then each opponent
//	who didn't draws a card.
//
// Implementation: end step (own turn only). Controller draws. Then for
// each player (opponents and self), if they have a land in hand, drop
// the lowest-impact basic. Each opponent who didn't draws a card.
func registerKynaiosAndTiroOfMeletis(r *Registry) {
	r.OnTrigger("Kynaios and Tiro of Meletis", "end_step", kynaiosEndStep)
}

func kynaiosFirstLandInHand(seat *gameengine.Seat) *gameengine.Card {
	if seat == nil {
		return nil
	}
	for _, c := range seat.Hand {
		if c != nil && cardHasType(c, "land") {
			return c
		}
	}
	return nil
}

func kynaiosEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kynaios_tiro_end_step"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	results := map[int]bool{}
	for i := range gs.Seats {
		s := gs.Seats[i]
		if s == nil || s.Lost {
			continue
		}
		land := kynaiosFirstLandInHand(s)
		if land != nil {
			moveCardBetweenZones(gs, i, land, "hand", "battlefield", "kynaios_land_drop")
			results[i] = true
		} else {
			results[i] = false
		}
	}
	for i, played := range results {
		if i == perm.Controller {
			continue
		}
		if !played {
			drawOne(gs, i, perm.Card.DisplayName())
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":           perm.Controller,
		"land_drops":     results,
	})
}
