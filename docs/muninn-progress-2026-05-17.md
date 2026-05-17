# Muninn Progress Report — 2026-05-17

Snapshot of `data/muninn/*.json` pulled from DARKSTAR at 2026-05-17 ~03:19 UTC, run through `hexdek-muninn --all --top 20`.

## Headline

**Parser-gap unique-card count: 167 → 167.** No handler impact is observable in the Muninn telemetry yet.

The reason: DARKSTAR's `hexdek-server` binary is timestamped **2026-05-16 11:53** — that predates the two handler merges shipped today (`b28bfdb` Muninn handlers #8-#12, `11ef8d8` Muninn snowflakes #13-#20). The process has been running continuously since 19:36 yesterday on the pre-merge binary, so every game played in the last 24h is still hitting the unpatched parser path.

Until the engine is cross-compiled and redeployed, the gap counts for the cards we wrote handlers for will keep climbing at full velocity. A delta snapshot taken ~2 hours apart shows exactly that — see below.

## Per-card status (cards shipped today)

| Card                                | Baseline (user-cited) | Current     | Δ in last ~2h | Status                             |
| ----------------------------------- | --------------------- | ----------- | ------------- | ---------------------------------- |
| The One Ring                        | ~1.4M                 | 1,431,535   | +25,595       | Still climbing — handler not live  |
| Land Tax                            | ~593K                 | 604,780     | +11,110       | Still climbing — handler not live  |
| Necromancy                          | (n/a)                 | 242,153     | +4,252        | Still climbing                     |
| Bloodchief Ascension                | (n/a)                 | 227,979     | +4,248        | Still climbing                     |
| Light-Paws, Emperor's Voice         | (n/a)                 | 147,875     | +2,619        | Re-tightened today (`2f4e47b`) — not live |
| Tiamat                              | (n/a)                 | 126,019     | +2,254        | Re-tightened today (`2f4e47b`) — not live |
| Kodama of the East Tree             | (n/a)                 | 124,324     | +2,621        | Still climbing                     |
| Great Hall of the Biblioplex (#8)   | (n/a)                 | 113,836     | +2,090        | Handler merged today — not live    |
| Acererak the Archlich (#9)          | (n/a)                 | 97,696      | +1,773        | Handler merged today — not live    |
| Knight of the White Orchid (#10)    | (n/a)                 | 95,029      | +1,626        | Handler merged today — not live    |
| Vibrance (#11)                      | (n/a)                 | 93,671      | +1,734        | Handler merged today — not live    |
| Oversold Cemetery (#12)             | (n/a)                 | 78,345      | +1,445        | Handler merged today — not live    |

Velocity ordering matches frequency ordering — i.e., the engine is still parsing these cards as gaps at the same rate as before the merges. Once DARKSTAR is redeployed, the expected signature of a working handler is: count stops climbing in the next Muninn snapshot, `last_seen` stays pinned at the pre-deploy timestamp.

(Per-card hit *reduction*, the metric the task asked for, is not yet measurable — counts are cumulative since `first_seen = 2026-05-05` and only stop growing once the engine no longer trips the gap path. The pre-deploy total is the baseline against which reduction will be measured in the next report after redeploy.)

## New gaps that appeared

Eight entries first-seen on or after 2026-05-15, all single-hit emergent cases:

| Card / snippet                                         | first_seen  | count |
| ------------------------------------------------------ | ----------- | ----- |
| Bygone Bishop                                          | 2026-05-15  | 1     |
| Ketramose, the New Dawn                                | 2026-05-15  | 1     |
| Phyrexian Dreadnought                                  | 2026-05-15  | 1     |
| Lathiel, the Bounteous Dawn (cascade)                  | 2026-05-15  | 1     |
| Noxious Gearhulk                                       | 2026-05-15  | 1     |
| Token                                                  | 2026-05-16  | 1     |
| Archpriest of Shadows                                  | 2026-05-16  | 1     |
| Eccentric Pestfinder // Turn Stones (cascade)          | 2026-05-17  | 1     |

These are all snowflakes, not recurring gaps — three are cascade/value triggers, two are ETB/death-trigger creatures, one is a generic "Token" snippet (the parser failed to attribute it to a source). No new high-frequency offenders introduced by today's merges.

## Other Muninn signals

- **Recurring crashes:** 791 total. Top entries are all the May 11 nil-deref burst already root-caused and patched in `b348f4a` + the `abdel_adrian.go` rewrite documented in `docs/may11-nil-deref-forensics.md`. Nothing newer than 2026-05-12. No new crash signatures since the last muninn cycle.
- **Dead triggers:** 1 entry — `The One Ring` triggered_ability, count=84, last seen 2026-04-30. Stale; will likely clear after redeploy now that The One Ring has scaffolding.
- **Concessions:** 1,054 records. Top commanders by concession volume: Marchesa the Black Rose (334, avg turn 41.2), Ayesha Tanaka (332, avg turn 39.1), Choco Seeker of Paradise (239), Jaxis the Troublemaker (149). All avg turns ≥39 — consistent with previously logged stall-out pattern rather than early scoops.

## Next action

Cross-compile and ship the post-merge binary to DARKSTAR (`./scripts/deploy.sh backend`), let it run for one full overnight grinder cycle, then re-pull `data/muninn/parser_gaps.json` and measure handler-by-handler velocity drop. Anything in the top-12 above that goes from `+1,000–25,000/2h` to `0` in the next snapshot is a confirmed kill.
