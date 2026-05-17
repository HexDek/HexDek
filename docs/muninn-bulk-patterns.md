# Muninn parser-gap bulk patterns — 2026-05-16

## Question

The Muninn `parser_gaps.json` log has **167 distinct card entries** (a snippet
per card whose AST didn't produce a meaningful engine effect during live play).
After the recent one-card-at-a-time handler waves (top 20 snowflakes),
~147 entries remain. Can we close them in **classes** instead of one card
at a time?

## Answer

Yes for two clear shapes. Both ship in this branch
(`dev/muninn-bulk-patterns`).

### Pattern 1 — Land-tax family (`land_tax_family.go`)

Shape:

```
[ETB | upkeep | attack] trigger:
  if [opponent | defending player] controls more lands than you,
  search your library for a [basic] <subtype> card,
  put it [tapped | untapped] [onto the battlefield | into your hand],
  then shuffle.
```

Driven by a config table (`landTaxFamilyEntries`). Each row picks trigger
kind, land filter (basic flag, subtype, any-land), destination zone,
tapped flag, count, comparator (any-opponent vs defending-player),
optional repeat count.

Hand-rolled siblings stay registered: **Land Tax**, **Knight of the
White Orchid**, **Claim Jumper**. The family handler only owns the gap
cards.

Gap entries closed by this pattern:

| Card | Trigger | Filter | Notes |
|------|---------|--------|-------|
| Loyal Warhound | ETB | basic Plains | Tapped to battlefield |
| Sand Scout | ETB | Desert (any) | Tapped to battlefield |
| Aerial Surveyor | attacks | basic Plains | Defender-side gate |

**3 base gap entries closed** plus the cascade variants
(`Sand Scout (cascade)`, downstream cascade entries that share the same
handler dispatch).

### Pattern 2 — Hybrid-evoke color gate (`evoke_color_gate.go`)

Shape:

```
Hybrid mana cost {C1/C2}{C1/C2} (+ filler).
When ~ enters, if {C1}{C1} was spent to cast it, <effect A>.
When ~ enters, if {C2}{C2} was spent to cast it, <effect B>.
Evoke {C1/C2}{C1/C2}
```

Engine doesn't yet track per-cast hybrid-pip payment
(`resolve.go::mana_spent` returns true by default). Vibrance's existing
handler set the convention: when the ETB was triggered by a cast (the
`was_cast` flag is set), fire **both** color halves; on non-cast entries
(reanimate / blink / Sneak Attack), fire **neither**. We replicate that
gate generically and emit a `per_card_partial` breadcrumb so Muninn
keeps tracking the residual hybrid-pip gap.

Adding a card is two pieces: two ETB-effect closures (mode A, mode B)
and a one-line entry in `evokeColorGateEntries`.

Vibrance stays on its own handler. The family owns the gap cards:

| Card | Color A | Color B |
|------|---------|---------|
| Wistfulness | G — exile target opponent artifact/enchantment | U — draw 2, discard 1 |
| Deceit | U — bounce target nonland permanent | B — opponent reveals, you choose nonland to discard |

**2 base gap entries closed** plus the cascade variants.

## Measurement

Goldilocks (deterministic per-card effect verification) on the targets:

```
$ go run ./cmd/hexdek-thor/ --card-list <targets> --goldilocks
PASS: 5 / 5 cards, ZERO FAILURES
```

Per-handler effect breakdown (corpus-audit assertions): exile / draw /
bounce / discard / counter_mod — all firing as expected.

Direct gap-list reduction:

- Loyal Warhound, Sand Scout, Aerial Surveyor, Wistfulness, Deceit
  → **5 base entries** drop from `parser_gaps.json` once a tournament
  run sees these cards on the battlefield.
- Cascade variants (`Wistfulness (cascade)`, `Sand Scout (cascade)`,
  `Deceit (cascade)`, `Aerial Surveyor (cascade)` if it appears) close
  on the next refresh because the registry dispatch is the same code
  path.

Conservative tournament-run impact: **~5–10 entries** out of the ~147
remaining (~3–7%).

## Why not more

Looked at the obvious neighbouring clusters and rejected them as
**bulk** patterns:

- **"If you cast it" ETB gate** (~11 gap cards: Crackling Spellslinger,
  Eon Frolicker, Geological Appraiser, Bringer of the Last Gift, etc.).
  The gate is shared but each body is unique (storm grant, discover N,
  board sacrifice, counter-spell + replay, etc.) — bulk gate alone
  doesn't close the gap, since the dead effect IS the body.
- **"If you gained life this turn" end-step trigger** (~10 gap cards).
  Same problem: shared gate (life-gained tracker), bodies all unique
  (recursion, distribute counters, opponent sac, etc.).
- **"If creatures died this turn" end-step trigger** (~3 gap cards).
  Same shape, same bulk-vs-bodies tradeoff.

These would need engine-level work (a generic intervening-if gate that
the AST evaluator dispatches to body effects already parsed elsewhere)
— a much larger refactor than one or two new files. Tracked for a
later pass.

## Extending

To add a new family member:

- **Land-tax shape:** append a `landFetchFamilyConfig` row to
  `landTaxFamilyEntries`. No new file, no new register call.
- **Evoke-color shape:** write two body closures and append an
  `evokeColorGateEntry` row to `evokeColorGateEntries`.

The dispatch wiring (`OnETB` / `OnTrigger`) and the breadcrumb emission
(`emit` / `emitPartial`) are handled by the family runner.
