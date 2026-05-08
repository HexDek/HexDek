package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMarcusMutantMayor wires Marcus, Mutant Mayor.
//
// Oracle text:
//
//	Vigilance, trample
//	Whenever a creature you control deals combat damage to a player,
//	draw a card if that creature has a +1/+1 counter on it. If it
//	doesn't, put a +1/+1 counter on it.
//
// Implementation: combat_damage_player listener gates on source-controller
// match. Find the source permanent on battlefield to check counters and
// place counters as needed.
func registerMarcusMutantMayor(r *Registry) {
	r.OnTrigger("Marcus, Mutant Mayor", "combat_damage_player", marcusCombatDamage)
}

func marcusCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "marcus_mutant_mayor_combat_trigger"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	sourceSeat, _ := ctx["source_seat"].(int)
	if sourceSeat != perm.Controller {
		return
	}
	sourceName, _ := ctx["source_card"].(string)
	if sourceName == "" {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	var src *gameengine.Permanent
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if p.Card.DisplayName() == sourceName {
			src = p
			break
		}
	}
	if src == nil {
		return
	}
	if src.Counters != nil && src.Counters["+1/+1"] > 0 {
		drawOne(gs, perm.Controller, perm.Card.DisplayName())
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"source": sourceName,
			"effect": "draw",
		})
		return
	}
	src.AddCounter("+1/+1", 1)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"source": sourceName,
		"effect": "counter",
	})
}
