package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerHerigastEruptingNullkite wires Herigast, Erupting Nullkite.
//
// Oracle text:
//
//	Emerge {6}{R}{R} (You may cast this spell by sacrificing a creature
//	and paying the emerge cost reduced by that creature's mana value.)
//	When you cast this spell, you may exile your hand. If you do,
//	draw three cards.
//	Flying
//	Each creature spell you cast has emerge. The emerge cost is equal
//	to its mana cost.
//
// Emerge static grant and the cast-time exile-hand trigger require
// engine surfaces that don't exist — emitPartial.
func registerHerigastEruptingNullkite(r *Registry) {
	r.OnETB("Herigast, Erupting Nullkite", herigastETB)
}

func herigastETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "herigast_partial"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"emerge_grant_and_exile_hand_to_draw_three_unimplemented")
}
