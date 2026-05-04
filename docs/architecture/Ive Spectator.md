# Ive Spectator

> Source: `internal/hat/ive.go` (236 lines)
> Status: **Production** — generates narratives for showmatch games

The Ive Spectator transforms raw game events into a three-act narrative structure. It takes Tesla's causal pivot and builds a human-readable story around it — setup, conflict, resolution.

Named after Jony Ive — making the complex feel inevitable through deliberate structure.

## Three-Act Structure

Every game narrative is divided around the Tesla pivot:

```
Act 1: Setup        │  turns 1 to pivot-2
Act 2: Conflict     │  pivot ± 2 turns
Act 3: Resolution   │  pivot+2 to end
```

```go
type GameNarrative struct {
    Acts       [3]Act
    Highlights []Highlight
    Winner     string
    WinnerSeat int
    TotalTurns int
    Pivot      CausalPivot
    Synopsis   string
}

type Act struct {
    Name      string  // "Setup" / "Conflict" / "Resolution"
    StartTurn int
    EndTurn   int
    Summary   string
}
```

## Highlights

```go
type Highlight struct {
    Turn        int
    Seat        int
    Kind        string  // "first_blood", "pivot", "elimination", "combo", "boardwipe"
    Description string
}
```

`extractHighlights()` scans all game events for notable moments:

| Kind | Detection |
|------|----------|
| `first_blood` | First damage event in the game |
| `pivot` | Always included — the Tesla causal pivot turn |
| `elimination` | Player eliminated (Lost = true transition) |
| `boardwipe` | Destroy/exile affecting 3+ creatures |
| `combo` | Combo assessment flipped to Executable |

## Composition

```go
func ComposeNarrative(pivot CausalPivot, events []GameEvent,
    seatNames [4]string, winnerSeat int, totalTurns int) GameNarrative
```

### Act 1: Setup
Summarizes who ramped fastest by comparing land counts at the pivot boundary. "Player X established an early mana advantage with N lands by turn T."

### Act 2: Conflict
Scans the ±2 turn window around the pivot for boardwipes, eliminations, and combo assemblies. "The game turned on turn T when Player X [action]."

### Act 3: Resolution
Notes how quickly the winner closed after the pivot. "Player X closed the game in N turns after the decisive moment."

### Synopsis
One sentence: "4-player game over T turns. X won after a pivotal turn P."

## Game Events

```go
type GameEvent struct {
    Turn   int
    Seat   int
    Kind   string  // "damage", "cast", "destroy", "exile", "eliminate", "combo", "boardwipe"
    Source string  // card name
    Amount int     // damage dealt, creatures destroyed, etc.
}
```

Events are simplified from the engine's full event stream — Ive doesn't need stack resolution details, just the narrative-relevant actions.

## Integration

- Called post-game by the tournament runner after Tesla's `ExtractPivot()`
- Output is JSON-serializable `GameNarrative` sent to spectator UI via WebSocket
- Showmatch path broadcasts the narrative on game end
- The pivot from Tesla is the structural spine — without it, Ive falls back to equal-thirds division

## Related Docs

- [Tesla Causal Pivots](Tesla%20Causal%20Pivots.md) — provides the pivot that structures the narrative
- [Tool - Heimdall](Tool%20-%20Heimdall.md) — observation routing
- [Tool - Server](Tool%20-%20Server.md) — WebSocket broadcast to spectators
