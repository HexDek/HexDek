package gameengine

import "strings"

// keywords_strive.go — CR §702.118 Strive (Journey into Nyx, 2014).
//
// Strive is an ADDITIONAL cost (CR §601.2g, NOT the §601.2f alt-cost
// path used by Flashback / Spectacle / Warp / Disguise). Per §702.118a:
//
//   "Strive — This spell costs {N} more to cast for each target beyond
//    the first."
//
// The spell's own mana cost remains the base; the caster declares the
// number of targets first, and the additional cost is added once per
// extra target. Strive spells may legally be cast with zero targets
// when their printed text allows it (e.g. "any number of target
// creatures"); in that case no additional cost is added.
//
// Surface:
//
//   - HasStrive(card)                          — keyword OR oracle-text
//                                                detection.
//   - StrivePerTargetCost(card)                — extract {N} from the
//                                                keyword args.
//   - StriveCastCost(card, strivePer, targets) — total cost to cast
//                                                including the
//                                                per-target additions.
//   - CastWithStrive(...)                      — full cast entry point
//                                                that pays the strive
//                                                total, stamps
//                                                CostMeta["strive_targets"]=N
//                                                on the StackItem, and
//                                                logs the cast trail.

// HasStrive reports whether the card has the strive keyword.
//
// We accept two detection paths because strive is older than the
// keyword-tagging coverage in HexDek's AST corpus: many Theros-block
// strive cards are parsed as static text that mentions "strive —" in
// the oracle string rather than a clean Keyword node.
//
//   1. cardHasKeywordByName(card, "strive") — modern AST tagging.
//   2. Oracle text contains the literal "strive —" or "strive -" prefix
//      — the printed Reminder-text introducer for the additional cost.
//
// Returns false for nil / AST-less cards (tokens, basic lands).
func HasStrive(card *Card) bool {
	if card == nil {
		return false
	}
	if cardHasKeywordByName(card, "strive") {
		return true
	}
	text := OracleTextLower(card)
	if text == "" {
		return false
	}
	// Match both em-dash ("strive —") and ASCII hyphen ("strive -")
	// renderings. Scryfall normalizes to em-dash but some older corpus
	// dumps use the hyphen.
	return strings.Contains(text, "strive —") || strings.Contains(text, "strive -")
}

// StrivePerTargetCost returns the {N} additional cost per extra target.
// Defaults to keywordArgCost("strive"), which itself falls back to the
// card's CMC when no numeric arg is parsed — that fallback is wrong for
// strive (most strive spells charge {1} or {2} per extra target, NOT
// their full CMC), so callers that aren't sure should consult the
// printed oracle text and pass the cost explicitly to StriveCastCost.
func StrivePerTargetCost(card *Card) int {
	return keywordArgCost(card, "strive")
}

// StriveCastCost returns the total mana cost to cast a strive spell
// with `numTargets` targets, given a base mana cost from card.CMC and
// an additional per-extra-target cost of `strivePerTarget`.
//
// Formula (CR §702.118a):
//
//   total = card.CMC + max(0, numTargets - 1) * strivePerTarget
//
// Zero-target case: spells that allow zero targets (e.g. "any number of
// target creatures") cast with `numTargets == 0` pay just the base
// cost — the `max(0, ...)` guard prevents a negative additional cost
// when the caller passes 0.
//
// Negative strivePerTarget is clamped to 0 (defensive — strive costs
// are always positive on real cards, but we don't want a malformed
// keyword arg to refund mana).
//
// Returns 0 for nil card.
func StriveCastCost(card *Card, strivePerTarget, numTargets int) int {
	if card == nil {
		return 0
	}
	if strivePerTarget < 0 {
		strivePerTarget = 0
	}
	extra := numTargets - 1
	if extra < 0 {
		extra = 0
	}
	return card.CMC + extra*strivePerTarget
}

// CastWithStrive casts `card` from `seatIdx`'s hand for its full
// strive-inclusive cost, with the caller-declared target count.
// CR §702.118a + §601.2g.
//
// Preconditions:
//   - card carries strive (HasStrive).
//   - card is in `seatIdx`'s hand.
//   - numTargets >= 0. We don't enforce an upper bound here — strive
//     spells typically allow targeting up to N matching permanents, but
//     the legal-target check belongs in the target-selection pipeline.
//   - The seat can afford the total cost computed by StriveCastCost.
//   - strivePerTarget should be the printed per-extra-target cost. Pass
//     -1 to fall back to StrivePerTargetCost(card) (note its CMC-fallback
//     caveat above).
//
// On success:
//   - Total cost (base + per-extra-target) is paid.
//   - The card is removed from hand.
//   - A StackItem is pushed with CostMeta:
//       "strive"              = true
//       "alt_cost"            = "strive"
//       "strive_targets"      = numTargets
//       "strive_per_target"   = strivePerTarget
//       "strive_total_cost"   = total
//       "strive_extra_paid"   = extra (the per-target add-on portion)
//     CastZone is ZoneHand. Note: strive is an ADDITIONAL cost
//     (§601.2g), not an alt cost — we use "alt_cost"="strive" for
//     consistency with the other alt-cost helpers' tagging scheme even
//     though the printed mana cost is still being paid (as part of the
//     total). The "strive" boolean is the load-bearing flag.
//   - A "strive_cast" event is logged with rule 702.118a, recording
//     amount = total cost paid.
//
// Returns a minimal CostPaymentResult on success, or a CastError with
// one of: no_strive_keyword, negative_targets, insufficient_mana,
// not_in_hand.
func CastWithStrive(gs *GameState, seatIdx int, card *Card, strivePerTarget, numTargets int) (*CostPaymentResult, error) {
	if gs == nil {
		return nil, &CastError{Reason: "nil game"}
	}
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return nil, &CastError{Reason: "invalid seat"}
	}
	if card == nil {
		return nil, &CastError{Reason: "nil card"}
	}
	if !HasStrive(card) {
		return nil, &CastError{Reason: "no_strive_keyword"}
	}
	if numTargets < 0 {
		return nil, &CastError{Reason: "negative_targets"}
	}
	if strivePerTarget < 0 {
		strivePerTarget = StrivePerTargetCost(card)
	}
	if strivePerTarget < 0 {
		strivePerTarget = 0
	}

	total := StriveCastCost(card, strivePerTarget, numTargets)
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return nil, &CastError{Reason: "nil seat"}
	}
	if seat.ManaPool < total {
		return nil, &CastError{Reason: "insufficient_mana"}
	}

	// Remove from hand BEFORE paying so a not_in_hand rejection cannot
	// leak mana. Matches CastWithSpectacle / CastFlashback / CastWarp.
	if !removeFromZone(seat, card, ZoneHand) {
		return nil, &CastError{Reason: "not_in_hand"}
	}

	// Pay the total. We don't split the payment into "base" + "additional"
	// log entries — the spell has a single overall mana cost that the
	// caster paid; the StackItem CostMeta records the breakdown so any
	// downstream consumer can reconstruct it.
	seat.ManaPool -= total
	SyncManaAfterSpend(seat)
	if total > 0 {
		gs.LogEvent(Event{
			Kind:   "pay_mana",
			Seat:   seatIdx,
			Amount: total,
			Source: card.DisplayName(),
			Details: map[string]interface{}{
				"reason":         "strive_cast",
				"keyword":        "strive",
				"rule":           "601.2g",
				"strive_targets": numTargets,
			},
		})
	}

	extra := numTargets - 1
	if extra < 0 {
		extra = 0
	}
	extraPaid := extra * strivePerTarget

	item := &StackItem{
		Card:       card,
		Controller: seatIdx,
		CastZone:   ZoneHand,
		Effect:     collectSpellEffect(card),
		CostMeta: map[string]interface{}{
			"strive":            true,
			"alt_cost":          "strive",
			"strive_targets":    numTargets,
			"strive_per_target": strivePerTarget,
			"strive_total_cost": total,
			"strive_extra_paid": extraPaid,
		},
	}
	PushStackItem(gs, item)

	// Per-turn flag so cards keying off "if a strive spell was cast
	// this turn" can read it. Mirrors SpellSpectacleThisTurn etc.
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	gs.Flags["spell_strive_this_turn:"+itoa(seatIdx)] = 1

	gs.LogEvent(Event{
		Kind:   "strive_cast",
		Seat:   seatIdx,
		Source: card.DisplayName(),
		Amount: total,
		Details: map[string]interface{}{
			"rule":           "702.118a",
			"strive_targets": numTargets,
			"extra_paid":     extraPaid,
		},
	})

	return &CostPaymentResult{}, nil
}

// SpellStriveThisTurn reports whether seatIdx cast at least one strive
// spell during the current turn. Mirrors SpellSpectacleThisTurn /
// SpellWarpedThisTurn.
func SpellStriveThisTurn(gs *GameState, seatIdx int) bool {
	if gs == nil || gs.Flags == nil {
		return false
	}
	return gs.Flags["spell_strive_this_turn:"+itoa(seatIdx)] > 0
}
