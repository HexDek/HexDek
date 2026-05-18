# Stack & Priority Audit — R37

**Date:** 2026-05-17
**Scope:** HexDek engine implementation of CR §117 (priority), §405 (stack), §608.2 (resolution / illegal targets), §603 (triggered abilities), §702.61 (split second).
**Files read in depth:** `internal/gameengine/stack.go` (1751 lines), `activation.go` (623), `triggers.go` (112), `trigger_stack_bridge.go` (92), `targets.go` (822), `phases.go` (937).

---

## (a) §117 Priority — **Correct**

- Active player receives priority first after each cast/activation/resolution via APNAP ordering in `apnapOrder(gs)` — `stack.go:591`.
- Opponents pass clockwise in seat order; the priority round walks the APNAP slice and short-circuits when all players have passed in succession.
- Priority window re-opens after each stack resolution per §117.5 — `stack.go:105–109` (the priority/resolve loop in the main game-loop driver).
- All-pass-on-empty-stack correctly transitions to the next step rather than infinite-looping — `stack.go:612–615`.
- Loop protection: 500-iteration cap on `PriorityRound` prevents pathological cycles without changing semantics — `stack.go:94–103`.
- Test coverage: `apnap_upkeep_test.go:114–237` exercises 4-seat APNAP under upkeep triggers.

## (b) §405.5 Stack Resolution — **Correct**

- LIFO discipline: only `gs.Stack[len-1]` ever resolves — `stack.go:888–889`.
- A priority round opens after each resolution before the next item is considered — `stack.go:108`.
- Empty-stack-with-all-pass transitions to next phase/step cleanly — `stack.go:612–615`.
- Counter-spell handling marks items `Countered` rather than splicing the slice, preserving stack-trace ordering — confirmed via `counter_resolve.go` integration.

## (c) §608.2 Spell/Ability Resolution — **Correct**

- Target legality is re-checked at resolution time, not at cast — `stack.go:937–957` via `CheckTargetLegality`.
- All-illegal-targets case: spell is removed from the stack and moved to its owner's graveyard (CR §608.2b "won't resolve") — `stack.go:949–954`.
- Partial-legal case: spell resolves applying only to surviving legal targets — `stack.go:955–957`.
- Distinction between "countered by rules" (all targets illegal) vs "countered by spell" is preserved for triggers that key on it.
- Caveat: `item.Targets = legalTargets` (stack.go:956) holds pointer references; safe in practice because of GC + single-threaded resolution, but worth a regression test for the destroy-during-resolution edge case.

## (d) §603 Triggered Abilities — **Partial** ⚠️

- Triggered abilities go on the stack rather than resolving immediately — correct in shape (`PushTriggeredAbility`, `stack.go:172–229`).
- APNAP-then-controller-order ordering helper exists: `OrderTriggersAPNAP` (`triggers.go:34–79`) and the batched entry point `PushSimultaneousTriggers` (`triggers.go:88–112`) implementing §603.3b.
- **Gap:** the batched entry point is **only called from tests and from `phases.go` upkeep flow**. Grep confirms zero non-test production callers beyond upkeep:
  ```
  apnap_upkeep_test.go:133, 191, 246, 312
  triggers_test.go:254
  ```
- All other production trigger fire paths (`FireCardTrigger`, `FireZoneChangeTriggers`, the ETB cascade at `stack.go:1349–1642`, per-card handlers via `trigger_stack_bridge.go:27–92`) push one trigger at a time through `PushTriggeredAbility`, which **opens a priority round and resolves the new top of stack inline** (`stack.go:222–227`):
  ```go
  PriorityRound(gs)
  if len(gs.Stack) > 0 && gs.Stack[len(gs.Stack)-1] == item {
      ResolveStackTop(gs)
  }
  ```
- This is depth-first per arrival order, not breadth-first §603.3b ordering. When N ETBs fire from one event (Cyclonic Rift wipe + persist returns; mass-blink), the first trigger pushed runs to full resolution — including cascading triggers — before the second trigger is even placed on the stack. The controller therefore cannot choose the order of simultaneous triggers, and APNAP between players is not applied to the batch.
- Severity: **Medium**. Most decks rarely see >2 simultaneous-controller triggers at once, but multi-ETB and "dies" pile-ups (Karmic Guide, persist combos, Sun-Titan stacks) will exhibit non-rules-legal ordering.

## (e) §702.61 Split Second — **Correct**

- `SplitSecondActive(gs)` walks the stack for an uncountered split-second item — `stack.go:1474–1532`.
- Cast path rejects non-mana spells while active — `stack.go:277–279` (CastError with `Reason: "split_second"`).
- Activation path rejects non-mana, non-special abilities via `StaxCheck` — `activation.go:193–195`.
- Mana abilities bypass the check (they resolve inline, never hit the priority gate) — consistent with §605 / §702.61b.
- Triggered abilities continue to fire and go on the stack — also rules-correct.
- Priority round short-circuits responder windows when split-second is active — `stack.go:619–627`.
- Countered split-second items are correctly ignored by the gate (line 1489) so they can't lock the stack forever.

---

## Bugs found

### BUG-37-A — Simultaneous triggers resolve inline instead of batching (Medium)
**Where:** `stack.go:222–227` (`PushTriggeredAbility`) combined with one-at-a-time callers in `stack.go:1330–1355`, `trigger_stack_bridge.go:88–91`, `per_card_hooks.go`, and every `FireCardTrigger`/`FireZoneChangeTriggers` site.
**Why it's wrong:** §603.3 says triggers wait until "the next time a player would receive priority"; §603.3b requires they then be placed on the stack as a *batch*, ordered APNAP-between-players + controller-choice-within-player. Current code pushes trigger 1, resolves it (cascading further triggers), then pushes trigger 2.
**Fix shape:** route all multi-trigger fire sites through `PushSimultaneousTriggers` (which already exists, documented, and tested in isolation). Either (1) refactor `FireCardTrigger`/`FireZoneChangeTriggers` to collect a slice and call the batched API, or (2) keep the per-call API but defer pushes into a pending-triggers queue drained at the next priority window. The infrastructure is present — this is a wiring problem, not a design problem.

### BUG-37-B — Potential pointer aliasing in `item.Targets` after partial-legal trim (Low)
**Where:** `stack.go:956`.
**Why it's a concern:** if a permanent is destroyed mid-resolution and its struct address is later reused, a stale `Targets` pointer could mis-target. Not currently exploitable due to GC lifetime, but a future allocator change would expose it.
**Fix shape:** swap pointer targets for opaque IDs (e.g. `PermanentID`) and resolve to pointers at use sites. Out of scope for R37.

---

## Proposed race-condition tests

1. **Multi-ETB §603.3b ordering** — board has two creatures with ETB-draw triggers (one per opponent's control); a third effect makes both ETB simultaneously. Assert the stack contains *both* items in APNAP order before either resolves. Today this test would fail because trigger 1 resolves before trigger 2 is pushed.

2. **Controller-choice of simultaneous own triggers** — single controller, two ETBs that fire on the same event but interact (e.g., one draws, one demands a discard count). Assert controller is offered an ordering decision and the chosen order is reflected in resolution sequence.

3. **Partial-legal target survives destroy-during-resolution** — three-target damage spell; one target is destroyed by the first damage event. Assert the remaining two targets still take damage and stack trace logs one `target_illegal` + two `damage_dealt`, with no nil-deref on the dead pointer.

4. **Split-second + cascade-of-triggers interaction** — split-second spell is on the stack; a triggered ability fires from a state-based source (e.g., legend rule on a flicker). Assert the trigger *is* placed on the stack (allowed under §702.61) but no opponent ability can respond to it.

5. **Counter-target during split-second window** — confirm that even though split-second blocks new casts, a previously-stacked counterspell from earlier in the turn (somehow on top) still resolves correctly without violating the lock.

---

## Summary

Four of the five rule areas are implemented correctly. The single material gap is **§603.3b trigger batching**: the batched API and the APNAP ordering helper both exist and are tested in isolation, but the production trigger-fire hot paths still push one-at-a-time and resolve inline. This is a wiring fix, not a design fix, and would be a high-value cleanup before the next per-card stub batches that lean heavily on multi-ETB interactions.
