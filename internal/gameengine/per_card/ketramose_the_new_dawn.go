package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKetramoseTheNewDawn wires Ketramose, the New Dawn.
//
// Oracle text:
//
//	Menace, lifelink, indestructible
//	Ketramose can't attack or block unless there are seven or more cards
//	in exile.
//	Whenever one or more cards are put into exile from graveyards and/or
//	the battlefield during your turn, you draw a card and lose 1 life.
//
// Implementation: zone_change observer for moves to exile from graveyard
// or battlefield. Triggers only on Ketramose's controller's turn. Effect
// is per-event (one trigger per zone-change event), matching the "one or
// more" wording. Attack/block restriction is engine-level — emitPartial
// for that clause.
func registerKetramoseTheNewDawn(r *Registry) {
	r.OnTrigger("Ketramose, the New Dawn", "zone_change", ketramoseZoneChange)
}

func ketramoseZoneChange(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ketramose_exile_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	if gs.Active != perm.Controller {
		return
	}
	to, _ := ctx["to"].(string)
	from, _ := ctx["from"].(string)
	if to != "exile" {
		return
	}
	if from != "graveyard" && from != "battlefield" {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	seat.Life--
	gs.LogEvent(gameengine.Event{
		Kind:   "lose_life",
		Seat:   perm.Controller,
		Target: perm.Controller,
		Source: perm.Card.DisplayName(),
		Amount: 1,
		Details: map[string]interface{}{
			"reason": "ketramose_self_pay",
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
		"from": from,
	})
	_ = gs.CheckEnd()
}
