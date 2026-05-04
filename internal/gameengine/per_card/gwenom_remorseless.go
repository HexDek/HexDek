package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGwenomRemorseless wires Gwenom, Remorseless.
//
// Oracle text:
//
//	Deathtouch, lifelink
//	Whenever Gwenom attacks, until end of turn, you may look at the
//	top card of your library any time and you may play cards from
//	the top of your library. If you cast a spell this way, pay life
//	equal to its mana value rather than pay its mana cost.
//
// The "play from top of library" alternative-cost is non-trivial; the
// engine doesn't yet expose that surface for AI use — emitPartial.
func registerGwenomRemorseless(r *Registry) {
	r.OnTrigger("Gwenom, Remorseless", "attacks", gwenomRemorselessAttacks)
}

func gwenomRemorselessAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "gwenom_remorseless_attacks"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	attackerSeat, _ := ctx["seat"].(int)
	if attackerSeat != perm.Controller {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"play_from_top_of_library_with_life_payment_unimplemented")
}
