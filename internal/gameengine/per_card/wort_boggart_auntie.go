package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerWortBoggartAuntie wires Wort, Boggart Auntie.
//
// Oracle text:
//
//	Fear (This creature can't be blocked except by artifact creatures
//	and/or black creatures.)
//	At the beginning of your upkeep, you may return target Goblin card
//	from your graveyard to your hand.
//
// Implementation:
//   - "upkeep_controller" gated on active_seat == perm.Controller:
//     scan graveyard for the highest-CMC Goblin card and return it to
//     hand. AI policy: always opt yes (pure value).
//   - Fear is engine-handled.
func registerWortBoggartAuntie(r *Registry) {
	r.OnTrigger("Wort, Boggart Auntie", "upkeep_controller", wortBoggartAuntieUpkeep)
}

func wortBoggartAuntieUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "wort_boggart_auntie_return_goblin"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	var pick *gameengine.Card
	bestCMC := -1
	for _, c := range seat.Graveyard {
		if c == nil || !cardHasType(c, "goblin") {
			continue
		}
		if cmc := cardCMC(c); cmc > bestCMC {
			bestCMC = cmc
			pick = c
		}
	}
	if pick == nil {
		return
	}
	moveCardBetweenZones(gs, perm.Controller, pick, "graveyard", "hand", "wort_return_goblin")
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"returned": pick.DisplayName(),
	})
}
