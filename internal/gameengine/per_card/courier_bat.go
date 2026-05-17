package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCourierBat wires Courier Bat (Muninn parser-gap #52,
// 15,457 hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{2}{B}
//	Creature — Bat
//	Flying
//	When this creature enters, if you gained life this turn, return up
//	to one target creature card from your graveyard to your hand.
//
// Implementation:
//   - Flying is AST-engine-side.
//   - ETB gated on seat.Turn.LifeGained > 0 (the canonical per-turn
//     life-gain accumulator — see state.go's GainLife). The check
//     happens at resolve time, so any life gained earlier this turn
//     (including from the Bat's own caster gaining life via another
//     trigger that resolved first) counts.
//   - "Up to one target creature card from your graveyard" — Hat
//     policy: pick the highest-CMC creature card (best to re-cast).
//     If no creature in graveyard, the trigger fizzles (no target).
func registerCourierBat(r *Registry) {
	r.OnETB("Courier Bat", courierBatETB)
}

func courierBatETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "courier_bat_etb_return"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	if seat.Turn.LifeGained <= 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "no_life_gained_this_turn",
		})
		return
	}
	var best *gameengine.Card
	bestCMC := -1
	for _, c := range seat.Graveyard {
		if c == nil || !cardHasType(c, "creature") {
			continue
		}
		cmc := cardCMC(c)
		if cmc > bestCMC {
			bestCMC = cmc
			best = c
		}
	}
	if best == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":         perm.Controller,
			"triggered":    true,
			"life_gained":  seat.Turn.LifeGained,
			"target":       "none_in_graveyard",
		})
		return
	}
	gameengine.MoveCard(gs, best, perm.Controller, "graveyard", "hand", slug)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":        perm.Controller,
		"triggered":   true,
		"returned":    best.DisplayName(),
		"life_gained": seat.Turn.LifeGained,
	})
}
