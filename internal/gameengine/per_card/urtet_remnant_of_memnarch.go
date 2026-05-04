package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerUrtetRemnantOfMemnarch wires Urtet, Remnant of Memnarch.
//
// Oracle text:
//
//	{3}
//	Legendary Artifact Creature — Myr
//	Whenever you cast a Myr spell, create a 1/1 colorless Myr artifact
//	  creature token.
//	At the beginning of combat on your turn, untap each Myr you control.
//	{W}{U}{B}{R}{G}, {T}: Put three +1/+1 counters on each Myr you control.
//	  Activate only during your turn.
//
// Implementation:
//   - "spell_cast" trigger gated to caster == controller and spell card
//     has type "myr": create a 1/1 colorless Myr artifact creature token.
//   - "begin_combat_controller" trigger gated to active_seat ==
//     controller: untap each Myr the controller controls.
//   - Activated WUBRG ability: emitPartial.
func registerUrtetRemnantOfMemnarch(r *Registry) {
	r.OnTrigger("Urtet, Remnant of Memnarch", "spell_cast", urtetMyrCast)
	r.OnTrigger("Urtet, Remnant of Memnarch", "begin_combat_controller", urtetBeginCombat)
	r.OnETB("Urtet, Remnant of Memnarch", urtetETB)
}

func urtetETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "urtet_etb", perm.Card.DisplayName(),
		"activated_wubrg_tap_three_plus1plus1_counters_on_each_myr_partial")
}

func urtetMyrCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "urtet_myr_cast_token"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	caster, _ := ctx["caster_seat"].(int)
	if caster != perm.Controller {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil {
		return
	}
	if !cardHasType(card, "myr") {
		return
	}
	token := &gameengine.Card{
		Name:          "Myr Token",
		Owner:         perm.Controller,
		BasePower:     1,
		BaseToughness: 1,
		Types:         []string{"token", "artifact", "creature", "myr"},
		Colors:        []string{},
		TypeLine:      "Token Artifact Creature — Myr",
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}

func urtetBeginCombat(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "urtet_begin_combat_untap_myr"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	untapped := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if !cardHasType(p.Card, "myr") {
			continue
		}
		if p.Tapped {
			p.Tapped = false
			untapped++
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"untapped": untapped,
	})
}
