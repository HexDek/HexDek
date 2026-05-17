package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTranscendentDragon wires Transcendent Dragon (Muninn parser-gap
// #43, 20,144 hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{4}{U}{U}
//	Creature — Dragon
//	Flash
//	Flying
//	When this creature enters, if you cast it, counter target spell. If
//	that spell is countered this way, exile it instead of putting it
//	into its owner's graveyard, then you may cast it without paying its
//	mana cost.
//
// Implementation:
//   - Flash + Flying are AST-side.
//   - ETB gate: perm.Flags["was_cast"] == 1 (stack.go stamps cast path).
//   - Counter pick: top-down findCounterableSpell — first legal opponent
//     spell on the stack. Tag CostMeta["exile_on_resolve"] = true and
//     stamp ExiledByTimestamp = perm.Timestamp so the engine's countered-
//     to-exile path routes correctly (Smirking Spelljacker precedent).
//   - "May cast without paying" clause: same engine gap as Smirking
//     Spelljacker — no free-cast-arbitrary-spell path exists, so the
//     candidate is logged via emitPartial. The exile linkage is fully
//     wired so when that path lands the cast will follow without code
//     changes here.
func registerTranscendentDragon(r *Registry) {
	r.OnETB("Transcendent Dragon", transcendentDragonETB)
}

func transcendentDragonETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "transcendent_dragon_etb_counter_exile"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil || perm.Flags["was_cast"] != 1 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "not_cast",
		})
		return
	}
	target := findCounterableSpell(gs, perm.Controller, nil)
	if target == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_opponent_spell_on_stack", nil)
		return
	}
	target.Countered = true
	exiledName := ""
	if target.Card != nil {
		exiledName = target.Card.DisplayName()
		if target.CostMeta == nil {
			target.CostMeta = map[string]interface{}{}
		}
		target.CostMeta["exile_on_resolve"] = true
		target.CostMeta["exiled_by_timestamp"] = perm.Timestamp
		target.Card.ExiledByTimestamp = perm.Timestamp
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":    perm.Controller,
		"countered": exiledName,
		"opp":     target.Controller,
		"link_ts": perm.Timestamp,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"may_cast_exiled_spell_without_paying_mana_no_engine_free_cast_path")
}
