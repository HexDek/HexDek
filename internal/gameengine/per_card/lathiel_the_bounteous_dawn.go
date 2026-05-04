package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerLathielTheBounteousDawn wires Lathiel, the Bounteous Dawn.
//
// Oracle text:
//
//	Lifelink
//	At the beginning of each end step, if you gained life this turn,
//	distribute up to that many +1/+1 counters among any number of other
//	target creatures.
//
// Implementation: track life-gained-this-turn via a per-permanent flag
// (life_gained event observer increments it; cleared at end step after
// distribution). At each end step, distribute counters across owned
// creatures (excluding Lathiel itself), one per creature in BasePower
// descending order, capped at gained.
func registerLathielTheBounteousDawn(r *Registry) {
	r.OnTrigger("Lathiel, the Bounteous Dawn", "life_gained", lathielTrackLifeGained)
	r.OnTrigger("Lathiel, the Bounteous Dawn", "end_step", lathielEndStep)
}

func lathielTrackLifeGained(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	seat, _ := ctx["seat"].(int)
	if seat != perm.Controller {
		return
	}
	amount, _ := ctx["amount"].(int)
	if amount <= 0 {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["lathiel_life_gained_this_turn"] += amount
}

func lathielEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "lathiel_distribute_counters"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	gained := perm.Flags["lathiel_life_gained_this_turn"]
	delete(perm.Flags, "lathiel_life_gained_this_turn")
	if gained <= 0 {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	placed := 0
	for _, p := range seat.Battlefield {
		if placed >= gained {
			break
		}
		if p == nil || p == perm || !p.IsCreature() || p.Card == nil {
			continue
		}
		if p.Counters == nil {
			p.Counters = map[string]int{}
		}
		p.Counters["+1/+1"]++
		placed++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"gained":   gained,
		"placed":   placed,
	})
}
