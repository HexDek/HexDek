# Goldilocks Round-5 Audit

**Date:** 2026-05-16
**Branch:** `dev/goldilocks-round5` (from `main` @ `3255991`)
**Tool:** `cmd/hexdek-thor` `--goldilocks --workers 10`
**Inputs:** `data/rules/ast_dataset.jsonl` (31,963 cards), `data/rules/oracle-cards.json` (35,708 cards)
**Raw failures CSV:** [`goldilocks_round5.csv`](goldilocks_round5.csv) (64 rows)
**Baseline:** `dev/goldilocks-postmerge` @ 2026-05-16 (commit `b6e6d43`) — 64 failures, 0 panics, 99.80% pass

## Headline

| Metric | Round 5 (this run) | Postmerge baseline | Δ |
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
| Wall clock | 593 ms | 577 ms | +16 ms |

**Pass rate: 99.80%. Zero panics. Zero drift versus the postmerge baseline.**

## Round-4 merge wave under test

Since the baseline was captured, `main` advanced from `b6e6d43` to `3255991` with five merges:

| Commit | Merge | Surface |
|---|---|---|
| `b22b39b` | `dev/tournament-pool-output-fix` | Tournament CLI dashboard rendering — no engine/card surface |
| `c9006e1` | `dev/muninn-bulk-patterns` | Bulk-pattern handlers for card families |
| `9dc606e` | `dev/muninn-handlers-31-40` | 8 per_card handlers for Muninn snowflakes #31–#40 |
| `ab191df` | `dev/muninn-handlers-21-30` | 9 per_card handlers (#21–#30 + stragglers) |
| `a05abd5` | `dev/muninn-progress` | Docs only |

All 17 new per_card handlers + the bulk-pattern handler set pass the AST-verified Goldilocks sweep. No new dead-effect, invariant, or panic signatures introduced.

## Diff vs. baseline (failure set)

`diff <(sort goldilocks_postmerge.csv) <(sort goldilocks_round5.csv)` is empty modulo one nondeterministic Go heap pointer in the Etali invariant message:

```
< (ptr 0xc005ba0e10) appears in both seat 0 battlefield and seat 0 battlefield
> (ptr 0xc00442a0f0) appears in both seat 0 battlefield and seat 0 battlefield
```

Every other failing card, kind, ability kind, and effect bucket is byte-identical.

### Failure breakdown (unchanged)

| Kind | Count |
|---|---:|
| `goldilocks_dead_effect` | 61 |
| `goldilocks_invariant` | 2 |
| `goldilocks_unverified` | 1 |

Dead-effect by effect bucket:

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

Round-4 merges (Muninn handlers #21–40 + bulk patterns + tournament pool output fix) introduce **zero regressions** to Goldilocks. The corpus result is identical to the postmerge baseline down to the failing-card list — only a heap address differs. Pass rate holds at 99.80% with zero panics and zero new invariant violations. The two pre-existing engine-side bugs (Etali, Phage) and the one verifier gap (`turn_face_up`) carry forward unchanged; recommendations R1–R3 from `goldilocks_postmerge.md` remain the next-action list.
