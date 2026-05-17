# May 11 Crash Cluster — Forensic Audit

## Headline

The May 11 grinder produced **324 panics on `data/muninn/crashes.json`** (timestamps `2026-05-11T01:05:40Z` ... onward), all carrying the same `runtime error: invalid memory address or nil pointer dereference` signature. The flood is one root cause hit 321 times in the same Loki/grinder run, plus 3 carry-over `ExtractPivot` panics from the previous binary.

Both root causes were *symptomatically* patched in `main` before May 11 ran, but the deployed binary on `DARKSTAR` was still the older artifact (no commits landed May 10–11 — the binary in use was built from the May 4–May 8 era of `main`). The patches that already shipped (`b348f4a` 2026-05-08 added a `perm != nil` guard around the LTB `FireCardTrigger`; `199837e` 2026-05-05 added a `winnerSeat < 0` early-return) close the panic, but the API misuse that produced the nil `perm` in the first place is still in the tree. This audit fixes that.

## Distribution

| Date | Count | Signature |
|------|------:|-----------|
| 2026-05-05 | 466 | `hat.ExtractPivot` at `tesla.go:55` — `curr[winnerSeat]` deref with `winnerSeat = -1` |
| 2026-05-11 | 321 | `per_card.abdelAdrianETB` → `moveCardBetweenZones` → `MoveCard` → `FireZoneChangeTriggers` (nil `perm`) |
| 2026-05-11 | 3 | residual `ExtractPivot` (same as May 5) |
| 2026-05-12 | 1 | residual |
| **total** | **791** | |

The `decks=[]` / `turn_count=0` in every record is an artifact of `tournament.runOneGameSafe`'s `recover()` block: it captures the panic but does not enrich the record with mid-game state, so the Muninn `(empty deck list, zero turns)` is the *normal* shape of a grinder-safe panic, not a separate bug.

## Root cause #1 — Abdel Adrian (the May 11 flood)

`internal/gameengine/per_card/abdel_adrian.go:64-65` previously did:

```go
removePermanent(gs, p)
moveCardBetweenZones(gs, seat, card, "battlefield", "exile", "abdel_adrian")
```

This violates the contract in `internal/gameengine/zone_move.go:51-55`:

> Callers moving FROM the battlefield should not use MoveCard directly: battlefield exits have their own Permanent-lifecycle semantics … MoveCard intentionally does nothing on `fromZone == "battlefield"`.

`MoveCard` for `fromZone="battlefield"` reaches `FireZoneChangeTriggers(gs, nil, card, "battlefield", dest)` — `perm` is `nil`. In the binary deployed on May 11 (built off commit `6920224` 2026-05-01 era), `zone_change.go:379` was:

```go
if fromZone == "battlefield" {
    FireCardTrigger(gs, "permanent_ltb", map[string]interface{}{
        "perm":            perm,
        "card":            card,
        "controller_seat": perm.Controller,   // ← nil deref
        ...
    })
}
```

Commit `b348f4a` (2026-05-08) hardened that branch to `if fromZone == "battlefield" && perm != nil`, which is why the panic does not reproduce against `main` today. But the **API contract is still violated** — `abdelAdrianETB` was bypassing:

- §614 `would_be_exiled` replacement chain
- §903.9b commander-zone redirect
- `UnregisterReplacementsForPermanent` / `UnregisterContinuousEffectsForPermanent`
- `detachAll` (auras attached to the exiled permanent stay glued to a phantom)
- The proper `LogEvent("exile")` audit record

So the panic was the loud symptom of a quieter correctness problem.

### Fix

Route the exile through `gameengine.ExilePermanent(gs, p, perm)` — the canonical battlefield-exit API. This:

1. Fires the would-be-exiled replacement chain (Rest in Peace will redirect, indestructible passes through since exile isn't destroy, etc.).
2. Calls `gs.removePermanent`, the full bookkeeping version (not the shallow per_card helper).
3. Unregisters replacements + continuous effects from the permanent.
4. Detaches auras / equipment.
5. Fires `FireZoneChangeTriggers` with the *real* perm so observer / self / LTB triggers (Blood Artist, Grave Pact, Karmic Guide ETB-target chains, the permanent's own LTB) all see a sensible source.
6. Returns `false` if the exile was cancelled (replacement effect, etc.), letting us count actually-exiled cards instead of token-creating against an inflated count.

Token count is now driven by `exiledCount` (cards that survived the §614 chain), not `len(picks)`.

## Root cause #2 — ExtractPivot (the May 5 flood, included for completeness)

`internal/hat/tesla.go:32-35` in the May-4 era was:

```go
func ExtractPivot(evalHistory map[int][]float64, winnerSeat int, gameTurns int) CausalPivot {
    if len(evalHistory) < 2 {
        return CausalPivot{Turn: gameTurns / 2, WinnerSeat: winnerSeat}
    }
```

`runOneGameFast` calls `ExtractPivot(collector.History(), winner, gs.Turn)`. `winner` is `-1` when the game terminated with no surviving seat (mutual annihilation in a 1-on-1 → 0-on-1 cascade, a stall hit `showmatchMaxTurn` with the conviction system having conceded everyone, etc.). With `winnerSeat = -1`, the guard `winnerSeat >= len(prev)` is false (`-1 >= 0`) so the code reaches `curr[winnerSeat]` → bounds panic surfaced by the runtime as the nil-pointer message via slice-header pun on Apple silicon.

**Already fixed in commit `199837e` (2026-05-05)** by tightening the early return to `if len(evalHistory) < 2 || winnerSeat < 0`. No further action required — this is documented here purely so the May 5 entries in `crashes.json` don't get re-investigated.

## Verification

```
go build ./...
go test ./internal/gameengine/... -count=1 -timeout 300s
```

Both clean.

## Residual API-misuse audit

The same `removePermanent(gs, p) → moveCardBetweenZones(gs, …, "battlefield", "exile", …)` anti-pattern still exists in:

| File | Operation | Correct API |
|------|-----------|-------------|
| `etrata_the_silencer.go:57` | exile target creature (defender hit) | `ExilePermanent` |
| `etrata_the_silencer.go:82` | shuffle Etrata into library | needs a dedicated battlefield→library helper |
| `zabaz_the_glimmerwasp.go:67` | destroy permanent | `DestroyPermanent` |
| `zimone_and_dina.go:135` | sacrifice creature | `SacrificePermanent` / `sacrificePermanentImpl` |
| `gen_bilbo_birthday_celebrant.go:112` | self-exile | `ExilePermanent` |
| `thassa_deep_dwelling.go:71-72` | blink (exile + return) | `ExilePermanent` then a return-from-exile helper |

These do **not** crash on `main` (the May 8 `perm != nil` guards across `FireZoneChangeTriggers` close every nil-deref path), but they all skip the replacement chain, aura detach, and bookkeeping. They are correctness — not stability — bugs. Listed here so the next handler-coverage pass can sweep them in one PR rather than discovering them one Goldilocks/Loki report at a time.

## Open follow-up

- Consider adding a `gs.LogEvent("internal_warning")` when `MoveCard` is invoked with `fromZone == "battlefield"` so future misuse gets caught at runtime instead of silently bypassing battlefield-exit semantics. Not included in this PR to keep the blast radius small.
