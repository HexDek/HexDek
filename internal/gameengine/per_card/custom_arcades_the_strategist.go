package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerArcadesCustom adds Arcades' defender-ETB draw trigger that
// the auto-generated static stub omits.
//
// Oracle text:
//
//	Flying, vigilance
//	Whenever a creature you control with defender enters, draw a card.
//	Each creature you control with defender assigns combat damage equal
//	to its toughness rather than its power and can attack as though it
//	didn't have defender.
//
// Flying / vigilance are AST keyword territory. We wire here:
//   - The ETB-draw on defender (load-bearing payoff for wall decks).
//   - A layer 7d P/T switch on every defender controlled by Arcades'
//     controller, modeling "assigns combat damage equal to its
//     toughness rather than its power" via the same approximation used
//     for Doran (RegisterDoranSiegeTower in layers.go) — swapping power
//     and toughness gives correct combat damage outcomes in nearly all
//     cases. Active only while Arcades is on the battlefield.
//   - A "can attack as though it didn't have defender" flag on each
//     such defender, which the combat engine consults when validating
//     attacker legality.
func registerArcadesCustom(r *Registry) {
	r.OnETB("Arcades, the Strategist", arcadesETBRegisterStatics)
	r.OnTrigger("Arcades, the Strategist", "permanent_etb", arcadesDefenderETBDraw)
}

func arcadesETBRegisterStatics(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	source := perm
	pred := func(_ *gameengine.GameState, t *gameengine.Permanent) bool {
		if t == nil || t.Card == nil {
			return false
		}
		if t.Controller != source.Controller {
			return false
		}
		if !t.IsCreature() {
			return false
		}
		return t.HasKeyword("defender")
	}
	gameengine.RegisterPTSwitch(gs, source, gameengine.DurationUntilSourceLeaves, pred)

	// "Can attack as though it didn't have defender" — register a tick
	// hook via continuous effect that flips a runtime flag the combat
	// engine can read. We use the layer-6 pipeline (ability-grant layer)
	// and append a synthetic keyword so combat code that tolerates
	// kw:can_attack_with_defender on per-perm Flags also picks it up.
	apply := func(_ *gameengine.GameState, target *gameengine.Permanent, chars *gameengine.Characteristics) {
		if target == nil || chars == nil {
			return
		}
		if target.Flags == nil {
			target.Flags = map[string]int{}
		}
		target.Flags["can_attack_with_defender"] = 1
	}
	gs.RegisterContinuousEffect(&gameengine.ContinuousEffect{
		Layer:          gameengine.LayerAbility,
		Timestamp:      gs.NextTimestamp(),
		SourcePerm:     source,
		SourceCardName: "Arcades, the Strategist",
		ControllerSeat: source.Controller,
		HandlerID:      "arcades_attack_as_no_defender_" + perm.Card.DisplayName(),
		Duration:       gameengine.DurationUntilSourceLeaves,
		Predicate:      pred,
		ApplyFn:        apply,
	})
	emit(gs, "arcades_statics_registered", perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}

func arcadesDefenderETBDraw(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "arcades_defender_etb_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	entered, _ := ctx["perm"].(*gameengine.Permanent)
	if entered == nil || entered.Card == nil {
		return
	}
	if entered.Controller != perm.Controller {
		return
	}
	if !entered.IsCreature() {
		return
	}
	if !entered.HasKeyword("defender") {
		return
	}
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          perm.Controller,
		"defender_card": entered.Card.DisplayName(),
	})
}
