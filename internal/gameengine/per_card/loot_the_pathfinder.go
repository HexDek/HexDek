package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerLootThePathfinder wires Loot, the Pathfinder.
//
// Oracle text:
//
//	Double strike, vigilance, haste
//	Exhaust — {G}, {T}: Add three mana of any one color.
//	Exhaust — {U}, {T}: Draw three cards.
//	Exhaust — {R}, {T}: Loot deals 3 damage to any target.
//
// Implementation: each exhaust ability can be activated once per game.
// Track via permanent flags. Ability index 0 = mana, 1 = draw 3,
// 2 = 3 damage to any target.
func registerLootThePathfinder(r *Registry) {
	r.OnActivated("Loot, the Pathfinder", lootActivated)
}

func lootActivated(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "loot_pathfinder_exhaust"
	if gs == nil || src == nil {
		return
	}
	if src.Tapped {
		emitFail(gs, slug, src.Card.DisplayName(), "already_tapped", nil)
		return
	}
	if src.Flags == nil {
		src.Flags = map[string]int{}
	}
	key := "exhaust_used_"
	switch abilityIdx {
	case 0:
		key += "mana"
	case 1:
		key += "draw"
	case 2:
		key += "damage"
	default:
		return
	}
	if src.Flags[key] > 0 {
		emitFail(gs, slug, src.Card.DisplayName(), "exhaust_already_used", map[string]interface{}{"key": key})
		return
	}
	src.Flags[key] = 1
	src.Tapped = true
	switch abilityIdx {
	case 0:
		seat := gs.Seats[src.Controller]
		if seat != nil && seat.Mana != nil {
			seat.Mana.Any += 3
			gameengine.SyncManaAfterSpend(seat)
		} else if seat != nil {
			seat.ManaPool += 3
		}
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":   src.Controller,
			"effect": "add_3_mana",
		})
	case 1:
		for i := 0; i < 3; i++ {
			drawOne(gs, src.Controller, src.Card.DisplayName())
		}
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":   src.Controller,
			"effect": "draw_3",
		})
	case 2:
		// Pick the highest-life opponent as default target.
		target := -1
		bestLife := -1
		for i, s := range gs.Seats {
			if s == nil || s.Lost || i == src.Controller {
				continue
			}
			if s.Life > bestLife {
				bestLife = s.Life
				target = i
			}
		}
		if target >= 0 {
			gameengine.DealDamage(gs, target, 3, src.Card.DisplayName())
			_ = gs.CheckEnd()
		}
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":        src.Controller,
			"effect":      "damage_3",
			"target_seat": target,
		})
	}
}
