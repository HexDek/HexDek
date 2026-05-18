package gameengine

// keywords_outlaw.go — Outlaw type-group surface (CR §205.3m, Outlaws
// of Thunder Junction 2024).
//
// CR §205.3m: "Outlaw" is a creature-type GROUP, not a subtype itself.
//              It comprises five subtypes: Assassin, Mercenary, Pirate,
//              Rogue, and Warlock. A creature is an "Outlaw" if its
//              type line contains at least one of those five subtypes.
//              Card text that refers to an "Outlaw" matches any
//              creature with one of those subtypes (an Assassin Rogue
//              counts once, not twice).
//
// Many OTJ-era cards key off Outlaws:
//
//   - "Whenever an Outlaw enters the battlefield under your control, do Y"
//     (Bonny Pall, Clearcutter; Annie Joins Up; Knockout Punch)
//   - "Outlaws you control get +1/+1" / "for each Outlaw you control"
//     (Trained Arynx; Vraska, the Silencer; Magda, the Hoardmaster)
//   - Cost reduction for casting an Outlaw spell (Freerunning's gate)
//
// Engine surface:
//
//   - IsOutlaw(card) bool                  — type-line predicate
//   - PermIsOutlaw(perm) bool              — battlefield-side wrapper
//   - CountOutlawsControlled(gs, seat) int — battlefield tally
//   - HasOutlawTrigger(card) bool          — oracle detector for
//                                            "an Outlaw enters" patterns
//   - FireOutlawETBTriggers(gs, perm)      — fan-out hook wired into
//                                            FirePermanentETBTriggers
//
// The five Outlaw subtypes:

import "strings"

// outlawSubtypes is the canonical list per CR §205.3m. Lowercased for
// matching against Card.Types and Card.TypeLine (both normalized
// lowercase elsewhere in the engine).
var outlawSubtypes = []string{"assassin", "mercenary", "pirate", "rogue", "warlock"}

// OutlawSubtypes returns the canonical CR §205.3m subtype list. Useful
// for tooling (Heimdall, Freya) that wants to enumerate the set
// rather than hardcode it.
func OutlawSubtypes() []string {
	return append([]string(nil), outlawSubtypes...)
}

// ---------------------------------------------------------------------------
// IsOutlaw / PermIsOutlaw
// ---------------------------------------------------------------------------

// IsOutlaw reports whether the card has at least one of the five
// Outlaw subtypes (Assassin, Mercenary, Pirate, Rogue, Warlock). Scans
// Card.Types and falls through to a TypeLine substring check so cards
// whose loader populated only the type line still match.
//
// CR §205.3m: a creature with multiple Outlaw subtypes still counts as
// ONE outlaw — the helper returns true on first match.
func IsOutlaw(card *Card) bool {
	if card == nil {
		return false
	}
	for _, t := range card.Types {
		tl := strings.ToLower(t)
		for _, want := range outlawSubtypes {
			if tl == want {
				return true
			}
		}
	}
	// TypeLine fallback: match whole-word against the printed line so
	// "Pirate" inside "Spirate" wouldn't false-positive (the printed
	// type lines from the corpus are space-separated subtypes).
	if card.TypeLine != "" {
		tl := strings.ToLower(card.TypeLine)
		for _, want := range outlawSubtypes {
			if containsWord(tl, want) {
				return true
			}
		}
	}
	return false
}

// containsWord returns true if `text` contains `word` as a whole token
// (separated by spaces, punctuation, or line boundaries). Used by the
// TypeLine fallback to avoid the "Pirate" substring matching "Spirate"
// or similar.
func containsWord(text, word string) bool {
	idx := 0
	for {
		i := strings.Index(text[idx:], word)
		if i < 0 {
			return false
		}
		start := idx + i
		end := start + len(word)
		left := start == 0 || !isWordChar(text[start-1])
		right := end == len(text) || !isWordChar(text[end])
		if left && right {
			return true
		}
		idx = end
	}
}

func isWordChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') || b == '_'
}

// PermIsOutlaw is the battlefield-side wrapper. Operates on the
// permanent's active face (perm.Card.AST is kept in sync with
// perm.Transformed by TransformPermanent).
func PermIsOutlaw(perm *Permanent) bool {
	if perm == nil {
		return false
	}
	return IsOutlaw(perm.Card)
}

// ---------------------------------------------------------------------------
// CountOutlawsControlled
// ---------------------------------------------------------------------------

// CountOutlawsControlled returns the number of permanents on
// `seatIdx`'s battlefield that satisfy PermIsOutlaw. Backs cards like
// "for each Outlaw you control."
func CountOutlawsControlled(gs *GameState, seatIdx int) int {
	if gs == nil || seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return 0
	}
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return 0
	}
	n := 0
	for _, p := range seat.Battlefield {
		if PermIsOutlaw(p) {
			n++
		}
	}
	return n
}

// ---------------------------------------------------------------------------
// HasOutlawTrigger — oracle detector
// ---------------------------------------------------------------------------

// HasOutlawTrigger reports whether `card`'s oracle text carries a
// "whenever an Outlaw enters" / "when an Outlaw enters" / "when ~ or
// another Outlaw enters" trigger preamble. Used by FireOutlawETBTriggers
// to decide which permanents to fan out to.
func HasOutlawTrigger(card *Card) bool {
	if card == nil {
		return false
	}
	text := strings.ToLower(OracleTextLower(card))
	if text == "" {
		return false
	}
	patterns := []string{
		"whenever an outlaw enters",
		"when an outlaw enters",
		"whenever another outlaw enters",
		"when another outlaw enters",
		"whenever you cast an outlaw spell",
		"whenever an outlaw you control",
	}
	for _, p := range patterns {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// FireOutlawETBTriggers — the ETB fan-out
// ---------------------------------------------------------------------------

// FireOutlawETBTriggers fans out an "outlaw_etb" trigger to every
// permanent on every battlefield whose oracle text watches for outlaw
// ETBs, when `perm` itself qualifies as an Outlaw.
//
// Two contracts:
//   - `perm` must be an Outlaw — if PermIsOutlaw(perm) is false this
//     is a no-op (fast path).
//   - Each watcher is notified via FireCardTrigger("outlaw_etb", ctx)
//     where ctx carries:
//       - "perm":              *Permanent — the entering outlaw
//       - "card":              *Card      — entering outlaw's card
//       - "card_name":         string
//       - "controller_seat":   int        — owner of the new outlaw
//       - "watcher":           *Permanent — the permanent whose
//                                            trigger is firing
//       - "watcher_seat":      int        — that watcher's controller
//
// The "another Outlaw" wording (cards that don't self-trigger) is
// handled by per-card handlers — they read ctx["perm"] and skip
// self-events. The engine fires for every watcher; "another" filtering
// is the watcher card's responsibility.
//
// Wired into FirePermanentETBTriggers (etb_dispatch.go) so every
// battlefield-entry path runs this once.
func FireOutlawETBTriggers(gs *GameState, perm *Permanent) {
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if !PermIsOutlaw(perm) {
		return
	}
	name := perm.Card.DisplayName()
	for _, seat := range gs.Seats {
		if seat == nil {
			continue
		}
		for _, watcher := range seat.Battlefield {
			if watcher == nil || watcher.Card == nil {
				continue
			}
			if !HasOutlawTrigger(watcher.Card) {
				continue
			}
			FireCardTrigger(gs, "outlaw_etb", map[string]interface{}{
				"perm":            perm,
				"card":            perm.Card,
				"card_name":       name,
				"controller_seat": perm.Controller,
				"watcher":         watcher,
				"watcher_seat":    watcher.Controller,
			})
		}
	}
}
