package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerDoomsdayExcruciator wires Doomsday Excruciator (Muninn parser-gap
// #89, ~5.6K hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{6}{B}{B}
//	Creature — Avatar Horror
//	Flying
//	When this creature enters, if it was cast, each player exiles all
//	but the bottom six cards of their library face down.
//	At the beginning of your upkeep, draw a card.
//
// Implementation:
//   - OnETB gated on was_cast: for each living seat, exile every card above
//     the bottom 6 of their library. Routed via MoveCard so zone-change
//     triggers (Mindcrank-style) fire correctly. Face-down distinction is
//     not modeled (engine has no face-down library/exile state) — partial.
//   - "upkeep_controller" gated on active_seat == controller: drawOne.
func registerDoomsdayExcruciator(r *Registry) {
	r.OnETB("Doomsday Excruciator", doomsdayExcruciatorETB)
	r.OnTrigger("Doomsday Excruciator", "upkeep_controller", doomsdayExcruciatorUpkeep)
}

func doomsdayExcruciatorETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "doomsday_excruciator_etb_exile_library_above_six"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil || perm.Flags["was_cast"] != 1 {
		emitFail(gs, slug, perm.Card.DisplayName(), "not_cast", nil)
		return
	}
	totals := map[int]int{}
	for i, seat := range gs.Seats {
		if seat == nil || seat.Lost {
			continue
		}
		if len(seat.Library) <= 6 {
			continue
		}
		// Exile every card except the bottom 6. The bottom-6 are the LAST
		// six entries of the Library slice (top of library = index 0).
		exileCount := len(seat.Library) - 6
		// MoveCard mutates the library each call; pull the top card
		// repeatedly so indexing stays stable.
		for j := 0; j < exileCount; j++ {
			top := seat.Library[0]
			if top == nil {
				seat.Library = seat.Library[1:]
				continue
			}
			gameengine.MoveCard(gs, top, i, "library", "exile", slug)
		}
		totals[i] = exileCount
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"exiled": totals,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"face_down_exile_state_unmodeled")
}

func doomsdayExcruciatorUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "doomsday_excruciator_upkeep_draw"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	activeSeat, ok := ctx["active_seat"].(int)
	if !ok || activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
		"draw": 1,
	})
}
