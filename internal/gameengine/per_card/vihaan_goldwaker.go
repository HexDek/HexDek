package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerVihaanGoldwaker wires Vihaan, Goldwaker.
//
// Oracle text:
//
//	Other outlaws you control have vigilance and haste. (Assassins,
//	Mercenaries, Pirates, Rogues, and Warlocks are outlaws.)
//	At the beginning of combat on your turn, you may have Treasures you
//	control become 3/3 Construct Assassin artifact creatures in addition
//	to their other types until end of turn.
//
// Implementation:
//   - "combat_begin" on Vihaan's controller's turn: animate every Treasure
//     we control as a 3/3 Construct Assassin until end of turn. The
//     animation is implemented by appending "creature", "construct",
//     "assassin" types and a +0/+0 Modification with BaseSet semantics
//     (we use Modifications with explicit power/toughness override).
//     Cleanup at next_end_step.
//   - The static "outlaws have vigilance/haste" anthem is left to the
//     engine's static-ability system; emitPartial flags the gap.
func registerVihaanGoldwaker(r *Registry) {
	r.OnTrigger("Vihaan, Goldwaker", "combat_begin", vihaanCombatBegin)
}

func vihaanCombatBegin(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "vihaan_goldwaker_animate_treasures"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}

	emitPartial(gs, slug, perm.Card.DisplayName(), "outlaw_anthem_static_not_implemented")

	type savedTypes struct {
		perm *gameengine.Permanent
		old  []string
	}
	var saved []savedTypes
	count := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		isTreasure := false
		for _, t := range p.Card.Types {
			if t == "treasure" || t == "Treasure" {
				isTreasure = true
				break
			}
		}
		if !isTreasure {
			continue
		}
		// Snapshot original types then add creature/construct/assassin.
		old := append([]string(nil), p.Card.Types...)
		saved = append(saved, savedTypes{perm: p, old: old})
		needAdd := []string{"creature", "construct", "assassin"}
		for _, want := range needAdd {
			has := false
			for _, t := range p.Card.Types {
				if t == want {
					has = true
					break
				}
			}
			if !has {
				p.Card.Types = append(p.Card.Types, want)
			}
		}
		// Set base P/T to 3/3 via modification.
		p.Modifications = append(p.Modifications, gameengine.Modification{
			Power:     3 - p.Card.BasePower,
			Toughness: 3 - p.Card.BaseToughness,
			Duration:  "until_end_of_turn",
		})
		count++
	}
	if count == 0 {
		return
	}
	gs.InvalidateCharacteristicsCache()

	gs.RegisterDelayedTrigger(&gameengine.DelayedTrigger{
		TriggerAt:      "next_end_step",
		ControllerSeat: perm.Controller,
		SourceCardName: perm.Card.DisplayName(),
		EffectFn: func(gs *gameengine.GameState) {
			for _, s := range saved {
				if s.perm == nil || s.perm.Card == nil {
					continue
				}
				s.perm.Card.Types = s.old
			}
			gs.InvalidateCharacteristicsCache()
		},
	})

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"animated": count,
	})
}
