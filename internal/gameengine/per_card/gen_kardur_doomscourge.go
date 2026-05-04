package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKardurDoomscourge wires Kardur, Doomscourge.
//
// Oracle text:
//
//   When Kardur enters, until your next turn, creatures your opponents control attack each combat if able and attack a player other than you if able.
//   Whenever an attacking creature dies, each opponent loses 1 life and you gain 1 life.
//
// Implementation:
//   - "creature_dies" gated to dying perm being an attacker: each
//     opponent loses 1, controller gains 1.
//   - The ETB goad-style attack-forcing is engine-state and not modeled
//     here — emitPartial.
func registerKardurDoomscourge(r *Registry) {
	r.OnETB("Kardur, Doomscourge", kardurDoomscourgeETB)
	r.OnTrigger("Kardur, Doomscourge", "creature_dies", kardurAttackingDies)
}

func kardurDoomscourgeETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "kardur_doomscourge_etb", perm.Card.DisplayName(),
		"opponents_creatures_must_attack_until_your_next_turn_unimplemented")
}

func kardurAttackingDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kardur_attacker_dies_drain"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	dyingPerm, _ := ctx["perm"].(*gameengine.Permanent)
	if dyingPerm == nil {
		return
	}
	if !dyingPerm.IsAttacking() {
		return
	}
	for _, opp := range gs.Opponents(perm.Controller) {
		s := gs.Seats[opp]
		if s == nil || s.Lost {
			continue
		}
		s.Life--
		gs.LogEvent(gameengine.Event{
			Kind:   "life_change",
			Seat:   opp,
			Source: perm.Card.DisplayName(),
			Amount: -1,
			Details: map[string]interface{}{
				"slug":   slug,
				"reason": "kardur_drain_loss",
			},
		})
	}
	gameengine.GainLife(gs, perm.Controller, 1, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"dying": dyingPerm.Card.DisplayName(),
	})
	_ = gs.CheckEnd()
}
