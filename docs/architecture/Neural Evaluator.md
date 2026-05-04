# Neural Evaluator

> Source: `internal/hat/neural.go` (383 lines)
> Status: **Production** — 80/20 heuristic/neural blend when model exists

The neural evaluator is a multi-layer perceptron that learns position evaluation from self-play data. It complements the hand-coded 20-dimension evaluator by learning patterns the heuristics miss.

## State Vector

Every game state is encoded as a fixed 92-dimensional vector:

```go
const (
    seatFeatures   = 22   // per seat
    globalFeatures = 4    // game-wide
    maxSeats       = 4
    StateVectorDim = 92   // 22 × 4 + 4
)
```

### Per-Seat Features (22 per seat, seat 0 = perspective player)

| # | Feature | Normalization |
|---|---------|--------------|
| 1 | Life | life / 40 |
| 2 | Poison counters | poison / 10 |
| 3 | Commander damage taken | max_cmdr_dmg / 21 |
| 4 | Hand size | cards / 10 |
| 5 | Library size | cards / 99 |
| 6 | Graveyard size | cards / 50 |
| 7 | Exile size | cards / 30 |
| 8 | Land count | lands / 15 |
| 9 | Creature count | creatures / 15 |
| 10 | Artifact count | artifacts / 10 |
| 11 | Enchantment count | enchantments / 10 |
| 12 | Planeswalker count | walkers / 3 |
| 13 | Total power | power / 50 |
| 14 | Total toughness | toughness / 50 |
| 15 | Available mana | mana / 15 |
| 16 | Commander on battlefield | 0 or 1 |
| 17 | Commander tax | tax / 10 |
| 18 | Spells cast this turn | spells / 5 |
| 19 | Has counterspell | 0 or 1 |
| 20 | Has removal | 0 or 1 |
| 21 | Lost | 0 or 1 |
| 22 | Is active player | 0 or 1 |

Seat rotation: the perspective player is always encoded at index 0, opponents rotated into slots 1–3. This makes the model position-invariant.

### Global Features (4)

| # | Feature | Normalization |
|---|---------|--------------|
| 1 | Turn number | turn / 50 |
| 2 | Stack depth | depth / 10 |
| 3 | Game stage | 0.0 (early) / 0.5 (mid) / 1.0 (late) |
| 4 | Living opponent count | alive / 3 |

## Architecture

```
Input (92) → Dense(256, ReLU) → Dense(128, ReLU) → Dense(64, ReLU) → Dense(1, Sigmoid)
```

Multi-layer MLP with ReLU activations on hidden layers, sigmoid on output. The output is a win probability in `[0, 1]`.

### Model Format

```go
type NeuralLayer struct {
    Weights [][]float64  // [out][in]
    Biases  []float64    // [out]
}

type NeuralEvaluator struct {
    Layers []NeuralLayer
}
```

Persisted as JSON (`data/training/model.json`). Supports both `"mlp"` (N hidden layers) and legacy `"mlp-1h"` (single hidden layer) formats for backward compatibility.

## Forward Pass

Pure Go inference — no CGo, no Python, no ONNX:

```go
func (ne *NeuralEvaluator) Evaluate(v StateVector) float64
```

For each hidden layer: `output = ReLU(W·input + b)`. Final layer: `output = sigmoid(W·input + b)`.

## Training Data Collection

```go
type TrainingCollector struct {
    snapshotInterval int  // default: every 5 turns
}
```

During each game, `Snapshot(gs, perspectiveSeat)` records the state vector at regular intervals. After the game ends, `Finalize(placement, gameTurns)` stamps all snapshots with the outcome:

```go
type TrainingSample struct {
    State     StateVector
    Placement float64  // 1.0 = 1st, 0.0 = last
    Turn      int
    GameTurn  int
}
```

Samples are enriched with Tesla causal pivot data before writing to `data/training/samples.jsonl`.

## Integration with Hand-Coded Evaluator

When a neural model exists, `evalPosition()` blends both evaluators:

```
score = 0.8 × heuristic_score + 0.2 × neural_score
```

The 80/20 split ensures the well-understood heuristic dominates while the neural model contributes learned nuance. As the model improves through more training data, this ratio can shift.

When no model file exists, `TryLoadNeuralEvaluator()` returns nil and the hat gracefully degrades to pure heuristic evaluation.

## Training Pipeline

Training happens externally via PyTorch:

```bash
python3 scripts/train_neural_eval.py \
    --samples data/training/samples.jsonl \
    --output data/training/model.json \
    --epochs 200
```

The script supports CUDA (4090) and CPU fallback, uses ReduceLROnPlateau scheduling and early stopping with checkpoints. The trained model is exported to the JSON format that `LoadNeuralEvaluator()` reads.

See [Self-Play Loop](Self-Play%20Loop.md) for the automated training coordination.

## Related Docs

- [Self-Play Loop](Self-Play%20Loop.md) — automated training cycle
- [Tesla Causal Pivots](Tesla%20Causal%20Pivots.md) — enriches training samples
- [Eval Weights and Archetypes](Eval%20Weights%20and%20Archetypes.md) — hand-coded evaluator
- [YggdrasilHat](YggdrasilHat.md) — consumer of both evaluators
