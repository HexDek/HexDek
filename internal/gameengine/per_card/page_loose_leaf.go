package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerPageLooseLeaf wires Page, Loose Leaf.
//
// Oracle text:
//
//	{T}: Add {C}.
//	Grandeur — Discard another card named Page, Loose Leaf: Reveal
//	cards from the top of your library until you reveal an instant or
//	sorcery card. Put that card into your hand and the rest on the
//	bottom of your library in a random order.
//
// Mana ability is handled by the engine's mana parser. The Grandeur
// activation is left as a parser gap (no grandeur dispatcher path).
func registerPageLooseLeaf(r *Registry) {
	r.OnETB("Page, Loose Leaf", pageLooseLeafETB)
}

func pageLooseLeafETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "page_loose_leaf_etb"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"grandeur_discard_for_instant_sorcery_tutor_unimplemented")
}
