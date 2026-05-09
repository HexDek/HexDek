package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerArcadesCustom adds Arcades' defender-ETB draw trigger that
// the auto-generated static stub omits.
//
// Oracle text:
//
//	Flying, vigilance
//	Whenever a creature you control with defender enters, draw a card.
//	Each creature you control with defender assigns combat damage equal
//	to its toughness rather than its power and can attack as though it
//	didn't have defender.
//
// Flying / vigilance / the combat statics are AST/engine territory. The
// ETB-draw is the load-bearing payoff for Arcades wall decks and lives
// here.
func registerArcadesCustom(r *Registry) {
	r.OnTrigger("Arcades, the Strategist", "permanent_etb", arcadesDefenderETBDraw)
}

func arcadesDefenderETBDraw(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "arcades_defender_etb_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	entered, _ := ctx["perm"].(*gameengine.Permanent)
	if entered == nil || entered.Card == nil {
		return
	}
	if entered.Controller != perm.Controller {
		return
	}
	if !entered.IsCreature() {
		return
	}
	if !entered.HasKeyword("defender") {
		return
	}
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          perm.Controller,
		"defender_card": entered.Card.DisplayName(),
	})
}
