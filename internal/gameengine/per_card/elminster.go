package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerElminster wires Elminster (planeswalker).
//
// Oracle text:
//
//	Whenever you scry, the next instant or sorcery spell you cast this
//	turn costs {X} less to cast, where X is the number of cards looked
//	at while scrying this way.
//	+2: Draw a card, then scry 2.
//	−3: Exile the top card of your library. Create a number of 1/1 blue
//	    Faerie Dragon creature tokens with flying equal to that card's
//	    mana value.
//	Elminster can be your commander.
//
// Planeswalker activations are out of scope; emitPartial.
func registerElminster(r *Registry) {
	r.OnETB("Elminster", elminsterStub)
}

func elminsterStub(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "elminster_planeswalker_partial"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"planeswalker_activations_and_scry_cost_reduction_unimplemented")
}
