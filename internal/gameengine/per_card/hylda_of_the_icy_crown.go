package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerHyldaOfTheIcyCrown wires Hylda of the Icy Crown.
//
// Oracle text:
//
//	Whenever you tap an untapped creature an opponent controls, you
//	may pay {1}. When you do, choose one —
//	  • Create a 4/4 white and blue Elemental creature token.
//	  • Put a +1/+1 counter on each creature you control.
//	  • Scry 2, then draw a card.
//
// "Whenever you tap an untapped creature an opponent controls" isn't a
// per-card hook the engine exposes generically — emitPartial.
func registerHyldaOfTheIcyCrown(r *Registry) {
	r.OnETB("Hylda of the Icy Crown", hyldaIcyCrownETB)
}

func hyldaIcyCrownETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "hylda_icy_crown_partial"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"opponent_creature_tap_trigger_unimplemented")
}
