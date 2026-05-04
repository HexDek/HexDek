package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerElugeTheShorelessSea wires Eluge, the Shoreless Sea.
//
// Oracle text:
//
//	Eluge's power and toughness are each equal to the number of Islands
//	you control.
//	Whenever Eluge enters or attacks, put a flood counter on target
//	land. It's an Island in addition to its other types for as long as
//	it has a flood counter on it.
//	The first instant or sorcery spell you cast each turn costs {U}
//	(or {1}) less to cast for each land you control with a flood
//	counter on it.
//
// CDA p/t and cost reduction are AST-resolved. We wire ETB/attack flood
// counter placement.
func registerElugeTheShorelessSea(r *Registry) {
	r.OnETB("Eluge, the Shoreless Sea", elugeFlood)
	r.OnTrigger("Eluge, the Shoreless Sea", "attacks", elugeFloodAttack)
}

func elugeFlood(gs *gameengine.GameState, perm *gameengine.Permanent) {
	elugePlaceFloodCounter(gs, perm)
}

func elugeFloodAttack(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if ctx == nil {
		return
	}
	attackerSeat, _ := ctx["seat"].(int)
	if attackerSeat != perm.Controller {
		return
	}
	elugePlaceFloodCounter(gs, perm)
}

func elugePlaceFloodCounter(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "eluge_flood_counter"
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	// Pick a non-Island land we control without a flood counter yet.
	for _, p := range seat.Battlefield {
		if p == nil || !p.IsLand() {
			continue
		}
		if p.Counters != nil && p.Counters["flood"] > 0 {
			continue
		}
		if cardHasType(p.Card, "island") {
			continue
		}
		p.AddCounter("flood", 1)
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"target": p.Card.DisplayName(),
		})
		return
	}
}
