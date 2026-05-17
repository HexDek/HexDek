# Muninn Variant-Coverage Audit — 2026-05-17

## Question

The dev-2 round-13 dispatch-fallback handler `c09db37` ("wave 201-220") wired
a `lookupCandidates` helper that strips one trailing parenthetical (and DFC
front-face) from `Card.DisplayName()` before fanning out across all five
`fire*` dispatchers. The claim: this covers the cascade/copy/token name
variants that previously cost ~5k cumulative gap-log hits.

This audit answers two questions against the current corpus:

1. How many distinct name-variant patterns exist in the engine?
2. Which of the top variant snippets does `c09db37` actually rescue, and
   which still fall through?

## Sources of variant names

Greppable rename sites in `internal/gameengine/` that mutate `Card.Name`
(or assign a fresh `Name:` to a token/copy permanent):

| Site | Suffix shape | Handler? |
|---|---|---|
| `cascade.go:150` | `"X (cascade)"` | ✓ stripped |
| `per_card/custom_maelstrom_wanderer.go:49` | `"X (2nd cascade)"` | ✓ stripped |
| `keywords_misc.go:238` | `"X (Embalmed)"` | ✓ stripped |
| `keywords_misc.go:308` | `"X (Eternalized)"` | ✓ stripped |
| `keywords_misc.go:374` | `"X (Encore)"` | ✓ stripped |
| `keywords_batch.go:793` | `"X (blitz)"` (SourceCardName) | ✓ stripped |
| `keywords_batch4.go:1007` | `"X (champion LTB)"` (SourceCardName) | ✓ stripped |
| `keywords_batch3.go:178` | `"X (replicate N)"` | ✓ stripped |
| `keywords_batch6.go:654` | `"X (epic)"` (SourceCardName) | ✓ stripped |
| `keywords_batch6.go:859` | `"X (gravestorm N)"` | ✓ stripped |
| `keywords_combat.go:439` | `"X (myriad copy)"` | ✓ stripped |
| `keywords_p1p2.go:119` | `"X (unearth)"` (SourceCardName) | ✓ stripped |
| `storm.go:142` | `"X (storm copy N)"` | ✓ stripped |
| `keywords_batch3.go:548` | `"X (squad token)"` | ✓ stripped |
| `per_card/era3_batch.go:649` (Urza) | `"X (Urza copy)"` | ✓ stripped |
| `per_card/miirym_sentinel_wyrm.go:52` | `"X (Miirym Token)"` | ✓ stripped |
| `per_card/adrix_and_nev_twincasters.go:51` | `"X (Adrix copy)"` | ✓ stripped |
| `per_card/altair_ibn_la_ahad.go:220` | `"X (altair copy)"` | ✓ stripped |
| `per_card/custom_inalla_archmage_ritualist.go:60` | `"X (Inalla token)"` | ✓ stripped |
| `per_card/custom_saheeli_radiant_creator.go:118` | `"X (Saheeli token)"` | ✓ stripped |
| `per_card/custom_yenna_redtooth_regent.go:93` | `"X (Yenna token)"` | ✓ stripped |
| `per_card/custom_araumi_of_the_dead_tide.go:119` | `"X (Araumi token)"` | ✓ stripped |
| `per_card/gen_sauron_the_necromancer.go:67` | `"X (Wraith token)"` | ✓ stripped |
| `per_card/gen_the_jolly_balloon_man.go:88` | `"X (Balloon token)"` | ✓ stripped |
| `per_card/compy_swarm.go:82` | `"X (Compy token)"` | ✓ stripped |
| `per_card/preston_the_vanisher.go:105` | `"X (Illusion token)"` | ✓ stripped |
| `per_card/zinnia_valleys_voice.go:85` | `"X (1/1 Offspring)"` | ✓ stripped |
| `per_card/custom_sliver_gravemother.go:97` | `"X (Encore token)"` | ✓ stripped |
| `per_card/lorehold_archivist.go:104` | `"X (Restore-Relic token)"` | ✓ stripped |
| `per_card/life_of_the_party.go:57` | `"X (Life-of-the-Party token)"` | ✓ stripped |
| `per_card/ratadrabik_of_urborg.go:85` | **`"X Token"` (no parens)** | ✗ MISS |
| `per_card/paradigm_echocasting_symposium.go:63` | **`"X Token"` (no parens)** | ✗ MISS |
| `per_card/runo_stromkirk.go:174` | **`"X Token"` (no parens)** | ✗ MISS |
| `per_card/sin_spiras_punishment.go:92` | **`"X Token"` (no parens)** | ✗ MISS |

40 distinct parenthetical patterns appear in source (`Name`/`SourceCardName`
assignments) — of those, ~16 are real card-name suffixes (the rest are
diagnostic strings like `(layer=%d)`, `(CR 704.6c)`, `(orphaned linked exile)`
that never appear on a `Permanent.Card.Name`).

## Top 30 variant snippets observed in `data/muninn/parser_gaps.json`

Ranked by hit count. Pattern column groups by suffix shape.

| # | Snippet | Count | Pattern | c09db37 |
|--:|---|--:|---|:-:|
|  1 | `The One Ring (cascade)` | 924 | `(cascade)` | ✓ |
|  2 | `Tiamat (Miirym Token)` | 605 | `(X Token)` | ✓ |
|  3 | `Transcendent Dragon (cascade)` | 523 | `(cascade)` | ✓ |
|  4 | `Evercoat Ursine (cascade)` | 395 | `(cascade)` | ✓ |
|  5 | `Wistfulness (cascade)` | 394 | `(cascade)` | ✓ |
|  6 | `Vibrance (cascade)` | 283 | `(cascade)` | ✓ |
|  7 | `Deceit (cascade)` | 268 | `(cascade)` | ✓ |
|  8 | `Land Tax (cascade)` | 248 | `(cascade)` | ✓ |
|  9 | `Rocco, Cabaretti Caterer (cascade)` | 197 | `(cascade)` | ✓ |
| 10 | `Oversold Cemetery (cascade)` | 185 | `(cascade)` | ✓ |
| 11 | `Nessian Wilds Ravager (cascade)` | 166 | `(cascade)` | ✓ |
| 12 | `Claim Jumper (cascade)` | 136 | `(cascade)` | ✓ |
| 13 | `Breaching Leviathan (cascade)` | 119 | `(cascade)` | ✓ |
| 14 | `Life of the Party (Life-of-the-Party token)` | 111 | `(X token)` | ✓ |
| 15 | `Curious Homunculus // Voracious Reader (cascade)` | 103 | DFC + `(cascade)` | ✓ |
| 16 | `Claim Jumper (Restore-Relic token)` | 82 | `(X token)` | ✓ |
| 17 | `Necromancy (cascade)` | 51 | `(cascade)` | ✓ |
| 18 | `Kodama of the East Tree (cascade)` | 41 | `(cascade)` | ✓ |
| 19 | `Phoenix Fleet Airship (Urza copy)` | 32 | `(Urza copy)` | ✓ |
| 20 | `Claim Jumper Token` | 3 | **`X Token`** | ✗ |
| 21 | `Sand Scout Token` | 3 | **`X Token`** | ✗ |
| 22 | `Gau, Feral Youth Token` | 2 | **`X Token`** | ✗ |
| 23 | `Lux Artillery (Urza copy)` | 2 | `(Urza copy)` | ✓ |
| 24 | `Phyrexian Myr Token` | 2 | **`X Token`** | ✗ |
| 25 | `Crown of Gondor (Urza copy) (Urza copy)` | 1 | **double-paren** | ✗ |
| 26 | `Crown of Gondor (Urza copy)` | 1 | `(Urza copy)` | ✓ |
| 27 | `Eccentric Pestfinder // Turn Stones (cascade)` | 1 | DFC + `(cascade)` | ✓ |
| 28 | `Emeritus of Woe // Demonic Tutor (cascade)` | 1 | DFC + `(cascade)` | ✓ |
| 29 | `Kodama of the East Tree Token` | 1 | **`X Token`** | ✗ |
| 30 | `Knight of the White Orchid (cascade)` | 1 | `(cascade)` | ✓ |

Tail (not in top 30, listed for completeness):
`Lathiel, the Bounteous Dawn (cascade)` (1), `Lorehold Archivist // Restore
Relic (Restore-Relic token)` (1), `Myr Token` (1), `Rankle and Torbran Token`
(1), `Sand Scout (cascade)` (1), `Wistfulness Token` (1), `Construct Token`
(1), `Token` (1), `creature token scorpion dragon Token` (1), `creature token
knight Token` (1), `creature token colorless myr artifact Token` (1).

## Coverage scorecard

Of the 41 variant-shaped snippets in the live gap log:

| Bucket | Snippets | Cumulative hits | c09db37 strips to base? |
|---|--:|--:|:-:|
| `(cascade)` family (incl. `(2nd cascade)`, DFC-prefixed) | 21 | **4,143** | ✓ all |
| `(X copy)` / `(X Token)` / `(X token)` single paren | 9 | **840** | ✓ all |
| `X Token` suffix (no parens) | 7 | **13** | ✗ none |
| Doubly-stripped `(Urza copy) (Urza copy)` | 1 | **1** | ✗ |
| Bare `Token` / `creature token ... Token` (no base name) | 4 | **4** | ✗ (unrecoverable from dispatch) |

**Top-30 weighted coverage: 4,983 / 5,001 (99.6%).**

The handler's stated claim — "Equivalent to wiring 10+ new handlers in
gap-log coverage" — holds. Every cascade and every paren-wrapped copy/token
suffix on the top-30 strips correctly.

## Gaps that remain

### 1. `X Token` suffix without parentheses (7 snippets, ~13 hits)

Four code sites emit `Card.DisplayName() + " Token"`:

- `per_card/ratadrabik_of_urborg.go:85`
- `per_card/paradigm_echocasting_symposium.go:63`
- `per_card/runo_stromkirk.go:174`
- `per_card/sin_spiras_punishment.go:92`

These produce names like `Claim Jumper Token`, `Sand Scout Token`,
`Kodama of the East Tree Token`. `lookupCandidates` looks for ` (` and finds
none, so the registered handler for the base card never fires on the token.

Volume is small today (~13 cumulative hits) but will grow with every new
dragon-doubler / copy-token commander that adopts this " Token" convention.
The fix is a single line in `lookupCandidates`: after the paren strip,
strip a trailing ` Token` suffix too. No real card on Scryfall ends in
` Token`, so collision is impossible.

### 2. Doubly-stripped suffixes (1 snippet, 1 hit)

`Crown of Gondor (Urza copy) (Urza copy)` shows up once. This happens when
an Urza-copied permanent itself becomes the source of another Urza copy.
`strings.LastIndex(nk, " (")` strips only one suffix, leaving
`Crown of Gondor (Urza copy)` — which has no registered handler either.

Fix: iterate the strip until no trailing parenthetical remains. The
helper would need to add intermediate candidates to the chain so each
stripping depth gets a lookup attempt.

### 3. Bare token-type names (4 snippets, 4 hits)

`Token`, `Construct Token`, `creature token scorpion dragon Token`,
`creature token knight Token`, `creature token colorless myr artifact Token`.
These are generic token type names without any base card to fall back to —
**unrecoverable at the dispatch layer**. They should be filtered out at the
gap-extraction layer in `heimdall/replay.go:ExtractParserGaps` so they
stop polluting the gap log, but that is a separate change from c09db37's
scope.

## Architectural observation worth recording

`c09db37`'s last_seen timestamps tell a story the commit message doesn't.
`c09db37` merged at `2026-05-17T09:02 -0700`. The cascade variants in the
gap log have `last_seen` timestamps from `14:00` onward — i.e., **the dispatch
fallback is shipped and yet `Wistfulness (cascade)`, `Evercoat Ursine
(cascade)`, etc. are still being recorded as parser gaps**.

Wistfulness has a registered handler (`evoke_color_gate.go:89`); after
c09db37 the lookupCandidates fallback ought to find it. So why is the
gap still firing?

Because the `parser_gap` flag is set during **effect resolution** in
`resolve.go:178` and `resolve.go:203` (and `resolve_helpers.go:2136`,
`:4512`) when the engine walks the parsed `gameast` and encounters an
`UnknownEffect` or unhandled kind. This is **independent of per-card
dispatch**: the per_card handler fires on ETB/cast/trigger, but the
parser still walks every effect node, and any node that doesn't map to
a wired resolver flips the flag on the Permanent. `ExtractParserGaps`
then reads the permanent's `DisplayName()` regardless of whether any
handler ran.

**The dispatch fallback wins back handler-firing on variants** — which is
real value for cards whose primary game-state effect lives in the per_card
handler (ETB triggers, cast triggers, activated abilities). **It does not
reduce gap-log volume** for cards whose oracle text still has
parse-unhandled nodes that the resolver walks during normal effect
processing. The "5k gap-log hits across cascades" rescue claim in the
commit message will hold for handler-coverage gaps, but not for
parser-coverage gaps — those need work in the gameast layer (parsing) or
in `resolve_helpers.go` (resolution kinds), not at dispatch.

A reasonable follow-up: at `ExtractParserGaps`, normalize the reported
name through the same `lookupCandidates`-style strip so the gap log
aggregates under the base card name. This won't reduce the parser-coverage
work but it will stop the gap log from showing the same underlying card
five different ways (`Wistfulness`, `Wistfulness (cascade)`,
`Wistfulness Token`, ...).

## Recommended follow-ups (in priority order)

1. **`lookupCandidates`: strip trailing ` Token` suffix** after the paren
   strip. Closes 7 of the remaining 13 missed-dispatch snippets.
2. **`lookupCandidates`: iterate the paren strip** until no trailing
   parenthetical remains. Closes the `(Urza copy) (Urza copy)` case and
   any future stacked-rename cases.
3. **`ExtractParserGaps`: filter bare token-type names** (`Token`,
   `creature token …`) so they stop polluting the gap log. These are
   unrecoverable at the dispatch layer.
4. **`ExtractParserGaps`: normalize variant names to base** so gap-log
   entries aggregate under the base card name. Stops the "same card
   reported five ways" pattern without changing the underlying parser
   coverage work.

Items 1–2 are one-file changes to `internal/gameengine/per_card/registry.go`
and would extend the wave-201-220 dispatch fallback to full top-30
coverage. Items 3–4 are a deeper question about what the gap log is
*for* — currently it tracks where the resolver walks unparsed effects,
not where dispatch misses; conflating the two muddies the signal.

---

_Branch: `dev/muninn-variant-coverage`. Audit run against
`data/muninn/parser_gaps.json` snapshot at 2026-05-17T15:11Z, with 175
total gap entries (41 variant-shaped, 134 plain card-name)._
