package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTheSwarmweaver wires The Swarmweaver.
//
// Oracle text (DSK, {2}{B}{G}, 2/3 Legendary Artifact Creature — Scarecrow):
//
//	When The Swarmweaver enters, create two 1/1 black and green Insect
//	creature tokens with flying.
//	Delirium — As long as there are four or more card types among
//	cards in your graveyard, Insects and Spiders you control get +1/+1
//	and have deathtouch.
//
// Implementation:
//   - ETB creates two 1/1 BG Insect tokens with flying (carried as a
//     type tag for the AST keyword pipeline).
//   - The Delirium-gated anthem is a continuous static; the AST engine
//     handles the conditional via its as-long-as layer. emitPartial
//     records the boundary so analytics can audit application.
func registerTheSwarmweaver(r *Registry) {
	r.OnETB("The Swarmweaver", theSwarmweaverETB)
}

func theSwarmweaverETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "the_swarmweaver_etb_insect_tokens"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	for i := 0; i < 2; i++ {
		gameengine.CreateCreatureToken(gs, seat, "Insect Token",
			[]string{"creature", "insect", "pip:B", "pip:G", "flying"}, 1, 1)
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   seat,
		"tokens": 2,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"delirium_static_anthem_handled_by_ast_engine_layers")
}
