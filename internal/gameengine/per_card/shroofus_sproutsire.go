package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerShroofusSproutsire wires Shroofus Sproutsire.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Trample
//	Whenever a Saproling you control deals combat damage to a player,
//	create that many 1/1 green Saproling creature tokens.
//
// Implementation:
//   - Trample handled by AST keyword pipeline.
//   - "combat_damage_player" trigger: when a Saproling Shroofus's
//     controller controls deals combat damage to a player, mint that
//     many Saproling tokens.
func registerShroofusSproutsire(r *Registry) {
	r.OnTrigger("Shroofus Sproutsire", "combat_damage_player", shroofusCombatDamage)
}

func shroofusCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "shroofus_saproling_combat_damage"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	srcSeat, _ := ctx["source_seat"].(int)
	if srcSeat != perm.Controller {
		return
	}
	srcName, _ := ctx["source_card"].(string)
	if srcName == "" {
		return
	}
	// Confirm source is a Saproling we control.
	isSap := false
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if !strings.EqualFold(p.Card.DisplayName(), srcName) {
			continue
		}
		for _, t := range p.Card.Types {
			if strings.EqualFold(t, "saproling") {
				isSap = true
				break
			}
		}
		if isSap {
			break
		}
	}
	if !isSap {
		return
	}
	amount, _ := ctx["amount"].(int)
	if amount <= 0 {
		return
	}
	for i := 0; i < amount; i++ {
		gameengine.CreateCreatureToken(gs, perm.Controller, "Saproling",
			[]string{"creature", "saproling", "pip:G"}, 1, 1)
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"source": srcName,
		"tokens": amount,
	})
}
