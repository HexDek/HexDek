package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTheDawningArchaic wires The Dawning Archaic.
//
// Oracle text:
//
//	{10}
//	Legendary Creature — Avatar
//	This spell costs {1} less to cast for each instant and sorcery card
//	  in your graveyard.
//	Reach
//	Whenever The Dawning Archaic attacks, you may cast target instant or
//	  sorcery card from your graveyard without paying its mana cost. If
//	  that spell would be put into your graveyard, exile it instead.
//
// Implementation:
//   - Cost reduction is done by the cost_modifiers.go scanner via
//     oracle-text inspection (a known pattern); not handled here.
//   - Reach via AST.
//   - Attack trigger: emitPartial — casting cards from graveyard without
//     paying their mana cost requires the spell-replay path which isn't
//     surfaced cleanly to per_card today.
func registerTheDawningArchaic(r *Registry) {
	r.OnETB("The Dawning Archaic", theDawningArchaicETB)
	r.OnTrigger("The Dawning Archaic", "creature_attacks", theDawningArchaicAttack)
}

func theDawningArchaicETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "the_dawning_archaic_etb", perm.Card.DisplayName(),
		"static_cost_reduction_per_is_in_graveyard_and_attack_replay_partial")
}

func theDawningArchaicAttack(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "the_dawning_archaic_attack_replay"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atkPerm, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atkPerm == nil || atkPerm != perm {
		return
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"may_cast_target_is_from_graveyard_without_paying_mana_partial")
}
