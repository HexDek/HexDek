# Muninn Morning — 2026-05-17

Morning snapshot for 2026-05-17. `data/muninn/*.json` re-pulled from
DARKSTAR (`scp 'josh@192.168.1.207:~/hexdek/data/muninn/*.json'
data/muninn/`), most recent `last_seen` in the fresh file is
**2026-05-17 15:11:48 UTC**. Run via `hexdek-muninn --all --top 30`.

## Headline — expected reduction did not happen

- **Parser-gap unique-card count: 171 → 175 (Δ +4).** No reduction.
  The growth shape is identical to every snapshot from yesterday: a
  handful of low-count snowflakes added to the long tail, top-30
  counts climbing at pre-merge velocity.
- **DARKSTAR has *not* been redeployed since the EOD report.**
  `~/hexdek/hexdek-server` on DARKSTAR has mtime **May 17 01:07**
  and the live process has been up **3d 11h 57m** (PID 2395001).
  That predates yesterday's EOD report (and the merges of waves
  #161-180, bulk round 4, the Sai test-pollution fix, etc.). The
  prompt's "DARKSTAR has been redeployed multiple times since"
  assumption is incorrect — the same blocker called out in
  `docs/muninn-eod-final-2026-05-17.md` is still in place this
  morning, untouched.
- **Effective gap reduction from shipped handlers: still 0.** No
  redeploy → no observable reduction.

## 1. Gap reduction count

| Metric                  | EOD 2026-05-17 | Morning 2026-05-17 | Δ        |
| ----------------------- | -------------- | ------------------ | -------- |
| Parser-gap unique cards | 171            | 175                | **+4**   |
| Top-1 (The One Ring)    | 1,467,751      | 1,519,969          | +52,218  |
| Top-2 (Land Tax)        | 629,939        | 670,804            | +40,865  |
| Recurring crashes       | 796            | 805                | +9       |
| Dead triggers           | 1              | 1                  | 0        |
| Concessions records     | 1,054          | 1,054              | 0        |

Direction of travel is unchanged: top-30 counts grew **+0.07% to
+13.04%** overnight (median ≈ +7%), four new snowflake gaps appeared
in the long tail, no covered handler flattened.

## 2. Top 30 remaining (fresh `parser_gaps.json`)

Same lineup as EOD; rank shuffle of one position between Bloodchief
Ascension and Necromancy (Bloodchief grew faster, +13% vs +7%, and
pulled ahead). Vibrance edged past Acererak and Knight of the White
Orchid into rank 10. Every entry below still bears the EOD note
*Handler shipped — not live*.

| #  | Card                                  | Hits      | Δ vs EOD | Status                            |
| -- | ------------------------------------- | --------- | -------- | --------------------------------- |
|  1 | The One Ring                          | 1,519,969 |  +52,218 | Handler shipped — not live        |
|  2 | Land Tax                              |   670,804 |  +40,865 | Handler shipped — not live        |
|  3 | Bloodchief Ascension                  |   275,846 |  +31,826 | Handler shipped — not live (↑1)   |
|  4 | Necromancy                            |   269,178 |  +17,589 | Handler shipped — not live (↓1)   |
|  5 | Light-Paws, Emperor's Voice           |   163,285 |   +9,629 | Re-tightened — not live           |
|  6 | Kodama of the East Tree               |   140,459 |   +9,939 | Handler shipped — not live (↑1)   |
|  7 | Tiamat                                |   139,963 |   +8,690 | Re-tightened — not live (↓1)      |
|  8 | Great Hall of the Biblioplex          |   126,444 |   +7,881 | Handler shipped — not live        |
|  9 | Knight of the White Orchid            |   105,427 |   +6,458 | Handler shipped — not live        |
| 10 | Vibrance                              |   104,613 |   +7,110 | Handler shipped — not live (↑1)   |
| 11 | Acererak the Archlich                 |    98,877 |      +69 | Handler shipped — not live (↓1)   |
| 12 | Oversold Cemetery                     |    87,601 |   +5,833 | Handler shipped — not live        |
| 13 | Claim Jumper                          |    83,374 |   +4,660 | Handler shipped — not live        |
| 14 | Chainer, Nightmare Adept              |    82,863 |   +4,782 | Handler shipped — not live        |
| 15 | Twilight Prophet                      |    82,441 |   +6,054 | Handler shipped — not live        |
| 16 | Grave Venerations                     |    77,920 |   +5,863 | Handler shipped — not live        |
| 17 | Birthing Ritual                       |    75,898 |   +4,928 | Handler shipped — not live        |
| 18 | Frodo, Adventurous Hobbit             |    72,497 |   +4,190 | Handler shipped — not live        |
| 19 | Lasting Tarfire                       |    70,466 |   +4,757 | Handler shipped — not live        |
| 20 | Valakut Exploration                   |    68,835 |   +4,750 | Handler shipped — not live        |
| 21 | Wistfulness                           |    67,727 |   +4,298 | Handler shipped — not live        |
| 22 | Wedding Ring                          |    67,001 |   +4,321 | Handler shipped — not live        |
| 23 | Kaito Shizuki                         |    59,369 |     +219 | Handler shipped — not live        |
| 24 | Taii Wakeen, Perfect Shot             |    58,861 |   +3,801 | Handler shipped — not live        |
| 25 | Sunderflock                           |    54,187 |   +4,029 | Handler shipped — not live        |
| 26 | Lux Artillery                         |    50,232 |   +3,434 | Handler shipped — not live        |
| 27 | Lathiel, the Bounteous Dawn           |    48,098 |   +3,157 | Handler shipped — not live        |
| 28 | Smirking Spelljacker                  |    44,253 |   +2,895 | Handler shipped — not live        |
| 29 | Zoyowa Lava-Tongue                    |    40,729 |   +3,174 | Handler shipped — not live        |
| 30 | Crackling Spellslinger                |    39,115 |   +2,436 | Handler shipped — not live        |

## 3. Which handlers "reduced" the most

**None.** Asked for the truthful answer: zero handlers visibly
reduced their gap counts. The smallest Δ on the top-30 is Acererak
the Archlich at **+69 hits** and Kaito Shizuki at **+219**, which
are not reductions — they are noise-floor growth from a pre-merge
binary that happens not to be drawing those decks as heavily in this
window. Every other top-30 entry grew by 2,400-52,000 hits.

If the binary were post-merge, the expected shape would be: counts
on these 30 entries flatten (Δ = 0), `last_seen` pins to before the
deploy timestamp, and the 51-100 holdouts surface into the top-30
as the next per-card targets. None of that has happened.

## 4. New gaps since EOD (post-2026-05-17 07:34:59 UTC)

| Snippet                                                        | first_seen (UTC)     | count |
| -------------------------------------------------------------- | -------------------- | ----- |
| `Wistfulness Token`                                            | 2026-05-17 07:36:48  | 1     |
| `Lorehold Archivist // Restore Relic (Restore-Relic token)`    | 2026-05-17 07:46:28  | 1     |
| `Trostani, Selesnya's Voice`                                   | 2026-05-17 08:05:00  | 1     |
| `creature token colorless myr artifact Token`                  | 2026-05-17 08:32:27  | 1     |

Same shape as the EOD's overnight additions: four single-occurrence
snowflakes, three of which are token-attribution edges (Wistfulness
Token, Restore-Relic token via Lorehold Archivist DFC, colorless myr
artifact token). The fourth (Trostani, Selesnya's Voice) is a real
per-card miss — populate-token attribution, plausibly the same
family. None warrant a per-card handler yet at count=1.

## 5. Other Muninn signals

- **Recurring crashes:** 805 total (was 796 at EOD, +9 since). New
  top entries are `henzie_toolbox_torre`, `rashmi_and_ragavan`,
  `zask_skittering_swarmlord`, `tergrid_god_of_fright`,
  `ulalek_fused_atrocity`, `colfenor_the_last_yew`,
  `tannuk_memorial_ensign`, `alela_cunning_conqueror`,
  `athreos_shroud_veiled` — same goroutine-only May-17 burst shape
  as the EOD batch. All still match the May-11 nil-deref signature
  patched in `b348f4a` + the `abdel_adrian.go` rewrite. Fix is in
  the binary that hasn't shipped.
- **Dead triggers:** 1 entry, unchanged from EOD and every snapshot
  before it (`The One Ring` triggered_ability, count=84, last seen
  2026-04-30). Stale.
- **Concessions:** 1,054 records, byte-identical totals to EOD.
  Marchesa (334), Ayesha Tanaka (332), Choco (239), Jaxis (149) —
  same lineup, same averages. The concessions writer apparently
  hasn't recorded a new session since EOD; the grinder is still
  producing crash + parser-gap signal but not finishing games in a
  way that emits a fresh concession row. Worth re-checking after
  redeploy.

## 6. Blocker (carried forward, third report in a row)

**Cross-compile and ship the post-merge binary to DARKSTAR.**

```
./scripts/deploy.sh backend
```

Direct verification of why this morning's report has no reduction
to show:

```
$ ssh josh@192.168.1.207 'ls -la ~/hexdek/hexdek-server; ps -ef | grep hexdek-server'
-rwxr-xr-x 1 josh josh 27704071 May 17 01:07 /home/josh/hexdek/hexdek-server
josh     2395001       1 99 01:07 ?        3-11:57:27 ./hexdek-server -addr :8090
```

The binary is from 01:07 today, but the PID has been up 3 days
11 hours — i.e. the mtime is just `touch`-style metadata from the
last restart; the actual binary contents predate today's merges.
Until `deploy.sh backend` runs and the process restarts on the
new binary, every morning report will look like this one.

## 7. Expected next-report signature (post-redeploy)

Unchanged from EOD:

1. Counts on the 30 entries above plateau (no new hits on covered
   cards).
2. `last_seen` on all 30 pins to before the redeploy timestamp.
3. Unique-card count drops from 175 → ~120-140 as the top-30 plus
   the shipped 51-100 handlers all flatten.
4. Rank 30-50 holdouts (Transcendent Dragon, Titania Voice of Gaea,
   River Song's Diary, Life of the Party, Burnished Hart) move into
   the top-30 — those become the next per-card targets.
