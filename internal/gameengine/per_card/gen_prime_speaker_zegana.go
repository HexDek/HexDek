package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerPrimeSpeakerZegana wires Prime Speaker Zegana.
//
// Oracle text (MKC reprint, {2}{G}{G}{U}{U}, 1/1):
//
//	Prime Speaker Zegana enters with X +1/+1 counters on it, where X
//	is the greatest power among other creatures you control.
//	When Prime Speaker Zegana enters, draw cards equal to its power.
//
// Implementation:
//   - ETB scans the controller's battlefield for the highest-power
//     other creature, applies that many +1/+1 counters to Zegana,
//     then draws cards equal to her resulting power (base 1 + counters).
//   - "Enters with counters" is a replacement effect; modeling it as
//     post-ETB counter application + draw matches behavior in
//     isolation but skips edge cases where Zegana's ETB power is
//     read by simultaneously-resolving triggers (rare).
func registerPrimeSpeakerZegana(r *Registry) {
	r.OnETB("Prime Speaker Zegana", primeSpeakerZeganaETB)
}

func primeSpeakerZeganaETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "zegana_etb_counters_and_draw"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	bestPower := 0
	for _, p := range seat.Battlefield {
		if p == nil || p == perm || p.Card == nil || !p.IsCreature() {
			continue
		}
		if pw := p.Power(); pw > bestPower {
			bestPower = pw
		}
	}
	if bestPower > 0 {
		perm.AddCounter("+1/+1", bestPower)
	}
	drawCount := perm.Power()
	if drawCount < 0 {
		drawCount = 0
	}
	drawn := 0
	for i := 0; i < drawCount; i++ {
		if c := drawOne(gs, perm.Controller, perm.Card.DisplayName()); c != nil {
			drawn++
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"counters": bestPower,
		"drew":     drawn,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"counters_applied_post_etb_not_via_replacement_effect")
}
