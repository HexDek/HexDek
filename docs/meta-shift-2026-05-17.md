# Meta-Shift Report — 2026-05-17 (vs morning bench)

## Goal

Measure whether today's merge wave — **100+ Muninn per_card handlers, the perf round-2 cache + combo-eval alloc fix, the goldilocks/Loki round-6 stability sweep, freya-quality-audit improvements, the visual/Heimdall side-tracks** — has measurably shifted the metagame compared to this morning's bench (`docs/postmerge-meta-bench-2026-05-16.md`).

## Method

- **Tool:** `cmd/hexdek-tournament`, hat = `yggdrasil` (budget 50, σ 0.2), commander on, seed 42, 16 workers
- **Pool:** Top **50** decks by `hex_rating` from a fresh `https://hexdek.dev/api/live/elo` snapshot (1,315-entry leaderboard, fetched at run start)
- **Games:** **1,000** in `--pool` mode, seats=4 — every game samples 4 random decks from the 50-deck pool
- **Decks:** all 50 located locally under `data/decks/moxfield{,_300}/`; all 50 Freya sidecars (`*_freya.md` + `*.strategy.json`) loaded cleanly — no "no Freya analysis" warnings
- **Engine head:** `419a054` (dev/meta-shift-report, branched from main `97f7e8f`; the head commit is an unrelated frontend polish landed by a sibling worker — backend identical to main)

Bench for comparison: 500 games, top-30 pool, branch `dev/postmerge-meta-bench` at `b6e6d43`.

### What is and isn't comparable

This is **not apples-to-apples** with the morning bench. Three things changed at once:

1. **Pool composition changed dramatically.** The morning top-30 collapsed to 12 unique commanders (Tymna ×8, Rograkh ×6, Kinnan ×5, …). Today's top-50 has **42 unique commanders** — 8 of them appear ≥2× (Kinnan/Rograkh/Tymna lead with 4/4/3 representatives), the other 34 are singletons. The field is **far broader**, which on its own pulls per-commander winrates toward the 25% no-skill mean.
2. **Top-of-board churned.** Of the 12 commanders that anchored the bench, **only 8 are still in this top-50** (Vial Smasher, Glarb, Inalla, Malcolm dropped out of the top-50 entirely). Hex-rating thresholds climbed from ~2974-3032 (bench) to **3018-3067** (now) — the live ELO board is in a different state than 16 hours ago.
3. **Pool mode still aggregates by commander, not by deck** (per bench Caveat #1). With 42 commanders sharing 50 deck-slots, per-commander signal is cleaner than it was at 12-of-30, but multi-deck commanders (Kinnan ≈4, Rograkh ≈4, Tymna ≈3) still aggregate across heterogeneous decklists.

Interpret every Δ below as **directional**, not causal.

## Field stability

| Metric | Bench (500g) | Today (1000g) | Δ |
|---|---:|---:|---:|
| Wall time | 47.0 s | 90.5 s | +93% (≈ linear scale w/ games) |
| Throughput | 10.6 g/s | **11.0 g/s** | **+4%** |
| Crashes | 0 | **0** | — |
| Concessions | 0 | **0** | — |
| Avg turns / game | 35.4 | **48.6** | **+13.2 (+37%)** |
| Unique commanders | 12 | 42 | +30 |

- **Zero crashes, zero concessions on 1,000 games and a more diverse pool** — no new failure modes from today's handler wave or perf cache surfaced. Goldilocks/Loki round-6 (the 0-panic clean) clearly held.
- **+4% throughput** is consistent with perf round 2's claimed −22% alloc / −13% CPU win (the headline number doesn't drop straight through to wall-clock because game-length grew).
- **Avg turns jumped from 35.4 → 48.6.** This is the largest signal in the run. Three plausible drivers; we cannot yet separate them:
  - **Pool-composition change** — the 50-deck pool draws in more midrange/value commanders (Yuna, Meren, Tayam, Y'shtola, The Wise Mothman) vs. the bench's Tymna/Rograkh/Kinnan-heavy cEDH spike. Longer games are expected.
  - **Handler wave makes more cards "work"** — newly-wired triggers/abilities convert previously-no-op cards into game-changing plays, extending decision space.
  - **Mana / threat reading improvements (Freya audit, hat archetype weights)** make AI play less suicidally aggressive, suppressing 20-turn race-outs.
  - **Suggested follow-up:** rerun the bench's exact top-30 pool at 1,000 games against today's `main`. If avg turns stays high there, it's a true engine shift; if it falls back to ~35, the diversity-of-pool hypothesis is the explanation.

## Top-50 pool (sorted by `hex_rating` desc, top 20 shown)

| Rk | Hex | LiveWR% | Games | Commander |
|---:|----:|--------:|------:|---|
|  1 | 3067.1 | 25.1 | 110,200 | Atraxa, Grand Unifier |
|  2 | 3057.6 | 24.7 | 120,585 | Niv-Mizzet Reborn |
|  3 | 3057.2 | 25.2 | 108,991 | Shiko, Paragon of the Way |
|  4 | 3055.6 | 25.1 | 109,926 | Rasputin Dreamweaver |
|  5 | 3055.1 | 24.6 | 121,422 | Ivy, Gleeful Spellthief |
|  6 | 3054.3 | 24.9 | 114,959 | Meren of Clan Nel Toth |
|  7 | 3053.4 | 24.4 | 132,154 | Kinnan, Bonder Prodigy |
|  8 | 3049.5 | 24.9 | 111,721 | The Reality Chip |
|  9 | 3048.6 | 24.6 | 125,003 | Kilo, Apogee Mind |
| 10 | 3048.4 | 24.7 | 119,272 | Raffine, Scheming Seer |
| 11 | 3047.4 | 25.5 | 130,186 | Maelstrom Wanderer |
| 12 | 3043.3 | 24.8 | 115,806 | Yuna, Grand Summoner |
| 13 | 3042.7 | 24.9 | 110,640 | Prismari, the Inspiration |
| 14 | 3042.6 | 25.1 | 108,955 | Emiel the Blessed |
| 15 | 3040.5 | 25.1 | 109,119 | Derevi, Empyrial Tactician |
| 16 | 3038.3 | 24.9 | 115,152 | Anje Falkenrath |
| 17 | 3038.0 | 24.9 | 114,469 | Esika, God of the Tree |
| 18 | 3035.9 | 24.8 | 115,591 | Abdel Adrian, Gorion's Ward |
| 19 | 3035.4 | 24.5 | 130,573 | Tymna the Weaver |
| 20 | 3034.8 | 24.6 | 129,710 | Shorikai, Genesis Engine |

Full pool: `/tmp/metashift/top50.json`. Freya archetype mix across the 50: midrange ×13, combo ×13, storm ×7, stax ×7, control ×4, artifacts ×4, aggro/voltron ×2.

## 1000-game winrates (full 42-commander table)

```
 1. Yuna, Grand Summoner                      37.5%  (30/80)
 2. Niv-Mizzet Reborn                         33.8%  (27/80)
 3. Magus Lucea Kane                          32.6%  (28/86)
 4. Gwenom, Remorseless                       32.3%  (21/65)
 5. Derevi, Empyrial Tactician                32.3%  (20/62)
 6. Edgar Markov                              31.4%  (22/70)
 7. Esika, God of the Tree                    31.4%  (27/86)
 8. Tayam, Luminous Enigma                    29.9%  (23/77)
 9. Gerrard, Weatherlight Hero                29.7%  (27/91)
10. Y'shtola, Night's Blessed                 29.0%  (27/93)
11. Abdel Adrian, Gorion's Ward               28.8%  (23/80)
12. Zurgo and Ojutai                          28.3%  (26/92)
13. The Wise Mothman                          28.0%  (23/82)
14. Rograkh, Son of Rohgahh                   27.9%  (77/276)
15. Rasputin Dreamweaver                      27.4%  (23/84)
16. Kinnan, Bonder Prodigy                    27.3%  (85/311)
17. Meren of Clan Nel Toth                    27.1%  (23/85)
18. Sidar Kondo of Jamuraa                    26.7%  (23/86)
19. Shiko, Paragon of the Way                 26.4%  (28/106)
20. Vivi Ornitier                             25.3%  (21/83)
21. Prismari, the Inspiration                 25.0%  (20/80)
22. Shorikai, Genesis Engine                  24.4%  (21/86)
23. Flubs, the Fool                           24.0%  (18/75)
24. Ivy, Gleeful Spellthief                   23.5%  (16/68)
25. Etali, Primal Conqueror                   23.3%  (31/133)
26. Raffine, Scheming Seer                    23.2%  (19/82)
27. Tymna the Weaver                          22.9%  (57/249)
28. Fire Lord Azula                           22.6%  (14/62)
29. Emiel the Blessed                         22.5%  (16/71)
30. Gorma, the Gullet                         22.4%  (17/76)
31. Hashaton, Scarab's Fist                   21.9%  (16/73)
32. Atraxa, Grand Unifier                     20.7%  (18/87)
33. The Reality Chip                          20.5%  (17/83)
34. Selvala, Explorer Returned                20.3%  (15/74)
35. Anje Falkenrath                           19.8%  (17/86)
36. Thrasios, Triton Hero                     18.2%  (16/88)
37. Kilo, Apogee Mind                         17.7%  (14/79)
38. Norman Osborn // Green Goblin             17.1%  (14/82)
39. Kenrith, the Returned King                16.4%  (11/67)
40. Maelstrom Wanderer                        16.0%  (12/75)
41. Jhoira, Weatherlight Captain              14.5%  (11/76)
42. Francisco, Fowl Marauder                   8.2%  (6/73)
```

Coverage: 42/42 commanders played at least 1 game. Game-count range 62–311 (multi-deck commanders ≫ singletons). 4-seat random pool no-skill mean is 25%, so the field is approximately symmetric (mean ≈ 23.8% over 42 commanders) with a clear top/tail spread of ~30pp.

## Bench-to-1k delta on the 12 commanders shared with the morning bench

8 of the bench's 12 commanders also appear in the new top-50 pool. The other 4 (Vial Smasher, Glarb, Inalla, Malcolm) fell out of the top-50 entirely and have no 1k comparison.

| Commander | Bench WR% | 1k WR% | Δwr | Bench Rk (of 12) | 1k Rk (of 42) | Bench %ile | 1k %ile | Δ%ile |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Thrasios, Triton Hero          | 32.3 | **18.2** | **−14.1** |  1 | 36 | 0.08 | 0.86 | **−0.77** |
| Hashaton, Scarab's Fist        | 15.9 | **21.9** | **+6.0**  | 12 | 31 | 1.00 | 0.74 | **+0.26** |
| Tymna the Weaver               | 26.7 | 22.9 | −3.8 |  3 | 27 | 0.25 | 0.64 | −0.39 |
| Rograkh, Son of Rohgahh        | 24.2 | 27.9 | +3.7 |  9 | 14 | 0.75 | 0.33 | +0.42 |
| Kinnan, Bonder Prodigy         | 23.8 | 27.3 | +3.5 | 10 | 16 | 0.83 | 0.38 | +0.45 |
| Y'shtola, Night's Blessed      | 25.9 | 29.0 | +3.1 |  5 | 10 | 0.42 | 0.24 | +0.18 |
| Norman Osborn // Green Goblin  | 20.0 | 17.1 | −2.9 | 11 | 38 | 0.92 | 0.90 | +0.01 |
| Shorikai, Genesis Engine       | 25.4 | 24.4 | −1.0 |  7 | 22 | 0.58 | 0.52 | +0.06 |
| Vial Smasher the Fierce        | 27.0 | — | — |  2 | — | 0.17 | — | — |
| Glarb, Calamity's Augur        | 26.5 | — | — |  4 | — | 0.33 | — | — |
| Inalla, Archmage Ritualist     | 25.8 | — | — |  6 | — | 0.50 | — | — |
| Malcolm, Keen-Eyed Navigator   | 24.6 | — | — |  8 | — | 0.67 | — | — |

`%ile` = rank / N within run; lower is better. `Δ%ile = bench%ile − 1k%ile`; positive = commander is **relatively higher-ranked** today than in the bench.

## Material shifts

### **Thrasios crashed −14.1pp (32.3% → 18.2%).** 
- Bench: rank 1 of 12 (top 8%), based on 65 games. 1k: rank 36 of 42 (bottom 14%), based on 88 games.
- The bench called this out explicitly: "Single deck, 65 games — wide CI". The 1k data confirms the bench number was small-sample noise on the high side, not a real merge-induced spike. **Thrasios's live-board winrate (16.5%) is closer to its true engine performance than the bench bar.**
- No causal attribution to merges; this is a Type-I-error correction.

### **Hashaton recovered +6.0pp (15.9% → 21.9%).** 
- Bench: rank 12 of 12 (worst), 69 games. 1k: rank 31 of 42 (bottom 26%), 73 games.
- Still below the 25% no-skill mean, but **no longer a tail outlier** (Hashaton is now ahead of Atraxa, The Reality Chip, Anje Falkenrath, Thrasios, Kilo, Norman Osborn, Kenrith, Maelstrom Wanderer, Jhoira, Francisco).
- This is consistent with the bench's flagged follow-up: "Hashaton is reanimator-adjacent; today's merges include `dev/muninn-top5-handlers` (Necromancy, Bloodchief, Kodama)…" Plus the Resolved-entry "Graveyard recursion intelligence gated behind ArchetypeReanimator" hat fix and the Dread shuffle-into-library invariant fix (2026-05-08). Whatever was kneecapping Hashaton's reanimator plumbing appears partially repaired.
- **Strongest defensible "merges helped" signal in the run.** Still 73 games — recommend a follow-up isolated rerun before claiming the recovery is fully attributable.

### **Kinnan / Rograkh: meaningful relative recovery (Δ%ile +0.45 / +0.42).**
- Both moved from bottom-third in the bench (relative to its 12-cmdr field) to mid-pack in the 1k (relative to its 42-cmdr field).
- Absolute Δwr is +3.5pp and +3.7pp respectively. Each commander has 4 distinct decks in the pool and 276-311 games — these are the **most sample-robust commanders in the run.**
- Interpretation: Kinnan and Rograkh both lean heavily on consistent ramp + threat sequencing. The combo-eval alloc fix, the bitmask color cache (perf round 2), and any handler wins that unblocked their main lines would all push WR up modestly. Plausible that merge wave helped, but with the pool-composition caveat above, the field around them also shifted.

### **Tymna regressed −3.8pp (Δ%ile −0.39).**
- Bench rank 3, 1k rank 27. 8 decks in bench → 3 in 1k (5 dropped off the leaderboard).
- The 3 surviving Tymna decks may simply be the more conservative builds; the cEDH-leaning ones that gave bench its 26.7% number aren't in today's pool.
- **Not interpretable as a merge regression.** Different decklist mix.

### **Yuna leads at 37.5%, well above any commander on the bench.**
- Yuna had 0 games in bench (commander not in top-30). At 80 games in the 1k pool the WR is well above the field mean. Worth a 1k isolated rerun (gauntlet) to confirm.

### **Francisco, Fowl Marauder tails at 8.2% (6/73).**
- Lowest WR in the field by 6pp. New commander, freya archetype unrevealed in this report — flag for Loki/Goldilocks targeted check that no Francisco-specific scaffold is mis-firing.

## Caveats and known limitations

1. **Pool composition is not held constant.** The leaderboard moved between morning and now. This is the dominant confound.
2. **Per-commander aggregation, not per-deck.** A 4-deck commander's 27.9% WR averages across heterogeneous decklists with different combo counts, archetypes, and mana bases.
3. **`runPool` still does not write a markdown report.** `internal/tournament/runner.go:1058` returns a `TournamentResult` but no `WriteMarkdown` call exists on the pool path. The CLI's `"Report written to /tmp/…"` message remains a false success (bench Caveat #2 still open). This report is composed from captured stdout (`/tmp/metashift/tourney.log`).
4. **`--matchup` still not produced in pool mode** — bench Caveat #3 still open. Not requested in this run.
5. **Determinism note.** Two consecutive 1k runs at seed=42 in this session produced different per-commander winrates (e.g., Yuna 26.2% vs 37.5%, Niv-Mizzet 35.0% vs 33.8%). Per-game outcomes are *not* reproducible across runs — most likely because worker goroutines pick games off the seed channel in non-deterministic order and the master seed only seeds the per-game seed *value*, not the per-game *assignment*. Bench reproducibility against today's run is therefore not guaranteed even on bench's exact pool. **Worth a separate ticket.** Suggested fix: seed per-game with `cfg.Seed + gameIdx` rather than relying on FIFO consumption order.
6. **Print cap.** `runner.go:1039` hard-codes a 30-commander print limit. The full-42 listing in this report was captured by a temporary local patch (since reverted) that printed the rank-31+ tail. Suggest replacing with `--top N` flag or dumping a JSON commander table by default.
7. **Game-count floor for shift interpretation.** Commanders with <80 games (24 of 42) have ±5pp 95%-CI on winrate; treat their shifts as noise unless ≥6pp absolute.

## Recommended follow-ups

1. **Bench-pool rerun.** Run today's `main` against the bench's exact 30-deck top-30 pool at 1,000 games. This isolates engine-change effect from pool-change effect. If avg-turns and per-commander Δwr persist, attribute to merges; if they don't, attribute to leaderboard composition.
2. **Hashaton isolated rerun.** 4-deck round-robin at 2,000 games, Hashaton vs. 3 strong baseline cEDH commanders. Confirm the +6.0pp is engine-level, not field-mix.
3. **Yuna and Francisco isolated reruns.** Both are likely small-sample artifacts (Yuna high, Francisco low) but need confirmation since both are non-trivial leaderboard fixtures now.
4. **Fix `runPool` → `WriteMarkdown`.** Both runs (bench and today) have lost the canonical machine-readable report. Should also emit per-commander JSON for diff tooling.
5. **Investigate game-length jump.** 35.4 → 48.6 turns is the most striking field-wide change. Either real engine shift (good, but worth understanding) or pool-mix artifact. Bench-pool rerun (#1) will disambiguate.
6. **Pool-mode determinism.** Pick per-game seed = `cfg.Seed + gameIdx` so two `--seed 42` runs of the same pool produce identical aggregate results. Currently they don't.

## Raw artifacts

- Tournament stdout: `/tmp/metashift/tourney.log`
- Leaderboard snapshot: `/tmp/metashift/elo.json` (1,315 entries, fetched 2026-05-16 23:21 PDT)
- Top-50 JSON: `/tmp/metashift/top50.json`
- Staged decks + Freya sidecars: `/tmp/metashift/decks/` (50 deck symlinks + `freya/` subdir with 50 `.md` + 50 `.strategy.json`)
- Tournament binary built from `419a054`: `/tmp/metashift/hexdek-tournament`
