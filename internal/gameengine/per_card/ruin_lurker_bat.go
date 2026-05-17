package per_card

import (
	"strconv"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRuinLurkerBat wires Ruin-Lurker Bat (Muninn parser-gap #87,
// ~6.1K hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{2}{B}
//	Creature — Bat
//	Flying, lifelink
//	At the beginning of your end step, if you descended this turn,
//	scry 1. (You descended if a permanent card was put into your
//	graveyard from anywhere.)
//
// Implementation:
//   - Flying / lifelink: AST keyword pipeline.
//   - "end_step" listener gated on active_seat == controller. Read the
//     engine-maintained gs.Flags["descended_N"] flag — set by zone_move.go
//     whenever a permanent card enters a player's graveyard — and Scry 1 if
//     it's set for the controller this turn.
func registerRuinLurkerBat(r *Registry) {
	r.OnTrigger("Ruin-Lurker Bat", "end_step", ruinLurkerBatEndStep)
}

func ruinLurkerBatEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ruin_lurker_bat_descend_scry"
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
	if gs.Flags == nil || gs.Flags["descended_"+strconv.Itoa(perm.Controller)] == 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "no_descend_this_turn",
		})
		return
	}
	gameengine.Scry(gs, perm.Controller, 1)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"triggered": true,
		"scry":      1,
	})
}
