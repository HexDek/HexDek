package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerThrunBreakerOfSilence wires Thrun, Breaker of Silence.
//
// Oracle text:
//
//   This spell can't be countered.
//   Trample
//   Thrun can't be the target of nongreen spells your opponents
//   control or abilities from nongreen sources your opponents control.
//   During your turn, Thrun has indestructible.
//
// R37 port:
//
//   - "This spell can't be countered": PORTED via OnCast hook. The
//     counter-spell resolver (counter_resolve.go's spellCannotBeCountered)
//     reads StackItem.CostMeta["cannot_be_countered"]=true and refuses
//     to counter the spell. We stamp that flag when Thrun is being
//     cast — at the cast hook the StackItem is already on the stack
//     (PushStackItem ran before InvokeCastHook fires it).
//   - Trample: AST keyword pipeline.
//   - "Can't be the target of nongreen spells/abilities from opponents":
//     a per-permanent shroud-style filter; the engine's target-legality
//     check supports color-based protection via AST keywords. Flagged
//     in emitPartial since the per_card layer doesn't independently
//     register this protection.
//   - "During your turn, Thrun has indestructible": conditional static
//     based on active player. Engine layer 6 / SBA path doesn't have
//     a per-permanent "indestructible while controller is active" hook
//     today; flagged.
func registerThrunBreakerOfSilence(r *Registry) {
	r.OnCast("Thrun, Breaker of Silence", thrunCastUncounterable)
	r.OnETB("Thrun, Breaker of Silence", thrunETBPartial)
}

func thrunCastUncounterable(gs *gameengine.GameState, item *gameengine.StackItem) {
	const slug = "thrun_cant_be_countered"
	if gs == nil || item == nil {
		return
	}
	if item.CostMeta == nil {
		item.CostMeta = map[string]interface{}{}
	}
	item.CostMeta["cannot_be_countered"] = true
	name := ""
	if item.Card != nil {
		name = item.Card.DisplayName()
	}
	emit(gs, slug, name, map[string]interface{}{
		"seat":          item.Controller,
		"stack_flagged": true,
	})
}

func thrunETBPartial(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "thrun_etb"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"target-protection vs nongreen opponents + active-turn indestructible not yet ported; cast-can't-be-countered handled by cast hook")
}
