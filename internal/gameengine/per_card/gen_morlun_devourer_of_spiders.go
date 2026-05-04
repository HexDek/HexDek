package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMorlunDevourerOfSpiders wires Morlun, Devourer of Spiders.
//
// Oracle text:
//
//   Lifelink
//   Morlun enters with X +1/+1 counters on him.
//   When Morlun enters, he deals X damage to target opponent.
//
// Cost: {X}{B}{B}. X is read from CMC minus the colored pips. The
// X-paid amount isn't surfaced cleanly to ETB handlers, so we use
// CMC-2 as a proxy. Lifelink is a printed keyword handled elsewhere.
func registerMorlunDevourerOfSpiders(r *Registry) {
	r.OnETB("Morlun, Devourer of Spiders", morlunDevourerOfSpidersETB)
}

func morlunDevourerOfSpidersETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "morlun_devourer_of_spiders_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	x := 0
	if perm.Card != nil {
		x = perm.Card.CMC - 2
	}
	if x < 0 {
		x = 0
	}
	if x > 0 {
		perm.AddCounter("+1/+1", x)
		gs.InvalidateCharacteristicsCache()
	}
	// Pick lowest-life opponent for X damage.
	target := -1
	bestLife := 1 << 30
	for _, oppIdx := range gs.Opponents(seat) {
		opp := gs.Seats[oppIdx]
		if opp == nil || opp.Lost {
			continue
		}
		if opp.Life < bestLife {
			bestLife = opp.Life
			target = oppIdx
		}
	}
	dealt := 0
	if target >= 0 && x > 0 {
		amt, cancelled := gameengine.FireDamageEvent(gs, perm, target, nil, x)
		if !cancelled && amt > 0 {
			gs.Seats[target].Life -= amt
			dealt = amt
			gs.LogEvent(gameengine.Event{
				Kind:   "damage",
				Seat:   seat,
				Target: target,
				Source: perm.Card.DisplayName(),
				Amount: amt,
			})
			_ = gs.CheckEnd()
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     seat,
		"x":        x,
		"counters": x,
		"target":   target,
		"damage":   dealt,
	})
}
