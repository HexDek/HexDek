# HexDek Engine Performance Profile — 2026-05-16

Profile target: `hexdek-tournament --decks data/decks/moxfield_300 --pool --games 500 --seats 4 --seed 42 --pprof`
on darwin/amd64 (Apple Silicon emulating amd64 via Rosetta), Go 1.25.0.

- 500 games of the 300-deck Moxfield pool (4 seats each, no Freya intel —
  Hat falls back to generic weights), seed 42 for reproducibility.
- Baseline wall-clock (pre-fix, no-pprof slot of the same run): **11.5 g/s, 43.5 s total**.
  CLAUDE.md banner ("~25 g/s on 4-seat") corresponds to the Freya-cached fast pool;
  the unfreyad pool exercises the full eval loop and is a stricter perf target.
- CPU profile: 30 s wall-clock sample captured via `/debug/pprof/profile?seconds=30`
  during the 43 s run. Total CPU samples: **78.84 s** across ~2.6 cores
  (workers default to NumCPU, 4 here).
- Alloc profile: cumulative `/debug/pprof/allocs` captured mid-run. Total
  alloc_objects: **36.77 M** / alloc_space: **1.84 GB** over ~25 s of game play.

---

## Top 10 CPU hotspots (flat %)

| Rank | Symbol                                                           | Flat    | Flat % | Cum    | Cum %  |
|------|------------------------------------------------------------------|---------|--------|--------|--------|
| 1    | `runtime.indexbody` *(internal/stringslite.Index inner loop)*    | 40.09 s | 50.85% | 40.09 s | 50.85% |
| 2    | `runtime.indexbytebody`                                          |  3.30 s |  4.19% |  3.30 s |  4.19% |
| 3    | `gameengine.(*Permanent).hasType`                                |  2.58 s |  3.27% |  4.05 s |  5.14% |
| 4    | `internal/stringslite.Index`                                     |  2.20 s |  2.79% | 46.09 s | 58.46% |
| 5    | `runtime.findObject` *(GC scan)*                                 |  2.01 s |  2.55% |  2.15 s |  2.73% |
| 6    | `strings.ToLower`                                                |  1.57 s |  1.99% |  2.80 s |  3.55% |
| 7    | `gameengine.cardIsToken`                                         |  1.55 s |  1.97% |  1.58 s |  2.00% |
| 8    | `runtime.memeqbody`                                              |  1.55 s |  1.97% |  1.55 s |  1.97% |
| 9    | `runtime.madvise` *(memory release)*                             |  1.28 s |  1.62% |  1.28 s |  1.62% |
| 10   | `runtime.memclrNoHeapPointers`                                   |  0.96 s |  1.22% |  0.96 s |  1.22% |

### What this actually says

`strings.Contains` consumes **48.68% of CPU** (the indexbody + Index nodes
are its inner machinery). Tracing callers via `pprof -peek`:

| Caller                                          | strings.Contains time | % of strings.Contains |
|-------------------------------------------------|----------------------:|----------------------:|
| `hat.scoreOpponentGraveyard`                    |                9.88 s |                25.74% |
| `hat.isManaOnlyAbility`                         |                6.42 s |                16.73% |
| `hat.scoreGraveyard`                            |                4.78 s |                12.45% |
| `gameengine.DFCCardMatchesName`                 |                3.00 s |                 7.82% |
| `hat.scoreStaxLock`                             |                2.85 s |                 7.43% |
| `hat.scoreDrainEngine`                          |                2.08 s |                 5.42% |
| `hat.scoreThreat`                               |                1.66 s |                 4.33% |
| 25+ other Hat scorers                           |                7.71 s |                20.08% |

Every scorer is doing the same shape of thing: take the cached lowercased
oracle text and run 3-10 `strings.Contains(ot, "<keyword>")` calls per card
per evaluation. The cached lowercase string elides the `ToLower` cost
(only 1.99% flat), but the substring searches themselves are the bottleneck.

`gameengine.(*Permanent).hasType` (#3, 3.27% flat) is the type-string
comparison loop hit constantly by combat, SBA, and target filtering.

`gameengine.IsCommanderCard` (cumulative 6.31 s, 8% of CPU, called from
`hat.detectPhase` 90.81% of the time) burns through `DFCCardMatchesName`
+ `DFCFrontFaceName` doing string compares on the same commander
identifier inside MCTS-style eval loops.

---

## Top 5 allocation hotspots (alloc_objects)

| Rank | Symbol                                                | Objects     | %      | Cum %  |
|------|-------------------------------------------------------|-------------|--------|--------|
| 1    | `strings.(*Builder).grow`                             | 15,362,854  | 41.78% | 41.78% |
| 2    | `reflect.New` *(json decode reflect path)*            |  2,031,637  |  5.52% | 47.30% |
| 3    | `gameengine.BaseCharacteristics`                      |  1,844,813  |  5.02% | 52.32% |
| 4    | `encoding/json.(*decodeState).object`                 |  1,714,875  |  4.66% | 56.98% |
| 5    | `encoding/json.Unmarshal`                             |  1,434,631  |  3.90% | 60.88% |

**`strings.Builder.grow` callers (alloc_objects):**

| Caller                                | Allocs     | %      |
|---------------------------------------|------------|--------|
| `per_card.normalizeName` *(via Grow)* |  6,818,533 | 44.38% |
| `strings.ToLower` *(via Grow)*        |  8,500,631 | 55.33% |

`per_card.normalizeName` is called from `fireTrigger` on **every
battlefield permanent of every seat for every trigger event**
(`registry.go:317`). The card-name space is tiny (~30K oracle entries),
but the function allocates a fresh `strings.Builder`-backed string per
call. Across a 500-game run that's 14M+ identical computations.

`encoding/json.Unmarshal` (#5, top-5 by space at 197 MB / 10.71%) is
hit from `astload` deserializing per-card AST fragments stored as
`RawMessage`. Not in the per-game hot path — pays once at deck-load —
so it dominates baseline allocation but doesn't scale with throughput.

---

## Cheap-win recommendations

### ✅ Applied in this branch — `per_card.normalizeName` memoization

Wrapped the function with a process-wide `sync.Map` cache (registry.go).
Card names are bounded and `normalizeName` is a pure function, so the
cache hit rate is ~100% after the first turn.

**Measured impact (post-fix re-run, same seed/decks/pool):**

| Metric                      | Before        | After         | Δ        |
|-----------------------------|---------------|---------------|----------|
| Total alloc_objects         | 36,773,335    | 18,868,982    | **−48.7%** |
| `strings.Builder.grow`      | 15,362,854    |  1,119,579    | **−92.7%** |
| `per_card.normalizeName`    | in top-3      | not in top-10 | gone     |

Wall-clock throughput improvement was not cleanly measurable on this run
(multiple concurrent tournament + Go-compile workers were sharing CPU);
GC pressure relief alone should give +5-15% on an unloaded box, more on
trigger-heavy decks.

### 🟡 High-value, slightly more work — Hat scorer trait flags

`hat.scoreOpponentGraveyard`, `scoreGraveyard`, `scoreStaxLock`,
`scoreDrainEngine`, `scoreThreat`, `isManaOnlyAbility`, and ~25 sibling
scorers collectively burn **48.68% of CPU** running `strings.Contains`
on the cached oracle text. Each card has a finite set of "interesting"
phrases (flashback, escape, unearth, retrace, "destroy", "exile", ":
add {", etc.).

**Suggested fix:** add a `CardTraitFlags uint64` field to `Card`,
populated lazily alongside `OracleTextCache` in `OracleTextLower`. The
~30 hot phrases each get a bit. Each scorer flips from
`strings.Contains(ot, "foo") && strings.Contains(ot, "bar")` to
`(card.CardTraitFlags & (TraitFoo|TraitBar)) == (TraitFoo|TraitBar)`.

Estimated impact: bringing strings.Contains from 48.68% to <5% of CPU
would buy ~1.5-1.8× tournament throughput. Mechanical refactor — touch
every scorer once, no behavior change because the boolean output is
identical to the substring search.

### 🟡 `IsCommanderCard` cache

`IsCommanderCard` (cumulative 8% CPU) is called from `hat.detectPhase`
on every position eval and re-runs DFC name matching every time. The
"is this perm my commander" answer is fixed for the lifetime of a
permanent on the battlefield.

**Suggested fix:** stamp `Permanent.Flags["is_commander"] = 1` on ETB
for any permanent that satisfies the commander-name match, and have
`IsCommanderCard` short-circuit on the flag. Estimated impact: ~5-8% CPU
recovered.

### 🟢 `gameengine.(*Permanent).hasType` (#3 hot, 3.27% flat)

Inner loop is `for _, t := range p.Card.Types { if t == name { return true } }`.
The number of types per permanent is small (3-6), but the function is
called millions of times across combat, SBA, and target filtering.

**Suggested fix:** at card-load time precompute a `TypeBits uint32` (each
of CR §300.1's card types = one bit) and a parallel `SubtypeBits`. Most
of the hot calls are `hasType("creature")`, `hasType("land")`, etc. — a
single bit-test replaces the string loop. The deferred-mutation path
(Mycosynth Lattice, layer-effect type-stripping) already rebuilds
characteristics through `BaseCharacteristics`, so the bitmap can be
recomputed there.

Estimated impact: ~3% CPU recovered, plus removes `hasType` from the
list of GC-walked string allocators.

### 🟢 `gameengine.BaseCharacteristics` (#3 alloc_space, 192 MB / 1.85M objs)

Returns a fresh `Characteristics` struct including freshly-allocated
slices for Types/Subtypes/Colors on every call. Most callers only need
read access.

**Suggested fix:** cache the layered characteristics on `Permanent`
keyed by a state version counter (already exists for layer rebuilds);
return a shared `*Characteristics` instead of a fresh value. Estimated
impact: ~10% drop in alloc_space, modest CPU win from reduced GC.

---

## Lower-priority observations

- `runtime.findObject` + `runtime.scanobject` + `runtime.greyobject` +
  `runtime.madvise` collectively account for ~5.5% CPU — GC overhead
  consistent with the alloc rate. The applied `normalizeName` cache
  + the trait-bits refactor would each cut this further.
- `gameengine.partitionTypes` (1.6% alloc_objects, called from layer
  effects) appends to fresh slices per call — pool-able with a
  per-game arena if it ever surfaces in CPU profile.
- Goroutine count was 5 (1 main + 4 game workers) at sample time — no
  goroutine leak indication.

## Reproduction

```bash
# Build instrumented binary
go build -o /tmp/hexdek-tournament ./cmd/hexdek-tournament

# Start a 500-game pool tournament with pprof
/tmp/hexdek-tournament --decks data/decks/moxfield_300 --pool --games 500 \
  --seats 4 --seed 42 --pprof --report /tmp/perf_report.md &

# Capture profiles while the run is in progress
curl -o /tmp/cpu.prof 'http://localhost:6060/debug/pprof/profile?seconds=30'
curl -o /tmp/allocs.prof 'http://localhost:6060/debug/pprof/allocs'
curl -o /tmp/heap.prof 'http://localhost:6060/debug/pprof/heap'

# Analyze
go tool pprof -top -nodecount=20 /tmp/hexdek-tournament /tmp/cpu.prof
go tool pprof -peek 'strings.Contains' /tmp/hexdek-tournament /tmp/cpu.prof
go tool pprof -top -sample_index=alloc_objects /tmp/hexdek-tournament /tmp/allocs.prof
```
