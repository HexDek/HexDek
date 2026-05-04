package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSunSpiderNimbleWebber wires Sun-Spider, Nimble Webber.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	During your turn, Sun-Spider has flying.
//	When Sun-Spider enters, search your library for an Aura or Equipment
//	  card, reveal it, put it into your hand, then shuffle.
//
// Implementation:
//   - OnETB: search our library for an Aura or Equipment card. Heuristic
//     pick: highest-CMC Equipment first, else highest-CMC Aura. Put it
//     into our hand. (Library shuffle is implicit — we don't reorder
//     post-removal.)
//   - "During your turn, has flying" handled at static-effect layer —
//     emitPartial.
func registerSunSpiderNimbleWebber(r *Registry) {
	r.OnETB("Sun-Spider, Nimble Webber", sunSpiderETB)
}

func sunSpiderETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "sun_spider_etb_tutor"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}

	var bestEq *gameengine.Card
	bestEqCMC := -1
	var bestAura *gameengine.Card
	bestAuraCMC := -1
	for _, c := range seat.Library {
		if c == nil {
			continue
		}
		cm := gameengine.ManaCostOf(c)
		if cardHasType(c, "equipment") && cm > bestEqCMC {
			bestEqCMC = cm
			bestEq = c
		} else if cardHasType(c, "aura") && cm > bestAuraCMC {
			bestAuraCMC = cm
			bestAura = c
		}
	}
	pick := bestEq
	if pick == nil {
		pick = bestAura
	}
	if pick == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_aura_or_equipment_in_library", nil)
		return
	}
	gameengine.MoveCard(gs, pick, perm.Controller, "library", "hand", "sun_spider_tutor")
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"found": pick.DisplayName(),
	})
	emitPartial(gs, "sun_spider_static_flying", perm.Card.DisplayName(),
		"during_your_turn_flying_continuous_static_not_modeled")
}
