package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKamizObscuraOculus wires Kamiz, Obscura Oculus.
//
// Oracle text:
//
//	Whenever you attack, target attacking creature can't be blocked
//	this turn. It connives. Then choose another attacking creature
//	with lesser power. That creature gains double strike until end
//	of turn.
//
// Connive (draw + discard, possibly +1/+1 counter) and the attacking
// power-comparison double-strike grant aren't supported by the per-card
// surface — emitPartial.
func registerKamizObscuraOculus(r *Registry) {
	r.OnTrigger("Kamiz, Obscura Oculus", "attacks", kamizAttacks)
}

func kamizAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kamiz_obscura_partial"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	attackerSeat, _ := ctx["seat"].(int)
	if attackerSeat != perm.Controller {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"unblockable_connive_and_double_strike_unimplemented")
}
