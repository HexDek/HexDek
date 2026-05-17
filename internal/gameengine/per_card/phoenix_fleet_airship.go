package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerPhoenixFleetAirship wires Phoenix Fleet Airship.
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	Flying
//	At the beginning of your end step, if you sacrificed a permanent
//	this turn, create a token that's a copy of this Vehicle.
//	As long as you control eight or more permanents named Phoenix Fleet
//	Airship, this Vehicle is an artifact creature.
//	Crew 1
//
// Implementation (Muninn gap #34 — 28K hits):
//   - Flying / Crew handled by AST keyword pipeline.
//   - OnTrigger("end_step") gated on controller == active seat and on
//     seat.Turn.Sacrificed > 0 (state.go's TurnCounters.Sacrificed).
//     Token copy via Card.DeepCopy + enterBattlefieldWithETB mirrors
//     resolve.go:resolveCreateTokenCopy (line 1755).
//   - The "becomes a creature when you control 8+" static type-changing
//     overlay needs the Phase 8 layers pass. emitPartial.
func registerPhoenixFleetAirship(r *Registry) {
	r.OnTrigger("Phoenix Fleet Airship", "end_step", phoenixFleetAirshipEndStep)
}

func phoenixFleetAirshipEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "phoenix_fleet_airship_copy_on_sacrifice"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	seatIdx := perm.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[seatIdx]
	if seat == nil || seat.Lost {
		return
	}
	if seat.Turn.Sacrificed <= 0 {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_permanent_sacrificed_this_turn", map[string]interface{}{
			"seat": seatIdx,
		})
		return
	}
	card := perm.Card.DeepCopy()
	hasToken := false
	for _, t := range card.Types {
		if strings.EqualFold(t, "token") {
			hasToken = true
			break
		}
	}
	if !hasToken {
		card.Types = append([]string{"token"}, card.Types...)
	}
	card.Owner = seatIdx
	token := enterBattlefieldWithETB(gs, seatIdx, card, false)
	if token == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "token_creation_failed", nil)
		return
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       seatIdx,
		"sacrificed": seat.Turn.Sacrificed,
		"token":      "Phoenix Fleet Airship (token copy)",
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"static_eight_copies_become_creature_needs_phase8_layers_overlay")
}
