package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAishaOfSparksAndSmoke wires Aisha of Sparks and Smoke.
//
// Oracle text:
//
//	Prowess
//	{R/W}: Aisha of Sparks and Smoke gains first strike until end of turn.
//	Whenever Aisha deals combat damage, you may cast a sorcery spell
//	from your hand with mana value less than or equal to that damage
//	without paying its mana cost.
//
// Implementation: prowess is handled by the AST engine; the activated
// first-strike grant and combat-damage free-cast are stubbed via
// emitPartial. The free cast requires a free-cast resolver path that
// the engine doesn't expose for instant/sorcery cards yet.
func registerAishaOfSparksAndSmoke(r *Registry) {
	r.OnTrigger("Aisha of Sparks and Smoke", "combat_damage_player", aishaCombatDamage)
}

func aishaCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "aisha_combat_free_sorcery"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	sourceSeat, _ := ctx["source_seat"].(int)
	if sourceSeat != perm.Controller {
		return
	}
	sourceName, _ := ctx["source_card"].(string)
	if sourceName != "" && sourceName != perm.Card.DisplayName() {
		return
	}
	amount, _ := ctx["amount"].(int)
	if amount <= 0 {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	// Best sorcery in hand <= amount.
	bestIdx := -1
	bestCMC := -1
	for i, c := range seat.Hand {
		if c == nil || !cardHasType(c, "sorcery") {
			continue
		}
		cmc := gameengine.ManaCostOf(c)
		if cmc <= amount && cmc > bestCMC {
			bestCMC = cmc
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"damage": amount,
			"cast":   false,
		})
		return
	}
	cardName := seat.Hand[bestIdx].DisplayName()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"damage":    amount,
		"candidate": cardName,
		"cmc":       bestCMC,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"sorcery_free_cast_resolution_shortcut_unimplemented")
}
