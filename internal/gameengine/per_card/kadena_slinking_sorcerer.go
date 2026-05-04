package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKadenaSlinkingSorcerer wires Kadena, Slinking Sorcerer.
//
// Oracle text:
//
//	The first face-down creature spell you cast each turn costs {3}
//	less to cast.
//	Whenever a face-down creature you control enters, draw a card.
//
// Face-down / morph isn't tracked at the engine level — emitPartial.
func registerKadenaSlinkingSorcerer(r *Registry) {
	r.OnETB("Kadena, Slinking Sorcerer", kadenaETB)
}

func kadenaETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "kadena_partial"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"face_down_morph_cost_reduction_and_etb_draw_unimplemented")
}
