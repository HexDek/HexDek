package gameengine

import (
	"regexp"
	"strings"

	"github.com/hexdek/hexdek/internal/gameast"
)

// affinityForRe extracts the type clause from an "affinity for <type>"
// keyword. Handles both the parsed-keyword form ("affinity for humans"
// with name == that whole string) and the raw-text form ("affinity for
// humans (this spell costs {1} less...)"). Captures the type word
// (singular form) for case-insensitive type matching.
//
// Examples:
//
//	"affinity for artifacts"        → "artifact"
//	"affinity for humans"           → "human"
//	"affinity for artifact creatures" → "artifact creature" (handled in CR §702.41
//	                                   note as a compound type — special-cased
//	                                   downstream)
var affinityForRe = regexp.MustCompile(`(?i)affinity\s+for\s+([a-z][a-z\s]*?)s?\b`)

// AffinityForType returns (hasIt, typeStr) for any "affinity for <type>"
// keyword found on the card. The returned typeStr is lowercase and
// singular (matches the convention of cardHasType / Card.Types entries).
//
// Replaces the previous N specific HasAffinityFor* functions with one
// parameterless detector that returns WHICH type matters. Callers count
// matching permanents on the controller's battlefield to compute the
// cost reduction.
//
// Examples:
//
//	Frogmite                                  → (true, "artifact")
//	Witherbloom, the Balancer                 → (true, "creature")
//	Urza, Chief Artificer                     → (true, "artifact creature")
//	Riders of the Mark (Affinity for Humans)  → (true, "human")
//	Anything without an affinity keyword      → (false, "")
//
// Multiple affinity-for keywords on the same card: returns the first
// match. Cards with multiple distinct affinity types are exceedingly
// rare in printed Magic; if WotC ever prints one, this would need to
// return a slice instead.
func AffinityForType(card *Card) (bool, string) {
	if card == nil || card.AST == nil {
		return false, ""
	}
	for _, ab := range card.AST.Abilities {
		kw, ok := ab.(*gameast.Keyword)
		if !ok {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(kw.Name))
		raw := strings.ToLower(strings.TrimSpace(kw.Raw))
		// Form 1: keyword name itself encodes the type
		// e.g. "affinity for humans"
		if strings.HasPrefix(name, "affinity for ") {
			typeStr := strings.TrimPrefix(name, "affinity for ")
			typeStr = strings.TrimSuffix(typeStr, "s") // singularize
			typeStr = strings.TrimSpace(typeStr)
			if typeStr != "" {
				return true, typeStr
			}
		}
		// Form 2: name is just "affinity", raw text carries the type
		if name == "affinity" {
			if m := affinityForRe.FindStringSubmatch(raw); m != nil {
				typeStr := strings.TrimSpace(strings.ToLower(m[1]))
				if typeStr != "" {
					return true, typeStr
				}
			}
		}
	}
	return false, ""
}

// CountPermanentsByType counts permanents on seatIdx's battlefield that
// have the given type (uses c.Types comprehensive slice — covers super
// types, main types, and subtypes uniformly).
//
// Generic counterpart to CountArtifacts / CountCreaturesOnBattlefield.
// Callers from the cost-modifier path use this to compute the per-type
// affinity reduction for any "affinity for <type>" keyword.
//
// Special case: "artifact creature" — the compound type used by Urza,
// Chief Artificer. A permanent qualifies only if it is BOTH an artifact
// AND a creature. The function detects the space in the typeStr and
// splits on it, requiring all component types to match.
func CountPermanentsByType(gs *GameState, seatIdx int, typeStr string) int {
	if gs == nil || seatIdx < 0 || seatIdx >= len(gs.Seats) || typeStr == "" {
		return 0
	}
	parts := strings.Fields(strings.ToLower(strings.TrimSpace(typeStr)))
	if len(parts) == 0 {
		return 0
	}
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return 0
	}
	count := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		ok := true
		for _, want := range parts {
			if !permanentHasTypeLower(p, want) {
				ok = false
				break
			}
		}
		if ok {
			count++
		}
	}
	return count
}

// permanentHasTypeLower is a package-private case-insensitive type
// check against a permanent's card types. Lives next to the generic
// affinity helpers so callers don't have to import per_card's helper
// of the same name.
func permanentHasTypeLower(p *Permanent, want string) bool {
	if p == nil || p.Card == nil {
		return false
	}
	w := strings.ToLower(want)
	for _, got := range p.Card.Types {
		if strings.ToLower(got) == w {
			return true
		}
	}
	return false
}
