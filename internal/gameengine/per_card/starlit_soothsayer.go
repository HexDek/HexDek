package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerStarlitSoothsayer wires Starlit Soothsayer (Muninn parser-gap
// #58, 12,879 hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{2}{B}
//	Creature — Bat Cleric
//	Flying
//	At the beginning of your end step, if you gained or lost life this
//	turn, surveil 1.
//
// Implementation:
//   - Flying is AST-engine-side.
//   - End-step trigger gated on active_seat == controller AND
//     (seat.Turn.LifeGained > 0 OR seat.Turn.LifeLost > 0).
//   - Surveil 1 via gameengine.Surveil — the hat's ChooseSurveil
//     handles top-vs-graveyard.
func registerStarlitSoothsayer(r *Registry) {
	// end_step normalizes to a single canonical event; one registration
	// avoids any alias-driven double-fire.
	r.OnTrigger("Starlit Soothsayer", "end_step", starlitSoothsayerEndStep)
}

func starlitSoothsayerEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "starlit_soothsayer_end_step_surveil"
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
	if seat.Turn.LifeGained <= 0 && seat.Turn.LifeLost <= 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
		})
		return
	}
	gameengine.Surveil(gs, perm.Controller, 1)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"triggered": true,
	})
}
