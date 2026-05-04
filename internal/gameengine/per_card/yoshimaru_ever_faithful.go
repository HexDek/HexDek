package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerYoshimaruEverFaithful wires Yoshimaru, Ever Faithful.
//
// Oracle text:
//
//	Whenever another legendary permanent you control enters, put a +1/+1
//	counter on Yoshimaru.
//	Partner (You can have two commanders if both have partner.)
//
// Implementation:
//   - "permanent_etb" filtered to entering perm being legendary, our
//     control, not Yoshimaru himself. Adds 1 +1/+1 counter.
//   - Partner is a deck-construction rule, not a runtime ability.
func registerYoshimaruEverFaithful(r *Registry) {
	r.OnTrigger("Yoshimaru, Ever Faithful", "permanent_etb", yoshimaruLegendaryETB)
}

func yoshimaruLegendaryETB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "yoshimaru_legendary_counter"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	controllerSeat, _ := ctx["controller_seat"].(int)
	if controllerSeat != perm.Controller {
		return
	}
	entered, _ := ctx["perm"].(*gameengine.Permanent)
	if entered == nil || entered == perm || entered.Card == nil {
		return
	}
	if !cardHasType(entered.Card, "legendary") {
		return
	}
	perm.AddCounter("+1/+1", 1)
	gs.InvalidateCharacteristicsCache()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":    perm.Controller,
		"entered": entered.Card.DisplayName(),
	})
}
