package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerBloodthirstyConqueror wires Bloodthirsty Conqueror (Muninn
// parser-gap rank ~143, drain-mirror finisher).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{3}{B}{B}
//	Creature — Vampire Knight
//	Flying, deathtouch
//	Whenever an opponent loses life, you gain that much life. (Damage
//	causes loss of life.)
//
// Implementation:
//   - Flying + deathtouch via AST keyword pipeline.
//   - OnTrigger("life_lost"): gated on the losing seat being an opponent
//     of perm.Controller. Mirror the amount lost into GainLife on
//     perm.Controller. Mirrors Vito/Sanguine Bond shape inverted —
//     instead of gain → opp loss, this is opp loss → our gain.
//   - Bloodthirsty Conqueror also pairs with Sanguine Bond / Exquisite
//     Blood for the classic infinite drain combo; if a Sanguine
//     Bond-equivalent is on our side the engine's regular trigger
//     depth cap will halt the chain rather than infinite-loop.
func registerBloodthirstyConqueror(r *Registry) {
	r.OnTrigger("Bloodthirsty Conqueror", "life_lost", bloodthirstyConquerorLifeLost)
}

func bloodthirstyConquerorLifeLost(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "bloodthirsty_conqueror_drain_mirror"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	loserSeat, _ := ctx["seat"].(int)
	if loserSeat == perm.Controller {
		return
	}
	if loserSeat < 0 || loserSeat >= len(gs.Seats) {
		return
	}
	amount, _ := ctx["amount"].(int)
	if amount <= 0 {
		return
	}
	gameengine.GainLife(gs, perm.Controller, amount, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"loser_seat": loserSeat,
		"gained":     amount,
	})
}
