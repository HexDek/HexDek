package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerNamorSubMariner wires Namor the Sub-Mariner.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Flying
//	Namor's power is equal to the number of Merfolk you control.
//	Whenever you cast a noncreature spell with one or more blue mana
//	  symbols in its mana cost, create that many 1/1 blue Merfolk
//	  creature tokens.
//
// Implementation:
//   - "noncreature_spell_cast": gate on caster_seat == perm.Controller.
//     Count {U} pips in cast card's mana cost. Mint that many 1/1 U
//     Merfolk tokens.
//   - PT-equals-merfolk-count is a continuous-effect characteristic —
//     emitPartial on ETB.
//   - Flying handled by AST keyword pipeline.
func registerNamorSubMariner(r *Registry) {
	r.OnETB("Namor the Sub-Mariner", namorETB)
	r.OnTrigger("Namor the Sub-Mariner", "noncreature_spell_cast", namorNoncreatureSpellCast)
}

func namorETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "namor_pt_static", perm.Card.DisplayName(),
		"power_equals_merfolk_count_continuous_static_not_modeled")
}

func namorNoncreatureSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "namor_blue_pip_merfolk"
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
	pips := countPipsOnCard(card, map[string]bool{"U": true})
	if pips <= 0 {
		return
	}
	for i := 0; i < pips; i++ {
		token := &gameengine.Card{
			Name:          "Merfolk Token (Namor)",
			Owner:         perm.Controller,
			BasePower:     1,
			BaseToughness: 1,
			Types:         []string{"token", "creature", "merfolk"},
			Colors:        []string{"U"},
			TypeLine:      "Token Creature — Merfolk",
		}
		enterBattlefieldWithETB(gs, perm.Controller, token, false)
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"pips":   pips,
		"spell":  card.DisplayName(),
		"tokens": pips,
	})
}
