package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTalionTheKindlyLord wires Talion, the Kindly Lord.
//
// Oracle text:
//
//	{2}{U}{B}
//	Legendary Creature — Faerie Noble
//	Flying
//	As Talion enters, choose a number between 1 and 10.
//	Whenever an opponent casts a spell with mana value, power, or
//	  toughness equal to the chosen number, that player loses 2 life
//	  and you draw a card.
//
// Implementation:
//   - Flying: AST keyword.
//   - "As enters, choose a number" — replacement-style choice; we pick a
//     fixed default of 3 (the most common Commander curve hit). The
//     chosen number is stashed on perm.Flags["talion_number"] so future
//     ability evaluations can read it. Stored on ETB.
//   - "spell_cast" trigger: gated to caster != controller. Compares the
//     spell card's CMC, BasePower, BaseToughness to the chosen number;
//     any match fires the drain+draw.
func registerTalionTheKindlyLord(r *Registry) {
	r.OnETB("Talion, the Kindly Lord", talionETB)
	r.OnTrigger("Talion, the Kindly Lord", "spell_cast", talionOnSpellCast)
}

func talionETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "talion_choose_number"
	if gs == nil || perm == nil {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	// Fixed pick: 3. Most common cMV hit and a common power/toughness.
	perm.Flags["talion_number"] = 3
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"number": 3,
	})
}

func talionOnSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "talion_drain_and_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	caster, _ := ctx["caster_seat"].(int)
	if caster == perm.Controller {
		return
	}
	if caster < 0 || caster >= len(gs.Seats) {
		return
	}
	chosen := perm.Flags["talion_number"]
	if chosen == 0 {
		chosen = 3
	}

	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil {
		return
	}
	match := false
	if cardCMC(card) == chosen {
		match = true
	} else if cardHasType(card, "creature") && (card.BasePower == chosen || card.BaseToughness == chosen) {
		match = true
	}
	if !match {
		return
	}

	target := gs.Seats[caster]
	if target == nil || target.Lost {
		return
	}
	gameengine.LoseLife(gs, caster, 2, perm.Card.DisplayName())
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":        perm.Controller,
		"caster_seat": caster,
		"chosen":      chosen,
		"spell":       card.DisplayName(),
	})
	_ = gs.CheckEnd()
}
