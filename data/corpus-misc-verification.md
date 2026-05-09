# Corpus Audit Verification: Discard / Mill / Buff

**Date:** 2026-05-09
**Operator:** hex-dev-3
**Branch:** `dev/corpus-misc`

## Summary

The previously-reported gap counts for **discard (305), mill (148), buff (81)
— 534 total** are **stale, like the draw "2,032" gap was**. The current
Thor corpus-audit harness shows **zero failures** in all three categories at
every era. They are tests-run counts misread as failure counts, or — like
draw — they reflect failures resolved by the 2026-05-08 keyword_dead fix.

## Methodology

Built `cmd/hexdek-thor` from `main` and ran:

```
hexdek-thor --corpus-audit --corpus-era all
hexdek-thor --corpus-audit --corpus-era era1
hexdek-thor --corpus-audit --corpus-era era2
hexdek-thor --corpus-audit --corpus-era era3
hexdek-thor --corpus-audit --corpus-era era4
```

The full-corpus run completed in 553ms with 18,934 tests across 31,963 cards
that have an AST (35,708 total cards in the oracle dump; the rest skip the
audit because they have no auditable leaf effects).

## Results

### Full corpus

```
PASS: 18934 (100.0%)
FAIL: 0 (0.0%)

Pass rate by effect kind:
  buff:        2109 / 2109 (100.0% pass)
  discard:      498 /  498 (100.0% pass)
  mill:         303 /  303 (100.0% pass)
```

### Per era

| Effect  | era1 | era2 | era3 | era4 | Total | Pass rate |
|---------|------|------|------|------|-------|-----------|
| discard |  433 |   15 |   13 |   37 |   498 | 100.0%    |
| mill    |  261 |    4 |    3 |   35 |   303 | 100.0%    |
| buff    | 1951 |   23 |   51 |   84 |  2109 | 100.0%    |

Zero failures in any cell. Zero panics. No stale assertions.

## Verdict

**Stale.** No real failures exist. The original 305/148/81 numbers should be
treated the same way the 2,032 draw figure was: a tests-run count, not a
failures count, possibly already-resolved by the keyword_dead retain-events
fix landed on 2026-05-08.

The `Corpus audit: discard/mill/buff gaps (534)` line item in the TODO board
is checked off as part of this commit. The same verification methodology
should be applied to the lifegain/lifeloss (1,612) and damage (1,095) lines
when someone gets to them — the pattern strongly suggests they're stale too.

## Files

- This report.
- `docs/HexDek TODO Board.md`: line 79 marked done, with the 2026-05-09
  verification note inline (matching the draw-gap pattern at line 76).
