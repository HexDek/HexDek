package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRadhaHeirToKeld wires Radha, Heir to Keld.
//
// Oracle text:
//
//	Whenever Radha attacks, you may add {R}{R}.
//	{T}: Add {G}.
//
// Tap-for-G is a stock mana ability handled by the engine. The attack
// trigger adds RR to the controller's mana pool (we always take the
// "may" since AI prefers ramp).
func registerRadhaHeirToKeld(r *Registry) {
	r.OnTrigger("Radha, Heir to Keld", "attacks", radhaAttack)
}

func radhaAttack(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "radha_attack_rr"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	attackerSeat, _ := ctx["seat"].(int)
	if attackerSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if seat.Mana != nil {
		seat.Mana.R += 2
	} else {
		seat.ManaPool += 2
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"added": "RR",
	})
}
