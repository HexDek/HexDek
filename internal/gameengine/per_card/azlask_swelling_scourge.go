package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAzlaskSwellingScourge wires Azlask, the Swelling Scourge.
//
// Oracle text:
//
//	Whenever Azlask or another colorless creature you control dies,
//	you get an experience counter.
//	{W}{U}{B}{R}{G}: Creatures you control get +X/+X until end of turn,
//	where X is the number of experience counters you have. Scions and
//	Spawns you control gain indestructible and annihilator 1 until end
//	of turn.
func registerAzlaskSwellingScourge(r *Registry) {
	r.OnTrigger("Azlask, the Swelling Scourge", "creature_dies", azlaskCreatureDies)
}

func azlaskCreatureDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "azlask_experience_counter"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	deadController, _ := ctx["controller_seat"].(int)
	if deadController != perm.Controller {
		return
	}
	deadCard, _ := ctx["card"].(*gameengine.Card)
	if deadCard != nil && len(deadCard.Colors) > 0 {
		return
	}
	if perm.Controller < 0 || perm.Controller >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if seat.Flags == nil {
		seat.Flags = map[string]int{}
	}
	seat.Flags["experience_counters"]++
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"experience": seat.Flags["experience_counters"],
	})
}
