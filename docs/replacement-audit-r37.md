# Replacement Effects Audit — R37

**Scope:** CR §614 (replacement effects) + §615 (prevention) + §616 (ordering).
**Files reviewed:**
- `internal/gameengine/replacement.go` (1294 LOC) — framework + 17 card handlers
- `internal/gameengine/replacement_test.go` (546 LOC)
- `internal/gameengine/zone_move.go` (285 LOC) — universal zone-change entry
- `internal/gameengine/commander.go` :258–328 — `FireZoneChange` dispatcher
- `internal/gameengine/sba.go` :1700–1822 — `destroyPermSBA` / `sacrificePermSBA` hooks
- `internal/gameengine/resolve.go` :733–830, 1040–1090 — noncombat damage / life events
- `internal/gameengine/combat.go` :1532–1780 — combat damage application
- `internal/gameengine/prevention.go` (327 LOC) — §615 prevention shields
- `internal/gameengine/etb_dispatch.go` (150 LOC) — ETB trigger cascade
- `internal/gameengine/per_card/custom_solphim_mayhem_dominus.go` — damage-doubler scaffold
- `internal/gameengine/per_card/yorvo_lord_of_garenbrig.go` — ETB-with-counters
- `internal/gameengine/per_card/ghave_guru_of_spores.go` — ETB-with-counters

Legend: ✅ implemented · ⚠️ partial / advisory · ❌ missing or unwired

---

## Framework summary

| Component | Location | Status |
|---|---|---|
| `ReplEvent` (mutable event wrapper) | replacement.go:65–142 | ✅ |
| `ReplacementEffect` (registry entry) | replacement.go:177–209 | ✅ |
| `gs.Replacements` slice | state.go (additive) | ✅ |
| `FireEvent` dispatcher | replacement.go:276–307 | ✅ |
| §616.1 category ordering | replacement.go:149–172, 337–349 | ✅ |
| §614.5 applied-once (`ev.AppliedIDs`) | replacement.go:98–102, 292–295 | ✅ |
| §616.1f iterate-until-quiescent | replacement.go:284–296, cap 64 | ✅ |
| APNAP tiebreak | replacement.go:343–348 | ⚠️ active-player first then timestamp only — no simultaneous-choice modeling |
| `Hat.OrderReplacements` delegation | replacement.go:351–360 | ✅ but consulted only for `ev.TargetSeat`'s hat (see §614.6 below) |
| `UnregisterReplacementsForPermanent` on LTB | sba.go:1733, 1796 | ✅ |

### Event types registered

| Event | Fired from | Implemented handlers |
|---|---|---|
| `would_draw` | `resolveDraw` (resolve.go:856-ish) | Lab Maniac, Jace WoM, Alhammarret's Archive, Notion Thief, Dredge |
| `would_gain_life` | `resolve.go:1048` | Alhammarret's, Boon Reflection, Rhox Faithmender |
| `would_lose_life` | `resolve.go:1084` | (no card handlers; framework only) |
| `would_be_dealt_damage` | `resolve.go:740, 807` (noncombat only) | (no card handlers — see Finding #1) |
| `would_put_counter` | `FirePutCounterEvent` callers | Doubling Season, Hardened Scales |
| `would_create_token` | `FireCreateTokenEvent` callers | Doubling Season |
| `would_fire_etb_trigger` | `FireETBTriggerEvent` callers | Panharmonicon, Yarok |
| `would_die` | `sba.go:1720, 1785` | Rest in Peace, Leyline of the Void, Anafenza, Dauthi Voidwalker |
| `would_be_put_into_graveyard` | `commander.go:282` (after `would_change_zone` lands on graveyard) | Rest in Peace, Leyline, Dauthi |
| `would_lose_game` | `sba.go:262` (704.5a 0-life SBA) | Platinum Angel |
| `would_win_game` | (not wired in production) | Platinum Angel handler — see Finding #3-adjacent |
| `would_change_zone` | `FireZoneChange` | (commander 903.9b internal handler only) |

---

## (a) ETB replacement effects — Mikaeus the Lunarch class

**Rule:** §614.1c — "this permanent enters with N +1/+1 counters on it" is a self-replacement on the put-counters event that fires *as part of* the permanent entering, before the entry is finished.

| Item | Status | Notes |
|---|---|---|
| Self-replacement category exists | ✅ | `CategorySelfReplacement` = 0 (replacement.go:150) |
| Doubling Season / Hardened Scales modify a `would_put_counter` event | ✅ | replacement.go:813–897 |
| `FirePutCounterEvent` dispatched as the standard counter-put path | ✅ | replacement.go:415–423 |
| Per-card "enters with N +1/+1" calls `FirePutCounterEvent` | ❌ | yorvo_lord_of_garenbrig.go:31 calls `perm.AddCounter("+1/+1", 4)` directly. ghave_guru_of_spores.go:48 same pattern (`AddCounter("+1/+1", 5)`). All bypass §614.1c. See Finding #2. |
| Generic "enters with" framework hooked at battlefield-entry | ❌ | No `applyEntersWithReplacements(perm)` helper; no `would_enter_with_counters` event type |
| Mikaeus the Lunarch handler | ❌ | No registered handler; card unsupported |
| Variable-counter ETB (e.g. Mikaeus the Lunarch's "X = pro-white count") | ❌ | No scaffolding |

**Verdict:** The §616.1a self-replacement *category* is exercised by a test (`layers_r38_test.go:183 TestReplOrder_SelfReplacementBeforeOther`) but **no production card uses it**. Every "enters with N counters" implementation in `per_card/` calls `AddCounter` directly, so neither Hardened Scales (+1) nor Doubling Season (×2) can interact with these ETBs.

Layer-7c counter resolution is unaffected (counters are applied; their P/T contribution is recomputed). But the *quantity* applied is wrong vs. CR.

---

## (b) Damage replacement — Furnace of Rath / Sulfuric Vortex / Boros Reckoner

**Rule:** §614 (general "if [damage] would be dealt") applies *in both combat and non-combat damage paths*.

| Item | Status | Notes |
|---|---|---|
| `FireDamageEvent` function exists | ✅ | replacement.go:403–411 |
| Noncombat damage routes through `FireDamageEvent` | ✅ | resolve.go:740 (seat target), 807 (permanent target) |
| Combat damage routes through `FireDamageEvent` | ❌ | `applyCombatDamageToPlayer` (combat.go:1544) and `applyCombatDamageToCreature` (combat.go:1697) skip directly to `PreventDamageToPlayer/Permanent` — no §614 chain |
| Damage doubler (Furnace of Rath, Gisela the Broken Blade, Solphim Mayhem Dominus) registered | ❌ | No `RegisterFurnaceOfRath`, no `RegisterGiselaTheBrokenBlade` damage handler. Solphim has an ETB scaffold (custom_solphim_mayhem_dominus.go:38) that increments `seat.Flags["noncombat_damage_doubler_count"]` — but the flag is **never read**. The handler emits `partial` acknowledging "noncombat damage doubling requires DealDamage replacement-effect hook" (line 43). See Finding #3. |
| Damage redirection (Stuffy Doll, Boros Reckoner) | ❌ | No `would_be_dealt_damage` handler redirects to another permanent |
| Lifegain prevention (Sulfuric Vortex "players can't gain life") | ⚠️ | Not a §614 replacement effect — it's a §613 continuous effect ("can't"). resolve_helpers.go:152 mentions `players_cant_gain_life` flag but `resolveGainLife` (resolve.go:1040+) does NOT check it before `FireGainLifeEvent`. Sulfuric Vortex is unimplemented end-to-end. |
| Combat damage prevention shields | ✅ | `PreventDamageToPlayer/Permanent` called from combat.go:1567, 1721 |
| Protection from color / "from everything" | ✅ | combat.go:1554, prevention.go:69 |
| Infect / Wither replace damage with -1/-1 counters | ✅ | combat.go:1582, 1726 (handled inline, not via §614 — acceptable per CR §702.90 modeling latitude) |
| Shield counter (§122.1b) | ✅ | prevention.go:171–189 |

**Verdict:** Noncombat damage replacement is plumbed correctly; combat damage is **a parallel un-instrumented path**. No damage-doubling card works in either combat or non-combat damage because no card has a registered `would_be_dealt_damage` handler at all.

---

## (c) §614.5 — self-replacement on a permanent's own ETB

**Rule:** §614.5 applied-once: a single replacement effect applies at most once per event chain. Implemented via `ev.AppliedIDs[HandlerID]`.

| Item | Status | Notes |
|---|---|---|
| `AppliedIDs` set on `ReplEvent` | ✅ | replacement.go:98–102 |
| `HandlerID` added BEFORE `ApplyFn` invoked (prevents reentrant self-fire) | ✅ | replacement.go:292–295 |
| Iteration cap | ✅ | 64 (replacement.go:255, 297–305 logs `replacement_iteration_cap`) |
| Test coverage | ✅ | `TestRepl_AppliedOnce_DoublerRunsOnlyOnce` :372, `TestRepl_SafetyCap_NoInfiniteLoop` :428 |
| Self-replacement (§616.1a) sorts ahead of Other (§616.1e) | ✅ | `TestReplOrder_SelfReplacementBeforeOther` (layers_r38_test.go:183) |
| `would_change_zone` self-replacement (Rest in Peace dies-to-exile during battlefield-entry) | ✅ | commander.go:281–294 — would_be_put_into_graveyard fires whenever dest lands on graveyard |
| Commander §903.9b zone replacement | ✅ | commander.go:196–242 (intentionally not in AppliedIDs — re-fires per §614.5 exception) |

**Verdict:** §614.5 is correctly modeled. No issues.

---

## (d) §614.6 — multiple replacement effects, affected player chooses order

**Rule:** When two or more replacement effects apply to the same event, the *affected player* (or *controller of the affected object*) chooses one to apply first; iterate.

| Item | Status | Notes |
|---|---|---|
| Sort by category, then APNAP, then timestamp | ✅ | replacement.go:337–349 |
| `Hat.OrderReplacements` consulted | ⚠️ | replacement.go:351–360 — **only `ev.TargetSeat`'s Hat is asked**. For seat-targeted events (would_draw, would_gain_life) this is correct. For permanent-targeted events the affected player is the permanent's controller, which usually equals `TargetSeat` — but the contract is implicit, not enforced. |
| Tied timestamps within category | ⚠️ | Deterministic (slice order) — no APNAP tiebreak for two non-active-player effects on the same timestamp. Real CR fallback is "affected player chooses arbitrarily." MVP works for current 17 handlers. |
| Iterate after each application | ✅ | replacement.go:284–296 |
| Re-evaluate `Applies` on each iteration | ✅ | replacement.go:317–331 |
| Chain example test (Alhammarret + Boon + Rhox = 3 → 24) | ✅ | replacement_test.go:124 `TestRepl_AllThreeLifeDoublers` |

**Verdict:** §614.6 ordering works for the canonical 17 handlers, but two issues to flag:

1. `Hat.OrderReplacements` is only invoked for `ev.TargetSeat`. For a `would_die` event affecting an opponent's creature (Anafenza on opponent's bear), `TargetSeat` is the opponent's seat, which is correct — but for an "affecting permanent" event where the controller and target seat diverge (none in current handler set), the wrong hat could be consulted.
2. The `Hat.OrderReplacements` reorder is consulted on every iteration of the dispatch loop, not just once. This is acceptable per CR (the rule says the player chooses an order, but re-asking each iteration is equivalent given deterministic state) but produces more Hat calls than necessary.

---

## (e) §614.7 — 'instead' vs 'as ... enters' distinction

**Rule:** §614.1a "if X would happen, instead Y" → mutate-or-cancel. §614.1b "X enters with …" / §614.1c "as X enters" — self-replacements that apply *as* the entry happens.

| Item | Status | Notes |
|---|---|---|
| Cancellation (`ev.Cancelled = true`) — "instead never happens" | ✅ | replacement.go:96–97; consumed by Lab Maniac alt-win, Notion Thief redirect, Platinum Angel |
| Mutation in place (`ev.SetCount`, `ev.Payload["to_zone"]`) — "Y happens instead of X" | ✅ | replacement.go:125–131; consumed by all doublers and graveyard-redirectors |
| "Enters with N counters" framework (§614.1b) | ❌ | No `would_enter_with_counters` event type; no `applyEntersWith` helper invoked from `FirePermanentETBTriggers` (etb_dispatch.go) |
| "Enters tapped" replacement (§614.1c) | ⚠️ | `resolve_helpers.go:3539 case "self_enters_tapped"` directly sets `Flags["enters_tapped"]=1` + taps. Not routed through `FireEvent`. Works for the common case but bypasses §616.1 ordering vs. other "enters tapped or untapped" effects (Amulet of Vigor, Vedalken Orrery-class). No interaction tests. |
| "Enters as a copy" (§614.1c, Clone class) | ⚠️ | Handled in `clone.go`. Not routed through replacement framework. CR allows Clone to be modeled either way; current implementation works in isolation. |
| Face-down ETB (§708.4) | ✅ | etb_dispatch.go:23 short-circuits trigger fire-out on face_down |
| Daybound/Nightbound first ETB (§726.2) | ✅ | etb_dispatch.go:53 `OnDayboundOrNightboundETB` |
| Saga first lore counter (§714.3a) | ✅ | etb_dispatch.go:97–149 |

**Verdict:** "Instead" replacements (cancel + mutate) are well-modeled. "As enters" replacements have no general framework — each card's ETB code adds counters / sets flags directly, bypassing the §614 chain. This is the largest single CR conformance gap.

---

## Coverage matrix — replacement-event-to-fire-site

| Event type | Producer | §614.5 applied-once? | §616.1 ordering? | Tests |
|---|---|---|---|---|
| `would_draw` | `resolveDraw` per-card | ✅ | ✅ | replacement_test.go:34–95 |
| `would_gain_life` | `resolveGainLife` | ✅ | ✅ | :97–134 |
| `would_lose_life` | `resolveLoseLife` | ✅ | ✅ (no handlers — pass-through) | — |
| `would_be_dealt_damage` (noncombat) | `applyDamage` | ✅ | ✅ | :508–515 |
| `would_be_dealt_damage` (combat) | — | ❌ NOT FIRED | n/a | — |
| `would_put_counter` | `FirePutCounterEvent` callers | ✅ | ✅ | layers_r38_test.go:183 |
| `would_create_token` | `FireCreateTokenEvent` callers | ✅ | ✅ | replacement_test.go:218 |
| `would_fire_etb_trigger` | combat.go ETB cascade | ✅ | ✅ | :257 |
| `would_die` | `destroyPermSBA`, `sacrificePermSBA` | ✅ | ✅ | :136–215 |
| `would_be_put_into_graveyard` | `FireZoneChange` when dest=graveyard | ✅ | ✅ | :140 |
| `would_change_zone` | `FireZoneChange` (all `MoveCard` callers) | ✅ | ✅ | commander_test.go, zone_change_test.go |
| `would_lose_game` | `StateBasedActions` 704.5a | ✅ | ✅ | :316–345 |
| `would_win_game` | — | ❌ NOT FIRED in production | n/a | :360 |

---

## Top 3 fixable issues

### Issue #1 — Combat damage skips the §614 replacement chain

**Symptom:** Damage-doubling cards (Furnace of Rath, Solphim Mayhem Dominus, Angrath's Marauders) and damage-redirection cards (Stuffy Doll, Boros Reckoner) cannot affect combat damage. Combat-damage tests pass only because no such card is registered.

**Root cause:** `applyCombatDamageToPlayer` (combat.go:1544) and `applyCombatDamageToCreature` (combat.go:1697) jump straight to `PreventDamageToPlayer` / `PreventDamageToPermanent` after the protection check. The `FireDamageEvent` call that the noncombat path uses (resolve.go:740, 807) is absent.

**Fix scope:** ~10 LOC. Insert at the top of both functions:
```go
modified, cancelled := FireDamageEvent(gs, src, seatIdx, target, amount)
if cancelled || modified <= 0 {
    return
}
amount = modified
```
Then audit that `IsCommanderCard` accumulation (combat.go:1573–1578) reads `amount` post-replacement — current code already uses the (now-replaced) local. Add a test mirroring `TestRepl_DamageEventBasic_NoReplacement` for the combat path with a registered damage doubler.

**Risk:** Low. Existing tests that don't register damage handlers will pass through unchanged (pickReplacement returns nil → FireDamageEvent is a no-op).

---

### Issue #2 — "Enters with N counters" bypasses §614.1c self-replacement

**Symptom:** Yorvo, Lord of Garenbrig's 4 +1/+1 counters do not interact with Hardened Scales (should add +1 → 5) or Doubling Season (should double → 8). Same for Ghave (5 counters), Cathars' Crusade triggered ETBs, persist/undying creatures (-1/-1 / +1/+1 counter on return), and any future "enters with N" card. Mikaeus the Lunarch is entirely unimplemented.

**Root cause:** `per_card/yorvo_lord_of_garenbrig.go:31` and similar ETB hooks call `perm.AddCounter("+1/+1", N)` directly, never through `FirePutCounterEvent`. The framework supports self-replacement category sorting (verified by `TestReplOrder_SelfReplacementBeforeOther`), but no card is registered as a self-replacement; instead they all skip the replacement chain entirely.

**Fix scope:** Two parts.
1. **Tactical (1 LOC per card):** Replace `perm.AddCounter("+1/+1", N)` with
   ```go
   final, cancelled := gameengine.FirePutCounterEvent(gs, perm, "+1/+1", N, perm)
   if !cancelled && final > 0 {
       perm.AddCounter("+1/+1", final)
   }
   ```
   in every ETB hook that adds counters (yorvo, ghave, custom_giada_font_of_hope, dyadrine_synthesis_amalgam, lux_artillery, plus the resolve_helpers.go and keywords_batch* sites that handle generic "enters with N counters" AST nodes).
2. **Structural (~30 LOC):** Add a `gameengine.EnterWithCounters(perm, kind, n)` helper that wraps `FirePutCounterEvent` with the §614.1c self-replacement contract (skip the normal "doubler" doubling — Doubling Season DOES double these per oracle, so actually let it through unchanged). Call from a new `applyEntersWithReplacements(gs, perm)` step inserted into `FirePermanentETBTriggers` (etb_dispatch.go) BEFORE the AST self-trigger loop so the counter-add fires before any "when this enters with counters" trigger (Hardened Scales doesn't trigger, so order is moot for it, but the Crusade-class triggers care).

**Risk:** Medium. ETB-counter additions are scattered; need to audit ~10 per-card files. Counter-add timing inside compound ETB effects (Mikaeus the Lunarch's "for each pro-white opponent" → variable N) needs care — the N must be computed BEFORE FirePutCounterEvent dispatches (so the replacement sees the right count). Recommend lifting the count-computation into a helper that returns N, then dispatching.

---

### Issue #3 — `noncombat_damage_doubler_count` is set but never read

**Symptom:** Solphim, Mayhem Dominus ETB scaffolds a per-seat flag (`custom_solphim_mayhem_dominus.go:38`) that no code path consumes. The handler emits `partial` saying so explicitly. era2_handlers_test.go:287 verifies the flag is set — but no integration test verifies damage is doubled, because it isn't.

**Root cause:** The Solphim author opted out of building a `would_be_dealt_damage` replacement handler and left a flag-and-partial as a marker. `resolve.go:733 applyDamage` does not consult the flag; `combat.go:1544 applyCombatDamageToPlayer` doesn't either (and wouldn't fire FireDamageEvent regardless — see Issue #1).

**Fix scope:** ~25 LOC. Replace the flag-set scaffold with a real `would_be_dealt_damage` replacement:
```go
gs.RegisterReplacement(&ReplacementEffect{
    EventType:      "would_be_dealt_damage",
    HandlerID:      handlerKey("Solphim, Mayhem Dominus", "noncombat_dbl", p),
    SourcePerm:     p,
    ControllerSeat: p.Controller,
    Timestamp:      p.Timestamp,
    Category:       CategoryOther,
    Applies: func(gs *GameState, ev *ReplEvent) bool {
        // "noncombat damage" — discriminator on ev.Payload["combat"]
        if ev.Source == nil || ev.Source.Controller != p.Controller {
            return false
        }
        if isCombatDamage(ev) {
            return false
        }
        return ev.Count() > 0
    },
    ApplyFn: func(gs *GameState, ev *ReplEvent) {
        ev.SetCount(ev.Count() * 2)
    },
})
```
Requires `FireDamageEvent` to thread a `combat=true|false` payload bit. Wire `UnregisterReplacementsForPermanent(p)` on LTB (already handled by the generic SBA path). Drop the dead `seat.Flags["noncombat_damage_doubler_count"]` flag and its partial emit. Fix the era2_handlers_test to assert post-replacement damage instead of flag presence.

**Risk:** Low for noncombat damage (already routed through FireDamageEvent). Will become low-medium for combat damage once Issue #1 is also fixed — the same Applies predicate then naturally extends to Furnace of Rath / Gisela. Recommend bundling Issue #1 and #3 in a single sweep, since the test for Solphim post-fix is identical to a Furnace of Rath test.

---

## Out of scope / advisory

- **Triggers-replaced-by-other-triggers** (Omen Machine, Possibility Storm): no support. Would need a `would_fire_triggered_ability` event type or stack-level interception.
- **Replacement-effect dependency ordering** (CR §613.8 applied to replacements, not just continuous effects): not modeled. Likely irrelevant in practice with the 17-handler set.
- **Notion Thief's "first draw in each draw step" interaction with Alhammarret's Archive**: §614.5 applied-once correctly prevents double-redirect, but the "first draw in draw step" flag (notion_thief_normal_draw_*) is set on `p.Flags` — fragile if Notion Thief leaves the battlefield mid-turn. Not blocking; advisory.
- **Affected-player choice across category boundaries**: CR §616.1 says the player chooses among *same-category* effects. Current impl correctly orders by category first, then APNAP within category. ✅
- **Replacement effect tied to copy effects** (Spitting Image, Quicksilver Amulet's "as a copy" line) — see `clone.go`; out of scope for §614 audit.

---

## Summary

Of 5 audit dimensions:

| Dimension | Status |
|---|---|
| (a) ETB replacement effects — Mikaeus class | ❌ generic framework absent; cards bypass `FirePutCounterEvent` |
| (b) Damage replacement (Furnace, Solphim) | ❌ combat path skips chain; no card handlers registered |
| (c) §614.5 self-replacement applied-once | ✅ correctly modeled and tested |
| (d) §614.6 ordering & affected-player choice | ⚠️ APNAP+timestamp+Hat hook works; minor edge cases |
| (e) §614.7 "instead" vs "as enters" | ⚠️ "instead" solid; "as enters" partially modeled outside framework |

Three issues identified, all fixable in <100 LOC each.
