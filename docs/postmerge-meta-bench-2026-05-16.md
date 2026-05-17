# Post-Merge Meta Benchmark — 2026-05-16

## Purpose

Quality benchmark to measure whether today's merge wave (Warp, scaffolds, per_card handlers, perf cache) measurably changed the metagame versus the live leaderboard.

## Method

- **Tool:** `cmd/hexdek-tournament` (Hat = `yggdrasil`, defaults: budget 50, σ 0.2, commander on)
- **Pool:** Top 30 decks by `hex_rating` from `https://hexdek.dev/api/live/elo`
- **Decks fetched from DARKSTAR** (`~/hexdek/data/decks/moxfield{,_300}/`), Freya strategy JSONs staged alongside (all 30 loaded successfully — no "no Freya analysis" warnings)
- **Games:** 500 in `--pool` mode, seats=4 (every game samples 4 random decks from the pool)
- **Flags:** `--matchup` requested (see Caveat below), `--report /tmp/report.md` requested
- **Stability:** 0 crashes, 0 concessions, avg 35.4 turns, 47.0 s wall, 10.6 g/s

Branch: `dev/postmerge-meta-bench` from `origin/main` (head `b6e6d43` — Light-Paws + Tiamat merge).

## Caveats — these matter for interpretation

1. **Pool mode aggregates by commander, not by deck.** The 30 distinct decklists collapsed to 12 unique commanders (Tymna ×8, Rograkh ×6, Kinnan ×5, Glarb/Norman ×2, the rest ×1). Per-deck winrates were not produced; per-commander winrates pool all decks sharing that commander.
2. **`runPool` does not call `WriteMarkdown`.** `internal/tournament/runner.go:101` dispatches to `runPool`, which never invokes `result.WriteMarkdown(cfg.ReportPath)` — only `runner.go:232` (non-pool) and `roundrobin.go:194` do. The CLI then prints `"Report written to /tmp/report.md"` unconditionally (`cmd/hexdek-tournament/main.go:344`), which is a false success message. **No `report.md` was actually written.** This report is hand-composed from stdout.
3. **`--matchup` produced no matrix in pool mode.** The post-pool block prints the legacy non-pool dashboard with zeroed fields; the matchup matrix lives in that legacy printer and was not populated. Same surface as the report-write bug.
4. **Sample size per commander is small** (Y'shtola, Inalla, Shorikai, Hashaton each ≈55–70 games). Treat their shifts as noisy; Tymna/Rograkh/Kinnan are the only commanders with >300 games of pool data.
5. **Production winrate is across all live games for that commander**, not against this exact 30-deck pool. Comparison is directional, not apples-to-apples.

## Top-30 Pool (sorted by `hex_rating` desc)

| Rk | Hex | LiveWR% | Games | Commander | deck_id (short) |
|---:|----:|--------:|------:|---|---|
|  1 | 3032.4 | 24.5 | 127207 | Kinnan, Bonder Prodigy | firedog1947 |
|  2 | 3032.2 | 24.5 | 125870 | Tymna the Weaver | ultrainstinctsol |
|  3 | 3026.9 | 24.5 | 126616 | Kinnan, Bonder Prodigy | apophismtg |
|  4 | 3022.4 | 24.6 | 126357 | Rograkh, Son of Rohgahh | devastationhour0202 |
|  5 | 3018.1 | 24.6 | 125062 | Tymna the Weaver | thigglesworth |
|  6 | 3016.2 | 24.6 | 126109 | Tymna the Weaver | kobeta |
|  7 | 3010.9 | 16.5 | 109893 | Thrasios, Triton Hero | registerformoxfield |
|  8 | 3006.6 | 24.6 | 124054 | Tymna the Weaver | devastationhour0202 |
|  9 | 3005.8 | 24.5 | 127480 | Inalla, Archmage Ritualist | yakuza |
| 10 | 3003.7 | 24.7 | 121332 | Hashaton, Scarab's Fist | drgwender |
| 11 | 3001.3 | 24.6 | 124106 | Y'shtola, Night's Blessed | lappland312 |
| 12 | 3000.1 | 24.6 | 125974 | Tymna the Weaver | zampmtg |
| 13 | 2999.8 | 24.5 | 126129 | Tymna the Weaver | bengordon6 |
| 14 | 2999.2 | 24.4 | 128043 | Tymna the Weaver | quincyhicks |
| 15 | 2998.8 | 24.5 | 126711 | Malcolm, Keen-Eyed Navigator | hdshadow |
| 16 | 2998.1 | 22.6 | 126984 | Tymna the Weaver | ezinho |
| 17 | 2995.5 | 24.6 | 125602 | Kinnan, Bonder Prodigy | xbonds16 |
| 18 | 2993.2 | 24.5 | 126524 | Kinnan, Bonder Prodigy | breadmaker207 |
| 19 | 2987.4 | 24.2 | 129748 | Rograkh, Son of Rohgahh | ocellblau |
| 20 | 2986.7 | 20.0 | 122912 | Glarb, Calamity's Augur | nomorefrenchtoast |
| 21 | 2984.4 | 24.4 | 128536 | Norman Osborn // Green Goblin | st33z |
| 22 | 2983.8 | 24.5 | 127569 | Rograkh, Son of Rohgahh | potatok1ng |
| 23 | 2982.8 | 24.5 | 126321 | Rograkh, Son of Rohgahh | kuyass |
| 24 | 2980.6 | 24.5 | 127769 | Rograkh, Son of Rohgahh | sparten407 |
| 25 | 2980.3 | 24.6 | 124144 | Shorikai, Genesis Engine | samuraichou |
| 26 | 2979.6 | 21.7 | 125240 | Vial Smasher the Fierce | thespeeze |
| 27 | 2979.3 | 24.6 | 125734 | Kinnan, Bonder Prodigy | tuberculosis |
| 28 | 2978.5 | 24.7 | 122669 | Rograkh, Son of Rohgahh | devynapse |
| 29 | 2977.7 | 22.3 | 126700 | Glarb, Calamity's Augur | kinkajou65 |
| 30 | 2974.5 | 24.6 | 125649 | Norman Osborn // Green Goblin | greengobbles |

Freya archetypes observed at deck load (mix across all 30): combo ×7, storm ×7, midrange ×11, stax ×4, artifacts ×1.

## Bench winrates (12 unique commanders)

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

Coverage: 12/12 commanders appeared in ≥1 game.

## Rank-Shift vs Production (per commander)

Production WR = pool-games-weighted average of `win_rate` across the commander's representatives in this top-30 list. Production rank = order by `hex_rating`-weighted average within this list (1 = highest).

| Commander | BenchWR% | ProdWR% | Δpp | BenchRk | ProdRk | Shift | Decks | BenchGames |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Thrasios, Triton Hero          | 32.3 | 16.5 | **+15.8** |  1 |  1 |   0 | 1 |  65 |
| Vial Smasher the Fierce        | 27.0 | 21.7 |  +5.3 |  2 | 11 |  +9 | 1 |  63 |
| Tymna the Weaver               | 26.7 | 24.3 |  +2.4 |  3 |  2 |  -1 | 8 | 559 |
| Glarb, Calamity's Augur        | 26.5 | 21.2 |  +5.3 |  4 |  9 |  +5 | 2 | 151 |
| Y'shtola, Night's Blessed      | 25.9 | 24.6 |  +1.3 |  5 |  6 |  +1 | 1 |  54 |
| Inalla, Archmage Ritualist     | 25.8 | 24.5 |  +1.3 |  6 |  3 |  -3 | 1 |  62 |
| Shorikai, Genesis Engine       | 25.4 | 24.6 |  +0.8 |  7 | 10 |  +3 | 1 |  63 |
| Malcolm, Keen-Eyed Navigator   | 24.6 | 24.5 |  +0.1 |  8 |  7 |  -1 | 1 |  69 |
| Rograkh, Son of Rohgahh        | 24.2 | 24.5 |  -0.3 |  9 |  8 |  -1 | 6 | 392 |
| Kinnan, Bonder Prodigy         | 23.8 | 24.5 |  -0.7 | 10 |  4 |  -6 | 5 | 328 |
| Norman Osborn // Green Goblin  | 20.0 | 24.5 |  -4.5 | 11 | 12 |  +1 | 2 | 125 |
| Hashaton, Scarab's Fist        | 15.9 | 24.7 | **-8.8** | 12 |  5 |  -7 | 1 |  69 |

`Shift = ProdRk - BenchRk`; `+` = bench ranks it **higher** than production does; `-` = bench ranks it **lower**.

## Material shifts (|Δwr| > 5pp OR |shift| ≥ 3)

- **Thrasios, Triton Hero — +15.8pp** (16.5 → 32.3). Lone deck in pool. Live deck is an outlier underperformer (16.5%) — bench treats it as the strongest commander in the field. Either the merges materially improved Thrasios lines (Freya combos=7, midrange archetype loaded cleanly) or the production deck is dragged down by matchups not represented in this 12-commander pool. Single deck, 65 games — wide CI.
- **Vial Smasher — +5.3pp, rank 11→2** (`+9`). One deck, 63 games. Storm archetype with 8 combos / 9 tutor targets per Freya. Possible Warp / scaffold beneficiary; worth a 2k-game re-run in isolation before claiming a real shift.
- **Glarb — +5.3pp, rank 9→4** (`+5`). 2 decks, 151 games. Combo archetype (one entry has combos=15). Same caveat — re-run for confirmation.
- **Inalla — rank 3→6** (`-3`). WR delta small (+1.3pp), but production-ranked higher than bench. Single deck, 62 games — likely noise.
- **Shorikai — rank 10→7** (`+3`). WR delta near zero (+0.8pp). Small-sample rank shuffle.
- **Kinnan — rank 4→10** (`-6`). The headline name on the live board (5 decks, 328 games here) plays at field-average winrate in the pool. Not a regression in absolute WR (-0.7pp), but it loses its leaderboard premium because every other commander is bunched near 24-26%. Suggests the live hex_rating premium for Kinnan reflects matchup spread the 12-commander pool can't reproduce, not a merge-induced regression.
- **Hashaton — -8.8pp, rank 5→12** (`-7`). The biggest absolute drop. 1 deck, 69 games — small sample. Loaded as `archetype=midrange, combos=12, tutor_targets=10`. **Flag for follow-up:** Hashaton is reanimator-adjacent; today's merges include `dev/muninn-top5-handlers` (Necromancy, Bloodchief, Kodama) and `dev/may11-nil-deref-forensics` (Abdel Adrian battlefield-exit). If any of those touched the discard / reanimate / graveyard plumbing Hashaton leans on, a regression is plausible.
- **Norman Osborn — -4.5pp**. 2 decks, 125 games. Below the 5pp threshold but worth noting; recently-printed commander, deck quality variance high.

## Field-stability read

- Engine ran 500 games clean: **0 crashes, 0 concessions, no Hat fatal errors, no Goldilocks-style invariant prints**. No new failure modes attributable to the merge wave surfaced at this scale.
- Average game length 35.4 turns — consistent with prior pool runs (no obvious stall pathology from the perf cache or scaffold work).
- Per-commander WR clusters tightly between 23.8% and 27.0% for the 10 commanders with ≥60 games (Thrasios high outlier above, Hashaton low outlier below). For 4-seat random pool the no-skill expectation is 25%, so the field is well-balanced; the AI is not exploiting any single archetype.

## Recommended follow-ups (not done in this run)

1. **Fix `runPool` to call `WriteMarkdown`** and to populate the matchup matrix. Today both silently no-op while the CLI prints success.
2. **Re-run Hashaton and Thrasios in isolation** (e.g., 4-deck round-robin at 2k games) before drawing merge-causation conclusions on those two; bench sample is 54–69 games.
3. **Switch to `--round-robin`** for benchmarks where per-deck (not per-commander) shifts matter, or run **per-deck pool with deck-keyed aggregation**, since pool mode's current collapse-by-commander loses the signal the leaderboard sorts on.

## Raw artifacts

- Tournament stdout/log: `/tmp/tournament.log`
- Leaderboard snapshot: `/tmp/elo.json` (1313 entries, taken pre-run)
- Top-30 JSON: `/tmp/top30.json`
- Rank-shift table source: `/tmp/rank_shift.json`
- Deck files used: `/tmp/postmerge-decks/all/*.txt` (30 files + freya/ subdir)
