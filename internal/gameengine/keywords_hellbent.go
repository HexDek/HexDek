package gameengine

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameast"
)

// keywords_hellbent.go — Hellbent (CR §702.45, Dissension 2006).
//
// CR §702.45a: "Hellbent — [effect]" is a static keyword ability that
//               applies as long as the controller has no cards in
//               hand. Functionally an "as long as" rider, same shape
//               as Threshold (CR §702.74) and Metalcraft (§702.105):
//               an always-on condition the resolver / per_card layer
//               consults at evaluation time, with no triggered ETB
//               component.
//
// Surface:
//
//   - HasHellbent(card)         — detector. AST keyword OR oracle
//                                 text scan (sibling of HasStrive's
//                                 two-path detector to handle both
//                                 the modern AST tagging and older
//                                 corpus dumps that emit the rider
//                                 as a Static.Raw line).
//   - HellbentActive(gs, seat)  — predicate. True iff the seat has
//                                 exactly zero cards in hand.
//
// Resolver wiring:
//
//   evalCondition (resolve.go) now recognises Condition{Kind:"hellbent"}.
//   AST parsers that want to emit a Hellbent-gated Conditional can use
//   that Kind; the evaluator routes to HellbentActive(gs, src.Controller).
//   Per_card handlers that have a Hellbent rider pre-computed can also
//   call HellbentActive directly without going through the AST.

// HasHellbent reports whether the card carries the §702.45 rider.
// Detection paths (mirrors HasStrive):
//
//  1. cardHasKeywordByName(card, "hellbent") — modern AST tagging.
//  2. Oracle text contains "hellbent —" or "hellbent -" — the printed
//     reminder-text introducer for the rider line. Some corpus dumps
//     also drop the em-dash and write "hellbent ", so we accept that
//     as a fallback when the dash forms don't match.
//
// Returns false for nil / AST-less cards.
func HasHellbent(card *Card) bool {
	if card == nil {
		return false
	}
	if cardHasKeywordByName(card, "hellbent") {
		return true
	}
	text := OracleTextLower(card)
	if text == "" {
		return false
	}
	if strings.Contains(text, "hellbent —") || strings.Contains(text, "hellbent -") {
		return true
	}
	// Dash-less fallback: "hellbent " followed by something. We require
	// the trailing space to avoid colliding with "hellbentish" etc.
	return strings.Contains(text, "hellbent ")
}

// HellbentActive reports whether seatIdx currently satisfies the
// Hellbent condition — i.e. has exactly zero cards in hand. Computed
// fresh from gs.Seats[seatIdx].Hand on every call so transient
// hand-emptying / hand-refilling effects (discard a card → draw a
// card on the same stack pop) flip the result correctly.
//
// Returns false for invalid seat indices and nil game.
func HellbentActive(gs *GameState, seatIdx int) bool {
	if gs == nil || seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return false
	}
	s := gs.Seats[seatIdx]
	if s == nil {
		return false
	}
	return len(s.Hand) == 0
}

// IsHellbent is a Permanent-facing convenience: "is this permanent's
// controller currently hellbent?" Mirrors IsMaxSpeed's shape from the
// round-25 max-speed rider hook so per_card resolvers can write the
// same predicate style across all rider keywords.
func IsHellbent(gs *GameState, perm *Permanent) bool {
	if perm == nil {
		return false
	}
	return HellbentActive(gs, perm.Controller)
}

// ---------------------------------------------------------------------------
// Rider executor (sibling of ApplyThresholdRider / ApplyMetalcraftRider /
// ApplyMaxSpeedRider).
// ---------------------------------------------------------------------------

// findHellbentRiderEffect walks the card's AST looking for the tagged
// "hellbent" rider payload. Tagging convention mirrors the other gated
// riders: Cost.Extra contains "hellbent_rider" OR Activated.Raw begins
// with "hellbent" (case-insensitive fallback for corpus dumps).
func findHellbentRiderEffect(card *Card) gameast.Effect {
	if card == nil || card.AST == nil {
		return nil
	}
	for _, ab := range card.AST.Abilities {
		a, ok := ab.(*gameast.Activated)
		if !ok || a == nil || a.Effect == nil {
			continue
		}
		if abilityIsTaggedHellbentRider(a) {
			return a.Effect
		}
	}
	return nil
}

func abilityIsTaggedHellbentRider(a *gameast.Activated) bool {
	if a == nil {
		return false
	}
	for _, extra := range a.Cost.Extra {
		if strings.EqualFold(extra, "hellbent_rider") {
			return true
		}
	}
	raw := strings.ToLower(strings.TrimSpace(a.Raw))
	return strings.HasPrefix(raw, "hellbent")
}

// ApplyHellbentRider executes the hellbent rider for `src`, if any.
// Returns true if the rider actually fired.
//
// Conditions to fire:
//   - src is non-nil and has a Card.
//   - HasHellbent(src.Card).
//   - HellbentActive(gs, src.Controller) — recomputed live, so a card
//     drawn mid-resolution correctly turns the rider off.
//
// Logs hellbent_rider on fire. If the AST carries a tagged rider
// effect, ResolveEffect runs it; otherwise we log
// hellbent_rider_pending so per_card handlers and corpus-backfill jobs
// can find affected spells.
func ApplyHellbentRider(gs *GameState, src *Permanent) bool {
	if gs == nil || src == nil || src.Card == nil {
		return false
	}
	if !HasHellbent(src.Card) {
		return false
	}
	if !HellbentActive(gs, src.Controller) {
		return false
	}

	handSize := 0
	if src.Controller >= 0 && src.Controller < len(gs.Seats) && gs.Seats[src.Controller] != nil {
		handSize = len(gs.Seats[src.Controller].Hand)
	}

	gs.LogEvent(Event{
		Kind:   "hellbent_rider",
		Seat:   src.Controller,
		Source: src.Card.DisplayName(),
		Details: map[string]interface{}{
			"rule":      "702.45",
			"hand_size": handSize,
		},
	})

	if eff := findHellbentRiderEffect(src.Card); eff != nil {
		ResolveEffect(gs, src, eff)
		return true
	}

	gs.LogEvent(Event{
		Kind:   "hellbent_rider_pending",
		Seat:   src.Controller,
		Source: src.Card.DisplayName(),
		Details: map[string]interface{}{
			"rule":   "702.45",
			"reason": "rider_payload_not_in_ast",
		},
	})
	return true
}
