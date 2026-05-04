package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSavraQueenOfTheGolgari wires Savra, Queen of the Golgari.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Whenever you sacrifice a black creature, you may pay 2 life. If
//	you do, each other player sacrifices a creature of their choice.
//	Whenever you sacrifice a green creature, you may gain 2 life.
//
// Implementation:
//   - "permanent_sacrificed" trigger: gate on the sacrificed permanent
//     being controlled by Savra's controller and being a creature. If
//     it's black, pay 2 life and force each opponent to sacrifice their
//     lowest-CMC creature. If it's green, gain 2 life. Cards that are
//     both colors trigger both halves.
func registerSavraQueenOfTheGolgari(r *Registry) {
	r.OnTrigger("Savra, Queen of the Golgari", "permanent_sacrificed", savraSacrifice)
}

func savraSacrifice(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "savra_sacrifice_trigger"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	controllerSeat, _ := ctx["controller_seat"].(int)
	if controllerSeat != perm.Controller {
		return
	}
	sacCard, _ := ctx["card"].(*gameengine.Card)
	if sacCard == nil || !cardHasType(sacCard, "creature") {
		return
	}
	isBlack := false
	isGreen := false
	for _, c := range sacCard.Colors {
		switch strings.ToUpper(c) {
		case "B":
			isBlack = true
		case "G":
			isGreen = true
		}
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if isBlack && seat.Life > 2 {
		seat.Life -= 2
		oppSacced := 0
		for _, oppIdx := range gs.Opponents(perm.Controller) {
			os := gs.Seats[oppIdx]
			if os == nil || os.Lost {
				continue
			}
			// Their choice: lowest-CMC creature.
			var pick *gameengine.Permanent
			low := 1 << 30
			for _, p := range os.Battlefield {
				if p == nil || p.Card == nil || !p.IsCreature() {
					continue
				}
				cm := cardCMC(p.Card)
				if cm < low {
					low = cm
					pick = p
				}
			}
			if pick != nil {
				gameengine.SacrificePermanent(gs, pick, "savra_black_trigger")
				oppSacced++
			}
		}
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":         perm.Controller,
			"color":        "B",
			"opp_sacced":   oppSacced,
			"life_paid":    2,
		})
	}
	if isGreen {
		seat.Life += 2
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"color":     "G",
			"life_gain": 2,
		})
	}
}
