# Performance Audit — Round 2 (2026-05-16)

Profiling run after the ~50 new handlers, 5 bulk-pattern families, and
the round-1 `normalizeName` memoization landed on main.

## Methodology

- Branch: `dev/perf-round-2` from `origin/main` (`4318132`)
- Workload: `hexdek-tournament --decks data/decks/moxfield_300 --games 500 --pool --seats 4 --pprof --max-turns 60 --game-timeout 60`
- Pool size: 300 decks, 231 unique commanders, all eligible
- Hat: yggdrasil, budget=50, turn-budget=0, noise=0.20
- Workers: `runtime.NumCPU()`
- Wall time: ~38s engine work (excluding AST/corpus load)
- Profiles: `/tmp/hexdek_cpu.prof`, `/tmp/hexdek_allocs.prof`,
  `/tmp/hexdek_heap_baseline.prof`, `/tmp/hexdek_heap_final.prof`

Tournament binary was extended in this branch so `--pprof` writes a
CPU profile + an allocs profile + before/after heap profiles to
`/tmp/hexdek_*.prof` (instead of leaving only the HTTP `:6060`
endpoint).

## Pre-fix CPU profile — top 10 flat

| # | Flat | % | Function | Notes |
|---|------|---|----------|-------|
| 1 | 69.32s | 49.46% | `indexbody` | runtime — bytes-level substring scan |
| 2 | 6.71s | 4.79% | `indexbytebody` | runtime — single-byte index |
| 3 | 4.28s | 3.05% | `runtime.findObject` | GC |
| 4 | 4.05s | 2.89% | `gameengine.(*Permanent).hasType` | hot, but linear over a 1-3 element slice |
| 5 | 3.87s | 2.76% | `internal/stringslite.Index` | std-lib glue around `indexbody` |
| 6 | 3.25s | 2.32% | `runtime.madvise` | system page release |
| 7 | 2.66s | 1.90% | `strings.ToLower` | called per-card per-evaluator pass |
| 8 | 2.38s | 1.70% | `memeqbody` | runtime memcmp |
| 9 | 2.29s | 1.63% | `gameengine.cardIsToken` | inline-only |
| 10 | 1.93s | 1.38% | `runtime.memclrNoHeapPointers` | GC clear |

`strings.Contains` itself was **57% of cumulative CPU**
(`79.28s / 140.14s`). Its top callers were all `hat` evaluator
functions — see allocations section for the culprit.

## Pre-fix top 5 alloc-space hotspots

| # | Bytes | % | Function |
|---|-------|---|----------|
| 1 | 1066 MB | 17.41% | `hat.landProducesColors` |
| 2 | 754 MB | 12.30% | `hat.buildZoneIndex.func1` |
| 3 | 594 MB | 9.70% | `hat.evaluateLine` |
| 4 | 501 MB | 8.18% | `gameengine.BaseCharacteristics` |
| 5 | 301 MB | 4.91% | `strings.(*Builder).grow` (mostly `OracleTextLower` first build + ad-hoc `ToLower`) |

Total allocated: **6125 MB** across the run.

## Triage

### Handler-introduced hot spots?

No. None of the top 10 CPU or top 5 alloc hot spots are new per-card
handler files. The handler dispatch path (`per_card/*`) does not show
up above 0.7s cumulative. The hot regression surface is entirely the
hat evaluator's per-evaluation `strings.Contains` workload, which has
been growing as evaluator dimensions have been added but pre-dates the
recent batch of handlers.

### Obvious-win category breakdown

1. **Repeated `strings.ToLower(TypeLine)` and map-per-call allocation
   in `hat.landProducesColors`** — top alloc hotspot, called per-color
   per-land per-evaluation pass. Each call allocated a
   `map[string]bool` and a fresh lowercased TypeLine. Cached output as
   a `uint8` bitmask on `Card`, plus added `TypeLineLower` cache.
2. **Map-per-piece allocation in `hat.evaluateLine`** — every combo
   piece built a `map[string]bool` from the line's accepted zones.
   Replaced with four local booleans; the accepted-zones slice is
   small (1-3 entries in practice) so a linear scan is faster than a
   hash.
3. **Redundant `strings.ToLower(name)` in basic-land scoring** — five
   independent `ToLower` calls on the same string, once per
   land-in-hand per land-pick evaluation. Hoisted to a single local.

### Not (yet) addressed — recommended for round 3

- `hat.scoreOpponentGraveyard`, `isManaOnlyAbility`, `scoreGraveyard`,
  `scoreStaxLock` — still ~30% of `strings.Contains` cumulative.
  Each runs 4–7 substring scans against `OracleTextLower` per card per
  evaluation. A per-card `HatTextFlags uint16` bitmask
  (lazy-populated, like `OracleTextCache`) covering recurring keywords
  (`flashback`, `escape`, `unearth`, `retrace`, `add {x}`, `graveyard
  + battlefield + return/put`, etc.) would eliminate the inner-loop
  substring work for known patterns. Estimated ~15-20% CPU recovery.
- `hat.buildZoneIndex` — rebuilds a per-seat zone map on every
  evaluation. Caching by `gs.Tick` (or a per-game-state version
  counter) would skip the rebuild when nothing changed since the last
  eval. Allocates ~754 MB per run.
- `hat.cardAliases` — invoked once per card per `buildZoneIndex` call,
  producing a fresh slice each time. Memoize on `Card` (single slice,
  computed once per game lifetime).
- Commander-name match loop in `scoreOpponentGraveyard` allocates a
  fresh lowercased string per commander per call; precompute
  `Seat.CommanderNamesLower []string` once.
- `gameengine.BaseCharacteristics` allocates ~501 MB. Most callers
  request only one or two of the characteristics it returns; an
  inlined "fast path" for the subset hat actually reads in the
  evaluator may be worth profiling.

## Fixes applied in this branch

| Change | File | Win |
|--------|------|-----|
| Add `TypeLineLowerCache`/`typeLineLowerReady` to `Card`; `TypeLineLower()` helper | `internal/gameengine/state.go`, `internal/gameengine/hat.go` | infrastructure |
| Add `ProducedColorsMask uint8`/`ProducedColorsReady bool` cache to `Card` | `internal/gameengine/state.go` | infrastructure |
| Rewrite `landProducesColors` → `landProducesColorsMask` (bitmask, cached on Card), retain map-returning wrapper for non-hot callers | `internal/hat/yggdrasil.go` | -1066 MB alloc, removes top alloc hotspot |
| Refactor land-scoring loop + `fieldColorSources` to consume the bitmask directly (no map lookups) | `internal/hat/yggdrasil.go` | secondary CPU saving |
| Replace `map[string]bool` per combo-piece with 4 local booleans in `evaluateLine` | `internal/hat/combo_sequencer.go` | -19 MB alloc, smaller allocator pressure |
| Hoist `strings.ToLower(name)` for basic-land name check | `internal/hat/yggdrasil.go` | minor CPU, no allocation |

## Measured impact (same workload, re-profile)

| Metric | Before | After | Δ |
|--------|--------|-------|---|
| Total CPU samples | 140.14s | 121.46s | **-13.3%** |
| Total alloc space | 6125 MB | 4746 MB | **-22.5%** |
| `strings.Contains` cumulative | 79.28s (56.6%) | 56.88s (46.8%) | -9.8pp |
| `landProducesColors` alloc | 1066 MB (#1) | absent from top 30 | gone |
| Top alloc hotspot | landProducesColors | buildZoneIndex (672 MB) | shifted |
| Wall time (engine work) | ~38s | ~40s | within run-to-run variance |

The wall-time delta is within noise on a 4-worker Mac NumCPU run
(both runs were ~40s ± 2s across repeated attempts). Sample time
shrinkage and allocator pressure reduction are the load-bearing wins
— this run was run on shared hardware. A clean run on the engine
host (DARKSTAR) is expected to show the throughput gain more
cleanly.

## Test verification

- `go test ./internal/hat/...` — PASS (1.4s)
- `go test ./internal/gameengine/...` — PASS (4.7s)
- `go test ./internal/gameengine/per_card/...` — PASS
- `go test ./internal/tournament/...` — PASS (135s; default 120s
  timeout in CI may need adjustment for `TestELO_*` long runs, but
  this is pre-existing and unrelated to this branch)
- Full 500-game pool tournament completes with no crashes, all 231
  commanders represented
