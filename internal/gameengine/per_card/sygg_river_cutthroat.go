package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSyggRiverCutthroat wires Sygg, River Cutthroat.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	At the beginning of each end step, if an opponent lost 3 or more
//	life this turn, you may draw a card.
//
// Implementation:
//   - "end_step" trigger: scan all opponents' seat.Flags["life_lost_this_turn"]
//     (which the engine maintains via life_loss event accounting). If any
//     opponent lost >= 3, draw a card for Sygg's controller.
func registerSyggRiverCutthroat(r *Registry) {
	r.OnTrigger("Sygg, River Cutthroat", "end_step", syggEndStep)
}

func syggEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "sygg_end_step_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	triggered := false
	for _, oppIdx := range gs.Opponents(perm.Controller) {
		s := gs.Seats[oppIdx]
		if s == nil {
			continue
		}
		if s.Turn.LifeLost >= 3 {
			triggered = true
			break
		}
	}
	if !triggered {
		return
	}
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}
