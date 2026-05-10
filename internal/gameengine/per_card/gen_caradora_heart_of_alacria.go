package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCaradoraHeartOfAlacria wires Caradora, Heart of Alacria.
//
// Oracle text:
//
//	When Caradora enters, you may search your library for a Mount or
//	Vehicle card, reveal it, put it into your hand, then shuffle.
//	If one or more +1/+1 counters would be put on a creature or
//	Vehicle you control, that many plus one +1/+1 counters are put on
//	it instead.
//
// Implementation:
//   - ETB tutor: scan library for highest-CMC Mount or Vehicle card,
//     move to hand, shuffle.
//   - +1/+1 replacement: engine-deep counter-placement hook (it has to
//     intercept every AddCounter call). Until that exists, set a
//     per-seat flag the engine can read and emit the partial.
func registerCaradoraHeartOfAlacria(r *Registry) {
	r.OnETB("Caradora, Heart of Alacria", caradoraETBTutorAndReplaceFlag)
}

func caradoraETBTutorAndReplaceFlag(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "caradora_etb_tutor_and_replace"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	var best *gameengine.Card
	bestCMC := -1
	for _, c := range seat.Library {
		if c == nil {
			continue
		}
		if !cardHasSubtype(c, "mount") && !cardHasSubtype(c, "vehicle") &&
			!cardHasType(c, "vehicle") {
			continue
		}
		if cmc := cardCMC(c); cmc > bestCMC {
			best = c
			bestCMC = cmc
		}
	}
	if best != nil {
		gameengine.MoveCard(gs, best, perm.Controller, "library", "hand", "caradora_tutor")
		if len(seat.Library) > 1 && gs.Rng != nil {
			gs.Rng.Shuffle(len(seat.Library), func(i, j int) {
				seat.Library[i], seat.Library[j] = seat.Library[j], seat.Library[i]
			})
		}
	}
	if seat.Flags == nil {
		seat.Flags = map[string]int{}
	}
	seat.Flags["caradora_plus_one_counter_replacement"] = 1
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"tutored":  best != nil,
		"target":   equipmentName(best),
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"+1/+1 counter +1 replacement needs AddCounter hook; flag set for downstream consumers")
}
