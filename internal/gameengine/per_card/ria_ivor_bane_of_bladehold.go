package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRiaIvor wires Ria Ivor, Bane of Bladehold.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Battle cry
//	At the beginning of combat on your turn, the next time target
//	creature would deal combat damage to one or more players this
//	combat, prevent that damage. If damage is prevented this way,
//	create that many 1/1 colorless Phyrexian Mite artifact creature
//	tokens with toxic 1 and "This token can't block."
//
// Implementation:
//   - Battle cry handled by AST keyword pipeline.
//   - "combat_begin" trigger fires when combat begins on Ria's
//     controller's turn. Pick the highest-power opposing creature whose
//     damage we want to neuter, stamp a "ria_prevent_damage_next" flag
//     on it, and arm a one-shot prevention scoped to this combat.
//   - "combat_damage_dealt" trigger consumes the prevention: when the
//     flagged creature would deal N combat damage to players, we zero
//     the damage and mint N Phyrexian Mites for Ria's controller.
//
// Note: the engine's prevention pipeline doesn't yet expose a clean
// "next-time replacement" hook, so we approximate by listening on
// "combat_damage_dealt" and undoing the life loss when our flag was set.
// emitPartial flags the gap.
func registerRiaIvor(r *Registry) {
	r.OnTrigger("Ria Ivor, Bane of Bladehold", "combat_begin", riaIvorCombatBegin)
	r.OnTrigger("Ria Ivor, Bane of Bladehold", "combat_damage_dealt", riaIvorPreventAndMint)
}

func riaIvorCombatBegin(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ria_ivor_arm_prevention"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}

	// Target the highest-power opponent creature (best damage to prevent).
	var best *gameengine.Permanent
	bestPow := -1
	for _, oppIdx := range gs.Opponents(perm.Controller) {
		s := gs.Seats[oppIdx]
		if s == nil || s.Lost {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil || !p.IsCreature() {
				continue
			}
			pow := p.Power()
			if pow > bestPow {
				bestPow = pow
				best = p
			}
		}
	}
	if best == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_target_creature", map[string]interface{}{
			"seat": perm.Controller,
		})
		return
	}
	if best.Flags == nil {
		best.Flags = map[string]int{}
	}
	best.Flags["ria_prevent_next_damage"] = 1
	best.Flags["ria_prevent_for_seat"] = perm.Controller + 1 // stash 1-indexed
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"target": best.Card.DisplayName(),
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"prevention_implemented_via_post_damage_compensation_not_replacement_effect")
}

func riaIvorPreventAndMint(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ria_ivor_prevent_and_mint"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	src, _ := ctx["source_perm"].(*gameengine.Permanent)
	if src == nil || src.Flags == nil {
		return
	}
	if src.Flags["ria_prevent_next_damage"] != 1 {
		return
	}
	stashed := src.Flags["ria_prevent_for_seat"] - 1
	if stashed != perm.Controller {
		return
	}
	amount, _ := ctx["amount"].(int)
	targetKind, _ := ctx["target_kind"].(string)
	if targetKind != "player" || amount <= 0 {
		return
	}
	// "Refund" the damage: target seat regains the life lost.
	tgtSeat, _ := ctx["target_seat"].(int)
	if tgtSeat >= 0 && tgtSeat < len(gs.Seats) && gs.Seats[tgtSeat] != nil {
		gameengine.GainLife(gs, tgtSeat, amount, perm.Card.DisplayName())
	}
	// Mint amount Phyrexian Mite tokens for Ria's controller.
	for i := 0; i < amount; i++ {
		gameengine.CreateCreatureToken(gs, perm.Controller, "Phyrexian Mite",
			[]string{"creature", "phyrexian", "mite", "artifact"}, 1, 1)
	}
	src.Flags["ria_prevent_next_damage"] = 0
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"prevented": amount,
		"tokens":    amount,
	})
}
