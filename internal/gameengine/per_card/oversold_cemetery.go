package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerOversoldCemetery wires Oversold Cemetery.
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	At the beginning of your upkeep, if you have four or more creature
//	cards in your graveyard, you may return target creature card from
//	your graveyard to your hand.
//
// Implementation (Muninn gap #12 — 76K hits):
//   - OnTrigger("upkeep") gated on active_seat == controller (CR
//     §603.1 — "your upkeep" only fires for the controller's own
//     upkeep step).
//   - Count creature cards in the controller's graveyard. If <4, no
//     trigger.
//   - "you may" + "target creature card from your graveyard": auto-accept
//     and pick the highest-CMC creature card (best body to re-cast next
//     turn). Move it back to hand.
//   - The choice gate is monotonic upside for the controller, so the
//     hat's "may" answer is always yes.
func registerOversoldCemetery(r *Registry) {
	r.OnTrigger("Oversold Cemetery", "upkeep", oversoldCemeteryUpkeep)
}

func oversoldCemeteryUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "oversold_cemetery_recur"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil || s.Lost {
		return
	}

	creatureCards := 0
	var best *gameengine.Card
	bestCMC := -1
	for _, c := range s.Graveyard {
		if c == nil || !cardHasType(c, "creature") {
			continue
		}
		creatureCards++
		cmc := gameengine.ManaCostOf(c)
		if cmc > bestCMC {
			bestCMC = cmc
			best = c
		}
	}
	if creatureCards < 4 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":            seat,
			"creature_cards":  creatureCards,
			"triggered":       false,
		})
		return
	}
	if best == nil {
		// Defensive: count >= 4 but no eligible target found.
		emitFail(gs, slug, perm.Card.DisplayName(), "no_eligible_creature_card", map[string]interface{}{
			"seat":           seat,
			"creature_cards": creatureCards,
		})
		return
	}

	gameengine.MoveCard(gs, best, seat, "graveyard", "hand", "oversold_cemetery_recur")

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":           seat,
		"creature_cards": creatureCards,
		"returned":       best.DisplayName(),
		"returned_cmc":   bestCMC,
		"triggered":      true,
	})
}
