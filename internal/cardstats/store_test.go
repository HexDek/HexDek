package cardstats

import "testing"

func TestRecord_DedupesByName(t *testing.T) {
	s := NewStore()
	// Deck running 7 Mountains shouldn't tally Mountain 7×.
	deck := []string{"Mountain", "Mountain", "Mountain", "Lightning Bolt"}
	s.Record(deck, true, map[string]int{"lightning bolt": 3})

	rows := s.Snapshot(0)
	got := map[string]CardRow{}
	for _, r := range rows {
		got[r.Name] = r
	}
	if got["Mountain"].Games != 1 {
		t.Fatalf("Mountain.Games=%d, want 1 (dedup)", got["Mountain"].Games)
	}
	if got["Mountain"].Wins != 1 {
		t.Fatalf("Mountain.Wins=%d, want 1", got["Mountain"].Wins)
	}
	if got["Mountain"].TimesPlayed != 0 {
		t.Fatalf("Mountain.TimesPlayed=%d, want 0 (not in firstTurn map)", got["Mountain"].TimesPlayed)
	}
	if got["Lightning Bolt"].TimesPlayed != 1 || got["Lightning Bolt"].AvgTurnPlayed != 3 {
		t.Fatalf("Bolt played stats: %+v", got["Lightning Bolt"])
	}
}

func TestRecord_WinLossAccumulates(t *testing.T) {
	s := NewStore()
	s.Record([]string{"Sol Ring"}, true, nil)
	s.Record([]string{"Sol Ring"}, true, nil)
	s.Record([]string{"Sol Ring"}, false, nil)
	rows := s.Snapshot(0)
	if len(rows) != 1 {
		t.Fatalf("rows=%d, want 1", len(rows))
	}
	r := rows[0]
	if r.Games != 3 || r.Wins != 2 || r.Losses != 1 {
		t.Fatalf("got %+v, want games=3 wins=2 losses=1", r)
	}
	if r.WinRate < 0.66 || r.WinRate > 0.67 {
		t.Fatalf("WinRate=%v, want ~0.667", r.WinRate)
	}
}

func TestSnapshot_RespectsMinGames(t *testing.T) {
	s := NewStore()
	for i := 0; i < 5; i++ {
		s.Record([]string{"Foo"}, true, nil)
	}
	for i := 0; i < 12; i++ {
		s.Record([]string{"Bar"}, false, nil)
	}
	rows := s.Snapshot(10)
	if len(rows) != 1 {
		t.Fatalf("rows=%d, want 1 (Foo<10 filtered)", len(rows))
	}
	if rows[0].Name != "Bar" {
		t.Fatalf("got %q, want Bar", rows[0].Name)
	}
}

func TestTopBottom_OrdersByWinRate(t *testing.T) {
	s := NewStore()
	// 10 wins, 0 losses → 1.0 WR
	for i := 0; i < 10; i++ {
		s.Record([]string{"Best"}, true, nil)
	}
	// 0 wins, 10 losses → 0.0 WR
	for i := 0; i < 10; i++ {
		s.Record([]string{"Worst"}, false, nil)
	}
	// 5/5 → 0.5 WR
	for i := 0; i < 5; i++ {
		s.Record([]string{"Mid"}, true, nil)
		s.Record([]string{"Mid"}, false, nil)
	}
	top, bot := s.TopBottom(2, 10)
	if len(top) != 2 || top[0].Name != "Best" {
		t.Fatalf("top=%+v", top)
	}
	if len(bot) != 2 || bot[0].Name != "Worst" {
		t.Fatalf("bottom=%+v", bot)
	}
}

func TestRecord_FirstTurnTakesEarliest(t *testing.T) {
	s := NewStore()
	// Two casts of the same card: the helper de-dupes by lowercased name,
	// so we expect the smaller turn to win.
	s.Record([]string{"Sol Ring"}, true, map[string]int{"sol ring": 2, "Sol Ring": 4})
	rows := s.Snapshot(0)
	if rows[0].AvgTurnPlayed != 2 {
		t.Fatalf("AvgTurnPlayed=%v, want 2", rows[0].AvgTurnPlayed)
	}
}
