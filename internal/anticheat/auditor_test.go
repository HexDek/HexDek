package anticheat

import (
	"context"
	"database/sql"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "audit.db") + "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func newTestAuditor(t *testing.T) *StatisticalAuditor {
	t.Helper()
	a, err := NewStatisticalAuditor(openTestDB(t))
	if err != nil {
		t.Fatalf("new auditor: %v", err)
	}
	return a
}

// playFairContributor records `games` games for the given contributor
// with a target win rate and turn count drawn from the supplied rng.
// Turns are drawn from N(turnMean, 4) — a wide-enough band that
// per-contributor sample-variance estimates have realistic spread,
// matching production where game lengths range from ~5 (combo wins)
// to ~30+ (control mirror) turns.
func playFairContributor(t *testing.T, a *StatisticalAuditor, id string, games int, winRate float64, turnMean float64, rng *rand.Rand) {
	t.Helper()
	for i := 0; i < games; i++ {
		won := rng.Float64() < winRate
		turns := int(turnMean + rng.NormFloat64()*4.0)
		if turns < 1 {
			turns = 1
		}
		_, err := a.RecordGame(context.Background(), Game{
			ContributorID: id, Won: won, Turns: turns,
		})
		if err != nil {
			t.Fatalf("record %s game %d: %v", id, i, err)
		}
	}
}

// findActiveFlag is a test helper that returns the active flag for one
// (contributor, metric) pair, or fails the test if there isn't exactly
// one.
func findActiveFlag(t *testing.T, a *StatisticalAuditor, contributor, metric string) Flag {
	t.Helper()
	flags, err := a.ListFlags(context.Background(), true, 0)
	if err != nil {
		t.Fatalf("list flags: %v", err)
	}
	var hits []Flag
	for _, f := range flags {
		if f.ContributorID == contributor && f.Metric == metric {
			hits = append(hits, f)
		}
	}
	if len(hits) != 1 {
		t.Fatalf("expected exactly 1 active flag for (%s, %s), got %d (%+v)", contributor, metric, len(hits), hits)
	}
	return hits[0]
}

// TestIdenticalPopulationNoFlags — when every contributor has IDENTICAL
// stats, the population stddev is zero and no flag can fire by design.
// This is the strongest no-false-positives guarantee available without
// depending on RNG luck (a normal-distribution test would still produce
// the textbook ~0.27% false positives per metric, which is a feature of
// 3-sigma thresholding, not a bug).
func TestIdenticalPopulationNoFlags(t *testing.T) {
	a := newTestAuditor(t)
	// 12 contributors, each playing the EXACT same alternating
	// win/loss/win/loss pattern with constant 14-turn games. Result:
	// every contributor has wins=20, games=40, total_turns=560,
	// total_turns_sq=14²·40 = 7840. Stddev across the population is 0.
	pattern := []bool{true, false, false, false} // 25% win rate
	for i := 0; i < 12; i++ {
		id := "owner-" + string(rune('a'+i))
		for game := 0; game < 40; game++ {
			if _, err := a.RecordGame(context.Background(), Game{
				ContributorID: id,
				Won:           pattern[game%4],
				Turns:         14,
			}); err != nil {
				t.Fatalf("record: %v", err)
			}
		}
	}

	flags, err := a.ListFlags(context.Background(), false, 0)
	if err != nil {
		t.Fatalf("list flags: %v", err)
	}
	if len(flags) != 0 {
		for _, f := range flags {
			t.Logf("unexpected flag: %+v", f)
		}
		t.Fatalf("expected 0 flags when entire population is identical, got %d", len(flags))
	}
}

// TestFairPopulationLowFlagRate — under a realistic fair distribution
// the flag count stays low. We don't assert zero (3-sigma thresholding
// produces a known ~0.27% false-positive rate per metric — that's the
// definition of the test, not a defect), but the rate should be
// bounded. With 25 contributors × 3 metrics over a wide-noise
// population, "many" flags would indicate a regression.
func TestFairPopulationLowFlagRate(t *testing.T) {
	a := newTestAuditor(t)
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 25; i++ {
		id := "owner-" + string(rune('a'+i%26)) + string(rune('z'-i/26))
		playFairContributor(t, a, id, 60, 0.25, 14, rng)
	}

	flags, _ := a.ListFlags(context.Background(), true, 0)
	// Hard ceiling: more than one flag per peer would mean the system
	// is over-firing. With 25 peers and 3 metrics, even worst-case
	// statistical noise should leave us comfortably under 25.
	if len(flags) > 12 {
		for _, f := range flags {
			t.Logf("flag: %+v", f)
		}
		t.Fatalf("fair population produced %d flags — over the noise ceiling", len(flags))
	}
	t.Logf("fair-population flag rate: %d/%d contributors flagged on at least one metric",
		distinctContributors(flags), 25)
}

func distinctContributors(flags []Flag) int {
	seen := map[string]bool{}
	for _, f := range flags {
		seen[f.ContributorID] = true
	}
	return len(seen)
}

// TestExtremeWinRateFlags — a contributor whose win rate is impossibly
// high relative to the population must be flagged.
func TestExtremeWinRateFlags(t *testing.T) {
	a := newTestAuditor(t)
	rng := rand.New(rand.NewSource(7))

	// Build a fair population at 25% win rate.
	for i := 0; i < 10; i++ {
		id := "owner" + string(rune('a'+i))
		playFairContributor(t, a, id, 50, 0.25, 14, rng)
	}

	// Cheater wins ~95% of their games — far above population mean.
	playFairContributor(t, a, "cheater", 50, 0.95, 14, rng)

	flags, err := a.ListFlags(context.Background(), true, 0)
	if err != nil {
		t.Fatalf("list flags: %v", err)
	}
	// New design keeps a single active flag per (contributor, metric) —
	// repeated audits escalate the existing row, not insert new ones.
	cheaterWinRateFlags := 0
	for _, f := range flags {
		if f.ContributorID == "cheater" && f.Metric == MetricWinRate {
			cheaterWinRateFlags++
			if f.ZScore < ZThreshold {
				t.Errorf("expected positive z >= %.1f for cheater win rate, got %.2f", ZThreshold, f.ZScore)
			}
			if f.Severity < 1 || f.Severity > MaxSeverity {
				t.Errorf("severity out of range: %d", f.Severity)
			}
		}
	}
	if cheaterWinRateFlags != 1 {
		t.Fatalf("expected exactly 1 active win_rate flag for cheater, got %d (flags: %+v)", cheaterWinRateFlags, flags)
	}
}

// TestExtremeTurnLengthFlags — a contributor whose games are
// dramatically shorter (instant-win pattern) than the population must
// be flagged on avg_game_length.
func TestExtremeTurnLengthFlags(t *testing.T) {
	a := newTestAuditor(t)
	rng := rand.New(rand.NewSource(11))

	// Population at ~14-turn games.
	for i := 0; i < 10; i++ {
		id := "owner" + string(rune('a'+i))
		playFairContributor(t, a, id, 50, 0.25, 14, rng)
	}

	// Cheater finishes every game in 3-4 turns.
	playFairContributor(t, a, "speedrun", 50, 0.25, 3, rng)

	flags, err := a.ListFlags(context.Background(), true, 0)
	if err != nil {
		t.Fatalf("list flags: %v", err)
	}
	hit := false
	for _, f := range flags {
		if f.ContributorID == "speedrun" && f.Metric == MetricAvgGameLength {
			hit = true
			if f.ZScore > -ZThreshold {
				t.Errorf("expected strongly negative z for speedrun length, got %.2f", f.ZScore)
			}
		}
	}
	if !hit {
		t.Fatalf("expected speedrun to be flagged on avg_game_length, got flags: %+v", flags)
	}
}

// TestSeverityEscalation — repeat audits on the same (contributor,
// metric) bump severity in place until it plateaus at MaxSeverity. The
// flag row count stays at 1 (single active row per metric); only the
// severity field climbs. Once a flag is resolved, the next flag starts
// fresh at severity 1 — exercised by TestSeverityResetsAfterResolve.
func TestSeverityEscalation(t *testing.T) {
	a := newTestAuditor(t)
	rng := rand.New(rand.NewSource(91))

	for i := 0; i < 10; i++ {
		id := "owner" + string(rune('a'+i))
		playFairContributor(t, a, id, 50, 0.25, 14, rng)
	}

	// Cheater enters: many extreme games. Severity should climb from
	// 1 to MaxSeverity over (MaxSeverity-1) flagged audits.
	for i := 0; i < 60; i++ {
		if _, err := a.RecordGame(context.Background(), Game{
			ContributorID: "cheater", Won: true, Turns: 14,
		}); err != nil {
			t.Fatalf("record: %v", err)
		}
	}

	flags, err := a.ListFlags(context.Background(), true, 0)
	if err != nil {
		t.Fatalf("list flags: %v", err)
	}
	winRateFlags := 0
	maxSev := 0
	for _, f := range flags {
		if f.ContributorID == "cheater" && f.Metric == MetricWinRate {
			winRateFlags++
			if f.Severity > maxSev {
				maxSev = f.Severity
			}
		}
	}
	if winRateFlags != 1 {
		t.Fatalf("expected exactly 1 active win_rate flag (escalation in place), got %d", winRateFlags)
	}
	if maxSev != MaxSeverity {
		t.Errorf("expected severity to plateau at MaxSeverity=%d, got %d", MaxSeverity, maxSev)
	}
}

// TestSeverityResetsAfterResolve — once an admin resolves a flag, the
// next flag on the same (contributor, metric) starts at severity 1.
// Past flags survive in the history so reviewers can audit decisions.
func TestSeverityResetsAfterResolve(t *testing.T) {
	a := newTestAuditor(t)
	rng := rand.New(rand.NewSource(202))

	for i := 0; i < 10; i++ {
		playFairContributor(t, a, "owner"+string(rune('a'+i)), 50, 0.25, 14, rng)
	}
	for i := 0; i < 40; i++ {
		if _, err := a.RecordGame(context.Background(), Game{
			ContributorID: "cheater", Won: true, Turns: 14,
		}); err != nil {
			t.Fatalf("record: %v", err)
		}
	}

	// Find and resolve the cheater's active win_rate flag specifically
	// — picking by index would risk grabbing an organic outlier from
	// the fair population.
	original := findActiveFlag(t, a, "cheater", MetricWinRate)
	if err := a.ResolveFlag(context.Background(), original.ID, "alice", "investigated"); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	// Continue extreme play. The audit fires on every game and would
	// escalate any active flag, so we play exactly 1 game to verify
	// "the next flag starts fresh at severity 1" without conflating
	// the reset with subsequent escalation.
	if _, err := a.RecordGame(context.Background(), Game{
		ContributorID: "cheater", Won: true, Turns: 14,
	}); err != nil {
		t.Fatalf("record after resolve: %v", err)
	}

	fresh := findActiveFlag(t, a, "cheater", MetricWinRate)
	if fresh.ID == original.ID {
		t.Errorf("expected a NEW flag row after resolve, but got same ID %d", fresh.ID)
	}
	if fresh.Severity != 1 {
		t.Errorf("post-resolve flag should reset to severity 1, got %d", fresh.Severity)
	}

	// Total flag history should be ≥2 (one resolved cheater flag, one
	// fresh active cheater flag, plus possibly organic outliers from
	// the fair population).
	all, _ := a.ListFlags(context.Background(), false, 0)
	winRateForCheater := 0
	for _, f := range all {
		if f.ContributorID == "cheater" && f.Metric == MetricWinRate {
			winRateForCheater++
		}
	}
	if winRateForCheater != 2 {
		t.Errorf("expected 2 total win_rate flags for cheater in history (1 resolved + 1 active), got %d", winRateForCheater)
	}
}

// TestResolveFlagClearsActiveSeverity — resolving a flag should both
// remove it from the only_active list and stop it counting toward
// future severity escalation.
func TestResolveFlagClearsActiveSeverity(t *testing.T) {
	a := newTestAuditor(t)
	rng := rand.New(rand.NewSource(13))

	for i := 0; i < 10; i++ {
		playFairContributor(t, a, "owner"+string(rune('a'+i)), 50, 0.25, 14, rng)
	}
	playFairContributor(t, a, "cheater", 50, 0.95, 14, rng)

	active, err := a.ListFlags(context.Background(), true, 0)
	if err != nil {
		t.Fatalf("list active: %v", err)
	}
	if len(active) == 0 {
		t.Fatal("expected at least one active flag after extreme cheater play")
	}

	// Resolve the first flag.
	first := active[0].ID
	if err := a.ResolveFlag(context.Background(), first, "alice", "false positive"); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	// Re-resolving the same flag must error — the audit history is
	// append-only, so a double-resolve would erase the original review.
	if err := a.ResolveFlag(context.Background(), first, "bob", ""); err == nil {
		t.Error("expected error when resolving an already-resolved flag")
	}
	if err := a.ResolveFlag(context.Background(), 999999, "bob", ""); err == nil {
		t.Error("expected error when resolving a nonexistent flag id")
	}

	// Active list shrinks; full list still has it.
	activeAfter, _ := a.ListFlags(context.Background(), true, 0)
	if len(activeAfter) >= len(active) {
		t.Errorf("active list should have shrunk, before=%d after=%d", len(active), len(activeAfter))
	}
	all, _ := a.ListFlags(context.Background(), false, 0)
	foundResolved := false
	for _, f := range all {
		if f.ID == first {
			foundResolved = true
			if f.ResolvedAt == nil {
				t.Error("resolved_at should be set on the resolved flag")
			}
			if f.ResolvedBy != "alice" || f.ResolvedNote != "false positive" {
				t.Errorf("resolved_by/resolved_note not persisted: by=%q note=%q", f.ResolvedBy, f.ResolvedNote)
			}
		}
	}
	if !foundResolved {
		t.Error("resolved flag should still appear when only_active=false")
	}
}

// TestNotEnoughGamesNoFlag — below MinGames, a contributor cannot be
// flagged regardless of how anomalous their early results look.
func TestNotEnoughGamesNoFlag(t *testing.T) {
	a := newTestAuditor(t)
	rng := rand.New(rand.NewSource(5))

	// Build a population first so the eligibility check has peers.
	for i := 0; i < 10; i++ {
		playFairContributor(t, a, "owner"+string(rune('a'+i)), 50, 0.25, 14, rng)
	}

	// Newcomer wins 5 games in a row — 100% win rate, but only 5 games.
	for i := 0; i < 5; i++ {
		flags, err := a.RecordGame(context.Background(), Game{
			ContributorID: "newbie", Won: true, Turns: 14,
		})
		if err != nil {
			t.Fatalf("record: %v", err)
		}
		if len(flags) > 0 {
			t.Fatalf("newcomer with %d games should never flag (got %+v)", i+1, flags)
		}
	}
}

// TestStatsAccumulation — verify the running counters match what we'd
// compute from a fresh in-memory walk of the same inputs.
func TestStatsAccumulation(t *testing.T) {
	a := newTestAuditor(t)
	games := []Game{
		{ContributorID: "x", Won: true, Turns: 10},
		{ContributorID: "x", Won: false, Turns: 14},
		{ContributorID: "x", Won: true, Turns: 12},
		{ContributorID: "x", Won: false, Turns: 18},
	}
	for _, g := range games {
		if _, err := a.RecordGame(context.Background(), g); err != nil {
			t.Fatalf("record: %v", err)
		}
	}
	s, ok, err := a.GetStats(context.Background(), "x")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !ok {
		t.Fatal("contributor x should exist")
	}
	if s.Games != 4 || s.Wins != 2 {
		t.Errorf("counters wrong: games=%d wins=%d", s.Games, s.Wins)
	}
	if s.AvgGameLength() != 13.5 {
		t.Errorf("avg length = %v, want 13.5", s.AvgGameLength())
	}
	// Hand-computed sample variance of [10,14,12,18]:
	// mean=13.5, deviations 3.5,-0.5,-1.5,4.5 → squares 12.25,0.25,2.25,20.25
	// sum=35.0, sample var = 35/3 ≈ 11.6667
	got := s.TurnVariance()
	want := 35.0 / 3.0
	if diff := got - want; diff < -1e-6 || diff > 1e-6 {
		t.Errorf("variance = %v, want ~%v", got, want)
	}
}

// TestThresholdInjection — drive the auditor with relaxed thresholds
// to verify SetThresholds works (used by the smoke test in the
// commit notes; also useful for ops tuning).
func TestThresholdInjection(t *testing.T) {
	a := newTestAuditor(t)
	a.SetThresholds(5, 1.5) // tiny eligibility, very loose z
	a.now = func() time.Time { return time.Unix(1700000000, 0) }
	rng := rand.New(rand.NewSource(3))
	for i := 0; i < 5; i++ {
		playFairContributor(t, a, "p"+string(rune('a'+i)), 5, 0.25, 14, rng)
	}
	playFairContributor(t, a, "spike", 5, 1.0, 14, rng)

	flags, _ := a.ListFlags(context.Background(), true, 0)
	if len(flags) == 0 {
		t.Fatalf("expected at least one flag with relaxed thresholds")
	}
	for _, f := range flags {
		if f.DetectedAt.Unix() != 1700000000 {
			t.Errorf("DetectedAt should reflect injected clock, got %v", f.DetectedAt)
		}
	}
}
