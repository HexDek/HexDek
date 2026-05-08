package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerZoralineCosmosCaller wires Zoraline, Cosmos Caller.
//
// Oracle text:
//
//	Flying, vigilance
//	Whenever a Bat you control attacks, you gain 1 life.
//	Whenever Zoraline enters or attacks, you may pay {W}{B} and 2 life.
//	When you do, return target nonland permanent card with mana value
//	3 or less from your graveyard to the battlefield with a finality
//	counter on it.
//
// Implementation:
//   - "creature_attacks": (a) if the attacker is a Bat we control, gain
//     1 life. (b) if the attacker is Zoraline herself, attempt
//     reanimation.
//   - OnETB: attempt reanimation.
//   - Reanimation: pay 2 life (no mana cost simulated), pick highest-CMC
//     nonland permanent card with CMC <= 3 from graveyard, return with
//     finality counter.
//   - Flying/vigilance are engine-handled.
func registerZoralineCosmosCaller(r *Registry) {
	r.OnETB("Zoraline, Cosmos Caller", zoralineETBReanimate)
	r.OnTrigger("Zoraline, Cosmos Caller", "creature_attacks", zoralineAttacks)
}

func zoralineETBReanimate(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	zoralineDoReanimate(gs, perm, "etb")
}

func zoralineAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk == nil || atk.Card == nil {
		return
	}
	if atk.Controller != perm.Controller {
		return
	}
	// Bat attacks: gain 1 life.
	if cardHasType(atk.Card, "bat") {
		gameengine.GainLife(gs, perm.Controller, 1, perm.Card.DisplayName())
		emit(gs, "zoraline_bat_attack_lifegain", perm.Card.DisplayName(), map[string]interface{}{
			"seat": perm.Controller,
			"bat":  atk.Card.DisplayName(),
		})
	}
	if atk == perm {
		zoralineDoReanimate(gs, perm, "attack")
	}
}

func zoralineDoReanimate(gs *gameengine.GameState, perm *gameengine.Permanent, kind string) {
	const slug = "zoraline_reanimate"
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	if seat.Life <= 2 {
		emitFail(gs, slug, perm.Card.DisplayName(), "life_too_low", map[string]interface{}{
			"life": seat.Life,
		})
		return
	}
	var pick *gameengine.Card
	bestCMC := -1
	for _, c := range seat.Graveyard {
		if c == nil {
			continue
		}
		if cardHasType(c, "land") {
			continue
		}
		isPerm := cardHasType(c, "creature") || cardHasType(c, "artifact") ||
			cardHasType(c, "enchantment") || cardHasType(c, "planeswalker") ||
			cardHasType(c, "battle")
		if !isPerm {
			continue
		}
		cmc := cardCMC(c)
		if cmc > 3 {
			continue
		}
		if cmc > bestCMC {
			bestCMC = cmc
			pick = c
		}
	}
	if pick == nil {
		return
	}
	gameengine.LoseLife(gs, perm.Controller, 2, perm.Card.DisplayName())
	moveCardBetweenZones(gs, perm.Controller, pick, "graveyard", "exile", "zoraline_reanimate_route")
	newPerm := enterBattlefieldWithETB(gs, perm.Controller, pick, false)
	if newPerm != nil {
		newPerm.AddCounter("finality", 1)
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"kind":     kind,
		"returned": pick.DisplayName(),
	})
}
