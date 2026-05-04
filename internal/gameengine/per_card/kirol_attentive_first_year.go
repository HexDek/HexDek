package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKirolAttentiveFirstYear wires Kirol, Attentive First-Year.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Tap two untapped creatures you control: Copy target triggered
//	  ability you control. You may choose new targets for the copy.
//	  Activate only once each turn.
//
// Copying triggered abilities requires stack-aware machinery that the
// per-card pipeline doesn't expose. Register an ETB partial flag and
// gate the activation on a turn-keyed flag.
func registerKirolAttentiveFirstYear(r *Registry) {
	r.OnETB("Kirol, Attentive First-Year", kirolETB)
	r.OnActivated("Kirol, Attentive First-Year", kirolActivate)
}

func kirolETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "kirol_static", perm.Card.DisplayName(),
		"trigger_copy_activation_not_modeled")
}

func kirolActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "kirol_copy_trigger"
	if gs == nil || src == nil {
		return
	}
	if src.Flags == nil {
		src.Flags = map[string]int{}
	}
	if src.Flags["kirol_used_turn"] == gs.Turn {
		emitFail(gs, slug, src.Card.DisplayName(), "already_activated_this_turn", nil)
		return
	}
	src.Flags["kirol_used_turn"] = gs.Turn
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat": src.Controller,
	})
	emitPartial(gs, slug, src.Card.DisplayName(),
		"trigger_copy_resolution_not_modeled")
}
