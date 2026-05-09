package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKarlovOfTheGhostCouncil wires Karlov of the Ghost Council.
//
// Oracle text:
//
//	Whenever you gain life, put two +1/+1 counters on Karlov.
//	{W}{B}, Remove six +1/+1 counters from Karlov: Exile target creature.
//
// Implementation:
//   - Lifegain trigger: when ctx["seat"] == controller, add TWO +1/+1
//     counters. (The auto-generated stub added one, missing the
//     "two" wording.)
//   - Activated abilityIdx 0: pay {W}{B} (2 generic mana from the
//     simplified pool), require ≥6 +1/+1 counters on Karlov, remove 6,
//     and exile ctx["target_perm"] (a *Permanent). When no target is
//     supplied, fall back to the largest opponent creature.
func registerKarlovOfTheGhostCouncil(r *Registry) {
	r.OnTrigger("Karlov of the Ghost Council", "life_gained", karlovOfTheGhostCouncilTrigger)
	r.OnActivated("Karlov of the Ghost Council", karlovActivate)
}

func karlovOfTheGhostCouncilTrigger(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "karlov_of_the_ghost_council_trigger"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	gainSeat, _ := ctx["seat"].(int)
	if gainSeat != perm.Controller {
		return
	}
	perm.AddCounter("+1/+1", 2)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"new_count": perm.Counters["+1/+1"],
	})
}

func karlovActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "karlov_exile_activate"
	if gs == nil || src == nil || src.Card == nil {
		return
	}
	if abilityIdx != 0 {
		return
	}
	seat := src.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil {
		return
	}
	const cost = 2 // {W}{B} ≈ 2 generic from the simplified pool
	if s.ManaPool < cost {
		emitFail(gs, slug, src.Card.DisplayName(), "insufficient_mana", map[string]interface{}{
			"seat":      seat,
			"required":  cost,
			"available": s.ManaPool,
		})
		return
	}
	if src.Counters == nil || src.Counters["+1/+1"] < 6 {
		have := 0
		if src.Counters != nil {
			have = src.Counters["+1/+1"]
		}
		emitFail(gs, slug, src.Card.DisplayName(), "insufficient_counters", map[string]interface{}{
			"seat":     seat,
			"required": 6,
			"have":     have,
		})
		return
	}

	// Resolve target.
	var target *gameengine.Permanent
	if ctx != nil {
		if t, ok := ctx["target_perm"].(*gameengine.Permanent); ok {
			target = t
		}
	}
	if target == nil {
		bestPow := -1
		for _, opp := range gs.Opponents(seat) {
			os := gs.Seats[opp]
			if os == nil {
				continue
			}
			for _, p := range os.Battlefield {
				if p == nil || p.Card == nil || !p.IsCreature() {
					continue
				}
				if p.Power() > bestPow {
					bestPow = p.Power()
					target = p
				}
			}
		}
	}
	if target == nil || !target.IsCreature() {
		emitFail(gs, slug, src.Card.DisplayName(), "no_legal_creature_target", nil)
		return
	}

	// Pay costs: spend mana, remove 6 counters.
	s.ManaPool -= cost
	gameengine.SyncManaAfterSpend(s)
	src.AddCounter("+1/+1", -6)

	gameengine.ExilePermanent(gs, target, src)
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":   seat,
		"target": target.Card.DisplayName(),
	})
}
