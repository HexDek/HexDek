package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerByrkeLongEarOfTheLaw wires Byrke, Long Ear of the Law.
//
// Oracle text (BLB, {4}{G}{W}, 4/4):
//
//	Vigilance
//	When Byrke enters, put a +1/+1 counter on each of up to two target
//	creatures.
//	Whenever a creature you control with a +1/+1 counter on it attacks,
//	double the number of +1/+1 counters on it.
//
// Implementation:
//   - ETB targets up to 2 of the controller's creatures, preferring
//     ones that already have counters (to seed Byrke's payoff trigger).
//   - "creature_attacks" trigger: if the attacker is controlled by
//     Byrke's controller and has +1/+1 counters, double them.
func registerByrkeLongEarOfTheLaw(r *Registry) {
	r.OnETB("Byrke, Long Ear of the Law", byrkeETB)
	r.OnTrigger("Byrke, Long Ear of the Law", "creature_attacks", byrkeAttackDouble)
}

func byrkeETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "byrke_etb_counters"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	var withCounters, others []*gameengine.Permanent
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil || !p.IsCreature() {
			continue
		}
		if p == perm {
			others = append(others, p)
			continue
		}
		if p.Counters != nil && p.Counters["+1/+1"] > 0 {
			withCounters = append(withCounters, p)
		} else {
			others = append(others, p)
		}
	}
	picks := append([]*gameengine.Permanent{}, withCounters...)
	for _, p := range others {
		if len(picks) >= 2 {
			break
		}
		picks = append(picks, p)
	}
	if len(picks) > 2 {
		picks = picks[:2]
	}
	if len(picks) == 0 && perm.IsCreature() {
		picks = append(picks, perm)
	}
	names := make([]string, 0, len(picks))
	for _, p := range picks {
		p.AddCounter("+1/+1", 1)
		names = append(names, p.Card.DisplayName())
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":    perm.Controller,
		"targets": names,
	})
}

func byrkeAttackDouble(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "byrke_attack_double_counters"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk == nil || atk.Card == nil {
		return
	}
	if atk.Controller != perm.Controller {
		return
	}
	if atk.Counters == nil {
		return
	}
	n := atk.Counters["+1/+1"]
	if n <= 0 {
		return
	}
	atk.AddCounter("+1/+1", n)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"attacker": atk.Card.DisplayName(),
		"before":   n,
		"after":    n * 2,
	})
}
