package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerArchmageAscension wires Archmage Ascension.
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	At the beginning of each end step, if you drew two or more cards
//	this turn, you may put a quest counter on this enchantment.
//	As long as this enchantment has six or more quest counters on it,
//	if you would draw a card, you may instead search your library for
//	a card, put that card into your hand, then shuffle.
//
// Implementation (Muninn gap #35 — 27K hits):
//   - OnTrigger("end_step"): "each end step" — fires for every player's
//     end step, gated on the controller having drawn 2+ this turn. AI
//     policy auto-accepts the "may" (monotone upside).
//   - The 6+ counter draw-replacement is a §614 replacement effect that
//     requires registering against the controller's draw event. The
//     engine's per-card replacement registration is limited to the
//     built-in table; emitPartial.
func registerArchmageAscension(r *Registry) {
	r.OnTrigger("Archmage Ascension", "end_step", archmageAscensionEndStep)
}

func archmageAscensionEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "archmage_ascension_quest"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	seatIdx := perm.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[seatIdx]
	if seat == nil || seat.Lost {
		return
	}
	drawn := seat.Turn.CardsDrawn
	if drawn < 2 {
		if seat.Flags != nil && seat.Flags["cards_drawn_this_turn"] >= 2 {
			drawn = seat.Flags["cards_drawn_this_turn"]
		} else {
			emitFail(gs, slug, perm.Card.DisplayName(), "drew_less_than_two_this_turn", map[string]interface{}{
				"seat":  seatIdx,
				"drawn": drawn,
			})
			return
		}
	}
	if perm.Counters == nil {
		perm.Counters = map[string]int{}
	}
	perm.AddCounter("quest", 1)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":           seatIdx,
		"drawn":          drawn,
		"quest_counters": perm.Counters["quest"],
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"draw_replacement_at_6_counters_needs_per_card_replacement_effect_registration")
}
