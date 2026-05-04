package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerColfenorTheLastYew wires Colfenor, the Last Yew.
//
// Oracle text:
//
//	Vigilance, reach
//	Whenever Colfenor or another creature you control dies, return up
//	to one other target creature card with lesser toughness from your
//	graveyard to your hand.
func registerColfenorTheLastYew(r *Registry) {
	r.OnTrigger("Colfenor, the Last Yew", "creature_dies", colfenorDies)
}

func colfenorDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "colfenor_creature_dies"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	deadController, _ := ctx["controller_seat"].(int)
	if deadController != perm.Controller {
		return
	}
	deadPerm, _ := ctx["perm"].(*gameengine.Permanent)
	if deadPerm == nil || deadPerm.Card == nil {
		return
	}
	deadTough := deadPerm.Toughness()
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	// Find creature card in graveyard with toughness < deadTough.
	var bestIdx = -1
	bestTough := -1
	for i, c := range seat.Graveyard {
		if c == nil || !cardHasType(c, "creature") {
			continue
		}
		if c.BaseToughness >= deadTough {
			continue
		}
		if c.BaseToughness > bestTough {
			bestTough = c.BaseToughness
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		return
	}
	card := seat.Graveyard[bestIdx]
	gameengine.MoveCard(gs, card, perm.Controller, "graveyard", "hand", "colfenor")
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"returned": card.DisplayName(),
	})
}
