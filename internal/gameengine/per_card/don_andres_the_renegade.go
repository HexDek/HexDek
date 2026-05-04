package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerDonAndresTheRenegade wires Don Andres, the Renegade.
//
// Oracle text:
//
//	Each creature you control but don't own gets +2/+2, has menace and
//	deathtouch, and is a Pirate in addition to its other types.
//	Whenever you cast a noncreature spell you don't own, create two
//	tapped Treasure tokens.
//
// Tracking ownership of cast spells across zones is non-trivial. We
// wire the trigger as a stub.
func registerDonAndresTheRenegade(r *Registry) {
	r.OnTrigger("Don Andres, the Renegade", "noncreature_spell_cast", donAndresNoncreature)
}

func donAndresNoncreature(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "don_andres_treasure"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	cardOwner, _ := ctx["card_owner"].(int)
	if cardOwner == perm.Controller {
		// Owned spell — no trigger.
		return
	}
	for i := 0; i < 2; i++ {
		gameengine.CreateTreasureToken(gs, perm.Controller)
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"treasures": 2,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"treasure_tokens_should_enter_tapped")
}
