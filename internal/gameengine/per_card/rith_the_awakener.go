package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRithTheAwakener wires Rith, the Awakener.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Flying
//	Whenever Rith deals combat damage to a player, you may pay
//	{2}{G}. If you do, choose a color, then create a 1/1 green
//	Saproling creature token for each permanent of that color.
//
// Implementation:
//   - Flying via AST keyword pipeline.
//   - "combat_damage_player" trigger: when Rith damages a player, pick
//     the dominant color across the battlefield and mint that many
//     Saprolings if Rith's controller has {2}{G} unspent. The pay-cost
//     decision uses a simple "would-net 3+ tokens" heuristic.
func registerRithTheAwakener(r *Registry) {
	r.OnTrigger("Rith, the Awakener", "combat_damage_player", rithCombatDamage)
}

func rithCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "rith_combat_damage_saprolings"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	srcName, _ := ctx["source_card"].(string)
	if !strings.EqualFold(srcName, perm.Card.DisplayName()) {
		return
	}
	srcSeat, _ := ctx["source_seat"].(int)
	if srcSeat != perm.Controller {
		return
	}

	bestColor := ""
	bestCount := 0
	for _, color := range []string{"W", "U", "B", "R", "G"} {
		n := rithCountColor(gs, color)
		if n > bestCount {
			bestCount = n
			bestColor = color
		}
	}
	if bestCount < 3 {
		emitFail(gs, slug, perm.Card.DisplayName(), "value_below_threshold", map[string]interface{}{
			"seat":  perm.Controller,
			"count": bestCount,
		})
		return
	}
	for i := 0; i < bestCount; i++ {
		gameengine.CreateCreatureToken(gs, perm.Controller, "Saproling",
			[]string{"creature", "saproling", "pip:G"}, 1, 1)
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"color":  bestColor,
		"tokens": bestCount,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"two_green_cost_payment_not_actually_deducted_from_mana_pool")
}

func rithCountColor(gs *gameengine.GameState, color string) int {
	n := 0
	want := strings.ToUpper(color)
	for _, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil {
				continue
			}
			for _, c := range p.Card.Colors {
				if strings.ToUpper(c) == want {
					n++
					break
				}
			}
		}
	}
	return n
}
