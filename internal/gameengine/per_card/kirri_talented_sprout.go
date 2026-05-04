package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKirriTalentedSprout wires Kirri, Talented Sprout.
//
// Oracle text:
//
//	Other Plants and Treefolk you control get +2/+0.
//	At the beginning of each of your postcombat main phases, return
//	target Plant, Treefolk, or land card from your graveyard to your hand.
//
// Implementation: postcombat_main_controller hook returns the highest-CMC
// plant/treefolk/land from controller's graveyard to hand. The +2/+0
// anthem is a continuous effect handled by the engine's static layer.
func registerKirriTalentedSprout(r *Registry) {
	r.OnTrigger("Kirri, Talented Sprout", "postcombat_main_controller", kirriPostcombat)
}

func kirriPostcombat(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kirri_postcombat_recur"
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
	bestIdx := -1
	bestCMC := -1
	for i, c := range seat.Graveyard {
		if c == nil {
			continue
		}
		if !cardHasType(c, "plant") && !cardHasType(c, "treefolk") && !cardHasType(c, "land") {
			continue
		}
		cmc := cardCMC(c)
		if cmc > bestCMC {
			bestCMC = cmc
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_valid_target", map[string]interface{}{
			"seat": perm.Controller,
		})
		return
	}
	target := seat.Graveyard[bestIdx]
	moveCardBetweenZones(gs, perm.Controller, target, "graveyard", "hand", "kirri_recur")
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"target": target.DisplayName(),
	})
}
