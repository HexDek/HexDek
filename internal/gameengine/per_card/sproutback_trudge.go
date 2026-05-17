package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSproutbackTrudge wires Sproutback Trudge
// (Muninn parser-gap #104, ~1.4K hits).
//
// Oracle text (Scryfall, verified 2026-05-17):
//
//	Sproutback Trudge — {7}{G}{G} Creature — Fungus Beast 4/5
//	This spell costs {X} less to cast, where X is the amount of life
//	you gained this turn.
//	Trample
//	At the beginning of your end step, if you gained life this turn,
//	you may cast this creature from your graveyard.
//
// Implementation:
//   - ETB hook is the only battlefield event we can hang behavior on.
//     Both the cost-reduction and the graveyard end-step cast option
//     live outside the battlefield: the former is a cast-pipeline
//     concern (alternative cost), the latter is the same graveyard-side
//     phase-trigger gap noted on Ichorid (registry.go iterates
//     Seat.Battlefield only). emitPartial both so Muninn sees the snippet
//     resolve to a known stub instead of an unhandled gap.
func registerSproutbackTrudge(r *Registry) {
	r.OnETB("Sproutback Trudge", sproutbackTrudgeETB)
}

func sproutbackTrudgeETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emit(gs, "sproutback_trudge_etb", "Sproutback Trudge", map[string]interface{}{
		"seat": perm.Controller,
	})
	emitPartial(gs, "sproutback_trudge", "Sproutback Trudge",
		"cost_reduction_by_life_gained_this_turn_requires_cast_pipeline_hook")
	emitPartial(gs, "sproutback_trudge", "Sproutback Trudge",
		"end_step_cast_from_graveyard_if_life_gained_requires_graveyard_phase_trigger_dispatch")
}
