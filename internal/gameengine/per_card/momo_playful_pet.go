package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMomoPlayfulPet wires Momo, Playful Pet.
//
// Oracle text:
//
//	Flying, vigilance
//	When Momo leaves the battlefield, choose one —
//	  • Create a Food token.
//	  • Put a +1/+1 counter on target creature you control.
//	  • Scry 2.
//
// AI policy: prefer +1/+1 counter on best creature; otherwise create
// Food; otherwise scry 2 (left as emitPartial — no scry helper here).
func registerMomoPlayfulPet(r *Registry) {
	r.OnTrigger("Momo, Playful Pet", "permanent_ltb", momoLTB)
}

func momoLTB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "momo_ltb_modal"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	deadController, _ := ctx["controller_seat"].(int)
	if deadController != perm.Controller {
		return
	}
	deadCard, _ := ctx["card"].(*gameengine.Card)
	if deadCard == nil || deadCard.DisplayName() != perm.Card.DisplayName() {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	var best *gameengine.Permanent
	for _, p := range seat.Battlefield {
		if p == nil || p == perm || !p.IsCreature() {
			continue
		}
		if best == nil || p.Power() > best.Power() {
			best = p
		}
	}
	if best != nil {
		best.AddCounter("+1/+1", 1)
		gs.InvalidateCharacteristicsCache()
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"mode":   "plus_one_counter",
			"target": best.Card.DisplayName(),
		})
		return
	}
	tok := gameengine.CreateCreatureToken(gs, perm.Controller, "Food",
		[]string{"artifact", "food"}, 0, 0)
	if tok != nil && tok.Card != nil {
		tok.Card.Types = []string{"token", "artifact", "food"}
		tok.Card.BasePower = 0
		tok.Card.BaseToughness = 0
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
		"mode": "food_token",
	})
}
