package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRankleAndTorbran wires Rankle and Torbran (Muninn parser-gap #90,
// ~5.4K hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{B}{R}
//	Legendary Creature — Faerie Dwarf
//	Flying, first strike, haste
//	Whenever Rankle and Torbran deals combat damage to a player or
//	battle, choose any number —
//	• Each player creates a Treasure token.
//	• Each player sacrifices a creature of their choice.
//	• If a source would deal damage to a player or battle this turn,
//	  it deals that much damage plus 2 instead.
//
// Implementation:
//   - Keywords (flying / first strike / haste): AST keyword pipeline.
//   - "combat_damage_player" trigger gated to Rankle and Torbran as the
//     source. AI choice: always pick mode 1 (each player creates a Treasure).
//     The other two modes are skipped:
//       • mode 2 (each player sacrifices) is a symmetric symmetric-sac that
//         tends to net-neutral or hurt us when we have the bigger board, so
//         a hat-aware choice would skip it most of the time anyway;
//       • mode 3 is a continuous damage replacement until end of turn
//         that requires a damage-replacement layer we don't model yet.
//     Modes 2 and 3 flagged emitPartial.
func registerRankleAndTorbran(r *Registry) {
	r.OnTrigger("Rankle and Torbran", "combat_damage_player", rankleAndTorbranCombatDamage)
}

func rankleAndTorbranCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "rankle_torbran_combat_damage_modes"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	sourceName, _ := ctx["source_card"].(string)
	if sourceName != "" && sourceName != perm.Card.DisplayName() {
		return
	}
	sourceSeat, _ := ctx["source_seat"].(int)
	if sourceSeat != perm.Controller {
		return
	}
	amount, _ := ctx["amount"].(int)
	if amount <= 0 {
		return
	}
	treasured := 0
	for i, seat := range gs.Seats {
		if seat == nil || seat.Lost {
			continue
		}
		gameengine.CreateTreasureToken(gs, i)
		treasured++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":             perm.Controller,
		"damage":           amount,
		"mode":             "treasure_each_player",
		"treasures_minted": treasured,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"sacrifice_mode_and_damage_plus_2_mode_unmodeled")
}
