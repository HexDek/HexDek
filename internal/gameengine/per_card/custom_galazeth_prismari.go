package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGalazethPrismariCustom wires the ETB Treasure for Galazeth
// Prismari. The auto-generated stub registerGalazethPrismari in
// gen_galazeth_prismari.go remains an inert breadcrumb.
//
// Oracle text (Strixhaven, {2}{U}{R}):
//
//	Flying
//	When Galazeth Prismari enters, create a Treasure token.
//	{T}: Add one mana of any color. Spend this mana only to cast an
//	instant or sorcery spell.
//
// Implementation:
//   - OnETB: create one Treasure token for the controller.
//   - The "{T}: Add mana, instant/sorcery only" mana ability is a static
//     restriction the AST mana system needs. Per-card hooks aren't on
//     the mana-add path; we emitPartial.
//   - Flying — AST keyword pipeline.
func registerGalazethPrismariCustom(r *Registry) {
	r.OnETB("Galazeth Prismari", galazethPrismariCustomETB)
}

func galazethPrismariCustomETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "galazeth_prismari_etb_treasure"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seatIdx := perm.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	gameengine.CreateTreasureToken(gs, seatIdx)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   seatIdx,
		"tokens": 1,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"tap_for_any_color_instant_or_sorcery_only_mana_restriction_not_modeled")
}
