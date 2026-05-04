package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerShannaPurifyingBlade wires Shanna, Purifying Blade.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Lifelink
//	At the beginning of your end step, you may pay {X}. If you do,
//	draw X cards. X can't be greater than the amount of life you
//	gained this turn.
//
// Implementation:
//   - Lifelink handled by AST keyword pipeline.
//   - "end_step_controller": pay X up to the lesser of life gained this
//     turn and a heuristic mana cap (we approximate available mana as
//     min(turn,5) since we don't have access to the precise free-mana
//     bookkeeping in this trigger window). Draw that many cards.
//
// Life-gained-this-turn tracking lives on seat.Flags["life_gained_this_turn"]
// — a counter the engine bumps on each gain_life event. We emitPartial
// when the engine doesn't expose this counter (older builds).
func registerShannaPurifyingBlade(r *Registry) {
	r.OnTrigger("Shanna, Purifying Blade", "end_step", shannaPurifyingEndStep)
}

func shannaPurifyingEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "shanna_purifying_end_step_draw_x"
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
	gained := 0
	if seat.Flags != nil {
		gained = seat.Flags["life_gained_this_turn"]
	}
	if gained <= 0 {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_life_gained_this_turn", map[string]interface{}{
			"seat": perm.Controller,
		})
		return
	}
	// Heuristic mana ceiling: 4 (typical end-step mana availability).
	cap := gained
	if cap > 4 {
		cap = 4
	}
	drawn := 0
	for i := 0; i < cap; i++ {
		if c := drawOne(gs, perm.Controller, perm.Card.DisplayName()); c != nil {
			drawn++
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         perm.Controller,
		"life_gained":  gained,
		"x_paid":       cap,
		"drawn":        drawn,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"x_payment_not_actually_deducted_from_mana_pool")
}
