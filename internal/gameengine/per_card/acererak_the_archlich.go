package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAcererakTheArchlichETB wires Acererak the Archlich's enter
// trigger. The attack trigger half is owned by era3_batch.go's
// registerAcererakEra3 — both registrations coexist on the same card key
// (ETB + creature_attacks).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	When Acererak enters, if you haven't completed Tomb of Annihilation,
//	return Acererak to its owner's hand and venture into the dungeon.
//	Whenever Acererak attacks, for each opponent, you create a 2/2 black
//	Zombie creature token unless that player sacrifices a creature of
//	their choice.
//
// Implementation (Muninn gap #9 — 95K hits; attack-trigger half already
// covered in internal/gameengine/per_card/era3_batch.go):
//   - OnETB: read seat.Flags["dungeon_completed"]. If zero, BouncePermanent
//     Acererak to its owner's hand and call VentureIntoDungeon. If the
//     dungeon is already completed, leave Acererak on the battlefield
//     (the conditional clause stops triggering — CR §603.4 intervening
//     "if" check).
//   - The engine's dungeon model is a simplified 4-room ladder (see
//     internal/gameengine/keywords_misc.go::VentureIntoDungeon); "Tomb of
//     Annihilation" specifically is not tracked separately, so any
//     completed dungeon satisfies the gate. Flagged via emitPartial.
//   - Dungeon room choice (Tomb has branching rooms) is approximated by
//     the linear advance baked into VentureIntoDungeon. Flagged via
//     emitPartial.
func registerAcererakTheArchlichETB(r *Registry) {
	r.OnETB("Acererak the Archlich", acererakTheArchlichETB)
}

func acererakTheArchlichETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "acererak_etb_venture"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil || s.Lost {
		return
	}

	completed := 0
	if s.Flags != nil {
		completed = s.Flags["dungeon_completed"]
	}
	if completed > 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":              seat,
			"dungeon_completed": completed,
			"bounced":           false,
		})
		return
	}

	// Bounce Acererak to its owner's hand BEFORE venturing — the venture
	// is a separate effect listed after the bounce in the oracle text,
	// but order doesn't matter for state since the dungeon advance lives
	// on the seat, not the permanent.
	gameengine.BouncePermanent(gs, perm, perm, "hand")

	room := gameengine.VentureIntoDungeon(gs, seat)

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":    seat,
		"bounced": true,
		"room":    room,
	})

	emitPartial(gs, slug, perm.Card.DisplayName(),
		"tomb_of_annihilation_specific_dungeon_choice_uses_engine_default_4_room_ladder")
}
