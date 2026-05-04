# Tesla Causal Pivots

> Source: `internal/hat/tesla.go` (164 lines)
> Status: **Production** — runs post-game in all 3 game paths

Tesla extracts the single most game-deciding moment from each completed game. This "causal pivot" identifies the turn where the winner's position swung most dramatically relative to the field — the moment the game was won.

Named after Nikola Tesla — finding the invisible current that drives the visible outcome.

## Pivot Extraction

```go
type CausalPivot struct {
    Turn       int
    WinnerSeat int
    DeltaScore float64
}
```

`ExtractPivot()` scans the eval history collected during the game:

1. For each turn, compute the winner's score delta (score[t] - score[t-1])
2. Subtract the field average delta (other seats' mean change)
3. The turn with the maximum relative swing is the pivot

```
Winner eval:  0.2  0.3  0.3  0.7  0.8  0.9
Field avg:    0.4  0.4  0.4  0.3  0.2  0.1
Relative Δ:   -   +0.2 -0.1 +0.5 +0.0 +0.0
                              ↑
                         Pivot: Turn 4
```

Falls back to game midpoint if fewer than 2 data points exist.

## Training Sample Enrichment

The pivot's primary purpose is improving neural evaluator training. Samples near the pivot are more instructive than samples from quiet board development turns.

```go
type PivotLabel struct {
    PivotTurn     int
    PivotDistance  float64  // |sample_turn - pivot_turn| / game_turns, normalized [0,1]
    IsPostPivot   bool
}

type PivotEnrichedSample struct {
    // TrainingSample fields +
    PivotTurn     int
    PivotDistance  float64
    IsPostPivot   bool
}
```

`LabelSamplesWithPivot()` stamps each training sample with its distance from the pivot. The PyTorch training script uses `pivot_distance` for weighted MSE loss — samples close to the pivot get higher loss weight, teaching the model to recognize *when* positions change rather than just who wins.

## Data Collection

```go
type EvalSnapshotCollector struct {
    history map[int][]float64  // turn → [4]float64 per-seat evals
}
```

The collector records all 4 seats' evaluator scores at each turn. This is separate from the `TrainingCollector` (which snapshots every 5 turns for state vectors) — the eval snapshot collector captures every turn for pivot detection precision.

```go
func (c *EvalSnapshotCollector) Record(turn int, scores []float64)
func (c *EvalSnapshotCollector) History() map[int][]float64
```

## Post-Game Pipeline

```
Game ends
    │
    ▼
EvalSnapshotCollector.History() → turn-by-turn eval scores
    │
    ▼
ExtractPivot(history, winnerSeat, gameTurns) → CausalPivot
    │
    ├──→ LabelSamplesWithPivot(trainingSamples, pivot)
    │        │
    │        ▼
    │    EnrichSamples(samples, labels) → []PivotEnrichedSample
    │        │
    │        ▼
    │    SelfPlayManager.WriteEnrichedSamples() → samples.jsonl
    │
    └──→ Ive.ComposeNarrative(pivot, ...) → GameNarrative
```

The pivot feeds both training (neural evaluator gets better at recognizing decisive moments) and spectating (Ive builds a three-act narrative around it).

## Related Docs

- [Neural Evaluator](Neural%20Evaluator.md) — consumer of enriched samples
- [Self-Play Loop](Self-Play%20Loop.md) — training coordination
- [Ive Spectator](Ive%20Spectator.md) — narrative built around pivots
- [Feynman Oracle](Feynman%20Oracle.md) — validates game before pivot extraction
