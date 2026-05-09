package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGhyrsonStarnCustom adds the "exact 1 damage → +2 from Ghyrson"
// trigger that the auto-generated static stub omits.
//
// Oracle text:
//
//	Ward {2}
//	Three Autostubs — Whenever another source you control deals exactly
//	1 damage to a permanent or player, Ghyrson Starn deals 2 damage to
//	that permanent or player.
//
// We listen for noncombat damage events; the engine's
// noncombat_damage_to_player and noncombat_damage_to_creature carry the
// amount in `ctx["amount"]`. The trigger fires when the source is
// controlled by Ghyrson's controller, the source is NOT Ghyrson itself
// (otherwise the rider would loop), and the damage amount is exactly 1.
// Combat-damage triggering of this ability is not yet wired.
func registerGhyrsonStarnCustom(r *Registry) {
	r.OnTrigger("Ghyrson Starn, Kelermorph", "noncombat_damage_to_player", ghyrsonOnNoncombatPlayer)
	r.OnTrigger("Ghyrson Starn, Kelermorph", "noncombat_damage_to_creature", ghyrsonOnNoncombatCreature)
}

func ghyrsonOnNoncombatPlayer(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ghyrson_autostubs_player"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	if amt, _ := ctx["amount"].(int); amt != 1 {
		return
	}
	if !ghyrsonSourceQualifies(perm, ctx) {
		return
	}
	target, ok := ctx["target_seat"].(int)
	if !ok || target < 0 || target >= len(gs.Seats) {
		return
	}
	gameengine.DealDamage(gs, target, 2, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":        perm.Controller,
		"target_seat": target,
		"damage":      2,
	})
}

func ghyrsonOnNoncombatCreature(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ghyrson_autostubs_creature"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	if amt, _ := ctx["amount"].(int); amt != 1 {
		return
	}
	if !ghyrsonSourceQualifies(perm, ctx) {
		return
	}
	target, _ := ctx["target_perm"].(*gameengine.Permanent)
	if target == nil {
		return
	}
	target.MarkedDamage += 2
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":        perm.Controller,
		"target_card": target.Card.DisplayName(),
		"damage":      2,
	})
}

func ghyrsonSourceQualifies(ghyrson *gameengine.Permanent, ctx map[string]interface{}) bool {
	src, _ := ctx["source_perm"].(*gameengine.Permanent)
	if src == nil {
		return false
	}
	if src == ghyrson {
		return false
	}
	if src.Controller != ghyrson.Controller {
		return false
	}
	return true
}
