package hat

import (
	"fmt"
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// Feynman Oracle — provably-correct invariant checker.
//
// Runs after each completed game and verifies that the engine's final
// state satisfies fundamental MTG rules invariants. Violations indicate
// engine bugs (SBA missed, combat damage wrong, zone tracking broken).
//
// Named after Feynman's "checking from a different direction" principle:
// the main engine plays forward fast; the oracle checks backward slow.

// OracleViolation describes a single rules invariant that was broken.
type OracleViolation struct {
	Rule        string // CR section (e.g., "704.5f")
	Description string
	Seat        int
	Severity    string // "critical", "warning", "info"
	Details     map[string]interface{}
}

func (v OracleViolation) String() string {
	return fmt.Sprintf("[%s] %s (seat %d): %s", v.Severity, v.Rule, v.Seat, v.Description)
}

// OracleResult is the output of a Feynman check on one completed game.
type OracleResult struct {
	Violations []OracleViolation
	GameTurns  int
	Checked    int // number of invariants checked
}

func (r OracleResult) Clean() bool { return len(r.Violations) == 0 }

// CheckGame runs all Feynman invariants on a completed game state.
// Call after the game loop exits but before cleanup.
func CheckGame(gs *gameengine.GameState) OracleResult {
	result := OracleResult{GameTurns: gs.Turn}
	checks := []func(*gameengine.GameState, *OracleResult){
		checkLifeSBA,
		checkToughnessSBA,
		checkPoisonSBA,
		checkCommanderDamageSBA,
		checkZoneAccounting,
		checkExactlyOneWinner,
		checkTurnBounds,
		checkNoNegativeCounters,
	}
	for _, check := range checks {
		check(gs, &result)
		result.Checked++
	}
	return result
}

// §704.5a — A player with 0 or less life loses the game.
func checkLifeSBA(gs *gameengine.GameState, r *OracleResult) {
	for i, s := range gs.Seats {
		if s == nil {
			continue
		}
		if s.Life <= 0 && !s.Lost {
			r.Violations = append(r.Violations, OracleViolation{
				Rule:        "704.5a",
				Description: fmt.Sprintf("seat %d has %d life but is not marked lost", i, s.Life),
				Seat:        i,
				Severity:    "critical",
				Details:     map[string]interface{}{"life": s.Life},
			})
		}
	}
}

// §704.5f — A creature with toughness 0 or less is put into its
// owner's graveyard. (Check no living creatures with ≤0 toughness.)
func checkToughnessSBA(gs *gameengine.GameState, r *OracleResult) {
	for i, s := range gs.Seats {
		if s == nil || s.Lost {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || !p.IsCreature() {
				continue
			}
			t := gs.ToughnessOf(p)
			if t <= 0 {
				r.Violations = append(r.Violations, OracleViolation{
					Rule: "704.5f",
					Description: fmt.Sprintf("creature %q has toughness %d on battlefield",
						p.Card.DisplayName(), t),
					Seat:     i,
					Severity: "critical",
					Details:  map[string]interface{}{"card": p.Card.DisplayName(), "toughness": t},
				})
			}
		}
	}
}

// §704.5c — A player with 10+ poison counters loses the game.
func checkPoisonSBA(gs *gameengine.GameState, r *OracleResult) {
	for i, s := range gs.Seats {
		if s == nil {
			continue
		}
		if s.PoisonCounters >= 10 && !s.Lost {
			r.Violations = append(r.Violations, OracleViolation{
				Rule:        "704.5c",
				Description: fmt.Sprintf("seat %d has %d poison but is not lost", i, s.PoisonCounters),
				Seat:        i,
				Severity:    "critical",
				Details:     map[string]interface{}{"poison": s.PoisonCounters},
			})
		}
	}
}

// §704.5v — Commander damage: a player who has been dealt 21+ combat
// damage by a single commander loses the game.
func checkCommanderDamageSBA(gs *gameengine.GameState, r *OracleResult) {
	for i, s := range gs.Seats {
		if s == nil || s.CommanderDamage == nil {
			continue
		}
		for _, cmdrMap := range s.CommanderDamage {
			for cmdrName, dmg := range cmdrMap {
				if dmg >= 21 && !s.Lost {
					r.Violations = append(r.Violations, OracleViolation{
						Rule: "704.5v",
						Description: fmt.Sprintf("seat %d has %d commander damage from %s but is not lost",
							i, dmg, cmdrName),
						Seat:     i,
						Severity: "critical",
						Details:  map[string]interface{}{"commander": cmdrName, "damage": dmg},
					})
				}
			}
		}
	}
}

// Zone accounting: every card should be in exactly one zone.
// Total cards per seat should equal starting deck size (99 + commanders).
func checkZoneAccounting(gs *gameengine.GameState, r *OracleResult) {
	for i, s := range gs.Seats {
		if s == nil {
			continue
		}
		total := len(s.Hand) + len(s.Library) + len(s.Graveyard) + len(s.Exile)
		for _, p := range s.Battlefield {
			if p != nil {
				total++
			}
		}
		// Commander zone cards.
		total += len(s.CommandZone)

		// Tokens on battlefield don't count toward the deck total.
		tokens := 0
		for _, p := range s.Battlefield {
			if p != nil && p.IsToken() {
				tokens++
			}
		}
		total -= tokens

		// Expected: 99-card deck + 1-2 commanders = 100-101.
		// Allow some tolerance for edge cases (partner commanders, companion).
		expected := 100
		if len(s.CommanderNames) > 1 {
			expected = 99 + len(s.CommanderNames)
		}

		diff := total - expected
		if diff < -3 || diff > 3 {
			r.Violations = append(r.Violations, OracleViolation{
				Rule: "zone_accounting",
				Description: fmt.Sprintf("seat %d has %d cards (expected ~%d, diff=%d)",
					i, total, expected, diff),
				Seat:     i,
				Severity: "warning",
				Details: map[string]interface{}{
					"hand": len(s.Hand), "library": len(s.Library),
					"graveyard": len(s.Graveyard), "exile": len(s.Exile),
					"battlefield": len(s.Battlefield) - tokens,
					"tokens": tokens, "command_zone": len(s.CommandZone),
				},
			})
		}
	}
}

// Exactly one winner: at game end, exactly N-1 seats should be lost.
func checkExactlyOneWinner(gs *gameengine.GameState, r *OracleResult) {
	lost := 0
	for _, s := range gs.Seats {
		if s != nil && s.Lost {
			lost++
		}
	}
	expected := len(gs.Seats) - 1
	if lost != expected {
		severity := "warning"
		if lost == 0 || lost == len(gs.Seats) {
			severity = "critical"
		}
		r.Violations = append(r.Violations, OracleViolation{
			Rule:        "game_end",
			Description: fmt.Sprintf("%d of %d seats lost (expected %d)", lost, len(gs.Seats), expected),
			Seat:        -1,
			Severity:    severity,
			Details:     map[string]interface{}{"lost": lost, "total": len(gs.Seats)},
		})
	}
}

// Sanity: games shouldn't run more than ~200 turns. Flag runaways.
func checkTurnBounds(gs *gameengine.GameState, r *OracleResult) {
	if gs.Turn > 200 {
		r.Violations = append(r.Violations, OracleViolation{
			Rule:        "turn_bound",
			Description: fmt.Sprintf("game ran %d turns (possible infinite loop)", gs.Turn),
			Seat:        -1,
			Severity:    "warning",
			Details:     map[string]interface{}{"turns": gs.Turn},
		})
	}
}

// No permanent should have negative counter values.
func checkNoNegativeCounters(gs *gameengine.GameState, r *OracleResult) {
	for i, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Counters == nil {
				continue
			}
			for kind, count := range p.Counters {
				if count < 0 {
					r.Violations = append(r.Violations, OracleViolation{
						Rule: "counter_negative",
						Description: fmt.Sprintf("%q has %d %s counters",
							p.Card.DisplayName(), count, kind),
						Seat:     i,
						Severity: "warning",
						Details: map[string]interface{}{
							"card": p.Card.DisplayName(),
							"counter": kind, "count": count,
						},
					})
				}
			}
		}
	}
}

// FormatViolations returns a human-readable summary of all violations.
func FormatViolations(violations []OracleViolation) string {
	if len(violations) == 0 {
		return "Feynman Oracle: all invariants satisfied"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Feynman Oracle: %d violation(s)\n", len(violations))
	for _, v := range violations {
		fmt.Fprintf(&b, "  %s\n", v)
	}
	return b.String()
}
