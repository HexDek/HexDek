package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerHenryWuIngenGeneticist wires Henry Wu, InGen Geneticist.
//
// Oracle text:
//
//	Henry Wu and other Human creatures you control have exploit.
//	(When a creature with exploit enters, you may sacrifice a creature.)
//	Whenever a creature you control exploits a non-Human creature,
//	draw a card. If the exploited creature had power 3 or greater,
//	create a Treasure token.
//
// The exploit static-grant and exploit-sacrifice trigger aren't tracked
// at the per-card hook surface — emitPartial.
func registerHenryWuIngenGeneticist(r *Registry) {
	r.OnETB("Henry Wu, InGen Geneticist", henryWuETB)
}

func henryWuETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "henry_wu_partial"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"exploit_grant_and_exploit_payoff_trigger_unimplemented")
}
