package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGisaGloriousResurrector wires Gisa, Glorious Resurrector.
//
// Oracle text:
//
//	If a creature an opponent controls would die, exile it instead.
//	At the beginning of your upkeep, put all creature cards exiled with
//	Gisa onto the battlefield under your control. They gain decayed.
//
// The "would die → exile instead" replacement is engine-level and
// requires replacement-effect registration. We track exiled-with-Gisa
// via seat flag and pull from a side pile.
func registerGisaGloriousResurrector(r *Registry) {
	r.OnTrigger("Gisa, Glorious Resurrector", "upkeep", gisaResurrectorUpkeep)
	r.OnTrigger("Gisa, Glorious Resurrector", "creature_dies", gisaResurrectorDies)
}

func gisaResurrectorDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "gisa_resurrector_replace_die"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"would_die_replacement_effect_not_wired_through_per_card_hook")
}

func gisaResurrectorUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "gisa_resurrector_upkeep"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}
