package gameengine

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameast"
)

// keywords_threshold.go — CR §702.72 Threshold (Odyssey, 2001).
//
// Threshold is a passive ability-word gating mechanic:
//
//   "Threshold — As long as seven or more cards are in your graveyard,
//    [effect]."
//
// Unlike triggered keywords, threshold is a STATIC riders-when-active
// rider: the printed effect is continuously available whenever the
// gating condition (seven+ cards in YOUR graveyard) holds, and goes
// silent the moment your graveyard drops below seven (e.g. via
// Tormod's Crypt, opponent's graveyard hate, your own recursion
// pulling cards back).
//
// Two distinct surfaces are exposed:
//
//   1. ThresholdActive(gs, seatIdx) — pure-function gating predicate
//      readable from anywhere. Cards with threshold-gated abilities
//      that aren't routed through the resolver-rider path (e.g. a
//      static buff on a creature) consult this directly.
//
//   2. ApplyThresholdRider(gs, src) — resolver-side rider executor,
//      modeled on ApplyMaxSpeedRider. When a spell with a printed
//      "Threshold — [effect]" rider resolves, and the controller's
//      graveyard has 7+ cards, the rider's tagged effect is resolved
//      as part of the spell's one-shot resolution (same source,
//      controller, targets).
//
// Wired into resolveSequence (resolve.go) at the outer-sequence
// boundary so a spell only re-fires the threshold rider once, never
// per nested Sequence node. We share the max-speed rider's depth
// gate convention with a dedicated counter
// "_threshold_rider_depth".
//
// Per-seat independence: ThresholdActive reads the SPECIFIC seat's
// graveyard. Seat 0 having 7 cards in their graveyard does NOT enable
// seat 1's threshold-printed spells; each seat counts only their own
// graveyard.

// ThresholdGraveCount is the number of cards in YOUR graveyard
// required for threshold-gated effects to be active. CR §702.72a.
const ThresholdGraveCount = 7

// HasThreshold reports whether the card prints a threshold ability.
// Detection paths (mirrors HasMaxSpeedRider):
//
//  1. cardHasKeywordByName(card, "threshold") — AST keyword tag for
//     modern corpus dumps that promote threshold to a Keyword node.
//  2. Oracle text contains "threshold —" or "threshold -" — the
//     printed ability-word prefix. Scryfall normalizes to em-dash; we
//     accept the ASCII hyphen for older dumps.
//
// Independent of game state — answers "does this card print
// threshold", not "is threshold currently active for the controller".
// Returns false for nil / AST-less cards.
func HasThreshold(card *Card) bool {
	if card == nil {
		return false
	}
	if cardHasKeywordByName(card, "threshold") {
		return true
	}
	text := OracleTextLower(card)
	if text == "" {
		return false
	}
	return strings.Contains(text, "threshold —") || strings.Contains(text, "threshold -")
}

// ThresholdActive reports whether `seatIdx`'s threshold condition is
// met right now: graveyard size >= 7. CR §702.72a.
//
// Per-seat: each seat consults their OWN graveyard. Threshold doesn't
// see opponent's graveyards. Returns false for invalid seat / nil gs.
//
// The check is recomputed on every call — no caching — so it stays
// honest across mid-turn graveyard mutations (Tormod's Crypt exiling
// your yard, Coffin Purge pulling cards back, etc.).
func ThresholdActive(gs *GameState, seatIdx int) bool {
	if gs == nil || seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return false
	}
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return false
	}
	return len(seat.Graveyard) >= ThresholdGraveCount
}

// findThresholdRiderEffect returns the Effect of the first Activated
// ability on `card.AST` tagged as the threshold rider payload, or nil
// when no tagged ability is present.
//
// Tagging convention (mirrors max-speed): either Cost.Extra contains
// the literal "threshold_rider" OR Activated.Raw begins with
// "threshold" (case-insensitive).
func findThresholdRiderEffect(card *Card) gameast.Effect {
	if card == nil || card.AST == nil {
		return nil
	}
	for _, ab := range card.AST.Abilities {
		a, ok := ab.(*gameast.Activated)
		if !ok || a == nil || a.Effect == nil {
			continue
		}
		if abilityIsTaggedThresholdRider(a) {
			return a.Effect
		}
	}
	return nil
}

// abilityIsTaggedThresholdRider returns true when an Activated ability
// carries the threshold rider-payload marker. Two recognized shapes:
//
//   1. Cost.Extra contains "threshold_rider" — explicit corpus tag.
//   2. Activated.Raw begins with "threshold" — fallback for corpus
//      dumps that preserved the printed rider line without an
//      extra-cost marker.
func abilityIsTaggedThresholdRider(a *gameast.Activated) bool {
	if a == nil {
		return false
	}
	for _, extra := range a.Cost.Extra {
		if strings.EqualFold(extra, "threshold_rider") {
			return true
		}
	}
	raw := strings.ToLower(strings.TrimSpace(a.Raw))
	return strings.HasPrefix(raw, "threshold")
}

// ApplyThresholdRider executes the threshold rider for `src`, if any.
// Returns true if the rider actually fired.
//
// Conditions to fire:
//   - src is non-nil and has a Card.
//   - HasThreshold(src.Card) — the card prints a threshold rider.
//   - ThresholdActive(gs, src.Controller) — the controller's
//     graveyard has 7+ cards RIGHT NOW. The check is recomputed
//     here (not snapshotted at cast time) so threshold turning off
//     between cast and resolve correctly suppresses the rider.
//
// On fire we log a threshold_rider event. If the AST carries a
// tagged threshold-rider effect, ResolveEffect runs it with `src`
// as the source. When no tagged effect is present (common today
// because most threshold cards predate the corpus tagging
// convention), we log a threshold_rider_pending event so per_card
// handlers and corpus-backfill jobs can find affected spells.
func ApplyThresholdRider(gs *GameState, src *Permanent) bool {
	if gs == nil || src == nil || src.Card == nil {
		return false
	}
	if !HasThreshold(src.Card) {
		return false
	}
	if !ThresholdActive(gs, src.Controller) {
		return false
	}

	gs.LogEvent(Event{
		Kind:   "threshold_rider",
		Seat:   src.Controller,
		Source: src.Card.DisplayName(),
		Details: map[string]interface{}{
			"rule":         "702.72",
			"grave_count":  len(gs.Seats[src.Controller].Graveyard),
			"grave_needed": ThresholdGraveCount,
		},
	})

	if eff := findThresholdRiderEffect(src.Card); eff != nil {
		ResolveEffect(gs, src, eff)
		return true
	}

	gs.LogEvent(Event{
		Kind:   "threshold_rider_pending",
		Seat:   src.Controller,
		Source: src.Card.DisplayName(),
		Details: map[string]interface{}{
			"rule":   "702.72",
			"reason": "rider_payload_not_in_ast",
		},
	})
	return true
}
