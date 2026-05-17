package gameengine

// keywords_behold.go — Behold (CR §701.4) as a real registry-driven
// keyword action with per-card "when you behold X" trigger fan-out.
//
// CR §701.4 ("Behold a [quality]"): As an additional cost or modal
// effect on Bloomburrow tribal/affinity cards, you "behold" a quality
// by EITHER:
//
//   - Revealing a card with that quality from your hand, OR
//   - Choosing a permanent with that quality that you control.
//
// The behold itself doesn't move cards or tap anything. It records, for
// the rest of this turn, that the active player beheld an object of
// that quality. Other cards trigger off the behold ("Whenever you
// behold a [quality], do Y") — that is what gives the action its game
// effect.
//
// Quality is typically a creature subtype (Dragon, Squirrel, Demon,
// Bat, Otter, Lizard …) but can be any oracle-named characteristic
// the cards print. This file matches qualities case-insensitively
// against Card.Types / Card.TypeLine via the engine-wide cardHasType
// helper, so any subtype the corpus loader populates works without
// per-quality wiring.
//
// Registry surface:
//
//   - Behold(gs, seat, quality, source)          → record + fire triggers
//   - BeholdRevealFromHand(gs, seat, quality, source, card)  bool
//   - BeholdChoosePermanent(gs, seat, quality, source, perm) bool
//   - HasBeheld(gs, seat, quality)               bool
//   - BeheldCount(gs, seat, quality)             int
//   - ClearBeholdRegistry(gs)                    wired into UntapAll
//
// Storage lives on gs.BeholdRegistry (state.go) — a per-seat
// map[string]int. Cleared at the start of every turn (UntapAll) so the
// "this turn" semantics close correctly.

import "strings"

// NormalizeBeholdQuality lowercases and trims the quality name so the
// registry uses a single canonical form. Callers that pass "Dragon",
// "dragon", " DRAGON " all hit the same bucket.
func NormalizeBeholdQuality(quality string) string {
	return strings.ToLower(strings.TrimSpace(quality))
}

// ---------------------------------------------------------------------------
// Quality matchers
// ---------------------------------------------------------------------------

// CardHasBeholdQuality reports whether `card` carries the named quality
// in a behold-eligible form. Quality matching uses cardHasType, which
// scans Card.Types AND Card.TypeLine — so a "Dragon" quality matches a
// card whose type line is "Creature — Elder Dragon" even if the corpus
// loader didn't explicitly split "elder dragon" into ["elder", "dragon"].
func CardHasBeholdQuality(card *Card, quality string) bool {
	q := NormalizeBeholdQuality(quality)
	if q == "" {
		return false
	}
	return cardHasType(card, q)
}

// PermHasBeholdQuality reports whether the permanent's active face
// carries the named quality. Used by the "choose a permanent" path of
// §701.4.
func PermHasBeholdQuality(perm *Permanent, quality string) bool {
	if perm == nil {
		return false
	}
	return CardHasBeholdQuality(perm.Card, quality)
}

// ---------------------------------------------------------------------------
// Recording primitive — Behold
// ---------------------------------------------------------------------------

// Behold records a behold of `quality` by `seat` and fires per-card
// "when you behold X" triggers. CR §701.4.
//
// This is the low-level recording primitive — it does not verify that
// the player actually has a card or permanent with the quality. Use
// BeholdRevealFromHand or BeholdChoosePermanent for the canonical
// gated paths.
//
// `source` is the human-readable name of the card that asked for the
// behold (e.g. "Mabel, Heir to Cragflame"). It is attached to the log
// event and to the trigger context so handlers can attribute correctly.
//
// Returns the new per-quality count for this seat-this-turn after the
// record. Cards that want a "first behold this turn" gate compare the
// returned value to 1.
func Behold(gs *GameState, seatIdx int, quality, source string) int {
	if gs == nil {
		return 0
	}
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return 0
	}
	q := NormalizeBeholdQuality(quality)
	if q == "" {
		return 0
	}
	if gs.BeholdRegistry == nil {
		gs.BeholdRegistry = map[int]map[string]int{}
	}
	seatReg, ok := gs.BeholdRegistry[seatIdx]
	if !ok {
		seatReg = map[string]int{}
		gs.BeholdRegistry[seatIdx] = seatReg
	}
	seatReg[q]++
	count := seatReg[q]

	gs.LogEvent(Event{
		Kind:   "behold",
		Seat:   seatIdx,
		Source: source,
		Amount: count,
		Details: map[string]interface{}{
			"quality":      q,
			"behold_count": count,
			"rule":         "701.4",
		},
	})

	FireCardTrigger(gs, "beheld", map[string]interface{}{
		"seat":         seatIdx,
		"quality":      q,
		"source":       source,
		"behold_count": count,
	})

	return count
}

// ---------------------------------------------------------------------------
// Gated paths — reveal from hand / choose a permanent
// ---------------------------------------------------------------------------

// BeholdRevealFromHand satisfies a Behold by revealing a hand card
// that carries the named quality (CR §701.4, reveal branch). The card
// stays in hand. Returns true on a successful behold, false when the
// card doesn't carry the quality or isn't actually in the seat's hand.
func BeholdRevealFromHand(gs *GameState, seatIdx int, quality, source string, card *Card) bool {
	if gs == nil || card == nil || seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return false
	}
	if !CardHasBeholdQuality(card, quality) {
		return false
	}
	// Verify the card is in the seat's hand. Behold is gated on a hand
	// reveal, so a caller that hands us a graveyard card or a stack
	// pointer should be rejected.
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return false
	}
	inHand := false
	for _, c := range seat.Hand {
		if c == card {
			inHand = true
			break
		}
	}
	if !inHand {
		return false
	}
	gs.LogEvent(Event{
		Kind:   "behold_reveal",
		Seat:   seatIdx,
		Source: source,
		Details: map[string]interface{}{
			"quality":   NormalizeBeholdQuality(quality),
			"card_name": card.DisplayName(),
			"path":      "reveal_from_hand",
			"rule":      "701.4a",
		},
	})
	Behold(gs, seatIdx, quality, source)
	return true
}

// BeholdChoosePermanent satisfies a Behold by choosing a permanent
// the seat controls that carries the named quality (CR §701.4, choose
// branch). The permanent stays on the battlefield as-is. Returns true
// on a successful behold, false when the perm doesn't carry the quality
// or isn't controlled by `seatIdx`.
func BeholdChoosePermanent(gs *GameState, seatIdx int, quality, source string, perm *Permanent) bool {
	if gs == nil || perm == nil || seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return false
	}
	if perm.Controller != seatIdx {
		return false
	}
	if !PermHasBeholdQuality(perm, quality) {
		return false
	}
	gs.LogEvent(Event{
		Kind:   "behold_choose",
		Seat:   seatIdx,
		Source: source,
		Details: map[string]interface{}{
			"quality":   NormalizeBeholdQuality(quality),
			"perm_name": perm.Card.DisplayName(),
			"path":      "choose_permanent",
			"rule":      "701.4a",
		},
	})
	Behold(gs, seatIdx, quality, source)
	return true
}

// ---------------------------------------------------------------------------
// Queries
// ---------------------------------------------------------------------------

// HasBeheld reports whether `seat` has beheld at least one object of
// `quality` during the current turn. CR §701.4.
func HasBeheld(gs *GameState, seatIdx int, quality string) bool {
	return BeheldCount(gs, seatIdx, quality) > 0
}

// BeheldCount returns the per-quality behold count for `seat` this
// turn. 0 when nothing has been beheld of that quality (including when
// the registry hasn't been initialized).
func BeheldCount(gs *GameState, seatIdx int, quality string) int {
	if gs == nil || gs.BeholdRegistry == nil {
		return 0
	}
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return 0
	}
	seatReg, ok := gs.BeholdRegistry[seatIdx]
	if !ok {
		return 0
	}
	return seatReg[NormalizeBeholdQuality(quality)]
}

// ---------------------------------------------------------------------------
// Turn-start reset hook
// ---------------------------------------------------------------------------

// ClearBeholdRegistry drops every per-seat per-quality behold count.
// Wired into UntapAll (phases.go) so the "this turn" window closes at
// each turn boundary. Reset is global (every seat), not per-seat,
// because §701.4 "this turn" follows the game-turn boundary rather
// than each player's own turn.
func ClearBeholdRegistry(gs *GameState) {
	if gs == nil {
		return
	}
	gs.BeholdRegistry = nil
}
