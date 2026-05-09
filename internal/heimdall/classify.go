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

// SnapshotTurnCounters captures the per-seat TurnCounters into a fixed-size
// array suitable for embedding in GameSeed. Seats beyond index 3 are ignored
// (HexDek is 4-player Commander). Nil seats produce a zero snapshot.
//
// NOTE: TurnCounters are reset every turn (see Reset() in UntapAll), so this
// captures only the FINAL turn's totals — not game-wide cumulative values.
// Callers that want game-wide aggregates must accumulate during play.
func SnapshotTurnCounters(gs *gameengine.GameState) [4]TurnCounterSnapshot {
	var out [4]TurnCounterSnapshot
	if gs == nil {
		return out
	}
	for i, seat := range gs.Seats {
		if i >= 4 || seat == nil {
			continue
		}
		out[i] = snapshotOne(&seat.Turn)
	}
	return out
}

func snapshotOne(tc *gameengine.TurnCounters) TurnCounterSnapshot {
	if tc == nil {
		return TurnCounterSnapshot{}
	}
	return TurnCounterSnapshot{
		LifeGained:          tc.LifeGained,
		LifeLost:            tc.LifeLost,
		DamageReceived:      tc.DamageReceived,
		LifePaid:            tc.LifePaid,
		CardsDrawn:          tc.CardsDrawn,
		SpellsCast:          tc.SpellsCast,
		CreaturesEntered:    tc.CreaturesEntered,
		ArtifactsEntered:    tc.ArtifactsEntered,
		EnchantmentsEntered: tc.EnchantmentsEntered,
		TokensCreated:       tc.TokensCreated,
		TreasuresCreated:    tc.TreasuresCreated,
		Sacrificed:          tc.Sacrificed,
		PermanentsLeft:      tc.PermanentsLeft,
		Discarded:           tc.Discarded,
		Milled:              tc.Milled,
		LandsPlayed:         tc.LandsPlayed,
		CreaturesDied:       tc.CreaturesDied,
		ExiledCards:         tc.ExiledCards,
		CastFromExile:       tc.CastFromExile,
		Descended:           tc.Descended,
		Attacked:            tc.Attacked,
	}
}
