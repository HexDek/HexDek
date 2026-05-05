# Learning Loop

> Sources: `internal/heimdall/observer.go` (117 lines), `internal/muninn/muninn.go`, `internal/hat/selfplay.go` (164 lines), `internal/hat/tesla.go` (164 lines), `internal/hat/feynman.go` (307 lines), `internal/hat/ive.go` (236 lines)
> Status: **Production** — fully wired in all 3 game paths

The learning loop is HexDek's post-game intelligence pipeline. Every completed game generates observations that flow through Heimdall to downstream sinks — each sink extracts a different kind of knowledge from the same game data.

## Architecture

```
                        ┌─────────────────────────────────┐
                        │         Game Completes          │
                        │  (grinder / bracket / showmatch)│
                        └────────────────┬────────────────┘
                                         │
                                         ▼
                        ┌─────────────────────────────────┐
                        │   Heimdall Observer (singleton)  │
                        │   internal/heimdall/observer.go  │
                        └─────┬───┬───┬───┬───┬───┬───────┘
                              │   │   │   │   │   │
              ┌───────────────┘   │   │   │   │   └────────────────┐
              ▼                   ▼   │   ▼   │                    ▼
        ┌──────────┐     ┌──────────┐│┌──────────┐        ┌──────────────┐
        │  Seeds   │     │  Huginn  │││  Muninn  │        │  GA4         │
        │ (disk)   │     │ (synergy)│││ (memory) │        │ (telemetry)  │
        └──────────┘     └─────┬────┘│└──────────┘        └──────────────┘
                               │     │
                               ▼     │
                        ┌──────────┐ │
                        │  Freya   │ │
                        │(strategy)│ │
                        └──────────┘ │
                                     │
              ┌──────────────────────┼──────────────────────┐
              ▼                      ▼                      ▼
        ┌──────────┐         ┌──────────┐           ┌──────────┐
        │  Tesla   │         │ Feynman  │           │Training  │
        │ (pivots) │         │ (oracle) │           │Collector │
        └────┬─────┘         └──────────┘           └────┬─────┘
             │                                           │
             └──────────────┬────────────────────────────┘
                            ▼
                    ┌──────────────┐
                    │ Self-Play    │
                    │ Manager      │
                    └──────┬───────┘
                           │ threshold met
                           ▼
                    ┌──────────────┐
                    │ PyTorch      │
                    │ Training     │
                    └──────┬───────┘
                           │ success
                           ▼
                    ┌──────────────┐
                    │ Hot Reload   │
                    │ Neural Model │
                    └──────────────┘
```

## Heimdall Observer

The singleton that receives game results from ALL game paths. Source: `internal/heimdall/observer.go`.

```go
type Observer struct {
    seedBuf   []GameSeed      // ring buffer, cap 1000
    huginn    HuginnSink
    muninn    MuninnSink
    telemetry TelemetrySink
    dataDir   string
    mu        sync.Mutex
}
```

### Seed Recording

`RecordSeed()` — called after EVERY game. Appends to a bounded ring buffer (1000 entries). When full, flushes to `heimdall/seeds.jsonl` on disk. Thread-safe via mutex.

```go
type GameSeed struct {
    RNGSeed    int64
    DeckKeys   [4]string
    Winner     int
    Turns      int
    KillMethod string
}
```

Seeds enable exact replay — given a seed + deck keys, the engine reproduces the exact game state for debugging.

### Observation Routing

`RecordObservation()` — routes data to downstream sinks without buffering:

```go
func (o *Observer) RecordObservation(obs Observation) {
    // Co-triggers → Huginn (synergy discovery)
    if o.huginn != nil && len(obs.CoTriggers) > 0 {
        o.huginn.IngestCoTriggers(obs.CoTriggers, deckNames)
    }
    // Parser gaps + dead triggers → Muninn (persistent memory)
    if o.muninn != nil {
        o.muninn.RecordParserGaps(obs.ParserGaps, gameID)
        o.muninn.RecordDeadTriggers(obs.DeadTriggers, gameID)
    }
}
```

Observations are dispatched immediately — no buffering. Sinks can be nil (observations silently dropped).

### Crash Recovery

`RecordCrash()` — called from panic recovery in the grinder. Routes stack traces + deck keys to Muninn for persistent crash memory.

### Health Pulse

`Pulse()` — forwards server health stats to GA4 telemetry (games played, uptime, memory usage).

## Sink Interfaces

```go
type HuginnSink interface {
    IngestCoTriggers(pairs []CoTriggerPair, deckNames []string)
}

type MuninnSink interface {
    RecordParserGaps(gaps []string, gameID string)
    RecordDeadTriggers(triggers []DeadTrigger, gameID string)
    RecordCrash(panicMsg string, stackTrace string, deckKeys []string)
}

type TelemetrySink interface {
    Pulse(stats HealthPulse)
}
```

## Data Flow Per Game

After each game completes, the tournament runner executes this sequence:

1. **Feynman Oracle** validates the game state (8 invariant checks)
2. **Tesla** extracts the causal pivot from eval history
3. **Training Collector** finalizes samples with placement scores
4. **Tesla** enriches samples with pivot distance
5. **Self-Play Manager** writes enriched samples, checks training threshold
6. **Curse Pool** records result, triggers evolution if threshold met
7. **Heimdall Observer** records seed + routes observation to Huginn/Muninn
8. **Ive** composes narrative (showmatch path only)

Steps 1–6 happen inline in the game goroutine. Step 7 is the Heimdall dispatch. Step 8 is showmatch-only.

## Huginn: Synergy Discovery

Huginn ingests co-trigger pairs — cards that triggered together during a game. Over many games, frequently co-occurring cards graduate through tiers:

| Tier | Threshold | Action |
|------|----------|--------|
| 0 | Observed | Stored |
| 1 | 10+ games | Pattern confirmed |
| 2 | 50+ games | Cross-deck validated |
| 3 | 100+ games | Exported to Freya |

Tier 3 patterns feed back into Freya's strategy analysis, creating a closed loop: games → observations → pattern discovery → strategy refinement → better hat decisions → more games.

See [Tool - Huginn](Tool%20-%20Huginn.md) for details.

## Muninn: Persistent Memory

Muninn persists three types of data:

- **Parser gaps**: card names the engine encountered but couldn't fully parse
- **Dead triggers**: per-card triggers that registered but never fired
- **Crash reports**: panic messages + stack traces + deck keys

All stored as append-only JSON files. Muninn's data drives handler development priority — cards that appear frequently in parser gaps get handlers first.

See [Tool - Muninn](Tool%20-%20Muninn.md) for details.

## Closed-Loop Feedback

The learning loop creates three feedback cycles:

### Cycle 1: Curse Evolution
```
Game → placement score → CursePool.RecordResult()
    → evolve() every 100 games
    → DimStats.WeightCorrections() → next game's eval weights
```
Timescale: ~100 games per evolution step (~1 minute at grinder speed).

### Cycle 2: Neural Training
```
Game → TrainingCollector snapshots → Tesla enrichment
    → SelfPlayManager.WriteEnrichedSamples()
    → PyTorch training (every 10,000 samples)
    → hot-reload NeuralEvaluator into running hats
```
Timescale: ~10,000 games per training cycle (~10 minutes at grinder speed).

### Cycle 3: Huginn → Freya
```
Game → Heimdall → Huginn co-trigger ingestion
    → Tier graduation (10 → 50 → 100 games)
    → Tier 3 export → Freya reads on next analysis
    → Updated strategy profiles → hat construction
```
Timescale: hundreds of games per new Tier 3 pattern (~hours).

## Related Docs

- [Tool - Heimdall](Tool%20-%20Heimdall.md) — the CLI analytics tool (separate from the observer)
- [Tool - Huginn](Tool%20-%20Huginn.md) — synergy discovery details
- [Tool - Muninn](Tool%20-%20Muninn.md) — persistent memory details
- [Tesla Causal Pivots](Tesla%20Causal%20Pivots.md) — pivot extraction
- [Feynman Oracle](Feynman%20Oracle.md) — game validation
- [Neural Evaluator](Neural%20Evaluator.md) — training data consumer
- [Self-Play Loop](Self-Play%20Loop.md) — training coordination
- [Ive Spectator](Ive%20Spectator.md) — narrative generation
- [Genetic Curse](Genetic%20Curse.md) — per-deck evolution
