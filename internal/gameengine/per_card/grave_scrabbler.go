package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGraveScrabbler wires Grave Scrabbler (Muninn parser-gap #92, 4.8K
// hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{3}{B}
//	Creature — Zombie
//	Madness {1}{B}
//	When this creature enters, if its madness cost was paid, you may
//	return target creature card from a graveyard to its owner's hand.
//
// Implementation:
//   - Madness keyword cast pipeline lives in keywords_batch6.go
//     (ActivateMadness — pushes StackItem with CostMeta["madness"]).
//   - The cast-time CostMeta does not currently propagate to perm.Flags
//     when the spell resolves into a permanent, so we approximate: gate
//     on perm.Flags["was_cast"] and emitPartial about the missing
//     madness-specific gate. The recurring-from-graveyard effect is
//     "may" so over-firing is strictly value-positive — the symmetric
//     case (paid hard cost instead of madness, no recursion) leaks
//     extra hand returns rather than illegal state.
//   - Target selection: prefer the controller's own graveyard, picking
//     the highest-CMC creature card (biggest tempo swing). Fall back
//     to any opponent's graveyard if our own is creature-empty.
func registerGraveScrabbler(r *Registry) {
	r.OnETB("Grave Scrabbler", graveScrabblerETB)
}

func graveScrabblerETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "grave_scrabbler_madness_recur"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil || perm.Flags["was_cast"] == 0 {
		emitFail(gs, slug, perm.Card.DisplayName(), "not_cast", map[string]interface{}{
			"seat": perm.Controller,
		})
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	target, targetSeat := graveScrabblerPickRecur(gs, seat)
	if target == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_creature_in_graveyards", map[string]interface{}{
			"seat": seat,
		})
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"madness_cost_paid_flag_not_propagated_from_cost_meta_gate_via_was_cast")
		return
	}
	gameengine.MoveCard(gs, target, targetSeat, "graveyard", "hand", "grave_scrabbler_recur")
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          seat,
		"target":        target.DisplayName(),
		"target_seat":   targetSeat,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"madness_cost_paid_flag_not_propagated_from_cost_meta_gate_via_was_cast")
}

// graveScrabblerPickRecur scans graveyards starting with our own, then
// opponents'. Returns the highest-CMC creature card found and the seat
// index of the graveyard it came from. (Returning to OWNER's hand per
// oracle — owner can differ from the graveyard seat only for stolen
// cards; we route MoveCard from the graveyard seat which is the
// physical zone the card lives in.)
func graveScrabblerPickRecur(gs *gameengine.GameState, ourSeat int) (*gameengine.Card, int) {
	var best *gameengine.Card
	bestCMC := -1
	bestSeat := -1
	tryScan := func(seatIdx int) {
		s := gs.Seats[seatIdx]
		if s == nil {
			return
		}
		for _, c := range s.Graveyard {
			if c == nil || !cardHasType(c, "creature") {
				continue
			}
			cmc := cardCMC(c)
			if cmc > bestCMC {
				bestCMC = cmc
				best = c
				bestSeat = seatIdx
			}
		}
	}
	tryScan(ourSeat)
	if best != nil {
		return best, bestSeat
	}
	for _, opp := range gs.Opponents(ourSeat) {
		tryScan(opp)
	}
	return best, bestSeat
}
