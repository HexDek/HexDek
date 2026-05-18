package per_card

import (
	"fmt"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTannukMemorialEnsign wires Tannuk, Memorial Ensign.
//
// Oracle text (Scryfall, verified):
//
//	Landfall — Whenever a land you control enters, Tannuk deals 1
//	damage to each opponent. If this is the second time this ability
//	has resolved this turn, draw a card.
//
// Implementation (R36 stub port):
//   - OnTrigger("permanent_etb"): standard landfall pattern (mirror
//     of Roil Elemental / Strider). Gate on entered perm being a land
//     controlled by perm.Controller.
//   - Per-turn resolution counter on perm.Flags keyed by gs.Turn+1
//     (same offset convention as Sand Scout / Ashling).
//   - Every fire: 1 damage to each living opponent.
//   - On 2nd fire only: draw a card.
func registerTannukMemorialEnsign(r *Registry) {
	r.OnTrigger("Tannuk, Memorial Ensign", "permanent_etb", tannukLandfall)
}

func tannukLandfall(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "tannuk_memorial_ensign_landfall"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	entered, _ := ctx["perm"].(*gameengine.Permanent)
	if entered == nil || entered.Card == nil || !entered.IsLand() {
		return
	}
	if entered.Controller != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}

	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	turnKey := tannukTurnKey(gs.Turn)
	perm.Flags[turnKey]++
	resolutions := perm.Flags[turnKey]

	// 1 damage to each living opponent every fire.
	hitOpps := 0
	for _, opp := range gs.Opponents(perm.Controller) {
		s := gs.Seats[opp]
		if s == nil || s.Lost {
			continue
		}
		gameengine.DealDamage(gs, opp, 1, perm.Card.DisplayName())
		hitOpps++
	}

	// 2nd resolution this turn: draw a card.
	drawnName := ""
	if resolutions == 2 {
		drawn := drawOne(gs, perm.Controller, perm.Card.DisplayName())
		if drawn != nil {
			drawnName = drawn.DisplayName()
		}
	}

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         perm.Controller,
		"resolutions":  resolutions,
		"hit_opponents": hitOpps,
		"landfall_from": entered.Card.DisplayName(),
		"drawn":        drawnName,
	})
}

func tannukTurnKey(turn int) string {
	return fmt.Sprintf("tannuk_resolutions_t%d", turn+1)
}
