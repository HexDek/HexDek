package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerByrkeLongEarOfTheLaw wires Byrke, Long Ear of the Law.
//
// Oracle text:
//
//   Vigilance
//   When Byrke enters, put a +1/+1 counter on each of up to two target creatures.
//   Whenever a creature you control with a +1/+1 counter on it attacks, double the number of +1/+1 counters on it.
func registerByrkeLongEarOfTheLaw(r *Registry) {
	r.OnETB("Byrke, Long Ear of the Law", byrkeLongEarOfTheLawETB)
	r.OnTrigger("Byrke, Long Ear of the Law", "creature_attacks", byrkeAttackDouble)
}

// byrkeLongEarOfTheLawETB: choose up to 2 creatures controlled by Byrke's
// controller (greedy: highest-power non-Byrke creatures) and put a +1/+1
// counter on each.
func byrkeLongEarOfTheLawETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "byrke_long_ear_of_the_law_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil {
		return
	}
	type cand struct {
		p     *gameengine.Permanent
		power int
	}
	var pool []cand
	for _, p := range s.Battlefield {
		if p == nil || !p.IsCreature() {
			continue
		}
		pool = append(pool, cand{p: p, power: p.Power()})
	}
	// Sort descending by power (greedy: counters go on biggest threats).
	for i := 0; i < len(pool); i++ {
		for j := i + 1; j < len(pool); j++ {
			if pool[j].power > pool[i].power {
				pool[i], pool[j] = pool[j], pool[i]
			}
		}
	}
	picked := 0
	var names []string
	for _, c := range pool {
		if picked >= 2 {
			break
		}
		c.p.AddCounter("+1/+1", 1)
		picked++
		if c.p.Card != nil {
			names = append(names, c.p.Card.DisplayName())
		}
	}
	if picked > 0 {
		gs.InvalidateCharacteristicsCache()
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     seat,
		"targeted": names,
		"count":    picked,
	})
}

// byrkeAttackDouble: when a creature controlled by Byrke's controller with
// at least one +1/+1 counter attacks, double those counters.
func byrkeAttackDouble(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "byrke_long_ear_of_the_law_attack"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk == nil || atk.Controller != perm.Controller {
		return
	}
	if atk.Counters == nil {
		return
	}
	cur := atk.Counters["+1/+1"]
	if cur <= 0 {
		return
	}
	atk.AddCounter("+1/+1", cur)
	gs.InvalidateCharacteristicsCache()
	name := ""
	if atk.Card != nil {
		name = atk.Card.DisplayName()
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          perm.Controller,
		"attacker":      name,
		"counters_before": cur,
		"counters_after":  atk.Counters["+1/+1"],
	})
}
