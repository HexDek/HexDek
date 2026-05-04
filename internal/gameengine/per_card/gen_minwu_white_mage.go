package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMinwuWhiteMage wires Minwu, White Mage.
//
// Oracle text:
//
//   Vigilance, lifelink
//   Whenever you gain life, put a +1/+1 counter on each Cleric you control.
//
// Implementation: on "life_gained" for Minwu's controller, place a
// +1/+1 counter on each Cleric (including Minwu if he is one). Vigilance
// and lifelink are AST keywords.
func registerMinwuWhiteMage(r *Registry) {
	r.OnTrigger("Minwu, White Mage", "life_gained", minwuWhiteMageTrigger)
}

func minwuWhiteMageTrigger(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "minwu_lifegain_cleric_counters"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	gainSeat, _ := ctx["seat"].(int)
	if gainSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	count := 0
	for _, p := range seat.Battlefield {
		if p == nil || !p.IsCreature() || p.Card == nil {
			continue
		}
		if !cardHasType(p.Card, "cleric") {
			continue
		}
		p.AddCounter("+1/+1", 1)
		count++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":    perm.Controller,
		"clerics": count,
	})
}
