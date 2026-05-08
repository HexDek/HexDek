package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMahadiEmporiumMaster wires Mahadi, Emporium Master.
//
// Oracle text:
//
//	At the beginning of your end step, create a Treasure token for each
//	creature that died this turn.
//
// Implementation: track creatures-died-this-turn via a perm flag
// (creature_dies observer increments). At end step, create that many
// treasure tokens, then reset the counter.
func registerMahadiEmporiumMaster(r *Registry) {
	r.OnTrigger("Mahadi, Emporium Master", "end_step", mahadiEndStep)
}

func mahadiEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "mahadi_end_step_treasures"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	count := 0
	for _, s := range gs.Seats {
		if s != nil {
			count += s.Turn.CreaturesDied
		}
	}
	if count <= 0 {
		return
	}
	for i := 0; i < count; i++ {
		gameengine.CreateTreasureToken(gs, perm.Controller)
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"treasures": count,
	})
}
