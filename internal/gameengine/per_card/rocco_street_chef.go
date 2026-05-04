package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRoccoStreetChef wires Rocco, Street Chef.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	At the beginning of your end step, each player exiles the top
//	card of their library. Until your next end step, each player may
//	play the card they exiled this way.
//	Whenever a player plays a land from exile or casts a spell from
//	exile, you put a +1/+1 counter on target creature and create a
//	Food token.
//
// Implementation:
//   - "end_step_controller": each seat exiles top card. We mark each
//     exile entry with a flag indicating the exiler so the cast/play
//     trigger can match.
//   - "spell_cast" + "land_entered_battlefield": if cast/play came from
//     the exile zone (ctx["cast_zone"] == "exile" or play_from_exile
//     flag), buff Rocco (+1/+1 on highest-power friendly creature) and
//     mint a Food.
//
// The "Until your next end step, each player may play..." permission
// is gated by the engine's exile machinery; we emitPartial when we
// can't model the temporary play permission cleanly.
func registerRoccoStreetChef(r *Registry) {
	r.OnTrigger("Rocco, Street Chef", "end_step", roccoEndStep)
	r.OnTrigger("Rocco, Street Chef", "spell_cast", roccoExileCast)
	r.OnTrigger("Rocco, Street Chef", "land_entered_battlefield", roccoExileLandPlay)
}

func roccoEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "rocco_end_step_exile_top"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	exiled := 0
	for i, s := range gs.Seats {
		if s == nil || s.Lost || len(s.Library) == 0 {
			continue
		}
		top := s.Library[0]
		gameengine.MoveCard(gs, top, i, "library", "exile", "rocco_end_step_exile")
		exiled++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"count": exiled,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"play_permission_until_next_end_step_not_modeled_as_alternative_cost_window")
}

func roccoExileCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	zone, _ := ctx["cast_zone"].(string)
	if zone != gameengine.ZoneExile {
		return
	}
	roccoBuffAndFood(gs, perm, "spell")
}

func roccoExileLandPlay(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	source, _ := ctx["from_zone"].(string)
	if source != gameengine.ZoneExile {
		return
	}
	roccoBuffAndFood(gs, perm, "land")
}

func roccoBuffAndFood(gs *gameengine.GameState, perm *gameengine.Permanent, source string) {
	const slug = "rocco_play_from_exile_reward"
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	// +1/+1 on highest-power friendly creature.
	var best *gameengine.Permanent
	bestPow := -1
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil || !p.IsCreature() {
			continue
		}
		if p.Power() > bestPow {
			bestPow = p.Power()
			best = p
		}
	}
	if best != nil {
		best.AddCounter("+1/+1", 1)
		gs.InvalidateCharacteristicsCache()
	}
	gameengine.CreateFoodToken(gs, perm.Controller)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"source": source,
		"buffed": cloudPermName(best),
	})
}
