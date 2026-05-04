package per_card

import (
	"math/rand"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKarumonixTheRatKing wires Karumonix, the Rat King.
//
// Oracle text:
//
//   Toxic 1 (Players dealt combat damage by this creature also get a poison counter.)
//   Other Rats you control have toxic 1.
//   When Karumonix enters, look at the top five cards of your library. You may reveal any number of Rat cards from among them and put the revealed cards into your hand. Put the rest on the bottom of your library in a random order.
//
// Implementation:
//   - Toxic 1 on Karumonix: AST keyword on the printed card.
//   - "Other Rats you control have toxic 1": granted by static ability.
//     Implementing as a runtime flag on every Rat we control at ETB and
//     refreshing on a creature_etb tick is more than this stub needs;
//     log the gap and rely on the layer system / AST anointed-keyword
//     pipeline if/when it lands. We mark the partial below.
//   - ETB look-five-reveal-Rats: take the top 5 (or fewer) cards, route
//     every Rat into hand, shuffle the rest and bottom them.
func registerKarumonixTheRatKing(r *Registry) {
	r.OnETB("Karumonix, the Rat King", karumonixTheRatKingETB)
}

func karumonixTheRatKingETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "karumonix_etb_top5_reveal_rats"
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

	n := 5
	if len(s.Library) < n {
		n = len(s.Library)
	}
	if n == 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":     seat,
			"revealed": 0,
		})
		emitPartial(gs, "karumonix_other_rats_toxic_1", perm.Card.DisplayName(),
			"static_grant_toxic_1_to_other_rats_unimplemented")
		return
	}

	revealed := make([]*gameengine.Card, n)
	copy(revealed, s.Library[:n])
	s.Library = s.Library[n:]

	taken := make(map[int]bool)
	var pulled []string
	for i, c := range revealed {
		if c == nil {
			continue
		}
		if !cardHasType(c, "rat") {
			continue
		}
		s.Hand = append(s.Hand, c)
		gs.LogEvent(gameengine.Event{
			Kind:   "draw",
			Seat:   seat,
			Source: perm.Card.DisplayName(),
			Details: map[string]interface{}{
				"slug":   slug,
				"reason": "karumonix_reveal_pick",
				"card":   c.DisplayName(),
			},
		})
		taken[i] = true
		pulled = append(pulled, c.DisplayName())
	}

	var remainder []*gameengine.Card
	for i, c := range revealed {
		if taken[i] || c == nil {
			continue
		}
		remainder = append(remainder, c)
	}
	rand.Shuffle(len(remainder), func(i, j int) {
		remainder[i], remainder[j] = remainder[j], remainder[i]
	})
	s.Library = append(s.Library, remainder...)

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":           seat,
		"revealed":       n,
		"taken_count":    len(pulled),
		"taken_cards":    pulled,
		"bottomed_count": len(remainder),
	})
	emitPartial(gs, "karumonix_other_rats_toxic_1", perm.Card.DisplayName(),
		"static_grant_toxic_1_to_other_rats_unimplemented")
}
