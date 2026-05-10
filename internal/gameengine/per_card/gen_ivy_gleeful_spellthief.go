package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerIvyGleefulSpellthief wires Ivy, Gleeful Spellthief.
//
// Oracle text:
//
//	Flying
//	Whenever a player casts a spell that targets only a single
//	creature other than Ivy, you may copy that spell. The copy
//	targets Ivy. (A copy of an Aura spell becomes a token.)
//
// Implementation:
//   - Flying: handled by the AST keyword pipeline.
//   - "spell_cast" trigger: inspect the cast spell's target list. We
//     count creature targets and require exactly one creature target,
//     not Ivy herself, and no non-creature targets. The "may copy /
//     copy targets Ivy" routing depends on the engine's spell-copy +
//     retarget pipeline, which doesn't yet expose to per_card; we
//     emit the trigger fire so observers see Ivy reacted.
//
// emitPartial: spell-copy-with-retarget is engine-side TODO.
func registerIvyGleefulSpellthief(r *Registry) {
	r.OnTrigger("Ivy, Gleeful Spellthief", "spell_cast", ivySpellCast)
}

func ivySpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ivy_copy_single_creature_target"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	// "any player" — no caster_seat filter.
	targets, _ := ctx["targets"].([]interface{})
	if len(targets) == 0 {
		// Try alternate keys some emitters use.
		if tperm, ok := ctx["target_perm"].(*gameengine.Permanent); ok && tperm != nil {
			targets = []interface{}{tperm}
		}
	}
	if len(targets) != 1 {
		return
	}
	tperm, ok := targets[0].(*gameengine.Permanent)
	if !ok || tperm == nil || !tperm.IsCreature() {
		return
	}
	if tperm == perm {
		return // can't be Ivy herself
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"original_target": tperm.Card.DisplayName(),
		"new_target":      perm.Card.DisplayName(),
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"spell_copy_and_retarget_pipeline_not_wired_for_per_card")
}
