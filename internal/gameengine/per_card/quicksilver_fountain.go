package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerQuicksilverFountain wires Quicksilver Fountain (Muninn parser-gap
// #101, 1.9K hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{3}
//	Artifact
//	At the beginning of each player's upkeep, that player puts a flood
//	counter on target non-Island land they control of their choice. That
//	land is an Island for as long as it has a flood counter on it.
//	At the beginning of each end step, if all lands on the battlefield
//	are Islands, remove all flood counters from them.
//
// Implementation (Muninn #101-120 wave):
//   - upkeep trigger fires once per player. The active player picks a
//     non-Island land they control and stamps a "flood" counter on it.
//     Picking-policy: prefer a tapped land (least tempo cost). If none
//     are tapped, pick the first non-Island land. If no non-Island
//     candidates exist (mono-Island board), no-op.
//   - "That land is an Island" is a continuous type-changing effect
//     belonging to the Phase 8 layers pass. We do not yet retype the
//     land in the engine — emitPartial so Muninn can keep tracking the
//     residual gap.
//   - end_step "if all lands are Islands, remove all flood counters":
//     scan the entire battlefield. If every land is already a base
//     Island (Card.Types contains "island" subtype), wipe all flood
//     counters across all seats. Since the type-changing layer isn't
//     modelled, the practical exit-criterion is "all lands had Island
//     subtype to begin with" — which the parser-gap source decks
//     virtually never satisfy, so flood counters accumulate in a way
//     Muninn / Heimdall can audit later.
func registerQuicksilverFountain(r *Registry) {
	r.OnTrigger("Quicksilver Fountain", "upkeep", quicksilverFountainUpkeep)
	r.OnTrigger("Quicksilver Fountain", "end_step", quicksilverFountainEndStep)
}

func quicksilverFountainUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "quicksilver_fountain_flood_counter"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	// "At the beginning of each player's upkeep" — fires for whoever's
	// upkeep it is. The "target … land they control of their choice"
	// belongs to the upkeep-stepping player.
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat < 0 || activeSeat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[activeSeat]
	if s == nil || s.Lost {
		return
	}
	var pickedTapped, pickedAny *gameengine.Permanent
	for _, p := range s.Battlefield {
		if p == nil || p.Card == nil || !p.IsLand() {
			continue
		}
		if cardHasType(p.Card, "island") {
			continue
		}
		if pickedAny == nil {
			pickedAny = p
		}
		if p.Tapped && pickedTapped == nil {
			pickedTapped = p
			break
		}
	}
	pick := pickedTapped
	if pick == nil {
		pick = pickedAny
	}
	if pick == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_non_island_land", map[string]interface{}{
			"seat": activeSeat,
		})
		return
	}
	pick.AddCounter("flood", 1)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         activeSeat,
		"target_land":  pick.Card.DisplayName(),
		"flood_total":  pick.Counters["flood"],
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"flood_counter_island_type_change_pending_layers_pipeline")
}

func quicksilverFountainEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "quicksilver_fountain_flood_clear"
	if gs == nil || perm == nil {
		return
	}
	allIslands := true
	totalLands := 0
	for _, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil || !p.IsLand() {
				continue
			}
			totalLands++
			if !cardHasType(p.Card, "island") {
				allIslands = false
			}
		}
	}
	if totalLands == 0 || !allIslands {
		return
	}
	cleared := 0
	for _, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Counters == nil {
				continue
			}
			if p.Counters["flood"] > 0 {
				cleared += p.Counters["flood"]
				delete(p.Counters, "flood")
			}
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"all_islands":   true,
		"flood_cleared": cleared,
	})
}
