package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMorlunDevourerOfSpidersCustom implements Morlun's enter-with-X
// counters + X damage ETB. The auto-generated stub is a no-op.
//
// Oracle text:
//
//	Lifelink
//	Morlun enters with X +1/+1 counters on him.
//	When Morlun enters, he deals X damage to target opponent.
//
// Implementation notes:
//   - Lifelink is a static keyword handled by the AST pipeline.
//   - X is the value paid for the {X} in Morlun's mana cost. The
//     engine does not yet thread X through to ETB hooks; we recover
//     a best-effort X from perm.Flags["x_paid"] (the standard X
//     conduit other handlers use). Falls back to 0 when unknown.
//   - Damage target = the opponent with lowest life, since X damage
//     is most valuable as a kill-shot and otherwise as a chunk to
//     the leader.
func registerMorlunDevourerOfSpidersCustom(r *Registry) {
	r.OnETB("Morlun, Devourer of Spiders", morlunETB)
}

func morlunETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "morlun_x_counters_and_damage"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seatIdx := perm.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}

	x := 0
	if perm.Flags != nil {
		x = perm.Flags["x_paid"]
	}
	if x <= 0 {
		// Defensive default: when X wasn't recorded, ETB at minimum.
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":  seatIdx,
			"x":     0,
			"note":  "x_unknown_at_etb",
		})
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"X paid value not threaded into ETB hook — using 0")
		return
	}

	perm.AddCounter("+1/+1", x)
	gs.InvalidateCharacteristicsCache()

	// Pick lowest-life opponent.
	target := -1
	lowest := 1 << 30
	for i, s := range gs.Seats {
		if i == seatIdx || s == nil || s.Lost || s.LeftGame {
			continue
		}
		if s.Life < lowest {
			lowest = s.Life
			target = i
		}
	}
	if target >= 0 {
		gameengine.DealDamage(gs, target, x, perm.Card.DisplayName())
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         seatIdx,
		"x":            x,
		"counters":     x,
		"damage":       x,
		"target_seat":  target,
	})
	_ = gs.CheckEnd()
}
