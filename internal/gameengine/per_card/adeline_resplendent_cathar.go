package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAdelineResplendentCathar wires Adeline, Resplendent Cathar.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	{1}{W}{W}
//	Legendary Creature — Human Knight
//	Vigilance
//	Adeline's power is equal to the number of creatures you control.
//	Whenever you attack, for each opponent, create a 1/1 white Human
//	creature token that's tapped and attacking that player or a
//	planeswalker they control.
//
// Implementation:
//   - ETB: refresh Adeline's temp_power to creature count (recompute on
//     ETB; layered static would need full layers wiring — emitPartial).
//   - declare_attackers gated to controller: mint one 1/1 W Human token
//     per living opponent, tapped + attacking.
func registerAdelineResplendentCathar(r *Registry) {
	r.OnETB("Adeline, Resplendent Cathar", adelineETB)
	r.OnTrigger("Adeline, Resplendent Cathar", "declare_attackers", adelineAttacks)
}

func adelineETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "adeline_resplendent_cathar_etb_buff"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	count := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if p.IsCreature() {
			count++
		}
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["temp_power"] += count - 1
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":           perm.Controller,
		"creature_count": count,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"power_equals_creature_count_only_refreshed_on_etb")
}

func adelineAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "adeline_attack_token_per_opponent"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	tokens := 0
	for i, opp := range gs.Seats {
		if opp == nil || opp.Lost || i == perm.Controller {
			continue
		}
		token := &gameengine.Card{
			Name:          "Human Token",
			Owner:         perm.Controller,
			BasePower:     1,
			BaseToughness: 1,
			Types:         []string{"token", "creature", "human"},
			Colors:        []string{"W"},
			TypeLine:      "Token Creature — Human",
		}
		enterBattlefieldWithETB(gs, perm.Controller, token, true)
		tokens++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"tokens": tokens,
	})
}
