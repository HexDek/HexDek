package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerIchorid wires Ichorid (Muninn parser-gap #64, ~10.3K hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{3}{B}
//	Creature — Horror
//	Haste
//	At the beginning of the end step, sacrifice this creature.
//	At the beginning of your upkeep, if this card is in your graveyard,
//	you may exile a black creature card other than Ichorid from your
//	graveyard. If you do, return Ichorid to the battlefield.
//
// Implementation:
//   - Haste is AST-side.
//   - end_step: while on the battlefield, sacrifice unconditionally
//     (printed text is "sacrifice this creature" — no opposing rider).
//   - Graveyard-side upkeep return: trigger dispatch in registry.go
//     iterates Seat.Battlefield only — there is no engine hook today
//     for "phase trigger on a card in a private zone". Same gap noted
//     on Quintorius (cards leaving graveyard). emitPartial.
func registerIchorid(r *Registry) {
	r.OnTrigger("Ichorid", "end_step", ichoridEndStepSac)
}

func ichoridEndStepSac(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ichorid_end_step_sac"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	// Printed text: "At the beginning of the end step" — fires every end
	// step, not just controller's. No active-seat gate.
	gameengine.SacrificePermanent(gs, perm, "ichorid_end_step_sac")
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"graveyard_upkeep_exile_black_creature_return_unimplemented")
}
