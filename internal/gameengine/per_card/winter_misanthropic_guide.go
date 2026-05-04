package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerWinterMisanthropicGuide wires Winter, Misanthropic Guide.
//
// Oracle text:
//
//	Ward {2}
//	At the beginning of your upkeep, each player draws two cards.
//	Delirium — As long as there are four or more card types among cards
//	in your graveyard, each opponent's maximum hand size is equal to
//	seven minus the number of those card types.
//
// Implementation:
//   - "upkeep_controller" on Winter's controller's upkeep: each player
//     draws 2 cards.
//   - The delirium hand-size static effect is tracked via emitPartial —
//     hand-size enforcement at end of cleanup is engine-side and not
//     configurable from a per-card hook here.
//   - Ward {2} is engine-handled.
func registerWinterMisanthropicGuide(r *Registry) {
	r.OnTrigger("Winter, Misanthropic Guide", "upkeep_controller", winterMisanthropicUpkeep)
}

func winterMisanthropicUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "winter_misanthropic_each_draws_two"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	for i, s := range gs.Seats {
		if s == nil || s.Lost {
			continue
		}
		drawOne(gs, i, perm.Card.DisplayName())
		drawOne(gs, i, perm.Card.DisplayName())
	}
	emitPartial(gs, slug, perm.Card.DisplayName(), "delirium_hand_size_static_not_implemented")
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}
