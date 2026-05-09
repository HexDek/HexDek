package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerChainerDementiaMasterCustom implements Chainer's reanimate
// activation. The auto-generated static stub omits the activated effect.
//
// Oracle text:
//
//	All Nightmares get +1/+1.
//	{B}{B}{B}, Pay 3 life: Put target creature card from a graveyard
//	onto the battlefield under your control. That creature is black
//	and is a Nightmare in addition to its other creature types.
//	When Chainer leaves the battlefield, exile all Nightmares.
//
// {B}{B}{B} is enforced by engine cost dispatch. The "Pay 3 life"
// secondary cost is NOT in the standard mana cost; we enforce it here
// since the audit confirms the gen_*.go template doesn't pay it. The
// graveyard search ranges over EVERY player's graveyard ("a graveyard"
// — CR §503.2 cf. ICR), not just Chainer's controller's. The leaves-
// the-battlefield exile rider is wired as a partial (LTB hook isn't
// exposed yet for this surface).
func registerChainerDementiaMasterCustom(r *Registry) {
	r.OnActivated("Chainer, Dementia Master", chainerReanimate)
	r.OnETB("Chainer, Dementia Master", func(gs *gameengine.GameState, perm *gameengine.Permanent) {
		emitPartial(gs, "chainer_ltb_exile", perm.Card.DisplayName(),
			"LTB exile-all-Nightmares rider needs engine-side LTB hook")
		emitPartial(gs, "chainer_nightmare_anthem", perm.Card.DisplayName(),
			"static +1/+1 anthem for Nightmares handled by AST/layers")
	})
}

func chainerReanimate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "chainer_reanimate"
	if gs == nil || src == nil {
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	// Pay 3 life as the additional cost — engine cost dispatch only
	// handles the {B}{B}{B} mana side. Reject if life would drop to 0
	// or below per §704.5a (Chainer's controller can't pay if it'd lose).
	if seat.Life <= 3 {
		emitFail(gs, slug, src.Card.DisplayName(), "insufficient_life", map[string]interface{}{
			"life": seat.Life,
		})
		return
	}
	gameengine.LoseLife(gs, src.Controller, 3, src.Card.DisplayName())

	// Find best creature card across ALL players' graveyards.
	var best *gameengine.Card
	bestSeat := -1
	bestPower := -1
	for i, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, c := range s.Graveyard {
			if c == nil {
				continue
			}
			if !cardHasType(c, "creature") {
				continue
			}
			if c.BasePower > bestPower {
				best = c
				bestSeat = i
				bestPower = c.BasePower
			}
		}
	}
	if best == nil {
		emitFail(gs, slug, src.Card.DisplayName(), "no_creature_in_any_graveyard", nil)
		return
	}
	// Remove from owner's graveyard.
	owner := gs.Seats[bestSeat]
	for i, c := range owner.Graveyard {
		if c == best {
			owner.Graveyard = append(owner.Graveyard[:i], owner.Graveyard[i+1:]...)
			break
		}
	}
	// Tag as black + Nightmare.
	best.Types = append(best.Types, "color:black", "nightmare")
	enterBattlefieldWithETB(gs, src.Controller, best, false)
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":           src.Controller,
		"reanimated":     best.DisplayName(),
		"original_owner": bestSeat,
		"life_paid":      3,
	})
}
