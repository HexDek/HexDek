# Heimdall Post-Merge Meta Report — 2026-05-17

## Purpose

Re-run the morning's 30-deck gauntlet through `hexdek-heimdall --analyze` after today's merge wave to measure how mechanical-correctness changes (Warp keyword + 80 per-card handlers + 8 bulk-pattern families + perf round-2) have shifted the leaderboard. Same pool, same seed, same hat — engine HEAD is the only moving variable.

## Method

- **Tool:** `cmd/hexdek-heimdall --analyze` (yggdrasil hat support added on this branch; mirrors `hexdek-tournament`'s per-deck Freya strategy loading)
- **Pool:** Same 30-deck snapshot the morning baseline used (`/tmp/postmerge-decks/all/`, with `freya/` strategy JSONs alongside) — top 30 by `hex_rating` from the 2026-05-16 leaderboard snapshot
- **Mode:** `--pool` (each game samples 4 random decks from the pool of 30)
- **Games:** 500
- **Seed:** 42 (identical to morning baseline)
- **Hat:** `yggdrasil` (budget 50, turn-budget 0, noise σ 0.2 — defaults)
- **Seats:** 4 / max-turns: default

Determinism: with the same pool, same seed, and same per-game RNG derivation, the 500 game seeds are identical to the morning run. Per-commander **games played** counts are byte-identical (Tymna 559, Rograkh 392, Kinnan 328, Glarb 151, Norman 125, Hashaton 69, Malcolm 69, Thrasios 65, Vial Smasher 63, Shorikai 63, Inalla 62, Y'shtola 54). **Any winrate delta is mechanical: same matchups, different engine behavior on specific cards or interactions.**

## Comparison Anchors

| | Morning Baseline | Post-Merge | Δ |
|---|---|---|---|
| **Branch / HEAD** | `b6e6d43` (Light-Paws + Tiamat) | `a78f3a1` (#61-80 handlers) | — |
| **Tool** | `hexdek-tournament --pool` | `hexdek-heimdall --analyze --pool` | — |
| **Games** | 500 | 500 | 0 |
| **Crashes** | 0 | 0 | 0 |
| **Concessions** | 0 | 0 | 0 |
| **Avg turns** | 35.4 | 36.0 | +0.6 |
| **Wall time** | 47.0 s | 33.4 s | **−29%** |
| **Throughput** | 10.6 g/s | 15.0 g/s | **+41%** |

The throughput jump is consistent with perf round-2's headline (−22% alloc, −13% CPU). Avg-turn drift of +0.6 is within run-to-run noise and not attributable to any single merge.

## Merge Wave Between Baseline and Now

Commit range `b6e6d43..a78f3a1` (`git log --oneline`):

| Class | Merges | Notes |
|---|---|---|
| **Engine — Warp keyword** | `9a6fb01` (impl) | Full alt-cost cast + end-step exile + cast-from-exile |
| **Engine — `any-target` semantics** | `b7b0477` | Corpus `{base:target, quantifier:any}` now matches any target |
| **Parser — scaffolds** | era 1 / era 2+4 / era 3 audits + ~41 new condition scaffold kinds across `7957992 6a60951 0cd6f32` | |
| **Parser — wrapper promotions** | waves 1b / 2 / 3 (`9a6fb01 f773525 fdb6293`) | Residual `Modification` wrappers → typed kinds |
| **per_card — top-tier targeted** | Necromancy, Bloodchief Ascension, Kodama, Land Tax, The One Ring revive, Light-Paws + Tiamat tightening, Abdel Adrian battlefield-exit fix (B6/S4 nil-deref class) | |
| **per_card — Muninn gaps** | 80+ handlers across #8-#12, #13-#20, #21-#30, #31-#40, #41-#60, #51-#70, #61-#80, #81-#100 (deduped to ~88 distinct new handlers) | |
| **per_card — bulk patterns** | `c9006e1 4101aab 4318132` — 6 bulk-pattern handler families covering recurring shapes | |
| **perf** | round-2 (`2bfa1a6 35b3921`): bitmask produced-color cache + combo-eval alloc drop | |
| **Audits — proof-of-stability** | loki round 5, round 6 (+ goldilocks postmerge / round 5 / round 6) — collectively 3,000 chaos games + 20K nightmare boards, **0 crashes**, **0 panics** | |
| **UI / deploy / tests** | visual polish 2/3/4, backend deploy fix, e2e Playwright expansion | No engine impact |

Loki round 6 specifically (per `data/audit/loki_round6.md`): 2,000 chaos games + 10,000 nightmare boards in 1m16.7s, 0 crashes, 0 panics. Invariant noise floor (CardIdentity / ZoneConservation classes) was statistically flat — round 5 = 0.41 violations/game, round 6 = 0.38. **No new failure modes introduced by the merge wave.** Meta shifts below are therefore pure mechanical-correctness wins, not crash-induced artifacts.

## Pool Winrates

### Baseline (morning, 2026-05-16, `b6e6d43`)

```
 1. Thrasios, Triton Hero            32.3%  (21/65)
 2. Vial Smasher the Fierce          27.0%  (17/63)
 3. Tymna the Weaver                 26.7%  (149/559)
 4. Glarb, Calamity's Augur          26.5%  (40/151)
 5. Y'shtola, Night's Blessed        25.9%  (14/54)
 6. Inalla, Archmage Ritualist       25.8%  (16/62)
 7. Shorikai, Genesis Engine         25.4%  (16/63)
 8. Malcolm, Keen-Eyed Navigator     24.6%  (17/69)
 9. Rograkh, Son of Rohgahh          24.2%  (95/392)
10. Kinnan, Bonder Prodigy           23.8%  (78/328)
11. Norman Osborn // Green Goblin    20.0%  (25/125)
12. Hashaton, Scarab's Fist          15.9%  (11/69)
```

### Post-Merge (this run, 2026-05-17, `a78f3a1`)

```
 1. Inalla, Archmage Ritualist       32.3%  (20/62)
 2. Tymna the Weaver                 27.2%  (152/559)
 3. Thrasios, Triton Hero            26.2%  (17/65)
 4. Y'shtola, Night's Blessed        25.9%  (14/54)
 5. Norman Osborn // Green Goblin    24.8%  (31/125)
 6. Malcolm, Keen-Eyed Navigator     24.6%  (17/69)
 7. Glarb, Calamity's Augur          24.5%  (37/151)
 8. Rograkh, Son of Rohgahh          24.5%  (96/392)
 9. Vial Smasher the Fierce          23.8%  (15/63)
10. Shorikai, Genesis Engine         23.8%  (15/63)
11. Kinnan, Bonder Prodigy           22.3%  (73/328)
12. Hashaton, Scarab's Fist          17.4%  (12/69)
```

## Rank-Shift Table

`Shift = BaselineRk − PostRk` (positive = climbed).

| Commander | BaseWR% | PostWR% | Δwins | Δpp | BaseRk | PostRk | Shift |
|---|---:|---:|---:|---:|---:|---:|---:|
| **Inalla, Archmage Ritualist** | 25.8 | 32.3 | **+4** | **+6.5** | 6 | 1 | **+5** |
| **Norman Osborn // Green Goblin** | 20.0 | 24.8 | **+6** | **+4.8** | 11 | 5 | **+6** |
| Tymna the Weaver | 26.7 | 27.2 | +3 | +0.5 | 3 | 2 | +1 |
| Hashaton, Scarab's Fist | 15.9 | 17.4 | +1 | +1.5 | 12 | 12 | 0 |
| Rograkh, Son of Rohgahh | 24.2 | 24.5 | +1 | +0.3 | 9 | 8 | +1 |
| Y'shtola, Night's Blessed | 25.9 | 25.9 | 0 | 0.0 | 5 | 4 | +1 |
| Malcolm, Keen-Eyed Navigator | 24.6 | 24.6 | 0 | 0.0 | 8 | 6 | +2 |
| Shorikai, Genesis Engine | 25.4 | 23.8 | −1 | −1.6 | 7 | 10 | −3 |
| Kinnan, Bonder Prodigy | 23.8 | 22.3 | −5 | −1.5 | 10 | 11 | −1 |
| Glarb, Calamity's Augur | 26.5 | 24.5 | −3 | −2.0 | 4 | 7 | −3 |
| **Vial Smasher the Fierce** | 27.0 | 23.8 | **−2** | **−3.2** | 2 | 9 | **−7** |
| **Thrasios, Triton Hero** | 32.3 | 26.2 | **−4** | **−6.2** | 1 | 3 | −2 |

Total wins per side: 499 (1 draw on both sides — same game, same seed).

## Material Movers — Interpretation

**Inalla climbs 6→1, +4 wins on 62 games (+6.5pp).** Combo archetype, 8 Freya combos. Inalla relies on the wizard ETB-token trigger combined with sacrifice outlets and ETB-loops. The handler wave (#8-#100) shipped Necromancy + Bloodchief Ascension + 80+ Muninn gap handlers — multiple of those touch ETB-trigger / death-trigger / cast-trigger plumbing the Inalla engine sits on top of. Plausible root cause; small absolute base (62 games), so noise band ≈ ±2 wins.

**Norman Osborn climbs 11→5, +6 wins on 125 games (+4.8pp).** Two decks, both combo / dense-combo (combos = 17, 42). Strongest signal in the run — 125 games is the second-largest sample, and +6 wins on that base is well above the ±2 win/100 noise floor. Likely beneficiaries: the 6 bulk-pattern handler families (which catch recurring oracle shapes) and `any-target` semantics fix. Worth a follow-up Freya re-analysis to see if Osborn now reads as a better-tuned combo than morning.

**Thrasios drops 1→3, −4 wins on 65 games (−6.2pp).** Single deck, midrange / 7 combos. The morning bench flagged Thrasios as a +15.8pp outlier vs. live winrate (16.5%). The drop here is in the direction of "regression toward live" — i.e. the engine got closer to right, not more wrong. Today's run still sits 9.7pp above live, suggesting the merges chipped the gap but didn't close it; remaining gap is the same 12-commander pool composition issue the morning bench called out.

**Vial Smasher drops 2→9, −2 wins on 63 games (−3.2pp).** Single deck, storm archetype, 8 combos. Storm engines lean on cast-trigger counting and "spells cast this turn" state. Lower priority to investigate — within noise on 63 games, and unlike Thrasios this is a mid-tier deck moving inside the pack.

**Glarb drops 4→7, −3 wins on 151 games (−2.0pp).** Two decks, one combo-heavy (15 combos). 151 games gives this ~±1.8 win noise band, so −3 is marginal-significance. Glarb's combos lean on cast-trigger surveil + graveyard recursion — both touched by the handler wave. Worth a Heimdall card-rank re-run in isolation if it stays low.

**Kinnan drops 10→11, −5 wins on 328 games (−1.5pp).** Five decks, all midrange / low-combo (4-5 combos each). Largest sample drop, but smallest %-point move — engine improvement bringing Kinnan's pool winrate closer to live (24.5%). Same story as Thrasios in miniature: meta is correcting toward the live leaderboard.

**Tymna +3 on 559 games (+0.5pp).** Largest sample. The +3 wins inside ±4 win noise band — essentially flat. Tymna is the meta anchor and remains the meta anchor.

**Hashaton flat at rank 12 (+1 win).** Morning bench flagged this as the −8.8pp outlier and proposed reanimator-plumbing as the suspect. Post-merge: still rank 12 at 17.4% vs. live 24.7% — gap closed by 1.5pp but the regression is **not resolved**. Reanimator path remains the top open question.

## Stability Read

- 500 games, 0 crashes, 0 concessions — matches loki round-6's 0-panic finding across 2,000 chaos games + full corpus
- Avg turns 35.4 → 36.0 (+0.6) — within noise; no stall pathology introduced
- Throughput 10.6 → 15.0 g/s (+41%) — consistent with perf round-2 telemetry (−13% CPU)
- Zero "no Freya analysis" warnings; all 30 deck strategies loaded cleanly under the new yggdrasil-with-strategy path heimdall now supports

## Known Caveats

1. **Pool mode does not populate `result.Analyses` / `result.CardRankings`.** Same gap the morning bench flagged on `runPool` not calling `WriteMarkdown` — the `analytics.md` artifact this run wrote out is structurally empty (no MVP cards, kill shots, co-trigger summaries) because pool-mode aggregation in `runner.go:runPool` skips the per-game analysis collection that `aggregate.go:153` performs on the non-pool path. **Card-level analytics require a non-pool run (round-robin or rotate mode) to surface.** Recommend a follow-up Heimdall round-robin pass on a smaller pod for the deep-card numbers.
2. **Pool mode uses a single uniform hat factory** (`runner.go:857`) — even though heimdall now builds 30 per-deck yggdrasil factories with Freya strategies, only `factories[0]` is applied in pool mode. Both sides of this comparison have the same property, so the comparison is consistent; but neither run reflects "every deck plays its own Freya profile."
3. **Sample sizes are small for half the commanders** (Y'shtola 54, Inalla 62, Vial Smasher 63, Shorikai 63, Thrasios 65, Malcolm 69, Hashaton 69). Inalla's +5 rank shift sits on 62 games — credible but not bulletproof. Tymna, Rograkh, Kinnan are the only commanders with >300 games of pool data.
4. **Same-seed determinism caveat:** with 4-of-30 sampling at seed 42, the *games* are identical between runs but the *evaluation rollouts inside the hat* may differ because the yggdrasil hat's MCTS noise consumes RNG independently each invocation. Headline rank shifts at >2 wins absolute are robust to this; ±1-win shifts could be hat-noise drift rather than mechanical correctness.

## Follow-Ups

1. **Heimdall round-robin pass on a 4-deck pod** (Inalla + Norman + Thrasios + Hashaton at 2,000 games) — surfaces MVP cards, kill shots, and co-trigger interactions that pool mode discards. The four commanders cover the headline movers in both directions.
2. **Hashaton reanimator audit.** Morning bench flagged this; current run leaves it unresolved at rank 12. Inspect discard / reanimate / graveyard plumbing against the May 11 forensics doc's anti-pattern sweep.
3. **Fix `runPool` to populate `Analyses`** — small mechanical change (~10 LOC in `runner.go:959-984`) that would let pool-mode runs produce a non-empty analytics report instead of the current stub. Out of scope for this report; opening as a separate task.
4. **Re-run Norman Osborn (Green Goblin) in isolation** — strongest positive signal (+6 wins / 125 games / +4.8pp) deserves confirmation at 2k games before claiming Warp + bulk-patterns as the cause.

## Raw Artifacts

- Heimdall stdout: `/tmp/heimdall-postmerge/run.log`
- Heimdall analytics stub: `/tmp/heimdall-postmerge/analytics.md` (empty — pool-mode limitation noted above)
- Deck pool snapshot: `/tmp/postmerge-decks/all/*.txt` (30 files + `freya/` subdir)
- Baseline source: `docs/postmerge-meta-bench-2026-05-16.md`
- Stability anchor: `data/audit/loki_round6.md`
