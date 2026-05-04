package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerZethiArcaneBlademaster wires Zethi, Arcane Blademaster.
//
// Oracle text:
//
//   Multikicker {W/U}
//   When Zethi, Arcane Blademaster enters, exile up to X target instant cards from your graveyard, where X is the number of times Zethi was kicked. Put a kick counter on each of them.
//   Whenever Zethi attacks, copy each exiled card you own with a kick counter on it. You may cast the copies.
//
// Implementation:
//   - ETB: exile up to X instant cards from graveyard (X = times kicked,
//     read from perm.Flags["kicked"]). Track exiled instants on the
//     controller seat as Flags so the attack trigger can reference them.
//   - Attack trigger: copy each tracked instant and cast the copy.
//     Casting copies isn't surfaced cleanly here, so we emit a partial
//     and just log the intent.
func registerZethiArcaneBlademaster(r *Registry) {
	r.OnETB("Zethi, Arcane Blademaster", zethiArcaneBlademasterETB)
	r.OnTrigger("Zethi, Arcane Blademaster", "attacks", zethiArcaneBlademasterAttack)
}

func zethiArcaneBlademasterETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "zethi_arcane_blademaster_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil {
		return
	}
	x := 0
	if perm.Flags != nil {
		x = perm.Flags["kicked"]
	}
	if x <= 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   seat,
			"x":      0,
			"exiled": 0,
		})
		return
	}
	exiled := []string{}
	for i := 0; i < x; i++ {
		idx := -1
		for j, c := range s.Graveyard {
			if c == nil {
				continue
			}
			if cardHasType(c, "instant") {
				idx = j
				break
			}
		}
		if idx < 0 {
			break
		}
		card := s.Graveyard[idx]
		gameengine.MoveCard(gs, card, seat, "graveyard", "exile", "zethi_exile")
		// Mark the card so the attack trigger can find it.
		if card != nil {
			card.Types = append(card.Types, "kick_counter:zethi")
		}
		exiled = append(exiled, card.DisplayName())
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   seat,
		"x":      x,
		"exiled": exiled,
	})
}

func zethiArcaneBlademasterAttack(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "zethi_arcane_blademaster_attack"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atkSeat, _ := ctx["seat"].(int)
	if atkSeat != perm.Controller {
		return
	}
	s := gs.Seats[perm.Controller]
	if s == nil {
		return
	}
	tagged := []string{}
	for _, c := range s.Exile {
		if c == nil {
			continue
		}
		for _, t := range c.Types {
			if t == "kick_counter:zethi" {
				tagged = append(tagged, c.DisplayName())
				break
			}
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"copies": tagged,
	})
	if len(tagged) > 0 {
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"copy_and_cast_exiled_instants_unimplemented")
	}
}
