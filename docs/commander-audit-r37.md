# Commander Format Audit — CR §903 (r37)

**Scope.** Engine coverage of the Commander variant's seven load-bearing
sub-rules:

| § | Topic | Locus |
|---|-------|-------|
| §903.4 | Color identity (deckbuild constraint) | `cmd/hexdek-freya/legality.go::checkColorIdentity` |
| §903.5 | 1 commander + 99 singleton (basic-land exempt) | `cmd/hexdek-freya/legality.go::checkCardCount`, `checkSingleton` |
| §903.8 | Commander tax (+{2} per previous cast) | `internal/gameengine/commander.go::CastCommanderFromCommandZone`, `CommanderCastCost` |
| §903.9a | Graveyard/exile → command zone (SBA, "may") | `internal/gameengine/sba.go::sba704_6d` |
| §903.9b | Hand/library → command zone (replacement, "may") | `internal/gameengine/commander.go::registerCommanderZoneReplacement` |
| §903.10a | 21+ commander damage → loss (SBA) | `internal/gameengine/sba.go::sba704_6c`, `commander.go::AccumulateCommanderDamage` |
| Banlist | RC-maintained, NOT in CR §903.12 | `cmd/hexdek-freya/legality.go::commanderBannedList` + `checkBanned` |

> Terminology note. The audit prompt cites "§903.12 banned cards check"
> — in CR `MagicCompRules-20260227.txt` line 6904, §903.12 is the
> **Brawl Option**, not a banlist. The Commander banlist is maintained
> by the Commander Rules Committee outside CR. The engine follows that
> convention: banlist enforcement lives in Freya as a deckbuild check.

---

## §903.4 — Color identity ✅ COVERED (deckbuild-time)

Color identity is a deck construction invariant (CR §903.4a: "Color
identity is established before the game begins"). The engine does not
re-validate at runtime, which is correct: once `SetupCommanderGame`
seeds the seat, color-identity violations are impossible.

| Sub-rule | Implementation | Status |
|----------|----------------|--------|
| §903.4   | Scryfall `color_identity` field via `oracleDB.lookup(...).ColorIdentity` | ✅ |
| §903.4a  | Snapshot at deck load; no runtime mutation | ✅ |
| §903.4b  | Color-choice commanders not modeled | ⚠️ Gap (rare) |
| §903.4c  | Reminder text already stripped by Scryfall | ✅ (delegated) |
| §903.4d  | DFC back face included by Scryfall | ✅ (delegated) |
| §903.4e  | Adventurer alt-characteristics aggregated by Scryfall | ✅ (delegated) |
| §903.4f  | "your commander's colors" undefined when no commander | n/a (always present) |

`checkColorIdentity` (legality.go:246) iterates the 99 and rejects any
card with a color in its identity not in the commander's identity.
**Partner color-identity union (§702.124c) is not combined** — single-
commander assumption; partner decks get a false-negative if the second
commander widens the allowed set.

## §903.5 — Deck construction (1+99, singleton, basic exempt) ✅ COVERED

| Sub-rule | Implementation | Status |
|----------|----------------|--------|
| §903.5a  | `checkCardCount` against `expected = 100`, partner-aware (98+2) | ✅ |
| §903.5b  | `checkSingleton` exempts 5 basics + wastes, snow-covered variants, "any number of" oracle text | ✅ |
| §903.5c  | Forwarded to `checkColorIdentity` | ✅ |
| §903.5d  | Implicitly via Scryfall `color_identity` (basic-land subtype colors already included) | ✅ |
| §903.5e  | No sideboard concept in engine | ✅ (n/a) |

**Caveat (§903.5b):** singleton check uses `qp.Profile.Name` without
normalising interchangeable names (CR §201.3). For Scryfall-sourced
data the lookup canonicalises; hand-assembled decks with set codes
baked into the name can false-positive. Low priority.

## §903.8 — Commander tax ✅ COMPREHENSIVE

`CastCommanderFromCommandZone` (commander.go:350) computes
`totalCost = baseCMC + 2 × seat.CommanderCastCounts[name]` and
increments the per-name counter **after** successful payment (matching
CR's "previous time" wording — the Nth cast pays 2(N−1)).

| Edge case | Test |
|-----------|------|
| First cast: 0 tax | `TestCastCommanderFromCommandZone_FirstCastZeroTax` |
| Nth cast: 2(N-1) | `TestCastCommanderFromCommandZone_SecondCastCostsPlusTwo` |
| Hand cast (not from CZ) → no tax | `TestHandCastCommander_NoTaxIncrement` |
| Partner pair → independent counters (CR §702.124d) | `TestCommanderCastCost_*` |
| Tax persists across zone cycles | `TestCommanderTax_PersistsAcrossZoneCycles` |
| Drannith Magistrate blocks CZ cast (§601.2a) | inline `drannithBlocksZoneCast` |
| DFC back-face recast (MDFC Esika→Bridge) | `DFCCardMatchesName` at commander.go:375 |
| Cast triggers fire (Storm, Rhystic, Esper Sentinel) | `IncrementCastCount` + `RecordCast` + `fireCastTriggersFromZone` + `FireCastTriggerObservers` |

No actionable gaps in §903.8.

## §903.9a — Graveyard/exile → command zone (SBA) ⚠️ GREEDY-ALLOW

`sba704_6d` (sba.go:1598) sweeps each seat's graveyard and exile every
SBA pass; any commander found is **unconditionally** moved to the
command zone.

CR §903.9a / §704.6d says *"its owner **may** put it into the command
zone"*. The "may" is a player choice. Engine policy is greedy-allow —
documented design call (sba.go:1595) but a deviation from CR. See
Issue #2.

Idempotency: the SBA correctly skips commanders already moved to the
command zone via `inCommandZone(c)`.

## §903.9b — Hand/library → command zone (replacement) ⚠️ DFC GAP

`registerCommanderZoneReplacement` (commander.go:199) wires a §614
replacement keyed on `(card_name, owner_seat, to_zone ∈ {hand,
library, library_top, library_bottom})`. On match, payload `to_zone`
is rewritten to `command_zone`.

| Sub-rule | Implementation | Status |
|----------|----------------|--------|
| §903.9b | Hand / library destinations | All four matched | ✅ |
| §903.9b | From anywhere | No `from_zone` filter | ✅ |
| §903.9b | Owner-keyed (CR §108.3) | `ev.Payload["owner_seat"] != ownerSeat` | ✅ |
| §903.9b | "May apply more than once" — exception to §614.5 | HandlerID intentionally **not** appended to `AppliedIDs` | ✅ |
| §903.9b | Player chooses ("may") | Engine always redirects | ⚠️ Policy (Issue #2) |
| §903.9b | DFC back-face commander | Raw string equality fails after Card.Name swap | ❌ **Issue #1** |
| §903.9c | Meld / merge components | No special handling | ⚠️ Gap (rare cards) |

Gilded Drake ownership preservation is covered
(`TestGildedDrake_OwnershipSurvivesControlSwap`): the replacement
keys off `Permanent.Owner`, not `Permanent.Controller`, so a traded
commander still redirects to the *original* owner's command zone.

## §903.10a — 21 commander damage ✅ COVERED

`sba704_6c` (sba.go:1547) walks `Seat.CommanderDamage`
(`map[dealerSeat]map[commanderName]int`) and triggers loss when any
single (dealer, name) bucket reaches 21.

| Sub-rule | Implementation | Status |
|----------|----------------|--------|
| §903.10a | 21+ from same commander → loss | Threshold + per-name bucketing | ✅ |
| §903.10a | "same commander" persists across zones | Bucketed by name, not by `*Card` pointer | ✅ |
| §702.124d | Partner: each commander tracked separately | Nested map keys per-name | ✅ |
| §903.10a | Combat damage only | `AccumulateCommanderDamage` called only from combat damage paths | ✅ |
| §903.10a | Dealer indexed | Outer map key is `dealerSeat` | ✅ |

Coverage is tight. The combat damage handler is the only writer; the
SBA is the only reader; 3 dedicated tests + integration coverage.

## Banlist — ✅ COVERED (deckbuild-time, ~47 entries)

`commanderBannedList` (legality.go:76) is a flat `map[string]bool`
keyed on lowercased oracle names. `checkBanned` normalises curly
quotes (`foldQuotes`), resolves through `oracle.lookup` to
canonicalise, and reports any match.

| Concern | Status |
|---------|--------|
| Up-to-date through April 2026 (Nadu, Dockside, Jeweled Lotus) | ✅ |
| Commander-specific bans (Erayo, Leovold, Rofellos, Hullbreacher) | ✅ |
| Pre-Modern bans (Channel, Ancestral, Black Lotus, all 5 Moxen) | ✅ |
| Bracket banlist (April 2025: Mana Crypt, Sol Ring banned at low brackets) | ❌ Not modeled |
| Banned-as-commander vs. banned-anywhere distinction | ❌ Not modeled |
| New-set adds (post-April 2026) | ⚠️ Manual maintenance |

Engine intentionally treats the banlist as a soft deckbuild warning,
not a runtime invariant — desirable so legacy / training decks with
banned cards still run.

---

## Top 3 fixable issues

### Issue #1 — §903.9b: DFC back-face commander bypasses zone redirect (HIGH)

**File:** `internal/gameengine/commander.go:211-223`
(`registerCommanderZoneReplacement.Applies`).

The replacement's match predicate uses raw string equality:

```go
if ev.String("card_name") != commanderName {
    return false
}
```

`commanderName` is the oracle name registered at setup (e.g.
`"Esika, God of the Tree // The Prismatic Bridge"`). When the back
face is cast and resolves, `Card.Name` is mutated to the back-face-
only name (per stack.go DFC handling — see in-source comment at
commander.go:373 confirming this swap). A `would_change_zone` event
for the back-face card carries `card_name = "The Prismatic Bridge"`
in its payload. `"The Prismatic Bridge" != "Esika... // The Prismatic
Bridge"` → predicate returns false → §903.9b **does not fire** → the
commander goes to hand or library instead of the command zone.

`IsCommanderCard` already solves the same recognition problem
correctly via `DFCCardMatchesName` (commander.go:505). The fix is to
reuse the same comparison here. Either factor out a
`nameMatchesCommander(eventName, registeredName string) bool` helper
(mirroring `DFCCardMatchesName`'s face-pair logic without requiring a
`*Card`), or pass the `*Card` pointer through the payload and call
`DFCCardMatchesName(card, commanderName)` directly. The helper route
is cleaner — payload-pointer route requires auditing every
`would_change_zone` caller.

**Test gap:** no test covers "back-face DFC commander bounced to
hand". Add a regression analogous to
`TestCommanderReturnToHand_RedirectsToCommandZone` with an MDFC
commander cast on its back face, then bounced via Boomerang.

### Issue #2 — §903.9a / §903.9b: greedy-allow forecloses player choice (MEDIUM)

**Files:**
- `internal/gameengine/sba.go:1598` (`sba704_6d`)
- `internal/gameengine/commander.go:225-238` (`registerCommanderZoneReplacement.ApplyFn`)

Both §903.9a (SBA) and §903.9b (replacement) are CR "may" effects —
the owner *chooses* whether to return the commander. The engine
implements both as unconditional auto-return.

For graveyard-recursion archetypes this is materially wrong:

- **Karador, Ghost Chieftain** wants the commander in the graveyard
  for graveyard-based casting (his own ability cares about creatures
  in graveyard).
- **Muldrotha, the Gravetide** wants permanents in graveyard so she
  can replay them — including herself if she dies cheap.
- **Meren of Clan Nel Toth**, **The Mimeoplasm**, reanimator
  packages — same logic.

Hat's existing `hasGraveyardRecursionValue` helper (added 2026-05-08,
see Issue Log) confirms the AI already wants this kind of decision;
the engine simply doesn't expose the choice.

**Fix shape:** per-seat policy hook:

```go
type CommanderReturnPolicy func(gs *GameState, seatIdx int,
    commanderName, fromZone string) bool

// On Seat:
ReturnFromGraveyardPolicy CommanderReturnPolicy
ReturnFromExilePolicy     CommanderReturnPolicy
ReturnFromHandPolicy      CommanderReturnPolicy  // 903.9b
ReturnFromLibraryPolicy   CommanderReturnPolicy  // 903.9b
```

Default to `func(...) bool { return true }` for backward compat.
`sba704_6d` and the replacement consult the relevant policy before
moving. Freya already classifies Reanimator-archetype decks; route
the classification into the seat's policy field at game setup so
Karador/Muldrotha decks return false from graveyard policy.

### Issue #3 — §903.5a partner detection misclassifies non-Partner-keyword pairs (LOW)

**File:** `cmd/hexdek-freya/legality.go:152-163` (in `CheckLegality`).

Partner detection uses `strings.Contains(oracleText, "partner")` —
catches:
- ✅ "Partner" (bare keyword)
- ✅ "Partner with X"
- ❌ "Friends forever" — no "partner" substring
- ❌ "Choose a Background" — no "partner" substring
- ❌ "Doctor's companion" — "companion" without "partner"

A Background-paired deck (e.g., Faceless Lord of Loss + Cultist of
the Absolute) has 98+2 = 100 cards but `hasPartner = false` sets
`expected = 99+1 = 100` — same number, so `checkCardCount` passes by
coincidence. The partner-aware message (`"100 cards (98 + 2 partner
commanders)"` vs `"100 cards (99 + 1 commander)"`) reports the wrong
configuration, and downstream tooling that branches on
`CardCount.HasPartner` gets a false negative.

`internal/gameengine/multiplayer.go:513` already has
`ReadPartnerInfo(card *Card)` that correctly parses all five
partner-family keywords from a card's AST. The Freya legality layer
should call that (or duplicate its substring set) rather than do its
own naive scan.

**Fix:** in `CheckLegality`, replace the `strings.Contains("partner")`
sniff with checks across `partner`, `partner with`, `friends
forever`, `choose a background`, `doctor's companion`, plus type-line
`background` / `doctor` checks on the commander entry. If the
deckparser surfaces the second commander in `qtyProfiles` (Moxfield
format), trust that count over substring inference.

---

## Test coverage summary

| File | Tests | Notes |
|------|-------|-------|
| `commander_test.go` | 26 | Setup, tax (zero/scaling/persistence/hand-cast no-tax), §903.9b hand+library redirect, §903.9a graveyard+exile SBA, 21-damage loss, partner separation, Gilded Drake ownership, IsCommanderCard, DFC name match |
| `commander_damage_combat_test.go` | 3 | Combat-driven damage accumulation, non-commander non-accumulation, partner per-key separation |
| `multiplayer_test.go` | 20 | APNAP, opponent iteration, seat elimination (idempotency, stack purge, owned vs. stolen permanents), CheckEnd (continue / single-winner / draw) |

**Total: 49 tests** across the three files covering §903 + §800 +
§101.4.

**Untested edges:**
1. DFC back-face commander bounce (Issue #1).
2. Meld/merge commander redirect (§903.9c — Brisela returning to CZ
   should leave non-commander half in owner's graveyard / battlefield).
3. Replacement re-application — "may apply more than once" exception
   to §614.5 has no dedicated test (commander bounced from hand by
   Boomerang chain after a Path-to-Exile redirect).
4. Background / Doctor-companion / Friends-Forever partner detection
   in Freya legality (Issue #3).

---

## Compliance score

| § | Sub-rules audited | Covered | Score |
|---|---|---|---|
| 903.4 (color identity) | 7 | 6 | 86% |
| 903.5 (deck construction) | 5 | 5 | 100% |
| 903.8 (commander tax) | 8 edges | 8 | 100% |
| 903.9a (graveyard/exile SBA) | 1 (`may` semantics) | 0 (greedy) | 0% (policy) |
| 903.9b (hand/library replacement) | 6 | 4 | 67% (Issues #1, #2) |
| 903.9c (meld/merge) | 1 | 0 | 0% (gap, rare) |
| 903.10a (21 commander damage) | 5 | 5 | 100% |
| Banlist (RC) | 1 | 1 | 100% |

**Weighted (excluding §903.9c rarity):** ≈ 87% conformance. Issue #1
(DFC redirect) closes the only correctness gap; Issues #2 and #3 are
policy / UX improvements.
