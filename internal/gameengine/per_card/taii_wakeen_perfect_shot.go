package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTaiiWakeenPerfectShot wires Taii Wakeen, Perfect Shot.
//
// Oracle text:
//
//	{R}{W}
//	Legendary Creature — Human Mercenary
//	Whenever a source you control deals noncombat damage to a creature
//	  equal to that creature's toughness, draw a card.
//	{X}, {T}: If a source you control would deal noncombat damage to a
//	  permanent or player this turn, it deals that much damage plus X
//	  instead.
//
// Implementation:
//   - "noncombat_damage_to_creature" trigger gated to source controller
//     == Taii's controller and damage_amount == defender_perm's
//     toughness: draw a card.
//   - Activated X+tap damage replacement: emitPartial.
func registerTaiiWakeenPerfectShot(r *Registry) {
	r.OnTrigger("Taii Wakeen, Perfect Shot", "noncombat_damage_to_creature", taiiWakeenOnNoncombatDamage)
	r.OnETB("Taii Wakeen, Perfect Shot", taiiWakeenETB)
}

func taiiWakeenETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "taii_wakeen_etb", perm.Card.DisplayName(),
		"activated_X_tap_noncombat_damage_replacement_plus_X_partial")
}

func taiiWakeenOnNoncombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "taii_wakeen_perfect_shot_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	srcCtrl, _ := ctx["source_controller"].(int)
	if srcCtrl != perm.Controller {
		return
	}
	target, _ := ctx["target_perm"].(*gameengine.Permanent)
	if target == nil || target.Card == nil {
		return
	}
	dmg, _ := ctx["amount"].(int)
	tough := target.Card.BaseToughness
	if target.Counters != nil {
		tough += target.Counters["+1/+1"]
		tough -= target.Counters["-1/-1"]
	}
	if dmg != tough {
		return
	}
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"target":    target.Card.DisplayName(),
		"toughness": tough,
	})
}
