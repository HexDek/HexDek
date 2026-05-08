package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerVaziKeenNegotiator wires Vazi, Keen Negotiator.
//
// Oracle text:
//
//	Haste
//	{T}: Target opponent creates X Treasure tokens, where X is the number
//	of Treasure tokens you created this turn.
//	Whenever an opponent casts a spell or activates an ability, if mana
//	from a Treasure was spent to cast it or activate it, put a +1/+1
//	counter on target creature, then draw a card.
//
// Implementation:
//   - Activated ability (idx 0): tap Vazi, mint X Treasures for the
//     lowest-life opponent. X = seat.Turn.TreasuresCreated.
//   - The opponent-cast-with-treasure trigger requires per-spell mana
//     provenance, which the engine doesn't track. emitPartial.
func registerVaziKeenNegotiator(r *Registry) {
	r.OnActivated("Vazi, Keen Negotiator", vaziActivate)
}

func vaziActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "vazi_keen_negotiator_treasure_gift"
	if gs == nil || src == nil {
		return
	}
	if abilityIdx != 0 {
		return
	}
	if src.Tapped {
		emitFail(gs, slug, src.Card.DisplayName(), "already_tapped", nil)
		return
	}
	src.Tapped = true

	x := gs.Seats[src.Controller].Turn.TreasuresCreated
	if x <= 0 {
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat": src.Controller,
			"x":    0,
		})
		return
	}

	// Pick lowest-life living opponent.
	target := -1
	bestLife := 1 << 30
	for _, opp := range gs.Opponents(src.Controller) {
		s := gs.Seats[opp]
		if s == nil || s.Lost {
			continue
		}
		if s.Life < bestLife {
			bestLife = s.Life
			target = opp
		}
	}
	if target < 0 {
		emitFail(gs, slug, src.Card.DisplayName(), "no_opponent", nil)
		return
	}
	for i := 0; i < x; i++ {
		gameengine.CreateTreasureToken(gs, target)
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":   src.Controller,
		"target": target,
		"x":      x,
	})
	emitPartial(gs, slug, src.Card.DisplayName(), "treasure_mana_provenance_for_opp_cast_trigger_not_tracked")
}
