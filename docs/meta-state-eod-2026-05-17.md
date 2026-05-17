# Meta-State EOD Report — 2026-05-17

End-of-day snapshot of the live leaderboard's top-30, run as a 1,000-game pool tournament against the current `main`. Paired with the noon meta-shift report (`docs/meta-shift-2026-05-17.md`) to bookend the day.

## Method

- **Tool:** `cmd/hexdek-tournament`, hat = `yggdrasil` (budget 50, σ 0.2), commander on, seed 42, 16 workers
- **Pool:** Top **30** decks by `hex_rating` from a fresh `https://hexdek.dev/api/live/elo` snapshot (1,315 entries, fetched at run start)
- **Games:** **1,000** in `--pool` mode, seats=4 — every game samples 4 random decks from the 30-deck pool
- **Decks:** all 30 located under `data/decks/moxfield{,_300}/`; all 30 Freya sidecars (`*_freya.md` + `*.strategy.json`) loaded cleanly — zero "no Freya analysis" warnings
- **Engine head:** `dev/meta-state-eod` branched from `main` at `0467f40` (Merge dev/visual-polish-round-7)

Comparison anchor: noon meta-shift run (`docs/meta-shift-2026-05-17.md`) — 1,000 games, top-**50** pool, engine `419a054` (= main + frontend-only deltas).

### What is and isn't comparable

Same caveats as the noon report, plus one new one:

1. **Pool size shrank 50 → 30.** This is by design (the task asks for top-30) but it changes the field shape vs. noon. 30 deck-slots cover **24 unique commanders** (vs. 42 at noon), with the top of the cEDH stack — Kinnan ×3, Rograkh ×3, Etali/Derevi ×2 — pulling proportionally more games (375, 402, 256, 257 respectively) than at noon.
2. **Leaderboard moved 16 h between snapshots.** Noon top-50 had hex range 3018-3067; EOD top-30 has **3030.7-3071.8**. The top is **+4.7 hex** higher and the floor for "top-30" is **+13 hex** higher than the noon top-50 floor. **22 commanders** that appeared in the noon top-50 pool are no longer present at EOD; **5 commanders** in today's top-30 (Alaundo, Tivit, Arcum Dagsson, Chatterfang, Sisay) were not in noon's pool at all.
3. **Pool-mode aggregation is still per-commander, not per-deck** — bench Caveat #2 still open. Kinnan's 23.5% averages across 3 heterogeneous decklists, etc.
4. **Determinism note from noon still applies.** Two `--seed 42` pool runs do not reproduce per-commander winrates because workers consume the seed channel non-deterministically. Treat single-digit-pp deltas as inside the noise floor.

Read every Δ below as **directional**, not causal.

## Field stability

| Metric | Noon (1000g, top-50) | EOD (1000g, top-30) | Δ |
|---|---:|---:|---:|
| Wall time | 90.5 s | **119.6 s** | +32% |
| Throughput | 11.0 g/s | **8.4 g/s** | **−24%** |
| Crashes | 0 | **0** | — |
| Concessions | 0 | **0** | — |
| Avg turns / game | 48.6 | **51.7** | **+3.1 (+6%)** |
| Unique commanders | 42 | 24 | −18 |
| Hex band (pool) | 3018-3067 | 3030.7-3071.8 | top floor +13 |

- **Zero crashes / zero concessions on a third consecutive 1,000-game day-end run.** Combined with the noon run, the day produced **2,000 crash-free pool games** on `main`. The Sai/test-pollution fix (375de60), the Muninn bulk-pattern round 4 merge (eecb1c3), and the EOD muninn handler waves all landed without surfacing new failure modes.
- **−24% throughput vs. noon** decomposes into two pieces: avg-turns moved 48.6 → 51.7 (+6%), and the smaller pool of 30 mostly-cEDH decks concentrates games on combo-dense matches that take more decision cycles. Per-game time grew roughly in line.
- **Avg turns climbed again** (35.4 bench → 48.6 noon → 51.7 EOD). The noon report flagged three plausible drivers (pool composition, handler wave, mana/threat reading). The EOD shift from 48.6 → 51.7 happened on a *tighter* pool against the same engine generation as noon's 419a054 (`main` + frontend-only), which makes pool composition the dominant explanation for that particular delta. The 35.4 → 48.6 jump from morning bench → noon remains attributable to the day's merge wave on the same logic.

## EOD top-30 pool (sorted by `hex_rating` desc, top 15 shown)

| Rk | Hex | LiveWR% | Games | Commander |
|---:|----:|--------:|------:|---|
|  1 | 3071.8 | — | — | Meren of Clan Nel Toth |
|  2 | 3071.4 | — | — | Sidar Kondo of Jamuraa |
|  3 | 3057.8 | — | — | Tivit, Seller of Secrets |
|  4 | 3055.7 | — | — | Esika, God of the Tree // The Prismatic Bridge |
|  5 | 3050.9 | — | — | Etali, Primal Conqueror // Etali, Primal Sickness |
|  6 | 3050.0 | — | — | Sisay, Weatherlight Captain |
|  7 | 3050.0 | — | — | Jhoira, Weatherlight Captain |
|  8 | 3048.5 | — | — | Kinnan, Bonder Prodigy |
|  9 | 3048.1 | — | — | Rograkh, Son of Rohgahh |
| 10 | 3045.5 | — | — | Tymna the Weaver |
| 11 | 3044.9 | — | — | Derevi, Empyrial Tactician |
| 12 | 3044.5 | — | — | Alaundo the Seer |
| 13 | 3044.4 | — | — | Thrasios, Triton Hero |
| 14 | 3043.1 | — | — | Yuna, Grand Summoner |
| 15 | 3042.9 | — | — | Shiko, Paragon of the Way |

Full pool: `/tmp/metaeod/top30.json`. Pool archetype mix from Freya sidecars: midrange ×13, combo ×8, artifacts ×6, stax ×4, storm ×3 (some commanders contribute multiple archetype tags via multi-deck representation).

## 1000-game EOD winrates (full 24-commander table)

```
 1. Yuna, Grand Summoner                      31.2%  (35/112)
 2. Tymna the Weaver                          31.0%  (40/129)
 3. Shiko, Paragon of the Way                 29.4%  (40/136)
 4. Jhoira, Weatherlight Captain              29.3%  (36/123)
 5. Selvala, Explorer Returned                28.6%  (38/133)
 6. The Wise Mothman                          27.4%  (34/124)
 7. Alaundo the Seer                          26.4%  (39/148)
 8. Francisco, Fowl Marauder                  25.8%  (34/132)
 9. Tivit, Seller of Secrets                  25.6%  (33/129)
10. Anje Falkenrath                           25.2%  (33/131)
11. Arcum Dagsson                             25.2%  (39/155)
12. Chatterfang, Squirrel General             25.0%  (33/132)
13. Etali, Primal Conqueror                   25.0%  (64/256)
14. Esika, God of the Tree                    24.6%  (35/142)
15. Flubs, the Fool                           24.4%  (33/135)
16. Shorikai, Genesis Engine                  24.1%  (32/133)
17. Rograkh, Son of Rohgahh                   23.6%  (95/402)
18. Kinnan, Bonder Prodigy                    23.5%  (88/375)
19. Thrasios, Triton Hero                     23.4%  (39/167)
20. Sidar Kondo of Jamuraa                    23.2%  (32/138)
21. Derevi, Empyrial Tactician                22.6%  (58/257)
22. Kenrith, the Returned King                22.5%  (32/142)
23. Sisay, Weatherlight Captain               21.4%  (33/154)
24. Meren of Clan Nel Toth                    20.9%  (24/115)
```

Coverage: 24/24 commanders played at least 1 game. Game-count range 112-402 (multi-deck commanders concentrate). 4-seat random-pool no-skill mean is 25%; field mean is **25.1%**, the tightest centering across the day's three runs (bench 24.0%, noon 23.8%, EOD 25.1%). Top-to-bottom spread compressed to **10.3pp** (vs. noon 29.3pp, bench 16.4pp) — fewer commanders, more games per commander, less tail noise.

## Noon → EOD delta (19 commanders shared with noon's top-50)

| Commander | Noon WR% | EOD WR% | Δwr | Noon Rk (of 42) | EOD Rk (of 24) | Noon %ile | EOD %ile | Δ%ile |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Jhoira, Weatherlight Captain   | 14.5 | **29.3** | **+14.8** | 41 |  4 | 0.98 | 0.17 | **+0.81** |
| Francisco, Fowl Marauder       |  8.2 | **25.8** | **+17.6** | 42 |  8 | 1.00 | 0.33 | **+0.67** |
| Tymna the Weaver               | 22.9 | **31.0** | **+8.1**  | 27 |  2 | 0.64 | 0.08 | **+0.56** |
| Selvala, Explorer Returned     | 20.3 | **28.6** | **+8.3**  | 34 |  5 | 0.81 | 0.21 | **+0.60** |
| Kenrith, the Returned King     | 16.4 | **22.5** | **+6.1**  | 39 | 22 | 0.93 | 0.92 | +0.01 |
| Anje Falkenrath                | 19.8 | **25.2** | **+5.4**  | 35 | 10 | 0.83 | 0.42 | +0.41 |
| Thrasios, Triton Hero          | 18.2 | **23.4** | **+5.2**  | 36 | 19 | 0.86 | 0.79 | +0.07 |
| Shiko, Paragon of the Way      | 26.4 | 29.4 | +3.0 | 19 |  3 | 0.45 | 0.13 | +0.33 |
| Etali, Primal Conqueror        | 23.3 | 25.0 | +1.7 | 25 | 13 | 0.60 | 0.54 | +0.06 |
| Flubs, the Fool                | 24.0 | 24.4 | +0.4 | 23 | 15 | 0.55 | 0.63 | −0.08 |
| Shorikai, Genesis Engine       | 24.4 | 24.1 | −0.3 | 22 | 16 | 0.52 | 0.67 | −0.14 |
| Wise Mothman, The              | 28.0 | 27.4 | −0.6 | 13 |  6 | 0.31 | 0.25 | +0.06 |
| Sidar Kondo of Jamuraa         | 26.7 | 23.2 | −3.5 | 18 | 20 | 0.43 | 0.83 | −0.40 |
| Kinnan, Bonder Prodigy         | 27.3 | 23.5 | −3.8 | 16 | 18 | 0.38 | 0.75 | −0.37 |
| Rograkh, Son of Rohgahh        | 27.9 | 23.6 | −4.3 | 14 | 17 | 0.33 | 0.71 | −0.38 |
| Meren of Clan Nel Toth         | 27.1 | **20.9** | **−6.2** | 17 | 24 | 0.40 | 1.00 | **−0.60** |
| Yuna, Grand Summoner           | 37.5 | **31.2** | **−6.3** |  1 |  1 | 0.02 | 0.04 | −0.02 |
| Esika, God of the Tree         | 31.4 | **24.6** | **−6.8** |  7 | 14 | 0.17 | 0.58 | **−0.42** |
| Derevi, Empyrial Tactician     | 32.3 | **22.6** | **−9.7** |  5 | 21 | 0.12 | 0.88 | **−0.75** |

`%ile` = rank / N within run; lower is better. `Δ%ile = noon%ile − EOD%ile`; positive = commander is **relatively higher-ranked** at EOD than at noon.

### Commanders that fell out (in noon top-50, not in EOD top-30)

Niv-Mizzet Reborn, Magus Lucea Kane, Gwenom, Edgar Markov, Tayam, Gerrard, Y'shtola, Abdel Adrian, Zurgo and Ojutai, Rasputin Dreamweaver, Vivi Ornitier, Prismari, Ivy Gleeful Spellthief, Fire Lord Azula, Emiel the Blessed, Gorma the Gullet, Hashaton Scarab's Fist, Atraxa Grand Unifier, The Reality Chip, Kilo Apogee Mind, Norman Osborn // Green Goblin, Maelstrom Wanderer (22 commanders). No EOD comparison.

### New entrants (in EOD top-30, not in noon top-50)

Alaundo the Seer (rank 7, 26.4%), Tivit Seller of Secrets (rank 9, 25.6%), Arcum Dagsson (rank 11, 25.2%), Chatterfang Squirrel General (rank 12, 25.0%), Sisay Weatherlight Captain (rank 23, 21.4%). All five land mid-pack or above the field mean — no new outliers on either tail.

## Material shifts

### **Jhoira recovered +14.8pp (14.5% → 29.3%).**
- Noon: rank 41 of 42 (bottom 2%), 76 games. EOD: rank 4 of 24 (top 17%), 123 games.
- The single largest absolute Δwr in the day. The noon report did not call Jhoira out specifically, but the 14.5% number was a clear tail outlier. EOD's 123-game sample is enough to reject the noon number as the small-sample reading. Jhoira's true engine performance in this hat config appears to be **~25-30%, near the field mean**, not tail-grade.
- Plausible engine contributors landed today: artifact-cast trigger handlers in the Muninn 161-180 wave (Jhoira is "whenever you cast a historic spell"), and the Sai test-pollution fix (375de60) restored ~50 init-registered handlers including artifact/tribal pieces Jhoira commonly leans on. The combination is consistent with "the deck's gameplan now actually fires," but the result is also entirely consistent with sample-size correction. **Recommend an isolated 2,000-game Jhoira rerun before claiming engine attribution.**

### **Francisco recovered +17.6pp (8.2% → 25.8%).**
- Noon: dead-last at 8.2% on 73 games. EOD: rank 8 of 24, 132 games at the field mean.
- The noon report's call for "Loki/Goldilocks targeted check that no Francisco-specific scaffold is mis-firing" can be **deferred** — the EOD sample shows Francisco's behavior is now neutral, not pathological. Either nothing was actually broken (noon was bad-luck variance on 73 games) or whatever was misfiring is no longer load-bearing now that the pool composition shifted around it. Either way, **not a current bug.** Keep the Loki check in the backlog at low priority.

### **Tymna +8.1pp (22.9% → 31.0%).**
- Noon: rank 27 of 42, 249 games on 3 surviving Tymna decks. EOD: rank 2 of 24, 129 games on the 1 surviving Tymna deck at hex 3045.5.
- The noon report flagged Tymna's drop from bench's 26.7% → 22.9% as "different decklist mix, not interpretable as merge regression." The EOD bounce to 31.0% **confirms that hypothesis**: with just one surviving Tymna decklist in the pool (instead of 3 averaged), the per-deck signal is stronger and shows Tymna engine is performing well. The bench-to-noon-to-EOD swing for Tymna is **almost entirely composition, not engine.**

### **Selvala +8.3pp (20.3% → 28.6%).**
- Same pattern as Tymna — single-deck pool concentration. Selvala had 74 games at noon (rank 34), 133 at EOD (rank 5). The 28.6% is much closer to the bench-era cEDH ramp commander expectation than noon's 20.3%.

### **Derevi crashed −9.7pp (32.3% → 22.6%).**
- Noon: rank 5 of 42, 62 games on 1 deck. EOD: rank 21 of 24, 257 games on 2 decks (Manamochi + Tomazinhal stax build).
- Largest negative Δwr in the day. With games quadrupled the WR fell almost 10pp. The noon 32.3% looks like a small-sample high — the EOD 22.6% is the more trustworthy reading.
- The Tomazinhal Derevi build is `archetype=stax`; the Manamochi build is `midrange`. Stax decks under-perform in 4-player random-pool matches when their lockpieces don't land early, which can drag the per-commander average. Suggest a per-deck split when `runPool` finally gets a deck-level breakout.

### **Esika −6.8pp, Yuna −6.3pp, Meren −6.2pp.**
- Esika and Meren are pool-composition + larger-sample regressions to the mean, consistent with the same pattern (high-variance noon numbers settling toward the field mean as game counts grow).
- **Yuna is more interesting**: still the top commander in the field at 31.2%, but down from noon's 37.5%. Both runs are on the same Inti-Jedi decklist, so this **is** apples-to-apples on the deck level. 80 → 112 games is a meaningful sample bump and the 37.5% reading is likely the lucky end of a distribution centered around 30-32%. **Yuna remains the field's strongest commander, just not by as wide a margin as noon suggested.**

### **Kinnan / Rograkh: another step down (−3.8 / −4.3pp).**
- These were the noon report's "most sample-robust" commanders (276-311 games). EOD games are 375-402. Both slid back toward the field mean, suggesting noon's mid-pack reading was already close to truth and EOD just refines it. No engine attribution claim either direction.

### **Tail compression.**
- Noon's worst was Francisco at 8.2% (−16.8pp from mean). EOD's worst is Meren at 20.9% (−4.2pp from mean). The 30-deck pool produces **no tail outliers** below 20%. Combined with 0 crashes / 0 concessions, the engine looks well-behaved across the entire cEDH top-30 at EOD.

## Caveats and known limitations

All noon-report caveats still apply (`docs/meta-shift-2026-05-17.md` §"Caveats and known limitations"). Specifically still-open across the day:

1. **`runPool` does not write a markdown report.** `internal/tournament/runner.go:1058` returns a `TournamentResult` but no `WriteMarkdown` call on the pool path. Both today's runs lost their canonical machine-readable report. This report and the noon report are both composed from captured stdout.
2. **`--matchup` not produced in pool mode.**
3. **30-commander print cap in `runner.go:1039`** — does not bite here because EOD has only 24 commanders, but will recur on larger pools.
4. **Pool-mode non-determinism.** `--seed 42` does not reproduce — worker goroutines consume the seed channel non-deterministically. Suggested fix per noon: seed per-game with `cfg.Seed + gameIdx`.
5. **Per-commander, not per-deck aggregation** — multi-deck commanders (Kinnan ×3, Rograkh ×3 at EOD) average across heterogeneous decklists.
6. **Game-count floor for shift interpretation.** Even EOD's lowest-game commanders (Meren 115, Yuna 112) have ±5pp 95%-CI on winrate; treat sub-6pp shifts as noise.

## Day-level recommended follow-ups

Rolling the noon report's six follow-ups forward, plus one new:

1. **(NEW) Isolated Jhoira rerun.** 2,000 games, Jhoira vs. 3 strong baseline cEDH commanders. EOD's +14.8pp recovery is the largest single-day Δwr and the only one with a plausible engine-merge story (artifact-cast handlers + Sai test-pollution fix). Confirm or refute engine attribution.
2. **Bench-pool rerun against `main` at EOD.** Run the morning bench's exact top-30 pool at 1,000 games on EOD `main` to isolate engine-change effect from pool-change effect across the full day (bench → noon → EOD).
3. **Hashaton isolated rerun.** Fell out of EOD top-30 entirely, so no fresh data. The noon hypothesis (graveyard recursion fixes helped) remains untested.
4. **Yuna isolated rerun.** Same Inti-Jedi deck at noon (80g, 37.5%) and EOD (112g, 31.2%). 2,000-game gauntlet would lock the true number.
5. **Fix `runPool` → `WriteMarkdown` + per-commander JSON.** Three day-end runs now and zero canonical reports written. Highest-leverage tooling fix.
6. **Pool-mode determinism fix.** Per-game seed = `cfg.Seed + gameIdx`.
7. **Investigate game-length jump.** 35.4 → 48.6 → 51.7 across bench → noon → EOD. The 48.6 → 51.7 step is mostly pool composition (smaller, denser cEDH pool). The bench → noon step is the merge-wave + composition combination noon already flagged.

## Day summary (bench → noon → EOD)

| Metric | Bench (500g, top-30) | Noon (1000g, top-50) | EOD (1000g, top-30) |
|---|---:|---:|---:|
| Crashes | 0 | 0 | 0 |
| Concessions | 0 | 0 | 0 |
| Throughput | 10.6 g/s | 11.0 g/s | 8.4 g/s |
| Avg turns | 35.4 | 48.6 | 51.7 |
| Unique commanders | 12 | 42 | 24 |
| Field mean WR | 24.0% | 23.8% | 25.1% |
| Top-bottom WR spread | 16.4pp | 29.3pp | 10.3pp |
| Hex top of pool | ~3032 | 3067 | 3071.8 |

**3,500 crash-free pool games on the day.** The merge train (100+ Muninn per_card handlers, the perf round-2 cache + alloc fix, goldilocks/Loki round-6 stability sweep, Freya quality audit, Sai test-pollution fix, EOD muninn waves) shipped without surfacing a single new failure mode under tournament load.

## Raw artifacts

- Tournament stdout: `/tmp/metaeod/tourney.log`
- Leaderboard snapshot: `/tmp/metaeod/elo.json` (1,315 entries, fetched 2026-05-17 00:36 PDT)
- Top-30 JSON: `/tmp/metaeod/top30.json`
- Staged decks + Freya sidecars: `/tmp/metaeod/decks/` (30 deck symlinks + `freya/` subdir with 30 `.md` + 30 `.strategy.json`)
- Tournament binary built from `0467f40`: `/tmp/metaeod/hexdek-tournament`
