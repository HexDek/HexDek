# State-Based Actions Audit — R37

**Date:** 2026-05-17
**Source:** `internal/gameengine/sba.go` (1862 LOC) vs. CR §704.5 / §704.6 from `data/rules/MagicCompRules-20260227.txt`.
**Scope:** Every §704.5 sub-letter (a–z, skipping l and o per CR preamble) plus §704.6c / §704.6d Commander variants.

> **Note on rule lettering.** The task prompt cited §704.5a–u with a non-current letter→rule mapping (e.g. "§704.5b = 10-poison", "§704.5c = legend rule"). The 2026-02-27 CR file in-repo uses the canonical lettering below; that is what `sba.go` targets and what this audit follows. The prompt's intent — verifying every state-based action — is preserved.

## Summary

| Status | Count | Rules |
|--------|-------|-------|
| Implemented (fully wired) | 16 | 704.5a, 704.5b, 704.5c, 704.5d, 704.5e, 704.5f, 704.5g, 704.5i, 704.5j, 704.5k, 704.5m, 704.5n, 704.5p, 704.5q, 704.5v, 704.5y, 704.6c, 704.6d |
| Implemented but data-source partial | 4 | 704.5h (deathtouch flag), 704.5r (oracle parse), 704.5s (saga ETB), 704.5z (Start-Your-Engines) |
| Implemented as best-effort heuristic | 2 | 704.5t (dungeon), 704.5w (battle protector), 704.5x (siege protector) |
| Pure stub | 1 | 704.5u (Space Sculptor) |
| N/A (letter skipped by CR) | 2 | 704.5l, 704.5o |

The §704.6a (mutate count limit) and §704.6b (host-and-augment) Commander-format clauses are out of scope for the engine (no mutate / host cards in pool) and not implemented; flagged only.

## Per-rule status

### §704.5a — Player at 0 or less life loses
**Status:** Implemented (`sba704_5a`, sba.go:247).
**Notes:** Fires `would_lose_game` replacement (Platinum Angel etc.) via `FireLoseGameEvent`. `SBA704_5a_emitted` is used as spam-suppression only; the gate is `!s.Lost`. Resets when life climbs back above 0. Solid.

### §704.5b — Drew from empty library since last SBA check
**Status:** Implemented (`sba704_5b`, sba.go:303).
**Notes:** Consumes and clears `Seat.AttemptedEmptyDraw`. Correct: the flag is set by the draw path when the library is empty rather than by checking library size at SBA time (CR is unambiguous that it is the attempt, not the empty state, that triggers loss — Laboratory Maniac decks pass because they replace the draw before the attempted-empty-draw is recorded).

### §704.5c — 10 or more poison counters
**Status:** Implemented (`sba704_5c`, sba.go:334).
**Notes:** Standard threshold check on `Seat.PoisonCounters`. No replacement-effect path (Solemnity, Melira) is wired here, but those operate at counter-placement time, not at the SBA, so this is correct.

### §704.5d — Token in non-battlefield zone ceases to exist
**Status:** Implemented (`sba704_5d`, sba.go:368).
**Notes:** Sweeps hand, graveyard, exile, library each pass. Token identity via `cardIsToken` checks `Card.Types` for the `"token"` entry. Caveat: this won't catch tokens that are reanimated (`Embalm`-style tokens of nontoken cards) until the next SBA pass after they leave the battlefield — fine, because death-trigger ordering already runs SBAs after stack resolution.

### §704.5e — Spell/card copies outside legal zones cease to exist
**Status:** Implemented (`sba704_5e`, sba.go:443).
**Notes:** Sweeps all four nonbattlefield zones for `Card.IsCopy`. Stack copies are not swept (correct).

### §704.5f — Creature with 0 or less toughness goes to graveyard
**Status:** Implemented (`sba704_5f`, sba.go:523).
**Notes:** Uses `gs.IsCreatureOf` and `gs.ToughnessOf` — layer-aware, so Opalescence-granted creatures are caught. Phased-out permanents skipped (§702.26). Regeneration explicitly cannot replace this event (passed through `destroyPermSBA` which fires §614 `would_die` but the resolver does not allow regeneration replacements for §704.5f — verify in `FireDieEvent`).
**Suggested follow-up:** Confirm that `FireDieEvent` rejects regeneration replacements when the death cause is `toughness_zero_or_less`. CR §704.5f explicitly says regen can't replace.

### §704.5g — Creature with lethal damage marked is destroyed
**Status:** Implemented (`sba704_5g`, sba.go:556).
**Notes:** Indestructible (`p.IsIndestructible()`) short-circuits per §702.12b. Layer-aware toughness query.

### §704.5h — Creature damaged by deathtouch source is destroyed
**Status:** Implemented but flag-dependent (`sba704_5h`, sba.go:596).
**Notes:** Reads `Permanent.Flags["deathtouch_damaged"]` set by combat (`combat.go:1634`) and by fight/bite resolution (`resolve.go:1636,1647`). Indestructible respected. Flag cleared on consume.
**Gap:** Non-combat damage sources (Pestilence, ping abilities, "deals damage" replacement effects, Lifelink-with-Deathtouch from triggered abilities) must also set the flag. Verify every code path that calls `DealDamage(...)` to a creature with `deathtouch` propagates the flag. Recommend introducing a single helper (e.g. `ApplyDamageToPermanent`) that owns the flag write so individual call sites can't forget.

### §704.5i — Planeswalker at 0 loyalty goes to graveyard
**Status:** Implemented (`sba704_5i`, sba.go:646).
**Notes:** Skips planeswalkers whose `Counters` map has no `"loyalty"` key (treated as ETB-pending). Once a value is recorded, ≤ 0 dies. Correct.

### §704.5j — Legend rule
**Status:** Implemented (`sba704_5j`, sba.go:684).
**Notes:** Per-controller grouping by `Card.DisplayName()`. Keeper = lowest timestamp (first-cast survives). CR allows the controller to pick; "earliest" is a deterministic policy mirroring Python reference. Acceptable for engine determinism; AI gets no choice but loses no information.
**Future polish:** Expose a choice hook for human players (`ChooseLegendKeeper`) so a Sun Titan flicker shenanigans player can pick the newer copy if desired.

### §704.5k — World rule
**Status:** Implemented (`sba704_5k`, sba.go:754).
**Notes:** Keeper = highest timestamp (shortest time as world). Tie → all die. Matches CR. Engine doesn't track gain-type effects, so timestamp == ETB is the proxy for "time as world." Fine until layer-changing world-grant effects exist in the pool (none currently).

### §704.5l — *(skipped letter, intentionally per CR preamble)*

### §704.5m — Aura illegally attached
**Status:** Implemented (`sba704_5m`, sba.go:818).
**Notes:** Treats any Aura on the battlefield with no `AttachedTo` (or whose target left) as illegal — including auras that were never attached. This is more aggressive than the original `aura_expects_attach` gate; correct per CR which makes no exception for "newly-resolved" auras.
**Gap:** Does not verify the *aura's enchant-restriction* (e.g. an Aura with `enchant creature` attached to a battle, a Pacifism on an indestructible-and-pacifism-immune creature). The aura's enchant clause should be checked against the target's current types/abilities. Suggested fix: extend with `auraTargetLegal(aura, target)` that consults the aura's parsed `enchant` AST.

### §704.5n — Equipment / Fortification illegally attached → unattach
**Status:** Implemented (`sba704_5n`, sba.go:849).
**Notes:** Detaches when (1) target gone, (2) Equipment attached to non-creature, (3) Fortification attached to non-land. CR-accurate.
**Gap:** Doesn't handle the "Equipment attached to a creature its controller does not control" case (Equipment can be attached cross-controller via Sword of the Animist tricks, control-change, etc.) — CR §301.5 actually does allow that. Currently the engine matches CR. No fix needed; documented for future readers.

### §704.5o — *(skipped letter, intentionally per CR preamble)*

### §704.5p — Battle/creature/other illegally attached → unattach
**Status:** Implemented (`sba704_5p`, sba.go:904).
**Notes:** Detaches any non-Aura, non-Equipment, non-Fortification permanent that has `AttachedTo != nil`. Matches CR.

### §704.5q — +1/+1 and -1/-1 counters annihilate
**Status:** Implemented (`sba704_5q`, sba.go:944).
**Notes:** Standard min(plus, minus) annihilation; deletes zero entries to keep the map clean.

### §704.5r — Counter cap from oracle text
**Status:** Implemented with limited parser (`sba704_5r`, sba.go:997).
**Notes:** Walks `Card.AST.Abilities` for `Static.Raw` containing `"can't have more than N <kind>"`. `parseCounterLimit` (sba.go:1051) recognizes number words one–fifteen plus digits.
**Gaps:**
- Recognises only the literal `"can't have more than"` phrasing. CR allows equivalent paraphrases ("This creature can't have more than two ki counters on it" works; "no more than" does not).
- Bails when the third token is `"counters"` (treats it as malformed). For phrasings like `"can't have more than two counters on it"` (no kind specified) the bail is correct, but it would also bail on `"can't have more than 2 counters"` written by a card-corpus typo — acceptable.
- Multi-word counter kinds (`"all counters"` is not a kind; `"feather counter"` is) — currently single-word kind only.
**Suggested fix:** Move parsing out of the SBA hot path: have the AST loader pre-compute a per-card `MaxCountersByKind map[string]int` during card load (Thor) and have §704.5r consult it directly. Lets SBA stay O(1) per perm and lets Thor enforce parse coverage during corpus build.

### §704.5s — Saga past final chapter sacrifices
**Status:** Implemented (`sba704_5s`, sba.go:1091); fed by `etb_dispatch.go` `initSagaLoreCounters`.
**Notes:** ETB hook walks the AST for `saga_chapter` modifications and sets `Counters["saga_final_chapter"]`. SBA then compares `Counters["lore"]` against it. The "is the source of a chapter ability still on the stack" clause is NOT checked — if the chapter ability is still triggering, the saga still dies one pass too early.
**Suggested fix:** Add a `Saga.PendingChapterAbilities int` counter incremented at trigger fire / decremented on resolution; skip §704.5s while > 0. Mirrors the parallel "battle/dungeon source-of-triggered-ability" clauses elsewhere in §704.5.

### §704.5t — Completed dungeon removed from game
**Status:** Best-effort heuristic (`sba704_5t`, sba.go:1133).
**Notes:** Reads `Seat.Flags["dungeon_level"]` and `dungeon_name` (1 → Mad Mage, 7 rooms; otherwise 4 rooms). On completion logs `dungeon_completed` and clears the flags. Dungeons are not modeled as command-zone cards — `data/rules/MagicCompRules-20260227.txt` permits this since dungeons are not a real zone occupant for our card pool.
**Gap:** Increments may be coming from two paths (`keywords_misc.go:1617` increments `dungeon_completed` directly; `resolve_helpers.go:1103, 4143` increments `dungeon_level`). The two paths are inconsistent. Recommended: standardize on `dungeon_level` and have completion detection live solely in `sba704_5t`.

### §704.5u — Space Sculptor sector designation
**Status:** Pure stub (`sba704_5u`, sba.go:1196 returns false).
**Notes:** Unfinity mechanic. Not in the 4-deck tournament pool. Leaving as stub is correct.
**Suggested fix path (only if Unfinity ever lands):** Add `Permanent.Sector string` plus a `Permanent.IsSpaceSculptor() bool`; on SBA, for any creature with empty `Sector` on a battlefield containing a space sculptor, prompt sector assignment from the appropriate seat per CR §704.5u ordering (non-sculptor controllers first, then sculptor controllers).

### §704.5v — Battle with 0 defense
**Status:** Implemented (`sba704_5v`, sba.go:1209).
**Notes:** Reads `Counters["defense"]` and destroys on ≤ 0. Uses `destroyPermSBA` (which fires §614 `would_die`, allowing exile redirects). The "is the source of an ability that has triggered but not yet left the stack" clause is not modeled — same gap as §704.5s. For battles this typically matters when the defense-reducing trigger and the chapter-style trigger are on the stack simultaneously.
**Suggested fix:** Same pending-trigger counter pattern proposed for §704.5s.

### §704.5w — Battle without protector
**Status:** Implemented heuristically (`sba704_5w`, sba.go:1245).
**Notes:** Verifies `Flags["protector_seat"]` still in game; if no protector, assigns the first living opponent of the battle's controller. If no opponent available, sacrifices the battle.
**Gaps:**
- Does not respect battle-type-specific protector selection rules. Siege protectors must be opponents of the controller; future battle types (e.g. "battle — quest") may have different valid-protector populations. Currently treats all battle subtypes identically.
- The "controller's choice" is replaced with "first opponent." CR allows the controller to choose; deterministic policy is acceptable but should be surfaced as a hook for AI policy.
- Does not pick "attacking creatures' controllers" as protectors when applicable per the CR phrasing.

### §704.5x — Siege controller == protector → reassign
**Status:** Implemented (`sba704_5x`, sba.go:1324).
**Notes:** Detects Siege subtype via `Card.Types` membership (`"siege"`/`"Siege"`). Reassigns to first living opponent if controller == protector. Sacrifices on no available opponent.
**Gap:** Identical limitation as §704.5w — no controller-choice hook.

### §704.5y — Multiple Roles on same permanent from same player
**Status:** Implemented (`sba704_5y`, sba.go:1403).
**Notes:** Per-(holder, controller) grouping; keeper = highest timestamp (newest). Older Roles die. CR-accurate. Iterates all seats' battlefields to find Roles attached anywhere — correct because Roles can be controlled by a different player than the holder.

### §704.5z — Start Your Engines! → speed = 1
**Status:** Implemented (`sba704_5z`, sba.go:1455).
**Notes:** Checks two paths: (1) any permanent with `"start your engines"` in `Card.Types` or `Flags["start_your_engines"] > 0`; (2) game-global `gs.Flags["start_your_engines"]`. Sets `Seat.Speed = 1` (and a parallel `Seat.Flags["speed"]` for the game-global path).
**Gap:** Two parallel speed-tracking surfaces (`Seat.Speed int` and `Seat.Flags["speed"]`) — pick one. Recommend `Seat.Speed`. The remainder of the engine has no speed-decrement / speed-cap / max-speed-4 enforcement, but that lives in damage-event handlers not SBAs.

---

### §704.6a — Commander mutate stack limit
**Status:** Not implemented. Mutate is not in pool. Flag only.

### §704.6b — Commander host/augment legality
**Status:** Not implemented. Unstable cards not in pool. Flag only.

### §704.6c — 21+ commander damage
**Status:** Implemented (`sba704_6c`, sba.go:1547).
**Notes:** Per-(dealer, commanderName) bucket; partner pair (Kraum+Tymna) accrue independently. Logs `sba_704_6c`. CR-accurate.

### §704.6d — Commander in graveyard/exile may go to command zone
**Status:** Implemented (`sba704_6d`, sba.go:1598).
**Notes:** Greedy-allow policy (every matching commander returns). Avoids double-append via `inCommandZone` check. CR allows the owner to *choose*; the engine takes the choice for them. For deck strategies that want to leave the commander in the graveyard (e.g. Karador/Muldrotha decks recasting from graveyard via the deck's own mechanics, not commander tax) this is the wrong default.
**Suggested fix:** Expose `Seat.CommanderReturnPolicy` (default `"always"`, alternatives `"never"` or per-instance prompt) so AI policies can keep the commander in graveyard for Muldrotha-style synergies. Currently this never matters for the tournament pool but will matter for graveyard-recursion decks.

---

## Implementation-quality notes (cross-cutting)

1. **§614 replacement coverage.** `destroyPermSBA` runs `FireDieEvent` (correct for §704.5f/g/i/m/v/y). `sacrificePermSBA` also runs `FireDieEvent` (sba.go:1786) — verify Saga sacrifice is the right surface for "would die" replacement (CR treats sacrifice as a separate event; `dies` triggers fire from either, but Anafenza-style exile-replacement is `dies`-specific). Currently both flows hit the same replacement chain. Acceptable, but document the semantics.
2. **Phased-out skipping.** §704.5f/g/h/i/m/v explicitly skip `PhasedOut`. §704.5j/k/n/p/q/r/s/y do *not*. Per CR §702.26 phased-out permanents are treated as though they don't exist for SBA purposes — extend the skip to the other helpers. Concretely: a phased-out legend should NOT trigger the legend rule against an in-phase copy of itself. Verify with a test.
3. **Pass cap (40) → game draw.** The infinite-loop fallback (sba.go:158) flags `game_draw` when 40 SBA passes don't quiesce. CR §104.4b is the right rule but the trigger condition is overly broad — 40 cascading deaths from a board wipe can legitimately consume many passes if attach detach cycles produce reactive triggers. Recommend raising the cap to 100 and instrumenting a `passes` histogram in `Event{Kind: "sba_cap_hit"}` so we can tune empirically.
4. **`ManaPool` reconciliation at sba.go:186–228.** This block doesn't implement an SBA — it reconciles legacy `s.ManaPool int` with the typed `s.Mana` struct. It runs every SBA call even when no SBA fired. Move it out of `StateBasedActions` to wherever mana-pool drift is actually being introduced (cleanup phase, end-of-step), or skip it when `!anyChange`. Currently it's hot-path overhead.

## Suggested order of attack

1. **Verify §704.5h coverage** — single source of truth for the deathtouch flag. (~1 day; medium-confidence.)
2. **Add pending-trigger guard for §704.5s / §704.5v** — saga/battle "source of triggered ability still on stack" clauses. (~1 day; high-confidence.)
3. **Pre-compute counter caps for §704.5r** during Thor card load; drop the runtime substring parser. (~½ day; high-confidence.)
4. **Phased-out skipping audit** across §704.5j/k/n/p/q/r/s/y. (~2 hours plus tests.)
5. **Mana-pool reconciliation extraction** out of `StateBasedActions`. (~1 hour.)
6. **Aura enchant-restriction check** for §704.5m. (~½ day; needs AST hook.)
7. **§704.6d return-policy hook** for graveyard-synergy commander decks. (~½ day; coordinated with AI policy.)
