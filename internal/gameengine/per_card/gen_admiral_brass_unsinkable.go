package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAdmiralBrassUnsinkable wires Admiral Brass, Unsinkable.
//
// Oracle text:
//
//   When Admiral Brass enters, mill four cards.
//   At the beginning of combat on your turn, you may return target Pirate creature card from your graveyard to the battlefield with a finality counter on it. It has base power and toughness 4/4. It gains haste until end of turn. (If a creature with a finality counter on it would die, exile it instead.)
//
// Implementation:
//   - ETB: mill 4 from controller's library.
//   - Combat-begin Pirate reanimation: emitPartial (begin-of-combat
//     observer + base-P/T override are not modeled here).
func registerAdmiralBrassUnsinkable(r *Registry) {
	r.OnETB("Admiral Brass, Unsinkable", admiralBrassUnsinkableETB)
}

func admiralBrassUnsinkableETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "admiral_brass_etb_mill4"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	milled := 0
	for i := 0; i < 4 && len(seat.Library) > 0; i++ {
		c := seat.Library[0]
		gameengine.MoveCard(gs, c, perm.Controller, "library", "graveyard", "admiral_brass_mill")
		milled++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"milled": milled,
	})
	emitPartial(gs, "admiral_brass_combat_pirate_reanimate", perm.Card.DisplayName(),
		"begin_of_combat_pirate_reanimate_with_base_pt_4_4_unimplemented")
}
