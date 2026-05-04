package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGyrudaDoomOfDepths wires Gyruda, Doom of Depths.
//
// Oracle text:
//
//   Companion — Your starting deck contains only cards with even mana values.
//   When Gyruda enters, each player mills four cards. Put a creature card
//   with an even mana value from among the milled cards onto the battlefield
//   under your control.
//
// Companion deckbuilding restriction is enforced (or not) at deck construction
// time, not at runtime — nothing to do for that clause here.
func registerGyrudaDoomOfDepths(r *Registry) {
	r.OnETB("Gyruda, Doom of Depths", gyrudaDoomOfDepthsETB)
}

func gyrudaDoomOfDepthsETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "gyruda_doom_of_depths_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}

	// Mill 4 from each player. Track even-CMC creatures milled across
	// all players so Gyruda's controller can pick one to reanimate.
	type milledEntry struct {
		card     *gameengine.Card
		fromSeat int
		cmc      int
	}
	var evenCreatures []milledEntry
	for i, s := range gs.Seats {
		if s == nil {
			continue
		}
		for j := 0; j < 4; j++ {
			if len(s.Library) == 0 {
				break
			}
			top := s.Library[0]
			gameengine.MoveCard(gs, top, i, "library", "graveyard", "gyruda_mill")
			gs.LogEvent(gameengine.Event{
				Kind:   "mill",
				Seat:   seat,
				Target: i,
				Source: perm.Card.DisplayName(),
				Amount: 1,
			})
			cmc := cardCMC(top)
			if cardHasType(top, "creature") && cmc%2 == 0 {
				evenCreatures = append(evenCreatures, milledEntry{card: top, fromSeat: i, cmc: cmc})
			}
		}
	}

	if len(evenCreatures) == 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":         seat,
			"reanimated":   "",
			"milled_total": 16,
		})
		return
	}

	// Greedy: pick the highest-CMC even creature (best ETB body).
	bestIdx := 0
	for i := 1; i < len(evenCreatures); i++ {
		if evenCreatures[i].cmc > evenCreatures[bestIdx].cmc {
			bestIdx = i
		}
	}
	chosen := evenCreatures[bestIdx]

	// Pull from owner's graveyard and ETB under Gyruda's controller.
	owner := gs.Seats[chosen.fromSeat]
	if owner != nil {
		for i, c := range owner.Graveyard {
			if c == chosen.card {
				owner.Graveyard = append(owner.Graveyard[:i], owner.Graveyard[i+1:]...)
				break
			}
		}
	}
	enterBattlefieldWithETB(gs, seat, chosen.card, false)

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":            seat,
		"reanimated":     chosen.card.DisplayName(),
		"reanimated_cmc": chosen.cmc,
		"from_seat":      chosen.fromSeat,
	})
}
