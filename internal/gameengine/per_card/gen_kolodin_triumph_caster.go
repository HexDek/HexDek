package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKolodinTriumphCaster wires Kolodin, Triumph Caster.
//
// Oracle text:
//
//   Mounts and Vehicles you control have haste.
//   Whenever a Mount you control enters, it becomes saddled until end of turn.
//   Whenever a Vehicle you control enters, it becomes an artifact creature until end of turn.
//
// The static "have haste" requires Layer 6 continuous-ability infrastructure.
// We approximate it by granting haste directly on the entering Mount/Vehicle
// inside each trigger so newly-entered Mounts/Vehicles can attack the same
// turn while Kolodin is on the battlefield.
func registerKolodinTriumphCaster(r *Registry) {
	r.OnTrigger("Kolodin, Triumph Caster", "permanent_etb", kolodinTriumphCasterTrigger1)
	r.OnTrigger("Kolodin, Triumph Caster", "permanent_etb", kolodinTriumphCasterTrigger2)
}

// Trigger 1: a Mount you control entered — it becomes saddled until end
// of turn (and is granted haste as a workaround for the static clause).
func kolodinTriumphCasterTrigger1(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kolodin_triumph_caster_mount_etb"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	controllerSeat, _ := ctx["controller_seat"].(int)
	if controllerSeat != perm.Controller {
		return
	}
	entering, _ := ctx["perm"].(*gameengine.Permanent)
	if entering == nil || entering == perm || entering.Card == nil {
		return
	}
	if !cardHasType(entering.Card, "mount") {
		return
	}
	if entering.Flags == nil {
		entering.Flags = map[string]int{}
	}
	entering.Flags["saddled"] = 1
	if !cardHasKeyword(entering.Card, "haste") {
		entering.GrantedAbilities = append(entering.GrantedAbilities, "haste")
		entering.SummoningSick = false
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"mount": entering.Card.DisplayName(),
	})
}

// Trigger 2: a Vehicle you control entered — it becomes an artifact
// creature until end of turn (mirrors CrewVehicle's grant logic).
func kolodinTriumphCasterTrigger2(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kolodin_triumph_caster_vehicle_etb"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	controllerSeat, _ := ctx["controller_seat"].(int)
	if controllerSeat != perm.Controller {
		return
	}
	entering, _ := ctx["perm"].(*gameengine.Permanent)
	if entering == nil || entering == perm || entering.Card == nil {
		return
	}
	if !gameengine.IsVehicle(entering) {
		return
	}
	if !entering.IsCreature() {
		entering.GrantedAbilities = append(entering.GrantedAbilities, "creature_type_granted")
		hasCreature := false
		for _, t := range entering.Card.Types {
			if t == "creature" {
				hasCreature = true
				break
			}
		}
		if !hasCreature {
			entering.Card.Types = append(entering.Card.Types, "creature")
		}
		if entering.Flags == nil {
			entering.Flags = map[string]int{}
		}
		entering.Flags["crewed_until_eot"] = 1
	}
	if !cardHasKeyword(entering.Card, "haste") {
		entering.GrantedAbilities = append(entering.GrantedAbilities, "haste")
		entering.SummoningSick = false
	}
	gs.InvalidateCharacteristicsCache()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":    perm.Controller,
		"vehicle": entering.Card.DisplayName(),
	})
}
