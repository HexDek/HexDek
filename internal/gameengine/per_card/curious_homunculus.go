package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCuriousHomunculus wires Curious Homunculus // Voracious Reader
// (Muninn parser-gap #66, ~9.8K hits).
//
// Oracle text (Scryfall, verified 2026-05-17):
//
//	Curious Homunculus — {U} Creature — Homunculus 1/2
//	{T}: Add {C}. Spend this mana only to cast an instant or sorcery
//	spell.
//	At the beginning of your upkeep, if there are three or more instant
//	and/or sorcery cards in your graveyard, transform Curious Homunculus.
//
//	Voracious Reader — Creature — Eldrazi Homunculus 3/2
//	Prowess
//	Instant and sorcery spells you cast cost {1} less to cast.
//
// Implementation:
//   - The mana ability ({T}: Add {C}, "spend only to cast I/S") is an
//     engine concern (mana_restriction), not a per-card handler.
//   - upkeep_controller: if controller has ≥3 instant/sorcery cards in
//     graveyard, emit a transform-eligible event. Actual DFC transform
//     requires gameengine.TransformPermanent (or equivalent) wired with
//     a back-face Card load — emitPartial flags the missing engine call.
//   - Static cost reduction on Voracious Reader side is layered cost
//     math — partial.
func registerCuriousHomunculus(r *Registry) {
	r.OnTrigger("Curious Homunculus", "upkeep_controller", curiousHomunculusUpkeep)
}

func curiousHomunculusUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "curious_homunculus_upkeep_transform_check"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	activeSeat, ok := ctx["active_seat"].(int)
	if !ok || activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	n := 0
	for _, c := range seat.Graveyard {
		if c == nil {
			continue
		}
		if cardHasType(c, "instant") || cardHasType(c, "sorcery") {
			n++
		}
	}
	if n < 3 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":             perm.Controller,
			"triggered":        false,
			"is_count":         n,
			"reason":           "fewer_than_three_is_in_graveyard",
		})
		return
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"triggered": true,
		"is_count":  n,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"transform_to_voracious_reader_engine_call_unimplemented")
}
