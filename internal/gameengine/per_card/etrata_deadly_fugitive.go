package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerEtrataDeadlyFugitive wires Etrata, Deadly Fugitive.
//
// Oracle text:
//
//	Deathtouch
//	Face-down creatures you control have "{2}{U}{B}: Turn this creature
//	face up. If you can't, exile it, then you may cast the exiled card
//	without paying its mana cost."
//	Whenever an Assassin you control deals combat damage to an
//	opponent, cloak the top card of that player's library.
//
// Cloak is a complex face-down mechanic; emitPartial.
func registerEtrataDeadlyFugitive(r *Registry) {
	r.OnTrigger("Etrata, Deadly Fugitive", "combat_damage_player", etrataDeadlyFugitiveCombat)
}

func etrataDeadlyFugitiveCombat(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "etrata_deadly_fugitive_cloak"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	sourceSeat, _ := ctx["source_seat"].(int)
	if sourceSeat != perm.Controller {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"cloak_face_down_mechanic_unimplemented")
}
