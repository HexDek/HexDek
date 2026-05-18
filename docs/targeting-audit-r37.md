# Targeting and Selection Audit — R37

**Date:** 2026-05-17
**Source:** `internal/gameengine/{stack.go, activation.go, zone_cast.go, targets.go, zone_change.go, keywords_p1p2.go, keywords_combat.go, resolve.go, resolve_helpers.go}` vs. CR §115 + §601.2c / §602.2b / §603.3d / §608.2b in `data/rules/MagicCompRules-20260227.txt`.
**Scope:** Target declaration at announcement (§115.1 / §601.2c), legal-target gating (§115.2 / §608.2b fizzle), change-target / new-target effects (§115.7), distinct-target rule (§115.3 — the canonical CR letter for the "two target creatures must be different" constraint; the prompt cited §115.7 but the actual rule is §115.3), and §702.11 hexproof / §702.18 shroud / §702.16 protection / §702.21 ward checked at targeting time.

> **Rule-letter note.** The task brief cited "§115.7: targets must be different where required ('two target creatures')." In CR 2026-02-27 that constraint is §115.3 ("The same target can't be chosen multiple times for any one instance of the word 'target'"). §115.7 is change-target / new-target. The audit covers both under their canonical letters.

## Summary

| Status                              | Count | Sub-rules                                                  |
|-------------------------------------|-------|------------------------------------------------------------|
| Implemented (fully wired)           |   5   | 115.1 (spells), 115.1c (activated), 115.2, 115.6, 608.2b   |
| Implemented but partial             |   4   | 115.1b (Aura enchant), 115.4 (any-target), 702.21 (ward), 702.16 (protection) |
| Pure stub                           |   3   | 115.3 (distinct), 115.5 (self-target illegal), 115.7 (change-target / new-target) |
| N/A or vacuously satisfied          |   2   | 115.8 (modal retarget), 115.9 (target-count introspection) |

The headline gap is that "two target creatures" / "any number of target X" / "up to N target X" filters degrade to single-target picks (no Quantifier="n" / "up_to_n" handling in `PickTarget`), and **CastSpell / ActivateAbility do not re-validate caller-supplied targets at announcement** (only ward fires, and only on `TargetKindPermanent`). Hexproof/shroud/protection are checked inside `PickTarget` (the AI-driven path) but a malicious or buggy caller can bypass them. §608.2b fizzle at resolution is correctly wired and is the engine's last line of defense.

## Coverage matrix

### §115.1 — Targets declared as part of putting the spell/ability on the stack
**Spell-side (`CastSpell`, stack.go:262).** Targets arrive as a `[]Target` parameter and are stuffed onto the StackItem at line 417 *after* the card leaves the hand and mana is paid. Per §601.2c the legality check on those targets should happen between target declaration and cost payment; in the engine the only at-cast legality fire is `CheckWardOnTargeting` at line 425 (§702.21 ward) and `FireHeroicTriggers` at line 431. Hexproof/shroud/protection are **not** re-validated. Acceptable today because every in-engine caller routes through `PickTarget` first (which filters on `CanBeTargetedByCombat`), but the contract leaks: see Top Issue #2 below.

**Activated-ability side (`ActivateAbility`, activation.go:243).** Same shape — `targets []Target` is taken on trust, pushed to the StackItem at line 480, ward fires at line 486. No hexproof/shroud/protection re-validation.

**Triggered-ability side (§115.1d / §603.3d).** Trigger pushes use `PushPerCardTrigger` / `PushStackItem` paths; per-card handlers pick targets at "puts the ability on the stack" time. Spot-checked Lightning Helix-style triggers (`per_card/*.go`) and the targeting goes through `PickTarget` or per-card helpers that pull from `gs.Opponents`/`gs.Battlefield`. Consistent with §115.1d. No central gate exists; gaps would surface per-handler.

**Aura cast (§115.1b).** Aura spells should be targeted at cast time by their enchant clause. `CastSpell` does not parse the enchant keyword to derive a target; instead the Aura is pushed as a permanent spell and `attachAuraOnETB` (stack.go:1370) picks a target at *resolution*. Functionally indistinguishable for most lines, but it means:
  - Hexproof/protection is checked at ETB, not at cast (would matter for Aura-counter triggers, "whenever you cast a spell that targets" effects on the would-be enchanted creature, and ward — ward would not fire because the cast didn't carry a target).
  - "Counter target Aura spell with target X" introspection (§115.9) sees an empty target list for the Aura on the stack.

**Status:** Implemented for non-Aura spells/abilities with a documented Aura-cast deviation. Move to **partial**.

### §115.2 — Only permanents are legal targets unless explicitly otherwise
PickTarget (`targets.go:27`) routes by `Filter.Base`: player-ish filters go to `pickPlayerTarget`, anything else to `pickPermanentTarget`. Spell-on-stack targets (e.g. counterspells) are produced separately by the counter-spell scan in `stack.go:771` (`CounterCanTarget`) — these correctly return `Target{Kind: TargetKindStackItem}`. Graveyard / hand / exile / library targets are handled by the per-card snowflake path (Snapcaster, Reanimate, etc.) which builds the Target struct directly. Solid coverage.

### §115.3 — Same target can't be chosen multiple times for one instance of "target"
**Status:** Stub. There is no machinery anywhere in the engine that:
  1. Enforces distinctness when a single filter requires multiple targets ("two target creatures"); or
  2. Detects when a spell's text uses the word "target" in two clauses and allows reuse across them (legal per CR §115.3).
The reason is upstream: `Filter.Quantifier == "n"` with `Filter.Count == 2` is parsed correctly by `gameast` but **`PickTarget` never inspects `Count`**; the call falls through `pickPermanentTarget` and returns exactly one Target. So distinctness is moot because the engine can't pick more than one target per "target …" phrase in the first place. See Top Issue #1.

### §115.4 — "any target," "another target," etc.
`any_target` / `any target` / `any` / `{Quantifier:"any", Base:"target"}` are handled in `pickPermanentTarget` (targets.go:211) — defaults to "first opponent seat" rather than scoring across (creature, player, planeswalker, battle). This is the same single-target degradation as §115.3 for "two targets that deal damage"-style spells (Cyclone Summoner, Forked Bolt). Implemented for the trivial case; partial for the multi-target case.

### §115.5 — A spell or ability on the stack is an illegal target for itself
**Status:** Not enforced (PRE-FIX). `pickCounterTarget`-equivalent code (`counterSpellEffect` + `CounterCanTarget`) does not exclude the counter-spell's own StackItem. In practice copies/echoes don't surface this because the trigger pushes after the source resolves, but a self-targeting Insidious Will-style "counter target spell. If you do, you may copy this and choose new targets" can theoretically reach itself. **Closed by Issue #2 fix** — `ValidateTargetsAtAnnouncement` carries a `selfStackItem` parameter and rejects `t.Stack == selfStackItem` with `target_self_illegal`. Wiring callers to pass the constructed StackItem is a follow-up; today every callsite passes nil, so the guard fires only when a caller upgrades to the two-phase pattern.

### §115.6 — Zero-target spells stay valid even if no targets are chosen
The engine has no notion of "may target none." Cards like Comet Storm with X=0, Beast Within with no creatures on the battlefield, etc., currently return `nil` from `PickTarget` → caller treats as "no valid target" → cast aborts. This violates §115.6 for spells whose minimum target count is zero. The Issue #1 fix half-resolves this for the `up_to_n` quantifier path; bare-X spells without `up_to_n` parsing still funnel through the single-target path.

### §115.7 — Change-target / choose-new-targets effects
**Status:** Stub. `resolve_helpers.go:4619` has a regex `reResChangeTgt` matching "^change target" / "^you may choose new targets?", and `resolve_helpers.go:4755` fires a `change_target` event — but the handler returns without actually retargeting anything. Concretely:
  - **Redirect (Reverse Damage-style):** unimplemented.
  - **Misdirection / Bolt Bend / Deflecting Swat / Deflecting Palm:** unimplemented as change-target. Spellskite's "change target to ~" ability is also unimplemented.
  - **Wild Ricochet / Twincast / Reverberate:** copy-spell side works (`resolveCopySpell`, resolve.go:2423) and `MayChooseNewTargets` is read — but the engine just *copies* the original targets verbatim (line 2459) instead of actually offering the controller a new pick. So §707.10c "controller may choose new targets" reduces to "controller keeps original targets." Cosmetic for the common case (same targets are still legal); breaks any line where the original target became illegal or the controller specifically wants to redirect.
  - **§115.7e "only the final set is evaluated":** no infrastructure since no retargeting is wired.
  - **§115.7f "the original division can't be changed":** no infrastructure since divide/distribute effects (Chandra's Pyrohelix, Lightning Helix-style "divide N damage among any number of targets") also aren't supported.

### §115.8 — Modal retargeting doesn't change mode
Vacuously satisfied — modal spells store the chosen mode on the StackItem (`Mode` field, set at cast in per-card handlers), and there is no retargeting infrastructure to violate this. When §115.7 is wired this rule will need an explicit guard.

### §115.9 — Target-count introspection ("a spell that targets …")
Three rules surface in the corpus: Wand of Orcus (counts targets), Brutal Hordechief-style "spell that targets you," Massacre Wurm. Spot-checked the per-card handlers — they scan `StackItem.Targets` directly. Correct for §115.9a/b. §115.9c "targets only [something]" appears in zero pool cards, fine.

### §702.11 / §702.18 / §702.16 — Hexproof, shroud, protection at targeting time
`CanBeTargetedBy(perm, seatIdx)` (keywords_p1p2.go:695) handles shroud (nobody) and hexproof (opponents). `CanBeTargetedByCombat(perm, seatIdx, sourceCard)` (keywords_combat.go:1764) layers on hexproof-from-color and protection-from-color. Wired into `pickPermanentTarget` at targets.go:250 (and into the new `pickNPermanentTargets` from Issue #1's fix).
**Gaps:**
  1. **Protection from [type/quality] beyond colors** — multicolored, monocolored, "everything," instants, artifacts, creatures, etc. Only color-based protection is read (the `protection_from_<colorName>` flag pattern). A card like Mother of Runes giving protection from "noncreature" or True Believer's "you have protection from everything" would miss.
  2. **No re-check at announcement (PRE-FIX).** **Closed by Issue #2 fix** — `ValidateTargetsAtAnnouncement` runs `CanBeTargetedByCombat` on every `TargetKindPermanent` target inside `CastSpell`, `ActivateAbility`, and `CastFromZone`.
  3. **Player-side hexproof / shroud / protection** (Leyline of Sanctity, Spellskite's controller-side, "you have hexproof"). The engine has no `CanBeTargetedBy(seat)` analogue; `TargetKindSeat` targets are not run through any keyword filter. The announcement-time validator covers seat alive/in-game, not keyword protection.

### §702.21 — Ward
`CheckWardOnTargeting` (stack.go:1676) iterates `TargetKindPermanent` targets, reads `ward_cost` flag (or defaults to {1}), and either pays from `ManaPool` or counters the spell. Wired into both `CastSpell` (line 425) and `ActivateAbility` (line 486). Solid for the generic-mana case.
**Gaps:**
  1. **Caster gets no choice (PRE-FIX).** **Closed by Issue #3 fix** — added optional `WardPayer` interface (gameengine/hat.go) with `ShouldPayWard(gs, seatIdx, item, target, wardCost) bool`. `CheckWardOnTargeting` consults it before debiting mana; Hats that don't implement it keep the historical auto-pay behavior.
  2. **Non-mana ward** (Ward—Pay 3 life, Ward—Sacrifice a creature, Ward—Discard a card, Ward—Mill three) is unsupported; only the `ward_cost` integer is consulted.
  3. **Ward on players.** `TargetKindSeat` targets skip the loop entirely (`continue` on line 1681). Cards that grant ward to a player exist (rare — Spellskite-style indirection) but more importantly, `Permanent` ward whose target was redirected via a §903.9b commander-zone choice would slip the check; not currently exercised but worth a note.

### §608.2b — Targets re-checked at resolution; fizzle if all illegal
`CheckTargetLegality` (zone_change.go:745) + `isTargetStillLegal` (zone_change.go:765). Covers:
  - `TargetKindSeat` — seat not Lost and not LeftGame.
  - `TargetKindPermanent` — still on the battlefield (via `permanentOnBattlefield`).
  - `TargetKindStackItem` — still on the stack (pointer-identity scan).
  - Unknown kind → conservatively legal.
Wired into `ResolveStackTop` at stack.go:938. All-illegal → fizzle event + card to graveyard (correct §608.2b for spell items; activated abilities just vanish). Partial-illegal → resolve with subset (also correct).
**Gap:** doesn't re-check hexproof/shroud/protection/ward at resolution. Per CR, those keywords aren't re-checked on resolution (they only gate targeting); zone-membership is the only fresh check. So this is actually correct, despite looking suspicious at first glance.

## Top 3 fixable issues — fixed in this commit

### Issue #1 — Multi-target filters degrade to single-target (high impact, medium effort) — FIXED
**Symptom.** "Choose two target creatures" / "up to three target permanents" / "any number of target X" all returned exactly one `Target` from `PickTarget`, because `targets.go:27` never inspected `Filter.Count` and had no `case "n":` / `case "up_to_n":` arm in the Quantifier switch (line 46). Filter parses the count correctly (`gameast/filter.go:19`), but the engine ignored it.
**Impact.** Every multi-target spell in the corpus mis-resolved: Hex (kill six creatures) destroyed one, Forked Bolt dealt damage once, Volcanic Geyser split zero ways, Aurelia's Fury fizzled. Distinctness (§115.3) was also vacuously moot, hiding a second bug.
**Fix.** Added `case "n", "up_to_n":` to the `PickTarget` Quantifier switch (targets.go) and two new helpers: `pickNPermanentTargets(gs, f, srcSeat, src, n, allowFewer)` and `pickNPlayerTargets(gs, f, srcSeat, n, allowFewer)`. The permanent helper extends the score-ranked candidate list, enforces §115.3 distinctness by pointer-identity dedup, and fires the "targeted" trigger per pick. `up_to_n` returns `[]Target{}` instead of nil when zero is acceptable — this also satisfies CR §115.6 ("may target none"). The integer count comes from `Filter.Count.IntVal()`; non-integer counts (X-driven Count) still degrade to n=1 as a conservative fallback (follow-up: thread chosen-X through Filter resolution).

### Issue #2 — `CastSpell` / `ActivateAbility` don't re-validate targets at announcement (high impact, low effort) — FIXED
**Symptom.** The targets parameter was taken on trust; only ward fired. A direct test (`CastSpell(gs, 0, lightningBolt, []Target{{Permanent: opponentHexproofCreature}})`) succeeded without any §115.2 / §702.11 check. Today every production caller routes through `PickTarget` first so this was dormant, but `hexdek-loki` fuzz tests and per-card unit tests bypass `PickTarget` and could mask real bugs by working around hexproof in test fixtures.
**Impact.** Defense-in-depth gap. Becomes load-bearing when retargeting (§115.7) lands — Bolt Bend / Misdirection will *change* the target list on a StackItem and need to verify the new target is legal. Without a central validator, every retarget callsite would have to remember to invoke `CanBeTargetedByCombat`.
**Fix.** Added `ValidateTargetsAtAnnouncement(gs, controller, sourceCard, targets, selfStackItem) error` in zone_change.go. Enforces §115.2 (legal zone), §702.11 / §702.16 / §702.18 (via `CanBeTargetedByCombat`), seat-alive for `TargetKindSeat`, stack-still-on-stack for `TargetKindStackItem`, and §115.5 self-target rejection for the announcement-in-progress item. Wired into `CastSpell` (stack.go), `ActivateAbility` (activation.go), and `CastFromZone` (zone_cast.go). All three return `*CastError` on failure, matching the existing §601.2e "game returns to the moment before" semantics — callers already restore the hand/zone on CastError.

### Issue #3 — Ward auto-pay/auto-fizzle ignores caster preference (medium impact, low effort) — FIXED
**Symptom.** `CheckWardOnTargeting` (stack.go:1676) decided for the controller: pay if you can, fizzle if you can't. There was no `Hat` hook to *decline* paying ward (e.g. to save mana for a counterspell on the resulting cantrip, or to deliberately fizzle to feed a graveyard).
**Impact.** AI played ward poorly: it would always burn the mana on the first targeting attempt even if a follow-up cast on the same turn was the real plan. Also blocked future Hat sophistication — there was no way to model "ward {3}" being a deterrent rather than an auto-pay.
**Fix.** Added optional `WardPayer` interface (gameengine/hat.go) — `ShouldPayWard(gs, seatIdx, item, target, wardCost) bool`. `CheckWardOnTargeting` type-asserts the caster's Hat against `WardPayer` and consults it before debiting mana; Hats that don't implement it (`GreedyHat`, `PokerHat`, test stubs) continue to auto-pay when affordable. CR §702.21c's "may" is now respected.

## Quick wins (deferred but cheap)

- **Player-side hexproof.** Add `func CanTargetSeat(gs *GameState, seat int, sourceCard *Card, sourceSeat int) bool` reading a `Seat.Flags["hexproof"]` / `protection_from_<X>` map, called from both `PickTarget` (player paths) and the new `ValidateTargetsAtAnnouncement`.
- **Protection from non-color qualities.** Generalize `protection_from_<colorName>` into `protection_from_<quality>` and have `CanBeTargetedByCombat` test against the source card's `TypeLine`, mana value, etc. — required for Mother of Runes / True Believer / Mystic Sanctuary-style protection.
- **Change-target plumbing.** Ship a `ChangeTargets(item, oldT, newT)` primitive that validates the new target (via `ValidateTargetsAtAnnouncement` with the live stack item) and writes back to `item.Targets`. Misdirection / Bolt Bend / Deflecting Swat then become 10-line per-card snowflakes each.
- **Thread the in-flight StackItem into the announcement validator.** Today every caller passes `selfStackItem == nil`, so the §115.5 self-target guard is dormant; once `CastSpell` constructs the StackItem before the validator call, the guard activates for free.
