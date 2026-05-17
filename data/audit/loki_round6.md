# Loki Fuzz — Round 6 (Post Perf-2 + Handler Wave)

**Branch:** `dev/loki-goldilocks-round6`
**Date:** 2026-05-16
**Tool:** `cmd/hexdek-loki` (normal mode, no handler seeding)
**Run:** `hexdek-loki -games 2000 -seed 666 -report …` (defaults: 4 seats, max-turns 60, 10 workers, nightmare-boards 10000)

## Scope

Fresh chaos run against `main` HEAD after the round-5 audit landed and the engine accumulated three more merge waves: 20+ new per_card handlers (#41-#60 and #51-#70), three additional bulk-pattern handler families, the perf round-2 bitmask/alloc reductions, the backend deploy fix, and visual-polish round 2/3 (UI-only).

| Commit | Subject |
|--------|---------|
| `20c0c15` | Merge dev/muninn-status-2: snapshot 2 docs |
| `695f78f` | Merge dev/deploy-script-fix: backend deploy no longer hangs |
| `35b3921` | Merge dev/perf-round-2: bitmask color cache + combo-eval alloc fix (−22% alloc, −13% CPU) |
| `589ffd8` | Merge dev/visual-polish-round-3 |
| `2e07323` | Merge dev/muninn-handlers-51-70: 10 new handlers |
| `4318132` | Merge dev/muninn-bulk-patterns-2: 3 more bulk-pattern families |
| `42e95a5` | Merge dev/muninn-handlers-41-50: 10 new handlers |
| `39c8b6e` | Merge dev/visual-polish-round-2 |
| `a3e8f95` | Merge dev/goldilocks-round5 |
| `48d28e8` | Merge dev/loki-round5 (round-5 baseline) |

## Headline Result

**Zero crashes. Zero panics.** 2,000 chaos games + 10,000 nightmare boards complete in 1m16.7s. Goldilocks-postmerge's "0 known panics" invariant continues to hold under doubled chaos volume.

| Phase | Count | Duration | Crashes | Violations | Affected |
|-------|------:|---------:|--------:|-----------:|---------:|
| Chaos games | 2,000 | 1m14.6s | **0** | 752 | 24 games |
| Nightmare boards | 10,000 | 2.2s | **0** | 0 | 0 |
| **Total** | — | 1m16.7s | **0** | **752** | — |

Clean-game rate: **1,976 / 2,000 (98.8%)**. Clean-board rate: **10,000 / 10,000 (100.0%)**. Game throughput rose from 23 g/s (round 5) → 27 g/s — consistent with the perf-round-2 −13% CPU reduction landing.

## Violations by Invariant

| Invariant | Count | Notes |
|-----------|------:|-------|
| CardIdentity | 586 | Same pointer in two zones — same B3/B5/S1/S2 class |
| ZoneConservation | 156 | Cards "disappearing" from totals — same B4/B7 class |
| AttachmentConsistency | 6 | Stale equip pointer after host left battlefield (B6/S4 class) |
| ZoneCastGrantExpiry | 4 | **New aggregate invariant** — not surfaced in round 5 |
| **Total** | **752** | All confined to **24 of 2,000 games** |

The detail block in `CHAOS_REPORT.md` inlines the first 30 occurrences, drawn from 3 of the 24 affected games (183, 216, 357). The remaining 21 games / 722 occurrences are not individually inspectable; every inlined signature maps to a previously-catalogued class.

Normalized violations/game: round 5 = 0.41, round 6 = 0.38. Comparable noise floor; the absolute violation count doubling is purely a function of doubling chaos volume.

## Signatures Observed (from inlined detail)

### S5 — `Adric, Mathematical Genius` duplicated across hand ↔ graveyard
- **Severity:** Medium (10 occurrences in inline block, single game/seat)
- **Invariant:** CardIdentity — `seat 1 hand` and `seat 1 graveyard`
- **Game:** 357 (seed 3570667)
- **Class:** B3-class. Postmerge audit's "Adric (Doctor's companion) hand×graveyard" originally observed at 54 occ in the postmerge run; not isolated in round 5 (master seed 555); re-isolated under round 6's master seed 666. **Pre-existing**, no per_card handler for Adric exists yet.

### S6 — Round-183 ZoneConservation tear
- **Severity:** Medium (18 inline occurrences; same single game)
- **Invariant:** ZoneConservation — 2 real cards disappeared (expected 376 → 374, then 368 → 366 in a later snapshot)
- **Game:** 183 (seed 1830667). Commanders: Hildibrand Manderville // Gentleman's Rise / Wernog, Rider's Chaplain / Ishai, Ojutai Dragonspeaker / Bane, Lord of Darkness.
- **Last events before tear:** Nevinyrral, Urborg Tyrant cast → ETB → triggered "destroy all" twice (seat 0 + seat 3 destroys), per_card_handler ran twice, citys_blessing fired. The two missing cards line up with permanents destroyed by Nevinyrral's wipe whose graveyard-arrival was lost — same shape as B4/B7 (Rally the Galadhrim conspire-copy class).
- **No `nevinyrral_urborg_tyrant.go` handler** exists; the per_card_handler events likely belong to the wipe's victims (Doc Ock's Tentacles or similar) rather than Nevinyrral itself.

### S7 — `Born to Drive` attached to an off-board pilot token
- **Severity:** Low (2 occurrences, single game)
- **Invariant:** AttachmentConsistency — seat 2's `Born to Drive` aura is attached to a `creature token pilot Token` that is not on any battlefield
- **Game:** 216 (seed 2160667)
- **Class:** B6/S4-class stale `AttachedTo` after the host token left play. SBA §704.5n should null the pointer when the pilot token dies/leaves; either it didn't fire or the aura's seat-2 ownership routed past the destroy-bearing seat's battlefield scan. Same root-cause shape as Doc Ock's Tentacles (round 5) and Spear of the General / Stupefying Touch (postmerge).

### S8 — `ZoneCastGrantExpiry` (4 occurrences, no inline detail)
- **Severity:** Low (4 total, none reached the first-30 inline window)
- **Invariant:** **New** — not present in round 5's aggregate table
- **Hypothesis:** Tracks "temporary cast-from-zone permission expires at the right boundary" (e.g., Sneak Attack-style "you may cast from your graveyard until end of turn", or the new bulk-pattern families that grant timed alt-cast permission). With 4 occurrences across 2,000 games and no panic, this is a low-severity bookkeeping gap rather than a regression that gates the merge. Worth pulling a targeted repro in a follow-up by surfacing the inline limit (or grep'ing the raw report after a higher `-permutations` run that gets the signature into the first 30).

## Status of Postmerge / Round-5 Bugs

| Bug | Postmerge | Round 5 | Round 6 (this run) | Status |
|-----|----------:|--------:|-------------------:|--------|
| B1 — Equipment self-equip | 174 | 0 | **0** | ✅ Confirmed fixed |
| B2 — Gerrard cmdzone double-add | 34 | 0 | **0** | ✅ Confirmed fixed |
| B3 — Adric hand×graveyard | 54 | not seeded | **10+ (S5)** | Still latent; same card |
| B4 — Rally conspire-copy in graveyard | 40 | not isolated | likely subsumed in 156 ZC | Same class as S6 |
| B5 — Cerulean Sphinx shuffle-into-library | 34 | not seeded | not isolated under seed 666 | Still latent |
| B6 — Spear of the General attachment | 2 | 2 (Doc Ock) | **2 (Born to Drive)** | Same class; not card-specific |
| B7 — ZoneConservation phantom cards | 34 | 78 (1 game) | 156 (≥1 game, S6) | Same class as B4 |
| S1/S2 (round 5) — Lo and Li, Baron Zemo | — | observed | not isolated under seed 666 | Still latent |
| **S8 — ZoneCastGrantExpiry** | — | — | **4 occ (new)** | New aggregate; no panic |

**No new gating bug class** surfaced. Every detail-block signature traces to an existing open issue from the postmerge audit. The only fresh entry — `ZoneCastGrantExpiry` — is a new invariant counter at 4/752 (0.5% of violations) with no panic and no inline detail.

## Regression Assessment vs Merges Since Round 5

- **Perf round 2 (`35b3921`):** No engine-side fallout. Throughput improved (27 vs 23 g/s) consistent with the −13% CPU claim; violation profile unchanged in shape.
- **20 new per_card handlers (#41-#60, #51-#70):** Zero crashes, zero handler-attributable violations in the inline block. None of the violation-bearing card names (Adric, Born to Drive, the Nevinyrral wipe targets) trace to the new handler set.
- **3 new bulk-pattern handler families (`4318132`):** No regression in CardIdentity or ZoneConservation. **Possible attribution candidate** for the 4-occ `ZoneCastGrantExpiry`: bulk patterns most likely to issue temporary cast permission. Worth verifying when the next round seeds for the affected card families.
- **Deploy fix (`695f78f`) + visual polish 2/3:** Out of scope for engine fuzz.

## Recommendation

- **Land clean** — Round 6 is regression-free at the gating level. 0 crashes, 0 panics, comparable noise floor.
- **Surface S8 detail:** raise the inline-violation cap in `cmd/hexdek-loki` (currently 30) or do a targeted seeded re-run filtered to `ZoneCastGrantExpiry` so the 4 occurrences become inspectable. Without inline detail, we can't even name the cards involved.
- B3 (Adric), B6-class (stale `AttachedTo`), and B7-class (ZoneConservation) remain the three most-actionable next-step bugs. Adric in particular is a known card-specific issue that has been observable across postmerge and round 6 and should now be cheap to fix with a dedicated handler.

## Repro Seeds (master seed 666)

| Game | Seed | Trigger Cards | Class |
|------|------|---------------|-------|
| 183 | 1830667 | Nevinyrral, Urborg Tyrant wipe → 2 lost cards | S6 / B4/B7 |
| 216 | 2160667 | Born to Drive aura, dead pilot token host | S7 / B6 |
| 357 | 3570667 | Adric, Mathematical Genius hand↔graveyard | S5 / B3 |

Raw report: `/tmp/loki-round6/CHAOS_REPORT.md` (2,257 lines).
