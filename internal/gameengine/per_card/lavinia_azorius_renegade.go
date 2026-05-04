package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerLaviniaAzoriusRenegade wires Lavinia, Azorius Renegade.
//
// Oracle text:
//
//	Each opponent can't cast noncreature spells with mana value greater
//	than the number of lands that player controls.
//	Whenever an opponent casts a spell, if no mana was spent to cast it,
//	counter that spell.
//
// Implementation: both clauses are static replacement / cost-checking
// effects that the engine doesn't yet plumb through. emitPartial.
func registerLaviniaAzoriusRenegade(r *Registry) {
	r.OnTrigger("Lavinia, Azorius Renegade", "spell_cast_by_opponent", laviniaSpellCast)
}

func laviniaSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "lavinia_free_spell_counter"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"counter_free_cast_and_mv_above_lands_restriction_unimplemented")
}
