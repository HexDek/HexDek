package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGoblinGoliath wires Goblin Goliath (Muninn parser-gap #61, ~12K
// hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{4}{R}{R}
//	Creature — Goblin Mutant
//	When this creature enters, create a number of 1/1 red Goblin
//	creature tokens equal to the number of opponents you have.
//	{3}{R}, {T}: If a source you control would deal damage to an
//	opponent this turn, it deals double that damage to that player
//	instead.
//
// Implementation:
//   - ETB: count living opponents, mint that many 1/1 red Goblin tokens.
//   - Activated double-damage ability: not implemented (continuous
//     replacement effect on damage assignment). emitPartial.
func registerGoblinGoliath(r *Registry) {
	r.OnETB("Goblin Goliath", goblinGoliathETB)
}

func goblinGoliathETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "goblin_goliath_etb_goblin_tokens"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	living := 0
	for _, opp := range gs.Opponents(perm.Controller) {
		s := gs.Seats[opp]
		if s == nil || s.Lost {
			continue
		}
		living++
	}
	for i := 0; i < living; i++ {
		token := &gameengine.Card{
			Name:          "Goblin Token",
			Owner:         perm.Controller,
			Types:         []string{"creature", "token", "goblin", "pip:R"},
			Colors:        []string{"R"},
			BasePower:     1,
			BaseToughness: 1,
		}
		enterBattlefieldWithETB(gs, perm.Controller, token, false)
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"opponents": living,
		"tokens":    living,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"activated_double_damage_replacement_unimplemented")
}
