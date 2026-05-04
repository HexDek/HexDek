package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTrevaTheRenewer wires Treva, the Renewer.
//
// Oracle text:
//
//	{3}{G}{W}{U}
//	Legendary Creature — Dragon
//	Flying
//	Whenever Treva deals combat damage to a player, you may pay {2}{W}.
//	  If you do, choose a color, then you gain 1 life for each permanent
//	  of that color.
//
// Implementation:
//   - Flying via AST.
//   - "creature_combat_damage_to_player" trigger gated to attacker ==
//     Treva. Heuristic: always "pay" (since handler can't manage mana
//     payment cleanly), pick the color with the largest permanent count
//     across all battlefields, gain that many life. emitPartial covers
//     the {2}{W} payment gate.
func registerTrevaTheRenewer(r *Registry) {
	r.OnTrigger("Treva, the Renewer", "creature_combat_damage_to_player", trevaOnCombatDamage)
}

func trevaOnCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "treva_lifegain"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk == nil || atk != perm {
		return
	}
	colors := []string{"W", "U", "B", "R", "G"}
	bestColor, bestCount := "", 0
	for _, c := range colors {
		count := trevaCountColor(gs, c)
		if count > bestCount {
			bestColor = c
			bestCount = count
		}
	}
	if bestCount > 0 {
		gameengine.GainLife(gs, perm.Controller, bestCount, perm.Card.DisplayName())
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"color": bestColor,
		"life":  bestCount,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"may_pay_2W_optional_payment_partial")
}

func trevaCountColor(gs *gameengine.GameState, color string) int {
	count := 0
	upper := strings.ToUpper(color)
	for _, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil {
				continue
			}
			for _, c := range p.Card.Colors {
				if strings.ToUpper(c) == upper {
					count++
					break
				}
			}
		}
	}
	return count
}
