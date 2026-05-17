# HexDek Fringe Stubs Audit — Round 22

**Date:** 2026-05-17
**Branch:** `dev/fringe-audit-r22`
**Working tree:** `/tmp/hexdek-fringe-r22` (off `origin/main` @ `9bec801`)
**Scope:** `internal/gameengine/keywords_*.go` plus the cross-file mechanic
verification list specified by the round-22 task.

This audit distinguishes three categories:

| Marker | Meaning |
|--------|---------|
| ✅ **impl** | Behavior matches the printed text or is a faithful engine model. |
| ⚠️ **partial** | Checkers/eligibility exist but the actual game effect is missing or trivialized to a log event. Card text would not resolve correctly if hit. |
| ❌ **stub** | Function exists but does nothing meaningful — logs an event and returns. |
| 🚫 **missing** | No keyword detection or apply function in the package. |

The "stub" and "partial" rows are the ones to prioritize. "Missing" rows are
informational — these mechanics simply haven't been on the roadmap yet.

---

## §1. Remaining stub helpers in `keywords_*.go`

These are functions that currently log an event and return without driving
any further game state — found by inspection of every `func` declared in
`internal/gameengine/keywords_*.go` (excluding `_test.go` files).

| § | Function | File:Line | Status | Suggested next step |
|---|----------|-----------|--------|---------------------|
| §702.140 | `ApplyMutatePlaceholder` | `internal/gameengine/keywords_batch6.go:339` | ❌ stub | Either delete (the real `ApplyMutate` at `keywords_batch6.go:264` is fully wired) or repoint legacy callers at `ApplyMutate` and remove. Currently logs `{stub: true}` only. |
| §702.178 | `HasMaxSpeed` | `internal/gameengine/keywords_batch6.go:1119` | ❌ stub | Add `PlayerSpeed`/`AdvanceSpeed` plus a `MaxSpeedActive(gs, seatIdx) bool` (true when speed counter == 4). Wire damage-to-player into speed advancement. Without the counter system, the keyword check is non-actionable. |
| §702.179 | `ApplyStartYourEngines` | `internal/gameengine/keywords_batch6.go:1132` | ⚠️ partial | Animates vehicles but never *un*-animates; the EOT cleanup that clears `start_your_engines` flag is missing. Add an `EndStepClearStartYourEngines` hook + counter+toughness restore. |
| §702.184 | `ApplyStation` (orphan) | `internal/gameengine/keywords_batch6.go:1496` | ❌ stub | Real impl is `ActivateStation` in `keywords_station.go:161`. Delete the orphan or repoint legacy callers — currently logs and exits. |
| §701.4   | `Behold` | `internal/gameengine/keywords_batch6.go:1855` | ❌ stub | Add a `beheld_cards` registry on `GameState` (or seat) so triggered abilities can ask "have you beheld an X." Reveal-from-hand mechanic is currently fire-and-forget. |
| §701.65  | `Airbend` | `internal/gameengine/keywords_batch6.go:2103` | ❌ stub | Set-specific Avatar bending — no effect to apply yet. Wait for printed card data or build a minimal target-bumping helper. |
| §701.66  | `Earthbend` | `internal/gameengine/keywords_batch6.go:2114` | ❌ stub | Same as Airbend — needs target framework before it can do anything. |
| §701.67  | `Waterbend` | `internal/gameengine/keywords_batch6.go:2125` | ❌ stub | Same as Airbend. |
| §701.68  | `Firebend` | `internal/gameengine/keywords_batch6.go:2136` | ❌ stub | Same as Airbend. |
| §702.47  | `ApplySplice` | `internal/gameengine/keywords_combat.go:1126` | ⚠️ partial | Pays the splice cost but does NOT actually append the spliced card's effect to the resolving spell. Should add the spliced effect chain to `item.Effect` before resolution or push a sequenced sub-effect onto the stack. Cards never gain the spliced rider. |
| §702.117 | `CanPaySurge` (no cast helper) | `internal/gameengine/keywords_combat.go:851` | ⚠️ partial | Eligibility predicate only — no `CastWithSurge` analog of `CastBuyback`/`CastFlashback`/`CastWarp`. Spells with surge can't be cast at the surge cost through the engine; only inspected. |
| §702.137 | `CanPaySpectacle` (no cast helper) | `internal/gameengine/keywords_combat.go:840` | ⚠️ partial | Same shape as Surge — eligibility check is there, `CastWithSpectacle` is missing. |
| §702.169 | `CanCastForFreerunning` (no cast helper) | `internal/gameengine/keywords_batch6.go:1072` | ⚠️ partial | Predicate only. Add `CastFreerunning(gs, seatIdx, card, freerunningCost)` mirroring `CastFlashback` so the alt-cost path can be exercised. |
| §702.184 | `ActivateStation` payoff dispatch | `internal/gameengine/keywords_station.go:159` | ⚠️ partial (TODO) | Marked in-source as "per_card payoff dispatch TODO when Aetherdrift cards land". Counter/threshold tracking is real; the actual per-card payoff hooks are not wired. |
| §702.183 | `keywords_job_select.go` self-described as STUB | `internal/gameengine/keywords_job_select.go:4` | ⚠️ partial | Cast-time choice + recording works; "job-specific characteristic dispatch" (riders, gating) is documented as deferred to per_card. |

**Stub count by file:**

- `keywords_batch6.go`: 8 (`ApplyMutatePlaceholder`, `HasMaxSpeed`, `ApplyStartYourEngines`, `ApplyStation`, `Behold`, 4× bend)
- `keywords_combat.go`: 3 (`ApplySplice`, `CanPaySurge`, `CanPaySpectacle`)
- `keywords_station.go`: 1 (`ActivateStation` payoff)
- `keywords_job_select.go`: 1 (per-card rider dispatch)

---

## §2. Fringe mechanic verification — round-22 request list

Status as of `9bec801`. Each row shows the request-list name, CR section,
status, and the canonical entry point.

| Mechanic | CR § | Status | Entry point |
|---|---|---|---|
| Surge | §702.117 | ⚠️ partial | `CanPaySurge` `keywords_combat.go:851` — missing `CastWithSurge` |
| Splice | §702.47 | ⚠️ partial | `ApplySplice` `keywords_combat.go:1126` — pays cost but doesn't append spliced effect |
| Spectacle | §702.137 | ⚠️ partial | `CanPaySpectacle` `keywords_combat.go:840` — missing `CastWithSpectacle` |
| Tempting offer | §702.74 | 🚫 missing | No `Tempting*` symbol in package |
| Will of the Council | §702.93? | 🚫 missing | No symbol; council-vote infra absent |
| Council's dilemma | §702.94? | 🚫 missing | No symbol |
| Conspire | §702.78 | ✅ impl | `ApplyConspire` `keywords_batch4.go:546` |
| Strive | §702.118 | 🚫 missing | No symbol |
| Replicate | §702.56 | ✅ impl | `ApplyReplicate` `keywords_batch3.go:151` (verified — copies pushed with correct §706.10 chars) |
| Forecast | §702.57 | ✅ impl | `ActivateForecast` `keywords_batch5.go:685` |
| Suspend | §702.62 | ✅ impl | `SuspendCard` `keywords_batch.go:226` + `TickSuspend` `keywords_batch.go:263` |
| Champion | §702.72 | ✅ impl | `ApplyChampion` `keywords_batch4.go:932` — exile + LTB return delayed trigger |
| Soulshift | §702.46 | ✅ impl | `CheckSoulshift` `keywords_batch.go:1156` |
| Bushido | §702.45 | ✅ impl | `ApplyBushido` `keywords_p1p2.go:465` + `FireBushidoTriggers` |
| Ninjutsu | §702.49 | ✅ impl | `CheckNinjutsuRefactored` `ninja_sneak.go:126`, wired in `combat.go:256` |
| Ripple | §702.92 | 🚫 missing | No symbol |
| Provoke | §702.40 | ✅ impl | `ApplyProvoke` `keywords_combat.go:717` + `FireProvokeTriggers` |
| Reinforce | §702.76 | ✅ impl | `ActivateReinforce` `keywords_misc.go:882` |
| Recover | §702.60 | ✅ impl | `CheckRecover` `keywords_batch6.go:696` |
| Retrace | §702.81 | ✅ impl | `CanRetrace` + `PayRetraceCost` + `CastWithRetrace` `keywords_misc.go:592` |
| Rebound | §702.88 | ✅ impl | `ApplyRebound` `keywords_batch5.go:815` |
| Wither | §702.80 | ✅ impl | `ApplyWitherDamageToCreature` `keywords_p1p2.go:325` |
| Persist | §702.79 | ✅ impl | `CheckPersist` `keywords_batch.go:67` |
| Modular | §702.43 | ✅ impl | `ApplyModularETB` + `ApplyModularDeath` `keywords_misc.go:1020` |
| Sunburst | §702.42 | ✅ impl | `ApplySunburst` `keywords_misc.go:1167` |
| Annihilator | §702.86 | ✅ impl | `ApplyAnnihilator` `keywords_combat.go:577` — verified, sacrifices N |
| Banding | §702.21 | ✅ impl | `ApplyBandingDamageRedistribution` `keywords_combat.go:234` — full damage-spread algorithm |
| Rampage | §702.23 | ✅ impl | `ApplyRampage` `keywords_combat.go:322` |
| Cumulative upkeep | §702.24 | ✅ impl | `ApplyCumulativeUpkeep` `keywords_batch.go:173` |
| Echo | §702.30 | ✅ impl | `CheckEcho` `keywords_batch.go:443` |
| Phasing | §702.26 | ✅ impl | `PhaseOut` + `PhaseIn` `phases.go:497/547`, indirect attached-perm phasing per §702.26d |
| Fading | §702.32 | ✅ impl | `ApplyFading` `keywords_batch.go:405` |
| Vanishing | §702.63 | ✅ impl | `ApplyVanishing` `keywords_batch.go:367` |
| Bestow | §702.103 | ✅ impl | `CastWithBestow` `keywords_p1p2.go:807`, `BestowFalloff` + `CheckBestowFalloffs` |
| Crew | §702.122 | ✅ impl | `CrewVehicle` `keywords_p1p2.go:316`, `UncrewVehiclesAtEOT` `keywords_p1p2.go:392` |
| Flagstones-class lands | — | 🚫 missing (no per-card handler for Flagstones of Trokair specifically) | Sibling pattern `internal/gameengine/per_card/fetchlands.go` is the closest analogue — replace with a "search on destroy" template if/when Flagstones is needed |

**Summary of the 36-row list:**

- ✅ impl: 24
- ⚠️ partial: 3 (Surge, Splice, Spectacle)
- 🚫 missing: 6 (Tempting offer, Will of the Council, Council's dilemma, Strive, Ripple, Flagstones-class)
- (Plus 3 partials surfaced in §1 that aren't on the request list — Freerunning cast, Station payoff, Job Select rider — relevant follow-ups)

---

## §3. Recommended priority order

If round 23 wants to keep moving the partial-impl tail, the lowest-effort,
highest-leverage knockouts are:

1. **Surge / Spectacle / Freerunning `CastWith*` helpers** — copy-paste of
   `CastBuyback` / `CastFlashback` with the appropriate eligibility
   predicate. Each is < 80 LOC + a 4-test file.
2. **`ApplyMutatePlaceholder` delete-or-repoint** — pure dead code cleanup.
   Audit callers first (no production callers expected since `ApplyMutate`
   already covers the live path).
3. **`ApplyStation` orphan delete** — same cleanup pattern.
4. **`ApplySplice` real effect-merging** — the only fringe stub on the
   request list where the gap is *behavioral* rather than missing. Append
   spliced effect to `item.Effect` chain at cast time; would require
   touching the effect-list type, so larger than the cast-helper trio.
5. **§702.178 `MaxSpeed` + §702.179 `StartYourEngines` EOT cleanup** —
   couple Aetherdrift Speed mechanic to a real player-side counter system.
   Touches phase cleanup; medium effort.
6. **Beholding registry** — small but unblocks every "have you beheld X"
   trigger.
7. **Avatar bending (Airbend/Earthbend/Waterbend/Firebend)** — defer until
   the set's card data lands; nothing useful to wire without targets.
8. **Tempting offer / Will of the Council / Council's dilemma / Strive /
   Ripple / Flagstones** — none have current corpus pressure. Pick up
   only if a specific commander or deck pulls them in.

---

## §4. Method notes

- The "stub" classification flagged any function whose body is essentially
  `LogEvent + return` with no state writes. Functions that move cards,
  pay mana, or modify permanents are counted as ✅ even when their effect
  is simplified (e.g. annihilator picks smallest creature as a policy).
- The mechanic-verification list cross-checked each name against
  `grep -rn "func.*<Name>" internal/gameengine/`; misses were re-grepped
  case-insensitively before being marked 🚫 missing.
- This audit operates against `origin/main` at commit `9bec801` — which
  includes the round-21 stubs closeout (mobilize, harmonize, solved,
  visit, mayhem) and the round-22a `dev/stubs-tail` merge (Tiered,
  Infinity, Space Sculptor).
