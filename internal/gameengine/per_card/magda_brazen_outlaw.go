package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMagdaBrazenOutlaw wires Magda, Brazen Outlaw.
//
// Oracle text:
//
//	Other Dwarves you control get +1/+0.
//	Whenever a Dwarf you control becomes tapped, create a Treasure token.
//	Sacrifice five Treasures: Search your library for an artifact or
//	Dragon card, put that card onto the battlefield, then shuffle.
//
// Implementation: listen on land_tapped_for_mana / artifact_tapped_for_mana
// is too narrow — we want any tap. Use becomes_tapped if the engine fires
// it; if not, hook tap-for-mana. The +1/+0 anthem is a static effect.
// The sacrifice ability is non-trivial library search — emitPartial.
func registerMagdaBrazenOutlaw(r *Registry) {
	r.OnTrigger("Magda, Brazen Outlaw", "becomes_tapped", magdaBrazenBecomesTapped)
	r.OnActivated("Magda, Brazen Outlaw", magdaBrazenActivated)
}

func magdaBrazenBecomesTapped(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "magda_brazen_dwarf_treasure"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	tappedSeat, _ := ctx["controller_seat"].(int)
	if tappedSeat != perm.Controller {
		return
	}
	tappedCard, _ := ctx["card"].(*gameengine.Card)
	if tappedCard == nil || !cardHasType(tappedCard, "dwarf") {
		return
	}
	gameengine.CreateTreasureToken(gs, perm.Controller)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"dwarf": tappedCard.DisplayName(),
	})
}

func magdaBrazenActivated(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "magda_brazen_sac_5_treasures_tutor"
	if gs == nil || src == nil {
		return
	}
	emitPartial(gs, slug, src.Card.DisplayName(),
		"sac_5_treasures_tutor_artifact_or_dragon_unimplemented")
}
