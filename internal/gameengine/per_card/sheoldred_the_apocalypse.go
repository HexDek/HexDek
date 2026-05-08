package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSheoldredTheApocalypse wires Sheoldred, the Apocalypse.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Deathtouch
//	Whenever you draw a card, you gain 2 life.
//	Whenever an opponent draws a card, they lose 2 life.
//
// Implementation:
//   - Deathtouch handled by AST keyword pipeline.
//   - "card_drawn" trigger: 2 life to the drawer if it's Sheoldred's
//     controller, 2 life loss to the drawer if it's an opponent.
func registerSheoldredTheApocalypse(r *Registry) {
	r.OnTrigger("Sheoldred, the Apocalypse", "card_drawn", sheoldredApocCardDrawn)
}

func sheoldredApocCardDrawn(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "sheoldred_apocalypse_card_drawn"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	drawerSeat, _ := ctx["seat"].(int)
	if drawerSeat < 0 || drawerSeat >= len(gs.Seats) {
		return
	}
	drawer := gs.Seats[drawerSeat]
	if drawer == nil {
		return
	}
	if drawerSeat == perm.Controller {
		gameengine.GainLife(gs, perm.Controller, 2, perm.Card.DisplayName())
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"effect": "gain_2",
		})
		return
	}
	gameengine.LoseLife(gs, drawerSeat, 2, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"target": drawerSeat,
		"effect": "lose_2",
	})
}
