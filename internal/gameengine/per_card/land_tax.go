package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerLandTax wires Land Tax.
//
// Oracle text:
//
//	At the beginning of your upkeep, if an opponent controls more lands
//	than you, you may search your library for up to three basic land
//	cards, reveal them, put them into your hand, then shuffle.
//
// Upkeep trigger, conditional on land-count comparison vs each opponent.
// May ability — Hat opts in (no value reason to decline a free basic tutor
// when the condition is satisfied).
func registerLandTax(r *Registry) {
	r.OnTrigger("Land Tax", "upkeep", landTaxUpkeep)
}

func landTaxUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "land_tax_upkeep"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}

	myLands := countBattlefieldLands(seat.Battlefield)
	opponentHasMore := false
	for i, s := range gs.Seats {
		if i == perm.Controller || s == nil {
			continue
		}
		if countBattlefieldLands(s.Battlefield) > myLands {
			opponentHasMore = true
			break
		}
	}
	if !opponentHasMore {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"my_lands":  myLands,
			"triggered": false,
		})
		return
	}

	found := []string{}
	for _, c := range seat.Library {
		if len(found) >= 3 {
			break
		}
		if c == nil {
			continue
		}
		if cardHasType(c, "basic") && cardHasType(c, "land") {
			gameengine.MoveCard(gs, c, perm.Controller, "library", "hand", "land_tax_search")
			found = append(found, c.DisplayName())
		}
	}
	shuffleLibraryPerCard(gs, perm.Controller)

	gs.LogEvent(gameengine.Event{
		Kind:   "search_library",
		Seat:   perm.Controller,
		Source: perm.Card.DisplayName(),
		Details: map[string]interface{}{
			"found":  found,
			"reason": "land_tax",
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"my_lands":  myLands,
		"triggered": true,
		"found":     found,
	})
}

func countBattlefieldLands(bf []*gameengine.Permanent) int {
	n := 0
	for _, p := range bf {
		if p != nil && p.IsLand() {
			n++
		}
	}
	return n
}
