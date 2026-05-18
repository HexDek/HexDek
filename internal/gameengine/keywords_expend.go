package gameengine

// keywords_expend.go — Expend (CR §702.190, Bloomburrow 2024).
//
// CR §702.190a: "Expend N — [effect]" is a triggered ability. It
//               means "When you've spent N or more mana this turn,
//               [effect]." The trigger fires the moment the
//               controller's running mana-spent total CROSSES the
//               threshold; it does not re-fire on subsequent spends
//               within the same turn, and it does not fire if the
//               permanent enters the battlefield after the threshold
//               has already been met (the trigger watches the
//               crossing, not the state).
// CR §702.190b: Each Expend N ability on a permanent is independent.
//               A creature with both "Expend 4 — ..." and "Expend 8
//               — ..." fires the first when the controller's total
//               crosses 4 and the second when it crosses 8 — both
//               can fire on the same turn, but each only once.
//
// Implementation surface:
//
//   TrackManaSpentThisTurn(gs, seatIdx, amount)
//       Canonical mana-spent increment. Adds `amount` to the seat's
//       TurnCounters.ManaSpent and fires expend triggers for every
//       threshold the spend just crossed.
//
//   HasExpendTrigger(card, threshold) → bool
//       True if the card has an "Expend N" trigger at exactly
//       `threshold`. Inspects keyword AST nodes with name "expend"
//       whose first arg is the threshold (float64 or int).
//
//   ExpendThresholds(card) → []int
//       Returns all expend thresholds on a card (in declaration order,
//       deduplicated). Empty slice for cards without expend.
//
//   FireExpendTriggers(gs, seatIdx, prevSpent, newSpent)
//       For each permanent on seatIdx's battlefield with an Expend N
//       threshold where prevSpent < N <= newSpent, fires
//       FireCardTrigger("expend", {source, controller, threshold,
//       total}). This is the threshold-crossing fan-out the per_card
//       layer listens on.
//
// Wiring: TrackManaSpentThisTurn is the recommended public entry
// point. Engine code that already spends mana via SpendMana /
// SyncManaAfterSpend can call TrackManaSpentThisTurn from the same
// caller frame; this file deliberately does NOT modify those helpers
// to avoid invasive changes across the hot mana path. Callers that
// want expend coverage on a specific cost-payment site invoke
// TrackManaSpentThisTurn at the point of payment.

import (
	"strconv"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// HasExpendTrigger / ExpendThresholds
// ---------------------------------------------------------------------------

// HasExpendTrigger returns true if the card has an "Expend N" trigger
// at exactly `threshold`. Returns false for nil cards or AST-less
// cards.
func HasExpendTrigger(card *Card, threshold int) bool {
	if threshold <= 0 {
		return false
	}
	for _, n := range ExpendThresholds(card) {
		if n == threshold {
			return true
		}
	}
	return false
}

// ExpendThresholds returns all expend thresholds declared on `card`
// in declaration order, deduplicated. Each Keyword AST node whose
// name matches "expend" and whose first arg is a positive number
// contributes its threshold. Returns an empty (non-nil) slice when
// the card has no expend triggers.
func ExpendThresholds(card *Card) []int {
	if card == nil || card.AST == nil {
		return []int{}
	}
	var out []int
	seen := map[int]bool{}
	for _, ab := range card.AST.Abilities {
		kw, ok := ab.(*gameast.Keyword)
		if !ok {
			continue
		}
		if !keywordNameEquals(kw, "expend") {
			continue
		}
		if len(kw.Args) == 0 {
			continue
		}
		n := 0
		switch v := kw.Args[0].(type) {
		case float64:
			n = int(v)
		case int:
			n = v
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				n = parsed
			}
		}
		if n <= 0 || seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n)
	}
	if out == nil {
		return []int{}
	}
	return out
}

// ---------------------------------------------------------------------------
// TrackManaSpentThisTurn
// ---------------------------------------------------------------------------

// TrackManaSpentThisTurn bumps the seat's per-turn mana-spent
// counter and fires expend triggers for every threshold the spend
// just crossed. Pass `amount` as a non-negative integer; <= 0 is a
// safe no-op.
//
// This is the canonical entry point for Expend tracking. Cost-
// payment sites (CastSpell, alt-cost helpers, activated abilities,
// etc.) that want to drive Expend should call this immediately
// after deducting from the seat's mana pool. The function returns
// the new running total so callers can chain other "you spent
// mana"-style state updates if needed.
//
// Returns the new running total (post-increment); returns the
// existing total unchanged on nil-safe / invalid-seat input.
func TrackManaSpentThisTurn(gs *GameState, seatIdx int, amount int) int {
	if gs == nil || seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return 0
	}
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return 0
	}
	if amount <= 0 {
		return seat.Turn.ManaSpent
	}
	prev := seat.Turn.ManaSpent
	seat.Turn.ManaSpent = prev + amount
	gs.LogEvent(Event{
		Kind:   "mana_spent",
		Seat:   seatIdx,
		Amount: amount,
		Details: map[string]interface{}{
			"total": seat.Turn.ManaSpent,
			"rule":  "702.190",
		},
	})
	FireExpendTriggers(gs, seatIdx, prev, seat.Turn.ManaSpent)
	return seat.Turn.ManaSpent
}

// ---------------------------------------------------------------------------
// FireExpendTriggers
// ---------------------------------------------------------------------------

// FireExpendTriggers scans seatIdx's battlefield for permanents with
// Expend N triggers and fires "expend" card triggers for every
// threshold N where prevSpent < N <= newSpent. CR §702.190a — the
// trigger fires at the moment of the crossing, not on subsequent
// spends.
//
// Side effects:
//   - emits "expend_trigger" event per fan-out (Seat=seatIdx,
//     Source=the permanent's name, Details carries threshold + total)
//   - calls FireCardTrigger("expend", ctx) for each permanent so
//     per_card handlers can implement card-specific payoffs
//
// ctx keys forwarded to handlers:
//
//	"source":     *Permanent  — the expend-bearing permanent
//	"controller": int          — seatIdx
//	"threshold":  int          — the N that just crossed
//	"total":      int          — newSpent (running mana-spent total)
//
// Nil-safe on gs / seat. prevSpent < 0 is clamped to 0 so callers
// that don't pre-read the previous total still get correct
// threshold-crossing semantics.
func FireExpendTriggers(gs *GameState, seatIdx int, prevSpent, newSpent int) {
	if gs == nil || seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	if gs.Seats[seatIdx] == nil {
		return
	}
	if prevSpent < 0 {
		prevSpent = 0
	}
	if newSpent <= prevSpent {
		return
	}
	for _, p := range gs.Seats[seatIdx].Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if p.PhasedOut {
			continue
		}
		for _, n := range ExpendThresholds(p.Card) {
			if n <= prevSpent || n > newSpent {
				continue
			}
			gs.LogEvent(Event{
				Kind:   "expend_trigger",
				Seat:   seatIdx,
				Source: p.Card.DisplayName(),
				Amount: n,
				Details: map[string]interface{}{
					"threshold": n,
					"total":     newSpent,
					"rule":      "702.190a",
				},
			})
			FireCardTrigger(gs, "expend", map[string]interface{}{
				"source":     p,
				"controller": seatIdx,
				"threshold":  n,
				"total":      newSpent,
			})
		}
	}
}
