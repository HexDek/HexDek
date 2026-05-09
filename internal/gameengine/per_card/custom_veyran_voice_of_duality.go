package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerVeyranCustom adds Veyran's magecraft pump trigger that the
// auto-generated static stub omits.
//
// Oracle text:
//
//	Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	Veyran gets +1/+1 until end of turn.
//	If you casting or copying an instant or sorcery spell causes a
//	triggered ability of a permanent you control to trigger, that ability
//	triggers an additional time.
//
// The pump is wired here. The "triggered abilities trigger an additional
// time" rider is engine-side and recorded as a parser_gap so the audit
// can find it.
func registerVeyranCustom(r *Registry) {
	r.OnTrigger("Veyran, Voice of Duality", "instant_or_sorcery_cast", veyranMagecraftPump)
}

func veyranMagecraftPump(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "veyran_magecraft_pump"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	caster, ok := ctx["caster_seat"].(int)
	if !ok {
		return
	}
	if caster != perm.Controller {
		return
	}
	perm.Modifications = append(perm.Modifications, gameengine.Modification{
		Power:     1,
		Toughness: 1,
		Duration:  "until_end_of_turn",
		Timestamp: gs.NextTimestamp(),
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"new_power": perm.Power(),
	})
	emitPartial(gs, "veyran_doubles_triggers", perm.Card.DisplayName(),
		"trigger-doubling rider needs engine-side hook into triggered-ability resolution")
}
