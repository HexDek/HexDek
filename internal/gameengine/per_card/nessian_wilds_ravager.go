package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerNessianWildsRavager wires Nessian Wilds Ravager (Muninn parser-gap
// #63, ~11.6K hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{4}{G}{G}
//	Creature — Hydra
//	Tribute 6 (As this creature enters, an opponent of your choice may
//	put six +1/+1 counters on it.)
//	When this creature enters, if tribute wasn't paid, you may have
//	this creature fight another target creature. (Each deals damage
//	equal to its power to the other.)
//
// Implementation:
//   - Tribute is not modeled by the engine — no opponent prompt exists
//     for the cast-time choice. We default to "tribute NOT paid" (Hat
//     opponents rationally refuse: a 5/5 fights for free vs. an 11/11
//     fights for free, and the fight target is usually their own best
//     creature — they decline). emitPartial flags it.
//   - On ETB, scan opposing battlefields for the highest-power creature
//     and fight it: Nessian and the target each take damage equal to the
//     other's power. Damage is marked via "damage_marked" counter
//     (Ulrich Back-Fight precedent); SBAs clear lethals later.
func registerNessianWildsRavager(r *Registry) {
	r.OnETB("Nessian Wilds Ravager", nessianWildsRavagerETB)
}

func nessianWildsRavagerETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "nessian_wilds_ravager_fight"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	var target *gameengine.Permanent
	for _, opp := range gs.Opponents(perm.Controller) {
		if opp < 0 || opp >= len(gs.Seats) {
			continue
		}
		s := gs.Seats[opp]
		if s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || !p.IsCreature() {
				continue
			}
			if target == nil || p.Power() > target.Power() {
				target = p
			}
		}
	}
	if target == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "no_opposing_creature",
		})
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"tribute_choice_unmodeled_assumed_not_paid")
		return
	}
	nessianPower := perm.Power()
	targetPower := target.Power()
	target.AddCounter("damage_marked", nessianPower)
	perm.AddCounter("damage_marked", targetPower)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          perm.Controller,
		"target":        target.Card.DisplayName(),
		"nessian_power": nessianPower,
		"target_power":  targetPower,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"tribute_choice_unmodeled_assumed_not_paid")
}
