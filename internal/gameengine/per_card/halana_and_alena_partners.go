package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerHalanaAndAlenaPartners wires Halana and Alena, Partners.
//
// Oracle text:
//
//	First strike
//	Reach
//	At the beginning of combat on your turn, put X +1/+1 counters on
//	another target creature you control, where X is Halana and
//	Alena's power. That creature gains haste until end of turn.
//
// Implementation:
//   - combat_begin trigger: pick the highest-power non-Halana creature
//     we control as the target. Add X +1/+1 counters where X is our
//     own current power.
//   - "Gains haste until end of turn" — set the kw:haste flag on the
//     target. (No until-end-of-turn cleanup; mirrors how other
//     handlers grant temporary haste.)
func registerHalanaAndAlenaPartners(r *Registry) {
	r.OnTrigger("Halana and Alena, Partners", "combat_begin", halanaAlenaCombatBegin)
}

func halanaAlenaCombatBegin(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "halana_alena_combat_buff"
	if gs == nil || perm == nil {
		return
	}
	if gs.Active != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	x := int(perm.Card.BasePower)
	if x <= 0 {
		return
	}
	var target *gameengine.Permanent
	bestPower := -1
	for _, p := range seat.Battlefield {
		if p == nil || p == perm || p.Card == nil {
			continue
		}
		if !cardHasType(p.Card, "creature") {
			continue
		}
		pp := int(p.Card.BasePower)
		if pp > bestPower {
			bestPower = pp
			target = p
		}
	}
	if target == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":     perm.Controller,
			"targeted": false,
		})
		return
	}
	target.AddCounter("+1/+1", x)
	if target.Flags == nil {
		target.Flags = map[string]int{}
	}
	target.Flags["kw:haste"] = 1
	gs.InvalidateCharacteristicsCache()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"target":   target.Card.DisplayName(),
		"counters": x,
	})
}
