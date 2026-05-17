package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerYathanRoadwatcher wires Yathan Roadwatcher (Muninn parser-gap
// #62, ~11K hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{1}{W}{B}{G}
//	Creature — Human Scout
//	When this creature enters, if you cast it, mill four cards. When
//	you do, return target creature card with mana value 3 or less from
//	your graveyard to the battlefield.
//
// Implementation:
//   - Gate on was_cast (cast clause).
//   - Mill the top 4 cards from controller's library to graveyard.
//   - Resolve the reflexive "when you do" by scanning the freshly-milled
//     pile plus the rest of the graveyard for creature cards with CMC<=3,
//     picking the highest-CMC such card and returning it to battlefield.
func registerYathanRoadwatcher(r *Registry) {
	r.OnETB("Yathan Roadwatcher", yathanRoadwatcherETB)
}

func yathanRoadwatcherETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "yathan_roadwatcher_mill4_reanimate"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil || perm.Flags["was_cast"] != 1 {
		emitFail(gs, slug, perm.Card.DisplayName(), "not_cast", nil)
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	milled := []string{}
	for i := 0; i < 4 && len(seat.Library) > 0; i++ {
		top := seat.Library[0]
		if top == nil {
			seat.Library = seat.Library[1:]
			continue
		}
		gameengine.MoveCard(gs, top, perm.Controller, "library", "graveyard", slug)
		milled = append(milled, top.DisplayName())
	}
	var best *gameengine.Card
	bestCMC := -1
	for _, c := range seat.Graveyard {
		if c == nil || !cardHasType(c, "creature") {
			continue
		}
		cmc := cardCMC(c)
		if cmc > 3 {
			continue
		}
		if cmc > bestCMC {
			bestCMC = cmc
			best = c
		}
	}
	if best == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":           perm.Controller,
			"milled":         milled,
			"reanimated":     "none",
			"reason":         "no_eligible_creature_in_graveyard",
		})
		return
	}
	gameengine.MoveCard(gs, best, perm.Controller, "graveyard", "battlefield", slug)
	enterBattlefieldWithETB(gs, perm.Controller, best, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"milled":     milled,
		"reanimated": best.DisplayName(),
		"cmc":        bestCMC,
	})
}
