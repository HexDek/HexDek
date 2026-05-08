package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerBetorKinToAll wires Betor, Kin to All.
//
// Oracle text:
//
//	Flying
//	At the beginning of your end step, if creatures you control have
//	total toughness 10 or greater, draw a card. Then if total toughness
//	20 or greater, untap each creature you control. Then if total
//	toughness 40 or greater, each opponent loses half their life,
//	rounded up.
func registerBetorKinToAll(r *Registry) {
	r.OnTrigger("Betor, Kin to All", "end_step", betorKinEndStep)
}

func betorKinEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "betor_kin_to_all_end_step"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	total := 0
	for _, p := range seat.Battlefield {
		if p == nil || !p.IsCreature() {
			continue
		}
		t := p.Toughness()
		if t > 0 {
			total += t
		}
	}
	drewCard := false
	untapped := 0
	lifeLost := 0
	if total >= 10 {
		drawOne(gs, perm.Controller, perm.Card.DisplayName())
		drewCard = true
	}
	if total >= 20 {
		for _, p := range seat.Battlefield {
			if p != nil && p.IsCreature() && p.Tapped {
				p.Tapped = false
				untapped++
			}
		}
	}
	if total >= 40 {
		for i, opp := range gs.Seats {
			if opp == nil || i == perm.Controller || opp.Lost {
				continue
			}
			loss := (opp.Life + 1) / 2
			if loss > 0 {
				gameengine.LoseLife(gs, i, loss, perm.Card.DisplayName())
				lifeLost += loss
			}
		}
		_ = gs.CheckEnd()
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"toughness": total,
		"drew_card": drewCard,
		"untapped":  untapped,
		"life_lost": lifeLost,
	})
}
