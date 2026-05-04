package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerBlimComedicGenius wires Blim, Comedic Genius.
//
// Oracle text:
//
//	Flying
//	Whenever Blim deals combat damage to a player, that player gains
//	control of target permanent you control. Then each player loses
//	life and discards cards equal to the number of permanents they
//	control but don't own.
//
// Control-change to opponents is non-trivial — emitPartial.
func registerBlimComedicGenius(r *Registry) {
	r.OnTrigger("Blim, Comedic Genius", "combat_damage_player", blimCombatDamage)
}

func blimCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "blim_comedic_genius_donate"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	sourceSeat, _ := ctx["source_seat"].(int)
	if sourceSeat != perm.Controller {
		return
	}
	sourceName, _ := ctx["source_card"].(string)
	if sourceName != "" && sourceName != perm.Card.DisplayName() {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"opponent_takes_control_and_donate_chain_unimplemented")
}
