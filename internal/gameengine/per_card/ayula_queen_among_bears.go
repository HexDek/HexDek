package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAyulaQueenAmongBears wires Ayula, Queen Among Bears.
//
// Oracle text:
//
//	Whenever another Bear you control enters, choose one —
//	• Put two +1/+1 counters on target Bear.
//	• Target Bear you control fights target creature you don't control.
//
// Heuristic: prefer counters mode (no risk) — buff the entering Bear
// with two +1/+1 counters. Fight mode (skipped) is a TODO that requires
// damage resolution between two creatures.
func registerAyulaQueenAmongBears(r *Registry) {
	r.OnTrigger("Ayula, Queen Among Bears", "permanent_etb", ayulaBearETB)
}

func ayulaBearETB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ayula_bear_etb_counters"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	controllerSeat, _ := ctx["controller_seat"].(int)
	if controllerSeat != perm.Controller {
		return
	}
	entered, _ := ctx["perm"].(*gameengine.Permanent)
	if entered == nil || entered == perm || entered.Card == nil {
		return
	}
	if !cardHasType(entered.Card, "bear") {
		return
	}
	entered.AddCounter("+1/+1", 2)
	gs.InvalidateCharacteristicsCache()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"target": entered.Card.DisplayName(),
	})
}
