package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerLordXanderTheCollector wires Lord Xander, the Collector.
//
// Oracle text:
//
//   When Lord Xander enters, target opponent discards half the cards in their hand, rounded down.
//   Whenever Lord Xander attacks, defending player mills half their library, rounded down.
//   When Lord Xander dies, target opponent sacrifices half the nonland permanents they control of their choice, rounded down.
//
// Implementation:
//   - ETB: pick the opponent with the largest hand and discard half (floor).
//   - Attack mill and dies-sacrifice triggers wired here as well.
func registerLordXanderTheCollector(r *Registry) {
	r.OnETB("Lord Xander, the Collector", lordXanderTheCollectorETB)
	r.OnTrigger("Lord Xander, the Collector", "attacks", lordXanderAttacks)
	r.OnTrigger("Lord Xander, the Collector", "dies", lordXanderDies)
}

func lordXanderTheCollectorETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "lord_xander_the_collector_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	// Pick opponent with the largest hand.
	target := -1
	best := 0
	for _, oppIdx := range gs.Opponents(seat) {
		opp := gs.Seats[oppIdx]
		if opp == nil || opp.Lost {
			continue
		}
		if len(opp.Hand) > best {
			best = len(opp.Hand)
			target = oppIdx
		}
	}
	if target < 0 || best <= 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      seat,
			"target":    target,
			"discarded": 0,
		})
		return
	}
	n := best / 2
	discarded := 0
	for i := 0; i < n; i++ {
		opp := gs.Seats[target]
		if len(opp.Hand) == 0 {
			break
		}
		card := opp.Hand[len(opp.Hand)-1]
		gameengine.DiscardCard(gs, card, target)
		discarded++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      seat,
		"target":    target,
		"discarded": discarded,
	})
}

func lordXanderAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "lord_xander_attacks_mill"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atkSeat, _ := ctx["seat"].(int)
	if atkSeat != perm.Controller {
		return
	}
	// Defending player: in 1v1 and most multiplayer attacks the
	// engine surfaces "defender" / "target_seat"; fall back to first
	// opponent.
	def := -1
	if v, ok := ctx["defender"].(int); ok {
		def = v
	} else if v, ok := ctx["target_seat"].(int); ok {
		def = v
	} else {
		opps := gs.Opponents(perm.Controller)
		if len(opps) > 0 {
			def = opps[0]
		}
	}
	if def < 0 || def >= len(gs.Seats) {
		return
	}
	defender := gs.Seats[def]
	if defender == nil || defender.Lost {
		return
	}
	n := len(defender.Library) / 2
	milled := 0
	for i := 0; i < n && len(defender.Library) > 0; i++ {
		c := defender.Library[0]
		gameengine.MoveCard(gs, c, def, "library", "graveyard", "lord_xander_attack_mill")
		milled++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"defender": def,
		"milled":   milled,
	})
}

func lordXanderDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "lord_xander_dies_sacrifice"
	if gs == nil || perm == nil {
		return
	}
	// Pick opponent with most nonland permanents.
	target := -1
	best := 0
	for _, oppIdx := range gs.Opponents(perm.Controller) {
		opp := gs.Seats[oppIdx]
		if opp == nil || opp.Lost {
			continue
		}
		count := 0
		for _, p := range opp.Battlefield {
			if p == nil || p.IsLand() {
				continue
			}
			count++
		}
		if count > best {
			best = count
			target = oppIdx
		}
	}
	if target < 0 || best <= 0 {
		return
	}
	n := best / 2
	opp := gs.Seats[target]
	// Sacrifice n nonland permanents (defender's choice — pick lowest CMC).
	sacrificed := 0
	for i := 0; i < n; i++ {
		var pick *gameengine.Permanent
		for _, p := range opp.Battlefield {
			if p == nil || p.IsLand() {
				continue
			}
			if pick == nil || cardCMC(p.Card) < cardCMC(pick.Card) {
				pick = p
			}
		}
		if pick == nil {
			break
		}
		gameengine.SacrificePermanent(gs, pick, "lord_xander_dies")
		sacrificed++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"target":     target,
		"sacrificed": sacrificed,
	})
}
