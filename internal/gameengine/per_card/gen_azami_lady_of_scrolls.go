package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAzamiLadyOfScrolls wires Azami, Lady of Scrolls.
//
// Oracle text:
//
//   Tap an untapped Wizard you control: Draw a card.
//
// Auto-generated activated ability handler.
func registerAzamiLadyOfScrolls(r *Registry) {
	r.OnActivated("Azami, Lady of Scrolls", azamiLadyOfScrollsActivate)
}

func azamiLadyOfScrollsActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "azami_lady_of_scrolls_activate"
	if gs == nil || src == nil {
		return
	}
	seatIdx := src.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[seatIdx]
	if seat == nil || seat.Lost {
		return
	}
	// Cost gate: "Tap an untapped Wizard you control."
	// Find an untapped Wizard the controller controls — Azami herself
	// counts (she's a Wizard, per oracle types). Prefer non-Azami
	// targets so the commander stays untapped for follow-up activations.
	var tapTarget *gameengine.Permanent
	for _, p := range seat.Battlefield {
		if p == nil || p == src || p.Card == nil {
			continue
		}
		if p.Tapped {
			continue
		}
		if !p.IsCreature() {
			continue
		}
		if !cardHasSubtype(p.Card, "wizard") {
			continue
		}
		tapTarget = p
		break
	}
	if tapTarget == nil {
		// Fall back to Azami herself if she's a Wizard and untapped.
		if !src.Tapped && src.Card != nil && cardHasSubtype(src.Card, "wizard") {
			tapTarget = src
		}
	}
	if tapTarget == nil {
		emitFail(gs, slug, src.Card.DisplayName(), "no_untapped_wizard", nil)
		return
	}
	tapTarget.Tapped = true
	drawOne(gs, seatIdx, src.Card.DisplayName())
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":       seatIdx,
		"tapped":     tapTarget.Card.DisplayName(),
		"is_self":    tapTarget == src,
	})
}
