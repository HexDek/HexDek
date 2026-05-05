package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAvatarAang wires Avatar Aang // Aang, Master of Elements.
//
// Front face — Avatar Aang:
//
//	Flying, firebending 2
//	Whenever you waterbend, earthbend, firebend, or airbend, draw a
//	card. Then if you've done all four this turn, transform Avatar Aang.
//
// Back face — Aang, Master of Elements:
//
//	Flying
//	Spells you cast cost {W}{U}{B}{R}{G} less to cast.
//	At the beginning of each upkeep, you may transform Aang, Master of
//	Elements. If you do, you gain 4 life, draw four cards, put four
//	+1/+1 counters on him, and he deals 4 damage to each opponent.
//
// Bending mechanics aren't tracked at the engine level — emitPartial.
func registerAvatarAang(r *Registry) {
	r.OnETB("Avatar Aang", avatarAangStub)
	r.OnETB("Avatar Aang // Aang, Master of Elements", avatarAangStub)
}

func avatarAangStub(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "avatar_aang_bending_partial"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"bending_triggers_and_dual_face_transform_not_tracked")
}
