package gameengine

// keywords_saddle.go — Saddle N (CR §702.171, OTJ Mounts) as a real
// explicit-tappers action.
//
// CR §702.171a: Saddle N is an activated ability of Mount permanents.
//                Activate by tapping any number of OTHER untapped
//                creatures you control whose total power is N or more.
// CR §702.171b: The Mount becomes "saddled" until end of turn.
// CR §702.171c: Saddled is a designation. Mounts have a "while saddled"
//                clause that turns on additional abilities once the
//                designation is set.
//
// Two activation paths exist in the engine:
//
//   - ActivateSaddle (keywords_batch4.go) — auto-greedy: the engine
//     picks the highest-power creatures and taps them down until the
//     saddle cost is met. Used by AI shortcuts and tests that don't
//     need fine-grained control.
//
//   - SaddleMount (this file) — explicit: the caller hands the engine
//     a specific list of tappers and the engine validates each one
//     (controller match, untapped, not the mount, IsCreature) and
//     commits the tap + designation atomically. This is the canonical
//     path for the Hat's tactical saddle decisions and for per-card
//     handlers that need to choose which creatures to pay with.
//
// Both paths converge on the same post-state:
//   - tappers are tapped
//   - mount.Flags["saddled"] = 1
//   - mount.SaddlersThisTurn carries the tapper pointers
//   - "saddle" event is logged
// The until-end-of-turn cleanup (UnsaddleAtEOT) drops the designation
// and clears the saddlers list — wired into ScanExpiredDurations from
// phases.go's cleanup pass.

// ---------------------------------------------------------------------------
// HasSaddle / SaddleCost
// ---------------------------------------------------------------------------

// HasSaddle returns true if the card has a saddle keyword.
func HasSaddle(card *Card) bool {
	return cardHasKeywordByName(card, "saddle")
}

// SaddleCost returns the N value of a card's "Saddle N" keyword (the
// total creature-power threshold). 0 when the keyword is absent or
// malformed.
func SaddleCost(card *Card) int {
	return keywordArgCost(card, "saddle")
}

// PermHasSaddle is the battlefield counterpart to HasSaddle — checks
// the permanent's active face. Cleaner than reaching through
// perm.Card at every call site.
func PermHasSaddle(perm *Permanent) bool {
	if perm == nil {
		return false
	}
	return HasSaddle(perm.Card)
}

// ---------------------------------------------------------------------------
// PermIsSaddled
// ---------------------------------------------------------------------------

// PermIsSaddled reports whether the Mount currently has the saddled
// designation. Backs the "while saddled" clause gate on every Mount
// card's printed body (CR §702.171c).
func PermIsSaddled(perm *Permanent) bool {
	if perm == nil || perm.Flags == nil {
		return false
	}
	return perm.Flags["saddled"] > 0
}

// ---------------------------------------------------------------------------
// SaddleMount — explicit-tapper activation
// ---------------------------------------------------------------------------

// SaddleMount attempts to saddle `mount` by tapping exactly the
// permanents in `tappers`. CR §702.171a.
//
// Validation (all-or-nothing — if any check fails the mount stays
// unsaddled and no creature gets tapped):
//
//   - mount must be a saddle-bearing permanent the seat controls
//   - each tapper must be non-nil, controlled by `seatIdx`, untapped,
//     a creature, and NOT the mount itself (no self-saddle)
//   - each tapper must be a distinct permanent (no double-counting)
//   - sum of tapper Power() values must be >= SaddleCost(mount.Card)
//
// On success:
//   - every tapper is tapped (perm.Tapped = true)
//   - mount.Flags["saddled"] = 1
//   - mount.SaddlersThisTurn appends the tapper pointers (so per-card
//     triggers like Gitrog's "creature that saddled it this turn" can
//     reference them)
//   - a "saddle" event is logged with the running total power
//
// Returns true on a successful saddle, false on any validation failure.
// The function is atomic: no partial taps on failure.
func SaddleMount(gs *GameState, seatIdx int, mount *Permanent, tappers []*Permanent) bool {
	if gs == nil || mount == nil {
		return false
	}
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return false
	}
	if mount.Controller != seatIdx {
		return false
	}
	cost := SaddleCost(mount.Card)
	if cost <= 0 {
		// No printed saddle keyword (or malformed); refuse rather than
		// silently saddling for free.
		return false
	}

	// Validate every tapper before mutating anything.
	seen := map[*Permanent]bool{}
	total := 0
	for _, t := range tappers {
		if t == nil {
			return false
		}
		if t == mount {
			return false
		}
		if seen[t] {
			return false
		}
		seen[t] = true
		if t.Controller != seatIdx {
			return false
		}
		if t.Tapped {
			return false
		}
		if !t.IsCreature() {
			return false
		}
		total += t.Power()
	}
	if total < cost {
		return false
	}

	// Commit.
	for _, t := range tappers {
		t.Tapped = true
	}
	if mount.Flags == nil {
		mount.Flags = map[string]int{}
	}
	mount.Flags["saddled"] = 1
	mount.SaddlersThisTurn = append(mount.SaddlersThisTurn, tappers...)

	gs.LogEvent(Event{
		Kind:   "saddle",
		Seat:   seatIdx,
		Source: mount.Card.DisplayName(),
		Amount: cost,
		Details: map[string]interface{}{
			"creatures_tapped": len(tappers),
			"total_power":      total,
			"path":             "explicit_tappers",
			"rule":             "702.171a",
		},
	})
	return true
}

// ---------------------------------------------------------------------------
// UnsaddleAtEOT — end-of-turn cleanup
// ---------------------------------------------------------------------------

// UnsaddleAtEOT drops the saddled designation and clears
// SaddlersThisTurn on every permanent on every battlefield. CR
// §702.171b — saddled lasts until end of turn.
//
// The existing inline cleanup in ScanExpiredDurations (phases.go)
// performs the same work; this helper makes the operation a named,
// callable entry point so tests can exercise the reset in isolation
// without having to drive a full phase tick.
//
// Idempotent: calling on a board with no saddled mounts is a cheap
// no-op.
func UnsaddleAtEOT(gs *GameState) {
	if gs == nil {
		return
	}
	for _, seat := range gs.Seats {
		if seat == nil {
			continue
		}
		for _, p := range seat.Battlefield {
			if p == nil {
				continue
			}
			if p.Flags != nil && p.Flags["saddled"] != 0 {
				delete(p.Flags, "saddled")
			}
			if len(p.SaddlersThisTurn) > 0 {
				p.SaddlersThisTurn = nil
			}
		}
	}
}
