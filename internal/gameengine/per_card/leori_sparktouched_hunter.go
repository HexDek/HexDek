package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerLeoriSparktouchedHunter wires Leori, Sparktouched Hunter.
//
// Oracle text:
//
//	Flying, vigilance
//	Whenever Leori deals combat damage to a player, choose a planeswalker
//	type. Until end of turn, whenever you activate an ability of a
//	planeswalker of that type, copy that ability. You may choose new
//	targets for the copies.
//
// Implementation: ability-copy delayed trigger requires plumbing
// activated-ability copies on the stack. emitPartial.
func registerLeoriSparktouchedHunter(r *Registry) {
	r.OnTrigger("Leori, Sparktouched Hunter", "combat_damage_player", leoriCombatDamage)
}

func leoriCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "leori_planeswalker_copy"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	sourceSeat, _ := ctx["source_seat"].(int)
	if sourceSeat != perm.Controller {
		return
	}
	sourceName, _ := ctx["source_card"].(string)
	if sourceName != perm.Card.DisplayName() {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"planeswalker_ability_copy_delayed_trigger_unimplemented")
}
