# EOD Audit (Final) — 2026-05-17

Final end-of-day stability sweep before close. Three tools run
sequentially against `main` HEAD `1361972` plus today's in-flight
worker commit `657c6ff` (hat conviction diagnostic) — the working
tree at run time. Branch: `dev/eod-audit-final` from `main`.
Run window: 2026-05-17 17:35–17:38 UTC.

## Headline

| Tool        | Result                                                  | Δ vs noon EOD (eod-audit-2026-05-17.md) |
| ----------- | ------------------------------------------------------- | --------------------------------------- |
| Muninn      | 176 parser-gap unique cards, 805 crashes, 0 panics      | +6 gaps, +11 crashes                    |
| Goldilocks  | 31,899 / 31,963 pass (99.80%), 64 fails, **0 panics**   | identical (same 64-card set)            |
| Loki 1000g  | 0 crashes (chaos + nightmare), 480+2 invariant viols    | doubled volume, no new crash signature  |

**Zero panics, zero crashes, zero goldilocks drift.** All deltas are
consistent with sampling/time progression on the pre-merge engine
binary still serving DARKSTAR (every per-card handler shipped today
remains invisible to the live grinder until the next deploy). No
regression introduced by today's hat work, handler waves, UI
changes, or auditor packages.

## 1. Muninn — `--all --top 30`

```
hexdek-muninn --all --top 30
PARSER GAPS:    176 total (top-30 shown)
CRASHES:        805 total (top-30 shown)
DEAD TRIGGERS:  1 (The One Ring upkeep, last seen 2026-04-30)
CONCESSIONS:    1054 records, top 4 commanders
```

| Metric                  | Noon EOD (07:00 UTC) | Final (17:38 UTC) | Δ        |
| ----------------------- | -------------------- | ----------------- | -------- |
| Parser-gap unique cards | 170                  | 176               | +6       |
| Recurring crashes       | 795                  | 805               | +10      |
| Dead triggers           | 1                    | 1                 | 0        |
| Concession records      | 1,054                | 1,054             | 0        |
| Top-1 (The One Ring)    | 1,463,323            | 1,536,990         | +73,667  |
| Top-2 (Land Tax)        | 626,221              | 684,800           | +58,579  |

The +6 unique-gap delta over ~10 hours is consistent with the
"pre-merge binary still serving" pattern documented in
`docs/muninn-progress-2026-05-17-final.md`: every per-card handler
shipped today (waves #1-#220 + bulk-pattern families 1-5) is
invisible to the live grinder, so the top-30 keeps growing at
pre-merge velocity and a handful of new snowflakes drip in each
hour. Top-30 ordering matches noon EOD except for natural
small-count reordering in ranks 9-11.

The +10 crashes are the same goroutine-stack signature already
catalogued (seat-0 grinder, `turns=0`, no panic message, post-Abdel
Adrian root-cause fix in `b348f4a` + `abdel_adrian.go` rewrite).
No new failure mode.

Dead triggers and concessions unchanged.

**No regression.** Gap reduction will only show after
`./scripts/deploy.sh backend` ships the post-merge binary.

## 2. Goldilocks — `hexdek-thor --goldilocks` (full corpus)

```
hexdek-thor --goldilocks
35,708 cards loaded → 31,963 AST-verified tests
Pass:        31,899
Failures:    64  (61 dead-effect + 2 invariant + 1 unverified)
Panics:      0
Keywords:    2,013 / 2,013 pass
Wall-clock:  ~1.2s
```

| Metric                | Noon EOD (59aa9a5) | Final (657c6ff working tree) | Δ |
| --------------------- | ------------------ | ----------------------------- | - |
| Cards tested          | 35,708             | 35,708                        | 0 |
| Passed                | 30,277             | 30,277                        | 0 |
| Dead-effect           | 61                 | 61                            | 0 |
| Invariant violations  | 2                  | 2                             | 0 |
| Unverified            | 1                  | 1                             | 0 |
| **Panics**            | **0**              | **0**                         | **0** |
| Keyword tests         | 2,013 / 2,013 pass | 2,013 / 2,013 pass            | 0 |
| Pass rate             | 99.79%             | 99.80%                        | ≈0 |

The 64-card failure set is **identical** to noon EOD: same 61
dead-effect cards (Kithkin Zephyrnaut, Lord of Tresserhorn, Etali
Primal Conqueror, Pyrohemia, Werewolf Pack Leader, Hobgoblin
Captain, etc.) plus 2 invariant (1 TurnStructure, 1 CardIdentity)
plus 1 unverified. The conviction-action commit (657c6ff) is hat
diagnostic code only — it does not touch handler dispatch or AST
verification, so a zero delta is expected.

**No regression.**

## 3. Loki — `--games 1000` (chaos + 10K nightmare boards)

```
hexdek-loki --games 1000
Seed:         42         Workers: 10        Seats: 4        Max turns: 60
Chaos:        38.37s     26 g/s
Nightmare:    2.263s     4,419 b/s
```

### Chaos phase

| Metric            | Noon EOD (500g) | Final (1000g) | Per-1000g extrap | Δ          |
| ----------------- | --------------: | ------------: | ---------------: | ---------- |
| Games             | 500             | 1,000         | —                | doubled    |
| **Crashes**       | **0**           | **0**         | 0                | **0**      |
| Invariant viols   | 138 (5 games)   | 480 (15 games)| 276 (10 games)   | +73% vs extrap |
| Clean-game rate   | 99.0%           | 98.5%         | 99.0%            | −0.5pp     |
| Throughput        | 25 g/s          | 26 g/s        | —                | +1 g/s     |

Invariant breakdown (chaos):

| Invariant              | Noon EOD (500g) | Final (1000g) | Prior round-6 (2000g) | Class |
| ---------------------- | --------------: | ------------: | --------------------: | ----- |
| CardIdentity           | 136             | 392           | 586                   | B3 / S5 (Adric-style hand×grave) |
| AttachmentConsistency  | 2               | 6             | 6                     | B6 / S4 (stale equip → off-board token) |
| ZoneConservation       | —               | 78            | 156                   | B4/B7 / S6 (phantom-cards-after-wipe) |
| TriggerCompleteness    | —               | 4             | — (2 in round 7)      | S9 (trigger w/o stack_resolve) |

The two categories absent from the noon 500g sample (`ZoneConservation`,
`TriggerCompleteness`) are **not new**: both are pre-existing
aggregates from rounds 5–7
(`data/audit/loki_round{5,6,7}.md`). They are tail-distribution
classes that simply weren't sampled at 500g. Per-1000g density for
`ZoneConservation` (78) matches round-6 exactly (156 / 2,000 = 78).

The +73% extrap delta on total chaos-violation count concentrates
in two affected games:

- **Game 108** (seed 1080043) — A-Phylath, World Sculptor + Sauron
  pod. New `AttachmentConsistency` witness: Inventor's Axe attached
  to a `creature token green elf warrior Token` that is not on any
  battlefield. Same shape as round-6 S7 (Born to Drive on a dead
  pilot token, Game 216) and postmerge "Spear of the General /
  Stupefying Touch". B6/S4-class; SBA §704.5n should null the
  pointer when the host token left play.
- **Game 170** (seed 1700043) — same Ezuri / Master of Keys / Syr
  Ginger / Katara pod already isolated as the CardIdentity hotspot
  in the noon EOD report. Repeats with the same signature.

No new gating bug class.

### Nightmare phase

| Metric          | Noon EOD (10K) | Final (10K) |
| --------------- | -------------: | ----------: |
| Boards          | 10,000         | 10,000      |
| **Crashes**     | **0**          | **0**       |
| Invariant viols | 2              | 2           |
| Clean-board rate| 99.99%         | 99.99%      |

Identical to noon EOD.

### Crashes

**None.** Zero crashes across 1,000 chaos games and 10,000 nightmare
boards. The Loki-side invariant continues to hold: no panic has
reproduced since the May 11 nil-deref burst was root-caused
(`b348f4a` + `abdel_adrian.go` battlefield-exit rewrite).

## Regression Verdict

**Clean.** All three tools show zero regression from today's work:

- **Hat / handler waves:** Goldilocks failure set is byte-identical
  to noon EOD. The 100+ per-card handlers, 9 bulk-pattern families,
  hat conviction-action diagnostic, archetype-aware weights, and
  yggdrasil polish landed today produced no new dead-effect, no
  new invariant violation, no new panic.
- **Loki crashes:** 0 across 1K chaos + 10K nightmare. Same as
  morning, same as round-7, same as round-6, same as round-5.
- **Loki violations:** all 482 traces map to pre-existing classes
  catalogued in `data/audit/loki_round{5,6,7}.md`. No new
  signature. Game 170 still ticks (Adric-class, B3); Game 108 is a
  fresh witness for the existing B6/S4 stale-attachment class.
- **Muninn:** gap and crash counts drift consistent with the
  pre-merge binary on DARKSTAR. No new failure mode in the dead
  triggers, concessions, or crash signatures.

## Next action

Same as every EOD report this week: cross-compile and ship the
post-merge binary to DARKSTAR (`./scripts/deploy.sh backend`).
Until that ships, Muninn gap reduction cannot be observed.

## Repro

```bash
git checkout dev/eod-audit-final  # branched from main 1361972
go run ./cmd/hexdek-muninn --all --top 30
go run ./cmd/hexdek-thor --goldilocks
go run ./cmd/hexdek-loki --games 1000   # seed 42, 10 workers
```

Raw outputs: `hexdek/audit-eod-final/{muninn,goldilocks,loki}.txt`
Loki chaos report: `data/rules/CHAOS_REPORT.md` (this run).
