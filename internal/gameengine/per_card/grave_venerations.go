package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGraveVenerations wires Grave Venerations.
//
// Oracle text (Scryfall, verified 2026-05-16):
//
//	When this enchantment enters, you become the monarch.
//	At the beginning of your end step, if you're the monarch, return up
//	to one target creature card from your graveyard to your hand.
//	Whenever a creature you control dies, each opponent loses 1 life
//	and you gain 1 life.
//
// Implementation:
//   - ETB: BecomeMonarch(controller).
//   - End step gated on active_seat == controller AND IsMonarch. Picks
//     the highest-CMC creature card in graveyard (most likely the most
//     valuable recursion target).
//   - Creature-dies aristocrat drain: filtered to creatures we control.
func registerGraveVenerations(r *Registry) {
	r.OnETB("Grave Venerations", graveVenerationsETB)
	r.OnTrigger("Grave Venerations", "end_step", graveVenerationsEndStep)
	r.OnTrigger("Grave Venerations", "creature_dies", graveVenerationsCreatureDies)
}

func graveVenerationsETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	gameengine.BecomeMonarch(gs, perm.Controller)
	emit(gs, "grave_venerations_etb_monarch", perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}

func graveVenerationsEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "grave_venerations_end_step_recur"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, ok := ctx["active_seat"].(int)
	if !ok || activeSeat != perm.Controller {
		return
	}
	if !gameengine.IsMonarch(gs, perm.Controller) {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	bestIdx := -1
	bestCMC := -1
	for i, c := range seat.Graveyard {
		if c == nil || !cardHasType(c, "creature") {
			continue
		}
		cmc := cardCMC(c)
		if cmc > bestCMC {
			bestCMC = cmc
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":     perm.Controller,
			"returned": nil,
		})
		return
	}
	card := seat.Graveyard[bestIdx]
	gameengine.MoveCard(gs, card, perm.Controller, "graveyard", "hand", "grave_venerations_recur")
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"returned": card.DisplayName(),
	})
}

func graveVenerationsCreatureDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "grave_venerations_drain"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	deadController, _ := ctx["controller_seat"].(int)
	if deadController != perm.Controller {
		return
	}
	drained := 0
	for i, s := range gs.Seats {
		if s == nil || s.Lost || i == perm.Controller {
			continue
		}
		gameengine.LoseLife(gs, i, 1, perm.Card.DisplayName())
		drained++
	}
	_ = gs.CheckEnd()
	if drained > 0 {
		gameengine.GainLife(gs, perm.Controller, 1, perm.Card.DisplayName())
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"opps_hit": drained,
	})
}
