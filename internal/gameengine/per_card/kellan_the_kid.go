package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKellanTheKid wires Kellan, the Kid.
//
// Oracle text:
//
//	Flying, lifelink
//	Whenever you cast a spell from anywhere other than your hand, you may
//	cast a permanent spell with equal or lesser mana value from your hand
//	without paying its mana cost. If you don't, you may put a land card
//	from your hand onto the battlefield.
//
// Implementation: cast-from-non-hand detection requires plumbing the cast
// source zone through every cast path. emitPartial.
func registerKellanTheKid(r *Registry) {
	r.OnTrigger("Kellan, the Kid", "spell_cast", kellanTheKidSpellCast)
}

func kellanTheKidSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kellan_the_kid_cast_from_nonhand"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"cast_from_nonhand_detection_unimplemented")
}
