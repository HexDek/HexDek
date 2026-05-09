package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKarumonixTheRatKingCustom wires Karumonix's ETB poison-dropper
// per Rat the controller has on the battlefield (per dev/etb-stub-handlers
// spec — a stylized variant of the printed "look at top 5, take Rats"
// effect that's a better fit for the engine's hard mechanics).
//
// Spec:
//
//	When Karumonix enters, each opponent gets a poison counter for each
//	Rat you control.
//
// Implementation:
//   - Walk the controller's battlefield. Count creatures with the "rat"
//     subtype (Karumonix herself counts — she's a Rat).
//   - For each living opponent, increment Seat.PoisonCounters by that
//     count. Engine SBA detects poison >= 10 and ends the game.
func registerKarumonixTheRatKingCustom(r *Registry) {
	r.OnETB("Karumonix, the Rat King", karumonixETBPoison)
}

func karumonixETBPoison(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "karumonix_etb_poison_per_rat"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	rats := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil || !p.IsCreature() {
			continue
		}
		if cardHasSubtype(p.Card, "rat") {
			rats++
		}
	}
	if rats == 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat": perm.Controller,
			"rats": 0,
			"note": "no_rats_no_poison",
		})
		return
	}
	dropped := 0
	for i, s := range gs.Seats {
		if s == nil || i == perm.Controller || s.Lost || s.LeftGame {
			continue
		}
		s.PoisonCounters += rats
		gs.LogEvent(gameengine.Event{
			Kind:   "poison_counter_added",
			Seat:   i,
			Source: perm.Card.DisplayName(),
			Amount: rats,
			Details: map[string]interface{}{
				"slug":   slug,
				"reason": "karumonix_etb_per_rat",
			},
		})
		dropped++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":            perm.Controller,
		"rats":            rats,
		"opponents_hit":   dropped,
		"poison_per_opp":  rats,
	})
}
