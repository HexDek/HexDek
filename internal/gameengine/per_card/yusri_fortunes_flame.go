package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerYusriFortunesFlame wires Yusri, Fortune's Flame.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	{1}{U}{R}
//	Legendary Creature — Efreet
//	2/3
//	Flying
//	Whenever Yusri attacks, choose a number between 1 and 5. Flip that
//	many coins. For each flip you win, draw a card. For each flip you
//	lose, Yusri deals 2 damage to you. If you won five flips this way,
//	you may cast spells from your hand this turn without paying their
//	mana costs.
//
// Implementation:
//   - creature_attacks gated on Yusri himself: pick number = 3 (the EV
//     sweet spot — 50% to draw 2+ vs 6 damage upper bound). Use
//     gs.RNG-equivalent fairness via gs.LogEvent — we model expected
//     value deterministically: 3 flips, 1.5 expected wins, rounded to
//     2 wins / 1 loss. Draw 2, take 2 damage. AI variance on coin
//     flips is poor signal anyway and the simulator favors stable EV.
//   - Free-cast bonus on five-win sweep is unreachable from EV path —
//     emitPartial flags it.
func registerYusriFortunesFlame(r *Registry) {
	r.OnTrigger("Yusri, Fortune's Flame", "creature_attacks", yusriFortunesFlameAttacks)
}

func yusriFortunesFlameAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "yusri_fortunes_flame_coin_flips"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk == nil || atk != perm {
		return
	}
	const wins = 2
	const losses = 1
	for i := 0; i < wins; i++ {
		drawOne(gs, perm.Controller, perm.Card.DisplayName())
	}
	damage := losses * 2
	gameengine.DealDamage(gs, perm.Controller, damage, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"chosen": wins + losses,
		"wins":   wins,
		"losses": losses,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"five_win_free_cast_unreachable_via_deterministic_ev_path")
	_ = gs.CheckEnd()
}
