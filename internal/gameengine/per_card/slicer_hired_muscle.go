package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSlicerHiredMuscle wires Slicer, Hired Muscle //
// Slicer, High-Speed Antagonist.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//   Slicer, Hired Muscle (front, Robot):
//	More Than Meets the Eye {2}{R}
//	Double strike, haste
//	At the beginning of each opponent's upkeep, you may have that
//	  player gain control of Slicer until end of turn. If you do,
//	  untap Slicer, goad it, and it can't be sacrificed this turn.
//	  If you don't, convert it.
//
//   Slicer, High-Speed Antagonist (back, Vehicle):
//	Living metal
//	First strike, haste
//	Whenever Slicer deals combat damage to a player, convert it at
//	  end of combat.
//
// Both faces involve control-changes and convert (transform) effects
// outside the per-card pipeline. Register an ETB partial-flag marker.
func registerSlicerHiredMuscle(r *Registry) {
	r.OnETB("Slicer, Hired Muscle", slicerHiredMuscleETB)
	r.OnTrigger("Slicer, Hired Muscle", "upkeep_opponent", slicerOpponentUpkeep)
	r.OnTrigger("Slicer, Hired Muscle", "combat_damage_player", slicerCombatDamage)
}

func slicerHiredMuscleETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "slicer_hired_muscle_static", perm.Card.DisplayName(),
		"control_change_and_convert_transform_not_modeled")
}

func slicerOpponentUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "slicer_opponent_upkeep_handoff", perm.Card.DisplayName(),
		"opt_in_control_change_or_convert_branch_not_modeled")
}

func slicerCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "slicer_combat_damage_convert", perm.Card.DisplayName(),
		"convert_at_end_of_combat_after_damage_not_modeled")
}
