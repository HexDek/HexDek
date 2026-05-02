package hat

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// ComboSequencer evaluates whether a seat can execute a combo win
// this turn. It reads Freya's combo packages (via StrategyProfile)
// and checks the player's zones for pieces and mana feasibility.
type ComboSequencer struct {
	Lines []ComboConstraint // loaded from Freya's combo packages
}

// NewComboSequencer builds a sequencer from a StrategyProfile's combo
// data. Returns nil when the profile has no combo plans, which callers
// treat as "no combos to evaluate."
func NewComboSequencer(sp *StrategyProfile) *ComboSequencer {
	if sp == nil || len(sp.ComboPieces) == 0 {
		return nil
	}
	lines := make([]ComboConstraint, 0, len(sp.ComboPieces))
	for _, cp := range sp.ComboPieces {
		if len(cp.Pieces) == 0 {
			continue
		}
		cc := ComboConstraint{
			Name:         comboName(cp),
			PiecesNeeded: cp.Pieces,
			ZonesAccepted: buildZonesAccepted(cp),
			ManaRequired: estimateManaRequired(cp),
			SequenceOrder: buildSequenceOrder(cp),
		}
		// Infinite combos often need protection to stick.
		if cp.Type == "infinite" {
			cc.NeedsProtection = true
		}
		lines = append(lines, cc)
	}
	if len(lines) == 0 {
		return nil
	}
	return &ComboSequencer{Lines: lines}
}

// comboName generates a readable name from the combo plan.
func comboName(cp ComboPlan) string {
	if len(cp.Pieces) == 0 {
		return "unknown"
	}
	parts := make([]string, len(cp.Pieces))
	copy(parts, cp.Pieces)
	name := strings.Join(parts, " + ")
	if cp.Type != "" {
		name += " (" + cp.Type + ")"
	}
	return name
}

// buildZonesAccepted creates the zone acceptance map. Most combo
// pieces need to be either in hand (castable) or on the battlefield
// (already resolved). Graveyard is accepted as a secondary source
// for reanimation-style combos. This is a practical 80% default.
func buildZonesAccepted(cp ComboPlan) map[string][]string {
	zones := make(map[string][]string, len(cp.Pieces))
	for _, piece := range cp.Pieces {
		zones[piece] = []string{"hand", "battlefield", "graveyard"}
	}
	return zones
}

// buildSequenceOrder returns the cast order for the combo, falling
// back to piece order if no explicit CastOrder is provided.
func buildSequenceOrder(cp ComboPlan) []string {
	if len(cp.CastOrder) > 0 {
		return cp.CastOrder
	}
	return cp.Pieces
}

// estimateManaRequired estimates the total mana needed to execute a
// combo from hand. We sum the costs of pieces that aren't already on
// the battlefield; since we don't have board state at construction
// time, we use len(pieces) * 2 as a rough heuristic (most combo
// pieces are 2-4 CMC). The Evaluate method does the real check.
func estimateManaRequired(cp ComboPlan) int {
	// Rough estimate: 2 mana per piece. Real check happens at eval time.
	return len(cp.Pieces) * 2
}

// Evaluate checks all combo lines against the current game state for
// a seat. Returns a ComboAssessment describing the best available
// combo line and whether it can be executed this turn.
func (cs *ComboSequencer) Evaluate(gs *gameengine.GameState, seatIdx int) ComboAssessment {
	if cs == nil || gs == nil || seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return ComboAssessment{}
	}
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return ComboAssessment{}
	}

	// Build zone indexes: card name -> set of zones it appears in.
	zoneIndex := buildZoneIndex(seat)
	availMana := gameengine.AvailableManaEstimate(gs, seat)

	var best *comboEval
	for i := range cs.Lines {
		ev := evaluateLine(&cs.Lines[i], zoneIndex, seat, availMana)
		if best == nil || ev.betterThan(best) {
			best = ev
		}
	}

	if best == nil {
		return ComboAssessment{}
	}

	result := ComboAssessment{
		BestLine:    best.line,
		PiecesFound: best.found,
		PiecesTotal: len(best.line.PiecesNeeded),
	}

	if best.executable {
		result.Executable = true
		result.NextAction = best.nextAction
	} else if best.missing == 1 && hasTutorInHand(seat) {
		result.Assembling = true
		result.MissingPiece = best.missingPiece
	}

	return result
}

// comboEval is the internal evaluation result for a single combo line.
type comboEval struct {
	line         *ComboConstraint
	found        int    // pieces found in acceptable zones
	missing      int    // pieces not found
	executable   bool   // all pieces + mana available
	nextAction   string // first piece to cast from hand
	missingPiece string // name of the first missing piece (for tutor)
	manaOK       bool   // enough mana to cast remaining pieces from hand
}

// betterThan returns true if this eval is strictly better than other.
// Priority: executable > more pieces found > fewer pieces needed.
func (e *comboEval) betterThan(other *comboEval) bool {
	if e.executable != other.executable {
		return e.executable
	}
	// Higher completion ratio is better.
	eRatio := float64(e.found) / float64(max(len(e.line.PiecesNeeded), 1))
	oRatio := float64(other.found) / float64(max(len(other.line.PiecesNeeded), 1))
	if eRatio != oRatio {
		return eRatio > oRatio
	}
	// Tie-break: fewer total pieces (simpler combo) wins.
	return len(e.line.PiecesNeeded) < len(other.line.PiecesNeeded)
}

// zoneSet tracks which zones a card name appears in.
type zoneSet struct {
	hand        bool
	battlefield bool
	graveyard   bool
}

// buildZoneIndex scans a seat's hand, battlefield, and graveyard and
// returns a map of card name -> which zones that card appears in.
func buildZoneIndex(seat *gameengine.Seat) map[string]*zoneSet {
	idx := make(map[string]*zoneSet)

	ensure := func(name string) *zoneSet {
		if z, ok := idx[name]; ok {
			return z
		}
		z := &zoneSet{}
		idx[name] = z
		return z
	}

	for _, c := range seat.Hand {
		if c == nil {
			continue
		}
		ensure(c.Name).hand = true
	}
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		ensure(p.Card.Name).battlefield = true
	}
	for _, c := range seat.Graveyard {
		if c == nil {
			continue
		}
		ensure(c.Name).graveyard = true
	}
	return idx
}

// evaluateLine checks a single combo line against the zone index and
// available mana.
func evaluateLine(line *ComboConstraint, zoneIndex map[string]*zoneSet, seat *gameengine.Seat, availMana int) *comboEval {
	ev := &comboEval{
		line: line,
	}

	// Track mana needed to cast pieces that are in hand (not yet on battlefield).
	manaCost := 0

	for _, piece := range line.PiecesNeeded {
		z, exists := zoneIndex[piece]
		if !exists {
			ev.missing++
			if ev.missingPiece == "" {
				ev.missingPiece = piece
			}
			continue
		}

		// Check if the piece is in an acceptable zone for this combo.
		accepted := line.ZonesAccepted[piece]
		inAcceptableZone := false
		needsCast := false

		for _, zone := range accepted {
			switch zone {
			case "hand":
				if z.hand {
					inAcceptableZone = true
					needsCast = true
				}
			case "battlefield":
				if z.battlefield {
					inAcceptableZone = true
					// Already on battlefield — no cast needed.
					needsCast = false
					break // battlefield is ideal, stop checking
				}
			case "graveyard":
				if z.graveyard {
					inAcceptableZone = true
					// In graveyard — may need recursion, treat as
					// needing a cast for mana purposes.
					needsCast = true
				}
			}
			// If on battlefield, that's the best case — break early.
			if z.battlefield && inAcceptableZone {
				break
			}
		}

		if inAcceptableZone {
			ev.found++
			if needsCast {
				manaCost += cardManaCost(seat, piece)
			}
		} else {
			ev.missing++
			if ev.missingPiece == "" {
				ev.missingPiece = piece
			}
		}
	}

	ev.manaOK = availMana >= manaCost
	ev.executable = ev.missing == 0 && ev.manaOK

	// Determine next action: first piece in sequence order that is in
	// hand (needs to be cast).
	if ev.executable {
		for _, name := range line.SequenceOrder {
			z, ok := zoneIndex[name]
			if !ok {
				continue
			}
			if z.hand && !z.battlefield {
				ev.nextAction = name
				break
			}
		}
		// If all pieces are on battlefield (activated combo), the next
		// action is the first piece in sequence order.
		if ev.nextAction == "" && len(line.SequenceOrder) > 0 {
			ev.nextAction = line.SequenceOrder[0]
		}
	}

	return ev
}

// cardManaCost looks up the mana cost for a card by name in the
// seat's hand. Returns 0 if not found (tokens, free spells).
func cardManaCost(seat *gameengine.Seat, name string) int {
	for _, c := range seat.Hand {
		if c != nil && c.Name == name {
			return gameengine.ManaCostOf(c)
		}
	}
	return 0
}

// hasTutorInHand checks if the seat has any tutor card in hand.
// Uses two detection methods:
//   1. StrategyProfile CardRoles (if the seat's Hat has one)
//   2. Oracle text heuristic: "search your library"
func hasTutorInHand(seat *gameengine.Seat) bool {
	for _, c := range seat.Hand {
		if c == nil {
			continue
		}
		// Heuristic: check oracle text for tutor-like effects.
		ot := gameengine.OracleTextLower(c)
		if strings.Contains(ot, "search your library") {
			return true
		}
		// Also check AST effect kinds for "tutor" kind.
		if isTutor(c) {
			return true
		}
	}
	return false
}

