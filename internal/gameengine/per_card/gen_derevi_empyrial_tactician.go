package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerDereviEmpyrialTactician wires Derevi, Empyrial Tactician.
//
// Oracle text:
//
//   Flying
//   When Derevi enters and whenever a creature you control deals combat damage to a player, you may tap or untap target permanent.
//   {1}{G}{W}{U}: Put Derevi onto the battlefield from the command zone.
//
// Implementation:
//   - Flying: AST keyword.
//   - ETB: pick a target permanent and toggle (tap-or-untap). The "may"
//     is auto-opted-in when there is a sensible choice — tap an
//     opponent's untapped permanent, or untap one of our own tapped
//     creatures with non-trivial board impact.
//   - combat_damage_player trigger: same toggle effect when any creature
//     Derevi's controller controls deals combat damage to a player.
//   - The command-zone alternate cost is a casting-side feature, not a
//     handler effect; the engine's casting pipeline owns it.
func registerDereviEmpyrialTactician(r *Registry) {
	r.OnETB("Derevi, Empyrial Tactician", dereviEmpyrialTacticianETB)
	r.OnTrigger("Derevi, Empyrial Tactician", "combat_damage_player", dereviCombatDamage)
}

func dereviEmpyrialTacticianETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "derevi_empyrial_tactician_etb_tap_untap"
	if gs == nil || perm == nil {
		return
	}
	dereviResolveTapUntap(gs, perm, slug, nil)
}

func dereviCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "derevi_empyrial_tactician_combat_tap_untap"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	sourceSeat, _ := ctx["source_seat"].(int)
	if sourceSeat != perm.Controller {
		return
	}
	dereviResolveTapUntap(gs, perm, slug, ctx)
}

// dereviResolveTapUntap picks a target and toggles its tapped state.
// Targeting priority:
//
//  1. ctx["target_perm"] if supplied.
//  2. An opponent's untapped non-land permanent (tap to neutralize).
//  3. Our own tapped creature (untap for an extra attack/block).
//  4. Our own tapped land (untap for ramp).
func dereviResolveTapUntap(gs *gameengine.GameState, src *gameengine.Permanent, slug string, ctx map[string]interface{}) {
	if ctx != nil {
		if v, ok := ctx["target_perm"].(*gameengine.Permanent); ok && v != nil {
			toggleTapped(v)
			emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
				"seat":   src.Controller,
				"target": v.Card.DisplayName(),
				"now":    tappedString(v),
			})
			return
		}
	}

	// Try opponents' untapped non-land permanents first (prefer creatures).
	var bestOpp *gameengine.Permanent
	bestOppScore := -1
	for i, s := range gs.Seats {
		if s == nil || i == src.Controller {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Tapped {
				continue
			}
			score := 1
			if p.IsCreature() {
				score = 5 + p.Power()
				if p.IsAttacking() {
					score += 50
				}
			} else if p.IsArtifact() || p.IsEnchantment() {
				score = 3
			} else if p.IsLand() {
				continue
			}
			if score > bestOppScore {
				bestOppScore = score
				bestOpp = p
			}
		}
	}
	if bestOpp != nil {
		bestOpp.Tapped = true
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":     src.Controller,
			"target":   bestOpp.Card.DisplayName(),
			"opponent": bestOpp.Controller,
			"now":      "tapped",
		})
		return
	}

	// Otherwise untap one of our own tapped permanents.
	mine := gs.Seats[src.Controller]
	if mine == nil {
		return
	}
	var bestOwn *gameengine.Permanent
	bestOwnScore := -1
	for _, p := range mine.Battlefield {
		if p == nil || !p.Tapped {
			continue
		}
		score := 1
		if p.IsCreature() {
			score = 5 + p.Power()
		} else if p.IsLand() {
			score = 2
		}
		if score > bestOwnScore {
			bestOwnScore = score
			bestOwn = p
		}
	}
	if bestOwn != nil {
		bestOwn.Tapped = false
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":   src.Controller,
			"target": bestOwn.Card.DisplayName(),
			"now":    "untapped",
		})
		return
	}

	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":   src.Controller,
		"target": "none",
	})
}

func toggleTapped(p *gameengine.Permanent) {
	if p == nil {
		return
	}
	p.Tapped = !p.Tapped
}

func tappedString(p *gameengine.Permanent) string {
	if p == nil {
		return ""
	}
	if p.Tapped {
		return "tapped"
	}
	return "untapped"
}
