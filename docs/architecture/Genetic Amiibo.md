# Genetic Amiibo

> Source: `internal/hat/amiibo.go` (377 lines)
> Status: **Production** — evolving per-deck DNA across all game paths

Genetic Amiibo is HexDek's per-deck personality evolution system. Each deck maintains a population of 8 DNA variants that compete, mutate, and evolve based on game outcomes. Over hundreds of games, each deck's Amiibo converges on the hat parameters that work best for *that specific pile*.

Named after Nintendo's Amiibo figures — physical objects that learn and adapt to their owner's playstyle.

## DNA Schema

```go
type AmiiboDNA struct {
    DeckKey         string
    Generation      int
    GamesPlayed     int
    Fitness         float64    // EMA of placement scores

    // Personality parameters [0, 1]
    Aggression       float64   // attack threshold, threat targeting bias
    ComboPat         float64   // combo patience (turns to wait before pivoting)
    ThreatParanoia   float64   // how aggressively to read opponents as threats
    ResourceGreed    float64   // weight on card/mana advantage vs tempo
    PoliticalMemory  float64   // grudge decay rate (0 = goldfish, 1 = elephant)
    DrainAffinity    float64   // aristocrats/drain engine weight nudge
    ArtifactAffinity float64   // artifact synergy weight nudge
}
```

Each parameter maps directly to a hat behavior. `Aggression` modulates attack thresholds in `chooseAttackers()`. `ComboPat` feeds the plan state machine's `assembleTimeout()` (3–8 turns before pivoting away from a stalled combo). `PoliticalMemory` controls the exponential decay rate on grudge scores.

## Population and Selection

Each deck gets a pool of 8 DNA variants:

```go
const AmiiboPopSize = 8
```

Before each game, `SelectForGame()` picks one variant via fitness-proportional roulette selection — fitter variants get picked more often, but all variants get chances. This prevents premature convergence while still biasing toward winners.

## Evolution Loop

Evolution triggers every 100 games (`AmiiboEvolveAt = 100`):

1. Sort population by fitness (descending)
2. Kill the bottom 2 (`AmiiboKillCount = 2`)
3. Clone the top 2 into the empty slots
4. Apply Gaussian mutation (σ = 0.05) to the clones
5. Every 10 generations, inject one fully random immigrant to prevent convergent collapse

```
Generation 0:  [random] [random] [random] [random] [random] [random] [random] [random]
                                    ↓ 100 games
Generation 1:  [best₁]  [best₂]  [mid₁]  [mid₂]  [mid₃]  [mid₄]  [clone₁'] [clone₂']
                                                                      ↑ mutated
```

## Fitness Scoring

```go
func PlacementScore(placement, totalPlayers int) float64
```

Converts 1st–4th finish to continuous `[0, 1]`:
- 1st place → 1.0
- 2nd place → 0.67
- 3rd place → 0.33
- 4th place → 0.0

This is smoother than binary win/loss — a DNA variant that consistently places 2nd accumulates more fitness than one that alternates between 1st and 4th.

Fitness uses an exponential moving average, so recent results matter more than ancient history.

## Dimension Stats (T3.2 Outcome-Correlated Learning)

Beyond the 7 personality params, each pool tracks `DimensionStats` — EMA-based running correlations between all 20 evaluator dimensions and game outcomes:

```go
type DimensionStats struct {
    MeanDim    [NumDimensions]float64
    MeanSqDim  [NumDimensions]float64
    CrossDim   [NumDimensions]float64
    MeanScore  float64
    MeanSqScore float64
    N          int
}
```

After each game, `RecordGame(dimMeans, outcome)` updates the EMA (α = 0.04, ~25-game half-life). `WeightCorrections()` returns per-dimension multiplicative adjustments (±20% swing) based on Pearson correlation. If BoardPresence correlates negatively with winning for a voltron deck, the Amiibo learns to weight it down.

These corrections are applied at hat construction, on top of the archetype defaults and Freya's static weights.

## Persistence

Each deck's pool saves to `data/amiibo/<deckkey>.json`. Auto-saved after every evolution step. Loaded at server startup.

```go
func SavePool(dir string, pool *AmiiboPool) error
func LoadPool(dir, deckKey string) (*AmiiboPool, error)
func SaveAllPools(dir string, pools map[string]*AmiiboPool) error
func LoadAllPools(dir string) (map[string]*AmiiboPool, error)
```

## Integration

- Tournament runner calls `SelectForGame()` pre-game, `RecordResult()` post-game
- `AmiiboDNA` fields read by YggdrasilHat via `h.DNA` to modulate real-time decisions
- `DimStats.WeightCorrections()` feeds the evaluator at game start
- `PlacementScore()` uses bracket-normalized fitness when bracket is set

## Related Docs

- [YggdrasilHat](YggdrasilHat.md) — consumes DNA parameters
- [Eval Weights and Archetypes](Eval%20Weights%20and%20Archetypes.md) — base weights that Amiibo modulates
- [Self-Play Loop](Self-Play%20Loop.md) — training loop that runs alongside evolution
