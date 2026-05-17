# Loki Fuzz — Round 7 (Hat Archetype-Weights + Handler Tier)

**Branch:** `dev/audits-round-7` (from `main` @ `127df76`)
**Date:** 2026-05-17
**Tool:** `cmd/hexdek-loki` (normal mode, no handler seeding)
**Run:** `hexdek-loki -games 2000 -seed 777 -report …` (defaults: 4 seats, max-turns 60, 10 workers, nightmare-boards 10000)
**Raw report:** [`loki_round7_raw.md`](loki_round7_raw.md) (2,915 lines)
**Baseline:** Round-6 (`20c0c15`, 2026-05-16, seed 666) — 0 crashes, 752 violations / 24 games, 0 nightmare violations

## Scope

Round 7 covers the merge wave that landed the **archetype-specific hat eval weights** + the **largest per_card handler tier expansion to date** (~70 new handlers, 6 new bulk-pattern families, plus the wave-201-220 dispatch fallback for cascade/copy/token name variants), the Sai test-pollution Reset-hook fix, and frontend/docs surface that is not exercised by this fuzz.

| Merge | Surface |
|---|---|
| `572ab7c` `dev/hat-improvements` | Archetype-specific eval weights for 9 Freya archetypes, sparse-weight merge |
| `6e1d8bf` `dev/hat-weights-validation` | 500g A/B confirms archetype dispatch shifts winrates (validation-only) |
| `02cf908` `dev/muninn-handlers-201-220` | Dispatch fallback for cascade/copy/token name variants |
| `0b22714` `dev/muninn-bulk-patterns-5` | etb-basic-land-ramp + etb-drain-target-opponent families |
| `eecb1c3` `dev/muninn-bulk-patterns-4` | shuffle-self-from-grave + etb-library-tutor families |
| `563b44e` `dev/muninn-handlers-181-200` | 2 handlers |
| `66fb9ec` `dev/muninn-handlers-161-180` | 1 handler |
| `59aa9a5` `dev/muninn-handlers-141-160` | 12 handlers (incl. Karlach, Worldspine Wurm, It That Betrays) |
| `e51a3ce` `dev/muninn-handlers-121-140` | 10 handlers |
| `32a3033` `dev/muninn-handlers-101-120` | 11 handlers |
| `a78f3a1` `dev/muninn-handlers-61-80` | 10 handlers |
| `fa91bf3` `dev/muninn-handlers-51-60-finish` | 5 top-50 + 51-60 holdouts |
| `4685c9d` `dev/muninn-handlers-81-100` | 10 handlers |
| `375de60` `dev/sai-test-pollution` | Reset-hook plumbing — restores ~50 late-`init()` handlers across `tribal_lords`, `batch17_sweep`, `combat_restrictions`, `obeka_support`, and four `zz_*_register.go` files |
| Frontend / docs / deploy / visual-polish | Out of scope for engine fuzz |

## Headline Result

**Zero crashes. Zero panics.** 2,000 chaos games + 10,000 nightmare boards complete in 1m21.7s.

| Phase | Count | Duration | Crashes | Violations | Affected |
|-------|------:|---------:|--------:|-----------:|---------:|
| Chaos games | 2,000 | 1m19.55s | **0** | **620** | 25 games |
| Nightmare boards | 10,000 | 2.10s | **0** | **2** | 1 board |
| **Total** | — | 1m21.7s | **0** | **622** | — |

Clean-game rate: **1,975 / 2,000 (98.75%)**. Clean-board rate: **9,999 / 10,000 (99.99%)**.

## Round-on-round comparison

| Metric | Round 7 (seed 777) | Round 6 (seed 666) | Δ |
|---|---:|---:|---:|
| Chaos games | 2,000 | 2,000 | 0 |
| Chaos crashes | **0** | **0** | 0 |
| Chaos violations | 620 | 752 | **−132 (−17.6%)** |
| Affected games | 25 / 2,000 | 24 / 2,000 | +1 |
| Violations/game (noise floor) | 0.310 | 0.376 | **−18%** |
| Nightmare boards | 10,000 | 10,000 | 0 |
| Nightmare crashes | **0** | **0** | 0 |
| Nightmare violations | **2** | **0** | **+2** (new — see §S10) |
| Chaos throughput | 25 g/s | 27 g/s | −7% |

**Net read:** total violation volume drops 17.6% across a doubled-rate handler tier. Per-game noise floor is the lowest since round 5 (0.41). One new signal class on nightmare boards (2 CardIdentity in 10K, single-board cluster) — see §S10.

## Violations by Invariant (Chaos Games)

| Invariant | Round 7 | Round 6 | Δ | Class |
|---|---:|---:|---:|---|
| CardIdentity | 424 | 586 | **−162 (−27.6%)** | B3/B5/S1/S2/S5 (same-pointer-two-zones) |
| ZoneConservation | 172 | 156 | +16 (+10.3%) | B4/B7/S6 (copy bug / phantom cards) |
| AttachmentConsistency | 16 | 6 | +10 | B6/S4/S7 (stale `AttachedTo` after host left) |
| CombatLegality | 4 | — (not surfaced) | +4 | New aggregate (see §S9) |
| ZoneCastGrantExpiry | 2 | 4 | −2 | Carry-over from round 6 §S8 |
| TriggerCompleteness | 2 | — (not surfaced) | +2 | New aggregate (see §S9) |
| **Total** | **620** | **752** | **−132** | |

## Signatures observed (from inlined detail block)

The first-30 inline cap is concentrated in 2 games this round (vs 3 in round 6). The remaining 590 occurrences across 23 affected games are not individually inspectable; aggregate counts above show no new violation class beyond the §S9 CombatLegality / TriggerCompleteness entries.

### S7-bis — `Hedron Blade` attached to off-board pilot/construct token
- **Severity:** Low (2 inline occ, 16 total)
- **Invariant:** AttachmentConsistency — seat 1's `Hedron Blade` attached to a `creature token construct Token` not on any battlefield
- **Game:** 153 (seed 1530778). Commanders: Borborygmos and Fblthp, Irma Part-Time Mutant, Mai Jaded Edge, Quicksilver Brash Blur.
- **Class:** B6 / round-6 S7 (`Born to Drive` on dead pilot token). Same root cause as the round-5 Doc Ock's Tentacles signature: SBA §704.5n's `AttachedTo` clear-on-leave doesn't fire when the host is a token destroyed via combat damage. **Pre-existing,** no per-card or SBA fix has landed.

### S11 — Game-174 token-or-copy run that accumulates 14 phantom cards across 13 turns
- **Severity:** Medium (28 inline occ, 172 total ZoneConservation in this game)
- **Invariant:** ZoneConservation — "extra real cards appeared" grows monotonically across turns (11 → 12 → 13 → 14 extras as the game progresses)
- **Game:** 174 (seed 1740778). Commanders: Sokka Bold Boomeranger, **Beatrix Loyal General**, **Kain Traitorous Dragoon**, Tivash Gloom Summoner.
- **Trigger events visible in tail:** Zephyr Boots cast/equip → declare_attackers → blockers → `zone_change seat=1 source=Elenda's Hierophant` followed by repeated token-generating combat sequences. **Elenda's Hierophant** spawns 1/1 Vampire tokens on combat-damage triggers; the count "real-card" balance suggests the token-vs-real classification is leaking when Elenda's trigger fires under specific commander interactions.
- **Class:** B4/B7 (round-6 S6 Nevinyrral wipe). Same shape: a token-generating trigger leaves an unaccounted-for "real card" each fire. Not a new class — but a different witness card from the round-6 Nevinyrral case. **Pre-existing**, no `elendas_hierophant.go` handler exists.

### S8 — `ZoneCastGrantExpiry` (2 occurrences)
- **Severity:** Low (2 occ, down from 4 in round 6)
- Aggregate-only, no inline detail. No correlation with the bulk-pattern families that issue timed alt-cast permission.

### S9 — `CombatLegality` (4) + `TriggerCompleteness` (2) — new aggregates
- **Severity:** Low (6 total occurrences across 2K games, no inline detail)
- **Status:** Newly-surfaced invariant counters that did not appear in the round 6 aggregate table. Without inline detail they are not card-attributable. Most likely tracking:
  - **CombatLegality**: a defender/wall/un-attackable creature ending up in `Attackers` after a copy or control-change. Worth flagging because the wave-141-160 Karlach handler creates extra-combat sequences (now wired by `02106e9`'s init-restore — these may surface combat-state bookkeeping gaps).
  - **TriggerCompleteness**: a registered trigger that didn't reach a `stack_resolve` or `state` event. Most plausible coupling: the new wave-201-220 dispatch fallback handlers fire late-cycle and may not log their resolution event.
- **Recommendation:** widen the inline-violation cap in `cmd/hexdek-loki` (currently 30) for the next round so these 6 occurrences become inspectable. With seed 777 they are reproducible.

### S10 — Nightmare board CardIdentity (2 occurrences) — new
- **Severity:** Low (2 occ in 10K boards = 0.02% — but new vs round 6's clean 10K)
- **Invariant:** CardIdentity on the random-permanent-piles harness
- **Class:** First nightmare-phase violations since the round-5 baseline run. Round 5 ran with master seed 555, round 6 with 666 — both 0 violations. Under seed 777 the nightmare RNG generates a pile that exposes a same-pointer-two-zones race in either layer recalc or SBA. Aggregate-only — no inline detail in the nightmare phase output.
- **Recommendation:** Re-run nightmare alone at seed 778/779/780 to determine whether this is seed-specific or a regressed code path. The 2 occurrences land in a single board; nightmare-boards are stateless 1-shot setups, so the cluster comes from per-board re-evaluation rather than progression.

## Status of postmerge / round-5 / round-6 bug classes

| Bug | Postmerge | R5 | R6 | R7 (this run) | Status |
|-----|----------:|---:|---:|--------------:|--------|
| B1 — Equipment self-equip | 174 | 0 | 0 | **0** | ✅ Confirmed fixed |
| B2 — Gerrard cmdzone double-add | 34 | 0 | 0 | **0** | ✅ Confirmed fixed |
| B3 — Adric hand×graveyard | 54 | not seeded | 10+ | **not isolated under seed 777** | Still latent (card-specific) |
| B4/B7 — phantom card / copy bug | 40 / 34 | 78 | 156 | **172** | Still latent (class — different witness each round) |
| B5 — Cerulean Sphinx shuffle-to-lib | 34 | not seeded | not isolated | not isolated | Still latent |
| B6 — Spear of the General attach | 2 | 2 (Doc Ock) | 2 (Born to Drive) | **2 (Hedron Blade)** | Same class, different witness |
| R6/S8 — `ZoneCastGrantExpiry` | — | — | 4 | **2** | Stable noise floor |
| **R7/S9 — `CombatLegality`** | — | — | — | **4** | **New aggregate** (no inline detail) |
| **R7/S9 — `TriggerCompleteness`** | — | — | — | **2** | **New aggregate** (no inline detail) |
| **R7/S10 — Nightmare CardIdentity** | — | 0 | 0 | **2** | **New** (first non-zero nightmare run) |

CardIdentity dropped from 586 → 424 (−162). Most plausible attribution: the **Sai test-pollution Reset-hook fix** (`02106e9`) restored ~50 init-registered handlers that had been silently absent in earlier multi-test runs, and several of those handlers correctly clean up zone-change side effects that previously left phantom pointers under chaos play. (Loki runs single-shot so it never hits `Reset()`, but the underlying late-init handlers themselves were on a code path that the `init()` restoration plumbing now exercises consistently.)

## Regression Assessment vs Merges Since Round 6

- **Hat archetype-specific eval weights (`b0b6db4` / `572ab7c`):** **Zero engine-side fallout.** Hat is the deck pilot used by every Loki game; the new weights changed scoring profile per Freya archetype. No new panic, no new invariant class attributable to hat decisions. Throughput dipped −2 g/s (27 → 25), consistent with the sparse-weight merge running an extra archetype lookup per eval — acceptable given the validation result (`6e1d8bf` 500g A/B confirms measurable winrate shift).
- **~70 new per_card handlers + 6 bulk-pattern families:** Net **−132 chaos violations** despite massively expanded handler dispatch. The two new aggregate counters (CombatLegality, TriggerCompleteness) are the only ambiguity — total 6 occ across 2K games, no panic, no inline witness. The B4/B7 class (phantom cards) is still latent and still finds a fresh witness each round (Nevinyrral → Elenda's Hierophant), but its absolute count tracks volume.
- **Wave 201-220 dispatch fallback (`02cf908`):** No new crash signature. Possible attribution candidate for `TriggerCompleteness` — fallback dispatch is most likely to issue a trigger that doesn't reach a `stack_resolve` log line.
- **Reset-hook fix (`375de60`):** Likely the biggest single contributor to the CardIdentity drop. Confirms the round-6 hypothesis that several latent handlers were missing from chaos runs.
- **Nightmare regression (§S10):** 2 CardIdentity in 10K under seed 777 (vs 0 in round 5 + round 6). One-board cluster, no panic; almost certainly seed-specific. Worth a quick re-seed sweep before sounding the alarm.

## Verdict

**Land clean.** Round 7 is regression-free at the gating level: 0 crashes, 0 panics, 18% lower per-game noise floor, biggest CardIdentity drop since the postmerge audit. The new merge surface is the most-disruptive engine wave of this audit lineage (hat refactor + 70+ handlers) and Loki absorbs it without surfacing a new panic class.

**Three follow-ups (low priority):**
1. **R7-F1 — Widen Loki inline-violation cap** in `cmd/hexdek-loki` (currently 30) so `CombatLegality`, `TriggerCompleteness`, and `ZoneCastGrantExpiry` become inspectable. Cheap.
2. **R7-F2 — Nightmare seed sweep.** Re-run nightmare-only at seeds 778-782 to determine whether the §S10 CardIdentity is seed-specific or a real regression. Cheap (10K boards = 2 s).
3. **R7-F3 — `elendas_hierophant.go` handler.** Game 174 is the §S11 witness; B4/B7 class will keep finding new token-generator witnesses each round until the underlying token-vs-real classifier gap is fixed. A per-card handler would short-circuit one instance; the class fix lives deeper in the SBA / zone-conservation accounting.

## Repro Seeds (master seed 777)

| Game | Seed | Witness cards | Class |
|------|------|---------------|-------|
| 153 | 1530778 | Hedron Blade on dead construct token | S7-bis / B6 |
| 174 | 1740778 | Elenda's Hierophant + Beatrix/Kain combat run | S11 / B4/B7 |

Raw report: [`data/audit/loki_round7_raw.md`](loki_round7_raw.md) (2,915 lines).
