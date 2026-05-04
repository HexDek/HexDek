package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerIllunaApexOfWishes wires Illuna, Apex of Wishes.
//
// Oracle text:
//
//	Mutate {3}{R/G}{U}{U}
//	Flying, trample
//	Whenever this creature mutates, exile cards from the top of your
//	library until you exile a nonland permanent card. Put that card
//	onto the battlefield or into your hand.
//
// Mutate isn't a tracked engine event — emitPartial.
func registerIllunaApexOfWishes(r *Registry) {
	r.OnETB("Illuna, Apex of Wishes", illunaApexETB)
}

func illunaApexETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "illuna_apex_partial"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"mutate_trigger_and_cheat_into_play_unimplemented")
}
