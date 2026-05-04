package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTheEverChangingDane wires The Ever-Changing 'Dane.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	{1}, Sacrifice another creature: The Ever-Changing 'Dane becomes a
//	  copy of the sacrificed creature, except it has this ability.
//
// The activated copy effect is layer-bound. Implement the sacrifice and
// emitPartial for the copy.
func registerTheEverChangingDane(r *Registry) {
	r.OnActivated("The Ever-Changing 'Dane", theEverChangingDaneActivate)
}

func theEverChangingDaneActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "ever_changing_dane_copy_sacrifice"
	if gs == nil || src == nil {
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	var victim *gameengine.Permanent
	bestPow := -1
	for _, p := range seat.Battlefield {
		if p == nil || p == src || !p.IsCreature() {
			continue
		}
		if pw := p.Power(); pw > bestPow {
			bestPow = pw
			victim = p
		}
	}
	if victim == nil {
		emitFail(gs, slug, src.Card.DisplayName(), "no_sacrifice_target", nil)
		return
	}
	victimName := victim.Card.DisplayName()
	gameengine.SacrificePermanent(gs, victim, "ever_changing_dane")
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":       src.Controller,
		"sacrificed": victimName,
	})
	emitPartial(gs, slug, src.Card.DisplayName(),
		"becomes_a_copy_of_sacrificed_creature_continuous_static_not_modeled")
}
