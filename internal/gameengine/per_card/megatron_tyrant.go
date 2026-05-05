package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMegatronTyrant wires Megatron, Tyrant // Megatron, Destructive
// Force (DFC, Transformer "convert" mechanic).
//
// Oracle text — front (Megatron, Tyrant):
//
//	More Than Meets the Eye {1}{R}{W}{B}
//	Your opponents can't cast spells during combat.
//	At the beginning of each of your postcombat main phases, you may
//	convert Megatron. If you do, add {C} for each 1 life your opponents
//	have lost this turn.
//
// Oracle text — back (Megatron, Destructive Force):
//
//	Living metal
//	Whenever Megatron attacks, you may sacrifice another artifact. When
//	you do, Megatron deals damage equal to the sacrificed artifact's
//	mana value to target creature. If excess damage would be dealt to
//	that creature this way, instead that damage is dealt to that
//	creature's controller and you convert Megatron.
//
// Implementation: convert (transform) and the static combat-cast
// restriction both require engine plumbing not currently available.
// emitPartial.
func registerMegatronTyrant(r *Registry) {
	r.OnTrigger("Megatron, Tyrant", "postcombat_main_controller", megatronPostcombat)
	r.OnTrigger("Megatron, Tyrant // Megatron, Destructive Force", "postcombat_main_controller", megatronPostcombat)
	r.OnTrigger("Megatron, Destructive Force", "attacks", megatronBackAttacks)
	r.OnTrigger("Megatron, Tyrant // Megatron, Destructive Force", "attacks", megatronBackAttacks)
}

func megatronPostcombat(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "megatron_tyrant_convert"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"convert_transform_and_colorless_mana_payoff_unimplemented")
}

func megatronBackAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "megatron_destructive_force_attack"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"sacrifice_artifact_damage_excess_convert_unimplemented")
}
