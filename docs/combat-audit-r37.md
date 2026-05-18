# Combat Phase Audit — R37

**Scope:** CR §506–§510 + §511 (combat phase, declare attackers, declare blockers, damage steps, end of combat).
**Source files reviewed:** `internal/gameengine/combat.go` (1940 LOC), `internal/gameengine/keywords_combat.go` (1944 LOC), `internal/gameengine/extra_combat.go`, `internal/gameengine/keywords_goad.go`, `internal/gameengine/keywords_battle.go`.

Legend: ✅ implemented, ⚠️ partial / advisory, ❌ missing / unwired.

---

## §506 — Combat phase structure

| Item | Status | Notes |
|---|---|---|
| Phase entry sequencing (begin → declare atk → declare blk → first-strike dmg → regular dmg → end) | ✅ | `CombatPhase` in `combat.go:195` |
| Skip-to-end-of-combat when no attackers (§506.1) | ✅ | Early return at `combat.go:222` |
| SBAs + priority between sub-steps | ✅ | `StateBasedActions` + `PriorityRound` called after each step (lines 210/225/264/298/308) |
| `gs.CheckEnd()` early-out after every priority round | ✅ | Lines 212/227/266/301/311 |
| Extra-combat queue drain | ✅ | `PendingExtraCombats` + `AddExtraCombat`; `passesCombatRestriction` gates extras |
| §506.4 combat-status flag clearing | ✅ | `EndOfCombatStep` clears `attacking`, `declared_attacker_this_combat`, `blocking`, `attacked_this_combat` |

---

## §506.1 / §507 — Beginning of combat step

| Item | Status | Notes |
|---|---|---|
| "At the beginning of combat" triggers fire | ✅ | `fireBeginningOfCombatTriggers` (line 325); pushed onto stack via `PushTriggeredAbility` per CR §603.3a |
| Phase 5 stack routing (not inline resolution) | ✅ | All triggers use `PushTriggeredAbility` |
| Card-trigger hook (`combat_begin`) | ✅ | `FireCardTrigger(gs, "combat_begin", ...)` for per_card handlers |
| Detect both `combat_start` and `begin_of_combat` aliases | ✅ | `isCombatBeginTrigger` (line 361) |

---

## §508 — Declare attackers

### Legality checks (`canAttack`, line 568)

| Item | Status | Notes |
|---|---|---|
| Tapped → can't attack | ✅ | |
| Summoning sick + no haste → can't attack | ✅ | |
| Defender keyword → can't attack | ✅ | |
| Phased out (§702.26) → can't attack | ✅ | |
| Detained (`Flags["detained"]==1`) → can't attack | ✅ | |
| Power ≤ 0 → can't attack | ⚠️ | Uses `p.Power()` not `gs.PowerOf(p)`; doesn't honor layer-applied buffs/penalties pending recomputation. Minor — most calls happen after `StateBasedActions`. |
| Land-creature restriction (Bumi Unleashed) | ✅ | `passesCombatRestriction` — only `land_creatures_only` tag implemented; unknown tags fail closed |

### Requirements & restrictions (§508.1c–§508.1d)

| Item | Status | Notes |
|---|---|---|
| Hat-driven attacker selection | ✅ | `seat.Hat.ChooseAttackers` (line 412) |
| Per-attacker defender choice (§506.1, multiplayer) | ✅ | `setAttackerDefender` + `flagDefenderSeat`; `pickAttackDefender` heuristic or Hat override |
| Battle as defender (§310.5) | ✅ | `SetAttackerDefenderBattle` / `AttackerDefenderBattle`; damage routing in `DealCombatDamageStep` lines 1230–1287 |
| Propaganda / Ghostly Prison tax (§508.1g) | ✅ | `propagandaTaxFor` — Propaganda, Ghostly Prison, Windborn Muse, Baird, Norn's Annex, Sphere of Safety |
| Goad — must-attack-if-able (§701.39) | ⚠️ | `MustAttackIfAble` / `CannotAttackGoader` exist in `keywords_goad.go` but **not enforced in `DeclareAttackers`**. Engine relies on Hat correctness; if Hat skips a goaded creature, engine silently allows it |
| Goad — can't-attack-goader | ⚠️ | Same — `CannotAttackGoader` defined; not gated in `DeclareAttackers` or `pickAttackDefender` |
| Silent Arbiter (1-attacker cap) | ✅ | Lines 416–426 |
| Cost ordering when multiple costs apply (§508.1f) | ❌ | Only Propaganda tax considered; no cumulative ordering of multiple cost effects |

### Token "enters tapped and attacking" (§506.3)

| Item | Status | Notes |
|---|---|---|
| Scoop-in pass after declared attackers | ✅ | Lines 517–527; tagged `entered_attacking` so own-attack triggers do NOT fire |
| Counted toward battalion / pack tactics totals | ✅ | `attackers` slice (declared+scoop) passed to `FireBattalionTriggers` / `FirePackTacticsForAttackers` |

### Attack-time keyword triggers

| Keyword | Status | Notes |
|---|---|---|
| Exalted (§702.83) | ✅ | `ApplyExalted` when `len(declared)==1` |
| Dethrone (§702.105) | ✅ | `FireDethroneTriggers` |
| Battalion (§702.101) | ✅ | `FireBattalionTriggers` — 3+ attacker threshold |
| Pack tactics (§702.149) | ✅ | `FirePackTacticsForAttackers` — power ≥ threshold |
| Battle cry (§702.91) | ✅ | `ApplyBattleCry` via `CheckAttackKeywordsCombat` |
| Myriad (§702.115) | ✅ | `ApplyMyriad` + `FireMyriadTriggers` |
| Melee (§702.121) | ✅ | `ApplyMelee` |
| Annihilator (§702.85) | ✅ | `FireAnnihilatorTriggers` — supports `annihilator N` from AST args |
| Provoke (§702.39) | ✅ | `ApplyProvoke` / `FireProvokeTriggers` — finds largest creature, marks `provoked_by` |
| Raid (§702.136) | ✅ | `seat.Flags["attacked_this_turn"]=1` for whole-turn predicate |
| Tap-on-attack (`tap_event`) | ✅ | `FireCardTrigger("tap_event", ...)` for Magda, Emmara, etc. |

### Self-attack triggers + ally-attack triggers

| Item | Status | Notes |
|---|---|---|
| Own-attack `whenever ~ attacks` (§603.3a) | ✅ | `fireAttackTriggers` step (1) — only for DECLARED attackers (§508.1) |
| Ally-attack `whenever a creature you control attacks` | ✅ | step (2); substring match on Raw |
| `another creature attacks` variant | ✅ | substring match |
| Stack routing (not inline) | ✅ | All via `PushTriggeredAbility` |

---

## §509 — Declare blockers

### Block legality (`canBlockGS`, line 968)

| Item | Status | Notes |
|---|---|---|
| Tapped blocker → can't block | ✅ | |
| Suspected creature can't block (§701.62a) | ✅ | `IsSuspected(blocker)` |
| Phased out → can't block | ✅ | both attacker and blocker checked |
| Flying — blocked only by flying/reach (§702.9b) | ✅ | |
| Horsemanship (§702.30) | ✅ | `CanBlockP1P2` |
| Intimidate, fear, shadow, skulk, daunt | ✅ | `CanBlockCombatKeywords` in `keywords_combat.go` |
| Landwalk (§702.14) | ✅ | `LandwalkType` + `DefenderControlsLandType`; requires non-nil `gs` (the `canBlock` legacy wrapper passes nil and skips) |
| Sidar Kondo (P≤2 can't be blocked by P≥3) | ✅ | Lines 1004–1008 |
| Unblockable | ✅ | |
| Protection from attacker's color (§702.16b) | ✅ | `attackerHasProtectionFrom` |
| Protection from non-color types ("from creatures", "from artifacts") | ❌ | `ProtectionTypes` exists in `keywords_combat.go:1661` but `attackerHasProtectionFrom` only checks color. Combat path ignores type-based protection |
| Menace — requires 2+ blockers (§702.110b) | ✅ | Both Hat path and fallback enforce |
| Lure / "must be blocked by all creatures able" | ❌ | No generic lure/Hunted Wumpus enforcement. Provoke is the only "force block" implemented |
| Provoke target's "must block if able" | ⚠️ | `provoked_by` flag is set but enforcement is left to Hat — `DeclareBlockers` does not require the provoked blocker to be assigned |

### Damage assignment order (§509.2 + §510.1c)

| Item | Status | Notes |
|---|---|---|
| Attacker assigns blocker order | ⚠️ | Order is **hardcoded ascending-toughness** at `combat.go:1327–1335`. CR §509.2 makes this an attacking-player choice. No `ChooseBlockerOrder` hook on Hat |
| Lethal-first damage allocation | ✅ | `lethalAmountGS` + per-blocker `give = min(remaining, need)` |
| Deathtouch sets lethal=1 | ✅ | `lethalAmountGS` line 1403 |
| Trample spillover to defender (§702.19b) | ✅ | Lines 1350–1352 |
| Trample over planeswalkers | ✅ | `HasTrampleOverPlaneswalkers` available (used elsewhere) |
| Trample spillover to battle (§310.5c) | ✅ | Lines 1282–1286 |
| Multi-blocker (3+) via fallback policy | ❌ | Fallback in `DeclareBlockers` only picks one blocker per attacker (two for menace). 3+ blockers only happen via Hat path |

### Ninjutsu / Sneak activation window

| Item | Status | Notes |
|---|---|---|
| Ninjutsu (§702.49) — return unblocked atk, ninja enters tapped+attacking | ✅ | `CheckNinjutsuRefactored`; ninja's attack triggers do NOT fire (§702.49a) |
| Sneak — same window, but IS a cast | ✅ | `CheckSneak`; increments commander tax, storm, fires cast triggers |

### Block-time keyword triggers

| Keyword | Status | Notes |
|---|---|---|
| Bushido (§702.45) | ✅ | `CheckCombatKeywordsP1P2` |
| Flanking (§702.25) | ✅ | `CheckCombatKeywordsP1P2` |
| Rampage (§702.23) | ✅ | `FireRampageTriggers` via `CheckCombatKeywordsCombat` |
| Afflict (§702.130) | ✅ | `FireAfflictTriggers` |
| Banding (§702.21) damage redistribution | ❌ | `ApplyBandingDamageRedistribution` defined in `keywords_combat.go:234` but **has no non-test caller**. Combat damage step does not invoke it. Banding declare-step controller-chooses-blocker is also not implemented |
| Block event trigger (`fireBlockTriggers`) + `becomes_blocked` | ✅ | Line 239 |

---

## §510.1 — First-strike combat damage step

| Item | Status | Notes |
|---|---|---|
| Skip step entirely if no FS/DS in pool | ✅ | `hasFS` detection lines 272–291 |
| `dealsInStep(p, true)` → FS or DS | ✅ | Line 1379 |
| SBAs between FS and regular step (§510.1c → §704.3) | ✅ | Line 298 |
| Priority after FS damage (§117.3) | ✅ | Line 300 |
| Damage triggers (`deals_combat_damage`) fire once per step | ✅ | `fireCombatDamageTriggers`; double-strike runs both steps → 2 trigger fires |

---

## §510 — Regular combat damage step

### Per-source damage application

| Item | Status | Notes |
|---|---|---|
| `dealsInStep(p, false)` → !FS or DS | ✅ | Line 1382 |
| Power read via `gs.PowerOf` (layer-aware) | ✅ | Line 1219 |
| Skip if attacker died before damage step | ✅ | `alive(gs, atk)` gate |
| Unblocked → all damage to defender | ✅ | |
| Blocked but all blockers died → no damage unless trample | ✅ | |
| Defender seat dead → damage not assigned (§800.4e) | ✅ | Lines 1297–1299 |
| Battle defender path | ✅ | Lines 1230–1287; unblocked → defense counters; blocked → through blockers, trample to battle |
| Blocker → attacker damage (Phase B) | ✅ | Lines 1356–1370 |
| Multiple blockers all hit same attacker | ✅ | Each blocker deals its full power to attacker |

### Damage effects (per-instance)

| Item | Status | Notes |
|---|---|---|
| Lifelink (§702.15) | ✅ | both creature- and player-damage paths |
| Deathtouch (§702.2) any nonzero = lethal | ✅ | Force-bumps `MarkedDamage` + sets `deathtouch_damaged` flag for §704.5h SBA |
| Infect (§702.90) — players get poison, creatures get -1/-1 counters | ✅ | `HasInfect(src)` branches |
| Wither (§702.80) — creatures get -1/-1 (players normal) | ✅ | `HasWither` branch + `ApplyWitherDamageToCreature` |
| Toxic N (§702.165) — poison ON TOP of normal damage | ✅ | `HasToxic` branch |
| Color protection (§702.16d) on player | ✅ | `Flags["protection_from_everything"]` + Teferi's Protection |
| Color protection on creature | ✅ | |
| Type/non-color protection on creature | ❌ | Combat path ignores `HasProtectionFromType` |
| Prevention shields (§615) | ✅ | `PreventDamageToPlayer` / `PreventDamageToPermanent` applied before damage |
| Fog (`prevent_all_combat_damage`) | ✅ | Lines 1196–1205 |
| Commander damage tracking (§903.10a) | ✅ | `AccumulateCommanderDamage` post-prevention, post-protection |
| Renown (§702.111) | ✅ | `ApplyRenownOnCombatDamage` |
| Monarch combat steal (§721.4) | ✅ | `CheckMonarchCombatSteal` |
| Speed advance (§702.179) — once per turn per dealer | ✅ | `SpeedDamageReporter` |
| Prowl per-card tracking (§702.74) | ✅ | `ctrlTurn.CombatDamageBy` append, dedup per turn |
| Bloodthirst (§702.54) `damage_taken_this_turn` flag | ✅ | Lines 1495–1499 |
| Basilisk-granted ability mark-for-destroy | ✅ | Lines 1661–1668 |
| `combat_damage_player` card-trigger hook | ✅ | For Fynn, Yuriko, etc. |

### Combat damage triggers (§510.2)

| Item | Status | Notes |
|---|---|---|
| `deals_combat_damage` triggers fire | ✅ | `fireCombatDamageTriggers` |
| `deals_damage` alias accepted | ✅ | Both events checked |
| Pushed onto stack (§603.3a) | ✅ | `PushTriggeredAbility` |
| Double-strikers fire damage triggers twice | ✅ | Step runs twice |

---

## §511 / §506.4 — End of combat

| Item | Status | Notes |
|---|---|---|
| "At end of combat" triggers fire | ✅ | `EndOfCombatStep` loops every seat's battlefield |
| Stack routing for end-of-combat triggers | ✅ | `PushTriggeredAbility` |
| Combat status flags cleared (§506.4) | ✅ | `attacking`, `declared_attacker_this_combat`, `blocking`, `attacked_this_combat` |
| `varina_triggered_this_combat` cleared | ✅ | Line 1773 |
| Until-end-of-combat continuous effects expire (§500.5a) | ✅ | `gs.ContinuousEffects` filtered |
| Until-end-of-combat modifications expire | ✅ | `p.Modifications` filtered |
| Characteristics cache invalidated on mod removal | ✅ | Required for 0-toughness SBA detection after FS buff drop |
| Marked damage **persists** until §514.2 cleanup | ✅ | Correct — not cleared here |
| Final SBA + priority pass after end-of-combat | ✅ | Lines 318–320 |

---

## Cross-cutting edge cases

| Case | Status | Notes |
|---|---|---|
| Extra combats (Aggravated Assault, Najeela, Moraug) | ✅ | `PendingExtraCombats` FIFO + per-combat `OnBegin` rider |
| Multi-defender combat (4-player pod) | ✅ | `DeclareBlockersMulti` partitions attackers by recorded defender seat; each defending seat declares own blockers (§509.1a) |
| Attacker → planeswalker (§508.1b) | ⚠️ | `pickAttackDefender` only picks player seats; planeswalker targeting via `flagDefenderSeat` is not in the legal-target enum. Per-card / Hat may set it externally |
| First-strike + deathtouch (kill before swung-back) | ✅ | SBAs run between steps; if FS deathtouch killed blocker, blocker doesn't hit back in regular step |
| Double-strike + deathtouch | ✅ | Both steps fire; both apply deathtouch lethality |
| Lifelink + double-strike (both hits gain life) | ✅ | `GainLife` per-instance, fires twice |
| Lifelink + infect (creature path) | ✅ | Lifelink branch runs inside infect early-return |
| First-strike + lifelink + bridge to opponent at 1 life | ✅ | Damage applied via `applyCombatDamageToPlayer` then SBA next |
| Trample + deathtouch (1 to each blocker, rest to defender) | ⚠️ | Hardcoded order via toughness; CR §702.2c allows attacker to assign 1 to each blocker then trample. Current code: `need = lethalAmountGS` (which returns 1 with deathtouch), so we'll spend 1 per blocker and trample the rest — **correct behavior** with deathtouch, but only by accident of the ordering policy |
| Banded attackers — damage redistribution by attacker controller | ❌ | Code path defined but unwired |
| Goaded creature controlled by Hat that ignores must-attack | ⚠️ | Engine has no fallback enforcement; relies on Hat compliance |
| Suspect + first-strike interaction | ✅ | Suspect's "can't block" handled at declare time; loss-of-FS for suspected creatures isn't an MTG rule, so no interaction needed |

---

## Identified gaps — buggy / missing

### High priority

1. **Banding damage redistribution unwired** (`keywords_combat.go:234`)
   - `ApplyBandingDamageRedistribution` has zero non-test callers.
   - Wire into `DealCombatDamageStep` at the start of Phase A and Phase B: collect banded attackers/blockers, redirect blocker damage to controller's choice within the band.

2. **Protection from non-color types ignored in combat** (`combat.go:1088`)
   - `attackerHasProtectionFrom` only consults `protectionColors`.
   - `HasProtectionFromType` exists but is unused in the combat path.
   - Cards like Mother of Runes ("protection from a color") work; True Believer ("protection from black") works; but "protection from creatures" (Dawn Charm), "protection from artifacts" (Mirran Crusader) silently fail to block-restrict.

3. **Damage assignment order is policy, not choice** (`combat.go:1327–1335`)
   - Attacker assigns ascending-toughness order — CR §509.2/§510.1c-d give this choice to the **attacking player**.
   - Add `Hat.ChooseBlockerOrder(attacker, blockers) []*Permanent` callback; fall back to current heuristic.

### Medium priority

4. **Goad enforcement is advisory** (`combat.go:DeclareAttackers`)
   - `MustAttackIfAble` / `CannotAttackGoader` defined but not gated in `DeclareAttackers` or `pickAttackDefender`.
   - Add a post-Hat-pick validation pass: for each `MustAttackIfAble(p)` not in `chosen`, audit-log; for each chosen attacker whose defender is the goader, redirect via `CannotAttackGoader`.

5. **Lure / generic must-be-blocked-by-all-able** — missing
   - Provoke target-marking exists; no generic lure (Lure, Hunted Wumpus, Bellowing Tanglewurm flavor) gating in `DeclareBlockers`.

6. **Fallback `DeclareBlockers` capped at 1 (or 2 for menace)** (`combat.go:866`)
   - Hat path supports arbitrary multi-block; fallback policy never assigns 3+ blockers, so deathtouch-trampler vs wall-of-1/1s defense plans fail when Hat is nil.

### Low priority / nice-to-have

7. **`canAttack` uses raw `Power()`, not `gs.PowerOf`** — could miss layer-applied buffs/penalties applied without an SBA between (rare).
8. **Cost-stacking order (§508.1f)** — only Propaganda taxes; no support for multiple attack-cost effects with ordering rules.
9. **Planeswalker-as-defender** — `flagDefenderSeat` enum doesn't model a PW target; relies on per_card flags.

---

## Sample edge-case tests proposed

```go
// 1. Protection-from-creatures blocker is illegal (MEDIUM)
//    Mirran Crusader (protection from black, white) attacks; black 5/5
//    cannot block. Currently passes by color. Add: artifact creature
//    with protection from creatures cannot be blocked.
func TestCombat_ProtectionFromCreatures_BlockerIllegal(t *testing.T) { ... }

// 2. Banding damage redistribution (HIGH)
//    Banded 2/2 + 3/3 attacking, blocked by 4/4. Defender deals 4 to band.
//    Attacker chooses to put all 4 on the 2/2 (3 wasted = 2/2 dies, 3/3
//    survives). Currently banding doesn't redirect.
func TestCombat_Banding_AttackerRedistributesDamage(t *testing.T) { ... }

// 3. Goad forces attack even when Hat would skip
//    Seat-0 goads seat-1's only creature. Seat-1 hat policy says "skip
//    combat". Engine MUST still send the goaded creature at a non-goader.
func TestCombat_Goad_EngineForcesAttackAfterHatSkip(t *testing.T) { ... }

// 4. Multi-blocker damage assignment order is attacker's choice
//    1/1 lifelink blocker + 5/5 vanilla blocker vs 5/5 attacker. CR says
//    attacker MUST be able to choose to kill the 5/5 first. Currently
//    forced to kill 1/1 first (ascending toughness).
func TestCombat_DamageAssignmentOrder_AttackerChoosesNotPolicy(t *testing.T) { ... }

// 5. Trample + deathtouch with hardcoded order — should still work
//    3-power deathtouch tramplers vs three 4/4 blockers + defender at 1.
//    Expect 1 dmg to each blocker (deathtouch lethal=1), 0 trample,
//    defender unharmed; deathtouch_damaged flag set on all 3 blockers.
func TestCombat_TrampleDeathtouch_OnePerBlockerNoSpillover(t *testing.T) { ... }

// 6. Double-strike + deathtouch fires triggers twice
//    Verify deals_combat_damage trigger fires once in FS step, once in
//    regular step; commander damage accumulates both hits; lifelink
//    gains twice.
func TestCombat_DoubleStrikeDeathtouch_TwoTriggerFires(t *testing.T) { ... }

// 7. Menace fallback when only 1 blocker available
//    Attacker has menace, defender has 1 untapped creature. Fallback
//    `DeclareBlockers` should NOT assign that creature (since menace
//    needs 2). Verified at combat.go:847.
func TestCombat_Menace_SoloBlockerCannotSatisfy(t *testing.T) { ... }

// 8. Battle attacker — blocker intercepts, trample spills to battle
//    5/5 trample attacker pointed at battle (3 defense counters),
//    blocked by 2/2. 2 dmg to blocker (kills), 3 trample to battle
//    (defense counters → 0, transforms via §310.10).
func TestCombat_BattleAttacker_BlockedTrampleSpillsToBattle(t *testing.T) { ... }

// 9. Goaded creature can't attack the goader
//    Seat-0 goads seat-1's creature; seat-1 only has seat-0 as opponent.
//    CannotAttackGoader(perm, 0) == true. Currently relies on Hat to
//    skip. Engine should drop the attack and log it.
func TestCombat_Goad_CannotAttackGoaderEvenIfOnlyOpp(t *testing.T) { ... }

// 10. Phasing out an attacker between declare and damage
//     Attacker phases out (Teferi's Protection on its controller) after
//     declare-blockers. `alive(gs, atk)` skips damage; combat status
//     flags clear on end-of-combat normally. Verify no nil-deref on
//     phase-out mid-combat.
func TestCombat_AttackerPhasedOutMidCombat_NoDamage(t *testing.T) { ... }
```

---

## Summary

The combat phase is **broadly correct** on the happy paths and on most common edge cases (FS/DS interaction, trample, lifelink, deathtouch, infect, wither, toxic, commander damage, battle attackers, multiplayer defender partition, Ninjutsu/Sneak windows, extra-combat queue). The three meaningful gaps are:

1. **Banding wired only at the keyword level** — handler exists, never called.
2. **Type-based protection ignored in combat** — only colors filter blockers and damage.
3. **Damage assignment order is fixed policy, not attacker choice** — works by happy coincidence with deathtouch + trample but is mechanically wrong for tactical multi-blocker scenarios.

Goad enforcement and the fallback-blocker policy are **advisory weaknesses** rather than outright bugs — current Hat implementations handle them correctly, but a misbehaving Hat or a nil-Hat code path would silently violate CR.

No nil-deref or invariant-violation patterns surfaced in this read.
