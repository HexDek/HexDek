package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKrosDefenseContractor wires Kros, Defense Contractor.
//
// Oracle text:
//
//	At the beginning of your upkeep, put a shield counter on target
//	creature an opponent controls.
//	Whenever you put one or more counters on a creature you don't
//	control, tap that creature and goad it. It gains trample until your
//	next turn.
//
// Implementation: upkeep places a shield counter on the highest-toughness
// opposing creature. Counter-on-other-creature trigger is non-trivial
// (would require listening on counter_placed with a "placed by my
// effect" filter); emitPartial for that clause.
func registerKrosDefenseContractor(r *Registry) {
	r.OnTrigger("Kros, Defense Contractor", "upkeep_controller", krosUpkeep)
	r.OnTrigger("Kros, Defense Contractor", "counter_placed", krosCounterPlaced)
}

func krosUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kros_upkeep_shield_counter"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	var target *gameengine.Permanent
	bestTough := -1
	for opp := range gs.Seats {
		if opp == perm.Controller {
			continue
		}
		os := gs.Seats[opp]
		if os == nil || os.Lost {
			continue
		}
		for _, p := range os.Battlefield {
			if p == nil || !p.IsCreature() || p.Card == nil {
				continue
			}
			t := p.Card.BaseToughness
			if t > bestTough {
				bestTough = t
				target = p
			}
		}
	}
	if target == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_opponent_creature", nil)
		return
	}
	if target.Counters == nil {
		target.Counters = map[string]int{}
	}
	target.Counters["shield"]++
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"target": target.Card.DisplayName(),
	})
}

func krosCounterPlaced(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kros_counter_goad"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"goad_on_counter_placed_by_self_unimplemented")
}
