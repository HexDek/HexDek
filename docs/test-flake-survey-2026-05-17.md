# Test Flake Survey — 2026-05-17

Branch: `dev/test-flake-survey` (from `main` @ `0467f40`)
Method: `go test ./... -count=1 -timeout 600s` × 3 consecutive runs, no parallelism between runs.

## Headline

Two failures appeared across three runs. **One is a true flake (1/3 fail), one is a deterministic regression (3/3 fail).** Everything else — 30 packages, ~110 tests — passed in all three runs.

| Test | Package | Run 1 | Run 2 | Run 3 | Verdict |
|------|---------|-------|-------|-------|---------|
| `TestStitchEndpoint` | `internal/pincer` | FAIL | ok | ok | **Flake** — async-vs-sync DB race |
| `TestLoadFullCorpus` | `internal/astload` | FAIL | FAIL | FAIL | **Not a flake** — stale assertion vs migrated corpus |

Suite wall-clock dominated by `internal/tournament`: 280.6s / 169.7s / 160.9s. The 281s spike on run 1 is cold-start cost (corpus parse + AST load not yet cached); subsequent runs reuse OS page cache. Not a flake — just I/O warming.

## Flake: `pincer.TestStitchEndpoint`

### Symptom (run 1)
```
--- FAIL: TestStitchEndpoint (0.05s)
    pincer_test.go:218: query: sql: Scan error on column index 0,
        name "user_id": converting NULL to string is unsupported
```

The row exists (otherwise we'd get `sql.ErrNoRows`), but `user_id` is NULL — meaning the stitch UPDATE matched zero rows.

### Root cause: race between async INSERT and sync UPDATE

`internal/pincer/pincer.go:151` fires the initial-session insert in a goroutine:

```go
// Best-effort upsert; never block the request on a tracking error.
go t.touchSession(sid, fresh, r.UserAgent(), clientIPHash(r))
```

`touchSession(fresh=true)` runs `INSERT OR IGNORE INTO pincer_session (id, user_id, ...) VALUES (?, NULL, ...)`.

The test's interleaving:

1. **Request 1** (`GET /`): middleware mints a cookie and schedules goroutine **A** (insert with `user_id=NULL`). Handler returns.
2. **Request 2** (`POST /api/pincer/stitch`): handler calls `Tracker.Stitch()` synchronously — `BEGIN; UPDATE pincer_session SET user_id='user-99' WHERE id=? AND (user_id IS NULL OR user_id=?); ...; COMMIT`.
   - If goroutine A has **not** run yet, the UPDATE matches **zero rows**. The follow-on `INSERT OR IGNORE INTO pincer_stitch ...` and `INSERT INTO pincer_event ...` still succeed (they don't depend on the session row).
3. **Goroutine A** finally runs `INSERT OR IGNORE INTO pincer_session (id, user_id, ...) VALUES (?, NULL, ...)` — row didn't exist, so it inserts with `user_id=NULL`.
4. Test's `SELECT user_id FROM pincer_session WHERE id=?` returns NULL → `Scan` into `var stored string` fails.

When goroutine A wins the schedule before request 2 hits Stitch (the common case under load on a modern Mac), the row exists, the UPDATE succeeds, and the test passes. That's why run 2 and run 3 passed.

### Suggested fixes (not applied — survey only)

Pick one — listed in order of preference:

1. **Make the initial insert synchronous for fresh sessions only.** The "never block" justification at line 150 is correct for the *recurring* `last_seen_at` update and the page-view events, but the very first insert is load-bearing for any subsequent Stitch. One extra SQLite insert on first hit is cheap.
2. **Test-side fix:** have the test wait for the insert. Adds a sync point the production code doesn't have — masks the prod race.
3. **Change Stitch to upsert** (`INSERT ... ON CONFLICT(id) DO UPDATE SET user_id=...`) so it doesn't depend on the row existing. Fixes the symptom but the audit-trail INSERT into `pincer_stitch` is still racy against the FK pattern if you ever add one.

Option 1 is the right fix: it removes the actual race in prod, not just the test's exposure of it. The same hazard exists for any real user who logs in within a few microseconds of their first page load on a slow disk.

### Reproduction
```bash
go test ./internal/pincer/ -run TestStitchEndpoint -count=200
```
Expect a small percentage of failures (single-digit %) on a quiet machine; higher under CPU pressure.

## Deterministic failure: `astload.TestLoadFullCorpus`

### Symptom (identical across all 3 runs)
```
astload_test.go:192: Lightning Bolt: expected spell_effect modification,
    got &{ModKind:typed_spell_effect Args:[...] Layer:}
astload_test.go:192: Counterspell: expected spell_effect modification,
    got &{ModKind:typed_spell_effect Args:[...] Layer:}
astload_test.go:192: Thassa's Oracle: ability 1 modification not spell_effect:
    got &{ModKind:typed_spell_effect Args:[...] Layer:}
```

### Root cause: corpus migrated `spell_effect` → `typed_spell_effect`, test never updated

The on-disk dataset (`data/rules/ast_dataset.jsonl`, 47.5 MiB, 31,963 cards) has:

```
$ grep -c "typed_spell_effect" data/rules/ast_dataset.jsonl
7877
$ grep -c '"kind":"spell_effect"' data/rules/ast_dataset.jsonl
0
```

Every `spell_effect` got renamed to `typed_spell_effect` at corpus-build time (Thor side). The loader at `internal/astload/loader.go:244` still produces literal `spell_effect` for its *synthetic Static wrappers*, but the bulk of the corpus flows through `decodeModification` which passes the `kind` field through verbatim — so cards come out as `typed_spell_effect`.

The test's golden assertions at `internal/astload/astload_test.go:220, 235, 261, 375` still compare against `"spell_effect"`.

Engine side appears to already be partially aware: `tests/golden/*.json` use the string `static:typed_spell_effect:_` (20+ files). But the live engine code at `internal/gameengine/stack.go:805`, `keywords_levelup.go:275`, and `resolve_helpers.go:2320` still switches on `"spell_effect"` — so this isn't *just* a test issue, it's a name-migration that didn't finish.

### Suggested fix (not applied — survey only)

Decide which name is canonical, then settle it consistently:

- **If `typed_spell_effect` is the new canonical name** (the corpus has migrated): update `astload_test.go` assertions and the three engine switch-sites (`stack.go`, `keywords_levelup.go`, `resolve_helpers.go`) plus the synthetic-wrapper kind in `loader.go:244`. ~5-line patch in each file.
- **If `spell_effect` should remain canonical**: add a kind-rename in `decodeModification` so `typed_spell_effect` → `spell_effect` on load, and re-run Thor on the corpus so disk gets back in sync.

Option 1 looks more likely to match intent — the goldens have already migrated, suggesting the rename was deliberate but engine-side cleanup stalled. Worth checking the Thor commit that introduced `typed_spell_effect` for the original motivation before picking.

This is not strictly an "engine flake" so it's somewhat out of scope for this survey, but the test fails every run and so was visible in the data.

## Methodology notes

- Single-machine, single-process runs. No `-race`. No `-parallel`.
- All three runs used the same `data/rules/` artifacts; no shuffling, no regenerated corpus between runs.
- Each run completed in ~6 minutes wall-clock. Total survey time ~20 minutes including overhead.
- Raw logs preserved at `/tmp/flake-survey/run{1,2,3}.log` during the session (not committed — large and machine-specific).

## What this survey did **not** cover

- **`-race` detector runs.** Worth a follow-up: `go test ./... -race -count=1` would likely surface the pincer race without needing the rare NULL scan to trigger.
- **`-shuffle` for intra-package ordering flakes.** Go 1.17+ flag, would catch tests that rely on file-declaration order.
- **Cross-test pollution.** The 2026-05-17 issue-log entry (Sai handler stripped by `Reset()` test pollution) suggests this class of bug is real here. A targeted `go test -run X -count=1` survey of the per-card suite would be a sensible next pass.
- **Long-tail rare flakes.** Three iterations catches anything with a per-run failure rate above ~30%. Lower-frequency flakes need 50–200 runs to surface reliably; cheap to run overnight if there's appetite.
