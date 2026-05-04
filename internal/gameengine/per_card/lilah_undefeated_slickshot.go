package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerLilahUndefeatedSlickshot wires Lilah, Undefeated Slickshot.
//
// Oracle text:
//
//	Prowess
//	Whenever you cast a multicolored instant or sorcery spell from your
//	hand, exile that spell instead of putting it into your graveyard as
//	it resolves. If you do, it becomes plotted. (You may cast it as a
//	sorcery on a later turn without paying its mana cost.)
//
// Implementation: replacement effect (graveyard → exile after resolve)
// and the "plot" mechanic both require engine plumbing not currently
// available. emitPartial.
func registerLilahUndefeatedSlickshot(r *Registry) {
	r.OnTrigger("Lilah, Undefeated Slickshot", "instant_or_sorcery_cast", lilahInstantOrSorcery)
}

func lilahInstantOrSorcery(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "lilah_plot_replacement"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"multicolor_spell_plot_replacement_unimplemented")
}
