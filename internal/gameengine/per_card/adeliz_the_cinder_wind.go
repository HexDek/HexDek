package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAdelizTheCinderWind wires Adeliz, the Cinder Wind.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	{1}{U}{R}
//	Legendary Creature — Human Wizard
//	Flying, haste
//	Whenever you cast an instant or sorcery spell, Wizards you control
//	get +1/+1 until end of turn.
//
// Implementation:
//   - instant_or_sorcery_cast trigger gated on caster_seat == controller:
//     pump every Wizard the controller controls (+1/+1 until end of
//     turn) via temp_power / temp_toughness flags.
func registerAdelizTheCinderWind(r *Registry) {
	r.OnTrigger("Adeliz, the Cinder Wind", "instant_or_sorcery_cast", adelizCastPump)
}

func adelizCastPump(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "adeliz_cinder_wind_wizard_pump"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	pumped := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil || !p.IsCreature() {
			continue
		}
		if !cardHasType(p.Card, "wizard") {
			continue
		}
		if p.Flags == nil {
			p.Flags = map[string]int{}
		}
		p.Flags["temp_power"]++
		p.Flags["temp_toughness"]++
		pumped++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          perm.Controller,
		"wizards_pumped": pumped,
	})
}
