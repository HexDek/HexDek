package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCecilyHauntedMage wires Cecily, Haunted Mage.
//
// Oracle text:
//
//   Your maximum hand size is eleven.
//   Whenever Cecily, Haunted Mage attacks, you draw a card and you lose 1 life. Then if you have eleven or more cards in your hand, you may cast an instant or sorcery spell from your hand without paying its mana cost.
//   Partner—Friends forever (You can have two commanders if both have this ability.)
//
// Implementation:
//   - "attacks" gated to perm.Controller: draw 1, lose 1 life. The
//     conditional free-cast at 11+ cards in hand is emitPartial.
//   - Hand-size cap and Partner are engine-side state.
func registerCecilyHauntedMage(r *Registry) {
	r.OnTrigger("Cecily, Haunted Mage", "attacks", cecilyHauntedMageAttacks)
}

func cecilyHauntedMageAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "cecily_haunted_mage_attack"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	attackerSeat, _ := ctx["seat"].(int)
	if attackerSeat != perm.Controller {
		return
	}
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	seat := gs.Seats[perm.Controller]
	if seat != nil {
		seat.Life--
		gs.LogEvent(gameengine.Event{
			Kind:   "life_change",
			Seat:   perm.Controller,
			Source: perm.Card.DisplayName(),
			Amount: -1,
			Details: map[string]interface{}{
				"slug":   slug,
				"reason": "cecily_attack_lose_1",
			},
		})
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"free_cast_when_11_plus_cards_in_hand_unimplemented")
}
