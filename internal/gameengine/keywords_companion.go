package gameengine

// keywords_companion.go — Companion (CR §702.139, Ikoria 2020).
//
// CR §702.139a: Companion is a static ability that functions outside
//                the game. "Companion — [restriction]" lets the owner
//                place this card in the companion zone before the
//                game if their starting deck satisfies the restriction.
// CR §702.139b: Once per game, any time the player could cast a
//                sorcery, they may pay {3} to put the companion from
//                the companion zone into their hand.
// CR §702.139c: A companion's deck-construction restriction is
//                checked once before the game begins and uses each
//                card's printed characteristics.
//
// Engine surface (canonical):
//
//   - HasCompanion(card) bool
//   - CompanionRestriction(card) string                — human-readable
//   - ValidateCompanionDeck(card, deck) bool           — table-driven
//   - PreGameExileCompanion(gs, seat, card)            — pre-game setup
//   - PayCompanionCost(gs, seat, card) error           — {3} once/game
//   - SeatCompanionExiled(seat) *Card                  — accessor
//   - SeatCompanionUsed(seat) bool                     — accessor
//
// The companion is stored on the seat under the existing
// Seat.Companion / Seat.CompanionMoved fields (kept under those names
// for backward compatibility with legacy DeclareCompanion /
// MoveCompanionToHand callers). The accessor functions surface them
// with the conceptual names from the §702.139 rule text
// (CompanionExiled / CompanionUsed) so downstream tooling can read
// the state with the canonical vocabulary.

import "strings"

// ---------------------------------------------------------------------------
// HasCompanion / CompanionRestriction
// ---------------------------------------------------------------------------

// HasCompanion returns true if the card carries the companion keyword.
func HasCompanion(card *Card) bool {
	return cardHasKeywordByName(card, "companion")
}

// CompanionRestriction returns a human-readable description of the
// deck-construction restriction `card` imposes. Empty string for
// non-companions.
func CompanionRestriction(card *Card) string {
	if card == nil {
		return ""
	}
	name := strings.ToLower(card.DisplayName())
	switch {
	case strings.Contains(name, "lurrus"):
		return "Each permanent card in your starting deck has mana value 2 or less."
	case strings.Contains(name, "yorion"):
		return "Your starting deck contains at least 20 cards more than the minimum deck size."
	case strings.Contains(name, "obosh"):
		return "Each nonland card in your starting deck has an odd converted mana cost."
	case strings.Contains(name, "gyruda"):
		return "Each card in your starting deck has an even converted mana cost."
	case strings.Contains(name, "kaheera"):
		return "Each creature card in your starting deck is a Cat, Elemental, Nightmare, Dinosaur, or Beast card."
	case strings.Contains(name, "jegantha"):
		return "No card in your starting deck has more than one of the same mana symbol in its mana cost."
	case strings.Contains(name, "keruga"):
		return "Each nonland card in your starting deck has converted mana cost 3 or greater."
	case strings.Contains(name, "umori"):
		return "Each nonland card in your starting deck shares a card type."
	case strings.Contains(name, "zirda"):
		return "Each permanent card in your starting deck has an activated ability."
	case strings.Contains(name, "lutri"):
		return "Each nonland card in your starting deck has a different name."
	}
	if HasCompanion(card) {
		return "Companion restriction (unwired)."
	}
	return ""
}

// ---------------------------------------------------------------------------
// ValidateCompanionDeck
// ---------------------------------------------------------------------------

// ValidateCompanionDeck reports whether `deck` satisfies the
// deck-construction restriction `card` imposes. CR §702.139c.
// Unknown companions (no validator wired) return true so the engine
// doesn't reject decks for partial coverage.
func ValidateCompanionDeck(card *Card, deck []*Card) bool {
	if card == nil {
		return false
	}
	return validateCompanionByName(card.DisplayName(), deck)
}

// validateCompanionByName is the table-driven dispatcher used by both
// ValidateCompanionDeck and the legacy CheckCompanionRestriction shim.
func validateCompanionByName(companionName string, deck []*Card) bool {
	name := strings.ToLower(companionName)
	switch {
	case strings.Contains(name, "lurrus"):
		for _, c := range deck {
			if companionCardIsPermanent(c) && c.CMC > 2 {
				return false
			}
		}
		return true

	case strings.Contains(name, "yorion"):
		return len(deck) >= 80

	case strings.Contains(name, "obosh"):
		for _, c := range deck {
			if c == nil {
				continue
			}
			if !cardHasType(c, "land") && c.CMC%2 == 0 {
				return false
			}
		}
		return true

	case strings.Contains(name, "gyruda"):
		for _, c := range deck {
			if c == nil {
				continue
			}
			if c.CMC%2 != 0 {
				return false
			}
		}
		return true

	case strings.Contains(name, "kaheera"):
		allowed := map[string]bool{
			"cat": true, "elemental": true, "nightmare": true,
			"dinosaur": true, "beast": true,
		}
		for _, c := range deck {
			if c == nil || !cardHasType(c, "creature") {
				continue
			}
			ok := false
			for _, t := range c.Types {
				if allowed[strings.ToLower(t)] {
					ok = true
					break
				}
			}
			if !ok {
				return false
			}
		}
		return true

	case strings.Contains(name, "jegantha"):
		// "No card has more than one of the same mana symbol in its
		// mana cost." Inspect each card's ManaCostString.
		for _, c := range deck {
			if c == nil {
				continue
			}
			if manaCostHasRepeatColoredPip(c.ManaCostString) {
				return false
			}
		}
		return true

	case strings.Contains(name, "keruga"):
		for _, c := range deck {
			if c == nil {
				continue
			}
			if !cardHasType(c, "land") && c.CMC < 3 {
				return false
			}
		}
		return true

	case strings.Contains(name, "umori"):
		// Each nonland card shares a card type. Intersect the
		// permanent/spell type sets across nonland cards.
		var candidate map[string]bool
		for _, c := range deck {
			if c == nil || cardHasType(c, "land") {
				continue
			}
			perms := map[string]bool{}
			for _, t := range c.Types {
				tl := strings.ToLower(t)
				switch tl {
				case "creature", "artifact", "enchantment", "planeswalker",
					"instant", "sorcery", "battle", "tribal":
					perms[tl] = true
				}
			}
			if len(perms) == 0 {
				return false
			}
			if candidate == nil {
				candidate = perms
				continue
			}
			for k := range candidate {
				if !perms[k] {
					delete(candidate, k)
				}
			}
			if len(candidate) == 0 {
				return false
			}
		}
		return true

	case strings.Contains(name, "zirda"):
		for _, c := range deck {
			if c == nil || !companionCardIsPermanent(c) {
				continue
			}
			if !cardHasAnyActivatedAbility(c) {
				return false
			}
		}
		return true

	case strings.Contains(name, "lutri"):
		seen := map[string]bool{}
		for _, c := range deck {
			if c == nil || cardHasType(c, "land") {
				continue
			}
			n := c.DisplayName()
			if seen[n] {
				return false
			}
			seen[n] = true
		}
		return true
	}
	// Unknown companion: assume valid.
	return true
}

// manaCostHasRepeatColoredPip scans a printed mana-cost string and
// reports whether any single colored pip appears more than once.
// Colored pips: W, U, B, R, G. Other pips (generic, X, S, C, hybrid,
// Phyrexian) are ignored.
func manaCostHasRepeatColoredPip(cost string) bool {
	counts := map[byte]int{}
	for i := 0; i < len(cost); i++ {
		ch := cost[i]
		switch ch {
		case 'W', 'U', 'B', 'R', 'G':
			counts[ch]++
			if counts[ch] > 1 {
				return true
			}
		}
	}
	return false
}

// cardHasAnyActivatedAbility reports whether the card's AST carries
// at least one Activated ability node. Backs Zirda's restriction
// (the printed-card check is "activated ability" rather than the
// parser's cost-less Activated bodies for spell effects, but the
// MVP accepts any Activated presence — corpus permanent ASTs
// separate cast bodies from real activations).
func cardHasAnyActivatedAbility(c *Card) bool {
	if c == nil || c.AST == nil {
		return false
	}
	for _, ab := range c.AST.Abilities {
		if ab == nil {
			continue
		}
		if ab.Kind() == "activated" {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// PreGameExileCompanion / PayCompanionCost
// ---------------------------------------------------------------------------

// PreGameExileCompanion stashes `card` as `seat`'s companion. Called
// during pre-game setup, AFTER mulligan resolution and BEFORE the
// first turn (CR §702.139a). Stores on the legacy Seat.Companion
// field; SeatCompanionExiled is the canonical accessor.
//
// Calling twice with the same card is a no-op; calling twice with
// different cards replaces the prior designation (defensive shape
// for unit tests).
func PreGameExileCompanion(gs *GameState, seatIdx int, card *Card) {
	if gs == nil || card == nil {
		return
	}
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return
	}
	seat.Companion = card
	seat.CompanionMoved = false
	gs.LogEvent(Event{
		Kind:   "companion_declared",
		Seat:   seatIdx,
		Source: card.DisplayName(),
		Details: map[string]interface{}{
			"rule": "702.139a",
		},
	})
}

// PayCompanionCost activates the §702.139b companion-to-hand ability.
// Pays {3} from seat.ManaPool, moves the companion from exile to the
// seat's hand, and flips Seat.CompanionMoved so the ability can't
// fire twice in one game.
//
// Validation:
//   - if `card` is non-nil, it must match seat.Companion
//   - seat.CompanionMoved must be false (once per game)
//   - seat.ManaPool >= 3
//   - sorcery-speed timing: gs.Active == seatIdx AND gs.Stack empty
//
// Returns nil on success, a CastError on any failure.
func PayCompanionCost(gs *GameState, seatIdx int, card *Card) error {
	if gs == nil || seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return &CastError{Reason: "invalid_state"}
	}
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return &CastError{Reason: "nil_seat"}
	}
	if seat.Companion == nil {
		return &CastError{Reason: "no_companion"}
	}
	if card != nil && card != seat.Companion {
		return &CastError{Reason: "companion_mismatch"}
	}
	if seat.CompanionMoved {
		return &CastError{Reason: "companion_already_used"}
	}
	const companionTax = 3
	if seat.ManaPool < companionTax {
		return &CastError{Reason: "insufficient_mana_for_companion"}
	}
	if gs.Active != seatIdx {
		return &CastError{Reason: "not_your_turn"}
	}
	if len(gs.Stack) > 0 {
		return &CastError{Reason: "stack_not_empty"}
	}

	seat.ManaPool -= companionTax
	SyncManaAfterSpend(seat)
	gs.LogEvent(Event{
		Kind:   "pay_mana",
		Seat:   seatIdx,
		Amount: companionTax,
		Source: seat.Companion.DisplayName(),
		Details: map[string]interface{}{
			"reason": "companion_tax",
			"rule":   "702.139b",
		},
	})

	moved := seat.Companion
	MoveCard(gs, moved, seatIdx, "companion", "hand", "companion-to-hand")
	seat.CompanionMoved = true
	gs.LogEvent(Event{
		Kind:   "companion_to_hand",
		Seat:   seatIdx,
		Source: moved.DisplayName(),
		Details: map[string]interface{}{
			"rule": "702.139b",
		},
	})
	return nil
}

// ---------------------------------------------------------------------------
// Canonical-name accessors
// ---------------------------------------------------------------------------

// SeatCompanionExiled returns the card currently in `seat`'s companion
// zone, or nil if no companion was declared. The "CompanionExiled"
// name follows the §702.139 conceptual model (the companion zone is
// implemented as a tagged exile subset).
func SeatCompanionExiled(seat *Seat) *Card {
	if seat == nil {
		return nil
	}
	return seat.Companion
}

// SeatCompanionUsed reports whether `seat` has already paid the {3}
// once-per-game companion-to-hand activation.
func SeatCompanionUsed(seat *Seat) bool {
	if seat == nil {
		return false
	}
	return seat.CompanionMoved
}
