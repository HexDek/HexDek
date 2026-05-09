package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerPrimeSpeakerZeganaCustom wires Zegana's ETB self-counter +
// scaling card draw. The auto-generated stub leaves the effect unwired.
//
// Oracle text:
//
//	When Prime Speaker Zegana enters, it enters with X +1/+1 counters
//	on it, where X is the greatest power among other creatures you
//	control. Then if X is 1 or more, draw X cards.
//
// Implementation:
//   - Walk the controller's battlefield, find max Power() among
//     creatures other than Zegana herself. X is that value, floored at 0.
//   - Add X +1/+1 counters to the entering Zegana.
//   - Draw X cards (drawOne in a loop) when X >= 1.
//
// "Enters with" counters are normally engine-resolved via ETB
// replacement, but the per-card hook fires after the perm is on the
// battlefield. Adding counters now produces the same end state for
// every effect that reads counters (power calculation, counter-removal
// triggers); the only edge case is a separate trigger that says "if
// Zegana entered with X counters" — we don't have any such card.
func registerPrimeSpeakerZeganaCustom(r *Registry) {
	r.OnETB("Prime Speaker Zegana", primeSpeakerZeganaETB)
}

func primeSpeakerZeganaETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "prime_speaker_zegana_etb"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	x := 0
	for _, p := range seat.Battlefield {
		if p == nil || p == perm || !p.IsCreature() {
			continue
		}
		if pw := p.Power(); pw > x {
			x = pw
		}
	}
	if x > 0 {
		perm.AddCounter("+1/+1", x)
		for i := 0; i < x && len(seat.Library) > 0; i++ {
			drawOne(gs, perm.Controller, perm.Card.DisplayName())
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"x":        x,
		"counters": x,
		"drawn":    x,
	})
}
