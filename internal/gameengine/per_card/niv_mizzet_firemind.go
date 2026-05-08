package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerNivMizzetTheFiremind wires Niv-Mizzet, the Firemind.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Flying
//	Whenever you draw a card, Niv-Mizzet deals 1 damage to any target.
//	{T}: Draw a card.
//
// Implementation mirrors Niv-Mizzet, Parun's draw-trigger ping but
// without the spell-cast draw payoff. Activated tap-to-draw is documented
// as the standard activation pipeline.
func registerNivMizzetTheFiremind(r *Registry) {
	r.OnTrigger("Niv-Mizzet, the Firemind", "player_would_draw", nivMizzetFiremindDraw)
	r.OnActivated("Niv-Mizzet, the Firemind", nivMizzetFiremindActivate)
}

func nivMizzetFiremindDraw(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "niv_mizzet_firemind_draw_damage"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	drawSeat, _ := ctx["draw_seat"].(int)
	if drawSeat != perm.Controller {
		return
	}
	opps := gs.Opponents(perm.Controller)
	if len(opps) == 0 {
		return
	}
	target := opps[0]
	bestLife := gs.Seats[target].Life
	for _, o := range opps[1:] {
		if gs.Seats[o].Life > bestLife {
			bestLife = gs.Seats[o].Life
			target = o
		}
	}
	gameengine.DealDamage(gs, target, 1, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":        perm.Controller,
		"target_seat": target,
	})
	_ = gs.CheckEnd()
}

func nivMizzetFiremindActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "niv_mizzet_firemind_tap_draw"
	if gs == nil || src == nil {
		return
	}
	if src.Tapped {
		return
	}
	src.Tapped = true
	drawn := drawOne(gs, src.Controller, src.Card.DisplayName())
	drawnName := ""
	if drawn != nil {
		drawnName = drawn.DisplayName()
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":  src.Controller,
		"drawn": drawnName,
	})
}
