package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMoxDiamond wires up Mox Diamond.
//
// Oracle text:
//
//   If Mox Diamond would enter the battlefield, you may discard a land
//   card. If you do, put Mox Diamond onto the battlefield. If you don't,
//   put it into its owner's graveyard.
//   {T}: Add one mana of any color.
//
// The replacement effect requires discarding a land as a cost to keep it
// on the battlefield. If no land is available, Mox Diamond goes to the
// graveyard instead.

func registerMoxDiamond(r *Registry) {
	r.OnETB("Mox Diamond", moxDiamondETB)
}

func moxDiamondETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "mox_diamond_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]

	landIdx := -1
	for i, c := range s.Hand {
		if c == nil {
			continue
		}
		for _, t := range c.Types {
			if strings.ToLower(t) == "land" {
				landIdx = i
				break
			}
		}
		if landIdx >= 0 {
			break
		}
	}

	if landIdx < 0 {
		gameengine.SacrificePermanent(gs, perm, "mox_diamond_no_land")
		emit(gs, slug, "Mox Diamond", map[string]interface{}{
			"seat":      seat,
			"discarded": false,
			"reason":    "no_land_in_hand",
		})
		return
	}

	card := s.Hand[landIdx]
	gameengine.MoveCard(gs, card, seat, "hand", "graveyard", "discard")

	gs.LogEvent(gameengine.Event{
		Kind:   "discard",
		Seat:   seat,
		Source: "Mox Diamond",
		Details: map[string]interface{}{
			"discarded_card": card.DisplayName(),
			"reason":         "mox_diamond_cost",
		},
	})
	emit(gs, slug, "Mox Diamond", map[string]interface{}{
		"seat":      seat,
		"discarded": true,
		"card":      card.DisplayName(),
	})
}
