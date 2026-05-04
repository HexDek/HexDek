package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCommodoreGuff wires Commodore Guff (planeswalker).
//
// Oracle text:
//
//	At the beginning of your end step, put a loyalty counter on another
//	target planeswalker you control.
//	+1: Create a 1/1 red Wizard creature token with "{T}: Add {R}.
//	    Spend this mana only to cast a planeswalker spell."
//	−3: You draw X cards and Commodore Guff deals X damage to each
//	    opponent, where X is the number of planeswalkers you control.
//	Commodore Guff can be your commander.
//
// Planeswalker activations are out of scope; we wire the end-step
// loyalty pump.
func registerCommodoreGuff(r *Registry) {
	r.OnTrigger("Commodore Guff", "end_step", commodoreGuffEndStep)
}

func commodoreGuffEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "commodore_guff_loyalty_pump"
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
		if p == nil || p == perm || p.Card == nil {
			continue
		}
		if !cardHasType(p.Card, "planeswalker") {
			continue
		}
		p.AddCounter("loyalty", 1)
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"target": p.Card.DisplayName(),
		})
		return
	}
}
