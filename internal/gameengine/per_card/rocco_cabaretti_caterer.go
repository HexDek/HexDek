package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRoccoCabarettiCaterer wires Rocco, Cabaretti Caterer
// (Muninn parser-gap #48, 16,322 hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{X}{R}{G}{W}
//	Legendary Creature — Elf Druid
//	When Rocco enters, if you cast it, you may search your library for
//	a creature card with mana value X or less, put it onto the
//	battlefield, then shuffle.
//
// Implementation:
//   - ETB-cast trigger: gate on perm.Flags["was_cast"] == 1. Non-cast
//     entries (reanimate, blink) fizzle.
//   - X value: read perm.Flags["x_paid"] (the standard X conduit;
//     see custom_morlun_devourer_of_spiders.go, bruce_banner.go).
//     If x_paid is 0 (engine didn't thread X through), we fall back
//     to the curve-cheapest creature (CMC ≤ 0) which is effectively
//     a "no tutor" — emitPartial flags the X-tracking gap.
//   - "you may" — Hat policy: always tutor (free creature is upside).
//   - Pick policy: highest-CMC creature with CMC ≤ X (Rocco rewards
//     spending X for value). Ties broken by library index (deterministic).
//   - Distinct from Rocco, Street Chef (handled separately).
//   - Note the Rocco STR registration: the engine routes "Rocco,
//     Cabaretti Caterer" to its dedicated handler; the bare "Rocco"
//     trigger key isn't shared with Street Chef.
func registerRoccoCabarettiCaterer(r *Registry) {
	r.OnETB("Rocco, Cabaretti Caterer", roccoCabarettiETB)
}

func roccoCabarettiETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "rocco_cabaretti_caterer_etb_tutor"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil || perm.Flags["was_cast"] != 1 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":     perm.Controller,
			"was_cast": false,
		})
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	x := perm.Flags["x_paid"]
	if x <= 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"x":      0,
			"reason": "x_unknown_at_etb",
		})
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"x_paid_not_threaded_into_etb_hook_no_tutor")
		return
	}
	var best *gameengine.Card
	bestCMC := -1
	for _, c := range seat.Library {
		if c == nil || !cardHasType(c, "creature") {
			continue
		}
		cmc := cardCMC(c)
		if cmc > x {
			continue
		}
		if cmc > bestCMC {
			bestCMC = cmc
			best = c
		}
	}
	if best == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat": perm.Controller,
			"x":    x,
			"found": "none",
		})
		shuffleLibraryPerCard(gs, perm.Controller)
		return
	}
	gameengine.MoveCard(gs, best, perm.Controller, "library", "battlefield", slug)
	shuffleLibraryPerCard(gs, perm.Controller)
	gs.LogEvent(gameengine.Event{
		Kind:   "search_library",
		Seat:   perm.Controller,
		Source: perm.Card.DisplayName(),
		Details: map[string]interface{}{
			"found":  []string{best.DisplayName()},
			"reason": slug,
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"x":         x,
		"into_play": best.DisplayName(),
		"cmc":       bestCMC,
	})
}
