package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerNivMizzetVisionary wires Niv-Mizzet, Visionary.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Flying
//	You have no maximum hand size.
//	Whenever a source you control deals noncombat damage to an opponent,
//	  you draw that many cards.
//
// Implementation:
//   - "noncombat_damage_to_player": filter to source.Controller ==
//     perm.Controller and target seat is an opponent. Draw the damage
//     amount. The engine doesn't yet expose a dedicated noncombat_damage
//     event, so we listen to "damage" with a "combat" flag absent.
//   - "no maximum hand size" handled at the SBA layer — emitPartial.
func registerNivMizzetVisionary(r *Registry) {
	r.OnETB("Niv-Mizzet, Visionary", nivMizzetVisionaryETB)
	r.OnTrigger("Niv-Mizzet, Visionary", "noncombat_damage_to_player", nivMizzetVisionaryNoncombat)
}

func nivMizzetVisionaryETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "niv_mizzet_visionary_no_max_hand", perm.Card.DisplayName(),
		"no_maximum_hand_size_static_not_modeled")
}

func nivMizzetVisionaryNoncombat(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "niv_mizzet_visionary_noncombat_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	srcSeat, _ := ctx["source_seat"].(int)
	if srcSeat != perm.Controller {
		return
	}
	tgtSeat, _ := ctx["target_seat"].(int)
	if tgtSeat == perm.Controller || tgtSeat < 0 || tgtSeat >= len(gs.Seats) {
		return
	}
	amt, _ := ctx["amount"].(int)
	if amt <= 0 {
		return
	}
	drawn := 0
	for i := 0; i < amt; i++ {
		if drawOne(gs, perm.Controller, perm.Card.DisplayName()) != nil {
			drawn++
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":        perm.Controller,
		"target_seat": tgtSeat,
		"damage":      amt,
		"drawn":       drawn,
	})
}
