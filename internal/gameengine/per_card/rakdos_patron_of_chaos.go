package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRakdosPatronOfChaos wires Rakdos, Patron of Chaos.
//
// Oracle text:
//
//	Flying, trample
//	At the beginning of your end step, target opponent may sacrifice
//	two nonland, nontoken permanents of their choice. If they don't,
//	you draw two cards.
//
// End-step observers aren't dispatched through per_card; recorded as a
// parser gap so Heimdall reports it. We still wire the registration so
// the card is recognized.
func registerRakdosPatronOfChaos(r *Registry) {
	r.OnETB("Rakdos, Patron of Chaos", rakdosPatronETB)
}

func rakdosPatronETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "rakdos_patron_etb"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"end_step_sacrifice_or_draw_two_unimplemented")
}
