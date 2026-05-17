# Muninn parser-gap bulk patterns — round 5 (2026-05-17)

## Question

After rounds 1-4 shipped nine bulk-pattern families (land_tax,
evoke_color_gate, lifegain_endstep, etb_tribe_gate, lifegain_counter,
gated_etb_effect, end_step_intervening_if, shuffle_self_from_grave,
etb_library_tutor) and 100+ per-card handlers closed the entire top-180
gap list, what card-family shapes still warrant a generic dispatcher?

Important context: the 167→171-card EOD parser_gaps.json snapshot is
STALE — DARKSTAR is still running the pre-merge binary, so every
post-2026-05-10 handler is invisible to the grinder. The remaining
"uncovered" tail is therefore a forward-looking question — what shapes
will appear once the post-deploy grinder runs against the current
corpus?

## Answer

Two new families ship in this branch (`dev/muninn-bulk-patterns-5`),
covering 10 cards total. Both are ETB-trigger shapes — common enough in
the broader corpus (75+ ETB-basic-land-ramp cards alone per Scryfall
oracle dump) that they belong in a family, not 10 hand-rolled files.

### Family 1 — ETB basic-land ramp (`etb_basic_land_ramp_family.go`)

Generic dispatcher for the un-gated single-basic ETB ramp shape:

```
When this creature enters, [you may] search your library for a basic
land card, put it onto the battlefield tapped / into your hand, then
shuffle.
```

Distinct from `land_tax_family.go`, which gates on opponent-controls-
more-lands. This family fires unconditionally on ETB.

| Card             | Destination          |
| ---------------- | -------------------- |
| Farhaven Elf     | battlefield (tapped) |
| Civic Wayfinder  | hand                 |
| Borderland Ranger| hand                 |
| Sylvan Ranger    | hand                 |
| Pilgrim's Eye    | hand                 |

Algorithm: walk the library in order, take the first basic land, route
via `MoveCard("library" → "battlefield_tapped" | "hand")` so §614
replacements + landfall observers fire, shuffle once regardless of
whether anything was found (CR §701.18b — searching a hidden zone
shuffles even on whiff).

Hat policy on the "you may" rider: always opt in. Hand-filtered ramp is
monotone upside in every archetype Yggdrasil scores.

Intentional skips with rationale in the family docstring:

- **Wood Elves** — Forest subtype filter, enters untapped; first member
  of a sibling family if more Forest-specific cards land.
- **Solemn Simulacrum** — same ETB shape, but bundles a die-draw
  trigger.
- **Yavimaya Granger** — adds Echo upkeep cost (engine doesn't track
  Echo distinctly today).
- **Gatecreeper Vine / District Guide / Scampering Surveyor** — filter
  alternation "basic land OR Gate/Cave".
- **Primeval Herald** — fires on ETB AND attack.
- **Loam Larva** — puts the land on TOP of the library.

### Family 2 — ETB drain target opponent (`etb_drain_target_opponent_family.go`)

Generic dispatcher for the vampire-cycle drain shape:

```
When this creature enters, target opponent loses N life and you gain
N life.
```

| Card                 | Amount |
| -------------------- | ------ |
| Skymarch Bloodletter | 1      |
| Vampire Sovereign    | 3      |
| Highway Robber       | 2      |
| Dakmor Ghoul         | 2      |
| Bloodborn Scoundrels | 2      |

Algorithm:

1. Pick the lowest-life living opponent (matches Athreos / Belakor /
   Ajani Nacatl Pariah heuristic — maximizes lethal-pressure stacking
   into win-line detection).
2. `gameengine.LoseLife` on the target (fires life_lost triggers →
   Bloodchief Ascension / Vito).
3. `gameengine.GainLife` on the controller (fires life_gained triggers
   → Soul Sister / Karlov / lifegain_counter_family).
4. `gs.CheckEnd()` so SBAs see any lethal life-total change.

Hand-rolled siblings with gates (Kalastria Healer ally, Tithe Drinker
extort, Blood Artist creature-died, Falkenrath Noble damage-trigger,
Vito conversion) keep their bespoke handlers — this family only owns
the un-gated shape.

## Investigation notes

Cards considered for families that turned out to be too niche or already
covered:

- **Burnished Hart / Sakura-Tribe Elder / Dawntreader Elk / Diligent
  Farmhand / Wayfarer's Bauble** — share a "sacrifice self, fetch basic
  land tapped" shape. These are activated abilities, not triggered, so
  they'd need an OnActivated family. Burnished Hart already has a
  bespoke handler (committed 2026-05-17 in wave 181-200). Sakura-Tribe
  Elder et al. fire via the engine's generic ramp-cost activation
  pipeline rather than per_card hooks. Deferring until a confirmed gap
  shows the bespoke pipeline missing.
- **Starlit Soothsayer / Star Charter / Starseer Mentor** — share a
  "end step, if you gained or lost life this turn, <effect>" gate. Only
  Starseer Mentor is uncovered; the other two have bespoke handlers
  already. Single uncovered member doesn't justify a new family;
  candidate for a future extension to `end_step_intervening_if_family`
  with a `gainedOrLostLife` gate.
- **Wood Elves + Forest-fetchers** — Forest subtype filter, untapped
  destination. Only one Muninn-relevant card in this shape today.
- **Soul Sister cluster (Soul Warden / Soul's Attendant / Auriok
  Champion / Essence Warden / Ajani's Welcome / Suture Priest)** —
  share "Whenever (another) creature enters, [you may] gain 1 life"
  shape. None in the current parser_gaps.json, none have handlers,
  but all are routed via permanent_etb listener with two-way
  controller-scope axes (your-only vs any) and an optional opponent
  drain rider (Suture Priest). Defer until any of them surfaces in a
  redeployed-DARKSTAR grinder cycle.

## Measurement

```
$ go test ./internal/gameengine/per_card/... -count=1
ok  github.com/hexdek/hexdek/internal/gameengine/per_card  0.89s

$ go test ./internal/gameengine/... -count=1
ok  github.com/hexdek/hexdek/internal/gameengine          3.91s
ok  github.com/hexdek/hexdek/internal/gameengine/per_card 0.89s
```

`bulk_patterns_5_test.go` — 9 new smoke tests, all passing:

- Farhaven Elf puts basic tapped onto battlefield (with tap state +
  library removal).
- Pilgrim's Eye puts basic into hand.
- Civic Wayfinder picks first basic (deterministic, library-order).
- Sylvan Ranger whiff emits `per_card_failed` event.
- Borderland Ranger sibling shape.
- Skymarch Bloodletter drains 1 from lowest-life opponent.
- Vampire Sovereign drains 3.
- Highway Robber / Dakmor Ghoul / Bloodborn Scoundrels share the
  2-life-drain shape.
- Drain skips dead opponents (Lost flag respected).

Direct gap-list impact once DARKSTAR redeploys + a grinder cycle runs:
the five drain creatures are common Vampire-tribe and ETB-drain
constructed staples (Skymarch Bloodletter is in 30+ EDH precons), and
the five basic-land ramp creatures are universal green ramp staples
(Farhaven Elf, Civic Wayfinder, Borderland Ranger, Sylvan Ranger,
Pilgrim's Eye all appear across hundreds of EDH decks). Expect ≥10
entries to vanish from `parser_gaps.json` as decks featuring them roll
through the post-deploy grinder.

## Extending

To add a new family member:

- **ETB basic-land ramp:** append an `etbBasicLandRampEntry` row with
  `cardName` and `dest`. Add a new `landRampDestination` enum + arm in
  `runEtbBasicLandRamp`'s destination switch if your card needs a new
  destination (e.g. top of library for Loam Larva, untapped for Wood
  Elves).
- **ETB drain:** append an `etbDrainEntry` row with `cardName` and
  `amount`. The lowest-life-opponent picker is shared; no per-card
  customization needed for the canonical shape.

The dispatcher handles wiring, slug emission, `gs.CheckEnd()`, `emit` /
`emitFail`, and life-zero / dead-opponent screening automatically.
