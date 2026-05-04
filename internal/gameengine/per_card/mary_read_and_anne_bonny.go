package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMaryReadAndAnneBonny wires Mary Read and Anne Bonny.
//
// Oracle text:
//
//	Haste
//	{T}: Draw a card, then discard a card.
//	Whenever you discard an Island, Pirate, or Vehicle card, create a
//	tapped Treasure token.
//
// Implementation: card_discarded observer that gates on Mary's controller
// and the discarded card's type. The {T}: rummage activated ability is
// engine-handled if dispatched through OnActivated; we provide it.
func registerMaryReadAndAnneBonny(r *Registry) {
	r.OnTrigger("Mary Read and Anne Bonny", "card_discarded", maryAndAnneDiscard)
	r.OnActivated("Mary Read and Anne Bonny", maryAndAnneActivated)
}

func maryAndAnneIsTriggerType(c *gameengine.Card) bool {
	if c == nil {
		return false
	}
	if cardHasType(c, "pirate") || cardHasType(c, "vehicle") {
		return true
	}
	// Island can be a basic land subtype.
	for _, t := range c.Types {
		if strings.EqualFold(t, "island") {
			return true
		}
	}
	return false
}

func maryAndAnneDiscard(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "mary_anne_discard_treasure"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	discardSeat, _ := ctx["seat"].(int)
	if discardSeat != perm.Controller {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil || !maryAndAnneIsTriggerType(card) {
		return
	}
	gameengine.CreateTreasureToken(gs, perm.Controller)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"discard":  card.DisplayName(),
	})
}

func maryAndAnneActivated(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "mary_anne_rummage"
	if gs == nil || src == nil {
		return
	}
	if src.Tapped {
		emitFail(gs, slug, src.Card.DisplayName(), "already_tapped", nil)
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	src.Tapped = true
	drawOne(gs, src.Controller, src.Card.DisplayName())
	// Discard worst card: prefer a triggering type to chain treasures.
	target := -1
	for i, c := range seat.Hand {
		if c != nil && maryAndAnneIsTriggerType(c) {
			target = i
			break
		}
	}
	if target < 0 && len(seat.Hand) > 0 {
		target = len(seat.Hand) - 1
	}
	if target >= 0 {
		c := seat.Hand[target]
		gameengine.DiscardCard(gs, c, src.Controller)
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat": src.Controller,
	})
}
