package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerErebosGodOfTheDead wires Erebos, God of the Dead.
//
// Oracle text:
//
//   Indestructible
//   As long as your devotion to black is less than five, Erebos isn't a creature. (Each {B} in the mana costs of permanents you control counts toward your devotion to black.)
//   Your opponents can't gain life.
//   {1}{B}, Pay 2 life: Draw a card.
//
// Auto-generated activated ability handler.
func registerErebosGodOfTheDead(r *Registry) {
	r.OnActivated("Erebos, God of the Dead", erebosGodOfTheDeadActivate)
}

func erebosGodOfTheDeadActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "erebos_god_of_the_dead_activate"
	if gs == nil || src == nil {
		return
	}
	seatIdx := src.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[seatIdx]
	if seat == nil || seat.Lost {
		return
	}
	// Cost gate: {1}{B}, Pay 2 life. Verify the seat can pay both the
	// mana and life portions before debiting either, so a half-paid
	// activation can never give a free draw.
	const manaCost = 2 // {1}{B} → 2 generic-equivalent
	const lifeCost = 2
	if seat.Life <= lifeCost {
		// Must have STRICTLY more life than the cost so the seat doesn't
		// kill itself paying — CR §119.4 (cost can't reduce life below 0
		// and an activation only resolves if all costs are paid in full).
		emitFail(gs, slug, src.Card.DisplayName(), "insufficient_life", map[string]interface{}{
			"life":      seat.Life,
			"life_cost": lifeCost,
		})
		return
	}
	if !gameengine.PayGenericCost(gs, seat, manaCost, "activated", "erebos_activate", src.Card.DisplayName()) {
		emitFail(gs, slug, src.Card.DisplayName(), "insufficient_mana", map[string]interface{}{
			"mana_pool": seat.ManaPool,
			"mana_cost": manaCost,
		})
		return
	}
	gameengine.LoseLife(gs, seatIdx, lifeCost, src.Card.DisplayName())
	drawOne(gs, seatIdx, src.Card.DisplayName())
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":     seatIdx,
		"mana_paid": manaCost,
		"life_paid": lifeCost,
	})
}
