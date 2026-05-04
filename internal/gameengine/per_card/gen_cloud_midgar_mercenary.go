package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCloudMidgarMercenary wires Cloud, Midgar Mercenary.
//
// Oracle text:
//
//   When Cloud enters, search your library for an Equipment card, reveal it, put it into your hand, then shuffle.
//   As long as Cloud is equipped, if a triggered ability of Cloud or an Equipment attached to it triggers, that ability triggers an additional time.
//
// The "trigger an additional time" static ability requires per-trigger
// duplication infrastructure (akin to Strionic Resonator on every trigger
// of attached equipment). Not currently wired — left as a partial.
func registerCloudMidgarMercenary(r *Registry) {
	r.OnETB("Cloud, Midgar Mercenary", cloudMidgarMercenaryETB)
}

func cloudMidgarMercenaryETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "cloud_midgar_mercenary_etb"
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
	bestIdx := -1
	bestCMC := -1
	for i, c := range s.Library {
		if c == nil {
			continue
		}
		if !cardHasType(c, "equipment") {
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
		emitFail(gs, slug, perm.Card.DisplayName(), "no_equipment_in_library", nil)
		emitPartial(gs, slug, perm.Card.DisplayName(), "static: equipped Cloud or attached Equipment triggers fire an additional time")
		return
	}
	card := s.Library[bestIdx]
	moveCardBetweenZones(gs, seat, card, "library", "hand", "cloud_midgar_tutor")
	shuffleLibraryPerCard(gs, seat)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  seat,
		"found": card.DisplayName(),
		"cmc":   bestCMC,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(), "static: equipped Cloud or attached Equipment triggers fire an additional time")
}
