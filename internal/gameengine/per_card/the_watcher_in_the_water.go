package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTheWatcherInTheWater wires The Watcher in the Water.
//
// Oracle text:
//
//	{3}{U}{U}
//	Legendary Creature — Kraken
//	The Watcher in the Water enters tapped with nine stun counters on it.
//	  (If a permanent with a stun counter would become untapped, remove
//	   one from it instead.)
//	Whenever you draw a card during an opponent's turn, create a 1/1 blue
//	  Tentacle creature token.
//	Whenever a Tentacle you control dies, untap up to one target Kraken
//	  and put a stun counter on up to one target nonland permanent.
//
// Implementation:
//   - ETB: tap and add 9 stun counters.
//   - "draw" trigger gated to drawer == controller AND active turn !=
//     controller: create 1/1 blue Tentacle token.
//   - "creature_dies" trigger gated to dying permanent's controller ==
//     The Watcher's controller and dying card has type "tentacle":
//     untap leftmost own Kraken; add stun counter to leftmost opponent
//     nonland permanent.
func registerTheWatcherInTheWater(r *Registry) {
	r.OnETB("The Watcher in the Water", theWatcherETB)
	r.OnTrigger("The Watcher in the Water", "draw", theWatcherOnDraw)
	r.OnTrigger("The Watcher in the Water", "creature_dies", theWatcherTentacleDies)
}

func theWatcherETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	perm.Tapped = true
	if perm.Counters == nil {
		perm.Counters = map[string]int{}
	}
	perm.Counters["stun"] += 9
	emit(gs, "the_watcher_etb", perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
		"stun": 9,
	})
}

func theWatcherOnDraw(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "the_watcher_tentacle_token"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	drawerSeat, _ := ctx["seat"].(int)
	if drawerSeat != perm.Controller {
		return
	}
	if gs.Active == perm.Controller {
		return
	}
	token := &gameengine.Card{
		Name:          "Tentacle Token",
		Owner:         perm.Controller,
		BasePower:     1,
		BaseToughness: 1,
		Types:         []string{"token", "creature", "tentacle"},
		Colors:        []string{"U"},
		TypeLine:      "Token Creature — Tentacle",
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}

func theWatcherTentacleDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "the_watcher_tentacle_died"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	ctrlSeat, _ := ctx["controller_seat"].(int)
	if ctrlSeat != perm.Controller {
		return
	}
	dying, _ := ctx["card"].(*gameengine.Card)
	if dying == nil || !cardHasType(dying, "tentacle") {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	// Untap one own Kraken (any tapped one).
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if p.Tapped && cardHasType(p.Card, "kraken") {
			p.Tapped = false
			break
		}
	}
	// Stun counter on first opponent nonland permanent.
	for _, opp := range gs.Opponents(perm.Controller) {
		s := gs.Seats[opp]
		if s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil {
				continue
			}
			if p.IsLand() {
				continue
			}
			if p.Counters == nil {
				p.Counters = map[string]int{}
			}
			p.Counters["stun"]++
			emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
				"seat":         perm.Controller,
				"stunned":      p.Card.DisplayName(),
				"stunned_seat": opp,
			})
			return
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"reason": "no_opp_nonland_perm_to_stun",
	})
}
