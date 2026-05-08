package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerPurphorosGodOfTheForge wires Purphoros, God of the Forge.
//
// Oracle text:
//
//	Indestructible
//	As long as your devotion to red is less than five, Purphoros isn't
//	a creature.
//	Whenever another creature you control enters, Purphoros deals 2
//	damage to each opponent.
//	{2}{R}: Creatures you control get +1/+0 until end of turn.
func registerPurphorosGodOfTheForge(r *Registry) {
	r.OnTrigger("Purphoros, God of the Forge", "permanent_etb", purphorosCreatureETB)
}

func purphorosCreatureETB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "purphoros_creature_ping"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	other, _ := ctx["perm"].(*gameengine.Permanent)
	if other == nil || other == perm {
		return
	}
	if other.Controller != perm.Controller {
		return
	}
	if other.Card == nil || !cardHasType(other.Card, "creature") {
		return
	}
	for _, opp := range gs.Opponents(perm.Controller) {
		s := gs.Seats[opp]
		if s == nil || s.Lost {
			continue
		}
		amt, cancelled := gameengine.FireDamageEvent(gs, perm, opp, nil, 2)
		if !cancelled && amt > 0 {
			gameengine.DealDamage(gs, opp, amt, perm.Card.DisplayName())
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"source": other.Card.DisplayName(),
	})
}
