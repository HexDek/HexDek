# Zone Accounting — Feynman Violation Analysis

**Date:** 2026-05-04
**Source log:** `josh@192.168.1.207:/tmp/hexdek-server.log` (~1 hour grinder run, 18:07–19:08 UTC).

## Headline numbers

| Metric                            | Value     |
| --------------------------------- | --------- |
| Total Feynman reports             | **11 774** |
| `zone_accounting` violations      | **9 135**  |
| Violation rate                    | **~78 %** of post-game checks flagged at least one seat |
| Span                              | 1 h 1 min |

## Violation-message shape

The `checkZoneAccounting` invariant lives in `internal/hat/feynman.go:174–229`. It compares the **observed** card count (`hand + library + graveyard + exile + non-token battlefield + command_zone`) to the **expected** count (`99 + len(CommanderNames)`, default 100, minus `seat.Flags["cards_left_game"]`). Tolerance is ±3.

Empirical diff distribution from the last 200 violations:

```
+4   39    +5   23    +6   28    +7   12    +8   16    +9    7
+10   4    +11   4    +12   5    +13   4
+14   3    +15   2    +16   2    +17   4    +18   3    +19   3
+20   2    +21   5    +24   1    +25   1    +36   1    +37   1
+84   1    +255  1
−4    8    −5    7    −6    6    −7    1    −8    2    −9    2
```

- **~80 % of violations are positive (overcounts → duplication).**
- **~20 % are negative (undercounts → loss).**
- Long tail: a few games hit `+255` and `+84`, suggesting a rare runaway path (probably reanimator chain).

## Hypothesis #1 — duplication via `MoveCard("battlefield") + enterBattlefieldWithETB`

This is the **dominant overcount source**.

### The bug

`gameengine.MoveCard` is the universal zone-change entry point (`internal/gameengine/zone_move.go`). For a `toZone == "battlefield"` move, it ultimately calls `gs.moveToZone(seat, card, "battlefield")` — but **`moveToZone` (state.go:1091) has no `"battlefield"` case**:

```go
switch zone {
case "hand":          ...
case "graveyard":     ...
case "exile":         ...
case "library_top":   ...
case "library_bottom":...
default:
    s.Graveyard = append(s.Graveyard, c)   // <-- silent fallthrough
}
```

So a call like `MoveCard(gs, c, seat, "graveyard", "battlefield", "...")`:
1. Removes `c` from graveyard (correct).
2. Falls into `default` and **re-appends `c` to the graveyard** (silent).

By itself, that's a correctness bug but not a duplication one — net zero, the card stays in graveyard.

The duplication appears when a per-card hook **chains** `MoveCard` with `enterBattlefieldWithETB`:

```go
// internal/gameengine/per_card/sidar_jabari_of_zhalfir.go:151
gameengine.MoveCard(gs, best, perm.Controller, "graveyard", "battlefield", "sidar_jabari_reanimate")
enterBattlefieldWithETB(gs, perm.Controller, best, false)
```

After both lines run:
- `best` is back in `seat.Graveyard` (via the `MoveCard` fallthrough).
- A new `Permanent{Card: best, …}` is on `seat.Battlefield` (via `enterBattlefieldWithETB → createPermanent`).

The Feynman check counts **both**: `len(s.Graveyard)` includes `best`, and the battlefield loop also counts the Permanent (`Card == best`). One reanimation = `+1` overcount.

### Affected call sites (engine + per-card)

`grep -rln 'MoveCard.*"battlefield"' internal/gameengine/` returns **75 hits**, of which **34 per-card files** also append/`enterBattlefieldWithETB` in the same handler. A non-exhaustive sample:

```
sidar_jabari_of_zhalfir.go      meren.go                 tayam.go
gen_eddie_brock_venom_lethal_protector.go   gerrard_weatherlight_hero.go
nicol_bolas.go                  athreos_shroud_veiled.go enchantment_toolbox.go
sefris_of_the_hidden_ways.go    excava_the_risen_past.go dr_eggman.go
sheoldred_true_scriptures.go    ayesha_tanaka_armorer.go bre_of_clan_stoutarm.go
moseo_veins_new_dean.go         squall_seed_mercenary.go shirei_shizos_caretaker.go
jodah_the_unifier.go            dr_madison_li.go         lord_windgrace.go
betor_ancestors_voice.go        sauron_lord_of_the_rings.go nine_fingers_keene.go
etali.go                        edea_possessed_sorceress.go terra_herald_of_hope.go
kenrith_returned_king.go        light_paws_emperors_voice.go gen_mayael_the_anima.go
tevesh_szat.go
```

Each activation of one of these abilities adds `+1` to the violator seat's overcount. A typical Meren game with 3–5 reanimations easily produces `diff=+5` to `+10`; a Sefris/Sidar-Jabari pile-up explains the `+13`/`+17` cluster; the rare `+84`/`+255` outliers are almost certainly Sauron-style infinite-recursion pathological games.

Engine-level (non-per-card) call sites with the same shape:
- `resolve_helpers.go:3977` — `MoveCard(c, seat, "hand", "battlefield", "cheat_creature")` (Sneak Attack family, Hypergenesis, etc.).
- `keywords_batch4.go:1001` — `MoveCard(exiledCard, ..., "battlefield", "exile", "champion")` — *different* bug: this is a battlefield→exile move; `MoveCard` is documented as a no-op for `fromZone="battlefield"`, so champion likely doesn't actually exile the championed creature.

### Recommended fix direction

1. **Add a `"battlefield"` case to `moveToZone`** that creates a `Permanent` (with proper Owner/Controller/Timestamp/Counters/Flags initialization), appends to the seat's `Battlefield`, registers replacements, and fires ETB triggers — i.e. fold `enterBattlefieldWithETB`'s logic into the universal entry point.
2. **Audit the 34 per-card hooks** to remove the redundant `enterBattlefieldWithETB` *or* the redundant `MoveCard`. Pick one canonical "card lands on battlefield" call and make every site use it.
3. **Add a regression test** in the gameengine suite that runs each per-card reanimate hook against a fresh fixture and asserts `seat.Graveyard` does NOT contain the reanimated card pointer afterward (and that the Battlefield contains exactly one Permanent referencing it).

A blunter interim mitigation: have `enterBattlefieldWithETB` defensively `removeFromZone(seat, card, "graveyard"|"hand"|"exile"|"library")` for all four non-battlefield zones before creating the Permanent. That swallows the duplicate without forcing the per-card audit.

## Hypothesis #2 — undercounts (~20 %)

Negative diffs are smaller and rarer (`−4` to `−9`). Likely sources, in priority order:

1. **§704.5e copy SBA over-sweep on the battlefield boundary.** `sba704_5e` (sba.go:417) sweeps `Hand/Graveyard/Exile/Library` for `IsCopy=true` cards. Non-token spell copies that resolved as permanents can survive on the battlefield with `IsCopy=true`. When such a permanent leaves the battlefield, it lands in graveyard and is immediately swept; `total` drops by 1 but `expected` doesn't move. Effect is small per game (`−1` per cleaned-up permanent copy).

2. **§800.4a "cards left game" miscount.** When a player concedes/loses, `multiplayer.go:391` sets `seat.Flags["cards_left_game"] = realCardsLeaving` and `expected -= cards_left_game`. If a per-card or rule path moves a card off the battlefield AFTER the seat is marked Lost without going through that bookkeeping, `total` drops below adjusted `expected`. Likely contributor on multi-eliminated games.

3. **Stack-resolution drop.** A spell resolves on the stack; the card pointer should land in graveyard for non-permanent spells. If a per-card snowflake sets `Card.Effect = nil` or otherwise short-circuits cleanup, the pointer can be lost.

These need further targeted instrumentation to confirm — they're outside the scope of this initial analysis.

## What's *not* the problem

- **`resolveTutor`** is correct: it removes from library before placing.
- **Tokens**: the Feynman check correctly subtracts `IsToken()` permanents on the battlefield.
- **Commander zone**: counted in `total` and the expected adjusts for partner pairs (`expected = 99 + len(CommanderNames)`).
- **MDFC back-face plays**: the runtime card identity is now the back face (commit `fa018fd`), but the same single `*Card` pointer is in exactly one zone — no duplication.

## Suggested next step

Open a tracking issue and attack hypothesis #1 first — fixing the `moveToZone("battlefield")` fallthrough should eliminate ~80 % of these violations in one stroke. The 34-file per-card audit is ugly but bounded; a single defensive `removeFromZone` sweep inside `enterBattlefieldWithETB` is a 5-line interim that should drop the violation rate dramatically while the deeper refactor lands.
