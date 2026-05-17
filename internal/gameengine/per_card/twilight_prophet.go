package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTwilightProphet wires Twilight Prophet.
//
// Oracle text (Scryfall, verified 2026-05-16):
//
//	Flying
//	Ascend (If you control ten or more permanents, you get the city's
//	blessing for the rest of the game.)
//	At the beginning of your upkeep, if you have the city's blessing,
//	reveal the top card of your library and put it into your hand.
//	Each opponent loses X life and you gain X life, where X is that
//	card's mana value.
//
// Implementation:
//   - Flying / Ascend: engine-side (FirePermanentETBTriggers calls
//     CheckAscend, AST keyword handles flying).
//   - Upkeep gated on active_seat == controller AND city's blessing.
//     Reveal the top card (the LogEvent of the move is the "reveal"),
//     move to hand, drain each living opponent for X = revealed.cmc,
//     then GainLife(X) once (CR §118 — single life-gain event).
func registerTwilightProphet(r *Registry) {
	r.OnTrigger("Twilight Prophet", "upkeep", twilightProphetUpkeep)
	r.OnTrigger("Twilight Prophet", "upkeep_controller", twilightProphetUpkeep)
}

func twilightProphetUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "twilight_prophet_upkeep"
	if gs == nil || perm == nil || ctx == nil {
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
	if !gameengine.HasCitysBlessing(gs, perm.Controller) {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "no_citys_blessing",
		})
		return
	}
	if len(seat.Library) == 0 {
		seat.AttemptedEmptyDraw = true
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": true,
			"empty":     true,
		})
		return
	}
	top := seat.Library[0]
	gs.LogEvent(gameengine.Event{
		Kind:   "reveal",
		Seat:   perm.Controller,
		Source: perm.Card.DisplayName(),
		Details: map[string]interface{}{
			"card": top.DisplayName(),
			"cmc":  cardCMC(top),
		},
	})
	gameengine.MoveCard(gs, top, perm.Controller, "library", "hand", "twilight_prophet_reveal")
	x := cardCMC(top)
	if x > 0 {
		for i, s := range gs.Seats {
			if s == nil || s.Lost || i == perm.Controller {
				continue
			}
			gameengine.LoseLife(gs, i, x, perm.Card.DisplayName())
		}
		gameengine.GainLife(gs, perm.Controller, x, perm.Card.DisplayName())
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"triggered": true,
		"revealed":  top.DisplayName(),
		"x":         x,
	})
}
