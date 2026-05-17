package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerClaimJumper wires Claim Jumper.
//
// Oracle text (Scryfall, verified 2026-05-16):
//
//	Vigilance
//	When this creature enters, if an opponent controls more lands than
//	you, you may search your library for a Plains card and put it onto
//	the battlefield tapped. Then if an opponent controls more lands
//	than you, repeat this process once. If you search your library
//	this way, shuffle.
//
// Implementation:
//   - Vigilance: AST keyword pipeline.
//   - ETB: two-pass land-tax-style search. Each pass re-evaluates the
//     opponent-controls-more-lands gate using *current* land counts,
//     so a Plains fetched on pass 1 narrows the gap before pass 2.
//   - We accept any Plains card (basic or non-basic — "Plains card" is
//     a subtype filter, not "basic Plains card"). enterBattlefieldWithETB
//     puts it onto the battlefield tapped, firing landfall + ETB triggers
//     so Valakut Exploration etc. see the land arrive.
//   - Single shuffle at the end iff at least one search happened (the
//     printed text only shuffles "if you search").
func registerClaimJumper(r *Registry) {
	r.OnETB("Claim Jumper", claimJumperETB)
}

func claimJumperETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "claim_jumper_etb_fetch_plains"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}

	found := []string{}
	for pass := 0; pass < 2; pass++ {
		if !claimJumperOpponentHasMoreLands(gs, perm.Controller) {
			break
		}
		plains := claimJumperFindPlains(seat.Library)
		if plains == nil {
			break
		}
		gameengine.MoveCard(gs, plains, perm.Controller, "library", "battlefield", "claim_jumper_search")
		enterBattlefieldWithETB(gs, perm.Controller, plains, true)
		found = append(found, plains.DisplayName())
	}
	if len(found) > 0 {
		shuffleLibraryPerCard(gs, perm.Controller)
	}

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"found": found,
	})
}

func claimJumperOpponentHasMoreLands(gs *gameengine.GameState, mySeat int) bool {
	if gs == nil || mySeat < 0 || mySeat >= len(gs.Seats) {
		return false
	}
	mine := countBattlefieldLands(gs.Seats[mySeat].Battlefield)
	for i, s := range gs.Seats {
		if i == mySeat || s == nil {
			continue
		}
		if countBattlefieldLands(s.Battlefield) > mine {
			return true
		}
	}
	return false
}

func claimJumperFindPlains(library []*gameengine.Card) *gameengine.Card {
	for _, c := range library {
		if c == nil {
			continue
		}
		if cardHasSubtype(c, "plains") {
			return c
		}
	}
	return nil
}
