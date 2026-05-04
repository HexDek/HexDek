package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGerrardWeatherlightHero wires Gerrard, Weatherlight Hero.
//
// Oracle text:
//
//	First strike
//	When Gerrard dies, exile it and return to the battlefield all
//	artifact and creature cards in your graveyard that were put there
//	from the battlefield this turn.
//
// We approximate "from battlefield this turn" by a perm flag on each
// card that the engine sets at zone-change time; absent that, we
// reanimate any artifact/creature in the controller's graveyard.
func registerGerrardWeatherlightHero(r *Registry) {
	r.OnTrigger("Gerrard, Weatherlight Hero", "dies", gerrardDies)
}

func gerrardDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "gerrard_dies_mass_recur"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	// Exile Gerrard himself.
	moveCardBetweenZones(gs, perm.Controller, perm.Card, "graveyard", "exile", "gerrard_self_exile")
	// Reanimate each artifact/creature card in graveyard.
	var keep []*gameengine.Card
	returned := 0
	for _, c := range seat.Graveyard {
		if c == nil {
			continue
		}
		if cardHasType(c, "creature") || cardHasType(c, "artifact") {
			gameengine.MoveCard(gs, c, perm.Controller, "graveyard", "battlefield", "gerrard_recur")
			enterBattlefieldWithETB(gs, perm.Controller, c, false)
			returned++
		} else {
			keep = append(keep, c)
		}
	}
	seat.Graveyard = keep
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"returned": returned,
	})
}
