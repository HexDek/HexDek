package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMikeyAndLeo wires Mikey & Leo, Chaos & Order.
//
// Oracle text:
//
//	Whenever you put a counter on a creature you control, draw a card.
//	This ability triggers only once each turn.
//
// We don't have a "counter_placed" trigger event, so the counter-placed
// detection is left as a parser gap. The handler still registers so the
// runtime knows the card was recognized.
func registerMikeyAndLeo(r *Registry) {
	r.OnETB("Mikey & Leo, Chaos & Order", mikeyAndLeoETB)
}

func mikeyAndLeoETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "mikey_and_leo_etb"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"counter_placed_trigger_event_not_wired")
}
