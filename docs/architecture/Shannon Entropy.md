# Shannon Entropy and Information Tracking

> Source: `internal/hat/yggdrasil.go` (3rd Eye subsystem), `internal/hat/information_set.go` (107 lines)
> Status: **Production** — active in all decision paths

Shannon entropy tracking gives YggdrasilHat a probabilistic model of what opponents are holding. Instead of treating opponent hands as black boxes, the hat builds per-opponent information profiles that gate critical decisions — most importantly, whether to fire a combo this turn.

## Two Complementary Systems

| System | Approach | Used for |
|--------|---------|----------|
| **3rd Eye** (heuristic) | Track observable events to estimate hand entropy | Real-time decision gating |
| **Information-Set MCTS** (search) | Determinize hidden state, run rollouts | Deep candidate evaluation |

## 3rd Eye: Opponent Hand Modeling

### Tracked State (per opponent)

```go
opponentHandEntropy  []float64          // [0,1] entropy estimate (0 = known, 1 = mystery)
opponentHeldMana     []int              // consecutive turns passing with 2+ open mana
opponentTutored      []bool             // has this opponent tutored this game?
opponentKnownCards   []map[string]bool  // cards confirmed in hand via reveal
opponentColors       []map[string]bool  // colors opponent has produced
cardsSeen            []map[string]int   // all cards observed per opponent
```

### Entropy Update Rules

Events that **decrease** entropy (hand becomes more predictable):
- **Tutor/search**: `entropy *= 0.6` (floor 0.1) — they found something specific
- **Reveal**: `entropy -= 0.15` — we saw a card directly
- **Cast**: remove from `knownCards`, reset `heldMana` counter

Events that **increase** entropy:
- **Draw**: `entropy += 0.08 × cards_drawn` (cap 1.0) — unknown cards entered

Passive tracking:
- **Upkeep**: if opponent has 2+ untapped mana and didn't cast → increment `heldMana`; else reset

### Query Methods

```go
opponentLikelyHasAnswer(oppSeat) bool
```
Returns true when: tutored AND held mana ≥ 2 turns AND has U/B colors. This is the primary combo-delay gate.

```go
opponentHasInteraction(gs, oppSeat) float64
```
Probability estimate (0–1) combining: open mana, colors, hand size, tutor flag, held mana turns, known interaction cards.

```go
tableInteractionRisk(gs, seatIdx) float64
```
Max interaction probability across all opponents. When high, the hat reorders spell casts to bait counters before committing key pieces.

## Combo Delay (COMBO-DELAY)

The most impactful integration. In the Execute plan:

```
for each interactive opponent:
    if opponentLikelyHasAnswer(opponent):
        entropyBlocked = true
        → delay combo execution this turn
```

The hat won't fire a combo into open U/B mana after a tutor. It waits for the opponent to tap out or for a protection spell.

## Spell Sequencing

When `tableInteractionRisk()` is high, the hat reorders casts:
1. Cast low-value spells first (bait counters)
2. Cast high-value spells after interaction is spent
3. Key combo pieces cast last when mana is committed elsewhere

## Attack Targeting

Entropy data flows into `seatThreat`:

```go
type seatThreat struct {
    HandEntropy    float64
    HeldManaTurns  int
    // ... other threat fields
}
```

Seats with confirmed removal (`heldMana` high, interaction likely) get attacked less aggressively — no point sending creatures into known removal.

## Information-Set MCTS

Source: `internal/hat/information_set.go`

For deeper decisions, the hat uses proper imperfect-information search:

```go
func determinize(gs *GameState, perspectiveSeat int, rng *rand.Rand) *GameState
```

For each opponent: shuffle their hand back into library, redeal the same count. This produces a "possible world" consistent with what the perspective player knows.

```go
func multiRolloutForCard(gs, seatIdx, card, n int) float64
```

Runs N determinized rollouts for a candidate action, returns the mean eval score. Each rollout sees a different possible arrangement of opponent hands. This is IS-MCTS — the standard approach for imperfect-information games (used by champion poker AIs).

3 rollouts per candidate is the current default, balancing accuracy against throughput.

## Related Docs

- [YggdrasilHat](YggdrasilHat.md) — parent system
- [Hat State Machine](Hat%20State%20Machine.md) — Execute plan gated by entropy
- [MCTS and Yggdrasil](MCTS%20and%20Yggdrasil.md) — rollout mechanics
