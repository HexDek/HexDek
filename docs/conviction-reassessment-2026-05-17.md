# Conviction System Reassessment

**Date:** 2026-05-17
**Branch:** dev/conviction-reassess
**Context:** Round 12 (commit b0b6db4) landed archetype-specific eval weights for 22 Freya archetypes plus sparse-weight merge. The question: does that change make the conviction-based `ShouldConcede` safe to re-enable?

## TL;DR

**No. Do not re-enable the old score-based conviction with the new weights.**

The new weights make `relativePosition` a *more accurate* board-state metric, which makes a score-threshold concession *more* dangerous, not less. A combo deck behind on board now correctly scores low on `BoardPresence` but its win path (resolving a 2-card combo) is unchanged by board state.

A safe-to-enable subset exists, but it must be **outs-based / survivability-based**, not score-based. Score is a meta-metric of board strength, not a probability-of-winning estimate, and EDH has too much variance (politics, topdecks, commander damage, latent combos) for "I feel behind for 4 turns" to mean "I cannot win."

## What was removed (commit `a9c157a`, 2026-05-04)

Old `ShouldConcede` in `internal/hat/yggdrasil.go`:

- 4-turn sliding window of `relativePosition(gs, seatIdx) = myScore - bestOppScore`
- Conceded after turn 10 if every window sample was `< -0.35`
- Killed because hat scooped at 38 life with full deck — purely positional sadness, no consideration of actual lethal state or path to a win.

Engine path is intact: `internal/tournament/turn.go:73` still calls `seat.Hat.ShouldConcede` at top of turn, `ConcedeGame` writes `LossReason="concession"` and fires a CR §104.3a event. Hooks for `Concessions` count and the per-game record are still threaded in `runner.go`. Only the YggdrasilHat logic is `return false`.

## What changed since (commit `b0b6db4`, 2026-05-17 — round 12)

1. Archetype-specific 20-dim eval weights for 22 archetypes (was 5 generic profiles), including Voltron CommanderProgress=2.0, Lifegain LifeResource=1.8, Stax StaxLockProgress=2.0, Combo ComboProximity=2.0, Mill DrainEngine=1.4, etc.
2. Sparse-weight merge: Freya's 8 serialized dimensions now overlay on the full archetype default instead of zeroing the other 12.
3. Sibling: opponent archetype recognition (`b46d077`) and strategy bridge already feed into hat decisions.

Net effect on `relativePosition`: it is now a more faithful measure of "how good is my board *for what my deck is trying to do*." Stax decks behind on board but with lock pieces in hand will not score as poorly. Combo decks will score lower on `BoardPresence` but higher on `ComboProximity` when pieces are visible — but this assumes pieces are *visible*; pieces still in hand or library do not move ComboProximity.

## Why score-based conviction still fails with the new weights

The fundamental gap: `relativePosition` summarizes the **observable board**, not the **distribution of futures**.

EDH-specific variance the score cannot see:

- **Topdeck wins.** A combo deck holding 0 pieces with 70 cards in library and an active tutor has a measurable shot. Score reads "behind."
- **Commander damage path.** Voltron deck behind on board but with a 2/2 commander, an equipment in hand, and 18 fresh attacks of headroom is *not* losing. CommanderProgress=2.0 only captures progress on existing pieces.
- **Politics.** A 3-player table can refocus on the leader at any moment. Conviction window doesn't model this.
- **Latent lethal.** Tendrils-style storm decks score terribly until they don't.
- **Stabilization.** Control with a sweeper in hand looks bad until end-of-opponent's-turn.

Making the score more accurate per-archetype *amplifies* these blind spots: a Voltron deck now correctly scores `CommanderProgress` low when the commander is in the command zone with 4 tax — but that's a deliberate setup turn for a 21-damage one-shot the next.

The old threshold (-0.35 sustained for 4 turns) didn't get less brittle. It got more brittle, because the weights now correlate more tightly with current-board, ignoring the deck's plan-state.

## Safe-to-enable subset (narrow)

If conviction must come back, the only conditions that survive scrutiny are **outs-based**, not score-based. Concede only when winning is *mechanically impossible*, not when the score is low:

### Condition 1: Library exhaustion against a non-mill deck
- Library size + cards-in-hand + active-recursion-options < cards-needed-to-survive-to-next-draw
- This is essentially already handled by SBA "deck out" (CR 704.5b). Probably no value to duplicate.

### Condition 2: Win-line extinction (archetype-aware, requires Strategy)
For decks with discrete win conditions, concede when **all** are mathematically gone:
- **Combo deck**: every card in `h.Strategy.ComboPieces` is in exile or removed-from-game across all instances *and* no tutor effects remain that could fetch a replacement. Requires tracking exile zones across all seats for combo pieces.
- **Voltron**: commander permanently removed-from-game (e.g., Solphim+Hostile Negotiations style RFG) AND commander tax > available mana ceiling for remainder of game.
- **Mill**: opponents collectively can't be milled out — their combined library + recursion exceeds your remaining mill output. Hard to compute reliably.

These three are real concession signals because they identify when the deck's plan is *gone*, not just behind. But:

- Detection is fragile (need to scan exile across seats, track tutor counts).
- False positives still costly: a "lost" combo piece sitting in graveyard with one Yawgmoth's Will in hand is still a win.
- Frequency is low — these states are rare enough that the value of detecting them is probably below the cost of the bug surface.

### Condition 3: Hard lockout for ≥3 turn cycles
- Hat has produced zero legal actions other than land drop / pass for 3 full turn rotations
- AND no card in hand can break the lock
- AND draw-step expected-value to break the lock is negligible

This catches stalled-against-Stax states. But this is also the textbook "stalemate" condition the engine should detect at a higher level (and partially does via `loop_shortcut`). Easier and safer to extend stalemate detection in the engine than to embed it in hat.

### What does **not** belong in conviction
- Score thresholds, regardless of archetype tuning.
- Life total alone (a 4-life hat with a Worldgorger combo in hand wins next turn).
- Eliminated-opponent counts (3-vs-1 with the 1 holding the win is still winnable).
- Turn count (long games aren't lost games).

## Recommendation

**Keep `ShouldConcede` returning `false` for YggdrasilHat.** The cost of over-concession (lost-game accounting noise, unstable winrate measurements, philosophical wrongness in EDH context) outweighs the benefit of letting hopeless games end early. Tournament throughput is already bounded by per-turn time budgets and `maxTurns`; a sad-but-alive hat does not block the game forever.

If we want to revisit, the right next steps are:

1. **Instrument first**, ship later. Add a *non-acting* `convictionDiagnostic` that records relative position, win-line presence, and lockout state every turn for completed games, then post-hoc analyze: across N games, when would each candidate trigger have fired, and what fraction of those games did the diagnosed-hopeless hat *actually win*?
2. **Compare against the engine**, not against intuition. Hoist Condition 3 (hard-lockout detection) into engine-level stalemate detection, not hat-level conviction.
3. **Outs-based only.** If any conviction logic ships, it must be Condition 2 (win-line extinction) and only after a tournament A/B against `return false` shows neither winrate distortion nor early-concession false positives.

## Files touched in this audit

- `internal/hat/yggdrasil.go` — read only (no change; conviction stub at L6361 stays).
- `internal/hat/eval_weights.go` — read only.
- `internal/hat/poker.go` — read only (archetype constants).
- `internal/gameengine/hat.go` — read only (interface + ConcedeGame).
- `internal/tournament/turn.go` — read only (caller).

No code change in this branch; this is an assessment doc only.
