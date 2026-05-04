package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSauronLordOfTheRings wires Sauron, Lord of the Rings.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	When you cast this spell, amass Orcs 5, mill five cards, then
//	return a creature card from your graveyard to the battlefield.
//	Trample
//	Whenever a commander an opponent controls dies, the Ring tempts
//	you.
//
// Implementation:
//   - "spell_cast" trigger: when this card is cast by its owner, fire
//     the cast-trigger cascade — amass Orcs 5, mill five, then reanimate
//     the highest-CMC creature from controller's graveyard. The cast
//     trigger fires from the stack, so we listen on spell_cast and
//     match on card identity.
//   - Trample handled by AST keyword pipeline.
//   - "creature_dies" trigger: when an opponent's commander dies, ring
//     tempts Sauron's controller.
func registerSauronLordOfTheRings(r *Registry) {
	r.OnTrigger("Sauron, Lord of the Rings", "spell_cast", sauronLOTRCast)
	r.OnTrigger("Sauron, Lord of the Rings", "creature_dies", sauronLOTROpponentCommanderDies)
}

func sauronLOTRCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "sauron_lotr_cast_trigger"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	cardCast, _ := ctx["card"].(*gameengine.Card)
	if cardCast == nil || !strings.EqualFold(cardCast.DisplayName(), perm.Card.DisplayName()) {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	for i := 0; i < 5; i++ {
		amassOrcs(gs, perm.Controller)
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	milled := 0
	for i := 0; i < 5 && len(seat.Library) > 0; i++ {
		top := seat.Library[0]
		gameengine.MoveCard(gs, top, perm.Controller, "library", "graveyard", "sauron_mill")
		milled++
	}
	// Reanimate highest-CMC creature.
	var best *gameengine.Card
	bestCMC := -1
	for _, c := range seat.Graveyard {
		if c == nil || !cardHasType(c, "creature") {
			continue
		}
		if cardCMC(c) > bestCMC {
			bestCMC = cardCMC(c)
			best = c
		}
	}
	if best != nil {
		gameengine.MoveCard(gs, best, perm.Controller, "graveyard", "battlefield", "sauron_reanimate")
		createPermanent(gs, perm.Controller, best, false)
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"amassed":    5,
		"milled":     milled,
		"reanimated": cardDisp(best),
	})
}

func sauronLOTROpponentCommanderDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "sauron_lotr_opp_commander_dies_ring_tempt"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	deadController, _ := ctx["controller_seat"].(int)
	if deadController == perm.Controller {
		return
	}
	deadCard, _ := ctx["card"].(*gameengine.Card)
	if deadCard == nil {
		return
	}
	// Is it a commander? Check the seat's commander identity.
	s := gs.Seats[deadController]
	if s == nil {
		return
	}
	isCommander := false
	for _, name := range s.CommanderNames {
		if strings.EqualFold(name, deadCard.DisplayName()) {
			isCommander = true
			break
		}
	}
	if !isCommander {
		return
	}
	gameengine.TheRingTemptsYou(gs, perm.Controller)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         perm.Controller,
		"dead_cmdr":    deadCard.DisplayName(),
		"opp_seat":     deadController,
	})
}
