package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKolodinTriumphCaster wires Kolodin, Triumph Caster.
//
// Oracle text:
//
//	Mounts and Vehicles you control have haste.
//	Whenever a Mount you control enters, it becomes saddled until end of turn.
//	Whenever a Vehicle you control enters, it becomes an artifact creature
//	until end of turn.
//
// Implementation:
//   - Single permanent_etb trigger that classifies the entering permanent
//     and applies the right one-shot. Unified into one fn so we don't
//     fire two parser_gap breadcrumbs per ETB.
//   - Mount → set "saddled" flag through end of turn. Saddle-checking
//     code (e.g. The Gitrog Ride) keys off this flag.
//   - Vehicle → set "kw:artifact_creature" flag and grant creature type
//     until end of turn (engine-side type-grant pipeline reads this).
//   - Static "haste on Mounts and Vehicles you control" handled by the
//     AST keyword pipeline; emitPartial flags the boundary.
func registerKolodinTriumphCaster(r *Registry) {
	r.OnETB("Kolodin, Triumph Caster", kolodinTriumphCasterETB)
	r.OnTrigger("Kolodin, Triumph Caster", "permanent_etb", kolodinTriumphCasterETBTrigger)
}

func kolodinTriumphCasterETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "kolodin_triumph_caster_etb"
	if gs == nil || perm == nil {
		return
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"mount_vehicle_haste_static_handled_by_ast_keyword_pipeline")
}

func kolodinTriumphCasterETBTrigger(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kolodin_triumph_caster_mount_vehicle_etb"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	controllerSeat, _ := ctx["controller_seat"].(int)
	if controllerSeat != perm.Controller {
		return
	}
	entered, _ := ctx["perm"].(*gameengine.Permanent)
	if entered == nil || entered == perm || entered.Card == nil {
		return
	}
	if entered.Controller != perm.Controller {
		return
	}
	if entered.Flags == nil {
		entered.Flags = map[string]int{}
	}
	if cardSubtypeMatches(entered.Card, "mount") {
		entered.Flags["saddled"] = 1
		// Track lifetime; cleared at end of turn by the cleanup pass.
		entered.Flags["saddled_until_eot"] = 1
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":     perm.Controller,
			"target":   entered.Card.DisplayName(),
			"effect":   "saddled_until_eot",
		})
		return
	}
	if cardSubtypeMatches(entered.Card, "vehicle") {
		entered.Flags["kw:artifact_creature_until_eot"] = 1
		// Add a "creature" type tag for the duration so combat code sees
		// the vehicle as a creature without needing crew.
		entered.Card.Types = append(entered.Card.Types, "creature_until_eot")
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":     perm.Controller,
			"target":   entered.Card.DisplayName(),
			"effect":   "becomes_artifact_creature_until_eot",
		})
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"vehicle_creature_type_grant_eot_cleanup_relies_on_engine_until_eot_pass")
		return
	}
}
