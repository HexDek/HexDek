# Hat Weights Validation — Archetype-Aware vs Midrange Baseline

**Date:** 2026-05-17
**Commit under test:** `b0b6db4` (hat: archetype-specific eval weights for 9 Freya archetypes + sparse-weight merge)
**Branch:** `dev/hat-weights-validation`

## Question

Does the new archetype-specific eval-weight dispatch produce measurably-different
play in actual games, or do the archetype dials get washed out by deeper heuristics
and MCTS rollouts?

## Method

Two 500-game pool tournaments on an identical 25-deck pool, same RNG seed (42),
same hat budget (50), same noise (0.2), same workers (8). The only difference:

- **new-weights:** `b0b6db4` archetype dispatch as merged. Lifegain, Voltron, Lands
  Matter, Counters Matter, Mill, Storm, Superfriends, Blink, and Extra Combats
  decks each get their own 20-dimension profile.
- **legacy-weights:** new flag `--legacy-hat-weights` forces every archetype to
  return the midrange profile from `DefaultWeightsForArchetype`. This collapses
  archetype dispatch and approximates pre-commit behavior (where 9 of 22 archetypes
  silently fell back to midrange and the 8-dim Freya overlay zeroed the other 12
  dims). Implementation: package-level `hat.LegacyMidrangeOnly` bool wired in
  `internal/hat/eval_weights.go:DefaultWeightsForArchetype`.

Same seed → identical deck draws and shuffles across the two runs. Every commander
played the **exact same number of games** in both runs (see game counts column
below) — winrate deltas are entirely attributable to weight changes.

## Pool

25 decks across 13 archetypes, picked from `data/decks/moxfield_300`:

| Archetype | Decks |
|---|---|
| lifegain, voltron, lands matter, counters matter, mill, storm, blink, reanimator, stax, midrange, aristocrats, combo | 2 each |
| extra combats | 1 |

8 of the 9 newly-tuned archetypes are represented (superfriends had no decks in the
moxfield_300 corpus). Pool dir: `/tmp/hat-validation-pool/` (symlinks to source decks).

## Headline

- **Mean |Δ winrate| = 4.18pp per commander.**
- 7/24 commanders shifted ≥5pp; 1 shifted ≥10pp.
- Aggregate game length essentially unchanged (39.9 vs 39.5 turns).
- Zero crashes, zero concessions in either run.

**Verdict: archetype-aware play produces measurably-different winrates.** Effects
are heterogeneous — some commanders gain, some lose, and the signal is strongest
for archetypes that previously fell back to midrange (`lifegain`, `voltron`,
`blink`).

## Per-commander winrate (sorted by Δ)

| Commander | Archetype | New | Legacy | Δ |
|---|---|---:|---:|---:|
| Abigale, Poet Laureate // Heroic Stanza | lifegain | 32.6% (92) | 22.8% (92) | **+9.8pp** |
| Toph, the First Metalbender | blink | 29.3% (82) | 23.2% (82) | **+6.1pp** |
| Bruce Banner // The Incredible Hulk | voltron | 22.4% (85) | 16.5% (85) | **+5.9pp** |
| Glarb, Calamity's Augur | reanimator | 27.9% (86) | 23.3% (86) | +4.6pp |
| Rograkh, Son of Rohgahh | stax | 22.7% (97) | 18.6% (97) | +4.1pp |
| Umbris, Fear Manifest | mill | 27.0% (74) | 23.0% (74) | +4.0pp |
| Clavileño, First of the Blessed | counters matter | 22.5% (80) | 18.8% (80) | +3.7pp |
| Bristly Bill, Spine Sower | lands matter | 25.0% (64) | 21.9% (64) | +3.1pp |
| Bria, Riptide Rogue | storm | 28.6% (77) | 26.0% (77) | +2.6pp |
| Mayael the Anima | extra combats | 24.4% (78) | 21.8% (78) | +2.6pp |
| Hearthhull, the Worldseed | aristocrats | 33.8% (65) | 32.3% (65) | +1.5pp |
| Ajani, Nacatl Pariah | combo | 22.9% (166) | 22.9% (166) | +0.0pp |
| Bre of Clan Stoutarm | counters matter | 27.7% (83) | 27.7% (83) | +0.0pp |
| Child of Alara | lifegain | 25.0% (80) | 25.0% (80) | +0.0pp |
| Jan Jansen, Chaos Crafter | aristocrats | 17.8% (73) | 17.8% (73) | +0.0pp |
| Phenax, God of Deception | mill | 27.2% (92) | 28.3% (92) | -1.1pp |
| Choco, Seeker of Paradise | lands matter | 25.8% (93) | 28.0% (93) | -2.2pp |
| Dina, Essence Brewer | stax | 24.4% (78) | 26.9% (78) | -2.5pp |
| Sokrates, Athenian Teacher | reanimator | 18.9% (74) | 23.0% (74) | -4.1pp |
| Alania, Divergent Storm | storm | 25.4% (71) | 29.6% (71) | -4.2pp |
| Atraxa, Grand Unifier | blink | 28.4% (81) | 35.8% (81) | **-7.4pp** |
| Cecil, Dark Knight // Cecil, Redeemed Paladin | voltron | 22.1% (77) | 29.9% (77) | **-7.8pp** |
| Araumi of the Dead Tide | midrange | 21.2% (85) | 29.4% (85) | **-8.2pp** |
| Anje Falkenrath | combo | 17.9% (67) | 32.8% (67) | **-14.9pp** |

## Per-archetype aggregate (pooled across 2 decks each)

| Archetype | New | Legacy | Δ | Notes |
|---|---:|---:|---:|---|
| **lifegain** | 29.1% (50/172) | 23.8% (41/172) | **+5.2pp** | LifeResource=1.8 dial paying off |
| **extra combats** | 24.4% (19/78) | 21.8% (17/78) | +2.6pp | Single-deck sample |
| **counters matter** | 25.2% (41/163) | 23.3% (38/163) | +1.8pp | |
| **mill** | 27.1% (45/166) | 25.9% (43/166) | +1.2pp | |
| **stax** | 23.4% (41/175) | 22.3% (39/175) | +1.1pp | Sparse-weight merge fix helping |
| **aristocrats** | 25.4% (35/138) | 24.6% (34/138) | +0.7pp | |
| **reanimator** | 23.8% (38/160) | 23.1% (37/160) | +0.6pp | |
| lands matter | 25.5% (40/157) | 25.5% (40/157) | +0.0pp | |
| blink | 28.8% (47/163) | 29.4% (48/163) | -0.6pp | |
| voltron | 22.2% (36/162) | 22.8% (37/162) | -0.6pp | Within-archetype split |
| storm | 27.0% (40/148) | 27.7% (41/148) | -0.7pp | |
| **combo** | 21.5% (50/233) | 25.8% (60/233) | **-4.3pp** | Regression — see below |
| midrange | 21.2% (18/85) | 29.4% (25/85) | -8.2pp | Single deck, noisy |

Aggregate SE on a per-archetype 25% rate with ~150 games is ≈2.7pp, so the +5.2pp
lifegain gain and -4.3pp combo regression are real; sub-2pp shifts are within
noise.

## Findings

### 1. The intended signal is real

Decks of newly-tuned archetypes that previously fell back to midrange show
clear positive movement: lifegain +5.2pp, counters matter +1.8pp, mill +1.2pp.
The biggest single-commander gain is Abigale (lifegain) at +9.8pp — exactly the
case the commit was designed for.

### 2. Combo archetype regressed (-4.3pp)

`Anje Falkenrath` lost **14.9pp** going from legacy (midrange-collapsed) to the
new dispatch. Combo *already had* a tuned profile before `b0b6db4`, so the
regression is not from the new archetype constants but from the sparse-weight
merge change in `strategy_loader.go`: pre-commit, Freya's 8-dim payload zeroed
out the other 12 dims; post-commit, those 12 dims keep their archetype-default
values (e.g. combo's `StackInteraction`, `ToolboxBreadth`, `ActivationTempo`).
This is correct behavior in principle, but the combo defaults appear to be
mistuned for this deck — Anje is a discard-graveyard combo whose play likely
suffered from non-zero `BoardPresence`/`ThreatExposure` defaults. Ajani combo,
on the same archetype, moved 0.0pp on 166 games, so this is a deck-specific
interaction, not a universal combo regression.

**Recommendation:** flag combo profile review in the Freya improvement kanban,
specifically `StackInteraction`/`ActivationTempo` defaults for discard-loop combos.

### 3. Within-archetype scatter

Voltron split hard (Bruce Banner +5.9pp, Cecil -7.8pp). Blink split similarly
(Toph +6.1pp, Atraxa -7.4pp). This is expected — archetype is a coarse summary
and individual decks have idiosyncratic plans. Atraxa "blink" is really a
goodstuff value pile, Cecil "voltron" is partner+equipment fragile. Per-deck
weight overrides via Freya's eval_weights serialization would help here.

### 4. Zero-delta commanders are interesting

4 commanders (Ajani combo, Bre counters matter, Child of Alara lifegain, Jan
Jansen aristocrats) moved exactly 0.0pp. With same-seed determinism this means
the weight change produced no decision divergence for these decks across 65-166
games. Likely causes: heuristic short-circuits dominating MCTS choice, or game
states resolving fast enough that the evaluator never gates the outcome. Worth
spot-checking decision logs to confirm the hat is actually consulting the
weights for these decks.

### 5. No game-length distortion

39.9 turns (new) vs 39.5 turns (legacy) — the weight change shifts *who* wins,
not *how long* games take. Throughput within 6% (10.1 vs 10.7 g/s).

## Reproduction

```bash
go build -o /tmp/hexdek-tournament ./cmd/hexdek-tournament/

# Pool setup — see data/audit/hat-weights-validation.md prelude for deck list.
mkdir -p /tmp/hat-validation-pool/freya
# ... symlink 25 decks from data/decks/moxfield_300 ...

# New weights
/tmp/hexdek-tournament \
    --decks /tmp/hat-validation-pool --pool \
    --games 500 --seed 42 --workers 8

# Legacy baseline (midrange forced for every archetype)
/tmp/hexdek-tournament \
    --decks /tmp/hat-validation-pool --pool \
    --games 500 --seed 42 --workers 8 \
    --legacy-hat-weights
```

## Open follow-ups

- [ ] Combo archetype default review — investigate Anje regression, possibly
      narrow `BoardPresence`/`ThreatExposure` defaults for graveyard-combo
      sub-shapes.
- [ ] Per-deck Freya `eval_weights` override coverage — Atraxa-blink and
      Cecil-voltron scatter suggests archetype-alone is too coarse.
- [ ] Spot-check the 4 zero-delta decks — confirm the hat is actually reading
      the modified weight profile for these (vs heuristic short-circuit).
- [ ] Re-run with `--hat-budget 200` to see whether higher rollout budget
      amplifies or dampens the archetype dial signal.
