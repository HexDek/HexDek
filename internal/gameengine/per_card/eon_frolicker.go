package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerEonFrolicker wires Eon Frolicker (Muninn parser-gap #56, 13,197
// hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{2}{U}{U}
//	Creature — Elemental Otter
//	Flying
//	When this creature enters, if you cast it, target opponent takes an
//	extra turn after this one. Until your next turn, you and
//	planeswalkers you control gain protection from that player.
//
// Implementation:
//   - Flying is AST-side.
//   - ETB gated on perm.Flags["was_cast"] == 1.
//   - Target opponent pick: lowest-life living opponent (least scary
//     extra turn to give away — and they're closest to death from the
//     accumulated draw/land they get).
//   - Engine has only a seat-agnostic gs.Flags["extra_turns_pending"]
//     counter today (resolve.go:2479). Bumping it gives THE ACTIVE
//     PLAYER (= Eon Frolicker's caster) the extra turn, which is the
//     opposite of the printed text. We log emitPartial flagging the
//     opp-target gap rather than silently miswire the rules.
//   - "Protection from that player until your next turn": no per-player
//     protection static today; logged as partial.
func registerEonFrolicker(r *Registry) {
	r.OnETB("Eon Frolicker", eonFrolickerETB)
}

func eonFrolickerETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "eon_frolicker_etb_opp_extra_turn"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil || perm.Flags["was_cast"] != 1 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "not_cast",
		})
		return
	}
	target := -1
	bestLife := 1 << 30
	for _, opp := range gs.Opponents(perm.Controller) {
		s := gs.Seats[opp]
		if s == nil || s.Lost {
			continue
		}
		if s.Life < bestLife {
			bestLife = s.Life
			target = opp
		}
	}
	if target < 0 {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_opponent", nil)
		return
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"target":    target,
		"opp_life":  bestLife,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"target_opponent_extra_turn_engine_lacks_seat_specific_extra_turn_queue")
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"protection_from_target_player_until_next_turn_no_engine_per_player_protection_static")
}
