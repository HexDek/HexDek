package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerIsshinCustom records Isshin's "attack triggers fire twice" as a
// game-state flag the engine can check when resolving combat triggers.
//
// Oracle text:
//
//	If a creature attacking causes a triggered ability of a permanent
//	you control to trigger, that ability triggers an additional time.
//
// True double-firing requires engine support that intercepts trigger
// resolution. As an interim, we set `gs.Flags["isshin_active_seat"]` to
// the controlling seat at attack time so the engine and observers can
// branch on it. The handler also emits a parser_gap so the audit can
// track this gap.
func registerIsshinCustom(r *Registry) {
	r.OnTrigger("Isshin, Two Heavens as One", "creature_attacks", isshinOnAttack)
}

func isshinOnAttack(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "isshin_attack_double_marker"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk == nil {
		return
	}
	if atk.Controller != perm.Controller {
		return
	}
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	gs.Flags["isshin_active_seat"] = perm.Controller + 1 // +1 so default 0 means "off"
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"attacker": atk.Card.DisplayName(),
	})
	emitPartial(gs, "isshin_double_attack_triggers", perm.Card.DisplayName(),
		"engine doesn't yet duplicate attack-trigger resolution; flag set for downstream consumers")
}
