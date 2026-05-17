package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKnightOfTheWhiteOrchid wires Knight of the White Orchid.
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	First strike
//	When this creature enters, if an opponent controls more lands than
//	you, you may search your library for a Plains card, put it onto the
//	battlefield, then shuffle.
//
// Implementation (Muninn gap #10 — 93K hits):
//   - OnETB: count battlefield lands for the controller and each
//     opponent. If any opponent has strictly more lands, scan the
//     controller's library for the first Plains (basic or otherwise —
//     "a Plains card" matches any card with the Plains subtype, CR
//     §201.3a), move it onto the battlefield via the standard ETB
//     pipeline, then shuffle.
//   - The "you may" choice is auto-accepted: a free untapped Plains is
//     pure upside whenever the condition is satisfied.
//   - First strike is an AST keyword; not modeled here.
func registerKnightOfTheWhiteOrchid(r *Registry) {
	r.OnETB("Knight of the White Orchid", knightOfTheWhiteOrchidETB)
}

func knightOfTheWhiteOrchidETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "knight_of_the_white_orchid_plains_search"
	if gs == nil || perm == nil || perm.Card == nil {
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

	myLands := countBattlefieldLands(s.Battlefield)
	opponentHasMore := false
	for i, os := range gs.Seats {
		if i == seat || os == nil || os.Lost {
			continue
		}
		if countBattlefieldLands(os.Battlefield) > myLands {
			opponentHasMore = true
			break
		}
	}
	if !opponentHasMore {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      seat,
			"my_lands":  myLands,
			"triggered": false,
		})
		return
	}

	var plains *gameengine.Card
	for _, c := range s.Library {
		if c == nil {
			continue
		}
		if cardHasSubtype(c, "plains") || cardHasType(c, "plains") {
			plains = c
			break
		}
	}
	if plains == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_plains_in_library", map[string]interface{}{
			"seat": seat,
		})
		shuffleLibraryPerCard(gs, seat)
		return
	}

	gameengine.MoveCard(gs, plains, seat, "library", "battlefield", "knight_white_orchid_search")
	shuffleLibraryPerCard(gs, seat)

	gs.LogEvent(gameengine.Event{
		Kind:   "search_library",
		Seat:   seat,
		Source: perm.Card.DisplayName(),
		Details: map[string]interface{}{
			"found":  []string{plains.DisplayName()},
			"reason": "knight_white_orchid",
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      seat,
		"my_lands":  myLands,
		"triggered": true,
		"found":     plains.DisplayName(),
	})
}
