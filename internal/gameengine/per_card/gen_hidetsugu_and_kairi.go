package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerHidetsuguAndKairi wires Hidetsugu and Kairi.
//
// Oracle text:
//
//   Flying
//   When Hidetsugu and Kairi enters, draw three cards, then put two cards
//   from your hand on top of your library in any order.
//   When Hidetsugu and Kairi dies, exile the top card of your library.
//   Target opponent loses life equal to its mana value. If it's an instant
//   or sorcery card, you may cast it without paying its mana cost.
func registerHidetsuguAndKairi(r *Registry) {
	r.OnETB("Hidetsugu and Kairi", hidetsuguAndKairiETB)
	r.OnTrigger("Hidetsugu and Kairi", "dies", hidetsuguAndKairiDies)
}

func hidetsuguAndKairiETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "hidetsugu_and_kairi_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	for i := 0; i < 3; i++ {
		drawOne(gs, seat, perm.Card.DisplayName())
	}
	// Put two cards from hand on top of library. Greedy: pick the two
	// highest-CMC cards in hand. The death trigger exiles the top of
	// library and burns target opponent for that card's mana value, so
	// stacking high-CMC cards on top maximizes damage.
	s := gs.Seats[seat]
	if s == nil {
		return
	}
	moved := []string{}
	for k := 0; k < 2; k++ {
		if len(s.Hand) == 0 {
			break
		}
		bestIdx := -1
		bestCMC := -1
		for i, c := range s.Hand {
			if c == nil {
				continue
			}
			cmc := cardCMC(c)
			if cmc > bestCMC {
				bestCMC = cmc
				bestIdx = i
			}
		}
		if bestIdx < 0 {
			break
		}
		card := s.Hand[bestIdx]
		gameengine.MoveCard(gs, card, seat, "hand", "library_top", "hidetsugu_kairi_topdeck")
		moved = append(moved, card.DisplayName())
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":             seat,
		"drew":             3,
		"top_of_library":   moved,
	})
}

func hidetsuguAndKairiDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "hidetsugu_and_kairi_dies"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil || len(s.Library) == 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   seat,
			"reason": "library_empty",
		})
		return
	}
	top := s.Library[0]
	if top == nil {
		return
	}
	gameengine.MoveCard(gs, top, seat, "library", "exile", "hidetsugu_kairi_death")
	cmc := cardCMC(top)

	// Target opponent: pick the highest-life living opponent.
	targetSeat := -1
	bestLife := -1
	for _, oppIdx := range gs.LivingOpponents(seat) {
		os := gs.Seats[oppIdx]
		if os == nil {
			continue
		}
		if os.Life > bestLife {
			bestLife = os.Life
			targetSeat = oppIdx
		}
	}
	if targetSeat >= 0 && cmc > 0 {
		os := gs.Seats[targetSeat]
		os.Life -= cmc
		gs.LogEvent(gameengine.Event{
			Kind:   "life_loss",
			Seat:   seat,
			Target: targetSeat,
			Source: perm.Card.DisplayName(),
			Amount: cmc,
			Details: map[string]interface{}{
				"reason": "hidetsugu_kairi_death_burn",
			},
		})
	}

	freeCast := false
	if cardHasType(top, "instant") || cardHasType(top, "sorcery") {
		// Free-cast shortcut not implemented for instants/sorceries — fallback
		// is to leave the card in exile (RAW), and emit a partial.
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"instant_or_sorcery_free_cast_resolution_shortcut_unimplemented")
	}

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         seat,
		"target_seat":  targetSeat,
		"exiled_card":  top.DisplayName(),
		"life_lost":    cmc,
		"free_cast":    freeCast,
	})
}
