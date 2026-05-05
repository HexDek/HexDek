package achievements

import (
	"path/filepath"
	"testing"
	"time"
)

func mkGame(turns int, seats ...SeatOutcome) GameOutcome {
	return GameOutcome{
		Turns:      turns,
		Seats:      seats,
		FinishedAt: time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC),
	}
}

func badgeIDs(detail []EarnedDetail) []string {
	ids := make([]string, 0, len(detail))
	for _, d := range detail {
		ids = append(ids, d.ID)
	}
	return ids
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

func TestFirstBloodOnFirstWin(t *testing.T) {
	tr, err := NewTracker(t.TempDir())
	if err != nil {
		t.Fatalf("NewTracker: %v", err)
	}

	if err := tr.OnGameComplete(mkGame(7,
		SeatOutcome{Owner: "alice", DeckKey: "alice/burn", Won: true, FinalLife: 12},
		SeatOutcome{Owner: "bob", DeckKey: "bob/control"},
	)); err != nil {
		t.Fatalf("OnGameComplete: %v", err)
	}

	got := badgeIDs(tr.Snapshot("alice").Badges)
	if !contains(got, "first_blood") {
		t.Errorf("expected first_blood after first win, got %v", got)
	}
	// Bob lost — must not earn first_blood.
	if contains(badgeIDs(tr.Snapshot("bob").Badges), "first_blood") {
		t.Errorf("bob lost; should not have first_blood")
	}
}

func TestComebackAndPerfectSweepLifeBands(t *testing.T) {
	tr, _ := NewTracker(t.TempDir())

	// Comeback: winner at 3 life.
	_ = tr.OnGameComplete(mkGame(11,
		SeatOutcome{Owner: "alice", DeckKey: "alice/burn", Won: true, FinalLife: 3},
		SeatOutcome{Owner: "bob"},
	))
	if !contains(badgeIDs(tr.Snapshot("alice").Badges), "comeback_sub5") {
		t.Errorf("expected comeback_sub5 at 3 life")
	}
	if contains(badgeIDs(tr.Snapshot("alice").Badges), "perfect_sweep") {
		t.Errorf("3-life win must not award perfect_sweep")
	}

	// Perfect sweep: winner at 40 life.
	_ = tr.OnGameComplete(mkGame(8,
		SeatOutcome{Owner: "carol", DeckKey: "carol/aggro", Won: true, FinalLife: 40},
		SeatOutcome{Owner: "dan"},
	))
	if !contains(badgeIDs(tr.Snapshot("carol").Badges), "perfect_sweep") {
		t.Errorf("expected perfect_sweep at 40 life")
	}
}

func TestEarlyWinAndLongHaulTurnBands(t *testing.T) {
	tr, _ := NewTracker(t.TempDir())

	_ = tr.OnGameComplete(mkGame(4,
		SeatOutcome{Owner: "alice", DeckKey: "alice/burn", Won: true, FinalLife: 28},
		SeatOutcome{Owner: "bob"},
	))
	if !contains(badgeIDs(tr.Snapshot("alice").Badges), "early_win") {
		t.Errorf("expected early_win for turn-4 win")
	}

	_ = tr.OnGameComplete(mkGame(22,
		SeatOutcome{Owner: "alice", DeckKey: "alice/burn", Won: true, FinalLife: 17},
		SeatOutcome{Owner: "bob"},
	))
	if !contains(badgeIDs(tr.Snapshot("alice").Badges), "long_haul") {
		t.Errorf("expected long_haul for turn-22 win")
	}
}

func TestStreakBadges(t *testing.T) {
	tr, _ := NewTracker(t.TempDir())

	for i := 0; i < 4; i++ {
		_ = tr.OnGameComplete(mkGame(8,
			SeatOutcome{Owner: "alice", DeckKey: "alice/burn", Won: true, FinalLife: 14},
			SeatOutcome{Owner: "bob"},
		))
	}
	if contains(badgeIDs(tr.Snapshot("alice").Badges), "iron_grip") {
		t.Errorf("iron_grip should not award before 5-streak")
	}

	_ = tr.OnGameComplete(mkGame(8,
		SeatOutcome{Owner: "alice", DeckKey: "alice/burn", Won: true, FinalLife: 14},
		SeatOutcome{Owner: "bob"},
	))
	if !contains(badgeIDs(tr.Snapshot("alice").Badges), "iron_grip") {
		t.Errorf("expected iron_grip at 5-streak")
	}

	// A loss must reset the streak (so further wins won't immediately re-trigger).
	_ = tr.OnGameComplete(mkGame(8,
		SeatOutcome{Owner: "alice", DeckKey: "alice/burn"},
		SeatOutcome{Owner: "bob", Won: true, FinalLife: 22},
	))
	got := tr.GetOwner("alice")
	if got.CurrentWinStreak != 0 {
		t.Errorf("expected current streak 0 after loss, got %d", got.CurrentWinStreak)
	}
	if got.MaxWinStreak != 5 {
		t.Errorf("expected max streak preserved at 5, got %d", got.MaxWinStreak)
	}
}

func TestOpponentMilestones(t *testing.T) {
	tr, _ := NewTracker(t.TempDir())

	for i := 0; i < 9; i++ {
		oppName := "opp" + string(rune('A'+i))
		_ = tr.OnGameComplete(mkGame(8,
			SeatOutcome{Owner: "alice", DeckKey: "alice/burn", Won: true, FinalLife: 18},
			SeatOutcome{Owner: oppName},
		))
	}
	if contains(badgeIDs(tr.Snapshot("alice").Badges), "ten_users") {
		t.Errorf("ten_users awarded too early at 9 unique opponents")
	}

	_ = tr.OnGameComplete(mkGame(8,
		SeatOutcome{Owner: "alice", DeckKey: "alice/burn"},
		SeatOutcome{Owner: "newcomer", Won: true, FinalLife: 18},
	))
	if !contains(badgeIDs(tr.Snapshot("alice").Badges), "ten_users") {
		t.Errorf("expected ten_users at 10 unique opponents")
	}

	// Repeat opponent must not double-count.
	pre := tr.Snapshot("alice").OpponentsFaced
	_ = tr.OnGameComplete(mkGame(8,
		SeatOutcome{Owner: "alice", DeckKey: "alice/burn"},
		SeatOutcome{Owner: "newcomer", Won: true, FinalLife: 18},
	))
	if got := tr.Snapshot("alice").OpponentsFaced; got != pre {
		t.Errorf("opponent count grew on repeat opponent: %d → %d", pre, got)
	}
}

func TestPersistenceRoundTrip(t *testing.T) {
	dir := t.TempDir()
	tr, _ := NewTracker(dir)
	_ = tr.OnGameComplete(mkGame(4,
		SeatOutcome{Owner: "alice", DeckKey: "alice/burn", Won: true, FinalLife: 28},
		SeatOutcome{Owner: "bob"},
	))

	// File for alice should be on disk.
	if _, err := filepath.Abs(filepath.Join(dir, "alice.json")); err != nil {
		t.Fatalf("abs: %v", err)
	}

	tr2, err := NewTracker(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	got := tr2.Snapshot("alice")
	if got.TotalWins != 1 || got.TotalGames != 1 {
		t.Errorf("reloaded state lost counters: %+v", got)
	}
	if !contains(badgeIDs(got.Badges), "first_blood") {
		t.Errorf("reloaded state lost first_blood: %v", badgeIDs(got.Badges))
	}
}

func TestAwardIsIdempotent(t *testing.T) {
	tr, _ := NewTracker(t.TempDir())

	for i := 0; i < 3; i++ {
		_ = tr.OnGameComplete(mkGame(8,
			SeatOutcome{Owner: "alice", DeckKey: "alice/burn", Won: true, FinalLife: 3},
			SeatOutcome{Owner: "bob"},
		))
	}
	count := 0
	for _, d := range tr.Snapshot("alice").Badges {
		if d.ID == "comeback_sub5" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected comeback_sub5 awarded exactly once, got %d", count)
	}
}

func TestSnapshotIncludesSortedCatalog(t *testing.T) {
	tr, _ := NewTracker(t.TempDir())
	snap := tr.Snapshot("nobody")
	if len(snap.Catalog) != len(Catalog) {
		t.Fatalf("snapshot catalog len %d != Catalog len %d", len(snap.Catalog), len(Catalog))
	}
	for i := 1; i < len(snap.Catalog); i++ {
		if rarityOrder(snap.Catalog[i-1].Rarity) > rarityOrder(snap.Catalog[i].Rarity) {
			t.Errorf("catalog not sorted by rarity at index %d (%s > %s)",
				i, snap.Catalog[i-1].Rarity, snap.Catalog[i].Rarity)
		}
	}
}
