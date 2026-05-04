package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGrevenPredatorCaptain wires Greven, Predator Captain.
//
// Oracle text:
//
//	Menace
//	Greven gets +X/+0, where X is the amount of life you've lost
//	this turn.
//	Whenever Greven attacks, you may sacrifice another creature.
//	If you do, you draw cards equal to that creature's power and
//	you lose life equal to that creature's toughness.
//
// Implementation:
//   - attacks trigger: pick a sacrificeable creature with the best
//     power-to-toughness ratio (max power, min toughness, never
//     Greven himself, prefer non-commander tokens). If sacrifice
//     would bring controller to <=0 life, skip.
//   - The +X/+0 static buff isn't applied here (engine doesn't
//     support per-card "life lost this turn" pumping at perm scope).
func registerGrevenPredatorCaptain(r *Registry) {
	r.OnTrigger("Greven, Predator Captain", "attacks", grevenAttacks)
}

func grevenAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "greven_attacks_sac_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	attackerSeat, _ := ctx["seat"].(int)
	if attackerSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}

	var pick *gameengine.Permanent
	bestScore := -1
	for _, p := range seat.Battlefield {
		if p == nil || p == perm || p.Card == nil {
			continue
		}
		if !cardHasType(p.Card, "creature") {
			continue
		}
		power := int(p.Card.BasePower)
		tough := int(p.Card.BaseToughness)
		if seat.Life-tough <= 0 {
			continue
		}
		score := power*10 - tough
		if score > bestScore {
			bestScore = score
			pick = p
		}
	}
	if pick == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":     perm.Controller,
			"sacrificed": false,
		})
		return
	}
	power := int(pick.Card.BasePower)
	tough := int(pick.Card.BaseToughness)
	if power < 0 {
		power = 0
	}
	if tough < 0 {
		tough = 0
	}
	removePermanent(gs, pick)
	if pick.Card != nil && !cardHasType(pick.Card, "token") {
		seat.Graveyard = append(seat.Graveyard, pick.Card)
	}
	for i := 0; i < power && len(seat.Library) > 0; i++ {
		card := seat.Library[0]
		seat.Library = seat.Library[1:]
		seat.Hand = append(seat.Hand, card)
	}
	seat.Life -= tough
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"sacrificed": true,
		"power":      power,
		"toughness":  tough,
	})
	emitPartial(gs, "greven_static_pump", perm.Card.DisplayName(),
		"plus_x_zero_from_life_lost_this_turn_unimplemented")
}
