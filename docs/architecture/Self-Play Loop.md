# Self-Play Loop

> Source: `internal/hat/selfplay.go` (164 lines)
> Status: **Production** — auto-triggers training when sample threshold is met

The self-play loop coordinates the full cycle: games generate training samples → samples accumulate → PyTorch training fires → new model hot-reloads into running hats. No manual intervention required.

## Architecture

```
Grinder/Bracket/Showmatch game
    │
    ▼
TrainingCollector snapshots (every 5 turns per seat)
    │
    ▼
Tesla enriches samples with causal pivot distance
    │
    ▼
SelfPlayManager.WriteEnrichedSamples()
    │
    ▼
RecordSamples(n) ──→ threshold check (10,000 samples)
    │                        │
    │                   [not met] → return
    │                        │
    │                   [met + cooldown elapsed]
    │                        │
    │                        ▼
    │               runTraining() (async goroutine)
    │                   python3 train_neural_eval.py
    │                        │
    │                   [success]
    │                        │
    │                        ▼
    │               TryLoadNeuralEvaluator()
    │                        │
    │                        ▼
    │               OnModelLoad(newModel)
    │                   ↓
    └──────── hat continues with new neural evaluator
```

## Configuration

```go
type SelfPlayConfig struct {
    SamplesPath    string  // data/training/samples.jsonl
    ModelPath      string  // data/training/model.json
    CheckpointDir  string  // data/training/checkpoints/
    TrainScript    string  // scripts/train_neural_eval.py
    TrainThreshold int     // 10,000 samples before training
    Epochs         int     // 200
    PythonBin      string  // python3
}
```

## Coordination

`SelfPlayManager` uses atomic operations for goroutine safety:

- `sampleCount int64` — total samples accumulated since last training
- `training int32` — flag preventing concurrent training runs
- `lastAttempt int64` — Unix timestamp enforcing 5-minute cooldown between attempts
- `generation int32` — increments on each successful training cycle

When `RecordSamples()` pushes the count over `TrainThreshold` and the cooldown has elapsed, training fires asynchronously. The calling goroutine (grinder worker) returns immediately — training runs in the background.

## Training Execution

`runTraining()` shells out to PyTorch:

```bash
python3 scripts/train_neural_eval.py \
    --samples data/training/samples.jsonl \
    --output data/training/model.json \
    --epochs 200
```

On success:
1. Load the new model via `TryLoadNeuralEvaluator()`
2. Fire `OnModelLoad` callback to hot-swap into all active hats
3. Advance `generation` counter
4. Reset `sampleCount`

On failure: log the error, advance `lastTrained` to prevent immediate re-trigger, try again after more samples accumulate.

## Hot Reload

The `OnModelLoad` callback is set by the tournament runner at startup. It swaps the new `NeuralEvaluator` pointer into the shared evaluator used by all hat instances. Games already in progress finish with their current evaluator; new games pick up the updated model.

## Status Monitoring

```go
func (sp *SelfPlayManager) Status() string
func (sp *SelfPlayManager) IsTraining() bool
func (sp *SelfPlayManager) Generation() int
func (sp *SelfPlayManager) SampleCount() int64
```

These are exposed via the server's health endpoint for monitoring.

## Related Docs

- [Neural Evaluator](Neural%20Evaluator.md) — the model being trained
- [Tesla Causal Pivots](Tesla%20Causal%20Pivots.md) — enriches samples with pivot distance
- [Feynman Oracle](Feynman%20Oracle.md) — validates games before samples are accepted
- [Tool - Heimdall](Tool%20-%20Heimdall.md) — observation routing
