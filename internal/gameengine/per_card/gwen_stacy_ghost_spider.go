package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGwenStacyGhostSpider wires Ghost-Spider, Gwen Stacy.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Menace
//	Whenever Ghost-Spider attacks, she deals X damage to defending
//	  player, where X is the number of attacking creatures.
//
// Implementation:
//   - "creature_attacks": gate on attacker_perm == perm. Find defender
//     via AttackerDefender, count attacking creatures controlled by us,
//     deal X damage to defender.
//   - Menace handled by AST keyword pipeline.
func registerGwenStacyGhostSpider(r *Registry) {
	r.OnTrigger("Ghost-Spider, Gwen Stacy", "creature_attacks", gwenStacyAttack)
	r.OnTrigger("Gwen Stacy // Ghost-Spider", "creature_attacks", gwenStacyAttack)
	r.OnTrigger("Gwen Stacy", "creature_attacks", gwenStacyAttack)
}

func gwenStacyAttack(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "gwen_stacy_attack_damage"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk != perm {
		return
	}
	defenderSeat, _ := gameengine.AttackerDefender(perm)
	if defenderSeat < 0 || defenderSeat >= len(gs.Seats) {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_defender", nil)
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	x := 0
	for _, p := range seat.Battlefield {
		if p == nil || !p.IsCreature() {
			continue
		}
		if p.IsAttacking() {
			x++
		}
	}
	if x <= 0 {
		return
	}
	gameengine.DealDamage(gs, defenderSeat, x, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":           perm.Controller,
		"target_seat":    defenderSeat,
		"attacker_count": x,
		"damage":         x,
	})
	_ = gs.CheckEnd()
}
