# Muninn Progress Report — 2026-05-17 (Final)

Final snapshot of the day. `data/muninn/*.json` pulled fresh from
DARKSTAR (`scp josh@192.168.1.207:~/hexdek/data/muninn/*.json
data/muninn/`) at 2026-05-17 06:21 UTC, run through
`hexdek-muninn --all --top 30`.

## Headline

- **Parser-gap unique-card count: 167 → 170 (Δ +3).** Gap count
  *grew* by three over the day rather than shrinking. DARKSTAR is
  still serving the pre-merge `hexdek-server` binary, so every handler
  shipped today (waves #1-#12, #13-#20, #21-#30, #31-#40, #41-#60,
  plus the two bulk-pattern families and the post-merge per-card
  batches) remains invisible to the grinder. Per-card counts on
  covered cards continue to climb at full pre-merge velocity and three
  brand-new snowflake entries first-seen overnight bring the total
  from 167 to 170.
- **Latest `last_seen` in the fresh data: 2026-05-17 06:21 UTC** —
  data is current.
- **Effective gap reduction from shipped handlers: 0.** Reduction
  cannot be measured until DARKSTAR is redeployed. See "Blocker /
  next action".

## 1. Gap reduction count

| Metric                  | Morning baseline | Fresh snapshot  | Δ   |
| ----------------------- | ---------------- | --------------- | --- |
| Parser-gap unique cards | 167              | 170             | +3  |
| Top-1 (The One Ring)    | ~1.43M           | 1,458,634       | +~28K |
| Top-2 (Land Tax)        | ~604K            | 622,221         | +~17K |
| Recurring crashes       | 791              | 794             | +3  |
| Dead triggers           | 1                | 1               | 0   |
| Concessions records     | 1,054            | 1,054           | 0   |

Reduction is **−3** (i.e. growth of 3). No covered card has flattened
because the post-merge engine has not shipped. The three additional
gap entries are all overnight first-seen snowflakes (counts 1, 5, 8 —
see §3).

## 2. Top 30 remaining (fresh `parser_gaps.json`)

Ranks are positions in the live `parser_gaps.json` top-170 by
cumulative hits since `first_seen`. All shipped handlers are present
because the deployed binary hasn't been updated; once redeploy lands,
expect 30+ of these to plateau and drop out of the top-30 immediately.

| #  | Card                                  | Hits      | First seen  | Status                            |
| -- | ------------------------------------- | --------- | ----------- | --------------------------------- |
|  1 | The One Ring                          | 1,458,634 | 2026-05-05  | Handler shipped — not live        |
|  2 | Land Tax                              |   622,221 | 2026-05-05  | Handler shipped — not live        |
|  3 | Necromancy                            |   248,667 | 2026-05-05  | Handler shipped — not live        |
|  4 | Bloodchief Ascension                  |   238,297 | 2026-05-05  | Handler shipped — not live        |
|  5 | Light-Paws, Emperor's Voice           |   151,825 | 2026-05-05  | Re-tightened today — not live     |
|  6 | Tiamat                                |   129,653 | 2026-05-05  | Re-tightened today — not live     |
|  7 | Kodama of the East Tree               |   128,561 | 2026-05-05  | Handler shipped — not live        |
|  8 | Great Hall of the Biblioplex          |   117,145 | 2026-05-05  | Handler shipped — not live        |
|  9 | Acererak the Archlich                 |    98,799 | 2026-05-05  | Handler shipped — not live        |
| 10 | Knight of the White Orchid            |    97,772 | 2026-05-05  | Handler shipped — not live        |
| 11 | Vibrance                              |    96,267 | 2026-05-05  | Handler shipped — not live        |
| 12 | Oversold Cemetery                     |    80,677 | 2026-05-05  | Handler shipped — not live        |
| 13 | Claim Jumper                          |    77,833 | 2026-05-05  | Handler shipped — not live        |
| 14 | Chainer, Nightmare Adept              |    77,209 | 2026-05-05  | Handler shipped — not live        |
| 15 | Twilight Prophet                      |    75,216 | 2026-05-05  | Handler shipped — not live        |
| 16 | Grave Venerations                     |    70,881 | 2026-05-05  | Handler shipped — not live        |
| 17 | Birthing Ritual                       |    70,096 | 2026-05-05  | Handler shipped — not live        |
| 18 | Frodo, Adventurous Hobbit             |    67,444 | 2026-05-05  | Handler shipped — not live        |
| 19 | Lasting Tarfire                       |    64,829 | 2026-05-05  | Handler shipped — not live        |
| 20 | Valakut Exploration                   |    63,239 | 2026-05-05  | Handler shipped — not live        |
| 21 | Wistfulness                           |    62,638 | 2026-05-05  | Handler shipped — not live        |
| 22 | Wedding Ring                          |    61,862 | 2026-05-06  | Handler shipped — not live        |
| 23 | Kaito Shizuki                         |    59,107 | 2026-05-05  | Handler shipped — not live        |
| 24 | Taii Wakeen, Perfect Shot             |    54,383 | 2026-05-05  | Handler shipped — not live        |
| 25 | Sunderflock                           |    49,425 | 2026-05-05  | Handler shipped — not live        |
| 26 | Lux Artillery                         |    46,164 | 2026-05-05  | Handler shipped — not live        |
| 27 | Lathiel, the Bounteous Dawn           |    44,308 | 2026-05-05  | Handler shipped — not live        |
| 28 | Smirking Spelljacker                  |    40,866 | 2026-05-05  | Handler shipped — not live        |
| 29 | Zoyowa Lava-Tongue                    |    36,905 | 2026-05-08  | Handler shipped — not live        |
| 30 | Crackling Spellslinger                |    36,262 | 2026-05-05  | Handler shipped — not live        |

For continuity with the snapshot-2 holdout list, the three top-50
"uncovered" cards from this morning — Transcendent Dragon (rank 43,
20,144 hits at midday), Titania, Voice of Gaea (rank 45), River
Song's Diary (rank 47) — still sit in the rank 30–50 band; they
move into the top-30 only after the post-merge engine flattens the
shipped-handler entries above them.

## 3. New gaps that appeared since morning baseline

Gaps with `first_seen` strictly after the morning report (03:19 UTC)
that were therefore absent from the 167-card baseline:

| Snippet                                    | first_seen (UTC)     | count |
| ------------------------------------------ | -------------------- | ----- |
| `creature token knight Token`              | 2026-05-17 04:58:20  | 1     |
| `Life of the Party (Life-of-the-Party token)` | 2026-05-17 05:46:25 | 8     |
| `Claim Jumper (Restore-Relic token)`       | 2026-05-17 05:47:23  | 5     |

Three new snowflakes, all token-attribution misses:

- **`creature token knight Token`** — single hit, parser failed to
  attribute a Knight creature token to its source (likely a Knight
  of the White Orchid trigger spawning a follow-up token; same family
  as the morning report's generic `Token` entry that first-seen
  2026-05-16).
- **`Life of the Party (Life-of-the-Party token)`** — eight hits.
  Life of the Party is already on the snapshot-2 51-100 uncovered
  list at 9,105 hits; this entry is the *token-side* of its
  party-mechanic scaling gap, distinct from the main card entry.
- **`Claim Jumper (Restore-Relic token)`** — five hits. Cross-card
  attribution: Claim Jumper appears at rank 13 in the top-30 above,
  and this entry pairs it with the Lorehold Archivist // Restore
  Relic adventure-DFC parse path (rank 27 on the morning uncovered
  list). The bulk-pattern family that covers token attribution from
  adventure-flavored split cards will collapse both this entry and
  the underlying Restore Relic gap.

No new high-frequency offenders introduced overnight — all three
first-seens are single-digit counts and all three are token-side
parser misses rather than new card-shape exposure.

## 4. Other Muninn signals

- **Recurring crashes:** 794 total (was 791 at snapshot 2, +3
  overnight). All three new entries are goroutine-stack signatures
  seen 2026-05-17 from `the_second_doctor`, `slicer_hired_muscle`, and
  `zask_skittering_swarmlord` deck configs. Counts and turns=0 plus
  the goroutine-only signature point at the same surface as the May
  11 nil-deref burst (root-caused and patched in `b348f4a` +
  `abdel_adrian.go` rewrite — fix is in the binary that hasn't
  shipped). Once redeploy lands these should not recur.
- **Dead triggers:** 1 entry, unchanged from earlier snapshots —
  `The One Ring` triggered_ability, count=84, last seen 2026-04-30.
  Stale; will clear after redeploy.
- **Concessions:** 1,054 records, identical totals and top-4
  ordering to snapshot 2. Marchesa the Black Rose (334, avg turn
  41.2), Ayesha Tanaka (332, avg turn 39.1), Choco Seeker of
  Paradise (239, avg turn 39.6), Jaxis the Troublemaker (149, avg
  turn 40.3). All avg-turns ≥39 — consistent stall-out pattern, no
  early-scoop signal.

## Blocker / next action

**Same blocker as the morning and snapshot-2 reports**: cross-compile
and ship the post-merge binary to DARKSTAR (`./scripts/deploy.sh
backend`). Until then, no handler reduction is observable and the
gap count will continue drifting upward by 2-4 snowflake entries per
day from token-attribution and cascade-resolution edges.

Next snapshot should be taken immediately after redeploy plus one
full overnight grinder cycle. The expected post-deploy signature:
counts on the 30 entries above plateau, `last_seen` pins to the
redeploy timestamp, and `170 → ~120-140` unique gaps once the
top-30 plus the shipped 51-100 handlers all flatten.
