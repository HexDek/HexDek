package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerChandraFireOfKaladesh wires Chandra, Fire of Kaladesh //
// Chandra, Roaring Flame.
//
// Front face — Chandra, Fire of Kaladesh:
//
//	Whenever you cast a red spell, untap Chandra.
//	{T}: Chandra deals 1 damage to target player or planeswalker. If
//	Chandra has dealt 3 or more damage this turn, exile her, then
//	return her to the battlefield transformed under her owner's
//	control.
//
// Back face — Chandra, Roaring Flame: Planeswalker abilities.
//
// Both faces are largely planeswalker / cumulative-damage logic that the
// engine doesn't yet support; emitPartial.
func registerChandraFireOfKaladesh(r *Registry) {
	r.OnETB("Chandra, Fire of Kaladesh", chandraFireStub)
}

func chandraFireStub(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "chandra_fire_of_kaladesh_partial"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"cumulative_damage_transform_and_planeswalker_face_unimplemented")
}
