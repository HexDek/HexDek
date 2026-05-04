package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerShireiShizosCaretaker wires Shirei, Shizo's Caretaker.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Whenever a creature with power 1 or less is put into your graveyard
//	  from the battlefield, you may return that card to the battlefield
//	  at the beginning of the next end step if Shirei is still on the
//	  battlefield.
//
// Implementation:
//   - "creature_dies": filter to creatures whose owner is Shirei's
//     controller (the dying card goes to that owner's graveyard) AND
//     dying perm.Power() <= 1 at the time of death. Tag the card with
//     a "shirei_pending" entry in Card.Types.
//   - "end_step_controller": for each tagged card in our graveyard,
//     route it back to the battlefield.
func registerShireiShizosCaretaker(r *Registry) {
	r.OnTrigger("Shirei, Shizo's Caretaker", "creature_dies", shireiOnDies)
	r.OnTrigger("Shirei, Shizo's Caretaker", "end_step", shireiOnEndStep)
}

const shireiPendingTag = "shirei_pending"

func shireiOnDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "shirei_dying_creature_tag"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	dyingPerm, _ := ctx["perm"].(*gameengine.Permanent)
	dyingCard, _ := ctx["card"].(*gameengine.Card)
	if dyingCard == nil {
		return
	}
	if dyingCard.Owner != perm.Controller {
		return
	}
	if dyingPerm != nil && dyingPerm.IsToken() {
		return
	}
	if dyingPerm != nil {
		if dyingPerm.Power() > 1 {
			return
		}
	} else if dyingCard.BasePower > 1 {
		return
	}
	if !cardHasType(dyingCard, "creature") {
		return
	}
	dyingCard.Types = append(dyingCard.Types, shireiPendingTag)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"creature": dyingCard.DisplayName(),
	})
}

func shireiOnEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "shirei_end_step_return"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	returned := []string{}
	// Walk a copy of graveyard since we mutate it.
	gy := append([]*gameengine.Card(nil), seat.Graveyard...)
	for _, c := range gy {
		if c == nil {
			continue
		}
		hasTag := false
		for _, t := range c.Types {
			if t == shireiPendingTag {
				hasTag = true
				break
			}
		}
		if !hasTag {
			continue
		}
		// Strip tag.
		newTypes := c.Types[:0]
		for _, t := range c.Types {
			if t != shireiPendingTag {
				newTypes = append(newTypes, t)
			}
		}
		c.Types = newTypes
		gameengine.MoveCard(gs, c, perm.Controller, "graveyard", "battlefield", "shirei_return")
		createPermanent(gs, perm.Controller, c, false)
		returned = append(returned, c.DisplayName())
	}
	if len(returned) > 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":     perm.Controller,
			"returned": returned,
		})
	}
}
