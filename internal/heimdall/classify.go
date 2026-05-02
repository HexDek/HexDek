package heimdall

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// ClassifyKill determines how the winner won the game by inspecting the
// LossReason of eliminated seats. Falls back to heuristic checks when
// LossReason is empty.
//
// Returns one of: "poison", "commander", "mill", "combat", "timeout", "combo".
func ClassifyKill(gs *gameengine.GameState, winner int) string {
	if gs == nil {
		return "combat"
	}

	for i, seat := range gs.Seats {
		if seat == nil || i == winner || !seat.Lost {
			continue
		}

		reason := strings.ToLower(seat.LossReason)

		// Check LossReason first -- the engine sets this with high fidelity.
		switch {
		case strings.Contains(reason, "poison"):
			return "poison"
		case strings.Contains(reason, "commander_damage") || strings.Contains(reason, "commander damage"):
			return "commander"
		case strings.Contains(reason, "empty library") || strings.Contains(reason, "704.5b"):
			return "mill"
		case strings.Contains(reason, "infinite") || strings.Contains(reason, "combo"):
			return "combo"
		}

		// Heuristic fallback: check game state directly.

		// Poison kill: 10+ poison counters.
		if seat.PoisonCounters >= 10 {
			return "poison"
		}

		// Commander damage kill: 21+ from any single commander.
		for _, cmdrDmg := range seat.CommanderDamage {
			for _, dmg := range cmdrDmg {
				if dmg >= 21 {
					return "commander"
				}
			}
		}

		// Mill kill: lost with positive life and no poison/commander kill.
		if seat.Life > 0 && seat.PoisonCounters < 10 {
			allCmdrBelow := true
			for _, cmdrDmg := range seat.CommanderDamage {
				for _, dmg := range cmdrDmg {
					if dmg >= 21 {
						allCmdrBelow = false
						break
					}
				}
				if !allCmdrBelow {
					break
				}
			}
			if allCmdrBelow {
				return "mill"
			}
		}
	}

	// Default: combat damage reduced life to 0.
	return "combat"
}

// ClassifyKillWithMaxTurns is like ClassifyKill but also detects timeout
// when the game hit the turn limit without a natural winner.
func ClassifyKillWithMaxTurns(gs *gameengine.GameState, winner, maxTurns int) string {
	if gs != nil && maxTurns > 0 && gs.Turn >= maxTurns {
		return "timeout"
	}
	return ClassifyKill(gs, winner)
}
