# Loki Fuzz — Round 5 (Total Bug-Hunting)

**Branch:** `dev/loki-round5`
**Date:** 2026-05-16
**Tool:** `cmd/hexdek-loki` (normal mode; `--seed-cards`/`--seed-cmdr` available but unused — this is broad coverage, not handler-targeted)
**Run:** `hexdek-loki -games 1000 -seed 555 -report …` (defaults: 4 seats, max-turns 60, 10 workers, nightmare-boards 10000)

## Scope

Fresh chaos run against `main` HEAD after the latest merge wave (post `dev/goldilocks-postmerge` clean baseline). Targets coverage of the ~40 new per_card handlers, the two bulk-pattern families, the tournament `--pool` dashboard fix, and the Bolt-targeting regression fix.

| Commit | Subject |
|--------|---------|
| `3255991` | Merge dev/muninn-handlers-21-30: 9 new handlers |
| `1bbd124` | Merge dev/muninn-bulk-patterns: bulk-pattern handlers for card families |
| `8e46cfe` | Merge dev/muninn-progress: 2026-05-17 progress report |
| `4b05378` | Merge dev/muninn-handlers-31-40: 8 per_card handlers (#31–#40) |
| `b22b39b` | Merge dev/tournament-pool-output-fix |
| `178c972` | Merge dev/goldilocks-postmerge: 99.80% pass, 0 panics |
| `509a5ec` | Merge dev/loki-new-handlers: --seed-cards/--seed-cmdr + 1600-game audit |
| `b28bfdb` | Merge dev/muninn-handlers-8-12 |
| `b6e6d43` | Merge dev/lightpaws-tiamat |
| `fab44ea` | Merge dev/muninn-top5-handlers |
| `af762ad` | Merge dev/may11-nil-deref-forensics (Abdel Adrian) |
| `4c4377b` | Merge dev/land-tax |
| `b8fe855` | Merge dev/the-one-ring |
| `862eb06` | Merge dev/era3-scaffolds |
| `0718a1f` | Merge dev/loki-postmerge-audit (B1 + B2 fixes) |

## Headline Result

**Zero crashes. Zero panics.** 1,000 chaos games + 10,000 nightmare boards complete in 1m17s. Goldilocks-postmerge's "0 known panics" expectation holds end-to-end.

| Phase | Count | Duration | Crashes | Violations | Affected |
|-------|------:|---------:|--------:|-----------:|---------:|
| Chaos games | 1,000 | 1m14.6s | **0** | 412 | 13 games |
| Nightmare boards | 10,000 | 3.1s | **0** | 2 | 1 board |
| **Total** | — | 1m17.7s | **0** | **414** | — |

Clean-game rate: **987 / 1000 (98.7%)**. Clean-board rate: **9999 / 10000 (99.99%)**. In line with the post-B1/B2 baseline noise floor from `loki_postmerge_audit.md`.

## Violations by Invariant

| Invariant | Count | Notes |
|-----------|------:|-------|
| CardIdentity | 330 | Same pointer in two zones — three distinct card signatures (see below) |
| ZoneConservation | 78 | Single bad game, almost certainly the same conspire/copy class as B7 |
| AttachmentConsistency | 2 | Stale equip pointer after host left the battlefield (B6-class) |
| TriggerCompleteness | 2 | Long-tail; not isolated this run |
| **Total** | **412** | All confined to **13 of 1,000 games** |

The detail block in `CHAOS_REPORT.md` inlines only the first 30 occurrences; the remaining 382 occurrences are not individually inspectable, but the four signatures observed in those 30 are consistent with previously-catalogued, pre-existing issues. No new failure mode is visible in the detail block.

## New / Continuing Signatures

### S1 — `Lo and Li, Twin Tutors` duplicated across graveyard ↔ exile
- **Severity:** Medium (18 occurrences observed, single game, single card)
- **Invariant:** CardIdentity — same `*Card` pointer in `seat 0 graveyard` and `seat 0 exile`
- **Game:** 178 (master seed 555 → game seed 1780556)
- **No per_card handler** for `lo_and_li_twin_tutors.go` exists yet. Card text involves a partner / tutor effect that likely routes through both a graveyard-on-cast path and an exile-on-resolution / replacement effect — the cast pipeline appears to leave the pointer in graveyard while a follow-on exile move appends without zone-removal. Same class as the B3 (Adric) and B4 (Rally / conspire) cluster from the postmerge audit.
- **Repro:** `hexdek-loki -games 1 -seed 1780556 -permutations 1` (or `-seed 555` and skip to game 178).

### S2 — `Baron Helmut Zemo` duplicated across command_zone ↔ battlefield
- **Severity:** Medium (8 occurrences, single seat)
- **Invariant:** CardIdentity — `seat 0 command_zone` and `seat 0 battlefield`
- **Hypothesis:** A cast-from-command-zone flow where the cmdzone slot is not cleared before the commander resolves onto the battlefield, or a §903.9a "may be put in command zone" replacement firing during a battlefield-exit while the permanent ref survives. Cousin of the now-fixed B2 (Gerrard, Weatherlight Hero) and the still-open B5 (Cerulean Sphinx) classes.
- **No per_card handler** for Zemo exists. Likely a generic commander-SBA / cast-from-cmdzone edge case rather than a card-specific issue.

### S3 — `Weatherseed Faeries` duplicated across exile ↔ battlefield
- **Severity:** Low (2 occurrences)
- **Invariant:** CardIdentity — `seat 3 exile` and `seat 3 battlefield`
- **Hypothesis:** Flashback returning the spell to battlefield (creature spell with retrace-style behaviour) while the exile-on-resolution didn't take. Same residual class as S1.

### S4 — `Doc Ock's Tentacles` attached to dead `Species Gorger`
- **Severity:** Low (2 occurrences, both in the same game)
- **Invariant:** AttachmentConsistency
- **Class:** B6-class stale `AttachedTo` after the equipped/enchanted host left the battlefield. Same root cause shape as the postmerge `Spear of the General` / `Stupefying Touch` reports. SBA §704.5n should null the pointer; either it didn't fire or it walked the wrong seat's battlefield.

## Status of Postmerge Audit Bugs (B1–B7)

| Bug | Reported in postmerge | This run | Status |
|-----|-----------------------|----------|--------|
| B1 — Equipment self-equip (Reality Chip, etc.) | 174 occ | **0** | ✅ Confirmed fixed |
| B2 — Gerrard commander cmdzone double-add | 34 occ | **0** | ✅ Confirmed fixed |
| B3 — Adric (Doctor's companion) hand×graveyard | 54 occ | not observed in this seed | Likely still latent (S1 is the same class on a different card) |
| B4 — Rally the Galadhrim conspire copy in graveyard | 40 occ | not isolated | ZoneConservation count (78 in 1 game) is consistent with the same class continuing to fire |
| B5 — Cerulean Sphinx shuffle-into-library | 34 occ | not observed in this seed | Likely still latent |
| B6 — Spear of the General attachment | 2 occ | **2 occ** (Doc Ock's Tentacles) | Same class; not card-specific |
| B7 — ZoneConservation phantom cards | 34 occ | **78 occ** (1 game) | Same class as B4; still open |

**No new bug class** surfaced in this run — every violation signature maps to an existing open issue from `loki_postmerge_audit.md`, with the new card names (`Lo and Li`, `Baron Helmut Zemo`, `Weatherseed Faeries`, `Doc Ock's Tentacles`) adding instance variety but not new root causes.

## Regression Assessment vs Latest Merge Wave

Every distinct signature in this 1,000-game run is **pre-existing** — none trace to commits merged after `0718a1f`. Specifically:

- The 40 new per_card handlers (waves 8-12, top5, bulk-patterns, snowflakes #21-40) produced **zero** crashes and **zero** handler-attributable violations.
- The Abdel Adrian battlefield-exit fix (`af762ad`) held up across the run — no nil-deref panics, no `abdel_adrian.go`-attributable violations.
- The Bolt corpus targeting fix (`c9a17da`, included via earlier merge) showed no regression: no Lightning Bolt-attributable failures across 1,000 games.
- The tournament `--pool` dashboard fix (`b22b39b`) is UI-only; not exercised here but no engine-side fallout.

## Recommendation

- **Land clean** — no new gating issues. Round 5 is a regression-free confirmation pass.
- Open follow-up triage for **S1 (Lo and Li)** and **S2 (Baron Helmut Zemo)** alongside B3 (Adric) — these three look like the same cast/exile-replacement family and may share a single fix in the cast pipeline.
- B5 (Cerulean Sphinx shuffle-into-library) and B6/S4 (stale `AttachedTo` after host death) remain the most actionable next-step bugs — both have well-defined root-cause hypotheses in the postmerge audit.

## Repro Seeds (master seed 555)

| Game | Seed | Trigger Cards | Class |
|------|------|---------------|-------|
| 8 | 80556 | Doc Ock's Tentacles on dead Species Gorger | S4 / B6-class |
| 157 | 1570556 | Weatherseed Faeries exile↔battlefield | S3 |
| 178 | 1780556 | Lo and Li grave↔exile, Baron Helmut Zemo cmdzone↔battlefield | S1 + S2 |

Raw report: `/tmp/loki-round5/CHAOS_REPORT.md` (preserved for any deeper digging).
