# Muninn Status — 2026-05-17 (Snapshot 2)

Second snapshot of the day. `data/muninn/*.json` pulled fresh from
DARKSTAR (`scp josh@192.168.1.207:~/hexdek/data/muninn/*.json
data/muninn/`) at 21:36 UTC, run through
`hexdek-muninn --all --top 50`.

## Headline

- **Parser-gap unique-card count: 167 → 167.** Identical to the morning
  baseline. DARKSTAR is still serving the pre-merge binary, so every
  handler shipped today (waves #1-#12, #13-#20, #21-#30, #31-#40,
  #41-#60, plus the two bulk-pattern families) is invisible to the
  grinder. Counts on covered cards continue to climb at full velocity.
- **Top-50 handler coverage: 47 / 50 (94%) after rounds 1-5.** Only
  three top-50 entries remain unimplemented (see "Top 30 remaining
  uncovered" below).
- **Effective baseline for the next round is the 51-100 band.**
  44 / 50 entries in that range are still uncovered. That's where the
  bulk of the remaining gap-volume lives once DARKSTAR is redeployed.

## 1. Total gap count vs. morning baseline

| Metric                 | Morning (snapshot 1) | Evening (snapshot 2) | Δ |
| ---------------------- | -------------------- | -------------------- | - |
| Parser-gap unique cards | 167                  | 167                  | 0 |
| Top-1 (The One Ring)   | ~1.43M               | 1,446,419            | +~14K |
| Top-2 (Land Tax)       | 604,780              | 611,703              | +6,923 |
| Recurring crashes      | 791                  | 791                  | 0 |
| Dead triggers          | 1                    | 1                    | 0 |
| Concessions records    | ~1,054               | 1,054                | 0 |

No new gap entries between the two snapshots; no count *decreases*
either, because the engine on DARKSTAR is the pre-merge binary. The
expected handler-impact signature — covered-card counts plateau,
`last_seen` pins to the redeploy timestamp — will only appear once
`./scripts/deploy.sh backend` ships the post-merge engine.

## 2. Top 30 remaining uncovered (top 100 by hits, minus shipped)

Ranks are the card's overall position in the live `parser_gaps.json`
top-100. Hit counts are cumulative since `first_seen = 2026-05-05`.

| #  | Rank | Card                                              | Hits   | Notes                              |
| -- | ---- | ------------------------------------------------- | ------ | ---------------------------------- |
|  1 | 43   | Transcendent Dragon                               | 20,144 | Top-50 holdout — only 3 left       |
|  2 | 45   | Titania, Voice of Gaea                            | 18,713 | Top-50 holdout                     |
|  3 | 47   | River Song's Diary                                | 16,900 | Top-50 holdout                     |
|  4 | 53   | Sméagol, Helpful Guide                            | 15,347 |                                    |
|  5 | 56   | Eon Frolicker                                     | 13,197 |                                    |
|  6 | 57   | Tombstone Stairwell                               | 13,124 | Upkeep-trigger token recursion     |
|  7 | 59   | Frodo, Sauron's Bane                              | 12,865 | LotR cycle (Frodo Adventurous shipped) |
|  8 | 61   | Goblin Goliath                                    | 12,155 |                                    |
|  9 | 62   | Yathan Roadwatcher                                | 11,700 |                                    |
| 10 | 63   | Nessian Wilds Ravager                             | 11,575 | Bestow / +1/+1 counters            |
| 11 | 64   | Ichorid                                           | 10,294 | Graveyard upkeep return (family)   |
| 12 | 65   | Breaching Leviathan                               | 10,003 | ETB tap-trigger                    |
| 13 | 66   | Curious Homunculus // Voracious Reader            |  9,770 | Transform DFC                      |
| 14 | 67   | Wary Farmer                                       |  9,682 |                                    |
| 15 | 68   | Bringer of the Last Gift                          |  9,630 |                                    |
| 16 | 69   | Compy Swarm                                       |  9,550 |                                    |
| 17 | 70   | Sarcomancy                                        |  9,507 | Upkeep-cost / zombie token         |
| 18 | 71   | Fearless Swashbuckler                             |  9,337 |                                    |
| 19 | 72   | Dreamcaller Siren                                 |  9,263 | Flash + tap-two-creatures          |
| 20 | 73   | Life of the Party                                 |  9,105 | Party-mechanic scaling             |
| 21 | 74   | Darigaaz Reincarnated                             |  8,947 | Death-counter resurrection         |
| 22 | 75   | Ghitu Journeymage                                 |  8,604 |                                    |
| 23 | 76   | Kami of Transience                                |  8,330 |                                    |
| 24 | 77   | Acclaimed Contender                               |  8,320 | Knight-tribal ETB tutor            |
| 25 | 78   | Feast of the Victorious Dead                      |  7,889 | EOT counter saga                   |
| 26 | 79   | Gau, Feral Youth                                  |  7,849 |                                    |
| 27 | 80   | Lorehold Archivist // Restore Relic               |  7,189 | Adventure-like split DFC           |
| 28 | 81   | Viconia, Drow Apostate                            |  7,109 |                                    |
| 29 | 82   | Emeritus of Woe // Demonic Tutor                  |  7,046 | DFC adventure-flavored             |
| 30 | 83   | Leonardo, Leader in Blue                          |  6,861 |                                    |

All ranks below 84 fall under 7K hits; tail-end mass per card drops
under 2K below rank 100.

## 3. Recommended next tranche

The right-sized next batch is **10 per-card handlers + 1 bulk-pattern
family**, taking advantage of the obvious clusters in this top-30.

### Tranche A — bulk-pattern candidate: Upkeep-graveyard-return cycle

Two clear shapes in the list above share the same engine seam:

- **Ichorid** — upkeep, exile target black creature card from grave,
  return Ichorid from grave to battlefield, exile at EOT.
- **Sarcomancy** — upkeep, sacrifice a Zombie token / take 1 life loss
  (gated on whether token exists), put 2/2 Zombie token onto
  battlefield (ETB).
- **Tombstone Stairwell** — upkeep token-army creation gated on
  graveyards, EOT cleanup. Same upkeep-conditional-zone shape.

All three currently exit the parser without firing a meaningful
effect. A small `upkeep_grave_return_family.go` (similar to
`land_tax_family.go`) with an entry table for source / cost predicate /
zone-result would cover the three above plus the cascade variants. Net
parser-gap reduction estimate: **3-5 entries**.

### Tranche B — 10 per-card handlers, by hit volume

Pure snowflakes that don't share enough shape for a family yet, ordered
by velocity:

1. Transcendent Dragon (20,144) — top-50 holdout
2. Titania, Voice of Gaea (18,713) — top-50 holdout, landfall token
3. River Song's Diary (16,900) — top-50 holdout, cumulative-counter draw
4. Sméagol, Helpful Guide (15,347)
5. Eon Frolicker (13,197)
6. Frodo, Sauron's Bane (12,865)
7. Goblin Goliath (12,155)
8. Yathan Roadwatcher (11,700)
9. Nessian Wilds Ravager (11,575)
10. Breaching Leviathan (10,003)

Closes all three remaining top-50 holdouts (which is the most visible
metric for next snapshot) and the next seven highest-volume snowflakes
in the 51-100 band.

### Rationale

- Closing the three top-50 holdouts gets us to **50/50 coverage** of
  the original task list — clean metric for the next status snapshot.
- The 10-card snowflake batch matches the per-tranche cadence of waves
  #21-#30, #31-#40, #41-#50.
- Pulling Ichorid / Sarcomancy / Tombstone Stairwell out via a single
  family handler is higher leverage than three more per-card files —
  the upkeep-graveyard pattern repeats elsewhere in the long tail
  (Bloodghast / Nether Spirit / Reassembling Skeleton variants will
  benefit from the same scaffold).

## Blocker / next action

Same as snapshot 1: redeploy DARKSTAR with the post-merge binary
(`./scripts/deploy.sh backend`) and let one full overnight grinder
cycle run. Snapshot 3 should be taken immediately after to (a) confirm
the 47 shipped handlers actually flatten their counts, and (b) drop
the top-50 holdouts to zero once Tranche B lands.
