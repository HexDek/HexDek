package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGeologicalAppraiser wires Geological Appraiser (Muninn parser-gap
// #95, ~4.3K hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{2}{R}
//	Creature — Human Artificer
//	When this creature enters, if you cast it, discover 3. (Exile cards
//	from the top of your library until you exile a nonland card with
//	mana value 3 or less. Cast it without paying its mana cost or put it
//	into your hand. Put the rest on the bottom in a random order.)
//
// Implementation:
//   - OnETB gated on was_cast. PerformDiscover(controller, 3) mirrors
//     Pantlaza's path.
func registerGeologicalAppraiser(r *Registry) {
	r.OnETB("Geological Appraiser", geologicalAppraiserETB)
}

func geologicalAppraiserETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "geological_appraiser_etb_discover_3"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil || perm.Flags["was_cast"] != 1 {
		emitFail(gs, slug, perm.Card.DisplayName(), "not_cast", nil)
		return
	}
	found := gameengine.PerformDiscover(gs, perm.Controller, 3)
	discovered := ""
	if found != nil {
		discovered = found.DisplayName()
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"discover_x": 3,
		"discovered": discovered,
	})
}
