package gameengine

import "strings"

// keywords_conjure.go — "Conjure a card named [X] into [zone]" (Alchemy
// / MTGA, CR §701.59-equivalent).
//
// Conjure creates a new real card instance with a named card's
// characteristics in a specified zone. The conjured card is a full Card
// object — NOT a token — so it can be cast, drawn, exiled, etc., and
// interacts with all normal card-handling code. The distinguishing
// runtime marker is Card.Meta["conjured"] = true, which cards or
// effects that key off conjure provenance can read.
//
// Architecture:
//
//   - ConjureCardLookup        package-level injectable card-by-name
//                              factory. Mirrors the per-card-hooks seam
//                              (ETBHook, TriggerHook, etc.): nil by
//                              default; tests inject an in-memory map,
//                              and production wires up a corpus-backed
//                              resolver. Keeping the lookup pluggable
//                              avoids importing the oracle corpus
//                              loader from gameengine.
//
//   - ConjureCard(gs, seat,    main entry. Resolves the card via
//     name, zone)              ConjureCardLookup, deep-copies the
//                              result (so each conjure produces an
//                              independent card object — repeated
//                              conjures of the same name don't share
//                              state), stamps Meta["conjured"]=true,
//                              re-homes Owner to `seat`, and places
//                              the card in `zone`. Returns the new
//                              Card on success or nil on lookup
//                              failure / invalid arguments.
//
//   - ConjureZone* constants   canonical zone string identifiers
//                              accepted by ConjureCard. We accept the
//                              raw strings ("hand", "library",
//                              "battlefield", "graveyard", "exile",
//                              "command") as well — the constants are
//                              for documentation / autocomplete.
//
//   - IsConjured(c) bool       predicate. Reads Meta["conjured"] and
//                              returns true when set. Safe on nil.

// Zone constants accepted by ConjureCard. The raw strings are also
// accepted (case-insensitive) for parity with the rest of the engine's
// zone-name conventions.
const (
	ConjureZoneHand        = "hand"
	ConjureZoneLibrary     = "library"
	ConjureZoneBattlefield = "battlefield"
	ConjureZoneGraveyard   = "graveyard"
	ConjureZoneExile       = "exile"
)

// ConjureCardLookup is the package-level seam that resolves a card
// name to a fresh Card. Production wires up a corpus-backed function;
// tests inject an in-memory map. Returns nil when the name is unknown.
//
// The function is expected to return a card with the requested
// characteristics PRE-RESOLUTION — ConjureCard takes responsibility
// for the deep copy + provenance stamp + zone placement.
var ConjureCardLookup func(name string) *Card

// ConjureCard creates a real card with `name`'s characteristics in
// `zone` controlled/owned by `seatIdx`. Returns the new Card on
// success.
//
// Returns nil and emits no events on:
//   - nil gs, invalid seat, empty name, unrecognized zone.
//   - ConjureCardLookup unset (no name-resolution backend wired up).
//   - ConjureCardLookup returning nil (unknown card name).
//
// On success:
//   - The looked-up card is deep-copied so repeated conjures of the
//     same name produce independent objects (no aliasing through
//     shared slices or AST pointers — the AST pointer IS shared since
//     AST is immutable, but Types/Colors/Meta are deep-copied).
//   - Card.Owner is reset to `seatIdx` (the lookup factory often
//     returns Owner=0 by default).
//   - Card.Meta["conjured"] is set to true (allocating Meta if nil).
//   - The card is placed in the requested zone:
//       hand        → seat.Hand append
//       library     → seat.Library prepend (top); matches
//                     the prompt's "appends to top by default"
//                     spec. Callers that want bottom-of-library
//                     placement should append after this call.
//       battlefield → seat.Battlefield via a fresh Permanent;
//                     RegisterReplacementsForPermanent +
//                     FirePermanentETBTriggers run so the
//                     conjured permanent's ETB triggers fire
//                     (and observers see it enter).
//       graveyard   → seat.Graveyard append
//       exile       → seat.Exile append
//   - A "conjure" event is logged with the card name, zone, and seat.
//
// Returns the new Card (or, for battlefield, the underlying Card of
// the Permanent — same pointer) so callers can chain further setup
// (e.g. enter the battlefield tapped, add counters).
func ConjureCard(gs *GameState, seatIdx int, name, zone string) *Card {
	if gs == nil || seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return nil
	}
	if strings.TrimSpace(name) == "" {
		return nil
	}
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return nil
	}
	zoneCanon := strings.ToLower(strings.TrimSpace(zone))
	switch zoneCanon {
	case ConjureZoneHand, ConjureZoneLibrary, ConjureZoneBattlefield,
		ConjureZoneGraveyard, ConjureZoneExile:
		// ok
	default:
		return nil
	}

	if ConjureCardLookup == nil {
		return nil
	}
	src := ConjureCardLookup(name)
	if src == nil {
		return nil
	}

	// Deep-copy so the conjured card is independent. DeepCopy clones
	// Types/Colors/Meta + the shared AST pointer is fine (AST is
	// immutable characteristic data).
	card := src.DeepCopy()
	card.Owner = seatIdx
	if card.Meta == nil {
		card.Meta = map[string]any{}
	}
	card.Meta["conjured"] = true

	// Mark the seat as having conjured a card this turn. Mirrors the
	// other "spell_X_this_turn" markers and is useful for analytics +
	// any future card that keys off "if a conjured card entered this
	// turn." Cleared in cleanup alongside other per-turn flags.
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	gs.Flags["conjured_card_this_turn:"+itoa(seatIdx)]++

	switch zoneCanon {
	case ConjureZoneHand:
		seat.Hand = append(seat.Hand, card)
	case ConjureZoneLibrary:
		// CR convention: "conjure into your library" places on TOP
		// of the library (the first card drawn). Per the prompt's
		// (e) — "conjuring into library appends to top by default."
		// Callers wanting bottom can shift to the tail.
		seat.Library = append([]*Card{card}, seat.Library...)
	case ConjureZoneGraveyard:
		seat.Graveyard = append(seat.Graveyard, card)
	case ConjureZoneExile:
		seat.Exile = append(seat.Exile, card)
	case ConjureZoneBattlefield:
		perm := &Permanent{
			Card:          card,
			Controller:    seatIdx,
			Owner:         seatIdx,
			SummoningSick: true,
			Timestamp:     gs.NextTimestamp(),
			Counters:      map[string]int{},
			Flags:         map[string]int{},
		}
		seat.Battlefield = append(seat.Battlefield, perm)
		RegisterReplacementsForPermanent(gs, perm)
		FirePermanentETBTriggers(gs, perm)
	}

	gs.LogEvent(Event{
		Kind:   "conjure",
		Seat:   seatIdx,
		Source: card.DisplayName(),
		Details: map[string]interface{}{
			"name": name,
			"zone": zoneCanon,
			"rule": "701.59",
		},
	})
	return card
}

// IsConjured reports whether `c` was created via ConjureCard (or
// otherwise tagged with Meta["conjured"]=true). Safe on nil.
func IsConjured(c *Card) bool {
	if c == nil || c.Meta == nil {
		return false
	}
	v, ok := c.Meta["conjured"].(bool)
	return ok && v
}

// ConjuredCardsThisTurn reports how many cards `seatIdx` conjured
// during the current turn. Reader for the
// "conjured_card_this_turn:<seat>" per-turn flag.
func ConjuredCardsThisTurn(gs *GameState, seatIdx int) int {
	if gs == nil || gs.Flags == nil {
		return 0
	}
	return gs.Flags["conjured_card_this_turn:"+itoa(seatIdx)]
}
