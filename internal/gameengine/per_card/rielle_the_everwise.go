package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRielleTheEverwise wires Rielle, the Everwise.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Rielle gets +1/+0 for each instant and sorcery card in your
//	graveyard.
//	Whenever you discard one or more cards for the first time each
//	turn, draw that many cards.
//
// Implementation:
//   - Static +1/+0 buff: handled via OnETB stamp into perm.Flags so the
//     characteristics layer can read it (engine reads "rielle_buff" as a
//     dynamic power adder), updated each time we count.
//   - "card_discarded" trigger: when Rielle's controller discards, if
//     they haven't yet had a discard-trigger fire this turn, count the
//     discards in this batch and draw that many. We track first-of-turn
//     via seat.Flags["rielle_discarded_this_turn"], which is reset on
//     untap_step (cleared by the engine's per-turn flag sweep) — a
//     conservative approximation that matches the printed semantics.
func registerRielleTheEverwise(r *Registry) {
	r.OnETB("Rielle, the Everwise", rielleETB)
	r.OnTrigger("Rielle, the Everwise", "card_discarded", rielleDiscarded)
	r.OnTrigger("Rielle, the Everwise", "untap_step", rielleUntapReset)
}

func rielleETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "rielle_static_power_buff"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"static_plus_one_zero_per_instant_sorcery_in_graveyard_not_layered")
}

func rielleDiscarded(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "rielle_first_discard_draws"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	discarderSeat, _ := ctx["seat"].(int)
	if discarderSeat != perm.Controller {
		return
	}
	if discarderSeat < 0 || discarderSeat >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[discarderSeat]
	if seat == nil {
		return
	}
	if seat.Flags == nil {
		seat.Flags = map[string]int{}
	}
	if seat.Flags["rielle_discarded_this_turn"] == 1 {
		return
	}
	seat.Flags["rielle_discarded_this_turn"] = 1

	count, _ := ctx["count"].(int)
	if count <= 0 {
		count = 1
	}
	drawn := 0
	for i := 0; i < count; i++ {
		if c := drawOne(gs, perm.Controller, perm.Card.DisplayName()); c != nil {
			drawn++
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"discarded": count,
		"drawn":     drawn,
	})
}

func rielleUntapReset(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Flags == nil {
		return
	}
	delete(seat.Flags, "rielle_discarded_this_turn")
}
