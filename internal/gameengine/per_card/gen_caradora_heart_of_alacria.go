package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCaradoraHeartOfAlacria wires Caradora, Heart of Alacria.
//
// Oracle text:
//
//   When Caradora enters, you may search your library for a Mount or Vehicle card, reveal it, put it into your hand, then shuffle.
//   If one or more +1/+1 counters would be put on a creature or Vehicle you control, that many plus one +1/+1 counters are put on it instead.
//
// The static replacement clause (extra +1/+1 counter) requires hooking into
// the engine's counter-placement replacement layer (CR §614). Not currently
// wired — left as a partial.
func registerCaradoraHeartOfAlacria(r *Registry) {
	r.OnETB("Caradora, Heart of Alacria", caradoraHeartOfAlacriaETB)
}

func caradoraHeartOfAlacriaETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "caradora_heart_of_alacria_etb"
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
	// Greedy tutor: prefer a Mount or Vehicle with the highest CMC (best
	// payoff). Falls back to any Mount/Vehicle if no CMC data is available.
	bestIdx := -1
	bestCMC := -1
	for i, c := range s.Library {
		if c == nil {
			continue
		}
		if !cardHasType(c, "mount") && !cardHasType(c, "vehicle") {
			continue
		}
		cmc := cardCMC(c)
		if cmc > bestCMC {
			bestCMC = cmc
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		shuffleLibraryPerCard(gs, seat)
		emitFail(gs, slug, perm.Card.DisplayName(), "no_mount_or_vehicle_in_library", nil)
		// Static replacement clause unimplemented (needs Layer/replacement
		// hook for +1/+1 counter placement).
		emitPartial(gs, slug, perm.Card.DisplayName(), "static replacement: extra +1/+1 counter on creature/Vehicle you control")
		return
	}
	card := s.Library[bestIdx]
	moveCardBetweenZones(gs, seat, card, "library", "hand", "caradora_tutor")
	shuffleLibraryPerCard(gs, seat)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  seat,
		"found": card.DisplayName(),
		"cmc":   bestCMC,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(), "static replacement: extra +1/+1 counter on creature/Vehicle you control")
}
