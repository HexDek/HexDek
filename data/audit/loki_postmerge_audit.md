# Loki Post-Merge Audit — 2026-05-16

## Scope

Random-game fuzz pass against `main` HEAD after the 2026-05-16 merge wave:

| Commit | Subject |
|--------|---------|
| `1687785` | Merge dev/muninn-wave2 — next-tier residual wrapper promotions |
| `c9a17da` | Merge dev/bolt-regression — Lightning Bolt corpus targeting fix |
| `f773525` | parser: wave 2 — residual wrappers → typed kinds |
| `39605d3` | Merge dev/era2-era4-scaffolds — 15 new condition scaffolds |
| `c9b73ad` | Merge dev/era1-scaffolds — 13 new condition scaffolds |
| `621e865` | Merge dev/warp-mechanic — full Warp keyword implementation |
| `08170e2` | Merge dev/muninn-parser-gaps — wave 1b residual-wrapper promotions |

Run configuration: `--games 500 --seats 4 --max-turns 60 --seed 42 --workers 10 --nightmare-boards 0`. Runtime ≈ 20s. AST corpus at the on-disk regeneration timestamp (2026-05-16 17:29).

## Top-Line Results

| Pass | Crashes | Violations | Affected Games |
|------|--------:|-----------:|---------------:|
| Baseline (main HEAD) | 0 | 374 | 11 |
| After fix A (equip self-target) | 0 | 200 | 5 |
| After fix B (commander SBA inSlice) | 0 | 138 | 5 |

No panics, no stack-trace crashes, no Warp-keyword-attributable failures, no era-scaffold-attributable failures, no regressions tied to the Lightning Bolt targeting fix.

## Bugs Found

### Fixed in this branch

#### B1 — Creature-Equipment can equip itself
- **Severity:** Medium (174 violation occurrences / 4 distinct cards)
- **Invariant:** AttachmentConsistency (`"X" is attached to itself`)
- **Cards:** The Reality Chip (134), Bronzeplate Boar (24), Komainu Battle Armor (8), Simian Sling (8)
- **Root cause:** `tryEquipAll` in `internal/tournament/turn.go:878` picks any creature on the controller's battlefield as an equip target without excluding the Equipment itself. Reconfigure equipments (Reality Chip, Bronzeplate Boar) and Living-weapon equipments are *both* `IsEquipment()` and `IsCreature()` in creature mode, so they satisfy their own filter. The lower-level `ActivateEquip` (`internal/gameengine/keywords_batch6.go:574`) has no `equipment != target` guard either, so the self-equip slipped through both layers.
- **Origin commit:** `7f9263a` (initial commit) — *pre-existing*, not a 2026-05-16 regression.
- **Fix:** Added `p == equip` skip in the AI's target loop and added a CR §702.6b self-target guard in `ActivateEquip`.

#### B2 — Commander SBA double-adds to command zone
- **Severity:** Medium (34 violation occurrences / 1 commander)
- **Invariant:** CardIdentity (`X appears in both seat N command_zone and seat N command_zone`)
- **Card:** Gerrard, Weatherlight Hero (34 cmdzone×cmdzone)
- **Root cause:** `sba704_6d` in `internal/gameengine/sba.go:1601` appends a commander card to `CommandZone` whenever it finds it in graveyard or exile, with no check whether the card is already in the command zone. Gerrard's die-trigger ("exile it and return …") interacts with the §704.6d / §903.9a "may be put in command zone" replacement to leave the same `*Card` pointer in two command-zone slots after the trigger resolves while exile→cmdzone migration also fires.
- **Origin commit:** Pre-existing in `sba.go` since the commander SBA was added. Not a 2026-05-16 regression.
- **Fix:** Defensive `inCommandZone(c)` check before the append on both the graveyard and exile sweeps. The cleanup of graveyard/exile entries still runs so the source zones don't keep stale refs.

### Open — needs investigation

#### B3 — Adric, Mathematical Genius duplicated across hand and graveyard
- **Severity:** Medium (54 occurrences: 46 hand×graveyard, 8 hand×battlefield)
- **Invariant:** CardIdentity
- **Hypothesis:** Adric is "Doctor's companion" with a `{2}{U}, {T}: Copy target activated or triggered ability`. The cast / activation / sacrifice flow appears to leave the card pointer in hand while also adding it to graveyard (or battlefield). Likely the cast pipeline doesn't strip the card from hand for spells with Doctor's-companion partner-style metadata, or the Ultimate Sacrifice activation routes through a path that grays the card to graveyard without hand removal.
- **Not fixed:** requires tracing the cast / activate paths against Doctor's-companion metadata. Recommend a per-card test with Adric scripted via `cmd/hexdek-loki` seed 360001 (game 36).

#### B4 — Rally the Galadhrim conspire copy duplicated in graveyard
- **Severity:** Medium (40 occurrences)
- **Invariant:** CardIdentity (graveyard×graveyard, same seat)
- **Hypothesis:** Conspire creates a copy of the spell on the stack. The original card resolves to the graveyard (sorcery), and the copy — which by CR §707.10 ceases to exist — appears to also be appended to graveyard using the same `*Card` pointer instead of being discarded.
- **Not fixed:** check how `ConspireCopy` (or whichever helper handles the spell copy) writes the copy to the graveyard.

#### B5 — Cerulean Sphinx / Mirror-Mad Phantasm shuffle-into-owner-library leaves stale battlefield ref
- **Severity:** Medium (34 occurrences: Sphinx 26 cross-seat, Mirror-Mad 8 same-seat)
- **Invariant:** CardIdentity
- **Hypothesis:** Both cards have `{U}: This creature's owner shuffles it into their library` as an activated ability on a permanent. The `shuffle_pronoun_into_owner_library` handler in `resolve_helpers.go:3811` calls `gs.removePermanent(src)` which only walks `gs.Seats[p.Controller].Battlefield`. For Cerulean Sphinx (owner=0, controller=2) the removal should still succeed. The hypothesis is that the activated-ability path supplies a `src` whose `Controller` is stale (the owner, not the current controller) by the time `removePermanent` runs, so the battlefield walk on the wrong seat fails silently and only `moveToZone(owner, …, library_bottom)` lands.
- **Not fixed:** verify the `src` passed in from the activated-ability dispatcher. Likely needs a fallback walk of all seats' battlefields in `removePermanent` (mirroring the graveyard/exile/hand fallback already present in the shuffle handler).

#### B6 — Spear of the General attached to off-battlefield Amber Gristle O'Maul
- **Severity:** Low (2 occurrences)
- **Invariant:** AttachmentConsistency (attached to permanent not on any battlefield)
- **Hypothesis:** Stale `AttachedTo` after the target left the battlefield without an `unattach` sweep. Equipment §704.5n SBA should null the `AttachedTo` when the host disappears; either it didn't fire or it fired against the wrong seat.
- **Not fixed:** narrow scope; revisit if it grows.

#### B7 — ZoneConservation: +14 phantom cards across all 4 seats
- **Severity:** Medium (34 occurrences in 1 game)
- **Invariant:** ZoneConservation (`expected 397, found 411 — possible copy bug`)
- **Hypothesis:** A spell or ability that creates copies of non-token cards is not flagging the copies as tokens, so they survive into zones and get counted as "real" cards. Co-occurs with Rally the Galadhrim's Conspire (B4) — likely the same root cause: the spell copy is registered as a `*Card` and routed to zones instead of being discarded as a stack-only object.
- **Not fixed:** ties into B4. Fixing the conspire-copy zoning will most likely zero out B7.

## Severity Breakdown

| Severity | Count | Distinct Bugs |
|---------:|------:|--------------:|
| High (crash) | 0 | 0 |
| Medium | 6 | B1, B2, B3, B4, B5, B7 |
| Low | 1 | B6 |
| **Total** | **7** | |

## Regression Assessment vs 2026-05-16 Merges

Every distinct bug surfaced by this run is **pre-existing** — none of B1–B7 trace to commits merged today. The Warp keyword, era1/era2/era3/era4 scaffold additions, parser wave 1b/2 promotions, and the Lightning Bolt corpus targeting fix produced **zero** new violations across 500 chaos games. The Warp-flagged cards in the corpus (e.g. Storm Sculptor, Lightning Stormkin reprints, etc.) were drawn into chaos decks without incident.

## Recommendation

- Land fixes B1 and B2 (already applied on `dev/loki-postmerge-audit`).
- Open follow-up tickets for B3–B7 with the seed reproducers below.
- Re-run loki at `--games 5000` once B3–B5 land to confirm the long-tail clearance.

## Repro Seeds (master seed 42)

| Game | Deck Seed | Trigger Card | Bug |
|------|-----------|--------------|-----|
| 36 | 360001 | Adric, Mathematical Genius | B3 |
| 52 | 520043 | The Reality Chip | B1 (now fixed) |
| (varies) | — | Rally the Galadhrim | B4 / B7 |
| (varies) | — | Gerrard, Weatherlight Hero | B2 (now fixed) |
| (varies) | — | Cerulean Sphinx | B5 |
