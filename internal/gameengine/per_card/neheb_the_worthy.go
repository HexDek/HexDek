package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerNehebTheWorthy wires Neheb, the Worthy.
//
// Oracle text:
//
//	First strike
//	Other Minotaurs you control have first strike.
//	As long as you have one or fewer cards in hand, Minotaurs you
//	control get +2/+0.
//	Whenever Neheb deals combat damage to a player, each player
//	discards a card.
//
// First strike + tribal first strike + tribal anthem are static effects
// best handled by the AST/static-effect engine. Combat damage trigger is
// wired here.
func registerNehebTheWorthy(r *Registry) {
	r.OnTrigger("Neheb, the Worthy", "combat_damage_player", nehebCombatDiscard)
}

func nehebCombatDiscard(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "neheb_combat_discard_each"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	sourceSeat, _ := ctx["source_seat"].(int)
	sourceName, _ := ctx["source_card"].(string)
	if sourceSeat != perm.Controller {
		return
	}
	if sourceName != perm.Card.DisplayName() {
		return
	}
	for i, s := range gs.Seats {
		if s == nil || s.Lost {
			continue
		}
		gameengine.DiscardN(gs, i, 1, "neheb_random")
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}
