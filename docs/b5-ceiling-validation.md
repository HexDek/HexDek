# B5 Combo Ceiling — Validation Report

**Date:** 2026-05-09
**Branch:** `dev/b5-validation`
**Fixes under test:** PR #15 (MDFC commander back-face cast) + PR #18 (combo sequencer command-zone + DFC face aliasing)

## Summary

Four-commander focused gauntlets confirm the B5 combo ceiling has lifted. Every PR-#18 problem
commander now wins its pod against three strong B5 opponents, in a uniform 4-deck round-robin
under the same hat policy and seed family.

| Commander | Pod WR (100 games) | Wins | Avg turn to win | Bracket of test deck |
| --- | ---: | ---: | ---: | --- |
| Najeela, the Blade-Blossom | **57.0%** | 57 | 49.2 | b2 |
| Kinnan, Bonder Prodigy | **52.0%** | 52 | 35.5 | b5 |
| Esika, God of the Tree // The Prismatic Bridge | **47.0%** | 47 | 55.7 | b4 |
| Korvold, Fae-Cursed King | **46.0%** | 46 | 44.3 | b3 |

PR #18 explicitly called out a "13–14% WR cluster" and PR #15 cited a 9.2% Esika regression. The
post-fix numbers above are 3–6× those baselines and well above the 25% even-share line for a 4-seat
pod.

## Methodology

- Tool: `cmd/hexdek-tournament` (Phase 11 parallel runner).
- Hat: `yggdrasil` default budget (50), default noise (0.20).
- Seats: 4. Games per pod: 100. CommanderMode: on.
- Each pod = the target commander + the same three fixed B5 opponents:
  - `atraxa_grand_unifier_b5_on_the_stack`
  - `arcum_dagsson_b5_diz`
  - `tymna_the_weaver_b5_ezinho` (from `data/decks/moxfield_300/`)
- Distinct master seed per pod (2001–2004) so RNG draws don't repeat.

Run script: `/tmp/b5-validation/run-all.sh`.
Per-pod reports: `/tmp/b5-validation/{esika,kinnan,najeela,korvold}.md`.

## Per-pod detail

### Esika, God of the Tree // The Prismatic Bridge (seed 2001)
```
1. Esika                      47.0%  (47/100)   avg turn 55.7
2. Atraxa, Grand Unifier      29.0%  (29/100)
3. Arcum Dagsson              16.0%  (16/100)
4. Tymna the Weaver            8.0%   (8/100)
```
Direct read on PR #15: the MDFC commander now actually casts as Bridge and resolves its upkeep
trigger, instead of entering as the front-face creature. The 9.2% regression flagged in PR #15 is
gone — Esika top-of-pod despite being the only b4 deck in a b5 field.

### Kinnan, Bonder Prodigy (seed 2002)
```
1. Kinnan                     52.0%  (52/100)   avg turn 35.5
2. Atraxa, Grand Unifier      25.0%  (25/100)
3. Arcum Dagsson              15.0%  (15/100)
4. Tymna the Weaver            8.0%   (8/100)
```
Fastest avg turn-to-win in the set (35.5) — consistent with Kinnan's mana-doubling combo lines
(Basalt Monolith, Great Whale lock, etc.) firing once command-zone availability is recognized.

### Najeela, the Blade-Blossom (seed 2003)
```
1. Najeela                    57.0%  (57/100)   avg turn 49.2
2. Atraxa, Grand Unifier      25.0%  (25/100)
3. Arcum Dagsson              13.0%  (13/100)
4. Tymna the Weaver            5.0%   (5/100)
```
Highest WR in the set despite the test deck only being b2. Najeela's combat-trigger combo loop
is now reachable from command zone.

### Korvold, Fae-Cursed King (seed 2004)
```
1. Korvold                    46.0%  (46/100)   avg turn 44.3
2. Atraxa, Grand Unifier      31.0%  (31/100)
3. Arcum Dagsson              15.0%  (15/100)
4. Tymna the Weaver            8.0%   (8/100)
```
b3 deck still tops a b5 pod. Smallest absolute lead in the set (b3 deck is the floor here),
but still 1.8× the next-best seat.

## Stability

- 0 crashes across all 400 games.
- 0 concessions.
- All games resolved in normal turn budgets (avg turns 39–57 across pods).
- DFC face aliases visibly resolved during Esika upkeep — no `CardIdentity` violations surfaced.

## Caveats

1. **Bracket asymmetry.** Of the four target decks, only Kinnan is true b5 in the local corpus.
   Esika is b4, Korvold is b3, Najeela is b2. Three of the four targets win their pods despite a
   bracket disadvantage, which strengthens the conclusion — but a true b5-on-b5-on-b5-on-b5 panel
   would be the cleanest proof. Add b5 versions of Esika/Korvold/Najeela when corpus permits.
2. **Freya strategy not loaded.** All four runs logged "no Freya analysis — Hat will play without
   strategy intelligence" for every seat. The warning is a UX artifact (`checkFreyaData` in
   `cmd/hexdek-tournament/main.go:51` looks at a different path than `hat.LoadStrategyFromFreya`),
   but the strategy load is genuinely empty in these runs. Because all four seats are equally
   affected within each pod, the relative WR comparison is valid; absolute numbers may shift once
   Freya plumbing is fixed.
3. **Single fixed opponent triple.** Same three seats across all four pods controls for opponent
   variance but doesn't sample matchups against, e.g., other ramp combo decks. The combo-ceiling
   claim is "these four commanders can execute their gameplan," which the data supports; a broader
   meta WR would need a wider opponent pool.
4. **Sample size.** 100 games per pod gives roughly ±10pp 95% CI on a 50% rate — comfortably tight
   relative to the 25%-baseline gap, but not tight enough to distinguish 47% from 52%.

## Conclusion

The B5 combo ceiling described in PR #18 is gone for the four named commanders. The combo
sequencer recognizes command-zone pieces, MDFC face aliases match in the zone index, and Esika's
back-face cast resolves correctly. No new crashes or invariant violations surfaced in 400 games.
