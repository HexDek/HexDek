package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMorlunDevourerOfSpiders wires Morlun, Devourer of Spiders.
//
// Oracle text (Scryfall, verified):
//
//	Lifelink
//	Morlun enters with X +1/+1 counters on him.
//	When Morlun enters, he deals X damage to target opponent.
//
// Implementation (R36 stub port):
//   - Lifelink is AST keyword pipeline.
//   - X-cost capture: OnCast stashes StackItem.ChosenX into the seat's
//     flag map as "_morlun_x_<seat>" so the ETB hook (which doesn't see
//     the stack item) can read it. Mirrors the Primo / Walking Ballista
//     pattern used elsewhere in per_card.
//   - ETB: read+consume the X flag, add X +1/+1 counters to Morlun,
//     deal X damage to the highest-life living opponent. Auto-target
//     selection matches the standard "damage target opponent" heuristic
//     used by per_card cards without a player-choice layer (Niv-Mizzet,
//     Parun's draw-damage handler picks the same way).
func registerMorlunDevourerOfSpiders(r *Registry) {
	r.OnCast("Morlun, Devourer of Spiders", morlunOnCastCaptureX)
	r.OnETB("Morlun, Devourer of Spiders", morlunETBCountersAndDamage)
}

func morlunOnCastCaptureX(gs *gameengine.GameState, item *gameengine.StackItem) {
	if gs == nil || item == nil {
		return
	}
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	gs.Flags["_morlun_x_"+intToStr(item.Controller)] = item.ChosenX
}

func morlunETBCountersAndDamage(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "morlun_devourer_of_spiders_etb"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seatIdx := perm.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}

	// Read+consume X.
	x := 0
	key := "_morlun_x_" + intToStr(seatIdx)
	if gs.Flags != nil {
		if v, ok := gs.Flags[key]; ok {
			x = v
			delete(gs.Flags, key)
		}
	}
	if x < 0 {
		x = 0
	}

	// +1/+1 counters on Morlun.
	if x > 0 {
		perm.AddCounter("+1/+1", x)
		gs.InvalidateCharacteristicsCache()
	}

	// Deal X damage to highest-life opponent (auto-targeting).
	targetSeat := -1
	bestLife := -1
	for _, opp := range gs.Opponents(seatIdx) {
		s := gs.Seats[opp]
		if s == nil || s.Lost {
			continue
		}
		if targetSeat < 0 || s.Life > bestLife {
			targetSeat = opp
			bestLife = s.Life
		}
	}
	dmgDealt := 0
	if x > 0 && targetSeat >= 0 {
		gameengine.DealDamage(gs, targetSeat, x, perm.Card.DisplayName())
		dmgDealt = x
	}

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         seatIdx,
		"x_value":      x,
		"counters_added": x,
		"target_seat":  targetSeat,
		"damage_dealt": dmgDealt,
	})
}
