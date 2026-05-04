package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSlimefootTheStowaway wires Slimefoot, the Stowaway.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Whenever a Saproling you control dies, Slimefoot deals 1 damage
//	to each opponent and you gain 1 life.
//	{4}: Create a 1/1 green Saproling creature token.
//
// Implementation:
//   - "creature_dies" trigger: when a Saproling controlled by Slimefoot's
//     controller dies, ping each opponent for 1 and gain 1 life.
//   - Activated ability: mint a Saproling token (mana cost gating handled
//     by activation pipeline).
func registerSlimefootTheStowaway(r *Registry) {
	r.OnTrigger("Slimefoot, the Stowaway", "creature_dies", slimefootSaprolingDies)
	r.OnActivated("Slimefoot, the Stowaway", slimefootActivate)
}

func slimefootSaprolingDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "slimefoot_saproling_died_drain"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	deadController, _ := ctx["controller_seat"].(int)
	if deadController != perm.Controller {
		return
	}
	deadCard, _ := ctx["card"].(*gameengine.Card)
	if deadCard == nil {
		return
	}
	isSap := false
	for _, t := range deadCard.Types {
		if strings.EqualFold(t, "saproling") {
			isSap = true
			break
		}
	}
	if !isSap {
		return
	}
	hits := 0
	for _, oppIdx := range gs.Opponents(perm.Controller) {
		s := gs.Seats[oppIdx]
		if s == nil || s.Lost {
			continue
		}
		s.Life--
		gs.LogEvent(gameengine.Event{
			Kind:   "damage",
			Seat:   perm.Controller,
			Target: oppIdx,
			Source: perm.Card.DisplayName(),
			Amount: 1,
		})
		hits++
	}
	gs.Seats[perm.Controller].Life++
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"pings": hits,
	})
}

func slimefootActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "slimefoot_activated_saproling"
	if gs == nil || src == nil {
		return
	}
	gameengine.CreateCreatureToken(gs, src.Controller, "Saproling",
		[]string{"creature", "saproling", "pip:G"}, 1, 1)
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat": src.Controller,
	})
}
