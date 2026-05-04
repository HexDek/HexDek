package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCalixGuidedByFate wires Calix, Guided by Fate.
//
// Oracle text:
//
//	Constellation — Whenever Calix or another enchantment you control
//	enters, put a +1/+1 counter on target creature.
//	Whenever Calix or an enchanted creature you control deals combat
//	damage to a player, you may create a token that's a copy of a
//	nonlegendary enchantment you control. Do this only once each turn.
//
// Constellation: pick best other creature to bulk up. Combat copy is
// non-trivial — emitPartial.
func registerCalixGuidedByFate(r *Registry) {
	r.OnETB("Calix, Guided by Fate", calixSelfETB)
	r.OnTrigger("Calix, Guided by Fate", "permanent_etb", calixConstellation)
	r.OnTrigger("Calix, Guided by Fate", "combat_damage_player", calixCombatCopy)
}

func calixSelfETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	calixAddCounter(gs, perm)
}

func calixConstellation(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if ctx == nil {
		return
	}
	controllerSeat, _ := ctx["controller_seat"].(int)
	if controllerSeat != perm.Controller {
		return
	}
	entered, _ := ctx["perm"].(*gameengine.Permanent)
	if entered == nil || entered == perm || entered.Card == nil {
		return
	}
	if !cardHasType(entered.Card, "enchantment") {
		return
	}
	calixAddCounter(gs, perm)
}

func calixAddCounter(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "calix_constellation"
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	var best *gameengine.Permanent
	bestPow := -1
	for _, p := range seat.Battlefield {
		if p == nil || !p.IsCreature() {
			continue
		}
		if pow := p.Power(); pow > bestPow {
			best = p
			bestPow = pow
		}
	}
	if best == nil {
		return
	}
	best.AddCounter("+1/+1", 1)
	gs.InvalidateCharacteristicsCache()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"target": best.Card.DisplayName(),
	})
}

func calixCombatCopy(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "calix_combat_enchantment_copy"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"combat_enchantment_token_copy_unimplemented")
}
