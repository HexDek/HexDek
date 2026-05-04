package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerLonisCryptozoologist wires Lonis, Cryptozoologist.
//
// Oracle text:
//
//	Whenever another nontoken creature you control enters, investigate.
//	{T}, Sacrifice X Clues: Target opponent reveals the top X cards of
//	their library. You may put a nonland permanent card with mana value
//	X or less from among them onto the battlefield under your control.
//	That player puts the rest on the bottom of their library in a random
//	order.
//
// Implementation: ETB-trigger creates a clue token. The activated
// {T}, sac X Clues ability is non-trivial — emitPartial.
func registerLonisCryptozoologist(r *Registry) {
	r.OnTrigger("Lonis, Cryptozoologist", "nonland_permanent_etb", lonisETB)
	r.OnActivated("Lonis, Cryptozoologist", lonisActivated)
}

func lonisETB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "lonis_investigate"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	enteringSeat, _ := ctx["controller_seat"].(int)
	if enteringSeat != perm.Controller {
		return
	}
	enteringCard, _ := ctx["card"].(*gameengine.Card)
	if enteringCard == nil {
		return
	}
	if !cardHasType(enteringCard, "creature") {
		return
	}
	if cardHasType(enteringCard, "token") {
		return
	}
	if enteringCard == perm.Card {
		return
	}
	gameengine.CreateClueToken(gs, perm.Controller)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"creature": enteringCard.DisplayName(),
	})
}

func lonisActivated(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "lonis_sac_clues_steal"
	if gs == nil || src == nil {
		return
	}
	emitPartial(gs, slug, src.Card.DisplayName(),
		"sac_x_clues_reveal_top_x_steal_permanent_unimplemented")
}
