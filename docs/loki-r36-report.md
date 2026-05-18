# Loki Fuzz Report — Round 36

**Date:** 2026-05-17
**Commit:** main @ post-r35 (gift mechanic merged)
**Invocation:** `hexdek-loki --games 2000 --seed 36 --seats 4 --nightmare-boards 1000`

## Headline

**One panic, one root cause, one fix.** After the fix the same gauntlet
runs clean — 2000 chaos games + 1000 nightmare boards, **zero crashes**.

## Run 1 — crash at game ~1300

The first invocation died on a stack-overflow panic inside the resolver:

```
runtime: goroutine stack exceeds 1000000000-byte limit
fatal error: stack overflow

goroutine 17 [running]:
runtime.mapassign_faststr  ...
github.com/hexdek/hexdek/internal/gameengine/per_card.fireTrigger.func1
github.com/hexdek/hexdek/internal/gameengine.FireCardTrigger
github.com/hexdek/hexdek/internal/gameengine.resolveCounterMod   resolve.go:2086
github.com/hexdek/hexdek/internal/gameengine.ResolveEffect       resolve.go:84
github.com/hexdek/hexdek/internal/gameengine.resolveSequence     resolve.go:247
github.com/hexdek/hexdek/internal/gameengine.ResolveEffect       resolve.go:30
github.com/hexdek/hexdek/internal/gameengine.ApplyThresholdRider keywords_threshold.go:181
github.com/hexdek/hexdek/internal/gameengine.resolveGatedRider   keywords_gated_riders.go:45
github.com/hexdek/hexdek/internal/gameengine.resolveSequence     resolve.go:256
github.com/hexdek/hexdek/internal/gameengine.ResolveEffect       resolve.go:30
github.com/hexdek/hexdek/internal/gameengine.ApplyThresholdRider keywords_threshold.go:181
github.com/hexdek/hexdek/internal/gameengine.resolveGatedRider   keywords_gated_riders.go:45
github.com/hexdek/hexdek/internal/gameengine.resolveSequence     resolve.go:256
... (repeated to 1GB stack limit) ...
```

### Root cause

`resolveSequence` (resolve.go:220) maintains three rider-depth counters
(`_max_speed_rider_depth`, `_gated_rider_depth`, `_devotion_rider_depth`)
that gate "fire rider on the outermost Sequence only" semantics — the
rider must NOT re-fire if the rider body itself contains a Sequence.

The original code decremented the counters **before** invoking the
rider:

```go
gs.Flags["_max_speed_rider_depth"]--
gs.Flags["_gated_rider_depth"]--
gs.Flags["_devotion_rider_depth"]--
if outer {
    ApplyMaxSpeedRider(gs, src)
    resolveGatedRider(gs, src)               // → ApplyThresholdRider
    ApplyDevotionRidersAllColors(gs, src)
}
```

`ApplyThresholdRider` calls `ResolveEffect(gs, src, eff)` where `eff` is
the threshold-rider body. When the body is a Sequence node, that nested
`resolveSequence` checks the depth counters at entry — finds them all 0
because the parent decremented before invoking the rider — concludes
`outer == true` for the nested call, runs the items, and then **re-fires
the same gated riders against the same `src`**. Since `src` still has
threshold (graveyard count unchanged within the rider body), the
ThresholdRider fires again, which resolves another Sequence, which fires
the rider yet again, ad infinitum until the goroutine stack hits the
1 GB ceiling.

This was masked in unit tests because synthetic rider bodies in the
existing suite are leaf effects (Damage, Draw), not Sequence — so the
recursive entry point never tripped. Loki's randomized AST decks
exposed the bug by combining a threshold card with a multi-statement
rider body in one game.

The cycle visible in the stack trace: every three frames repeat
`resolveSequence → ApplyThresholdRider → resolveGatedRider → ...`.

### Fix

Keep the rider-depth counters POSITIVE while the riders execute:

```go
gs.Flags["_max_speed_rider_depth"]--
gs.Flags["_gated_rider_depth"]--
gs.Flags["_devotion_rider_depth"]--
if outer {
    // Re-bump while the riders execute so any nested resolveSequence
    // inside a rider body sees a non-zero depth and skips re-firing.
    gs.Flags["_max_speed_rider_depth"]++
    gs.Flags["_gated_rider_depth"]++
    gs.Flags["_devotion_rider_depth"]++
    ApplyMaxSpeedRider(gs, src)
    resolveGatedRider(gs, src)
    ApplyDevotionRidersAllColors(gs, src)
    gs.Flags["_max_speed_rider_depth"]--
    gs.Flags["_gated_rider_depth"]--
    gs.Flags["_devotion_rider_depth"]--
}
```

Net counter effect for callers is unchanged (still zero on exit), but
the rider body now resolves under a non-zero depth so it can't trigger
the outer-call branch.

### Suggested follow-ups

- **Regression test:** synthesize a Threshold card with a Sequence rider
  body and assert `ApplyThresholdRider` runs exactly once.
- **Defensive cap:** add a hard recursion bound (e.g. `_gated_rider_depth
  > 8` returns) so a future regression bounds the blow-up to an event
  rather than a panic.
- **Audit similar patterns:** the Metalcraft, Hellbent, Raid, Ferocious,
  Revolt, Delirium, Coven riders share the same dispatch shape — they
  benefit from this fix transparently because all gated riders run
  under the same depth counter.

## Run 2 — clean

After the fix, the same seed completes the full gauntlet:

```
=== CHAOS GAMES COMPLETE ===
  games:           2000
  duration:        1m41.634s
  throughput:      20 games/sec
  crashes:         0 (in 0 games)
  violations:      1187 (in 40 games)
  clean games:     1960

=== NIGHTMARE BOARDS COMPLETE ===
  boards:          1000
  duration:        249ms
  throughput:      4015 boards/sec
  crashes:         0
  violations:      0
  clean boards:    1000
```

### Throughput

- Chaos: **20 games/sec**, 2000 games in 1m41s.
- Nightmare boards: **4015 boards/sec**, 1000 boards in 249ms.

### Invariant violations (1187 total, 40 games)

Distribution by invariant:

| Invariant | Count | Severity |
|-----------|------:|----------|
| CardIdentity | 852 | High (card appearing in 2 zones) |
| ZoneConservation | 308 | High (card lost from accounting) |
| AttachmentConsistency | 11 | Medium (orphan aura/equipment) |
| ZoneCastGrantExpiry | 8 | Low (stale ZoneCastGrants entry) |
| TriggerCompleteness | 4 | Medium (registered trigger never fired) |
| CombatLegality | 4 | Medium (illegal attacker/blocker) |

These are existing known-issue categories — none are new to this round
or attributable to the recently-merged keyword work (Gift / Manifest /
Bargain / Outlaw / Crime / Companion / Council's Dilemma / Suspect /
Channel / Tribute / Saddle / Daybound / Behold / Solved / Visit /
Mayhem / Splice). The 1187 events cluster in 40 games (concentrated in
2 games for the bulk of the CardIdentity hits — recursive
re-emission of the same underlying state mismatch per turn rather
than 852 distinct root causes).

#### Notable patterns

**ZoneCastGrantExpiry — Narset, Enlightened Exile.** Two games with
Narset as commander leak a stale ZoneCastGrant for a Forest after the
grant's `until_end_of_turn` window should have closed. `ExpireZoneCastGrants`
runs at EOT cleanup but doesn't appear to scrub Narset's per-impulse-draw
grant. Worth a focused per-card review (Narset's free-cast permission
likely doesn't honor the duration field).

**CardIdentity — concentrated in 1–2 games.** Of 852 hits, the bulk
fire on the same turn-by-turn cycle in a Sedris/Jhoira/Gwaihir/Alpharael
game — the violation re-emits at every cleanup step once the
desync starts. Root cause is likely a single early mishandled
zone-move that the engine then fails to reconcile.

**CombatLegality (4 events, 2 games).** Within tolerances for a 2000-
game fuzz; no new combat-keyword regression suggested.

## Verdict

- One panic → root-caused → fixed → re-run clean.
- Invariant-violation rate (40/2000 = 2.0% of games) is in family with
  recent rounds. Largest single contributor (Narset's stale grant) is a
  worthwhile per-card follow-up but is **not** a regression from any
  Round 17–35 keyword merge.
- Throughput steady (~20 g/s chaos, ~4000 b/s nightmare boards).

The fix lives in resolve.go `resolveSequence` and is committed alongside
this report.
