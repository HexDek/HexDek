package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSythisHarvestsHand wires Sythis, Harvest's Hand.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Whenever you cast an enchantment spell, you gain 1 life and draw a
//	  card.
//
// Implementation:
//   - "spell_cast": gate on caster_seat == perm.Controller and the cast
//     card is an enchantment. GainLife(1) + drawOne.
func registerSythisHarvestsHand(r *Registry) {
	r.OnTrigger("Sythis, Harvest's Hand", "spell_cast", sythisOnSpellCast)
}

func sythisOnSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "sythis_enchantment_cast"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	caster, _ := ctx["caster_seat"].(int)
	if caster != perm.Controller {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil || !cardHasType(card, "enchantment") {
		return
	}
	gameengine.GainLife(gs, perm.Controller, 1, perm.Card.DisplayName())
	drawn := drawOne(gs, perm.Controller, perm.Card.DisplayName())
	drawnName := ""
	if drawn != nil {
		drawnName = drawn.DisplayName()
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"spell": card.DisplayName(),
		"drawn": drawnName,
	})
}
