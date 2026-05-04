package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerBugenhagenWiseElder wires Bugenhagen, Wise Elder.
//
// Oracle text:
//
//	Reach
//	At the beginning of your upkeep, if you control a creature with
//	power 7 or greater, draw a card.
//	{T}: Add one mana of any color.
func registerBugenhagenWiseElder(r *Registry) {
	r.OnTrigger("Bugenhagen, Wise Elder", "upkeep", bugenhagenUpkeep)
}

func bugenhagenUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "bugenhagen_power_7_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	for _, p := range seat.Battlefield {
		if p != nil && p.IsCreature() && p.Power() >= 7 {
			drawOne(gs, perm.Controller, perm.Card.DisplayName())
			emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
				"seat":     perm.Controller,
				"big_creature": p.Card.DisplayName(),
				"power":    p.Power(),
			})
			return
		}
	}
}
