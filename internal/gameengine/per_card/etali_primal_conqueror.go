package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerEtaliPrimalConqueror wires Etali, Primal Conqueror // Etali,
// Primal Sickness.
//
// Front face — Etali, Primal Conqueror:
//
//	Trample
//	When Etali enters, each player exiles cards from the top of their
//	library until they exile a nonland card. You may cast any number
//	of spells from among the nonland cards exiled this way without
//	paying their mana costs.
//	{9}{G/P}: Transform Etali. Activate only as a sorcery.
//
// Back face — Etali, Primal Sickness:
//
//	Trample, indestructible
//	Whenever Etali deals combat damage to a player, they get that many
//	poison counters.
//
// Both faces depend on engine machinery (free-cast-from-exile, poison
// counters, transform) that isn't fully wired — emitPartial.
func registerEtaliPrimalConqueror(r *Registry) {
	r.OnETB("Etali, Primal Conqueror", etaliPrimalConquerorETB)
	r.OnETB("Etali, Primal Conqueror // Etali, Primal Sickness", etaliPrimalConquerorETB)
}

func etaliPrimalConquerorETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "etali_primal_conqueror_etb"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"each_player_exile_until_nonland_and_free_cast_unimplemented")
}
