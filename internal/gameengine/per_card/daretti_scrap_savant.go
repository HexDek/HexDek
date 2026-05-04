package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerDarettiScrapSavant wires Daretti, Scrap Savant.
//
// Oracle text:
//
//	+2: Discard up to two cards, then draw that many cards.
//	−2: Sacrifice an artifact. If you do, return target artifact card
//	    from your graveyard to the battlefield.
//	−10: You get an emblem with "Whenever an artifact is put into your
//	     graveyard from the battlefield, return that card to the
//	     battlefield at the beginning of the next end step."
//	Daretti, Scrap Savant can be your commander.
//
// Planeswalker activations are out of scope; emitPartial.
func registerDarettiScrapSavant(r *Registry) {
	r.OnActivated("Daretti, Scrap Savant", darettiActivated)
}

func darettiActivated(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "daretti_scrap_savant_planeswalker"
	if gs == nil || src == nil {
		return
	}
	emitPartial(gs, slug, src.Card.DisplayName(),
		"planeswalker_activations_unimplemented")
}
