# Goldilocks Round-6 Audit

**Date:** 2026-05-16
**Branch:** `dev/loki-goldilocks-round6` (from `main` @ `20c0c15`)
**Tool:** `cmd/hexdek-thor --goldilocks --workers 10`
**Inputs:** `data/rules/ast_dataset.jsonl` (31,963 cards), `data/rules/oracle-cards.json` (35,708 cards)
**Raw failures CSV:** [`goldilocks_round6.csv`](goldilocks_round6.csv) (64 rows)
**Baseline:** Round-5 (`a3e8f95`, 2026-05-16) — 64 failures, 0 panics, 99.80% pass

## Headline

| Metric | Round 6 (this run) | Round 5 baseline | Δ |
|---|---:|---:|---:|
| Cards loaded | 35,708 | 35,708 | 0 |
| AST-verified tests run | 31,963 | 31,963 | 0 |
| Passed | 30,277 | 30,277 | 0 |
| Dead-effect | 61 | 61 | 0 |
| Invariant violations | 2 | 2 | 0 |
| Unverified (verifier gap) | 1 | 1 | 0 |
| **Panics** | **0** | **0** | **0** |
| Skipped (no abilities) | 4,106 | 4,106 | 0 |
| Keyword tests | 2,013 pass / 0 fail | 2,013 pass / 0 fail | 0 |
| Wall clock | 1.639 s | 0.593 s | +1.05 s (Goldilocks-only, see note) |
| Throughput | 19,486 tests/s | 55,335 tests/s | — |

> **Wall-clock note:** Round 5 reported the full Thor process at 593 ms because it benefitted from a warm `go build` cache. This round was a cold first-run on the new branch (3.0 s AST load + 2.0 s oracle load + 1.6 s goldilocks). Re-running back-to-back drops the goldilocks phase to ~600 ms; no real slowdown.

**Pass rate: 99.80%. Zero panics. Zero drift versus the round-5 baseline.**

## Merge wave under test

Since round 5 was captured at `a3e8f95`, `main` advanced to `20c0c15` with these merges:

| Commit | Merge | Surface |
|---|---|---|
| `20c0c15` | `dev/muninn-status-2` | Docs only |
| `695f78f` | `dev/deploy-script-fix` | Ops only — no engine/card surface |
| `35b3921` | `dev/perf-round-2` | Hot-path microopts: bitmask color cache + combo-eval alloc fix (−22% alloc, −13% CPU) |
| `589ffd8` | `dev/visual-polish-round-3` | Frontend only |
| `2e07323` | `dev/muninn-handlers-51-70` | 10 new per_card handlers |
| `4318132` | `dev/muninn-bulk-patterns-2` | 3 more bulk-pattern handler families |
| `42e95a5` | `dev/muninn-handlers-41-50` | 10 new per_card handlers |
| `39c8b6e` | `dev/visual-polish-round-2` | Frontend only |

Net engine surface: **20 new per_card handlers + 3 new bulk-pattern families + the perf-2 hot-path edits**. All pass the AST-verified Goldilocks sweep. No new dead-effect, invariant, or panic signatures introduced.

## Diff vs. baseline (failure set)

`diff <(sort goldilocks_round5.csv) <(sort goldilocks_round6.csv)`:

```
< "Etali, Primal Conqueror // Etali, Primal Sickness", … (ptr 0xc00442a0f0) appears in both seat 0 battlefield and seat 0 battlefield
> "Etali, Primal Conqueror // Etali, Primal Sickness", … (ptr 0xc006cbd400) appears in both seat 0 battlefield and seat 0 battlefield
```

Single byte-level diff: a nondeterministic Go heap pointer in the Etali invariant message. Every other failing card, kind, ability kind, and effect bucket is byte-identical to round 5.

### Failure breakdown (unchanged)

| Kind | Count |
|---|---:|
| `goldilocks_dead_effect` | 61 |
| `goldilocks_invariant` | 2 |
| `goldilocks_unverified` | 1 |

Dead-effect by effect bucket (unchanged):

| Effect | Count |
|---|---:|
| `ability_word` (static) | 45 |
| `sacrifice` (4 triggered + 1 static) | 5 |
| `modification_effect` (triggered) | 4 |
| `create_token` (triggered) | 2 |
| `exile` (triggered) | 2 |
| `lose_life` (static) | 2 |
| `destroy` (activated) | 1 |

Invariants (unchanged carry-over):

1. **`Etali, Primal Conqueror // Etali, Primal Sickness` — CardIdentity.** Same card pointer indexed twice on the same battlefield. Recommendation R1 from the postmerge audit (transform/exile-and-cast dedupe) still open.
2. **`Phage the Untouchable` — TurnStructure.** `lose_game` fires without setting `LossReason`. Recommendation R2 from the postmerge audit still open.

Unverified (unchanged):

3. **`Expose the Culprit` — `turn_face_up`** not in Goldilocks verifiable set. Recommendation R3 still open.

## Conclusion

The round-5 → round-6 merge wave (20 new per_card handlers, 3 bulk-pattern families, perf round 2, and ops/UI changes) introduces **zero regressions** to Goldilocks. Failure set is byte-identical to round 5 except for one Go heap address. Pass rate holds at 99.80% with zero panics. The two pre-existing engine-side bugs (Etali, Phage) and the one verifier gap (`turn_face_up`) carry forward unchanged; recommendations R1–R3 from `goldilocks_postmerge.md` remain the next-action list.

Notably, the perf-round-2 edits — which touched mana-color cache invalidation paths and combo-eval allocation patterns — produced **no Goldilocks-observable behavior change**, supporting the claim that the optimization is semantics-preserving.
