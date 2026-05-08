package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerZoyowaLavaTongue wires Zoyowa Lava-Tongue.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	{B}{R}
//	Legendary Creature — Goblin Warlock
//	2/2
//	Deathtouch
//	At the beginning of your end step, if you descended this turn, each
//	opponent may discard a card or sacrifice a permanent of their
//	choice. Zoyowa deals 3 damage to each opponent who didn't. (You
//	descended if a permanent card was put into your graveyard from
//	anywhere.)
//
// Implementation:
//   - end_step_controller: check if controller descended this turn by
//     scanning the graveyard for any permanent card whose Owner ==
//     controller. (We don't have a turn-scoped "descended_this_turn"
//     counter — emitPartial flags this; the heuristic over-fires after
//     turn 1 but mostly correctly identifies sac-fueled or Saga-cycling
//     decks.) Each opponent loses 3 life (we model "didn't pay" since
//     opponents in simulation prefer not to give up cards/permanents).
func registerZoyowaLavaTongue(r *Registry) {
	r.OnTrigger("Zoyowa Lava-Tongue", "end_step_controller", zoyowaLavaTongueEnd)
}

func zoyowaLavaTongueEnd(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "zoyowa_lava_tongue_descended_drain"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if !seat.Turn.Descended {
		return
	}
	hits := 0
	for i, opp := range gs.Seats {
		if opp == nil || opp.Lost || i == perm.Controller {
			continue
		}
		gameengine.DealDamage(gs, i, 3, perm.Card.DisplayName())
		hits++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         perm.Controller,
		"opponents_hit": hits,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"opponents_assumed_to_not_pay")
	_ = gs.CheckEnd()
}
