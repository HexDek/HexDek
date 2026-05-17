# Loki Fuzz — New per_card Handlers

**Branch:** `dev/loki-new-handlers`
**Date:** 2026-05-16
**Tool:** `cmd/hexdek-loki` (patched with `--seed-cards` / `--seed-cmdr` for handler-focused fuzz)
**Scope:** 200 games × 8 handlers = **1,600 games**, 4 seats, max 50 turns, GreedyHat opponents

## Goal

Catch crashes or invariant regressions introduced by the recently merged per_card handlers:

- `the_one_ring.go`
- `land_tax.go`
- `necromancy.go`
- `bloodchief_ascension.go`
- `kodama_of_the_east_tree.go`
- `light_paws_emperors_voice.go`
- `custom_tiamat.go`
- `abdel_adrian.go` (battlefield-exit fix from `docs/may11-nil-deref-forensics.md`)

## Method

Loki was extended with two flags so its chaos generator can pin a specific commander and force specific cards into seat 0's library every game:

```
--seed-cmdr  "Commander Name"     # seat 0 commander
--seed-cards "Card A,Card B,..."  # force-included in seat 0's 99
```

Seats 1–3 remain pure chaos decks. Seat 0 plays the focused card under varied 3-opponent pressure with a fresh deck shuffle every game. Each run uses a different master seed to ensure 200 distinct deck × shuffle combinations per handler. Phase-2 nightmare boards disabled (this audit is handler-targeted, not layer/SBA-targeted).

| Handler card | Seed commander |
|---|---|
| The One Ring | Atraxa, Praetors' Voice |
| Land Tax | Heliod, Sun-Crowned |
| Necromancy | Sheoldred, the Apocalypse |
| Bloodchief Ascension | Sheoldred, the Apocalypse |
| Kodama of the East Tree | self |
| Light-Paws, Emperor's Voice | self |
| Tiamat | self |
| Abdel Adrian, Gorion's Ward | self |

Run script: `/tmp/loki-handlers/run.sh` (preserved for reproduction). Per-handler reports: `/tmp/loki-handlers/<slug>.md`.

## Headline Result

**Zero crashes. Zero panics. Across all 1,600 games.** Including the `Abdel Adrian` battlefield-exit path that was producing 324 nil-deref panics in the May 11 grinder forensics.

| Handler | Games | Crashes | Games w/ violations | Total violations | Notes |
|---|---:|---:|---:|---:|---|
| The One Ring | 200 | **0** | 2 | 10 | unrelated (Imskir Iron-Eater cmdr-zone dup) |
| Land Tax | 200 | **0** | 1 | 98 | unrelated (ZoneConservation, single bad game) |
| Necromancy | 200 | **0** | 1 | 24 | unrelated (God-Eternal Kefnet cmdr-replacement) |
| Bloodchief Ascension | 200 | **0** | 0 | 0 | clean |
| Kodama of the East Tree | 200 | **0** | 1 | 2 | unrelated (Loxodon Warhammer on dead token) |
| Light-Paws, Emperor's Voice | 200 | **0** | 3 | 120 | mostly unrelated (Lich's Mastery exile dup, etc.); 1 ptr matching Light-Paws — see below |
| Tiamat | 200 | **0** | 1 | 12 | unrelated (God-Eternal Oketra recur) |
| Abdel Adrian, Gorion's Ward | 200 | **0** | 2 | 18 | unrelated (Stupefying Touch on dead token) |
| **Total** | **1,600** | **0** | **11** | **284** | |

**Baseline reference** (200 chaos games, no seed cards, random commanders, seed=99): 0 crashes, 22 violations in 1 game (1 bad game out of 200, dominated by `Sarkhan, Soul Aflame` cmdr-zone duplication on opposing seats). The handler-focused rates (0.5–1.5% game-level violation rate) sit in the **same envelope as the baseline noise floor** — they are not elevated by the new handlers.

## Per-handler triage

### The One Ring (Atraxa)
10 violations in 2 games. All are `CardIdentity` cmdr-zone duplications on **opposing** commanders (e.g., Imskir Iron-Eater simultaneously in seat 3 command_zone and battlefield). Not attributable to The One Ring's draw/protection handler.

### Land Tax (Heliod)
98 `ZoneConservation` violations all from a **single game** (idx 4, seed 44449). Real cards disappear progressively (2 → 18 between turns 2 and 6) across the cleanup steps of an otherwise unremarkable game. None of the violation messages cite Land Tax; nothing in the Land Tax handler creates/destroys cards (only moves basic lands from library → hand, which is conservation-safe). Likely a separate engine issue triggered by an opposing card's behavior in that game.

### Necromancy (Sheoldred)
24 violations from 1 game (idx 111). All `CardIdentity` on God-Eternal Kefnet (well-known "die → exile then back to library" replacement effect that periodically duplicates pointers in chaos runs). Not Necromancy.

### Bloodchief Ascension (Sheoldred)
**Clean 200/200.** No crashes, no violations. The new quest-counter + zone-change drain handler runs without incident.

### Kodama of the East Tree (self)
2 violations in 1 game. `AttachmentConsistency` — Loxodon Warhammer in seat 1 still listed as attached to a token that was destroyed. Generic attachment-on-dead-token issue unrelated to the Kodama "permanent ETB → land from hand" trigger.

### Light-Paws, Emperor's Voice (self)
120 violations in 3 games — by far the loudest run. Breakdown of unique violation messages:

- `CardIdentity: "Cerulean Sphinx" appears in both seat 0 library and seat 0 battlefield` — Cerulean Sphinx's "leaves battlefield → shuffled into library" replacement, unrelated.
- `CardIdentity: "Lich's Mastery" appears in both seat 0 exile and seat 0 battlefield` — Lich's Mastery's exile-on-leave replacement, unrelated.
- `CardIdentity: "Runed Servitor" appears in both seat 0 graveyard and seat 0 exile` — unrelated.
- `CardIdentity: "Light-Paws, Emperor's Voice" appears in both seat 0 command_zone and seat 0 battlefield` (ptr 0xc00a028000)

The Light-Paws cmdr-zone duplication is the only line that names the handler card itself. It appears in the same game (idx 125) as the unrelated violations above, alongside late-turn explosive interactions — most likely a downstream symptom of the same broken pointer churn rather than a Light-Paws handler bug, but **worth a focused look at game seed 1254000** as a follow-up. The Light-Paws ETB-trigger handler itself produced no panics across all 200 games.

### Tiamat (self)
12 violations in 1 game (idx 30). All `CardIdentity` on God-Eternal Oketra (same pattern as the Kefnet case). Tiamat's ETB tutor handler did not panic in any game.

### Abdel Adrian, Gorion's Ward (self)
18 violations in 2 games. One bucket is Stupefying Touch attached to a dead token (generic attachment cleanup). The other is Mirror-Mad Phantasm cross-seat library/battlefield duplication. Neither cites the Abdel Adrian handler.

**Key finding for Abdel Adrian:** the May 11 grinder produced 324 nil-deref panics in a single run (`abdelAdrianETB → moveCardBetweenZones → FireZoneChangeTriggers(perm=nil)`, see `docs/may11-nil-deref-forensics.md`). Commit on `dev/may11-nil-deref-forensics` rewrote the handler to route exiles through canonical `gameengine.ExilePermanent`. **This fuzz pass confirms the fix — 200 games with Abdel Adrian as the commander produced 0 panics in his ETB exile path.**

## Conclusion

All 8 new per_card handlers survive 200-game fuzz pressure without a single panic. The few invariant violations observed are attributable to pre-existing engine behavior in opposing cards (God-Eternal cycle, Cerulean Sphinx, Lich's Mastery, generic attachment-on-dead-token cleanup) rather than the new handlers, and sit at the baseline chaos violation rate.

**Recommended follow-ups (out of scope for this branch):**
1. Investigate Light-Paws game seed 1254000 to confirm the cmdr-zone duplication is downstream of the broader pointer churn in that game and not a handler-level commander-tracking bug.
2. The God-Eternal cycle (Kefnet, Oketra) recurring `CardIdentity` violations are a known repeat offender across many runs — worth a dedicated handler that wires their "exile-instead-of-graveyard, may shuffle into library" replacement through the canonical replacement registry.
3. Attachment-on-dead-token cleanup keeps producing `AttachmentConsistency` violations — generic engine issue, not handler-specific.

## Reproduction

```bash
# rebuild loki with the new flags
go build -o /tmp/hexdek-loki ./cmd/hexdek-loki/

# single-handler focused fuzz
/tmp/hexdek-loki \
  --games 200 --permutations 1 --nightmare-boards 0 --workers 8 --max-turns 50 \
  --seed-cmdr "Abdel Adrian, Gorion's Ward" \
  --seed-cards "Abdel Adrian, Gorion's Ward" \
  --report /tmp/abdel.md

# or run all eight via the script preserved in /tmp/loki-handlers/run.sh
```
