package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMishraClaimedByGix wires Mishra, Claimed by Gix.
//
// Oracle text:
//
//	Whenever you attack, each opponent loses X life and you gain X life,
//	where X is the number of attacking creatures.
//	If Mishra, Claimed by Gix and a creature named Phyrexian Dragon
//	Engine are attacking, and you both own and control them, exile them,
//	then meld them into Mishra, Lost to Phyrexia.
//
// Meld is left as a parser gap. Drain is wired on the attacks trigger.
func registerMishraClaimedByGix(r *Registry) {
	r.OnTrigger("Mishra, Claimed by Gix", "attacks", mishraGixAttacks)
}

func mishraGixAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "mishra_gix_drain"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	attackerSeat, _ := ctx["seat"].(int)
	if attackerSeat != perm.Controller {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	if perm.Flags["mishra_gix_drained_turn"] == gs.Turn {
		return
	}
	perm.Flags["mishra_gix_drained_turn"] = gs.Turn

	x := 0
	for _, p := range gs.Seats[attackerSeat].Battlefield {
		if p == nil {
			continue
		}
		if p.IsAttacking() {
			x++
		}
	}
	if x <= 0 {
		return
	}
	for _, opp := range gs.Opponents(perm.Controller) {
		if opp < 0 || opp >= len(gs.Seats) {
			continue
		}
		s := gs.Seats[opp]
		if s == nil || s.Lost {
			continue
		}
		gameengine.FireLoseLifeEvent(gs, opp, x, perm)
	}
	gameengine.GainLife(gs, perm.Controller, x, "Mishra, Claimed by Gix")
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"attackers": x,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(), "meld_with_phyrexian_dragon_engine_unimplemented")
}
