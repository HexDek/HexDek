package gameengine

// keywords_mount.go — Mount creature subtype (Outlaws of Thunder
// Junction 2024). Mount is a creature subtype that integrates with the
// existing §702.171 Saddle mechanic in keywords_saddle.go /
// keywords_batch4.go.
//
// Surface:
//
//   - IsMount(card)                     — true iff the card has the
//                                         "mount" subtype on Types
//                                         (or TypeLine fallback).
//   - PermIsMount(perm)                 — Permanent-level convenience.
//   - CountMountsControlled(gs, seat)   — number of Mount permanents
//                                         controlled by seat. Doesn't
//                                         require the permanent to be
//                                         a creature right now — a
//                                         Mount that lost its types
//                                         via Humility or similar
//                                         still counts as a Mount per
//                                         CR §301.4 type lookup rules.
//                                         (The OTJ Mount package
//                                         doesn't ship any "lose all
//                                         types" interaction so this
//                                         is a theoretical edge.)
//   - FireMountSaddledTriggers(gs, mount) — fires when a Mount becomes
//                                         saddled. Called from both
//                                         SaddleMount (explicit-tapper
//                                         path) and ActivateSaddle
//                                         (greedy auto-tap path).
//                                         No-op when `mount` isn't a
//                                         Mount, so wiring is safe
//                                         even if a non-mount somehow
//                                         carries the Saddle ability
//                                         (test (e)).

// IsMount returns true if the card's printed type line includes the
// Mount subtype. Detection delegates to cardHasSubtype, which checks
// both Card.Types (canonical for runtime cards) and Card.TypeLine
// (fallback for corpus-loaded cards that haven't been re-parsed into
// the Types slice).
func IsMount(card *Card) bool {
	return cardHasSubtype(card, "mount")
}

// PermIsMount returns true if the permanent's card is a Mount. Nil-
// safe — returns false for nil perm or nil perm.Card.
func PermIsMount(perm *Permanent) bool {
	if perm == nil {
		return false
	}
	return IsMount(perm.Card)
}

// CountMountsControlled returns the number of Mount permanents
// controlled by seatIdx on the battlefield. Excludes phased-out
// permanents per CR §702.26 ("treated as though they don't exist").
//
// Returns 0 for nil game / out-of-range seat.
func CountMountsControlled(gs *GameState, seatIdx int) int {
	if gs == nil || seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return 0
	}
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return 0
	}
	n := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.PhasedOut {
			continue
		}
		if PermIsMount(p) {
			n++
		}
	}
	return n
}

// FireMountSaddledTriggers fires the "a Mount you control becomes
// saddled" trigger fan-out. Called from the Saddle success path
// (SaddleMount + ActivateSaddle) AFTER the saddled flag is set, so
// observers reading PermIsSaddled(mount) inside their trigger handler
// see the up-to-date state.
//
// No-op when `mount` isn't a Mount — this gates the trigger so a
// non-Mount with a granted Saddle ability (test (e)) does NOT
// publish a mount_saddled event.
//
// Emits:
//
//   - "mount_saddled" log event with rule citation
//   - "mount_saddled" FireCardTrigger for per_card observers that
//     watch for "whenever a Mount you control becomes saddled" payoffs
func FireMountSaddledTriggers(gs *GameState, mount *Permanent) {
	if gs == nil || mount == nil || mount.Card == nil {
		return
	}
	if !PermIsMount(mount) {
		return
	}
	gs.LogEvent(Event{
		Kind:   "mount_saddled",
		Seat:   mount.Controller,
		Source: mount.Card.DisplayName(),
		Details: map[string]interface{}{
			"rule": "702.171",
		},
	})
	FireCardTrigger(gs, "mount_saddled", map[string]interface{}{
		"mount":      mount,
		"controller": mount.Controller,
	})
}
