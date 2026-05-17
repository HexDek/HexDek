package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerWildPair wires Wild Pair (Muninn parser-gap #51, 15,459 hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{4}{G}{G}
//	Enchantment
//	Whenever a creature enters, if you cast it from your hand, you may
//	search your library for a creature card with the same total power
//	and toughness, put it onto the battlefield, then shuffle.
//
// Implementation:
//   - permanent_etb on Wild Pair's controller. Gate on:
//       (a) entering.IsCreature()
//       (b) entering.Controller == perm.Controller
//       (c) entering.Flags["was_cast"] == 1 AND
//           entering.Flags["cast_from_hand"] == 1
//   - Compute target total = entering.BasePower + entering.BaseToughness.
//     Search the controller's library for the first creature card whose
//     BasePower + BaseToughness equals the target. Skip the entering
//     creature itself (defensive — shouldn't appear in library).
//   - Move the find from library to battlefield via the standard
//     MoveCard("battlefield") path so ETB triggers cascade. Shuffle.
//   - "you may" — Hat policy: accept if any match exists (always upside).
//   - Re-entrancy guard via perm.Flags["wild_pair_in_trigger"] so the
//     fetched creature's own ETB can't recursively invoke Wild Pair
//     (the fetched creature wasn't cast from hand, so it wouldn't gate
//     in anyway, but the guard is cheap and matches the convention in
//     ragost_combo.go's nuka-cola token re-entry).
func registerWildPair(r *Registry) {
	r.OnTrigger("Wild Pair", "permanent_etb", wildPairPermETB)
}

func wildPairPermETB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "wild_pair_creature_pair_search"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	if perm.Flags["wild_pair_in_trigger"] == 1 {
		return
	}
	entering, _ := ctx["perm"].(*gameengine.Permanent)
	if entering == nil {
		entering, _ = ctx["permanent"].(*gameengine.Permanent)
	}
	if entering == nil || entering == perm || entering.Card == nil {
		return
	}
	if !entering.IsCreature() {
		return
	}
	if entering.Controller != perm.Controller {
		return
	}
	if entering.Flags == nil ||
		entering.Flags["was_cast"] != 1 ||
		entering.Flags["cast_from_hand"] != 1 {
		return
	}

	target := int(entering.Card.BasePower) + int(entering.Card.BaseToughness)
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}

	var match *gameengine.Card
	for _, c := range seat.Library {
		if c == nil || c == entering.Card {
			continue
		}
		if !cardHasType(c, "creature") {
			continue
		}
		if int(c.BasePower)+int(c.BaseToughness) == target {
			match = c
			break
		}
	}

	if match == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"entering":  entering.Card.DisplayName(),
			"target_pt": target,
			"found":     false,
		})
		// Still shuffle per "search your library, ..., then shuffle"
		// (CR §701.19c). The shuffle runs even on whiff.
		shuffleLibraryPerCard(gs, perm.Controller)
		return
	}

	perm.Flags["wild_pair_in_trigger"] = 1
	gameengine.MoveCard(gs, match, perm.Controller, "library", "battlefield", "wild_pair_search")
	shuffleLibraryPerCard(gs, perm.Controller)
	perm.Flags["wild_pair_in_trigger"] = 0

	gs.LogEvent(gameengine.Event{
		Kind:   "search_library",
		Seat:   perm.Controller,
		Source: perm.Card.DisplayName(),
		Details: map[string]interface{}{
			"found":  []string{match.DisplayName()},
			"reason": slug,
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"entering":  entering.Card.DisplayName(),
		"target_pt": target,
		"found":     true,
		"into_play": match.DisplayName(),
	})
}
