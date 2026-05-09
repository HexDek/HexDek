package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAmaliaBenavidesCustom upgrades the auto-generated Amalia stub
// (which only stuck a +1/+1 counter on every life-gain trigger) with a
// proper Explore + power-20 board-wipe trigger.
//
// Oracle text:
//
//	Ward—Pay 3 life.
//	Whenever you gain life, Amalia Benavides Aguirre explores. Then
//	destroy all other creatures if its power is exactly 20. (To have
//	this creature explore, reveal the top card of your library. Put
//	that card into your hand if it's a land. Otherwise, put a +1/+1
//	counter on this creature, then put the card back or put it into
//	your graveyard.)
//
// Explore is modeled deterministically: peek the top card; if it's a
// land, put it in hand. Otherwise, +1/+1 counter on Amalia and put the
// card on top (we never mill — keeping the card on top is the
// EV-optimal choice when the next draw is the same card revealed).
// After the explore, if Amalia's power == 20, destroy all OTHER
// creatures across all seats.
//
// Ward—Pay 3 life is engine-side (covered by ward layer) — we surface
// it as an emitPartial breadcrumb only when the auto-gen registered as
// "amalia_ward".
func registerAmaliaBenavidesCustom(r *Registry) {
	r.OnTrigger("Amalia Benavides Aguirre", "life_gained", amaliaExploreOnLifeGain)
}

func amaliaExploreOnLifeGain(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "amalia_explore_on_lifegain"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	gainSeat, _ := ctx["seat"].(int)
	if gainSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	// Explore: peek top card.
	wasLand := false
	if len(seat.Library) > 0 {
		top := seat.Library[0]
		if top != nil && cardHasType(top, "land") {
			gameengine.MoveCard(gs, top, perm.Controller, "library", "hand", "amalia_explore_land")
			wasLand = true
		} else {
			perm.AddCounter("+1/+1", 1)
			// Card stays on top — modeled as no-op.
		}
	} else {
		// Empty library — still grants the counter per CR 701.40.
		perm.AddCounter("+1/+1", 1)
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"was_land": wasLand,
		"power":    perm.Power(),
	})
	// Power == 20 → destroy all OTHER creatures.
	if perm.Power() == 20 {
		killed := 0
		for _, s := range gs.Seats {
			if s == nil {
				continue
			}
			survivors := s.Battlefield[:0]
			for _, p := range s.Battlefield {
				if p == nil {
					continue
				}
				if p == perm || !p.IsCreature() {
					survivors = append(survivors, p)
					continue
				}
				if p.Card != nil {
					gameengine.MoveCard(gs, p.Card, p.Controller, "battlefield", "graveyard", "amalia_pow20_wipe")
				}
				killed++
			}
			s.Battlefield = survivors
		}
		emit(gs, "amalia_power20_wipe", perm.Card.DisplayName(), map[string]interface{}{
			"seat":            perm.Controller,
			"creatures_wiped": killed,
		})
	}
}
