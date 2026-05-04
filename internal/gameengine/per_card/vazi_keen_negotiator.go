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
//     lowest-life opponent (helps the most-pressured player to bait deals).
//     X = perm.Flags["vazi_treasures_minted_turn"+turnKey] proxy: we read
//     ctx["treasures_this_turn"] when present; otherwise default 0.
//   - "token_created": when *we* create a Treasure, increment our
//     per-turn counter (engine doesn't track this natively).
//   - The opponent-cast-with-treasure trigger requires per-spell mana
//     provenance, which the engine doesn't track. emitPartial.
func registerVaziKeenNegotiator(r *Registry) {
	r.OnActivated("Vazi, Keen Negotiator", vaziActivate)
	r.OnTrigger("Vazi, Keen Negotiator", "token_created", vaziTokenCreated)
}

func vaziTreasuresKey(turn int) string {
	return "vazi_treasures_this_turn"
}

func vaziTokenCreated(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	creatorSeat, _ := ctx["controller_seat"].(int)
	if creatorSeat != perm.Controller {
		return
	}
	types, _ := ctx["types"].([]string)
	isTreasure := false
	for _, t := range types {
		if t == "treasure" || t == "Treasure" {
			isTreasure = true
			break
		}
	}
	if !isTreasure {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	// Reset per-turn counter on turn change.
	if perm.Flags["vazi_turn"] != gs.Turn {
		perm.Flags["vazi_turn"] = gs.Turn
		perm.Flags[vaziTreasuresKey(gs.Turn)] = 0
	}
	perm.Flags[vaziTreasuresKey(gs.Turn)]++
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

	x := 0
	if src.Flags != nil {
		x = src.Flags[vaziTreasuresKey(gs.Turn)]
	}
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
