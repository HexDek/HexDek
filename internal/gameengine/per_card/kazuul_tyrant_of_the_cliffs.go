package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKazuulTyrantOfTheCliffs wires Kazuul, Tyrant of the Cliffs.
//
// Oracle text:
//
//	Whenever a creature an opponent controls attacks, if you're the
//	defending player, create a 3/3 red Ogre creature token unless
//	that creature's controller pays {3}.
//
// Implementation:
//   - attacks trigger from any creature: gate on (attacker is opponent
//     of Kazuul's controller) AND (defending_seat == Kazuul's
//     controller). AI policy: opponents never pay the {3} toll, so
//     always create the 3/3 R Ogre.
func registerKazuulTyrantOfTheCliffs(r *Registry) {
	r.OnTrigger("Kazuul, Tyrant of the Cliffs", "attacks", kazuulAttacks)
}

func kazuulAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kazuul_ogre_token"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	attackerSeat, _ := ctx["seat"].(int)
	if attackerSeat == perm.Controller {
		return
	}
	defenderSeat := -1
	if v, ok := ctx["defender_seat"].(int); ok {
		defenderSeat = v
	} else if v, ok := ctx["defending_seat"].(int); ok {
		defenderSeat = v
	} else if v, ok := ctx["target"].(int); ok {
		defenderSeat = v
	}
	if defenderSeat != perm.Controller {
		return
	}
	token := &gameengine.Card{
		Name:          "Ogre Token",
		Owner:         perm.Controller,
		BasePower:     3,
		BaseToughness: 3,
		Types:         []string{"token", "creature", "ogre"},
		Colors:        []string{"R"},
		TypeLine:      "Token Creature — Ogre",
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"attacker": attackerSeat,
	})
}
