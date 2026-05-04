package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerJinGitaxiasProgressTyrant wires Jin-Gitaxias, Progress Tyrant.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Whenever you cast an artifact, instant, or sorcery spell, copy that
//	  spell. You may choose new targets for the copy. This ability
//	  triggers only once each turn. (A copy of a permanent spell becomes
//	  a token.)
//	Whenever an opponent casts an artifact, instant, or sorcery spell,
//	  counter that spell. This ability triggers only once each turn.
//
// Implementation:
//   - "spell_cast": filter to artifact/instant/sorcery types; route
//     copy-or-counter based on caster identity. Each branch is gated on
//     a turn-keyed flag (jin_copy_turn / jin_counter_turn) for the
//     "only once each turn" clause.
//   - Copy creation and counter resolution depend on stack-aware
//     pipelines that aren't fully exposed at the per-card layer; we
//     emitPartial for actual stack manipulation but record the intent.
func registerJinGitaxiasProgressTyrant(r *Registry) {
	r.OnTrigger("Jin-Gitaxias, Progress Tyrant", "spell_cast", jinGitaxiasOnSpellCast)
}

func jinGitaxiasOnSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "jin_gitaxias_progress_tyrant"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	caster, _ := ctx["caster_seat"].(int)
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil {
		return
	}
	if !cardHasType(card, "artifact") && !cardHasType(card, "instant") && !cardHasType(card, "sorcery") {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	if caster == perm.Controller {
		if perm.Flags["jin_copy_turn"] == gs.Turn {
			return
		}
		perm.Flags["jin_copy_turn"] = gs.Turn
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"branch": "copy",
			"spell":  card.DisplayName(),
		})
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"copy_spell_creation_not_modeled")
		return
	}
	// Opponent cast: counter.
	if perm.Flags["jin_counter_turn"] == gs.Turn {
		return
	}
	perm.Flags["jin_counter_turn"] = gs.Turn
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":        perm.Controller,
		"branch":      "counter",
		"spell":       card.DisplayName(),
		"caster_seat": caster,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"counter_spell_resolution_not_modeled")
}
