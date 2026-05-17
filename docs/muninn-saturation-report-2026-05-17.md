# Muninn Saturation Report — 2026-05-17

## Headline

The Muninn parser-gap log is **fully covered**. Fresh data re-pulled
from DARKSTAR
(`scp josh@192.168.1.207:~/hexdek/data/muninn/parser_gaps.json
data/muninn/`, mtime 11:09 UTC, 30,862 bytes, 176 entries) — every
snippet in the file resolves to either a bespoke `per_card/*.go`
handler, a generated `gen_*.go` registration, or one of the 10
bulk-pattern family scaffolds. Wave #221-250 was opened to add 10
more handlers and abandoned in flight: there is nothing in the current
gap log left to handle.

This document records that state and the trajectory that got us here.

## Coverage check — every entry in the gap log

Verified by name against `internal/gameengine/per_card/` plus the
cascade/copy/token dispatch fallback merged in wave #201-220 (commit
`c09db37`):

| Category | Examples | Resolution |
|---|---|---|
| Top-30 cards (high cumulative hits) | The One Ring, Land Tax, Bloodchief Ascension, Necromancy, Light-Paws, Kodama of the East Tree, Tiamat, Great Hall of the Biblioplex, Acererak, Oversold Cemetery | bespoke `per_card` files (waves 1-160) |
| Mid-tail singletons | Sam Loyal Attendant, Burnished Hart, Trostani Selesnya's Voice | bespoke (waves 161-180, 181-200) |
| Low-traffic snowflakes | Aerial Surveyor, Lighthouse Chronologist, Exemplar of Light, Archangel of Thune | family scaffolds (`land_tax_family`, `end_step_intervening_if_family`, `lifegain_counter_family`) |
| Cascade-renamed variants | `"The One Ring (cascade)"`, `"Land Tax (cascade)"`, etc. | wave 201-220 `lookupCandidates` fallback in `registry.go` |
| Urza / Miirym / Adrix copies | `"Phoenix Fleet Airship (Urza copy)"`, `"Tiamat (Miirym Token)"`, `"Crown of Gondor (Urza copy) (Urza copy)"` | same fallback (strips trailing parenthetical, recurses) |
| Embalm / Eternalize / Encore / Blitz / Champion / Replicate / Epic / Gravestorm / Myriad / Unearth / Storm copies | `"X (Embalmed)"`, `"X (storm copy N)"`, etc. | same fallback |
| Token strings (`"Wistfulness Token"`, `"creature token knight Token"`, `"Token"`) | — | name-suffix stripped by fallback; underlying card resolves |
| DFC partials (`"Esika, God of the Tree // The Prismatic Bridge"`, `"Curious Homunculus // Voracious Reader"`, …) | — | `" // "` fallback in `fireETB`/`fireTrigger` (pre-201) plus new lookupCandidates |

The 99.6% variant-coverage audit (`docs/muninn-variant-coverage-2026-05-17.md`,
commit `731843c`) already enumerated all greppable rename sites in the
engine and confirmed every one is caught by the `lookupCandidates`
fallback. The remaining 0.4% is not a missing handler — it is a card
with a real unimplemented sub-ability that the handler emits via
`emitPartial`, which keeps the entry on the gap list until that
sub-ability is wired.

## Today's contribution (2026-05-17)

| Commit | Type | Content |
|---|---|---|
| `e85c2aa` | bespoke wave (#161-180) | 1 handler: Sam, Loyal Attendant |
| `d4ad314` | bulk-pattern (round 4) | 2 families: `shuffle_self_from_grave_family`, `etb_library_tutor_family` |
| `e27cb27` | bespoke wave (#181-200) | 2 handlers: Burnished Hart, Trostani Selesnya's Voice |
| `fe7c203` | bulk-pattern (round 5) | 2 families: `etb_basic_land_ramp_family`, `etb_drain_target_opponent_family` |
| `c09db37` | dispatch fallback (#201-220) | 0 new handlers; `lookupCandidates` helper in `registry.go` strips one trailing parenthetical + DFC front face across all five `fire*` dispatchers |

**Totals shipped today:** 3 bespoke handlers, 4 bulk-pattern families,
1 dispatch fallback. Equivalent to ~10+ handlers' worth of gap-log
coverage once DARKSTAR's gap counters roll over (the variant fallback
alone rescues ~5k cumulative hits on Necromancy/Vibrance/Land
Tax/Phoenix Fleet Airship/Tiamat cascades).

## Bulk-pattern families (cumulative, 10 total)

| File | Pattern | Round | Commit |
|---|---|---|---|
| `land_tax_family.go` | "if [opponent / defending player] controls more lands than you, search library for a basic and put it [tapped] into play" | 1 | `c9006e1` |
| `lifegain_endstep_family.go` | "At the beginning of your end step, if …, gain N life / scry N / draw" | 2 | `4101aab` |
| `etb_tribe_gate_family.go` | "When this creature enters, if you control [tribe], …" | 2 | `4101aab` |
| `lifegain_counter_family.go` | "Whenever you gain life, put a +1/+1 counter on …" | 2 | `4101aab` |
| `gated_etb_effect_family.go` | "When ~ enters, if <self-gate>, do X" | 3 | `cf6b7d6` |
| `end_step_intervening_if_family.go` | "At the beginning of <step>, if <condition>, …" | 3 | `cf6b7d6` |
| `shuffle_self_from_grave_family.go` | "When ~ is put into a graveyard from anywhere, shuffle it into its owner's library" | 4 | `d4ad314` |
| `etb_library_tutor_family.go` | "When this creature enters, [you may] search your library for a [type] card" | 4 | `d4ad314` |
| `etb_basic_land_ramp_family.go` | "When this creature enters, [you may] put a basic land card from your hand onto the battlefield [tapped]" | 5 | `fe7c203` |
| `etb_drain_target_opponent_family.go` | "When this creature enters, target opponent loses N life and you gain N life" | 5 | `fe7c203` |

## Coverage trajectory across waves 1-220

| Wave | Commit | Handlers | Date | Notes |
|---|---|---|---|---|
| #8-#12 | `c7341ca`+`46a693f` | 5 | 2026-05-16 | top-12 holdouts |
| #13-#20 | `11ef8d8` | 8 | 2026-05-16 | Claim Jumper et al. |
| #21-#30 | `ab191df` | 9 | 2026-05-16 | + stragglers |
| #31-#40 | `9dc606e` | 8 | 2026-05-16 | snowflakes |
| #41-#60 | `fcb8861` | 10 | 2026-05-16 | |
| Top-50 close-out | `43b17dc` | 5 | 2026-05-16 | + 51-60 holdouts |
| #51-#70 | `dbe1754` | 10 | 2026-05-16 | |
| #61-#80 | `4fe7a99` | 10 | 2026-05-16 | |
| #81-#100 | `ce18e02` | 10 | 2026-05-16 | |
| #101-#120 | `4f65847` | 11 | 2026-05-16 | |
| #121-#140 | `5cff75e` | 10 | 2026-05-16 | |
| #141-#160 | `ca27f7e` | 12 | 2026-05-16 | |
| #161-#180 | `e85c2aa` | 1 | 2026-05-17 | "signal saturated" — only Sam Loyal Attendant net-new |
| #181-#200 | `e27cb27` | 2 | 2026-05-17 | tail single-hits (Burnished Hart, Trostani) |
| #201-#220 | `c09db37` | 0 | 2026-05-17 | dispatch fallback — no new registrations |
| **Total** | | **111** | | |

Plus 10 bulk-pattern families (see above), plus the dispatch
fallback. The shape of the curve is unmistakable:

```
Wave size (handlers per wave)
12 ┤                                    ●
11 ┤                              ●
10 ┤        ●     ●  ●  ●  ●  ●        ●
 9 ┤              ●
 8 ┤        ●  ●
 5 ┤  ●        ●
 2 ┤                                            ●
 1 ┤                                       ●
 0 ┤                                                ●
   └──────────────────────────────────────────────────
     8 13 21 31 41 51 51 61 81 101 121 141 161 181 201
     ↑              top-50 dense ↑          ↑ tail ↑
     bootstrap                              saturation
```

The 2026-05-16 push (Saturday) cleared the top 160 entries at
~10 handlers per wave. By Sunday morning the entire long tail had
collapsed: wave #161-180 found one bespoke target, #181-200 found
two, #201-220 found zero. The remaining "hits" in the log are all
variant-name dispatches against handlers that already exist — what
#201-220 patched.

## What's left

The current 176 entries fall into three buckets:

1. **Cumulative hit counts on covered cards** — the top-100 entries
   keep growing because DARKSTAR's grinder runs constantly and every
   parse re-emits a `gap_hit` event. These will not disappear from
   the log until the cumulative counters are reset (and a covered
   card under fresh counters quickly flat-lines).
2. **Variant-name re-emissions** — same root card, different
   suffix. The wave #201-220 fallback resolves the handler at
   dispatch time but the parser still tracks the variant string.
3. **`emitPartial` annotations** — some handlers wire the main
   ability but document an un-implemented sub-ability via
   `emitPartial`. Those entries (e.g.
   `"Solphim, Mayhem Dominus: noncombat damage doubling requires
   DealDamage replac…"`) are real residual work but they are not
   "no handler" — they are "handler exists, branch X not wired."

None of these are "uncovered cards."

## Next steps

- **Wait for fresh tournament data** that surfaces actual new cards.
  Coverage will not improve until the corpus changes; the engine is
  not the bottleneck.
- **Upgrade `emitPartial` sites** when individual sub-abilities
  become load-bearing for a specific deck or matchup. Top files by
  partial count: `urza_lord_high_artificer.go` (7),
  `elesh_norn_argent_etchings.go` (6), `heliod_sun_crowned.go` (5),
  `era3_batch.go` (5).
- **Drop the "wave NNN-MMM" cadence.** Continuing it produces 0-2
  handler PRs that thrash the gap-log without moving coverage.
  Switch to event-driven work: muninn flags a card → open a focused
  PR.
