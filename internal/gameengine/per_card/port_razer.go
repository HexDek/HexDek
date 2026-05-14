package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerPortRazer wires Port Razer.
//
// Oracle text (Scryfall, verified 2026-05-14):
//
//	{3}{R}{R}
//	Creature — Orc Pirate
//	Whenever this creature deals combat damage to a player, untap each
//	creature you control. After this phase, there is an additional
//	combat phase.
//	This creature can't attack a player it has already attacked this
//	turn.
//
// Engine handling:
//
//   - Combat damage trigger: this handler.
//     1. Untap all creatures controlled by Port Razer's controller —
//        modeled here as an OnBegin hook on the extra combat (the
//        oracle wording reads "untap each creature you control" as
//        part of the trigger effect that happens BEFORE the new
//        combat phase, but functionally there's no observable
//        difference between firing the untap immediately on the
//        trigger vs at the very start of the new combat; we choose
//        the OnBegin form to demonstrate the hook and to make the
//        untap visible in the per-combat event stream).
//     2. Queue an extra combat phase (vanilla — no attacker
//        restriction; Port Razer's extra combats let everyone attack).
//   - "Can't attack a player it has already attacked this turn":
//     this is an attack restriction on Port Razer specifically.
//     Handled at the engine's DeclareAttackers logic via the per-turn
//     "attacked_player" flag tracking (gs.Seats[seat].Flags or a
//     per-permanent flag). Not implemented in this handler — that
//     restriction is more general than Port Razer (many cards have
//     "can't attack same player twice" clauses) and should live in
//     the combat layer, not in a per_card hook. Logged here for
//     visibility into a future combat-layer pass.
//
// This handler uses the typed extra-combat queue introduced in the
// extra-combat-restrictions refactor — Port Razer's vanilla extra
// combat with OnBegin untap-each-creature hook is the canonical
// example of "no restriction, but has a beginning-of-combat rider."
func registerPortRazer(r *Registry) {
	r.OnTrigger("Port Razer", "combat_damage_player", portRazerCombatDamage)
}

func portRazerCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "port_razer_extra_combat"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	// Gate to Port Razer's own combat damage. Same context-key pattern
	// as other combat-damage triggers (source_card name + source_seat).
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
	controller := perm.Controller
	if controller < 0 || controller >= len(gs.Seats) {
		return
	}

	owner := controller
	gs.AddExtraCombat(gameengine.PendingExtraCombat{
		SourceCard: perm.Card.DisplayName(),
		// OnBegin: untap each creature you control before the new
		// combat phase declares attackers. Captured controller via
		// closure so the rider follows the original Port Razer's
		// controller even if Port Razer changes hands.
		OnBegin: func(g *gameengine.GameState) {
			if g == nil || owner < 0 || owner >= len(g.Seats) {
				return
			}
			s := g.Seats[owner]
			if s == nil {
				return
			}
			for _, p := range s.Battlefield {
				if p == nil || p.Card == nil {
					continue
				}
				if !p.IsCreature() {
					continue
				}
				if p.Tapped {
					p.Tapped = false
				}
			}
		},
	})

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          controller,
		"damage":        amount,
		"extra_combats": len(gs.PendingExtraCombats),
	})
}
