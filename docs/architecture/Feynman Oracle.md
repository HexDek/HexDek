# Feynman Oracle

> Source: `internal/hat/feynman.go` (307 lines)
> Status: **Production** — runs post-game in all 3 game paths

The Feynman Oracle validates every completed game against the comprehensive rules. Eight invariant checks catch engine bugs, handler errors, and impossible states that would contaminate training data or produce wrong ratings.

Named after Richard Feynman — "the first principle is that you must not fool yourself, and you are the easiest person to fool."

## Invariant Checks

```go
func CheckGame(gs *GameState) OracleResult
```

| # | Check | CR Reference | Severity | What it catches |
|---|-------|-------------|----------|-----------------|
| 1 | `checkLifeSBA` | §704.5a | critical/info | Player at 0 life not marked Lost (downgrades to "info" if cant-lose effect detected) |
| 2 | `checkToughnessSBA` | §704.5f | critical | Living creature with toughness ≤ 0 |
| 3 | `checkPoisonSBA` | §704.5c | critical | Player with 10+ poison not marked Lost |
| 4 | `checkCommanderDamageSBA` | §704.5v | critical | Player with 21+ commander damage not marked Lost |
| 5 | `checkZoneAccounting` | — | warning | Total cards per seat outside expected range (±3 tolerance for tokens, copy effects, cards-left-game) |
| 6 | `checkExactlyOneWinner` | — | warning | Not exactly N-1 seats Lost at game end (catches SBA-cap draws and turn-cap timeouts) |
| 7 | `checkTurnBounds` | — | warning | Game exceeded 200 turns (possible infinite loop) |
| 8 | `checkNoNegativeCounters` | — | warning | Any permanent with a negative counter value |

## Violation Structure

```go
type OracleViolation struct {
    Rule        string                 // e.g. "704.5f"
    Description string
    Seat        int                    // -1 for game-wide
    Severity    string                 // "critical", "warning", "info"
    Details     map[string]interface{}
}

type OracleResult struct {
    Violations []OracleViolation
    GameTurns  int
    Checked    int
}
```

## Cant-Lose Detection

```go
func hasCantLoseEffect(gs *GameState, seat int) bool
```

Checks `gs.Replacements` for `"would_lose_game"` replacement effects on a seat (Platinum Angel, Lich's Mastery). When detected, §704.5a violations downgrade from "critical" to "info" — the player is legitimately at 0 life but protected.

## Zone Accounting

The trickiest check. Expected card count per seat is the starting deck size (typically 100 for Commander). Tolerance of ±3 accounts for:

- Tokens created during the game (not real cards)
- Copy effects creating additional permanents
- `cards_left_game` tracking from `HandleSeatElimination` (§800.4a — eliminated player's cards leave)

Positive diffs from prolific copy/clone effects are known outliers and logged at "warning" level.

## Integration

- Called by the tournament runner after the game loop exits, before cleanup
- Violations logged to stdout with `feynman:` prefix for grep-ability
- Critical violations flag the game as invalid for neural training data
- Zero-tolerance for critical violations in the test suite (Thor/Odin/Loki use Feynman as a regression gate)
- Warning-level violations are tracked for monitoring but don't invalidate games

## Known Acceptable Violations

Two categories of expected warnings:

1. **game_end violations** (2 of 3 expected deaths): occur when games hit the turn cap or SBA iteration cap, ending with fewer eliminations than a natural game
2. **Zone accounting positive diffs**: copy/clone effects create cards that don't exist in the original deck, producing small positive diffs (typically 4–8 cards)

Both are logged for monitoring but don't indicate engine bugs.

## Related Docs

- [State-Based Actions](State-Based%20Actions.md) — the SBAs being validated
- [Self-Play Loop](Self-Play%20Loop.md) — training data gated by Feynman
- [Tesla Causal Pivots](Tesla%20Causal%20Pivots.md) — runs after Feynman validates
- [Tool - Odin](Tool%20-%20Odin.md) — the fuzz checker that also uses invariants
