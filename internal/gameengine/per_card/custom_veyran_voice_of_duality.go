package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerVeyranVoiceOfDualityCustom wires the magecraft trigger for
// Veyran, Voice of Duality. The auto-generated stub registerVeyranVoiceOfDuality
// in gen_veyran_voice_of_duality.go remains an inert breadcrumb.
//
// Oracle text (Strixhaven / Commander 2021, {U}{R}):
//
//	Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	Veyran, Voice of Duality gets +1/+1 until end of turn.
//	If you casting or copying an instant or sorcery spell causes a
//	triggered ability of a permanent you control to trigger, that
//	ability triggers an additional time.
//
// Implementation:
//   - "instant_or_sorcery_cast" trigger gated on caster_seat ==
//     controller — pump Veyran +1/+1 UEOT via temp_power / temp_toughness.
//   - The "triggers an additional time" clause is a static replacement
//     that operates on the trigger queue. The engine's per-card hook
//     pipeline is not the right place to express that — the AST /
//     trigger-multiplier framework would need to recognize Veyran the
//     same way it does Strionic Resonator. We emitPartial it.
func registerVeyranVoiceOfDualityCustom(r *Registry) {
	r.OnTrigger("Veyran, Voice of Duality", "instant_or_sorcery_cast", veyranMagecraftPump)
}

func veyranMagecraftPump(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "veyran_voice_of_duality_magecraft_pump"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["temp_power"]++
	perm.Flags["temp_toughness"]++

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"power": perm.Flags["temp_power"],
	})
	// Trigger doubler is a separate static — flag the gap.
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"trigger_multiplier_static_not_implemented_in_per_card_pipeline")
}
