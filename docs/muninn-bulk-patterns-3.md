# Muninn parser-gap bulk patterns — round 3 (2026-05-16)

## Question

After rounds 1 and 2 shipped five bulk-pattern families (land_tax,
evoke_color_gate, lifegain_endstep, etb_tribe_gate, lifegain_counter)
and waves #21-#100 closed ~80 per-card snowflakes, what other
card-family shapes are still investigatable in the remaining Muninn
parser gaps?

## Answer

Two more shapes ship in this branch (`dev/muninn-bulk-patterns-3`). A
third candidate was investigated and rejected.

### Family 1 — Gated ETB effect (`gated_etb_effect_family.go`)

Generic dispatcher for the very common shape

```
When ~ enters, if <self-gate>, <effect>.
```

The `<self-gate>` is one of four canonical checks on the entering
permanent itself, all already maintained by the engine:

| Gate enum            | Reads                                  | Source         |
| -------------------- | -------------------------------------- | -------------- |
| `etbGateNone`        | always true                            | —              |
| `etbGateWasCast`     | `perm.Flags["was_cast"] != 0`          | `stack.go`     |
| `etbGateCastFromHand`| `perm.Flags["cast_from_hand"] != 0`    | `stack.go`     |
| `etbGateNotToken`    | `"token"` absent from `perm.Card.Types`| AST + tokens   |
| `etbGateSneakEntry`  | `perm.Flags["sneak_entry"] != 0`       | `ninja_sneak.go` |

The `<effect>` is unique per card, so each entry plugs in its own body
closure. Identical to the closure-per-entry shape already used by
`lifegain_endstep_family.go`.

Entries shipped this round:

| Card                       | Gate         | Effect                                  |
| -------------------------- | ------------ | --------------------------------------- |
| Weftwalking                | was_cast     | Shuffle hand + graveyard into library, draw 7 |
| Leonardo, Leader in Blue   | sneak_entry  | Creatures you control get +2/+0 EOT     |

Hand-rolled siblings already shipped on bespoke handlers (Skitterbeam
Battalion, Crackling Spellslinger, Eon Frolicker, Geological Appraiser,
Bringer of the Last Gift, Breaching Leviathan, Tiamat, Gruff Triplets)
all match one of the gates above and could be migrated to the family in
a future cleanup pass — out of scope here because (a) the bespoke
handlers all carry extra mechanic-specific bookkeeping (prototype, evolve,
discover, etc.) that doesn't fit a generic body closure, and (b)
migration risks regressions on already-working effects with no coverage
gain.

### Family 2 — End-step intervening-if effect (`end_step_intervening_if_family.go`)

Generic dispatcher for

```
At the beginning of [each|your] end step, if <condition>, <effect>.
```

The scope axis (each-end-step vs your-end-step) and the gate axis are
independent — both pluggable.

| Gate enum                            | Reads                                |
| ------------------------------------ | ------------------------------------ |
| `endStepGateCastNoncreatureThisTurn` | `seat.Turn.Casts` filtered           |
| `endStepGateNotMyTurn`               | `active_seat != perm.Controller`     |

Entries shipped this round:

| Card                       | Scope | Gate          | Effect                |
| -------------------------- | ----- | ------------- | --------------------- |
| Lighthouse Chronologist    | each  | not_my_turn   | Take an extra turn    |

Hand-rolled siblings (Hurkyl Master Wizard, Phoenix Fleet Airship, Lord
Jyscal Guado, Feast of the Victorious Dead) all match this shape, but
each carries its own per-turn flag-tracking machinery for the gate
(distinct-types-of-noncreature-spells-cast, sacrificed-permanent count,
counter-on-creature flag, creatures-died count) that isn't a simple
read from `Turn.Casts`. The shared scaffold is genuine — the gate
helpers are not. Migration deferred.

### Rejected — "If you cast it" cast-time storm/seqthing family

Investigated as a third family. The gate check is simple (`was_cast`),
but every body in the gap list (Skitterbeam → self-copy, Weftwalking →
shuffle/draw, Eon Frolicker → extra turn + protection, Geological
Appraiser → discover 3, Bringer of the Last Gift → mass sac + reanimate,
Crackling Spellslinger → grant storm, Breaching Leviathan → mass tap +
skip-untap) is genuinely unique with no shared sub-shape. The dispatch
saving is sub-trivial (one `if` per handler).

Pulled in via the `gated_etb_effect_family.go` scaffold instead for the
two new cards (Weftwalking, Leonardo), letting future entries plug in
without a separate file. Same answer as round 2's note on this shape —
verified still correct against the current corpus.

## Investigation notes

Cards considered for a family that turned out to be too niche to bulk:

- **Sproutback Trudge** — end-step intervening-if (lifegain) trigger
  fires while the card is in the graveyard. Engine trigger dispatch
  iterates `Seat.Battlefield` only — there is no hook today for "phase
  trigger on a card in a private zone" (same gap noted on Ichorid's
  upkeep-from-grave return and Quintorius's cards-leaving-graveyard
  listener). Needs engine-level work; not a per-card handler.
- **Lighthouse Chronologist's extra-turn body** — engine has a single
  global `gs.Flags["extra_turns_pending"]` counter (`resolve.go:2479`),
  not a seat-keyed queue. Routing the extra turn to "the controller of
  Lighthouse Chronologist" rather than "the active player" collapses
  to the right answer in this card's case because the trigger fires on
  an opponent's end step → the next active turn IS the controller —
  but is recorded for future seat-target extra-turn cards (Eon Frolicker
  has the same partial today).
- **Light-Paws, Emperor's Voice** — "Whenever an Aura you control
  enters, if you cast it, search your library for an Aura with mana
  value less than or equal…" — listener-on-other plus a tutoring body
  with cost-bound search. Bespoke handler material; nothing else in the
  gap list shares this shape.
- **Senu, Keen-Eyed Protector** — exile-and-return-on-other-attacks,
  unique exiled-card-trigger plumbing.
- **Aradesh, the Founder** — gates on the enlist mechanic which the
  engine doesn't track at all; needs engine work, not a card handler.

Concurrent waves landed per-card handlers for the other candidates that
would have fit either family (Gruff Triplets, Hurkyl Master Wizard,
Skitterbeam Battalion, Quicksilver Fountain, Oracle of Bones). The
family scaffolds note those siblings in their entry-list docstring
comments so a future migration pass can find them.

## Measurement

Thor on the three round-3 target cards (`hexdek-thor --card-list
/tmp/muninn_round3_cards.txt --goldilocks --corpus-audit`):

```
=== THOR COMPLETE ===
  cards tested:  3
  total tests:   72
  failures:      0
  ZERO FAILURES — fully sterile.
```

Per-effect corpus-audit assertions: `buff` 100%, `grant_ability` 100%.
The goldilocks driver also passed all 3 cards (Weftwalking shuffle/draw
fires, Leonardo buffs the team, Lighthouse Chronologist queues the
extra turn).

Full `go test ./internal/gameengine/... -count=1` is clean (4.88s
`gameengine`, 1.46s `per_card`).

Direct gap-list reduction:

- Weftwalking, Leonardo Leader in Blue, Lighthouse Chronologist
  → **3 base entries** drop from `parser_gaps.json` once DARKSTAR is
  redeployed and a tournament cycle sees these cards in play.

Conservative tournament-run impact: **~3-5 entries** including cascade
variants (Sand Scout-style cascade duplicates).

## Extending

To add a new family member:

- **Gated-ETB shape:** append a `gatedEtbEffectEntry` row with
  `cardName`, `gate`, a body closure `func(gs, perm)`, and an optional
  `partial` reason. Add a new gate enum + matching arm in
  `etbGateOpen`/`etbGateName`/`etbGateFailReason` if your card needs
  a new self-condition.
- **End-step shape:** append an `endStepInterveningEntry` row with
  `cardName`, `scope`, `gate`, and a body closure `func(gs, perm, ctx)`.
  Add a new gate enum + matching arm in `endStepGateOpen` /
  `endStepGateName` / `endStepGateFailReason` if your card needs a new
  intervening-if check.

The dispatcher handles wiring, slug emission, `gs.CheckEnd()`,
`emit`/`emitPartial`, and active-seat scoping automatically.
