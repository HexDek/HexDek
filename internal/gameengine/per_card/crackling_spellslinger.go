package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCracklingSpellslinger wires Crackling Spellslinger
// (Muninn parser-gap #30, 34,645 hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{3}{R}{R}
//	Creature — Human Wizard
//	Flash
//	When this creature enters, if you cast it, the next instant or
//	sorcery spell you cast this turn has storm. (When you cast that
//	spell, copy it for each spell cast before it this turn. You may
//	choose new targets for the copies.)
//
// Implementation:
//   - OnETB gated on was_cast: mark perm.Flags["crackling_armed_turn"] =
//     gs.Turn + 1 so the next instant/sorcery the controller casts gets
//     storm.
//   - OnTrigger("spell_cast"): if the caster is our controller, the
//     spell is an instant or sorcery, and the armed flag matches the
//     current turn, push storm copies onto the stack via
//     gameengine.ApplyStormCopies (the same path the printed Storm
//     keyword uses). Disarm by clearing the flag so only the FIRST
//     post-ETB instant/sorcery gets the effect.
//   - "If you cast it" guard mirrors the gating cyclone_summoner.go uses.
func registerCracklingSpellslinger(r *Registry) {
	r.OnETB("Crackling Spellslinger", cracklingSpellslingerETB)
	r.OnTrigger("Crackling Spellslinger", "spell_cast", cracklingSpellslingerSpellCast)
}

func cracklingSpellslingerETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "crackling_spellslinger_arm"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	if perm.Flags["was_cast"] == 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"armed":  false,
			"reason": "not_cast",
		})
		return
	}
	perm.Flags["crackling_armed_turn"] = gs.Turn + 1
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"armed": true,
		"turn":  gs.Turn,
	})
}

func cracklingSpellslingerSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "crackling_spellslinger_storm_grant"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
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
	if card == perm.Card {
		return
	}
	if !cardHasType(card, "instant") && !cardHasType(card, "sorcery") {
		return
	}
	if perm.Flags == nil || perm.Flags["crackling_armed_turn"] != gs.Turn+1 {
		return
	}
	// Find the spell's StackItem and push storm copies.
	var stackItem *gameengine.StackItem
	for i := len(gs.Stack) - 1; i >= 0; i-- {
		si := gs.Stack[i]
		if si == nil || si.Card != card {
			continue
		}
		stackItem = si
		break
	}
	if stackItem == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "spell_not_on_stack", map[string]interface{}{
			"spell": card.DisplayName(),
		})
		return
	}
	delete(perm.Flags, "crackling_armed_turn")

	copies := gameengine.ApplyStormCopies(gs, stackItem, perm.Controller)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"spell":  card.DisplayName(),
		"copies": copies,
	})
}
