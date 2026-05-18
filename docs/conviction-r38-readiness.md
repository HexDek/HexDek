# Conviction R38 Re-Enable Readiness Assessment

**Date:** 2026-05-17
**Branch:** `dev/conviction-r38`
**Context:** Round 24 added a non-acting `convictionDiagnostic` + telemetry ring + admin endpoint. Round 38's job: evaluate whether the telemetry now justifies re-enabling `ShouldConcede`.

## TL;DR

**Not ready. The data the readiness gate requires has not been collected yet.**

Round 15's readiness doc (`docs/conviction-reassessment-2026-05-17.md`) was explicit about the criterion:

> 1. **Instrument first, ship later.** Add a *non-acting* `convictionDiagnostic`
>    that records relative position, win-line presence, and lockout state every
>    turn for completed games, then post-hoc analyze: across N games, when
>    would each candidate trigger have fired, and what fraction of those games
>    did the diagnosed-hopeless hat *actually win*?

The instrumentation shipped in Round 24 (this is verifiable — `internal/hat/conviction.go` is non-acting, emits `conviction_diagnostic` events every turn, and pushes structured records into a 1024-entry ring buffer that the admin endpoint serves). The **analysis** has not.

Without the empirical false-positive rate, re-enabling either trigger is the same coin-flip the original removal warned against — just with more accurate score weights underneath, which the audit doc itself flagged as making score-based concession *more* dangerous, not less.

## (a) Current state assessment

### What's wired

| Component | File | State |
|-----------|------|-------|
| `YggdrasilHat.ShouldConcede` | `internal/hat/yggdrasil.go:6374` | Always returns `false`; calls `recordConvictionSample` for diagnostics. |
| `recordConvictionSample` | `internal/hat/conviction.go:70` | Computes both candidate triggers (score-window + win-line-extinct), emits a `conviction_diagnostic` event AND pushes a structured `ConvictionEvent` into the global ring. |
| `convictionRing` (lossy ring buffer, 1024 events) | `internal/hat/conviction_telemetry.go` | Process-wide, mutex-guarded, sequence-numbered. Wraps on overflow. |
| `SnapshotConvictionEvents(sinceSeq, limit)` | `internal/hat/conviction_telemetry.go:123` | Public read API. Returns events newest-trimmed when `limit` is set. |
| `GET /api/admin/conviction-events` | `internal/hexapi/admin_conviction.go` | Admin-gated HTTP endpoint. Accepts `?since=<seq>&limit=<n>&triggered=1`. |
| Engine path (`ConcedeGame`, `LossReason="concession"`, tournament hook) | `internal/tournament/turn.go:73`, `internal/gameengine/hat.go:385` | Intact — only the hat decision is stubbed. Re-enable is a one-function change. |

### What's NOT wired

1. **Aggregation pipeline.** Every game emits ~150–400 diagnostic events into the game log. Nothing in `tournament/runner.go` aggregates these into a post-game summary — there's a `// Track min relative position for conviction calibration` comment at `runner.go:645` with an **empty for-loop body** (SMELL surfaced below). The data is captured but never harvested.

2. **Per-game outcome join key.** `TournamentResult` aggregates wins by commander name across all games; there's no per-seed → winner-seat map. So the standard "trigger fired on this seat in this game → did this seat win this game?" correlation can't be run from `Result` alone without joining through `Result.Analyses` (which requires `AnalyticsEnabled` and adds non-trivial work).

3. **Empirical FP rate calculation.** No script, no test, no dashboard query. The readiness doc's gating analysis (Step 1) is the missing piece.

## (b) Telemetry findings

### Production endpoint

Attempted: `curl https://dev.hexdek.dev/api/admin/conviction-events`
Result: `403 forbidden` (admin auth gate, as designed).

I do not have admin credentials for the dev server, so production telemetry cannot be inspected from this branch. The doc therefore reports **structural readiness only**, not measured behavior.

### Local harvest attempted

A `TestConvictionHarvest_R38` (gated behind `HEXDEK_CONVICTION_HARVEST=1`) was drafted to:

1. Reset the global conviction ring.
2. Run a 4-seat × 60-game tournament with `NewYggdrasilHat`.
3. Snapshot the ring and report per-trigger fire counts + stickiness (samples-per-unique-seat-game).

The test file was repeatedly deleted by concurrent workers operating on the same worktree (see SMELL #7 below); attempts to re-restore it were lost to subsequent branch churn. The structural finding from the partially-completed run is:

- **30-game tournament with `--hat yggdrasil`, `--audit` flag set, 0 concessions.** Confirmed `ShouldConcede` is still gated (`Concessions: 0` in the report). Sample emission happens but the binary doesn't expose the ring to its CLI users.

### Code-level analysis of the two triggers

I can reason about expected fire characteristics from reading the trigger logic:

**Score-window trigger** (`conviction.go:99-108`)
- Fires when every sample in a 4-turn rolling window has `relativePosition < -0.35`, gated on `turn ≥ 10`.
- The doc's original concern stands: this is an "I feel behind" signal, not a "I cannot win" signal.
- The new archetype weights amplify the signal accuracy on the metric it measures — board state — but board state is a poor proxy for win probability in EDH. **Expected pathology: high fire rate on combo/control decks behind on board pre-win-turn.**

**Win-line-extinct trigger** (`conviction.go:165-211`)
- Strict: every named card from `h.Strategy.ComboPieces` AND `h.Strategy.FinisherCards` must be in some seat's exile.
- Conservative-by-design — a single piece in graveyard means "recoverable" and the trigger doesn't fire.
- Disabled entirely when `h.Strategy == nil` (`return false, "no_strategy"`) or when the deck has no declared win lines.
- **Expected pathology: very low fire rate.** This trigger is the "shouldn't false-positive" one but also probably rarely fires in practice.

The asymmetry is significant: the score-window trigger is the dangerous-but-frequent one; the win-line-extinct is the safe-but-rare one. Re-enabling them together brings back exactly the failure mode the original removal addressed.

## (c) Recommended re-enable conditions

The doc's Condition 2 (win-line extinction) is the only candidate that survived first-principles scrutiny. Round 38 should not relax that gate.

**Re-enable READY criteria — all four must hold:**

1. **Empirical FP rate measured.** Either harvest from production (~1000 games), or build an `internal/tournament` harness that runs N≥500 games, snapshots the ring, AND has a per-game winner-seat join. Report:
   - `score_trigger_seat_game_fp_rate` over seat-games (NOT over samples — sticky triggers inflate the sample-level rate).
   - `winline_trigger_seat_game_fp_rate` same.
   - Per-archetype breakdown (combo decks are the riskiest; the doc explicitly flags storm/Tendrils as a class that scores low until lethal).

2. **FP rate ≤ 1%** for the trigger being re-enabled. The original implementation that was removed was killed because a single false positive at 38 life was visibly wrong; FP rates above 1% would mean the same noise at scale. **Only re-enable the trigger(s) that clear this bar.**

3. **A/B winrate test against `return false`.** Per the doc Step 3 explicitly. The threshold of "no winrate distortion" needs to be precise: an A/B run at N≥1000 games per arm with the same deck pool, the same hat budget, the same seeds. Difference in winrate per commander must not exceed sampling noise (~2σ binomial). Concession-eligible commanders losing more than the noise floor = block re-enable.

4. **Telemetry endpoint accessible to the operator running the re-enable.** Admin auth is currently a hard block on doing the analysis from this branch. Either the operator has the credentials, or a one-shot harvest export needs to be plumbed through `cmd/hexdek-tournament` (e.g., `--conviction-harvest path.json` that dumps `SnapshotConvictionEvents` after the run).

**Re-enable NOT-READY criteria — any single one blocks:**

- Score-window FP rate > 1% on the empirical run.
- Score-window fires on any seat that later wins by ≥4 turns post-fire (those are the "topdeck recovery" wins the doc warned against).
- `h.Strategy == nil` for any deck the hat plays. The win-line-extinct trigger silently returns `false` when Strategy is missing, so the re-enable would be a no-op for those decks; we should know how many decks fall in that bucket before shipping.
- Concession rate per game > 5% across the harvest. EDH games are long; many concessions distort tournament throughput accounting.

## (d) Test plan if re-enabling

### Phase A — Harvest

1. Run `HEXDEK_CONVICTION_HARVEST=1 go test ./internal/tournament/ -run TestConvictionHarvest_R38 -count=1 -timeout 1200s` with `N=500` games. (Requires the harvest test landing cleanly — see SMELL #7.)
2. Export the harvest summary to `data/rules/conviction-harvest-r38.txt`.
3. Inspect: stickiness, fire rate per seat-game, fire rate per archetype.

### Phase B — Per-game winner join

If `TournamentResult.Analyses` is enabled, extend the harvest to read `Analyses[i].SeatOutcomes[j].Won` and produce the per-trigger TP/FP confusion matrix. Acceptance: FP rate ≤ 1% per trigger considered for re-enable.

### Phase C — A/B winrate

Run two tournaments back-to-back with identical seeds, identical decks, identical hats:

- Arm 1: `ShouldConcede` returns `false` (current behavior).
- Arm 2: `ShouldConcede` returns the win-line-extinct trigger only.

N≥1000 games per arm. Compare per-commander winrate. Any commander whose winrate drops by more than 2σ in Arm 2 — block re-enable for the entire mechanism.

### Phase D — Targeted scenarios

Hand-built fixtures the harvest can't cover:

- 4-life YggdrasilHat with Worldgorger Dragon + Animate Dead in hand → score trigger should NOT fire, win-line-extinct should NOT fire.
- All combo pieces in opponent's exile, 80 cards left in library, no tutor → win-line-extinct SHOULD fire.
- 3-vs-1 lategame with the 1 holding the win-line in hand → neither trigger should fire.

These three are the original removal's regression scenarios. Codify as tests under `internal/hat/conviction_regression_test.go` so re-enable can't bypass them.

## Smells surfaced during this audit

1. **`runner.go:645-650` — dead loop.** "Track min relative position for conviction calibration" comment with an empty `for _, s := range gs.Seats { if s == nil { continue } }` body. Either complete the calibration (write min-relpos to the per-game record) or delete the placeholder. Currently it claims to do something it doesn't.

2. **Telemetry has no aggregation path.** Round 24 shipped a per-process ring + admin endpoint, but no `cmd/hexdek-tournament` flag to export harvest data, and no test/script that consumes the ring after a Run(). This is the missing piece between "instrument" and "ship."

3. **Branch race / shared worktree.** During this audit the worktree was switched out from under me three times (dev/hat-audit-r37 → dev/conviction-r38 → dev/layers-r38-fix). My harvest test file was deleted twice between writes by concurrent workers landing in the same checkout. The BlueFrog parallel-workers model needs either:
   - Separate worktrees per worker (true isolation), OR
   - A coordination protocol that doesn't checkout-switch with uncommitted other-worker files in the index.
   Surfaced for QUORUM consideration; not blocking on this PR but blocking on the operator wanting to do *any* multi-step single-branch work.

4. **Carry-over from Round 37 Etali task.** Round 37 was pivoted away from before completion. The Etali DFC duplicate-pointer test file (`dfc_transform_identity_test.go`) was never written. **QUORUM REQUEST: Round 37 Etali bug — should I retry it in a follow-up PR before chains of dependent rounds (e.g., any future DFC-touching task) build on top of unfixed behavior?**

## Recommendation

**Do NOT re-enable conviction in Round 38.**

Land this readiness doc as the explicit gate — it converts the doc's "instrument first, ship later" advice into a checklist with measurable thresholds. Add the harvest test under a gating env var so the next operator who can hit the admin endpoint (or run the local harness) has a one-command path to the FP-rate number.

When the data is ready (Phase A–D pass), re-enable ONLY the win-line-extinct trigger first, with the score-window trigger explicitly excluded. The doc's first-principles reasoning stands: a more accurate `relativePosition` makes score-based concession more brittle in EDH variance, not less.
