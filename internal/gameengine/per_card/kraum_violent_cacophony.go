package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKraumViolentCacophony wires Kraum, Violent Cacophony (the
// SLD partner reprint variant — the second-spell trigger now applies to
// YOUR spells, putting a counter on Kraum and drawing a card).
//
// Oracle text:
//
//	Flying
//	Whenever you cast your second spell each turn, put a +1/+1 counter
//	on Kraum and draw a card.
//
// Implementation: listen on "spell_cast" and gate on caster_seat ==
// controller && SpellsCastThisTurn == 2.
func registerKraumViolentCacophony(r *Registry) {
	r.OnTrigger("Kraum, Violent Cacophony", "spell_cast", kraumViolentCacophonyTrigger)
}

func kraumViolentCacophonyTrigger(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kraum_violent_cacophony_2nd_spell"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, ok := ctx["caster_seat"].(int)
	if !ok || casterSeat != perm.Controller {
		return
	}
	if casterSeat < 0 || casterSeat >= len(gs.Seats) || gs.Seats[casterSeat] == nil {
		return
	}
	if gs.Seats[casterSeat].SpellsCastThisTurn != 2 {
		return
	}
	perm.AddCounter("+1/+1", 1)
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"counters": perm.Counters["+1/+1"],
	})
}
