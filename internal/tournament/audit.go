package tournament

import (
	"time"

	"github.com/hexdek/hexdek/internal/deckparser"
)

// RunOneGameForAudit is an exported wrapper around runOneGameSafe so
// the cmd/tournament-audit one-shot auditor can drive its own pool loop
// while still piggy-backing on the engine's recover()/timeout harness.
//
// Contract sealing is intentionally disabled (zero-value contractParams);
// audit runs do not participate in seedcontract chains.
func RunOneGameForAudit(
	gameIdx int,
	decks []*deckparser.TournamentDeck,
	hats []HatFactory,
	nSeats int,
	masterSeed int64,
	maxTurns int,
	gameTimeoutSec int,
	commanderMode, auditEnabled, analyticsEnabled bool,
) GameOutcome {
	timeout := 180 * time.Second
	if gameTimeoutSec > 0 {
		timeout = time.Duration(gameTimeoutSec) * time.Second
	}
	if maxTurns <= 0 {
		maxTurns = defaultMaxTurns
	}
	return runOneGameSafe(
		gameIdx, decks, hats, nSeats, masterSeed, maxTurns, timeout,
		commanderMode, auditEnabled, analyticsEnabled, contractParams{},
	)
}
