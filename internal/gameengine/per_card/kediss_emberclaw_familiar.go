package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKedissEmberclawFamiliar wires Kediss, Emberclaw Familiar.
//
// Oracle text:
//
//	Whenever a commander you control deals combat damage to an opponent,
//	it deals that much damage to each other opponent.
//	Partner
//
// Implementation: listens on combat_damage_player. When the source is a
// commander controlled by Kediss's controller, deal that much damage to
// each *other* opponent of the controller.
func registerKedissEmberclawFamiliar(r *Registry) {
	r.OnTrigger("Kediss, Emberclaw Familiar", "combat_damage_player", kedissCombatDamage)
}

func kedissCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kediss_commander_damage_redirect"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	sourceSeat, _ := ctx["source_seat"].(int)
	if sourceSeat != perm.Controller {
		return
	}
	defenderSeat, _ := ctx["defender_seat"].(int)
	amount, _ := ctx["amount"].(int)
	if amount <= 0 {
		return
	}
	sourceName, _ := ctx["source_card"].(string)
	// Source must be a commander Kediss's controller controls.
	var src *gameengine.Permanent
	for _, p := range gs.Seats[perm.Controller].Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if p.Card.DisplayName() == sourceName && gameengine.IsCommanderCard(gs, perm.Controller, p.Card) {
			src = p
			break
		}
	}
	if src == nil {
		return
	}
	dealt := 0
	for opp := range gs.Seats {
		if opp == perm.Controller || opp == defenderSeat {
			continue
		}
		os := gs.Seats[opp]
		if os == nil || os.Lost {
			continue
		}
		gameengine.DealDamage(gs, opp, amount, src.Card.DisplayName())
		dealt++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          perm.Controller,
		"source":        sourceName,
		"defender_seat": defenderSeat,
		"amount":        amount,
		"opponents_hit": dealt,
	})
	_ = gs.CheckEnd()
}
