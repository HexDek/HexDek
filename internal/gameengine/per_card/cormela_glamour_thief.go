package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCormelaGlamourThief wires Cormela, Glamour Thief.
//
// Oracle text:
//
//	Haste
//	{1}, {T}: Add {U}{B}{R}. Spend this mana only to cast instant
//	and/or sorcery spells.
//	When Cormela dies, return up to one target instant or sorcery card
//	from your graveyard to your hand.
func registerCormelaGlamourThief(r *Registry) {
	r.OnTrigger("Cormela, Glamour Thief", "dies", cormelaDies)
}

func cormelaDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "cormela_dies_recursion"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	bestIdx := -1
	bestCMC := -1
	for i, c := range seat.Graveyard {
		if c == nil {
			continue
		}
		if !cardHasType(c, "instant") && !cardHasType(c, "sorcery") {
			continue
		}
		cmc := gameengine.ManaCostOf(c)
		if cmc > bestCMC {
			bestCMC = cmc
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		return
	}
	card := seat.Graveyard[bestIdx]
	gameengine.MoveCard(gs, card, perm.Controller, "graveyard", "hand", "cormela_dies")
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"returned": card.DisplayName(),
	})
}
