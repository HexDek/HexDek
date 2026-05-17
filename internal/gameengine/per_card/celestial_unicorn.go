package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCelestialUnicorn wires Celestial Unicorn (Muninn parser-gap, ~150
// hits across recurring lifegain decks).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{2}{W}
//	Creature — Unicorn
//	Whenever you gain life, put a +1/+1 counter on this creature.
//
// Implementation (mirrors Aerith Gainsborough's life-gained self-buff
// pattern):
//   - life_gained gated on seat == controller. Single +1/+1 counter per
//     trigger regardless of amount gained (CR §603.4 — the trigger fires
//     once per gain event, not per life point).
func registerCelestialUnicorn(r *Registry) {
	r.OnTrigger("Celestial Unicorn", "life_gained", celestialUnicornLifeGained)
}

func celestialUnicornLifeGained(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "celestial_unicorn_counter_on_life_gain"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	seat, _ := ctx["seat"].(int)
	if seat != perm.Controller {
		return
	}
	amount, _ := ctx["amount"].(int)
	if amount <= 0 {
		return
	}
	perm.AddCounter("+1/+1", 1)
	gs.InvalidateCharacteristicsCache()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"amount": amount,
	})
}
