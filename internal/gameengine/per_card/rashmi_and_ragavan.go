package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRashmiAndRagavan wires Rashmi and Ragavan.
//
// Oracle text:
//
//	Whenever you cast your first spell during each of your turns,
//	exile the top card of target opponent's library and create a
//	Treasure token. Then you may cast the exiled card without paying
//	its mana cost if it's a spell with mana value less than the
//	number of artifacts you control. If you don't cast it this way,
//	you may cast it this turn.
//
// We exile the top of an opponent's library and create a Treasure on
// the controller's first own-turn spell. The "cast for free" path is
// flagged as parser-gap (cross-zone cast machinery is non-trivial here).
func registerRashmiAndRagavan(r *Registry) {
	r.OnTrigger("Rashmi and Ragavan", "spell_cast", rashmiRagavanCast)
}

func rashmiRagavanCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "rashmi_ragavan_first_spell"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	if gs.Active != perm.Controller {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	if perm.Flags["rashmi_first_spell_turn"] == gs.Turn {
		return
	}
	perm.Flags["rashmi_first_spell_turn"] = gs.Turn

	var targetOpp int = -1
	for _, opp := range gs.Opponents(perm.Controller) {
		s := gs.Seats[opp]
		if s == nil || s.Lost || len(s.Library) == 0 {
			continue
		}
		targetOpp = opp
		break
	}
	if targetOpp >= 0 {
		s := gs.Seats[targetOpp]
		c := s.Library[0]
		gameengine.MoveCard(gs, c, targetOpp, "library", "exile", "rashmi_exile")
	}
	tok := gameengine.CreateCreatureToken(gs, perm.Controller, "Treasure",
		[]string{"artifact", "treasure"}, 0, 0)
	if tok != nil && tok.Card != nil {
		tok.Card.Types = []string{"token", "artifact", "treasure"}
		tok.Card.BasePower = 0
		tok.Card.BaseToughness = 0
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"target":    targetOpp,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(), "free_cast_of_exiled_card_unimplemented")
}
