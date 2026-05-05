package hat

import (
	"math"
	"testing"
)

// recordN feeds n games into the detector for deckID with a given win
// rate. Used to seed a population baseline before checking anomalies.
func recordN(d *AnomalyDetector, deckID string, games int, winRate float64) {
	wins := int(math.Round(float64(games) * winRate))
	for i := 0; i < games; i++ {
		d.Record(deckID, i < wins)
	}
}

func TestPerContributorStats_RecordGame(t *testing.T) {
	s := &PerContributorStats{DeckID: "alice"}
	for i := 0; i < 10; i++ {
		s.RecordGame(i%2 == 0) // 5 wins / 5 losses
	}
	if s.Games != 10 {
		t.Fatalf("Games=%d, want 10", s.Games)
	}
	if s.Wins != 5 {
		t.Fatalf("Wins=%d, want 5", s.Wins)
	}
	if got, want := s.WinRate(), 0.5; got != want {
		t.Fatalf("WinRate=%.3f, want %.3f", got, want)
	}
}

func TestPerContributorStats_RollingWindowTrim(t *testing.T) {
	s := &PerContributorStats{DeckID: "bob"}
	// Push more than the window cap; only the last AnomalyRollingWindow
	// entries should remain.
	for i := 0; i < AnomalyRollingWindow+25; i++ {
		s.RecordGame(true)
	}
	if got := len(s.Recent); got != AnomalyRollingWindow {
		t.Errorf("len(Recent)=%d, want %d", got, AnomalyRollingWindow)
	}
	// All-wins → rolling rate is 1.0.
	if got := s.RollingWinRate(); got != 1.0 {
		t.Errorf("RollingWinRate=%.3f, want 1.0", got)
	}
}

func TestPerContributorStats_RollingDivergesFromLifetime(t *testing.T) {
	// Lifetime: 200 games, 50% win rate. Recent window: AnomalyRollingWindow
	// games of all-wins. RollingWinRate should reflect the recent
	// streak, not the lifetime average.
	s := &PerContributorStats{DeckID: "carol"}
	for i := 0; i < 200; i++ {
		s.RecordGame(i%2 == 0)
	}
	for i := 0; i < AnomalyRollingWindow; i++ {
		s.RecordGame(true)
	}
	if got := s.RollingWinRate(); got != 1.0 {
		t.Errorf("RollingWinRate=%.3f, want 1.0 (recent streak)", got)
	}
	// Lifetime rate climbs from 0.5 toward (100+50)/(200+50)=0.6.
	wantLifetime := 150.0 / 250.0
	if got := s.WinRate(); math.Abs(got-wantLifetime) > 0.001 {
		t.Errorf("WinRate=%.3f, want ~%.3f", got, wantLifetime)
	}
}

func TestCheckAnomaly_BelowMinGames_Skips(t *testing.T) {
	d := NewAnomalyDetector()
	// Seed a 2-deck population so n>=2 isn't the failure cause.
	recordN(d, "pop1", 50, 0.30)
	recordN(d, "pop2", 50, 0.30)
	// Probe with a hypothetical at gameCount < AnomalyMinGames.
	if got := d.CheckAnomaly("test", 0.99, AnomalyMinGames-1); got != nil {
		t.Errorf("expected nil flag below AnomalyMinGames; got %+v", got)
	}
}

func TestCheckAnomaly_TooFewPopulationDecks_Skips(t *testing.T) {
	d := NewAnomalyDetector()
	// Only one eligible deck — n<2 means we can't compute a meaningful
	// stddev across the population, so we refuse to flag.
	recordN(d, "pop1", 50, 0.50)
	if got := d.CheckAnomaly("test", 0.99, 50); got != nil {
		t.Errorf("expected nil flag with single-deck population; got %+v", got)
	}
}

func TestCheckAnomaly_ZeroStdDev_Skips(t *testing.T) {
	d := NewAnomalyDetector()
	// Two decks with identical win rates → stddev is 0; any test deck
	// would have an infinite z-score. We refuse to flag.
	recordN(d, "pop1", 50, 0.50)
	recordN(d, "pop2", 50, 0.50)
	if got := d.CheckAnomaly("test", 0.99, 50); got != nil {
		t.Errorf("expected nil flag with zero population stddev; got %+v", got)
	}
}

func TestCheckAnomaly_FlagsHighOutlier(t *testing.T) {
	d := NewAnomalyDetector()
	// Population: 5 decks clustered around 25% win rate (the natural
	// 4-player pod baseline). Small spread.
	recordN(d, "pop1", 100, 0.24)
	recordN(d, "pop2", 100, 0.25)
	recordN(d, "pop3", 100, 0.26)
	recordN(d, "pop4", 100, 0.25)
	recordN(d, "pop5", 100, 0.24)
	// Test deck with 60% win rate — way beyond 3σ.
	flag := d.CheckAnomaly("cheater", 0.60, 100)
	if flag == nil {
		t.Fatalf("expected flag for 60%% WR vs ~25%% population; got nil")
	}
	if flag.DeckID != "cheater" {
		t.Errorf("flag.DeckID=%q, want cheater", flag.DeckID)
	}
	if flag.ZScore <= AnomalyZScore {
		t.Errorf("flag.ZScore=%.2f, want > %.1f", flag.ZScore, AnomalyZScore)
	}
	if flag.WinRate != 0.60 {
		t.Errorf("flag.WinRate=%.3f, want 0.60", flag.WinRate)
	}
	if flag.Games != 100 {
		t.Errorf("flag.Games=%d, want 100", flag.Games)
	}
	if flag.Detected.IsZero() {
		t.Errorf("flag.Detected should be set")
	}
}

func TestCheckAnomaly_FlagsLowOutlier(t *testing.T) {
	d := NewAnomalyDetector()
	recordN(d, "pop1", 100, 0.24)
	recordN(d, "pop2", 100, 0.25)
	recordN(d, "pop3", 100, 0.26)
	recordN(d, "pop4", 100, 0.25)
	recordN(d, "pop5", 100, 0.24)
	// 0% win rate — symmetric outlier on the low side.
	flag := d.CheckAnomaly("doormat", 0.0, 100)
	if flag == nil {
		t.Fatalf("expected flag for 0%% WR vs ~25%% population; got nil")
	}
	if flag.ZScore >= -AnomalyZScore {
		t.Errorf("flag.ZScore=%.2f, want < -%.1f", flag.ZScore, AnomalyZScore)
	}
}

func TestCheckAnomaly_NormalDeck_NotFlagged(t *testing.T) {
	d := NewAnomalyDetector()
	recordN(d, "pop1", 100, 0.24)
	recordN(d, "pop2", 100, 0.26)
	recordN(d, "pop3", 100, 0.25)
	recordN(d, "pop4", 100, 0.27)
	recordN(d, "pop5", 100, 0.23)
	// 28% — within reasonable bounds.
	if flag := d.CheckAnomaly("normal", 0.28, 100); flag != nil {
		t.Errorf("expected no flag for in-bounds WR; got %+v", flag)
	}
}

func TestRecord_AppendsToFlagsOnDetection(t *testing.T) {
	d := NewAnomalyDetector()
	// Seed population.
	recordN(d, "pop1", 100, 0.24)
	recordN(d, "pop2", 100, 0.25)
	recordN(d, "pop3", 100, 0.26)
	recordN(d, "pop4", 100, 0.25)
	recordN(d, "pop5", 100, 0.24)
	// Record cheater going 100% — should add a flag entry.
	for i := 0; i < AnomalyMinGames+10; i++ {
		d.Record("cheater", true)
	}
	flags := d.Flags()
	if len(flags) == 0 {
		t.Fatalf("expected at least one flag in detector.Flags()")
	}
	// Last flag should match cheater.
	last := flags[len(flags)-1]
	if last.DeckID != "cheater" {
		t.Errorf("last flag DeckID=%q, want cheater", last.DeckID)
	}
}

func TestRecord_NoFlagsForNormalDeck(t *testing.T) {
	d := NewAnomalyDetector()
	recordN(d, "pop1", 100, 0.24)
	recordN(d, "pop2", 100, 0.25)
	recordN(d, "pop3", 100, 0.26)
	recordN(d, "pop4", 100, 0.25)
	recordN(d, "pop5", 100, 0.24)
	flagsBefore := len(d.Flags())
	// Record a normal-rate deck.
	for i := 0; i < AnomalyMinGames+10; i++ {
		d.Record("normal", i%4 == 0) // 25%
	}
	if got := len(d.Flags()); got != flagsBefore {
		t.Errorf("normal deck should not raise flags; before=%d after=%d",
			flagsBefore, got)
	}
}

func TestPopulationMean_ExcludesIneligible(t *testing.T) {
	d := NewAnomalyDetector()
	// Two eligible decks at ~0.25 + ~0.50, plus an ineligible deck
	// (Games < AnomalyMinGames) that should NOT contribute to the mean.
	// recordN rounds wins to int (50 × 0.25 → 13 wins → 0.26 actual,
	// not 0.25), so the expected mean is the actual stored rate
	// average: (13/50 + 25/50) / 2 = 0.38.
	recordN(d, "pop1", 50, 0.25)
	recordN(d, "pop2", 50, 0.50)
	recordN(d, "small", AnomalyMinGames-1, 0.99) // excluded
	mean, n := d.PopulationMean()
	if n != 2 {
		t.Errorf("PopulationMean n=%d, want 2 (small excluded)", n)
	}
	want := 0.38
	if math.Abs(mean-want) > 0.001 {
		t.Errorf("PopulationMean=%.3f, want %.3f", mean, want)
	}
}

func TestStats_DefensiveCopy(t *testing.T) {
	d := NewAnomalyDetector()
	d.Record("alice", true)
	d.Record("alice", false)
	snap := d.Stats()
	cp := snap["alice"]
	cp.Games = 999 // mutate the copy
	cp.Recent = append(cp.Recent, true, true)
	// Detector's internal state should be untouched.
	live := d.Stats()
	if live["alice"].Games != 2 {
		t.Errorf("internal Games mutated: got %d, want 2", live["alice"].Games)
	}
	if len(live["alice"].Recent) != 2 {
		t.Errorf("internal Recent mutated: len=%d, want 2", len(live["alice"].Recent))
	}
}

func TestPackageLevelCheckAnomaly_RoutesViaDefault(t *testing.T) {
	// Verifies the package-level CheckAnomaly delegates to
	// DefaultAnomalyDetector. We seed with a unique deckID prefix to
	// avoid bleed from other tests, and use spread rates so the
	// population stddev is non-zero (identical seeds → stddev=0 →
	// the zero-stddev gate trips and no flag is raised).
	const prefix = "pkglvl_test_"
	recordN(DefaultAnomalyDetector, prefix+"a", 50, 0.20)
	recordN(DefaultAnomalyDetector, prefix+"b", 50, 0.25)
	recordN(DefaultAnomalyDetector, prefix+"c", 50, 0.30)
	flag := CheckAnomaly(prefix+"outlier", 0.95, 50)
	if flag == nil {
		t.Fatalf("package-level CheckAnomaly should flag a 95%% WR vs ~25%% population")
	}
	if flag.DeckID != prefix+"outlier" {
		t.Errorf("flag.DeckID=%q, want %s", flag.DeckID, prefix+"outlier")
	}
}
