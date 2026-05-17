# Goldilocks Round-7 Audit

**Date:** 2026-05-17
**Branch:** `dev/audits-round-7` (from `main` @ `127df76`)
**Tool:** `cmd/hexdek-thor --goldilocks --workers 10`
**Inputs:** `data/rules/ast_dataset.jsonl` (31,963 cards), `data/rules/oracle-cards.json` (35,708 cards)
**Raw failures CSV:** [`goldilocks_round7.csv`](goldilocks_round7.csv) (62 rows)
**Raw markdown report:** [`goldilocks_round7_raw.md`](goldilocks_round7_raw.md)
**Baseline:** Round-6 (`20c0c15`, 2026-05-16) — 64 failures, 0 panics, 99.80% pass

## Headline

| Metric | Round 7 | Round 6 | Δ |
|---|---:|---:|---:|
| Cards loaded | 35,708 | 35,708 | 0 |
| Tests run (Thor total) | 31,963 | 31,963 | 0 |
| Failures | 62 | 64 | **−2** |
| Dead-effect | 57 | 61 | **−4** |
| Invariant violations | 4 | 2 | **+2** |
| Unverified (verifier gap) | 1 | 1 | 0 |
| **Panics** | **0** | **0** | **0** |
| Skipped (no abilities) | 4,106 | 4,106 | 0 |
| Keyword tests | 2,013 pass / 0 fail | 2,013 pass / 0 fail | 0 |
| Goldilocks wall clock | 1.233 s | ~0.6 s warm | within noise |

**Pass rate: 99.81% (+0.01 pp). Zero panics. Net failure count down 2.**

## Merge wave under test

`main` advanced from round-6 base `20c0c15` → `127df76` (35 merges). Engine-relevant surface:

| Surface | Detail |
|---|---|
| `dev/hat-improvements` (`572ab7c`) + `dev/hat-weights-validation` (`6e1d8bf`) | Archetype-specific eval weights for 9 Freya archetypes + sparse-weight merge; 500g A/B confirms archetype dispatch measurably shifts winrates |
| Muninn handler waves 51-60-finish, 61-80, 81-100, 101-120, 121-140, 141-160, 161-180, 181-200, 201-220 | ~70+ new per_card handlers + dispatch fallback for cascade/copy/token name variants |
| Muninn bulk-pattern families 3, 4, 5 | end-step-intervening-if, gated-etb-effect, shuffle-self-from-grave, etb-library-tutor, etb-basic-land-ramp, etb-drain-target-opponent (6 new families) |
| `dev/sai-test-pollution` (`375de60`) | Reset hooks for late-`init()` handler registrations — restores ~50 handlers stripped by `per_card.Reset()` after `TestAllRegisteredTriggersAreDispatched` |
| Frontend, docs, deploy | Hot Cards widget, Similar Decks widget, visual polish 4/5/6/7/EOD, era audit docs, day summaries — out of scope for Goldilocks |

Net engine surface vs round 6: **~70 new per_card handlers, 6 new bulk-pattern families, the hat archetype-weight refactor, and the Reset-hook fix.**

## Diff vs. round-6 baseline (failure set)

`diff <(sort goldilocks_round7.csv) <(sort goldilocks_round6.csv)`:

```
< Etali, Primal Conqueror // Etali, Primal Sickness — CardIdentity (ptr 0xc003f3f4a0)        # heap-addr churn
> Etali, Primal Conqueror // Etali, Primal Sickness — CardIdentity (ptr 0xc006cbd400)

< Finest Hour                      goldilocks_invariant TurnStructure "begin_combat" invalid for phase "combat"   # NEW
< Karlach, Fury of Avernus         goldilocks_invariant TurnStructure "begin_combat" invalid for phase "combat"   # NEW

> Pestilence                       goldilocks_dead_effect sacrifice/triggered/self           # FIXED
> Pyrohemia                        goldilocks_dead_effect sacrifice/triggered/self           # FIXED
> Task Mage Assembly               goldilocks_dead_effect sacrifice/triggered/self           # FIXED
> Withering Wisps                  goldilocks_dead_effect sacrifice/triggered/self           # FIXED
```

Net: **4 sacrifice/triggered/self dead-effects resolved, 2 new TurnStructure invariants on extra-combat handlers, plus the usual non-deterministic Go heap pointer in the Etali message.** Every other failing card is byte-identical to round 6.

### Failure breakdown

| Kind | Round 7 | Round 6 | Δ |
|---|---:|---:|---:|
| `goldilocks_dead_effect` | 57 | 61 | −4 |
| `goldilocks_invariant` | 4 | 2 | +2 |
| `goldilocks_unverified` | 1 | 1 | 0 |

Dead-effect by effect bucket:

| Effect | Round 7 | Round 6 | Δ |
|---|---:|---:|---:|
| `ability_word` (static) | 45 | 45 | 0 |
| `modification_effect` (triggered) | 4 | 4 | 0 |
| `lose_life` (static) | 2 | 2 | 0 |
| `create_token` (triggered) | 2 | 2 | 0 |
| `exile` (triggered) | 2 | 2 | 0 |
| `sacrifice` (triggered + static) | **1** | **5** | **−4** |
| `destroy` (activated) | 1 | 1 | 0 |

Invariant violations:

| Card | Invariant | Status |
|---|---|---|
| `Etali, Primal Conqueror // Etali, Primal Sickness` | CardIdentity (battlefield double-index) | Carry-over (R1 from postmerge audit; transform/exile-and-cast dedupe still open) |
| `Phage the Untouchable` | TurnStructure (`lose_game` fires without `LossReason`) | Carry-over (R2 still open) |
| **`Finest Hour`** | **TurnStructure (`[untap]` step `begin_combat` invalid for phase `combat`)** | **New** — see Triage below |
| **`Karlach, Fury of Avernus`** | **TurnStructure (`[untap]` step `begin_combat` invalid for phase `combat`)** | **New** — see Triage below |

Unverified:

| Card | Effect | Status |
|---|---|---|
| `Expose the Culprit` | `turn_face_up` not in verifier set | Carry-over (R3 still open) |

## Triage: Finest Hour / Karlach TurnStructure invariants

Both cards generate **additional combat phases** via a new wave-141-160 handler (`karlach_fury_of_avernus.go` landed in `ca27f7e`). The invariant fires when the test scaffold's `untap` step is invoked against a phase context still labelled `"combat"` from the prior extra combat. Three observations:

1. **Same signature, distinct cards.** Identical message string suggests a single root cause: the Goldilocks scaffold doesn't reset phase state cleanly when a handler injects an extra combat into the per-card harness.
2. **Engine side is consistent with the new dispatch.** No regression in live-game Loki paths — Round 7 chaos shows 0 occurrences of this `TurnStructure` signature in 2,000 games. The failure is harness-only.
3. **Both cards previously read as goldilocks-passing because no handler existed.** Pre-wave-141-160, Karlach's extra-combat trigger was a dead-effect (and would have shown as `dead_effect` in round 6 if attempted) — except it wasn't even reached because no per_card handler dispatched the trigger. The new handler exposes a pre-existing scaffold gap.

**Severity: Low (scaffold gap, not engine bug).** Fix path lives in `cmd/hexdek-thor/goldilocks.go` (phase/step reset after a handler-injected combat sequence). Same shape as the wave-201-220 dispatch-fallback work already in this batch — narrow follow-up.

## Carry-over status (postmerge audit recommendations)

| Recommendation | Status |
|---|---|
| R1 — Etali transform/exile-and-cast dedupe | Open (carries) |
| R2 — Phage `lose_game` `LossReason` plumbing | Open (carries) |
| R3 — `turn_face_up` verifier gap for `Expose the Culprit` | Open (carries) |
| **R4 — Goldilocks scaffold extra-combat phase reset (new)** | Open (new from this round) |

## Verdict on the round-7 merge wave

- **Hat archetype-weight refactor (`b0b6db4` / `572ab7c`):** Zero Goldilocks-observable behavior change. Hat logic is not exercised by per-card Goldilocks (no AI in the scaffold), so this is expected — but it confirms the refactor didn't accidentally bleed into per_card dispatch.
- **~70 new per_card handlers + 6 bulk-pattern families:** Net **−4 dead-effects** with **−2 sacrifice-bucket fails** specifically resolved. The new wave-141-160 Karlach handler surfaced one new scaffold gap (R4) — a pre-existing latent issue, not a regression.
- **Reset-hook fix (`02106e9` / `375de60`):** Goldilocks runs single-shot, so it never hits a `per_card.Reset()`, but the change is exercised by `TestAllRegisteredTriggersAreDispatched` which now passes 3× clean (per the test-flake survey doc `13a0a6a`).
- **Dispatch fallback wave 201-220:** No new dead-effects from cascade/copy/token-name variants.

**Conclusion: Round 7 is regression-free at the gating level.** Pass rate moved 99.80% → 99.81%. The two new TurnStructure entries are harness-side and traceable to a real-engine improvement (Karlach now actually fires its extra-combat). Recommendations R1–R3 carry forward; R4 is added.

## Throughput

| Metric | Round 7 | Round 6 (cold) | Round 5 (warm) |
|---|---:|---:|---:|
| Goldilocks wall | 1.233 s | 1.639 s | 0.593 s |
| Goldilocks rate | 26,244 tests/s | 19,486 tests/s | 55,335 tests/s |

Cold-cache numbers comparable to round 6; warm reruns drop to ~600 ms.
