package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCharixTheRagingIsleCustom implements Charix's pump activation.
// The auto-generated activated stub leaves the effect side as a partial.
//
// Oracle text:
//
//	Spells your opponents cast that target Charix cost {2} more to cast.
//	{3}: Charix gets +X/-X until end of turn, where X is the number of
//	Islands you control.
//
// The {2}-tax static is engine territory (cost modifier scan) and noted
// as a partial. The activation paid cost ({3}) is enforced by the engine
// before dispatch — we just resolve the modifier swing here. X = islands
// controlled at activation time per CR §608.2b.
func registerCharixTheRagingIsleCustom(r *Registry) {
	r.OnActivated("Charix, the Raging Isle", charixActivatePump)
	r.OnETB("Charix, the Raging Isle", func(gs *gameengine.GameState, perm *gameengine.Permanent) {
		emitPartial(gs, "charix_target_tax", perm.Card.DisplayName(),
			"{2}-tax for opp spells targeting Charix needs cost-modifier hook")
	})
}

func charixActivatePump(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "charix_pump"
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
		// Per CR an Island is anything with the Island subtype, basic or
		// not (Tropical Island, Underground Sea, Hallowed Fountain, etc.).
		if cardHasType(p.Card, "land") && cardHasSubtype(p.Card, "island") {
			x++
		} else {
			// Scryfall sometimes encodes subtypes as a single comma-joined
			// token; check defensively.
			for _, t := range p.Card.Types {
				if strings.EqualFold(t, "island") {
					x++
					break
				}
			}
		}
	}
	if x == 0 {
		emitFail(gs, slug, src.Card.DisplayName(), "no_islands", nil)
		return
	}
	src.Modifications = append(src.Modifications, gameengine.Modification{
		Power:     x,
		Toughness: -x,
		Duration:  "until_end_of_turn",
		Timestamp: gs.NextTimestamp(),
	})
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":  src.Controller,
		"x":     x,
		"power": src.Power(),
	})
}
