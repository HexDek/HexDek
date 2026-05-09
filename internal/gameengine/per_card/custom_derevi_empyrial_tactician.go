package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerDereviEmpyrialTacticianCustom implements Derevi's ETB tap/untap
// trigger and adds the parser_gap for the combat-damage trigger and
// command-zone alt-cost.
//
// Oracle text:
//
//	Flying
//	When Derevi enters and whenever a creature you control deals combat
//	damage to a player, you may tap or untap target permanent.
//	{1}{G}{W}{U}: Put Derevi onto the battlefield from the command zone.
//
// The ETB branch fires here. The command-zone alt-cost is a casting
// path the engine doesn't yet expose; partial. The combat-damage trigger
// is wired as a separate trigger handler.
//
// Heuristic: tap an opponent's largest untapped creature if available;
// otherwise untap our own largest tapped creature.
func registerDereviEmpyrialTacticianCustom(r *Registry) {
	r.OnETB("Derevi, Empyrial Tactician", dereviETBTapOrUntap)
	r.OnTrigger("Derevi, Empyrial Tactician", "combat_damage_to_player", dereviCombatDamageTapOrUntap)
}

func dereviETBTapOrUntap(gs *gameengine.GameState, perm *gameengine.Permanent) {
	dereviPickAndToggle(gs, perm, "etb")
}

func dereviCombatDamageTapOrUntap(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk == nil || atk.Controller != perm.Controller {
		return
	}
	dereviPickAndToggle(gs, perm, "combat")
}

func dereviPickAndToggle(gs *gameengine.GameState, perm *gameengine.Permanent, source string) {
	const slug = "derevi_tap_or_untap"
	if gs == nil || perm == nil {
		return
	}
	// Prefer tapping the largest untapped opponent creature.
	var tapTarget *gameengine.Permanent
	tapTargetSeat := -1
	tapPower := -1
	for i, s := range gs.Seats {
		if s == nil || i == perm.Controller {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Tapped || !p.IsCreature() {
				continue
			}
			if p.Power() > tapPower {
				tapTarget = p
				tapTargetSeat = i
				tapPower = p.Power()
			}
		}
	}
	if tapTarget != nil {
		tapTarget.Tapped = true
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":        perm.Controller,
			"source":      source,
			"action":      "tap",
			"target_seat": tapTargetSeat,
			"target":      tapTarget.Card.DisplayName(),
		})
		return
	}
	// Otherwise untap our biggest tapped creature.
	var untapTarget *gameengine.Permanent
	untapPower := -1
	if seat := gs.Seats[perm.Controller]; seat != nil {
		for _, p := range seat.Battlefield {
			if p == nil || !p.Tapped || !p.IsCreature() {
				continue
			}
			if p.Power() > untapPower {
				untapTarget = p
				untapPower = p.Power()
			}
		}
	}
	if untapTarget != nil {
		untapTarget.Tapped = false
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"source": source,
			"action": "untap",
			"target": untapTarget.Card.DisplayName(),
		})
		return
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"source": source,
		"action": "none",
		"note":   "no_legal_target",
	})
}
