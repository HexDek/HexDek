package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMuddleEverChanging wires Muddle, the Ever-Changing.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Whenever you cast an instant or sorcery spell, Muddle becomes a
//	  copy of up to one target nonlegendary creature you control until
//	  end of turn, except it has myriad.
//
// Copy-becoming requires the layers pipeline. Register the trigger and
// emitPartial — the actual copy-of effect can't be applied at the
// per-card layer alone.
func registerMuddleEverChanging(r *Registry) {
	r.OnTrigger("Muddle, the Ever-Changing", "instant_or_sorcery_cast", muddleSpellCast)
}

func muddleSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "muddle_copy_creature_ueot"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	caster, _ := ctx["caster_seat"].(int)
	if caster != perm.Controller {
		return
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"copy_of_target_nonlegendary_with_myriad_ueot_not_modeled")
}
