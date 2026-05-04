package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMagnusTheRed wires Magnus the Red.
//
// Oracle text:
//
//   Flying
//   Unearthly Power — Instant and sorcery spells you cast cost {1} less to cast for each creature token you control.
//   Blade of Magnus — Whenever Magnus the Red deals combat damage to a player, create a 3/3 red Spawn creature token.
//
// Implementation:
//   - Cost reduction (Unearthly Power) handled by AST cost-mod pipeline.
//   - "combat_damage_player" gated to Magnus as source: create a 3/3 red
//     Spawn token.
func registerMagnusTheRed(r *Registry) {
	r.OnTrigger("Magnus the Red", "combat_damage_player", magnusTheRedSpawn)
}

func magnusTheRedSpawn(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "magnus_the_red_spawn"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	srcPerm, _ := ctx["source_perm"].(*gameengine.Permanent)
	if srcPerm != perm {
		// Fall back to source_card name match if perm pointer isn't set.
		srcName, _ := ctx["source_card"].(string)
		if srcName != perm.Card.DisplayName() {
			return
		}
	}
	token := &gameengine.Card{
		Name:          "Spawn Token",
		Owner:         perm.Controller,
		BasePower:     3,
		BaseToughness: 3,
		Types:         []string{"token", "creature", "spawn"},
		Colors:        []string{"R"},
		TypeLine:      "Token Creature — Spawn",
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}
