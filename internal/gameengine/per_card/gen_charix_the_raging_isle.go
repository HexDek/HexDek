package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCharixTheRagingIsle wires Charix, the Raging Isle.
//
// Oracle text:
//
//   Spells your opponents cast that target Charix cost {2} more to cast.
//   {3}: Charix gets +X/-X until end of turn, where X is the number of Islands you control.
//
// Implementation:
//   - {3}: Charix gets +X/-X until end of turn, X = Islands controlled.
//   - The opponent-target cost increase is handled by the cost-mod
//     pipeline.
func registerCharixTheRagingIsle(r *Registry) {
	r.OnActivated("Charix, the Raging Isle", charixTheRagingIsleActivate)
}

func charixTheRagingIsleActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "charix_pump_x_islands"
	if gs == nil || src == nil {
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	x := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if cardHasType(p.Card, "island") {
			x++
		}
	}
	if x <= 0 {
		emitFail(gs, slug, src.Card.DisplayName(), "no_islands", map[string]interface{}{
			"seat": src.Controller,
		})
		return
	}
	src.Modifications = append(src.Modifications, gameengine.Modification{
		Power:     x,
		Toughness: -x,
		Duration:  "until_end_of_turn",
		Timestamp: gs.NextTimestamp(),
	})
	gs.InvalidateCharacteristicsCache()
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat": src.Controller,
		"x":    x,
	})
}
