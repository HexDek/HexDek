package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMayaelTheAnima wires Mayael the Anima.
//
// Oracle text:
//
//   {3}{R}{G}{W}, {T}: Look at the top five cards of your library. You may put a creature card with power 5 or greater from among them onto the battlefield. Put the rest on the bottom of your library in any order.
//
// Implementation: scan top 5 of controller's library, drop the highest-
// power creature (power ≥ 5) onto the battlefield, send the rest to the
// bottom of the library.
func registerMayaelTheAnima(r *Registry) {
	r.OnActivated("Mayael the Anima", mayaelTheAnimaActivate)
}

func mayaelTheAnimaActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "mayael_top5_creature_5plus"
	if gs == nil || src == nil {
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	look := 5
	if look > len(seat.Library) {
		look = len(seat.Library)
	}
	if look == 0 {
		return
	}
	top := make([]*gameengine.Card, 0, look)
	for i := 0; i < look; i++ {
		top = append(top, seat.Library[0])
		seat.Library = seat.Library[1:]
	}
	bestIdx := -1
	bestPow := -1
	for i, c := range top {
		if c == nil || !cardHasType(c, "creature") {
			continue
		}
		p := int(c.BasePower)
		if p < 5 {
			continue
		}
		if p > bestPow {
			bestPow = p
			bestIdx = i
		}
	}
	if bestIdx >= 0 {
		c := top[bestIdx]
		gameengine.MoveCard(gs, c, src.Controller, "library", "battlefield", "mayael")
		enterBattlefieldWithETB(gs, src.Controller, c, false)
		top = append(top[:bestIdx], top[bestIdx+1:]...)
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":   src.Controller,
			"played": c.DisplayName(),
			"power":  bestPow,
		})
	} else {
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":   src.Controller,
			"played": nil,
		})
	}
	seat.Library = append(seat.Library, top...)
}
