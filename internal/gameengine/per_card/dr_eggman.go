package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerDrEggman wires Dr. Eggman.
//
// Oracle text:
//
//	Flying
//	At the beginning of your end step, draw a card. Then each opponent
//	faces a villainous choice — That player discards a card, or you may
//	put a Construct, Robot, or Vehicle card from your hand onto the
//	battlefield.
//
// Implementation: draw a card; for each opponent, if controller has a
// Construct/Robot/Vehicle in hand we want to drop, deploy it; otherwise
// force opponent to discard.
func registerDrEggman(r *Registry) {
	r.OnTrigger("Dr. Eggman", "end_step", drEggmanEndStep)
}

func drEggmanEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "dr_eggman_villainous_choice"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	deployed := 0
	discarded := 0
	for i, opp := range gs.Seats {
		if opp == nil || i == perm.Controller || opp.Lost {
			continue
		}
		// Find a deployable card in our hand.
		me := gs.Seats[perm.Controller]
		var target *gameengine.Card
		var targetIdx = -1
		bestCMC := -1
		for j, c := range me.Hand {
			if c == nil {
				continue
			}
			if !cardHasType(c, "construct") && !cardHasType(c, "robot") && !cardHasType(c, "vehicle") {
				continue
			}
			cmc := gameengine.ManaCostOf(c)
			if cmc > bestCMC {
				bestCMC = cmc
				target = c
				targetIdx = j
			}
		}
		if target != nil && targetIdx >= 0 {
			gameengine.MoveCard(gs, target, perm.Controller, "hand", "battlefield", "dr_eggman")
			enterBattlefieldWithETB(gs, perm.Controller, target, false)
			deployed++
		} else if len(opp.Hand) > 0 {
			c := opp.Hand[len(opp.Hand)-1]
			gameengine.DiscardCard(gs, c, i)
			discarded++
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"deployed":  deployed,
		"discarded": discarded,
	})
}
