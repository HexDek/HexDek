# Day Summary — 2026-05-17

End-of-day report covering all commits landed on `main` between
`00:00` and `23:59` local on 2026-05-17. Ten commits total: five
merges and five underlying changesets. Branch: `dev/day-summary`,
cut from `0467f40`.

## Commits by area

### Handlers / per_card (6 commits)

| Hash      | Subject                                                              |
|-----------|----------------------------------------------------------------------|
| `02106e9` | fix(per_card): restore init-registered handlers after Reset()        |
| `375de60` | Merge dev/sai-test-pollution: restore init handlers after Reset()    |
| `d4ad314` | per_card: 2 bulk-pattern families for Muninn gap shapes (round 4)    |
| `eecb1c3` | Merge dev/muninn-bulk-patterns-4: shuffle-self-from-grave + etb-library-tutor families |
| `e85c2aa` | per_card: 1 handler for Muninn parser-gap wave #161-180              |
| `66fb9ec` | Merge dev/muninn-handlers-161-180: 1 handler (gap log saturated)     |

### UI / Frontend (1 commit + merge)

| Hash      | Subject                                                              |
|-----------|----------------------------------------------------------------------|
| `9c1b322` | visual polish r7: gauntlet panel — distinct running badge, 3-tile stat row on mobile, drop duplicate ELO DELTA row |
| `0467f40` | Merge dev/visual-polish-round-7: gauntlet panel polish               |

### Ops / Docs (1 commit + merge)

| Hash      | Subject                                                              |
|-----------|----------------------------------------------------------------------|
| `e6d2e6f` | docs: eod audit 2026-05-17 — muninn/goldilocks/loki                  |
| `cfc0ce5` | Merge dev/eod-audit: EOD stability audit 2026-05-17                  |

No engine, parser (Thor/Odin), tournament, or deploy commits today —
the day was tightly focused on per_card coverage, a UI polish pass,
and stability auditing.

## Biggest wins

### 1. Test-pollution root-cause closed (`02106e9`)

`TestSai_ArtifactCastCreatesThopter` was passing in isolation and
failing in the full per_card suite. Bisection identified
`TestAllRegisteredTriggersAreDispatched` calling `per_card.Reset()`,
which rebuilt the global registry from `registerDefaults()` only.
Sai (and ~50 sibling handlers wired via `init()` in `tribal_lords.go`,
`obeka_support.go`, `combat_restrictions.go`, `batch17_sweep.go`,
and four `zz_*_register.go` files) were permanently stripped after
the first `Reset()` because Go runs each `init()` exactly once per
process.

Fix: added `AddResetHook(fn)` in `registry.go`; `Reset()` now
re-invokes every registered hook on the fresh registry. Each init()
that populates the registry now adds a hook (bodies extracted into
named `registerXxx` functions where needed).

Latent sibling bug caught in the same pass:
`custom_saheeli_radiant_creator.go` registered `OnTrigger("...",
"begin_combat", ...)` but the engine fires `combat_begin`. Added
`"begin_combat": {"combat_begin"}` alias in `event_aliases.go` so
registrations canonicalize to dispatched names.

This is the second test-pollution class bug closed this week —
worth keeping a checklist for future Reset()-style globals.

### 2. Muninn bulk-pattern families round 4 (`d4ad314`)

Two new family handlers covering 13 cards across two recurring
oracle shapes, bringing the bulk-pattern family count from 8 to 10:

- **`shuffle_self_from_grave_family.go`** — Planar Chaos eternal
  cycle (Dread, Purity, Guile, Vigor). Hooks `creature_dies`
  (battlefield→graveyard path). Mill / discard / exile→graveyard
  paths still need a graveyard-arrival hook (documented via
  `emitPartial`, same compromise Worldspine Wurm already makes).
- **`etb_library_tutor_family.go`** — covers 9 ETB tutors with
  pluggable filters on type, subtype, legendary, CMC, and printed
  power (Trophy Mage, Treasure Mage, Trinket Mage, Stoneforge
  Mystic, Heliod's Pilgrim, Spellseeker, Imperial Recruiter,
  Fierce Empath, Thalia's Lancers). Rune-Scarred Demon kept its
  bespoke handler — it wants a highest-CMC chooser.

### 3. Muninn parser-gap log saturated (`e85c2aa`)

After the 141-160 wave merged, the only top-170 parser_gaps entry
without a bespoke or family-backed handler is Sam, Loyal Attendant.
Every other tail entry is covered. **The Muninn signal has
effectively saturated against the current corpus** — additional tail
coverage now requires fresh tournament runs to surface new gaps.

### 4. Gauntlet panel mobile UX (`9c1b322`)

- Running badge now warn-yellow vs complete solid-ok (previously
  both rendered greenish via `.tag--solid` override — you couldn't
  tell run state at a glance).
- New `.gauntlet-stat-grid` keeps WIN RATE / RECORD / ELO DELTA in
  3 columns on mobile (values are short — "0%", "0W—5L", "-27");
  the old `.grid.col-3` collapsed to a vertical stack of three
  full-width slabs that dominated the panel.
- Dropped the duplicate ELO DELTA row from the KV block — already
  has its own color-treated tile.

### 5. End-of-day stability audit clean (`e6d2e6f`)

| Tool        | Result                                            |
|-------------|---------------------------------------------------|
| Muninn      | 170 parser-gap unique cards (Δ0 vs morning)       |
| Goldilocks  | 30,277 / 30,340 pass (99.79%), 0 panics, Δ0       |
| Loki (500g) | 0 crashes, 138 invariant violations in 5 games    |

Zero new crashes, zero new panics, zero goldilocks drift. Loki
violation density per-1000g slightly lower than round-6 (276 vs 376,
within sampling noise). No regression from any of today's handler
waves or test-pollution fix.

## Open / outstanding issues

### Blocker (ops)

- **DARKSTAR still serving pre-merge binary.** The 100+ handlers
  shipped today (and over the prior days) are invisible to the live
  grinder. Gap-count progress on Muninn will not be observable until
  the next `./scripts/deploy.sh backend` redeploy. This was flagged
  in both the morning-final progress report and the EOD audit.

### Known / accepted

- **CardIdentity hotspot in Loki game 170 (seed 1700043).** Same
  4-commander pod (Ezuri / Master of Keys / Syr Ginger / Katara)
  that round-6 surfaced. Signature ("card appears in both hand and
  battlefield") matches the `abdel_adrian.go`-class pattern but does
  not panic and does not match an open handler bug. Worth a trace,
  not a blocker.
- **Goldilocks 64-card failure set unchanged** (61 dead-effect + 2
  invariant + 1 unverified). Same set as round-5 / round-6 /
  post-merge — Tiger-Tribe Hunter, Voidforged Titan, Demonic Hordes,
  Emberheart Challenger, Phage the Untouchable, Lord of
  Tresserhorn, etc.
- **Shuffle-self-from-grave family covers only the
  battlefield→graveyard path.** Mill / discard / exile→graveyard
  trigger paths need a separate graveyard-arrival hook (per-card
  dispatcher iterates battlefield only). Same scoped compromise as
  Worldspine Wurm.
- **Sam, Loyal Attendant partial coverage.** Combat-begin Food
  trigger lands; partner-with ETB and Food-activation cost reduction
  documented as declarative / static-effect gaps via `emitPartial`.
- **4,190 unbucketed condition/trigger nodes** across all 4 eras
  (33.9% of 12,363 total). Standing corpus-audit finding from
  2026-05-08; not touched today.

## Tomorrow's priorities

1. **Deploy the backend** — `./scripts/deploy.sh backend`. Until
   DARKSTAR runs today's binary, none of the handler work landed
   over the last 24+ hours can produce a measurable Muninn signal.
   Single biggest leverage move available.
2. **Post-deploy Muninn re-baseline.** Once the new binary serves
   for a few hundred grinder games, snapshot fresh `data/muninn/*`
   from DARKSTAR and re-rank. Expect the top-30 entries to plateau
   and unique-card gap count to drop from 170 toward ~120–140.
   This is what unlocks the next wave of handler work — the current
   gap log is saturated against pre-merge state.
3. **Loki CardIdentity trace on game 170 / seed 1700043.** Persistent
   non-panic hotspot across rounds 5, 6, and the EOD 500g sweep.
   Targeted repro (`-seed 1700043 -seat 0 -games 1`) is fast; root
   cause likely a per_card handler bypassing canonical zone-change
   plumbing (same shape as the Abdel Adrian fix on 2026-05-16).
4. **Sweep the 6 sibling handlers flagged in
   `docs/may11-nil-deref-forensics.md`** (etrata, zabaz, zimone+dina,
   bilbo, thassa). They don't crash post-`b348f4a` but bypass the
   same machinery Abdel Adrian did and should be routed through
   `ExilePermanent` / canonical battlefield-exit APIs.
5. **Triple-combo cycle ordering** (Freya Kanban "Ready" item).
   Only tests 2/6 orderings — misses valid 3-card combos. Concrete,
   well-scoped, and the next Freya ship now that the 27-item Done
   list from late April through May is fully cleared.

## Stats at a glance

- Commits today: **10** (5 merges, 5 work)
- Handler commits: **6** (2 work, 4 if you count both family
  registrations)
- Cards covered by today's per_card work: **~14** (13 from family
  expansion + 1 bespoke Sam handler)
- Bulk-pattern family count: **8 → 10**
- Test-pollution bugs closed: **1** (with ~50 latently-missing
  handlers restored)
- Goldilocks regression: **0**
- Loki crashes introduced: **0**
- UI polish rounds shipped: **1** (round 7)
- Backend deploys: **0** ← this is the next move
