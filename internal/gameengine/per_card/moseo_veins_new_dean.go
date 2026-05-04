package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMoseoVeinsNewDean wires Moseo, Vein's New Dean.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Flying
//	When Moseo enters, create a 1/1 black and green Pest creature token
//	  with "Whenever this token attacks, you gain 1 life."
//	Infusion — At the beginning of your end step, if you gained life
//	  this turn, return up to one target creature card with mana value
//	  X or less from your graveyard to the battlefield, where X is the
//	  amount of life you gained this turn.
//
// Implementation:
//   - OnETB: mint a 1/1 B/G Pest token. The token's attack-to-gain-1
//     trigger isn't tracked separately (creature_attacks for tokens
//     without per-card handlers won't fire), so we tag the token with a
//     Card.Types entry "moseo_pest" used as a marker only.
//   - "end_step_controller": if life_gained_this_turn > 0, find the
//     highest-CMC creature in our graveyard with CMC <= X, return it.
//     Tracking life_gained_this_turn requires accumulating life-gain
//     events on perm.Flags via "life_gained" trigger.
//   - Flying handled by AST keyword pipeline.
func registerMoseoVeinsNewDean(r *Registry) {
	r.OnETB("Moseo, Vein's New Dean", moseoETB)
	r.OnTrigger("Moseo, Vein's New Dean", "life_gained", moseoLifeGained)
	r.OnTrigger("Moseo, Vein's New Dean", "end_step", moseoEndStep)
}

func moseoETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "moseo_pest_token"
	if gs == nil || perm == nil {
		return
	}
	token := &gameengine.Card{
		Name:          "Pest Token (Moseo)",
		Owner:         perm.Controller,
		BasePower:     1,
		BaseToughness: 1,
		Types:         []string{"token", "creature", "pest", "moseo_pest"},
		Colors:        []string{"B", "G"},
		TypeLine:      "Token Creature — Pest",
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"pest_token_attack_to_gain_1_life_trigger_not_modeled")
}

func moseoLifeGained(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	gainSeat, _ := ctx["seat"].(int)
	if gainSeat != perm.Controller {
		return
	}
	amount, _ := ctx["amount"].(int)
	if amount <= 0 {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	turnKey := "moseo_lg_turn"
	totalKey := "moseo_lg_total"
	if perm.Flags[turnKey] != gs.Turn {
		perm.Flags[turnKey] = gs.Turn
		perm.Flags[totalKey] = 0
	}
	perm.Flags[totalKey] += amount
}

func moseoEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "moseo_infusion_reanimate"
	if gs == nil || perm == nil {
		return
	}
	if perm.Flags == nil {
		return
	}
	if perm.Flags["moseo_lg_turn"] != gs.Turn {
		return
	}
	x := perm.Flags["moseo_lg_total"]
	if x <= 0 {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	var pick *gameengine.Card
	bestCMC := -1
	for _, c := range seat.Graveyard {
		if c == nil || !cardHasType(c, "creature") {
			continue
		}
		cm := gameengine.ManaCostOf(c)
		if cm > x {
			continue
		}
		if cm > bestCMC {
			bestCMC = cm
			pick = c
		}
	}
	if pick == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_eligible_creature", map[string]interface{}{
			"seat": perm.Controller,
			"x":    x,
		})
		return
	}
	gameengine.MoveCard(gs, pick, perm.Controller, "graveyard", "battlefield", "moseo_infusion")
	createPermanent(gs, perm.Controller, pick, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"x":          x,
		"reanimated": pick.DisplayName(),
	})
}
