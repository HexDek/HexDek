package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerWinterCynicalOpportunist wires Winter, Cynical Opportunist.
//
// Oracle text:
//
//	Deathtouch
//	Whenever Winter attacks, mill three cards.
//	Delirium — At the beginning of your end step, you may exile any
//	number of cards from your graveyard with four or more card types
//	among them. If you do, put a permanent card from among them onto
//	the battlefield with a finality counter on it.
//
// Implementation:
//   - "creature_attacks" gated on atk == perm: mill the top 3 cards.
//   - "end_step" gated on active_seat == perm.Controller: if our
//     graveyard has cards covering 4+ distinct card types, exile a
//     curated subset that achieves the threshold and reanimate one
//     permanent card with a finality counter.
//   - Deathtouch is engine-handled.
func registerWinterCynicalOpportunist(r *Registry) {
	r.OnTrigger("Winter, Cynical Opportunist", "creature_attacks", winterCynicalAttacks)
	r.OnTrigger("Winter, Cynical Opportunist", "end_step", winterCynicalEndStep)
}

func winterCynicalAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "winter_cynical_attacks_mill"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk != perm {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	milled := 0
	for i := 0; i < 3 && len(seat.Library) > 0; i++ {
		c := seat.Library[0]
		moveCardBetweenZones(gs, perm.Controller, c, "library", "graveyard", "winter_mill")
		milled++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"milled": milled,
	})
}

var winterDeliriumTypes = []string{"creature", "instant", "sorcery", "artifact", "enchantment", "planeswalker", "land", "battle"}

func winterCynicalEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "winter_cynical_delirium_reanimate"
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

	// Count distinct card types present in graveyard.
	present := map[string]bool{}
	for _, c := range seat.Graveyard {
		if c == nil {
			continue
		}
		for _, want := range winterDeliriumTypes {
			if cardHasType(c, want) {
				present[want] = true
			}
		}
	}
	if len(present) < 4 {
		return
	}

	// Find best permanent card to reanimate (highest CMC permanent).
	var pick *gameengine.Card
	bestCMC := -1
	for _, c := range seat.Graveyard {
		if c == nil {
			continue
		}
		isPermanent := cardHasType(c, "creature") || cardHasType(c, "artifact") ||
			cardHasType(c, "enchantment") || cardHasType(c, "planeswalker") ||
			cardHasType(c, "land") || cardHasType(c, "battle")
		if !isPermanent {
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

	// Exile one card per type slot to satisfy "exile any number with 4+
	// types among them" — we just satisfy by exiling at minimum 4 cards.
	exiled := 0
	usedTypes := map[string]bool{}
	for _, c := range append([]*gameengine.Card(nil), seat.Graveyard...) {
		if c == nil || c == pick {
			continue
		}
		var matched string
		for _, t := range winterDeliriumTypes {
			if cardHasType(c, t) && !usedTypes[t] {
				matched = t
				break
			}
		}
		if matched == "" {
			continue
		}
		usedTypes[matched] = true
		moveCardBetweenZones(gs, perm.Controller, c, "graveyard", "exile", "winter_delirium_fuel")
		exiled++
		if len(usedTypes) >= 4 {
			break
		}
	}

	// Bring pick back to battlefield with finality counter.
	moveCardBetweenZones(gs, perm.Controller, pick, "graveyard", "exile", "winter_delirium_pick_route")
	newPerm := enterBattlefieldWithETB(gs, perm.Controller, pick, false)
	if newPerm != nil {
		newPerm.AddCounter("finality", 1)
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"exiled":     exiled,
		"reanimated": pick.DisplayName(),
	})
}
