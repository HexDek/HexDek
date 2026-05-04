package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTheCouncilOfFour wires The Council of Four.
//
// Oracle text:
//
//	{3}{W}{U}
//	Legendary Creature — Human Noble
//	Whenever a player draws their second card during their turn, you
//	  draw a card.
//	Whenever a player casts their second spell during their turn, you
//	  create a 2/2 white Knight creature token.
//
// Implementation:
//   - "second_draw_this_turn" trigger (canonical): controller draws.
//   - "second_spell_this_turn" trigger (canonical): controller creates a
//     2/2 white Knight token.
//   - These canonical events are the same names emitted by the storm /
//     wheel pipeline (see cast_counts.go). Per-turn nth-draw tracking is
//     done by the engine's draw counters; for "second" we gate by
//     ctx["nth"] == 2 if the engine emits the generic "draw" event.
func registerTheCouncilOfFour(r *Registry) {
	r.OnTrigger("The Council of Four", "draw", theCouncilOfFourOnDraw)
	r.OnTrigger("The Council of Four", "spell_cast", theCouncilOfFourOnCast)
}

func theCouncilOfFourOnDraw(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "the_council_of_four_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	nth, _ := ctx["nth_this_turn"].(int)
	if nth != 2 {
		return
	}
	drawerSeat, _ := ctx["seat"].(int)
	// Must be drawer's turn (oracle: "during their turn").
	if drawerSeat != gs.Active {
		return
	}
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":        perm.Controller,
		"drawer_seat": drawerSeat,
	})
}

func theCouncilOfFourOnCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "the_council_of_four_cast_knight"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	caster, _ := ctx["caster_seat"].(int)
	if caster != gs.Active {
		return
	}
	nth, _ := ctx["casts_this_turn"].(int)
	if nth != 2 {
		return
	}
	token := &gameengine.Card{
		Name:          "Knight Token",
		Owner:         perm.Controller,
		BasePower:     2,
		BaseToughness: 2,
		Types:         []string{"token", "creature", "knight"},
		Colors:        []string{"W"},
		TypeLine:      "Token Creature — Knight",
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":        perm.Controller,
		"caster_seat": caster,
	})
}
