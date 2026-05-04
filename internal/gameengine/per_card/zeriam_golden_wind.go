package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerZeriamGoldenWind wires Zeriam, Golden Wind.
//
// Oracle text:
//
//	Flying
//	Whenever a Griffin you control deals combat damage to a player,
//	create a 2/2 white Griffin creature token with flying.
//
// Implementation:
//   - "combat_damage_player": filter to dealer.Controller == perm.Controller
//     and dealer is Griffin. Create a 2/2 white Griffin token with flying.
//   - Flying on Zeriam herself is engine-handled.
func registerZeriamGoldenWind(r *Registry) {
	r.OnTrigger("Zeriam, Golden Wind", "combat_damage_player", zeriamGriffinDamage)
}

func zeriamGriffinDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "zeriam_griffin_token_on_damage"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	dealer, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if dealer == nil {
		dealer, _ = ctx["source_perm"].(*gameengine.Permanent)
	}
	if dealer == nil || dealer.Card == nil {
		return
	}
	if dealer.Controller != perm.Controller {
		return
	}
	if !cardHasType(dealer.Card, "griffin") {
		return
	}
	token := &gameengine.Card{
		Name:          "Griffin Token",
		Owner:         perm.Controller,
		BasePower:     2,
		BaseToughness: 2,
		Types:         []string{"token", "creature", "griffin"},
		Colors:        []string{"W"},
		TypeLine:      "Token Creature — Griffin",
	}
	t := enterBattlefieldWithETB(gs, perm.Controller, token, false)
	if t != nil {
		if t.Flags == nil {
			t.Flags = map[string]int{}
		}
		t.Flags["kw:flying"] = 1
	}
	gs.InvalidateCharacteristicsCache()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"dealer": dealer.Card.DisplayName(),
	})
}
