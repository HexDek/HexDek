# Muninn EOD Final — 2026-05-17

Closing snapshot for 2026-05-17. `data/muninn/*.json` re-pulled from
DARKSTAR (`scp 'josh@192.168.1.207:~/hexdek/data/muninn/*.json'
data/muninn/`), most recent `last_seen` in the file is
**2026-05-17 07:34:59 UTC**. Run via `hexdek-muninn --all --top 30`.

## Headline

- **Parser-gap unique-card count: 167 → 171 (Δ +4).** Net growth of
  four over the day, not the reduction one would expect from the
  ~100+ per-card handlers and 9 bulk-pattern families landed since
  the morning baseline. **The reason is unchanged from the snapshot-2
  and earlier "final" reports**: DARKSTAR is still serving the
  pre-merge `hexdek-server` binary, so every handler shipped today
  remains invisible to the grinder. Top-30 counts continue to climb
  at pre-merge velocity, and four overnight snowflakes pushed the
  unique-card total from 167 → 171.
- **Effective gap reduction from shipped handlers: 0.** Reduction
  cannot be observed in Muninn until DARKSTAR is redeployed
  (`./scripts/deploy.sh backend`).
- **Engine work is real and ready to go live.** 100+ per-card
  handlers across waves #1-#180 plus 9 bulk-pattern families are
  merged on `main`. Once the redeployed binary runs a single
  overnight grinder cycle, expect a sharp drop in the top-30 (most
  shipped entries plateau, then fall out as covered) and a
  171 → ~120-140 unique-gap floor as the 50-100 holdout range
  collapses too.

## 1. Gap reduction count

| Metric                  | Morning baseline | EOD final       | Δ        |
| ----------------------- | ---------------- | --------------- | -------- |
| Parser-gap unique cards | 167              | 171             | +4       |
| Top-1 (The One Ring)    | ~1.43M           | 1,467,751       | +~38K    |
| Top-2 (Land Tax)        | ~604K            | 629,939         | +~26K    |
| Recurring crashes       | 791              | 796             | +5       |
| Dead triggers           | 1                | 1               | 0        |
| Concessions records     | 1,054            | 1,054           | 0        |

Reduction is **−4** (i.e. growth of 4) on uniques, and every top-30
entry grew by single-digit thousands of hits as the pre-merge binary
keeps re-tripping the same parser branches. No covered card has
flattened. Reduction will only show after redeploy.

## 2. Top 30 remaining (fresh `parser_gaps.json`)

Ranks 9 and 10 swapped vs. the earlier "final" report (Knight of the
White Orchid edged past Acererak the Archlich on cumulative hits).
Every shipped-handler entry remains because the deployed binary has
not been updated.

| #  | Card                                  | Hits      | First seen  | Status                            |
| -- | ------------------------------------- | --------- | ----------- | --------------------------------- |
|  1 | The One Ring                          | 1,467,751 | 2026-05-05  | Handler shipped — not live        |
|  2 | Land Tax                              |   629,939 | 2026-05-05  | Handler shipped — not live        |
|  3 | Necromancy                            |   251,589 | 2026-05-05  | Handler shipped — not live        |
|  4 | Bloodchief Ascension                  |   244,020 | 2026-05-05  | Handler shipped — not live        |
|  5 | Light-Paws, Emperor's Voice           |   153,656 | 2026-05-05  | Re-tightened today — not live     |
|  6 | Tiamat                                |   131,273 | 2026-05-05  | Re-tightened today — not live     |
|  7 | Kodama of the East Tree               |   130,520 | 2026-05-05  | Handler shipped — not live        |
|  8 | Great Hall of the Biblioplex          |   118,563 | 2026-05-05  | Handler shipped — not live        |
|  9 | Knight of the White Orchid            |    98,969 | 2026-05-05  | Handler shipped — not live        |
| 10 | Acererak the Archlich                 |    98,808 | 2026-05-05  | Handler shipped — not live        |
| 11 | Vibrance                              |    97,503 | 2026-05-05  | Handler shipped — not live        |
| 12 | Oversold Cemetery                     |    81,768 | 2026-05-05  | Handler shipped — not live        |
| 13 | Claim Jumper                          |    78,714 | 2026-05-05  | Handler shipped — not live        |
| 14 | Chainer, Nightmare Adept              |    78,081 | 2026-05-05  | Handler shipped — not live        |
| 15 | Twilight Prophet                      |    76,387 | 2026-05-05  | Handler shipped — not live        |
| 16 | Grave Venerations                     |    72,057 | 2026-05-05  | Handler shipped — not live        |
| 17 | Birthing Ritual                       |    70,970 | 2026-05-05  | Handler shipped — not live        |
| 18 | Frodo, Adventurous Hobbit             |    68,307 | 2026-05-05  | Handler shipped — not live        |
| 19 | Lasting Tarfire                       |    65,709 | 2026-05-05  | Handler shipped — not live        |
| 20 | Valakut Exploration                   |    64,085 | 2026-05-05  | Handler shipped — not live        |
| 21 | Wistfulness                           |    63,429 | 2026-05-05  | Handler shipped — not live        |
| 22 | Wedding Ring                          |    62,680 | 2026-05-06  | Handler shipped — not live        |
| 23 | Kaito Shizuki                         |    59,150 | 2026-05-05  | Handler shipped — not live        |
| 24 | Taii Wakeen, Perfect Shot             |    55,060 | 2026-05-05  | Handler shipped — not live        |
| 25 | Sunderflock                           |    50,158 | 2026-05-05  | Handler shipped — not live        |
| 26 | Lux Artillery                         |    46,798 | 2026-05-05  | Handler shipped — not live        |
| 27 | Lathiel, the Bounteous Dawn           |    44,941 | 2026-05-05  | Handler shipped — not live        |
| 28 | Smirking Spelljacker                  |    41,358 | 2026-05-05  | Handler shipped — not live        |
| 29 | Zoyowa Lava-Tongue                    |    37,555 | 2026-05-08  | Handler shipped — not live        |
| 30 | Crackling Spellslinger                |    36,679 | 2026-05-05  | Handler shipped — not live        |

The rank 30-50 band is still dominated by the snapshot-2 holdouts —
Transcendent Dragon, Titania Voice of Gaea, River Song's Diary — and
they only enter the top-30 after the post-merge engine flattens the
shipped-handler entries above them.

## 3. New gaps that appeared today

Gaps with `first_seen` after 2026-05-17 00:00 UTC, absent from the
167-card morning baseline:

| Snippet                                          | first_seen (UTC)     | count |
| ------------------------------------------------ | -------------------- | ----- |
| `Eccentric Pestfinder // Turn Stones (cascade)`  | 2026-05-17 00:55:58  | 1     |
| `creature token knight Token`                    | 2026-05-17 04:58:20  | 1     |
| `Life of the Party (Life-of-the-Party token)`    | 2026-05-17 05:46:25  | 23    |
| `Claim Jumper (Restore-Relic token)`             | 2026-05-17 05:47:23  | 21    |
| `Burnished Hart`                                 | 2026-05-17 07:01:21  | 1     |

Four of the five are single-digit-or-low snowflakes (all
token-attribution or cascade edge cases). Two delta vs. the earlier
"final" report:

- **Eccentric Pestfinder // Turn Stones (cascade)** — adventure-DFC
  cascade-resolution edge, same family as the Claim Jumper /
  Restore-Relic entry. Will collapse into the adventure-token
  bulk-pattern family.
- **Burnished Hart** — newest entry (07:01 UTC, count=1). Single
  occurrence so far; classic land-fetch sacrifice creature, looks
  like a missed sac-cost attribution. Watch for repeats.

The two earlier-today counts climbed since 06:21 UTC: Life of the
Party token went 8 → 23 (+15), Claim Jumper / Restore-Relic token
went 5 → 21 (+16). Same pre-merge grinder still tripping the same
token-attribution branches.

## 4. Other Muninn signals

- **Recurring crashes:** 796 total (was 794 at the previous "final",
  +2 since). The new top entry is `teysa_orzhov_scion_b2_mollymauk76`
  (goroutine-only stack, turns=0, seen 2026-05-17), same shape as
  the existing May-17 burst from `the_second_doctor`,
  `slicer_hired_muscle`, and `zask_skittering_swarmlord`. All share
  the May-11 nil-deref signature already root-caused and patched in
  commit `b348f4a` plus the `abdel_adrian.go` rewrite — fix is in
  the binary that hasn't shipped. Redeploy should clear the trend.
- **Dead triggers:** 1 entry, unchanged across every snapshot today
  (`The One Ring` triggered_ability, count=84, last seen
  2026-04-30). Stale; will clear after redeploy.
- **Concessions:** 1,054 records, identical totals and top-4
  ordering to every prior snapshot today. Marchesa the Black Rose
  (334, avg turn 41.2), Ayesha Tanaka (332, avg turn 39.1), Choco
  Seeker of Paradise (239, avg turn 39.6), Jaxis the Troublemaker
  (149, avg turn 40.3). All avg-turns ≥39 — consistent stall-out
  pattern, no early-scoop signal.

## Day-in-review: engine work landed on `main`

Even though Muninn cannot see it yet, the day's per-card and
bulk-pattern progress is substantial. Walking `git log` since
yesterday's EOD:

- Per-card handler waves: #61-#80, #81-#100, #101-#120, #121-#140,
  #141-#160, #161-#180 (plus top-50 + 51-60 holdout closure).
  Roughly 60+ individual per-card handlers shipped today on top of
  the pre-existing waves, putting day-total handler additions at
  100+.
- Bulk-pattern families: rounds 3 and 4 — `shuffle-self-from-grave`
  and `etb-library-tutor` are the two most recent (commits
  `eecb1c3`, `d4ad314`); rounds 1-2 landed earlier in the week.
  Nine families now active.
- Sai test-pollution fix (`375de60` / `02106e9`) — `Reset()` now
  re-invokes `init()`-registered handlers via the new
  `AddResetHook(fn)` mechanism. Logged in CLAUDE.md issue resolution
  for 2026-05-17.
- Goldilocks: 1,795 keyword_dead failures driven to 0 earlier this
  week (combat-scaffold + `RetainEvents`) is still holding clean.
- Loki: 2000-game chaos run on the full corpus (round 6) still at
  0 panics on the pre-merge binary; post-merge has more handlers
  but also more guard-paths, will be re-verified after redeploy.

## Blocker / next action

**Single blocker carried across every snapshot today**: cross-compile
and ship the post-merge binary to DARKSTAR.

```
./scripts/deploy.sh backend
```

Until that happens, no handler reduction is observable in Muninn and
the gap count will drift upward by 2-5 snowflake entries per day
from token-attribution, adventure-DFC, and cascade-resolution edges.

Next Muninn snapshot should be taken **immediately after redeploy
plus one full overnight grinder cycle**. Expected post-deploy
signature:

1. Counts on the 30 entries above plateau (no new hits accumulate
   on covered cards).
2. `last_seen` on all 30 pins to before the redeploy timestamp.
3. Unique-card count drops from 171 → ~120-140 as the top-30 plus
   the shipped 51-100 handlers all flatten.
4. The rank 30-50 holdouts (Transcendent Dragon, Titania Voice of
   Gaea, River Song's Diary, Life of the Party, Burnished Hart) move
   into the top-30 — those become tomorrow's per-card targets.
