package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTeferiTemporalArchmage wires Teferi, Temporal Archmage.
//
// Oracle text:
//
//	{4}{U}{U}
//	Legendary Planeswalker — Teferi
//	+1: Look at the top two cards of your library. Put one of them into
//	    your hand and the other on the bottom of your library.
//	-1: Untap up to four target permanents.
//	-10: You get an emblem with "You may activate loyalty abilities of
//	     planeswalkers you control on any player's turn any time you
//	     could cast an instant."
//	Teferi, Temporal Archmage can be your commander.
//
// Implementation:
//   - Loyalty abilities are not yet routed through per_card; the engine's
//     planeswalker activation pipeline handles loyalty bookkeeping
//     generically. We register an ETB stub that emits a parser_gap note
//     so analysis tooling sees the snowflake; the +1 (impulse top-2) and
//     -1 (untap up to 4) effects are common enough that the engine's
//     activated-ability pipeline could resolve them via AST in the future.
func registerTeferiTemporalArchmage(r *Registry) {
	r.OnETB("Teferi, Temporal Archmage", teferiTemporalArchmageETB)
}

func teferiTemporalArchmageETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "teferi_temporal_archmage_etb", perm.Card.DisplayName(),
		"loyalty_abilities_plus1_impulse_minus1_untap_minus10_emblem_partial")
}
