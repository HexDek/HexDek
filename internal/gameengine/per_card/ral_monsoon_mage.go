package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRalMonsoonMage wires Ral, Monsoon Mage // Ral, Leyline Prodigy.
//
// Front face — Ral, Monsoon Mage:
//
//	Instant and sorcery spells you cast cost {1} less to cast.
//	Whenever you cast an instant or sorcery spell during your turn,
//	flip a coin. If you lose the flip, Ral deals 1 damage to you. If
//	you win the flip, you may exile Ral. If you do, return him to the
//	battlefield transformed under his owner's control.
//
// The cost-reduction static effect is best handled by the AST. The
// coinflip transform pipeline is a parser gap. We register a stub so
// the card is recognized and report the gap.
func registerRalMonsoonMage(r *Registry) {
	r.OnETB("Ral, Monsoon Mage", ralMonsoonMageETB)
}

func ralMonsoonMageETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "ral_monsoon_mage_etb"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"coinflip_transform_to_leyline_prodigy_unimplemented")
}
