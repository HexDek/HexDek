package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerJacobFrye wires Jacob Frye.
//
// Oracle text:
//
//	Partner with Evie Frye
//	Whenever one or more Assassins you control deal combat damage to
//	a player, exile up to one target Assassin card or card with
//	freerunning from your graveyard. If you do, copy it. You may
//	cast the copy.
//
// Partner-with tutoring and graveyard-cast-copy are non-trivial — emitPartial.
func registerJacobFrye(r *Registry) {
	r.OnETB("Jacob Frye", jacobFryeETB)
}

func jacobFryeETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "jacob_frye_partial"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"partner_tutor_and_assassin_combat_copy_cast_unimplemented")
}
