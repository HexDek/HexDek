package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerDarkLeoAndShredder wires Dark Leo & Shredder.
//
// Oracle text:
//
//	Sneak {W}{B}
//	Attacking Ninjas you control have deathtouch.
//	Whenever Dark Leo & Shredder deal combat damage to a player, create
//	a 1/1 black Ninja creature token. Then if you control five or more
//	Ninjas, that player loses half their life, rounded up.
func registerDarkLeoAndShredder(r *Registry) {
	r.OnTrigger("Dark Leo & Shredder", "combat_damage_player", darkLeoCombat)
}

func darkLeoCombat(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "dark_leo_shredder_ninja_token"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	sourceSeat, _ := ctx["source_seat"].(int)
	if sourceSeat != perm.Controller {
		return
	}
	sourceName, _ := ctx["source_card"].(string)
	if sourceName != "" && sourceName != perm.Card.DisplayName() {
		return
	}
	defenderSeat, _ := ctx["defender_seat"].(int)
	if defenderSeat < 0 || defenderSeat >= len(gs.Seats) {
		return
	}
	gameengine.CreateCreatureToken(gs, perm.Controller, "Ninja Token",
		[]string{"creature", "ninja", "pip:B"}, 1, 1)
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	ninjas := 0
	for _, p := range seat.Battlefield {
		if p == nil || !p.IsCreature() || p.Card == nil {
			continue
		}
		if cardHasType(p.Card, "ninja") {
			ninjas++
		}
	}
	halfLife := false
	if ninjas >= 5 {
		def := gs.Seats[defenderSeat]
		if def != nil {
			loss := (def.Life + 1) / 2
			if loss > 0 {
				def.Life -= loss
				gs.LogEvent(gameengine.Event{
					Kind:   "life_lost",
					Seat:   defenderSeat,
					Source: perm.Card.DisplayName(),
					Amount: loss,
				})
				halfLife = true
			}
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":        perm.Controller,
		"ninjas":      ninjas,
		"halved_life": halfLife,
	})
}
