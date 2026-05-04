package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerPrimeSpeakerZegana wires Prime Speaker Zegana.
//
// Oracle text:
//
//   Prime Speaker Zegana enters with X +1/+1 counters on it, where X is the greatest power among other creatures you control.
//   When Prime Speaker Zegana enters, draw cards equal to its power.
//
// Implementation collapses the "enters with" replacement and the ETB
// trigger into a single ETB handler: stamp X +1/+1 counters first
// (where X is the greatest power among OTHER creatures you control)
// then draw cards equal to Zegana's resulting power (printed power +
// counters).
func registerPrimeSpeakerZegana(r *Registry) {
	r.OnETB("Prime Speaker Zegana", primeSpeakerZeganaETB)
}

func primeSpeakerZeganaETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "prime_speaker_zegana_etb"
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
	// X = greatest power among OTHER creatures you control.
	x := 0
	for _, p := range s.Battlefield {
		if p == nil || p == perm || !p.IsCreature() {
			continue
		}
		if pow := p.Power(); pow > x {
			x = pow
		}
	}
	if x > 0 {
		perm.AddCounter("+1/+1", x)
		gs.InvalidateCharacteristicsCache()
	}
	// Draw cards equal to Zegana's power.
	pow := perm.Power()
	if pow < 0 {
		pow = 0
	}
	drawn := 0
	for i := 0; i < pow; i++ {
		if c := drawOne(gs, seat, perm.Card.DisplayName()); c != nil {
			drawn++
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     seat,
		"x":        x,
		"power":    pow,
		"drawn":    drawn,
	})
}
