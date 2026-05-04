package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerScionOfTheUrDragon wires Scion of the Ur-Dragon.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Flying
//	{2}: Search your library for a Dragon permanent card and put it
//	  into your graveyard. If you do, Scion of the Ur-Dragon becomes a
//	  copy of that card until end of turn. Then shuffle.
//
// Implementation:
//   - Activated tutor + UEOT copy effect requires the layers pipeline.
//     Register as ETB partial-flag marker. The activation can perform
//     the tutor portion (move highest-power Dragon from library to
//     graveyard) but the copy effect itself is handled at layers.
func registerScionOfTheUrDragon(r *Registry) {
	r.OnETB("Scion of the Ur-Dragon", scionOfUrDragonETB)
	r.OnActivated("Scion of the Ur-Dragon", scionOfUrDragonActivate)
}

func scionOfUrDragonETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "scion_ur_dragon_static", perm.Card.DisplayName(),
		"copy_target_dragon_ueot_continuous_static_not_modeled")
}

func scionOfUrDragonActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "scion_ur_dragon_dragon_tutor"
	if gs == nil || src == nil {
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	var pick *gameengine.Card
	bestPow := -1
	for _, c := range seat.Library {
		if c == nil {
			continue
		}
		if !cardHasType(c, "dragon") {
			continue
		}
		// "permanent card": creature, artifact, enchantment, land, planeswalker, battle.
		if !cardHasType(c, "creature") && !cardHasType(c, "artifact") &&
			!cardHasType(c, "enchantment") && !cardHasType(c, "land") &&
			!cardHasType(c, "planeswalker") {
			continue
		}
		if c.BasePower > bestPow {
			bestPow = c.BasePower
			pick = c
		}
	}
	if pick == nil {
		emitFail(gs, slug, src.Card.DisplayName(), "no_dragon_in_library", nil)
		return
	}
	gameengine.MoveCard(gs, pick, src.Controller, "library", "graveyard", "scion_tutor")
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":   src.Controller,
		"dragon": pick.DisplayName(),
	})
	emitPartial(gs, slug, src.Card.DisplayName(),
		"becomes_a_copy_ueot_not_modeled")
}
