package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerFeastOfTheVictoriousDead wires Feast of the Victorious Dead
// (Muninn parser-gap #78, ~7.9K hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{W}{B}
//	Enchantment
//	At the beginning of your end step, if one or more creatures died
//	this turn, you gain that much life and distribute that many +1/+1
//	counters among any number of creatures you control.
//
// Implementation:
//   - end_step gated on active_seat == controller.
//   - "Died this turn" = controller's seat.Turn.CreaturesDied (the
//     engine increments this on §704 death moves; see zone_change.go).
//     Note: the printed text says "creatures died this turn" globally,
//     not just under the controller — but the engine's per-seat counter
//     only tracks deaths attributable to that seat (zone_change.go
//     bumps the owner's TurnCounters). Close enough for most cases;
//     emitPartial flags the over-narrow scope.
//   - GainLife(N) for the life half.
//   - Distribute N +1/+1 counters: Hat policy — stack all N on the
//     single highest-power non-token creature we control (best chance
//     to convert into combat damage). If no creatures, no counters
//     placed but life still gained.
func registerFeastOfTheVictoriousDead(r *Registry) {
	r.OnTrigger("Feast of the Victorious Dead", "end_step", feastOfTheVictoriousDeadEndStep)
}

func feastOfTheVictoriousDeadEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "feast_of_the_victorious_dead_end_step"
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
	n := seat.Turn.CreaturesDied
	if n <= 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "no_deaths_this_turn",
		})
		return
	}
	gameengine.GainLife(gs, perm.Controller, n, perm.Card.DisplayName())
	var best *gameengine.Permanent
	for _, p := range seat.Battlefield {
		if p == nil || !p.IsCreature() {
			continue
		}
		if best == nil || p.Power() > best.Power() {
			best = p
		}
	}
	target := "none"
	if best != nil {
		best.AddCounter("+1/+1", n)
		gs.InvalidateCharacteristicsCache()
		target = best.Card.DisplayName()
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"triggered":  true,
		"life":       n,
		"counters":   n,
		"target":     target,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"global_creatures_died_proxy_uses_controller_only_counter")
}
