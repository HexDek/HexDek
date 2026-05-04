package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerFinneasAceArcher wires Finneas, Ace Archer.
//
// Oracle text:
//
//	Vigilance, reach
//	Whenever Finneas attacks, put a +1/+1 counter on each other
//	creature you control that's a token or a Rabbit. Then if creatures
//	you control have total power 10 or greater, draw a card.
func registerFinneasAceArcher(r *Registry) {
	r.OnTrigger("Finneas, Ace Archer", "attacks", finneasAttacks)
}

func finneasAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "finneas_attack_pump"
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
	count := 0
	for _, p := range seat.Battlefield {
		if p == nil || p == perm || !p.IsCreature() || p.Card == nil {
			continue
		}
		isToken := cardHasType(p.Card, "token")
		isRabbit := cardHasType(p.Card, "rabbit")
		if !isToken && !isRabbit {
			continue
		}
		p.AddCounter("+1/+1", 1)
		count++
	}
	if count > 0 {
		gs.InvalidateCharacteristicsCache()
	}
	totalPower := 0
	for _, p := range seat.Battlefield {
		if p != nil && p.IsCreature() {
			pow := p.Power()
			if pow > 0 {
				totalPower += pow
			}
		}
	}
	drewCard := false
	if totalPower >= 10 {
		drawOne(gs, perm.Controller, perm.Card.DisplayName())
		drewCard = true
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"buffed":    count,
		"power":     totalPower,
		"drew_card": drewCard,
	})
}
