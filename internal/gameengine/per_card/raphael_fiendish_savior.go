package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRaphaelFiendishSavior wires Raphael, Fiendish Savior.
//
// Oracle text:
//
//	Flying
//	Other Demons, Devils, Imps, and Tieflings you control get +1/+1
//	and have lifelink.
//	At the beginning of each end step, if a creature card was put
//	into your graveyard from anywhere this turn, create a 1/1 red
//	Devil creature token with "When this token dies, it deals 1
//	damage to any target."
//
// Tribal anthem is a static effect handled by AST. The end-step
// observer + per-turn graveyard tracking is left as a parser gap.
func registerRaphaelFiendishSavior(r *Registry) {
	r.OnETB("Raphael, Fiendish Savior", raphaelETB)
}

func raphaelETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "raphael_etb"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"end_step_creature_died_devil_token_unimplemented")
}
