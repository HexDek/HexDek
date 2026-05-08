package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAyeshaTanakaArmorer wires Ayesha Tanaka, Armorer.
//
// Oracle text:
//
//	Whenever Ayesha Tanaka attacks, look at the top four cards of your
//	library. You may put any number of artifact cards with mana value
//	less than or equal to Ayesha Tanaka's power from among them onto
//	the battlefield tapped. Put the rest on the bottom of your library
//	in a random order.
//	Ayesha Tanaka can't be blocked as long as defending player controls
//	three or more artifacts.
func registerAyeshaTanakaArmorer(r *Registry) {
	r.OnTrigger("Ayesha Tanaka, Armorer", "attacks", ayeshaAttacks)
}

func ayeshaAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ayesha_tanaka_top4_artifacts"
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
	power := perm.Power()
	look := 4
	if look > len(seat.Library) {
		look = len(seat.Library)
	}
	playedCount := 0
	rest := []*gameengine.Card{}
	for i := 0; i < look; i++ {
		c := seat.Library[0]
		seat.Library = seat.Library[1:]
		if c == nil {
			continue
		}
		if cardHasType(c, "artifact") && gameengine.ManaCostOf(c) <= power {
			gameengine.MoveCard(gs, c, perm.Controller, "library", "battlefield", "ayesha_tanaka")
			playedCount++
		} else {
			rest = append(rest, c)
		}
	}
	seat.Library = append(seat.Library, rest...)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"looked":    look,
		"played":    playedCount,
		"power_cap": power,
	})
}
