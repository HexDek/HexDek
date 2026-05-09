package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMayaelCustom implements Mayael the Anima's activated ability
// that the auto-generated stub leaves as `emitPartial` only.
//
// Oracle text:
//
//	{3}{R}{G}{W}, {T}: Look at the top five cards of your library. You
//	may put a creature card with power 5 or greater from among them
//	onto the battlefield. Put the rest on the bottom of your library in
//	any order.
//
// Picks the highest-power creature among the top 5 with power ≥ 5; if
// none qualifies, the top 5 just go to the bottom in their original
// order. The cost ({3}{R}{G}{W} + tap) is enforced by the engine before
// dispatch — this handler resolves the effect.
func registerMayaelCustom(r *Registry) {
	r.OnActivated("Mayael the Anima", mayaelLookFive)
}

func mayaelLookFive(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "mayael_look_five"
	if gs == nil || src == nil {
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil || len(seat.Library) == 0 {
		emitFail(gs, slug, src.Card.DisplayName(), "library_empty", nil)
		return
	}
	// Cost gates: {3}{R}{G}{W} = 6 generic from the engine's pool, plus
	// {T}. Defensive check for callers that bypass the engine activation
	// dispatcher.
	if src.Tapped {
		emitFail(gs, slug, src.Card.DisplayName(), "already_tapped", nil)
		return
	}
	if !payManaFromPool(seat, 6) {
		emitFail(gs, slug, src.Card.DisplayName(), "insufficient_mana", map[string]interface{}{
			"required":  6,
			"mana_pool": seat.ManaPool,
		})
		return
	}
	src.Tapped = true
	n := 5
	if n > len(seat.Library) {
		n = len(seat.Library)
	}
	top := append([]*gameengine.Card(nil), seat.Library[:n]...)
	var pick *gameengine.Card
	pickIdx := -1
	bestPower := 4 // strictly greater than 4 → ≥5
	for i, c := range top {
		if c == nil {
			continue
		}
		if !cardHasType(c, "creature") {
			continue
		}
		if c.BasePower > bestPower {
			pick = c
			pickIdx = i
			bestPower = c.BasePower
		}
	}
	if pick != nil {
		// Remove from library at pickIdx; the rest of `top` go to the bottom.
		seat.Library = seat.Library[n:]
		// Bottom the unpicked top cards in original order.
		for i, c := range top {
			if i == pickIdx {
				continue
			}
			seat.Library = append(seat.Library, c)
		}
		enterBattlefieldWithETB(gs, src.Controller, pick, false)
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":  src.Controller,
			"into_play": pick.DisplayName(),
			"power": bestPower,
		})
		return
	}
	// Nothing qualifies — bottom all five.
	seat.Library = append(seat.Library[n:], top...)
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":      src.Controller,
		"into_play": "",
		"note":      "no_qualifying_creature",
	})
}
