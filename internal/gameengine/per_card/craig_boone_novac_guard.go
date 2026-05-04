package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCraigBooneNovacGuard wires Craig Boone, Novac Guard.
//
// Oracle text:
//
//	Reach, lifelink
//	One for My Baby — Whenever you attack with two or more creatures,
//	put two quest counters on Craig Boone. When you do, Craig Boone
//	deals damage equal to the number of quest counters on it to up to
//	one target creature unless that creature's controller has Craig
//	Boone deal that much damage to them.
//
// Implementation: track "attack_with_count >= 2" via the engine's
// declare_attackers event ctx.
func registerCraigBooneNovacGuard(r *Registry) {
	r.OnTrigger("Craig Boone, Novac Guard", "declare_attackers", craigBooneAttack)
}

func craigBooneAttack(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "craig_boone_quest_counters"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	attackerSeat, _ := ctx["seat"].(int)
	if attackerSeat != perm.Controller {
		return
	}
	count, _ := ctx["count"].(int)
	if count < 2 {
		return
	}
	perm.AddCounter("quest", 2)
	dmg := perm.Counters["quest"]
	// Damage to best opposing creature; planar choice = creature.
	var best *gameengine.Permanent
	bestPow := -1
	for i, opp := range gs.Seats {
		if opp == nil || i == perm.Controller || opp.Lost {
			continue
		}
		for _, p := range opp.Battlefield {
			if p == nil || !p.IsCreature() {
				continue
			}
			if pow := p.Power(); pow > bestPow {
				best = p
				bestPow = pow
			}
		}
	}
	target := ""
	if best != nil {
		gameengine.FireDamageEvent(gs, perm, best.Controller, best, dmg)
		target = best.Card.DisplayName()
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"counters": perm.Counters["quest"],
		"damage":   dmg,
		"target":   target,
	})
}
