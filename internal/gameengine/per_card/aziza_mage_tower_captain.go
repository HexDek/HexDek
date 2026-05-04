package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAzizaMageTowerCaptain wires Aziza, Mage Tower Captain.
//
// Oracle text:
//
//	Whenever you cast an instant or sorcery spell, you may tap three
//	untapped creatures you control. If you do, copy that spell. You
//	may choose new targets for the copy.
//
// Spell-copy mechanics aren't simulated at this level — emitPartial.
func registerAzizaMageTowerCaptain(r *Registry) {
	r.OnTrigger("Aziza, Mage Tower Captain", "instant_or_sorcery_cast", azizaSpellCopy)
}

func azizaSpellCopy(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "aziza_spell_copy"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"spell_copy_with_tap_three_creatures_not_simulated")
}
