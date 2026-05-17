package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerElderscaleWurm wires Elderscale Wurm (Muninn parser-gap #98,
// ~3.6K hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{4}{G}{G}{G}
//	Creature — Wurm
//	Trample
//	When this creature enters, if your life total is less than 7, your
//	life total becomes 7.
//	As long as you have 7 or more life, damage that would reduce your
//	life total to less than 7 reduces it to 7 instead.
//
// Implementation:
//   - Trample: AST keyword pipeline.
//   - OnETB: if controller's life < 7, set life to 7 (a one-shot, not the
//     continuous floor effect). LogEvent with Kind "life_set" so downstream
//     state-based action checks and the spectator stream can see the swing.
//   - The continuous "damage that would reduce your life to less than 7
//     reduces it to 7 instead" is a damage replacement effect (CR §615)
//     gated on the controlling player having ≥7 life. This requires the
//     damage-replacement layer that other floor effects (Worship-style)
//     also need — partial.
func registerElderscaleWurm(r *Registry) {
	r.OnETB("Elderscale Wurm", elderscaleWurmETB)
}

func elderscaleWurmETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "elderscale_wurm_etb_life_to_seven"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	if seat.Life >= 7 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"life":      seat.Life,
		})
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"continuous_damage_floor_replacement_unmodeled")
		return
	}
	before := seat.Life
	seat.Life = 7
	gs.LogEvent(gameengine.Event{
		Kind:   "life_set",
		Seat:   perm.Controller,
		Source: perm.Card.DisplayName(),
		Amount: 7,
		Details: map[string]interface{}{
			"before": before,
			"after":  7,
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"triggered": true,
		"before":    before,
		"after":     7,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"continuous_damage_floor_replacement_unmodeled")
}
