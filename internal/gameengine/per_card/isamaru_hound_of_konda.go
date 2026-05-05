package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerIsamaruHoundOfKonda wires Isamaru, Hound of Konda.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	{W}
//	Legendary Creature — Dog
//	2/2
//	(no abilities)
//
// Vanilla 2/2 for one mana — historically the strongest 1-CMC legendary
// available as a commander. The AST keyword pipeline already handles
// vanilla creatures' combat math; this handler exists only so the
// commander shows up as "registered" in eligibility audits and the AI
// player's 3rd-eye archetype tagger can look up a slug for it.
func registerIsamaruHoundOfKonda(r *Registry) {
	r.OnETB("Isamaru, Hound of Konda", isamaruHoundOfKondaETB)
}

func isamaruHoundOfKondaETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "isamaru_hound_of_konda_vanilla", perm.Card.DisplayName(),
		"vanilla_creature_no_triggered_abilities_register_only_stub")
}
