package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTasigurTheGoldenFang wires Tasigur, the Golden Fang.
//
// Oracle text:
//
//   Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)
//   {2}{G/U}{G/U}: Mill two cards, then return a nonland card of an opponent's choice from your graveyard to your hand.
//
// Auto-generated activated ability handler.
func registerTasigurTheGoldenFang(r *Registry) {
	r.OnActivated("Tasigur, the Golden Fang", tasigurTheGoldenFangActivate)
}

func tasigurTheGoldenFangActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "tasigur_the_golden_fang_activate"
	if gs == nil || src == nil {
		return
	}
	seatIdx := src.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seatIdx]
	if s == nil || s.Lost {
		return
	}
	// Cost gate: {2}{G/U}{G/U}. Hybrid pips can be paid with either G or
	// U; we ask the typed mana pool to spend any-color and fall through
	// to the generic pool for the full 4 cost.
	const manaCost = 4
	if !gameengine.PayGenericCost(gs, s, manaCost, "activated", "tasigur_activate", src.Card.DisplayName()) {
		emitFail(gs, slug, src.Card.DisplayName(), "insufficient_mana", map[string]interface{}{
			"mana_pool": s.ManaPool,
			"mana_cost": manaCost,
		})
		return
	}
	for i := 0; i < 2; i++ {
		if len(s.Library) == 0 {
			break
		}
		card := s.Library[0]
		gameengine.MoveCard(gs, card, seatIdx, "library", "graveyard", "mill")
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":      seatIdx,
		"mana_paid": manaCost,
	})
}
