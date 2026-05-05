# MicroNet: Per-Deck Neural Net (Level 7)

## Purpose

MicroNet gives each deck its own tiny neural network that learns deck-specific
position evaluation from game outcomes. Where the global neural evaluator
(Level 5) learns universal position patterns across all decks, MicroNet captures
what "winning position" looks like for *this specific deck* — whether that's
having combo pieces assembled, maintaining a mana advantage, or building a
critical mass of board presence unique to its strategy.

## Architecture

```
Input (30-dim)                    Output
┌──────────────────────────────┐
│ [0-19]  20 eval dimension    │
│         scores (tanh-normed) │
│ [20]    game stage (0/0.5/1) │
│ [21]    current plan / 5.0   │    30 → 64 → 48 → 16 → 1
│ [22]    combo proximity      │      ReLU   ReLU   ReLU  sigmoid
│ [23]    life / 40            │
│ [24]    position vs field    │    ~5,905 parameters per deck
│ [25]    mana pool / 15       │
│ [26]    hand size / 10       │    Output: [0,1] predicted placement
│ [27]    turn normalized      │
│ [28]    alive players / 4    │
│ [29]    library size / 99    │
└──────────────────────────────┘
```

**Hidden layers:** 64 → 48 → 16 neurons, ReLU activations.
**Output:** Single sigmoid neuron predicting normalized placement score.

## Training

- **Algorithm:** Full-batch SGD with momentum (0.9), learning rate 0.001
- **Epochs:** 80 per training cycle
- **Trigger:** Every 200 samples accumulated for a deck
- **Buffer cap:** 5,000 samples per deck (FIFO eviction)
- **Initialization:** He/Kaiming normal (√(2/fan_in))
- **Loss:** Mean squared error on placement prediction
- **Backpropagation:** Implemented in pure Go — no PyTorch dependency

### Sample Collection

During each game, every 3rd turn a snapshot is recorded per seat:
1. The hat's evaluator produces a detailed 20-dimension breakdown
2. `EncodeMicroInput()` combines this with game context (stage, plan, life, etc.)
3. Post-game, the collector stamps all snapshots with the seat's final placement score

### Training Pipeline

```
Game Turn Loop
    └── every 3rd turn: EncodeMicroInput() → MicroCollector.Record()
                                                    │
Game End                                            ▼
    └── MicroCollector.Finalize(placement) → []MicroSample
                                                    │
                                                    ▼
                                        MicroTrainer.AddSamples()
                                            │
                                            ├── buffer < 200: accumulate
                                            └── buffer ≥ 200: trainDeck()
                                                    │
                                                    ├── Train(samples, 80 epochs, 0.001 LR)
                                                    └── SaveMicroNet() → data/micronets/{deck}.json
```

## Integration with Evaluation

MicroNet output is blended into `evalPosition()`:

| Configuration | Blend |
|--------------|-------|
| Heuristic only (no micro, no neural) | 100% heuristic |
| Heuristic + neural (Level 5) | 80% heuristic, 20% neural |
| Heuristic + micro (Level 7) | 70% heuristic, 30% micro |
| Heuristic + micro + neural | 60% heuristic, 20% micro, 20% neural |

The micro-net gets a higher weight than the global neural evaluator in the
full-stack blend because it has deck-specific knowledge that the global model
lacks.

## Persistence

Each deck's trained net is saved as JSON in `data/micronets/{deck_key}.json`:

```json
{
  "layers": [
    {"weights": [[...]], "biases": [...]},
    {"weights": [[...]], "biases": [...]},
    {"weights": [[...]], "biases": [...]},
    {"weights": [[...]], "biases": [...]}
  ],
  "deck_key": "tinybones-trinket-1a2b3c",
  "generation": 15,
  "train_loss": 0.042
}
```

On startup, `MicroTrainer.LoadAll()` reads all persisted nets. New nets are
created automatically when a deck accumulates its first 200 training samples.

## Key Types

| Type | Role |
|------|------|
| `MicroNet` | Single deck's neural net (layers, weights, forward/train) |
| `MicroSample` | One training datum: 30-dim input + placement label |
| `MicroCollector` | Per-seat game recorder, stamps placement on finalize |
| `MicroTrainer` | Manages all deck buffers, triggers training, persists nets |

## Scale

With 654 decks in the pool and ~5,905 params each, the total parameter budget
is ~3.9M — trivial for any modern machine. Training 80 epochs on 200 samples
takes ~1ms per deck in Go. The system scales linearly with deck count.

## Relationship to Other Systems

- **Genetic Amiibo (Level 4):** GA evolves 7 personality floats; MicroNet learns
  nonlinear patterns the GA can't represent. They're complementary — GA nudges
  behavior, MicroNet refines evaluation.
- **Global Neural Evaluator (Level 5):** Learns universal patterns across all
  decks. MicroNet learns deck-specific patterns. Both feed into evalPosition.
- **DimensionStats (T3.2):** Linear weight corrections from outcome correlation.
  MicroNet is the nonlinear successor — it subsumes what DimStats does but with
  a much richer representational capacity.

## Source Files

- `internal/hat/micronet.go` — Implementation
- `internal/hat/micronet_test.go` — 10 tests
- `internal/hat/yggdrasil.go` — evalPosition() blend logic
- `internal/hexapi/showmatch.go` — Game loop wiring (grinder, bracket, showmatch)
