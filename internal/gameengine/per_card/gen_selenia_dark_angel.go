package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSeleniaDarkAngel wires Selenia, Dark Angel.
//
// Oracle text:
//
//   Flying
//   Pay 2 life: Return Selenia to its owner's hand.
//
// Implementation: bounce Selenia to her owner's hand. The 2-life cost
// is presumed paid by the activation pipeline.
func registerSeleniaDarkAngel(r *Registry) {
	r.OnActivated("Selenia, Dark Angel", seleniaDarkAngelActivate)
}

func seleniaDarkAngelActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "selenia_dark_angel_bounce"
	if gs == nil || src == nil {
		return
	}
	gameengine.BouncePermanent(gs, src, src, "hand")
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat": src.Controller,
	})
}
