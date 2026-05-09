# Thor 2.1 Corpus Audit — New Structure Detection

**Date:** 2026-05-08
**Engine commits:** 661f545..18b2152 (MoveResult, ExileLinked, ZoneCastPermission, TurnCounters, CastRecord, Adventure/Prepared/Paradigm)
**Corpus:** 37,384 oracle cards, 31,963 with AST coverage

---

## Executive Summary

The scaffold system (goldilocks.go + conditional_setup.go) was priming conditions using **legacy Flags** but the engine now reads from **TurnCounters** for morbid and other turn-scoped checks. This caused 15 morbid cards to silently fail (engine checked `Turn.CreaturesDied`, scaffold only set `gs.Flags["creature_died_this_turn"]`).

**Fixed:** All 14 existing scaffold kinds now dual-write to both TurnCounters and legacy Flags.
**Added:** 7 new `condScaffoldKind` entries for engine structures that had no scaffold at all.
**Result:** Dead-effect failures reduced from 68 → 53 (morbid fully resolved).

---

## Baseline Results

| Module | Tests | Pass | Fail | Rate |
|--------|-------|------|------|------|
| corpus-audit (correctness) | 18,934 | 18,933 | 1 (Dread) | 99.99% |
| goldilocks (liveness) — BEFORE | 31,963 | 30,049 | 1,914 | 94.01% |
| goldilocks (liveness) — AFTER | 31,963 | 30,064 | 1,899 | 94.06% |

---

## Category A: TurnCounters Migration Gaps (FIXED)

The scaffold system set legacy `gs.Flags` / `seat.Flags` but NOT the new `seat.Turn.*` counters. Per-card handlers reading `Turn.*` saw zeroed values and evaluated conditions as false.

### Scaffolds Updated (dual-write to TurnCounters + legacy Flags)

| Scaffold | Legacy Field | TurnCounter Field | Cards Affected |
|----------|-------------|-------------------|----------------|
| `condScaffoldCreatureDiedThisTurn` | `gs.Flags["creature_died_this_turn"]` | `seat.Turn.CreaturesDied`, `seat.Turn.PermanentsLeft` | 15 (morbid) |
| `condScaffoldCastSpellThisTurn` | `seat.Flags["cast_spell_this_turn"]` | `seat.Turn.SpellsCast`, `seat.Turn.Casts` | — |
| `condScaffoldSacrificedThisTurn` | `gs.Flags["sacrificed_this_turn"]` | `seat.Turn.Sacrificed`, `seat.Turn.PermanentsLeft` | — |
| `condScaffoldAttackedThisTurn` | `seat.Flags["attacked_this_turn"]` | `seat.Turn.Attacked` | — |
| `condScaffoldDiscardedThisTurn` | `seat.Flags["discarded_this_turn"]` | `seat.Turn.Discarded` | — |
| `condScaffoldLandfallThisTurn` | `seat.Flags["landfall_this_turn"]` | `seat.Turn.LandsPlayed` | — |
| `condScaffoldDrawnCardThisTurn` | `seat.Flags["drawn_card_this_turn"]` | `seat.Turn.CardsDrawn` | — |
| `condScaffoldCreatureETBThisTurn` | `gs.Flags["creature_etb_this_turn"]` | `seat.Turn.CreaturesEntered` | — |

### Goldilocks ability_word Setup (FIXED)

The `ability_word` case in goldilocks.go now also sets:
- `seat.Turn.CreaturesDied = 1` (morbid via TurnCounters)
- `seat.Turn.PermanentsLeft = 1` (disappear/void via TurnCounters)
- `gs.Flags["permanent_left_bf"] = 1` (disappear/void via legacy)
- `seat.Turn.Attacked = true` (pack tactics via TurnCounters)

---

## Category B: New condScaffoldKind Entries (ADDED)

7 new scaffold kinds for engine structures that had no detection or priming:

### `condScaffoldPermanentLeftBF`
- **Detection:** `"permanent left"` or `"permanent left the battlefield"` (excluding revolt)
- **Mutation:** `Turn.PermanentsLeft++`, `gs.Flags["permanent_left_bf"]=1`, sacrifice event log
- **Cards:** Disappear (6) + Void (7) = 13 cards. Currently still failing because the engine's `permanent_left_this_turn` condition evaluator reads `gs.Flags["permanent_left_bf"]` which is never written by `zone_change.go` — the engine sets `Turn.PermanentsLeft` but the evaluator doesn't read it. **Engine bug filed separately.**

### `condScaffoldSecondSpellThisTurn`
- **Detection:** regex `(?:second|third)\s+(?:spell|creature|noncreature|instant|sorcery|...)`
- **Mutation:** `Turn.SpellsCast=2`, appends 2 `CastRecord` entries, sets legacy flags
- **Cards:** 88 cards (Rashmi, Sunstar Lightsmith, Lotho, Kraum, etc.)

### `condScaffoldDescendedThisTurn`
- **Detection:** `"descended"` or `"descend"`
- **Mutation:** `Turn.Descended=true`, legacy `DescendedThisTurn=true`, creature in GY
- **Cards:** 12 cards (Child of the Volcano, Enterprising Scallywag, etc.)

### `condScaffoldLifeLostThisTurn`
- **Detection:** `"lost life"` + `"this turn"` (self, not opponent)
- **Mutation:** `Turn.LifeLost=3`, legacy `life_lost_this_turn` / `lost_life_this_turn` flags
- **Cards:** Greven, Sygg, Strefan, Rowan (per-card handlers reading `Turn.LifeLost`)

### `condScaffoldTokensCreatedCount`
- **Detection:** `"tokens"` + `"created this turn"`
- **Mutation:** `Turn.TokensCreated=3`, `Turn.TreasuresCreated=2`
- **Cards:** 3 cards (Vazi, Thalisse, Ellyn Harbreeze)

### `condScaffoldCastFromExile`
- **Detection:** `"cast"` + `"from exile"` (excluding `"you may cast"`)
- **Mutation:** `Turn.CastFromExile++`
- **Cards:** Cards that check cast-from-exile counts (Party Thrasher, etc.)

### `condScaffoldExileLinkedReturn`
- **Detection:** `"exiled with"` + (`"return"` or `"leaves"`)
- **Mutation:** Appends card to `srcPerm.LinkedExile`, places card in opponent exile
- **Cards:** 89 exile-until-leaves cards (Oblivion Ring, Fiend Hunter, Trapjaw Tyrant, etc.)

---

## Category C: Remaining Dead-Effect Failures (53 cards — NOT scaffold-fixable)

These failures require goldilocks EVENT FIRING changes, not scaffold priming:

### Kinship (12 cards)
**Pattern:** "look at the top card of your library. If it shares a creature type with this creature..."
**Why failing:** Goldilocks doesn't set up library top card to share creature type with source.
**Fix needed:** `goldilocks.go` ability_word setup: peek library[0], match source's creature subtype.

### Valiant (10 cards)
**Pattern:** "whenever this creature becomes the target of a spell or ability you control for the first time each turn"
**Why failing:** Goldilocks doesn't fire a targeting event on the source.
**Fix needed:** `fireTriggerEvent` or board setup to cast a spell targeting the source creature.

### Pack Tactics (7 cards)
**Pattern:** "if you attacked with creatures with total power 6 or greater this combat"
**Why failing:** Board setup places creatures but doesn't give them enough total power.
**Fix needed:** Place 2x 4/4 creatures with `attacking` flag on seat 0.

### Void (7 cards)
**Pattern:** "if a nonland permanent left the battlefield this turn or a spell was warped this turn"
**Why failing:** Engine condition evaluator reads `gs.Flags["permanent_left_bf"]` but `zone_change.go` only writes `Turn.PermanentsLeft`. Scaffold now sets the flag, but goldilocks board setup doesn't (fixed in goldilocks.go but these cards use the event-based path).
**Fix needed:** Engine bug — `zone_change.go` should also set `gs.Flags["permanent_left_bf"]`.

### Disappear (6 cards)
**Pattern:** "if a permanent left the battlefield under your control this turn"
**Same root cause as Void.** Engine doesn't write the flag that the condition evaluator reads.

### Join Forces (3 cards)
**Pattern:** "each player may pay any amount of mana"
**Why failing:** Requires multi-player mana payment interaction not modeled in goldilocks.
**Fix needed:** Specialized goldilocks setup with `ManaPool > 0` on all seats.

### Attacked This Turn (2 cards)
**Pattern:** "if you attacked this turn" (ETB conditional)
**Why failing:** Goldilocks sets `attacked_this_turn` flag but doesn't set `Turn.Attacked`. Now fixed by goldilocks.go change — **may resolve on re-test with event path**.

### Unless Control (2 cards)
**Pattern:** "unless you control another [Pirate/colorless creature]"
**Why failing:** Board setup doesn't include the required second permanent matching the subtype.
**Fix needed:** Parse oracle text for the required companion permanent type.

### Unclassified (4 cards)
- **Soul-Guide Lantern** — exile target from graveyard (needs graveyard setup with specific card)
- **Lord of Tresserhorn** — sacrifice two creatures (needs 3 creatures on board)
- **Demonic Hordes** — destroy target land (needs opponent land)
- **Scourge of Numai** — unless you control an Ogre (subtype check)

---

## Oracle Pattern Coverage by New Engine Structure

| Engine Structure | Oracle Cards | With AST | Scaffold Coverage |
|-----------------|-------------|----------|-------------------|
| ExileLinked (O-Ring) | 89 | 84 (94%) | `condScaffoldExileLinkedReturn` NEW |
| ZoneCastPermission (impulse draw) | 228 | 215+ (94%) | Handled by existing zone-cast deep rules |
| TurnCounters.LifeGained | 117 | 117 (100%) | `primeGainedLife` (already calls `GainLife` → `Turn.LifeGained`) |
| TurnCounters.SpellsCast / CastRecord | 88 | 88 (100%) | `condScaffoldSecondSpellThisTurn` NEW |
| TurnCounters.Descended | 12 | 12 (100%) | `condScaffoldDescendedThisTurn` NEW |
| TurnCounters.TokensCreated | 3 | 3 (100%) | `condScaffoldTokensCreatedCount` NEW |
| TurnCounters.CastFromExile | ~100 | ~95 | `condScaffoldCastFromExile` NEW |
| CastRecord.ManaValue | 619 | 600+ | Handled by existing AST MV threshold logic |
| Adventure/Prepared/Paradigm | 27 | 27 (100%) | Handled by deep rules + ZoneCastPermission |

---

## Per-Card Handler → TurnCounters Dependency Map

Handlers that read `Turn.*` and now have scaffold support:

| Handler | TurnCounter Read | Scaffold |
|---------|-----------------|----------|
| `gorma.go` | `Turn.CreaturesDied` | `condScaffoldCreatureDiedThisTurn` ✅ |
| `mahadi_emporium_master.go` | `Turn.CreaturesDied` | `condScaffoldCreatureDiedThisTurn` ✅ |
| `lathiel_the_bounteous_dawn.go` | `Turn.LifeGained` | `primeGainedLife` (via `GainLife`) ✅ |
| `sorin_of_house_markov.go` | `Turn.LifeGained` | `primeGainedLife` ✅ |
| `shanna_purifying_blade.go` | `Turn.LifeGained` | `primeGainedLife` ✅ |
| `moseo_veins_new_dean.go` | `Turn.LifeGained` | `primeGainedLife` ✅ |
| `aetherflux_reservoir.go` | `Turn.SpellsCast` | `condScaffoldCastSpellThisTurn` ✅ |
| `chaos_cascade.go` | `Turn.NthCastOfType` | `condScaffoldSecondSpellThisTurn` ✅ |
| `greven_predator_captain.go` | `Turn.LifeLost` | `condScaffoldLifeLostThisTurn` ✅ |
| `sygg_river_cutthroat.go` | `Turn.LifeLost` | `condScaffoldLifeLostThisTurn` ✅ |
| `strefan_maurer_progenitor.go` | `Turn.LifeLost`, `Turn.DamageReceived` | `condScaffoldLifeLostThisTurn` ✅ |
| `rowan_scion.go` | `Turn.LifeLost` | `condScaffoldLifeLostThisTurn` ✅ |
| `vazi_keen_negotiator.go` | `Turn.TreasuresCreated` | `condScaffoldTokensCreatedCount` ✅ |
| `zoyowa_lava_tongue.go` | `Turn.Descended` | `condScaffoldDescendedThisTurn` ✅ |

---

## Engine Bug: `permanent_left_bf` Flag Never Written

**Impact:** 13 cards (void + disappear ability words)
**Root cause:** `zone_change.go` writes `Turn.PermanentsLeft` but NOT `gs.Flags["permanent_left_bf"]`. The condition evaluator at `resolve.go:402` reads the flag, not the counter. The scaffold now writes the flag, but the engine's normal game loop doesn't.
**Fix:** In `zone_change.go`, add `gs.Flags["permanent_left_bf"]++` alongside `Turn.PermanentsLeft++`. Or migrate the evaluator to read `Turn.PermanentsLeft`.

---

## Files Changed

| File | Changes |
|------|---------|
| `cmd/hexdek-thor/conditional_setup.go` | 7 new scaffold kinds, TurnCounters dual-write on 8 existing kinds, `secondSpellRe` regex |
| `cmd/hexdek-thor/goldilocks.go` | ability_word setup: added Turn.CreaturesDied, Turn.PermanentsLeft, Turn.Attacked, permanent_left_bf flag |
