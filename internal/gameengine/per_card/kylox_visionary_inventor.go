package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKyloxVisionaryInventor wires Kylox, Visionary Inventor.
//
// Oracle text:
//
//	Menace, ward {2}, haste
//	Whenever Kylox attacks, sacrifice any number of other creatures, then
//	exile the top X cards of your library, where X is their total power.
//	You may cast any number of instant and/or sorcery spells from among
//	the exiled cards without paying their mana costs.
//
// Implementation: on attack, choose any creature with power <= 2 (sub-
// optimal but cheap heuristic) and sacrifice them. Exile X cards from
// the top of the library. Free-casting from exile is non-trivial — we
// log the exiled cards and emitPartial for the casting clause.
func registerKyloxVisionaryInventor(r *Registry) {
	r.OnTrigger("Kylox, Visionary Inventor", "attacks", kyloxAttacks)
}

func kyloxAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kylox_attack_sac_exile"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	attackerSeat, _ := ctx["seat"].(int)
	if attackerSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	totalPower := 0
	victims := []*gameengine.Permanent{}
	for _, p := range seat.Battlefield {
		if p == nil || p == perm || !p.IsCreature() || p.Card == nil {
			continue
		}
		// Heuristic: sacrifice tokens and small creatures to fuel the dig.
		isToken := cardHasType(p.Card, "token")
		if !isToken && p.Card.BasePower > 3 {
			continue
		}
		victims = append(victims, p)
		totalPower += p.Card.BasePower
	}
	for _, v := range victims {
		gameengine.SacrificePermanent(gs, v, "kylox_attack_sac")
	}
	exiled := 0
	for i := 0; i < totalPower && len(seat.Library) > 0; i++ {
		top := seat.Library[0]
		moveCardBetweenZones(gs, perm.Controller, top, "library", "exile", "kylox_attack_exile")
		exiled++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         perm.Controller,
		"sacrificed":   len(victims),
		"total_power":  totalPower,
		"exiled":       exiled,
	})
	if exiled > 0 {
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"free_cast_instants_sorceries_from_exile_unimplemented")
	}
}
