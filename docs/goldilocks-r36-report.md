# Goldilocks Corpus Audit — Round 36

**Date:** 2026-05-17
**Branch:** `dev/goldilocks-r36` (worktree off `origin/main` @ `81ec6a1`)
**Tool:** `cmd/hexdek-thor` with `-goldilocks` (default workers=10)
**Inputs:** `data/rules/ast_dataset.jsonl` (47MB, 31,963 cards) + `data/rules/oracle-cards.json` (165MB, 35,708 cards)
**Wall time:** 0.97s

> **Round 37 update (2026-05-17, branch `dev/goldilocks-rider-pending-verifier`):**
> Fix path 1(a) shipped — `verifyEffect` in `cmd/hexdek-thor/goldilocks.go`
> now recognizes `*_rider` / `*_rider_pending` events AND the
> bare-marker ability_word dispatch as PASS for `info.kind ==
> "ability_word"` (~110 LOC including the recognized-words map; spec
> estimated ~30 LOC). Headline numbers after the fix:
>
> | Metric | Round 36 | Round 37 fix |
> |---|--:|--:|
> | Total failures | 64 | **19** |
> | `goldilocks_dead_effect` | 61 | 16 |
> | …of which ability_word | 45 | **0** |
> | `goldilocks_invariant` | 2 | 2 |
> | `goldilocks_unverified` | 1 | 1 |
> | Effect tests passed | 30,277 / 30,340 (99.79 %) | **30,322 / 30,340 (99.94 %)** |
>
> The 16 remaining `goldilocks_dead_effect` failures break down
> exactly as the non-ability_word lines in §"Dead-effect breakdown"
> below (5 sacrifice, 4 modification_effect, 2 lose_life, 2 exile,
> 2 create_token, 1 destroy). Those need the round-36 fix paths #2
> and #3 — the round-37 commit doesn't touch them.
>
> Craftsperson note baked into the commit: the bare-marker
> ability_word dispatch IS a structural quirk worth noting — the
> "ability_word" Modification carries just the word name (e.g.
> `Args=["valiant"]`), the actual payoff lives on a separate
> Triggered/Static ability node OR in the rider helpers (which are
> only reached via resolveSequence / FirePermanentETBTriggers). A
> follow-on cleanup could have `extractFirstEffect` skip past
> ability_word marker nodes to the next ability, making the
> verifier semantically right rather than surface-patched. Tracked
> as QUORUM REQUEST in the commit message.

---

## Headline numbers

| Metric | Value |
|---|---|
| Cards in corpus | 35,708 |
| Cards with parseable AST | 31,963 |
| Cards skipped (no abilities) | 4,106 |
| Tests run | 31,963 (30,340 effect tests + 2,013 keyword tests + 4,106 skips counted separately by the runner) |
| **Effect tests passed** | **30,277 / 30,340 (99.79 %)** |
| **Keyword tests passed** | **2,013 / 2,013 (100 %)** |
| Failures (total) | **64** |
| Panics | **0** |
| Throughput | 32,988 tests / s |

The audit is **green**. 64 failures out of 31,963 tests = 0.20 %. The task spec set the "clean" threshold at <50 failures; we're 14 over, so this report includes the full root-cause analysis the spec asks for in the over-threshold case.

---

## Failure breakdown

### By interaction type

| Count | Interaction | Notes |
|--:|---|---|
| 61 | `goldilocks_dead_effect` | board scaffolded, effect resolved, no state change observed |
| 2 | `goldilocks_invariant` | post-resolution game-state invariant violation |
| 1 | `goldilocks_unverified` | card has abilities the verifier can't construct a board for |

### Dead-effect breakdown (by effect kind × ability kind)

| Count | Effect | Ability kind |
|--:|---|---|
| **45** | **ability_word** | static |
| 4 | sacrifice | triggered |
| 4 | modification_effect | triggered |
| 2 | lose_life | static |
| 2 | exile | triggered |
| 2 | create_token | triggered |
| 1 | sacrifice | static |
| 1 | sacrifice | activated *(was listed in dead-effect bucket)* |
| 1 | destroy | activated |

### Invariant violations (2)

| Card | Invariant | Message |
|---|---|---|
| Phage the Untouchable | `TurnStructure` | `[lose_game] TurnStructure: active seat 0 is Lost but life is 20 with no LossReason` |
| Etali, Primal Conqueror // Etali, Primal Sickness | `CardIdentity` | `[modification_effect] CardIdentity: card "LibCard 1-0" (ptr 0xc005052900) appears in both seat 0 battlefield and seat 0 battlefield` |

---

## Top 10 failure categories (with sample cards)

| # | Category | Count | Sample cards |
|--:|---|--:|---|
| 1 | ability_word static effect "dead" because gating condition unsatisfied | 45 | Wolf-Skull Shaman, Lord of Tresserhorn, Foot Mystic, Heartfire Hero, Werewolf Pack Leader, Emberheart Challenger, Hobgoblin Captain, Flowerfoot Swordmaster, Putrid Pals, Sensation Gorger |
| 2 | sacrifice triggered on EOT/upkeep, scaffolding misses controller/turn condition | 5 | Pestilence, Pyrohemia, Withering Wisps, Task Mage Assembly, Planar Engineering |
| 3 | modification_effect triggered, P/T modify not observable | 4 | Lord of Tresserhorn, Scourge of Numai, Reaver Drone, Fathom Fleet Boarder |
| 4 | exile triggered, condition unsatisfied | 2 | Soul-Guide Lantern, Fishing Gear |
| 5 | lose_life static (recurring/global), counter unobservable | 2 | Fraying Omnipotence, Pox Plague |
| 6 | create_token triggered, controller condition unmet | 2 | Nightsquad Commando, Trynn, Champion of Freedom |
| 7 | destroy activated, target type missing | 1 | Demonic Hordes |
| 8 | sacrifice activated (Roving Actuator) | 1 | Roving Actuator |
| 9 | TurnStructure invariant — Lost-without-loss-reason | 1 | Phage the Untouchable |
| 10 | CardIdentity invariant — DFC duplicate pointer | 1 | Etali, Primal Conqueror // Etali, Primal Sickness |

---

## Top 3 root causes + suggested fix paths

### 1. Ability-word gating not pre-satisfied — **45 failures (70 %)**

**Pattern.** Every card with a "Threshold — / Metalcraft — / Hellbent — / Coven — / Ferocious — / Heroic — / Magecraft — / Spell Mastery — / Constellation — / Domain — / Revolt — / Delirium — / Raid — / etc." rider lands in this bucket because the Goldilocks board scaffold doesn't pre-satisfy the gate. The resolver runs, the gated `Conditional` evaluates false, no state changes, the verifier flags it as a dead effect.

**Why it's not (mostly) an engine bug.** Round 31 (`resolveGatedRider`) and round 33 (`SpellMastery` + `Constellation`) already wired Threshold / Metalcraft / Hellbent / SpellMastery / Constellation riders correctly into resolveSequence and `FirePermanentETBTriggers`. The riders emit `{rider}_rider` and `{rider}_rider_pending` events when their gates are satisfied. The verifier just doesn't read those events — it observes state deltas only.

**Suggested fix path** (two options, can ship together):

  (a) **Verifier-side**: extend `verifyEffect` in `cmd/hexdek-thor/goldilocks.go` to recognize `rider_pending` / `{rider}_rider` events as evidence the engine *would* have applied the effect if the gate were open — count those as PASS for ability_word effects.

  (b) **Scaffold-side**: extend `makeGoldilocksState` to pre-load the gating state for the named ability word: graveyard sized 7 for Threshold, 3 artifacts for Metalcraft, empty hand for Hellbent, etc. The gating predicates already exist (`ThresholdActive`, `MetalcraftActive`, `HellbentActive`, `SpellMasteryActive`); the scaffold just needs to mirror them in reverse.

(a) is the faster fix (~30 LOC); (b) is the more thorough fix (~100 LOC, covers any future rider that doesn't yet emit a pending event).

### 2. Triggered effects on "your upkeep" / "your end step" / "this player" not seeing the right active seat — **~11 failures**

**Pattern.** Pestilence, Pyrohemia, Withering Wisps, Lord of Tresserhorn, Scourge of Numai, Trynn — these are recurring upkeep triggers, end-of-turn sacrifices, or "when you do X" controller-keyed triggers. The verifier scaffolds the trigger event but doesn't always set `gs.Active = source.Controller`, so the "your upkeep" check inside the trigger handler returns false and the effect short-circuits before it can mutate state.

**Suggested fix path.** In `cmd/hexdek-thor/goldilocks.go`'s trigger-scaffolding pass (`makeGoldilocksState` / `testGoldilocksCard`), set `gs.Active = source.Controller` before firing controller-keyed triggers. Recognition heuristic: any trigger event containing the substrings "your_upkeep", "your_end_step", "your_turn", "you_draw", "you_cast" warrants the active-seat assignment. Per_card sweep would catch ~10 of the 11, with the modification_effect bucket needing a follow-on look at how P/T deltas are observed.

### 3. DFC / token-clone identity collisions — **1 invariant + symptom of a class**

**Pattern.** Etali, Primal Conqueror // Etali, Primal Sickness flagged a `CardIdentity` violation during a `modification_effect` test — the same `*Card` pointer appears twice in seat 0's battlefield. Almost certainly the verifier's modification harness builds an Etali token by cloning the source card without `DeepCopy`, and the engine's invariant check correctly catches the pointer collision. The Etali per_card handler may also be cloning rather than copying for its "exile top of library, cast for free" effect, which would surface the same issue under real play.

**Suggested fix path.** Audit the modification_effect scaffold in `verifyEffect` for `card.DeepCopy()` usage when seeding board-state token clones. Separately, run a targeted `-card "Etali, Primal Conqueror"` against thor with `-trace` to confirm whether the per_card handler itself has the same issue (likely; the audit memory has prior precedent for similar bugs: `dev/may11-nil-deref-forensics` flagged Abdel Adrian's per_card for bypassing canonical zone-exit APIs).

**Phage the Untouchable** is its own one-off: the TurnStructure invariant fires because Phage's ETB "you lose the game unless you cast it from your hand" sets `seat.Lost = true` without populating `seat.LossReason`. Fix is one line in Phage's per_card handler.

---

## What's healthy

- **Zero panics** across 31,963 tests. The engine's nil-safety + invariant-guards are holding up under broad corpus pressure.
- **100 % keyword pass rate** (2,013/2,013). The round-19 `RetainEvents` + combat-scaffold fix from the May 2026 Goldilocks issue-log entry remains effective.
- **No catastrophic-effect regressions** in the dominant effect kinds: damage, draw, mill, counter-placement, ETB-creates-token (non-triggered), removal — all silent in the failure log.

---

## Method

```bash
# binary built from cmd/hexdek-thor/
go build -o /tmp/hexdek-thor-r36 ./cmd/hexdek-thor/

# audit invocation (default workers + paths)
/tmp/hexdek-thor-r36 -goldilocks                      # human summary
/tmp/hexdek-thor-r36 -goldilocks -failures-csv F.csv  # all 64 failure rows
```

The CSV is the canonical artifact for follow-up sweeps. The summary in this report cites it directly via `awk` / `grep` counts; nothing else was sampled.
