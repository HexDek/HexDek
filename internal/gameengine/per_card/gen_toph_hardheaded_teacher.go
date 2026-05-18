package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTophHardheadedTeacher wires Toph, Hardheaded Teacher.
//
// Oracle text (Scryfall, verified):
//
//	When Toph enters, you may discard a card. If you do, return target
//	instant or sorcery card from your graveyard to your hand.
//	Whenever you cast a spell, earthbend 1. If that spell is a Lesson,
//	put an additional +1/+1 counter on that land. (Target land you
//	control becomes a 0/0 creature with haste that's still a land. Put
//	a +1/+1 counter on it. When it dies or is exiled, return it to
//	the battlefield tapped.)
//
// Implementation (R36 stub port):
//   - ETB "may-discard-to-rebuy-instant/sorcery" is the primary ability
//     this port models. AI policy: opt in IFF the controller has BOTH
//     a discardable card in hand AND an instant/sorcery card in their
//     graveyard (otherwise the trade is pure loss). Picks the
//     lowest-MV hand card to discard and the highest-MV i/s in the
//     graveyard to return — maximises swap value.
//   - Earthbend trigger ("Whenever you cast a spell, earthbend 1") is
//     emitPartial — earthbending (turn-a-land-into-a-0/0-creature) is
//     a discrete mechanic that would need its own helper; the per-card
//     port flags the gap.
func registerTophHardheadedTeacher(r *Registry) {
	r.OnETB("Toph, Hardheaded Teacher", tophETBDiscardRebuy)
}

func tophETBDiscardRebuy(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "toph_hardheaded_teacher_etb_discard_rebuy"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seatIdx := perm.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[seatIdx]
	if seat == nil || seat.Lost {
		return
	}

	// Pick the highest-MV instant/sorcery in graveyard as the return target.
	// If no eligible card, we won't opt-in.
	var returnTarget *gameengine.Card
	bestCMC := -1
	for _, c := range seat.Graveyard {
		if c == nil {
			continue
		}
		if !cardHasType(c, "instant") && !cardHasType(c, "sorcery") {
			continue
		}
		cmc := cardCMC(c)
		if returnTarget == nil || cmc > bestCMC {
			returnTarget = c
			bestCMC = cmc
		}
	}
	// Pick the lowest-MV hand card as the discard victim.
	var discardVictim *gameengine.Card
	discardCMC := -1
	for _, c := range seat.Hand {
		if c == nil {
			continue
		}
		cmc := cardCMC(c)
		if discardVictim == nil || cmc < discardCMC {
			discardVictim = c
			discardCMC = cmc
		}
	}

	// AI policy: opt in only when BOTH conditions met.
	if returnTarget == nil || discardVictim == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      seatIdx,
			"opted_in":  false,
			"reason":    tophDeclineReason(returnTarget, discardVictim),
		})
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"earthbend_on_cast_trigger_not_implemented_separate_helper_needed")
		return
	}

	// Discard the chosen hand card.
	gameengine.DiscardCard(gs, discardVictim, seatIdx)
	// Return the chosen graveyard card to hand.
	gameengine.MoveCard(gs, returnTarget, seatIdx, "graveyard", "hand", "toph_etb_return")

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       seatIdx,
		"opted_in":   true,
		"discarded":  discardVictim.DisplayName(),
		"returned":   returnTarget.DisplayName(),
		"return_cmc": bestCMC,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"earthbend_on_cast_trigger_not_implemented_separate_helper_needed")
}

func tophDeclineReason(returnTarget, discardVictim *gameengine.Card) string {
	switch {
	case returnTarget == nil && discardVictim == nil:
		return "no_eligible_graveyard_card_and_empty_hand"
	case returnTarget == nil:
		return "no_instant_or_sorcery_in_graveyard"
	default:
		return "empty_hand_no_discard_victim"
	}
}
