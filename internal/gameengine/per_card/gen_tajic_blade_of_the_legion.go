package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTajicBladeOfTheLegion wires Tajic, Blade of the Legion.
//
// Oracle text:
//
//   Indestructible
//   Battalion — Whenever Tajic and at least two other creatures attack, Tajic gets +5/+5 until end of turn.
//
// Implementation: on Tajic attacking, count total attackers controlled
// by Tajic's controller. If 3+, grant +5/+5 UEOT.
func registerTajicBladeOfTheLegion(r *Registry) {
	r.OnTrigger("Tajic, Blade of the Legion", "attacks", tajicBattalion)
}

func tajicBattalion(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "tajic_battalion"
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
	atkCount := 0
	for _, p := range seat.Battlefield {
		if p == nil || !p.IsCreature() {
			continue
		}
		if p.IsAttacking() {
			atkCount++
		}
	}
	if atkCount < 3 {
		return
	}
	perm.Modifications = append(perm.Modifications, gameengine.Modification{
		Power:     5,
		Toughness: 5,
		Duration:  "until_end_of_turn",
		Timestamp: gs.NextTimestamp(),
	})
	gs.InvalidateCharacteristicsCache()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"attackers": atkCount,
	})
}
