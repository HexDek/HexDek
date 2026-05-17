package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// etb_basic_land_ramp_family.go — generic handler for the
// "When this creature enters, [you may] search your library for a basic
// land card, put it [onto the battlefield tapped|reveal it, put it into
// your hand], then shuffle" family.
//
// Shape (Farhaven Elf, Civic Wayfinder, Pilgrim's Eye, ...):
//
//	When this creature enters, [you may] search your library for a basic
//	land card, put that card [onto the battlefield tapped | into your
//	hand], then shuffle.
//
// One algorithm: walk the library in order, take the first basic land,
// route it through MoveCard("library"→"battlefield_tapped" or "hand")
// so §614 replacements + landfall observers fire, then shuffle once
// regardless of whether anything was found (CR §701.18b — searching a
// hidden zone shuffles even on whiff).
//
// Hat policy on the "you may" rider: always opt in. Hand-filtered ramp
// is monotone upside in every archetype Yggdrasil scores, the only
// downside being an empty-library shuffle which is irrelevant under
// our cost model. Mandatory variants (none in this initial batch) would
// run the same path.
//
// Cards intentionally NOT in this family:
//
//   - Wood Elves — searches for a Forest specifically (subtype filter)
//     and enters untapped. Could be added with a subtype field, but
//     keep the family scope tight; first member would be a separate
//     family if more cards in the same shape land.
//   - Solemn Simulacrum — same ETB shape, but also has a die-draw
//     trigger that needs its own creature_dies hook. Bespoke.
//   - Yavimaya Granger — same ETB shape, but layered with Echo which
//     the engine doesn't bookkeep separately from upkeep cost. Bespoke.
//   - Gatecreeper Vine / District Guide / Scampering Surveyor — search
//     for "basic land OR Gate/Cave" (filter alternation). Pulled into
//     a sibling family only if a third gating-subtype card emerges.
//   - Primeval Herald — fires on ETB AND attack ("enters or attacks").
//     Bespoke or future shared multi-trigger scaffold.
//   - Loam Larva — puts the land on TOP of the library instead of into
//     hand/battlefield. Niche enough to keep bespoke.
//
// Adding a new family member is one row in etbBasicLandRampEntries.

// landRampDestination chooses where the searched basic land ends up.
type landRampDestination int

const (
	// Onto the battlefield tapped (Farhaven Elf, Solemn Simulacrum-style).
	landRampDestBattlefieldTapped landRampDestination = iota
	// Into the controller's hand (Civic Wayfinder, Pilgrim's Eye, Sylvan
	// Ranger, Borderland Ranger).
	landRampDestHand
)

type etbBasicLandRampEntry struct {
	cardName string
	dest     landRampDestination
}

var etbBasicLandRampEntries = []etbBasicLandRampEntry{
	{
		// Farhaven Elf — {2}{G}, 1/1 Elf Druid.
		//   When this creature enters, you may search your library for a
		//   basic land card, put it onto the battlefield tapped, then
		//   shuffle.
		cardName: "Farhaven Elf",
		dest:     landRampDestBattlefieldTapped,
	},
	{
		// Civic Wayfinder — {2}{G}, 2/2 Elf Warrior.
		//   When this creature enters, you may search your library for a
		//   basic land card, reveal it, put it into your hand, then
		//   shuffle.
		cardName: "Civic Wayfinder",
		dest:     landRampDestHand,
	},
	{
		// Borderland Ranger — {2}{G}, 2/2 Human Scout.
		//   When this creature enters, you may search your library for a
		//   basic land card, reveal it, put it into your hand, then
		//   shuffle.
		cardName: "Borderland Ranger",
		dest:     landRampDestHand,
	},
	{
		// Sylvan Ranger — {1}{G}, 1/1 Elf Scout.
		//   When this creature enters, you may search your library for a
		//   basic land card, reveal it, put it into your hand, then
		//   shuffle.
		cardName: "Sylvan Ranger",
		dest:     landRampDestHand,
	},
	{
		// Pilgrim's Eye — {3}, 1/1 Thopter (artifact creature) with flying.
		//   When this creature enters, you may search your library for a
		//   basic land card, reveal it, put it into your hand, then
		//   shuffle.
		cardName: "Pilgrim's Eye",
		dest:     landRampDestHand,
	},
}

func registerEtbBasicLandRampFamily(r *Registry) {
	for _, e := range etbBasicLandRampEntries {
		e := e
		r.OnETB(e.cardName, func(gs *gameengine.GameState, perm *gameengine.Permanent) {
			runEtbBasicLandRamp(gs, perm, e)
		})
	}
}

func runEtbBasicLandRamp(gs *gameengine.GameState, perm *gameengine.Permanent, e etbBasicLandRampEntry) {
	slug := "etb_basic_land_ramp_family:" + landFetchSlug(e.cardName)
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil || s.Lost {
		return
	}

	land := pickFirstBasicLand(s.Library)
	if land == nil {
		// CR §701.18b: an unsuccessful search still shuffles.
		shuffleLibraryPerCard(gs, seat)
		emitFail(gs, slug, perm.Card.DisplayName(), "no_basic_land_in_library", map[string]interface{}{
			"seat": seat,
		})
		return
	}

	dest := "hand"
	if e.dest == landRampDestBattlefieldTapped {
		dest = "battlefield_tapped"
	}
	gameengine.MoveCard(gs, land, seat, "library", dest, slug+"_search")
	shuffleLibraryPerCard(gs, seat)

	gs.LogEvent(gameengine.Event{
		Kind:   "search_library",
		Seat:   seat,
		Source: perm.Card.DisplayName(),
		Details: map[string]interface{}{
			"found":  []string{land.DisplayName()},
			"reason": slug,
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  seat,
		"land":  land.DisplayName(),
		"dest":  dest,
	})
}

// pickFirstBasicLand returns the first basic land in the library, or nil.
// Walks in library order so the choice is deterministic.
func pickFirstBasicLand(library []*gameengine.Card) *gameengine.Card {
	for _, c := range library {
		if c == nil {
			continue
		}
		if isBasicLand(c) {
			return c
		}
	}
	return nil
}
