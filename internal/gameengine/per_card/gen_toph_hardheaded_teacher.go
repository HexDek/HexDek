package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTophHardheadedTeacher wires Toph, Hardheaded Teacher.
//
// Oracle text:
//
//   When Toph enters, you may discard a card. If you do, return target instant or sorcery card from your graveyard to your hand.
//   Whenever you cast a spell, earthbend 1. If that spell is a Lesson, put an additional +1/+1 counter on that land. (Target land you control becomes a 0/0 creature with haste that's still a land. Put a +1/+1 counter on it. When it dies or is exiled, return it to the battlefield tapped.)
//
// Implementation:
//   - ETB: if there's an instant or sorcery in graveyard, discard the
//     lowest-value card in hand and return that instant/sorcery.
//   - Cast trigger: earthbend 1 (turn one of our lands into a 0/0
//     creature with a +1/+1 counter and temp_haste). Lesson rider
//     emits a partial — Lesson detection isn't tracked.
func registerTophHardheadedTeacher(r *Registry) {
	r.OnETB("Toph, Hardheaded Teacher", tophHardheadedTeacherETB)
	r.OnTrigger("Toph, Hardheaded Teacher", "spell_cast", tophHardheadedSpellCast)
}

func tophHardheadedTeacherETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "toph_hardheaded_teacher_etb"
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
	// Find an instant or sorcery in graveyard.
	var target *gameengine.Card
	targetIdx := -1
	for i, c := range s.Graveyard {
		if c == nil {
			continue
		}
		if cardHasType(c, "instant") || cardHasType(c, "sorcery") {
			target = c
			targetIdx = i
			break
		}
	}
	if target == nil || len(s.Hand) == 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":     seat,
			"returned": false,
		})
		return
	}
	// Discard the lowest-CMC card in hand.
	var pickIdx = -1
	var pickCMC = 1 << 30
	for i, c := range s.Hand {
		if c == nil {
			continue
		}
		if cmc := cardCMC(c); cmc < pickCMC {
			pickCMC = cmc
			pickIdx = i
		}
	}
	if pickIdx < 0 {
		return
	}
	discarded := s.Hand[pickIdx]
	gameengine.DiscardCard(gs, discarded, seat)
	// Return the instant/sorcery to hand. DiscardCard may have shifted
	// graveyard indices, so re-locate by pointer.
	for i, c := range s.Graveyard {
		if c == target {
			targetIdx = i
			break
		}
	}
	if targetIdx >= 0 && targetIdx < len(s.Graveyard) && s.Graveyard[targetIdx] == target {
		gameengine.MoveCard(gs, target, seat, "graveyard", "hand", "toph_hardheaded_return")
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      seat,
		"discarded": discarded.DisplayName(),
		"returned":  target.DisplayName(),
	})
}

func tophHardheadedSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "toph_hardheaded_earthbend_1"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	// Pick a non-creature land we control to earthbend.
	var target *gameengine.Permanent
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if p.IsLand() && !p.IsCreature() {
			target = p
			break
		}
	}
	if target == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"reason": "no_land_target",
		})
		return
	}
	if target.Flags == nil {
		target.Flags = map[string]int{}
	}
	target.Flags["earthbent"] = 1
	target.Flags["temp_haste"] = 1
	if target.Counters == nil {
		target.Counters = map[string]int{}
	}
	target.Counters["+1/+1"]++
	target.SummoningSick = false
	gs.InvalidateCharacteristicsCache()
	gs.LogEvent(gameengine.Event{
		Kind: "earthbend", Seat: perm.Controller,
		Source: perm.Card.DisplayName(), Amount: 1,
		Details: map[string]interface{}{
			"target": target.Card.DisplayName(),
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"target": target.Card.DisplayName(),
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"lesson_extra_counter_rider_unimplemented")
}
