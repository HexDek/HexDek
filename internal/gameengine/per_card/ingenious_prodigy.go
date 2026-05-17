package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerIngeniousProdigy wires Ingenious Prodigy (Muninn parser-gap
// #54, 14,325 hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{X}{U}
//	Creature — Human Wizard
//	Skulk (This creature can't be blocked by creatures with greater
//	power.)
//	This creature enters with X +1/+1 counters on it.
//	At the beginning of your upkeep, if this creature has one or more
//	+1/+1 counters on it, you may remove a +1/+1 counter from it. If
//	you do, draw a card.
//
// Implementation:
//   - Skulk is engine-side (AST keyword pipeline).
//   - ETB: read perm.Flags["x_paid"] and AddCounter("+1/+1", x).
//     If x_paid is missing/0, the prodigy enters as a 0/0 + 0 counters
//     and dies to SBAs — that's the rules-correct outcome.
//   - Upkeep trigger gated on active_seat == controller AND
//     Counters["+1/+1"] >= 1. "you may" — Hat always accepts (draw
//     a card for a counter is upside when counters > 1; even at 1
//     counter, drawing now beats letting the prodigy linger as a 1/1).
//   - Remove one +1/+1 counter, draw one card via MoveCard from library
//     top to hand. If library is empty, set AttemptedEmptyDraw (the
//     SBA harness handles loss-on-draw).
func registerIngeniousProdigy(r *Registry) {
	r.OnETB("Ingenious Prodigy", ingeniousProdigyETB)
	// "upkeep" aliases to "upkeep_controller" — registering only one
	// prevents double-fire.
	r.OnTrigger("Ingenious Prodigy", "upkeep", ingeniousProdigyUpkeep)
}

func ingeniousProdigyETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "ingenious_prodigy_etb_x_counters"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	x := 0
	if perm.Flags != nil {
		x = perm.Flags["x_paid"]
	}
	if x > 0 {
		perm.AddCounter("+1/+1", x)
		gs.InvalidateCharacteristicsCache()
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"x":        x,
		"counters": perm.Counters["+1/+1"],
	})
	if x <= 0 {
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"x_paid_not_threaded_into_etb_hook_enters_at_zero_counters")
	}
}

func ingeniousProdigyUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ingenious_prodigy_upkeep_remove_counter_draw"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	activeSeat, ok := ctx["active_seat"].(int)
	if !ok || activeSeat != perm.Controller {
		return
	}
	if perm.Counters == nil || perm.Counters["+1/+1"] < 1 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
		})
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	perm.AddCounter("+1/+1", -1)
	gs.InvalidateCharacteristicsCache()
	drew := false
	if len(seat.Library) > 0 {
		c := seat.Library[0]
		gameengine.MoveCard(gs, c, perm.Controller, "library", "hand", slug)
		drew = true
		gs.LogEvent(gameengine.Event{
			Kind:   "draw",
			Seat:   perm.Controller,
			Source: perm.Card.DisplayName(),
			Details: map[string]interface{}{
				"reason": slug,
			},
		})
	} else {
		seat.AttemptedEmptyDraw = true
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"triggered": true,
		"drew":      drew,
		"remaining": perm.Counters["+1/+1"],
	})
}
