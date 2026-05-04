package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerNicolBolasTheRavager wires Nicol Bolas, the Ravager // the Arisen.
//
// Front face — Nicol Bolas, the Ravager:
//
//	Flying
//	When Nicol Bolas enters, each opponent discards a card.
//	{4}{U}{B}{R}: Exile Nicol Bolas, then return him to the
//	battlefield transformed under his owner's control. Activate only
//	as a sorcery.
//
// We wire the ETB discard. The transform activation is left as a
// parser gap (no DFC-transform-via-activation pipeline).
func registerNicolBolasTheRavager(r *Registry) {
	r.OnETB("Nicol Bolas, the Ravager", nicolBolasRavagerETB)
}

func nicolBolasRavagerETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "nicol_bolas_ravager_discard"
	if gs == nil || perm == nil {
		return
	}
	for _, opp := range gs.Opponents(perm.Controller) {
		s := gs.Seats[opp]
		if s == nil || s.Lost {
			continue
		}
		gameengine.DiscardN(gs, opp, 1, "nicol_bolas_random")
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(), "transform_activation_to_arisen_unimplemented")
}
