package analytics

import (
	"testing"
)

func TestPersistRivalries_CreateAndLoad(t *testing.T) {
	dir := t.TempDir()

	wins := map[string]map[string]int{
		"Alesha":  {"Krenko": 2},
		"Krenko":  {"Alesha": 1},
		"Edgar":   {"Krenko": 1},
	}
	games := map[string]map[string]int{
		"Alesha":  {"Krenko": 3},
		"Krenko":  {"Alesha": 3, "Edgar": 1},
		"Edgar":   {"Krenko": 1},
	}

	if err := PersistRivalries(dir, wins, games); err != nil {
		t.Fatalf("PersistRivalries: %v", err)
	}

	got, err := LoadRivalries(dir)
	if err != nil {
		t.Fatalf("LoadRivalries: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 rivalries, got %d: %+v", len(got), got)
	}

	// Find Alesha-Krenko (canonical: Alesha < Krenko).
	var ak *Rivalry
	for i := range got {
		if got[i].CommanderA == "Alesha" && got[i].CommanderB == "Krenko" {
			ak = &got[i]
		}
	}
	if ak == nil {
		t.Fatalf("Alesha-Krenko rivalry not found")
	}
	if ak.AWins != 2 || ak.BWins != 1 || ak.TotalGames != 3 {
		t.Errorf("Alesha-Krenko: got AWins=%d BWins=%d Total=%d, want 2/1/3", ak.AWins, ak.BWins, ak.TotalGames)
	}
	if ak.LastPlayed == "" {
		t.Errorf("LastPlayed should be set")
	}
}

func TestPersistRivalries_Merge(t *testing.T) {
	dir := t.TempDir()

	first := map[string]map[string]int{"A": {"B": 1}}
	games1 := map[string]map[string]int{"A": {"B": 2}, "B": {"A": 2}}
	if err := PersistRivalries(dir, first, games1); err != nil {
		t.Fatalf("first persist: %v", err)
	}

	second := map[string]map[string]int{"A": {"B": 2}, "B": {"A": 1}}
	games2 := map[string]map[string]int{"A": {"B": 3}, "B": {"A": 3}}
	if err := PersistRivalries(dir, second, games2); err != nil {
		t.Fatalf("second persist: %v", err)
	}

	got, _ := LoadRivalries(dir)
	if len(got) != 1 {
		t.Fatalf("expected 1 rivalry after merge, got %d", len(got))
	}
	r := got[0]
	if r.AWins != 3 || r.BWins != 1 || r.TotalGames != 5 {
		t.Errorf("merged rivalry: got AWins=%d BWins=%d Total=%d, want 3/1/5", r.AWins, r.BWins, r.TotalGames)
	}
}

func TestPersistRivalries_Empty(t *testing.T) {
	dir := t.TempDir()
	if err := PersistRivalries(dir, nil, nil); err != nil {
		t.Errorf("empty input should be no-op, got error: %v", err)
	}
	got, _ := LoadRivalries(dir)
	if len(got) != 0 {
		t.Errorf("expected no rivalries from empty input, got %d", len(got))
	}
}

func TestTopRivals_Sorting(t *testing.T) {
	rivalries := []Rivalry{
		{CommanderA: "Alesha", CommanderB: "Bob", AWins: 1, BWins: 1, TotalGames: 2},
		{CommanderA: "Alesha", CommanderB: "Cora", AWins: 5, BWins: 5, TotalGames: 10},
		{CommanderA: "Alesha", CommanderB: "Dan", AWins: 3, BWins: 2, TotalGames: 5},
		{CommanderA: "Edgar", CommanderB: "Frank", AWins: 7, BWins: 0, TotalGames: 7}, // unrelated
	}

	top := TopRivals(rivalries, "Alesha", 0)
	if len(top) != 3 {
		t.Fatalf("expected 3 rivals, got %d", len(top))
	}
	if top[0].Opponent != "Cora" || top[0].TotalGames != 10 {
		t.Errorf("most-played should be Cora, got %s with %d games", top[0].Opponent, top[0].TotalGames)
	}
	if top[1].Opponent != "Dan" {
		t.Errorf("second should be Dan, got %s", top[1].Opponent)
	}

	// Verify swap-side reflection: when Alesha is CommanderA, AWins are Alesha's wins.
	if top[0].Wins != 5 || top[0].Losses != 5 {
		t.Errorf("Alesha vs Cora: Wins=%d Losses=%d, want 5/5", top[0].Wins, top[0].Losses)
	}
	if top[0].WinRate != 0.5 {
		t.Errorf("WinRate: got %v, want 0.5", top[0].WinRate)
	}
}

func TestTopRivals_LimitN(t *testing.T) {
	rivalries := []Rivalry{
		{CommanderA: "A", CommanderB: "B", TotalGames: 5},
		{CommanderA: "A", CommanderB: "C", TotalGames: 3},
		{CommanderA: "A", CommanderB: "D", TotalGames: 1},
	}
	got := TopRivals(rivalries, "A", 2)
	if len(got) != 2 {
		t.Errorf("expected 2 (capped), got %d", len(got))
	}
	if got[0].Opponent != "B" || got[1].Opponent != "C" {
		t.Errorf("wrong order: %v", got)
	}
}

func TestTopRivals_OpponentSidePerspective(t *testing.T) {
	// Bob is CommanderB — top rivals for Bob should report Alesha's BWins as Bob's wins.
	rivalries := []Rivalry{
		{CommanderA: "Alesha", CommanderB: "Bob", AWins: 4, BWins: 6, TotalGames: 10},
	}
	got := TopRivals(rivalries, "Bob", 0)
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	if got[0].Opponent != "Alesha" || got[0].Wins != 6 || got[0].Losses != 4 {
		t.Errorf("Bob view: got %+v, want Opp=Alesha Wins=6 Losses=4", got[0])
	}
}

func TestTopRivals_NoMatches(t *testing.T) {
	rivalries := []Rivalry{{CommanderA: "X", CommanderB: "Y", TotalGames: 5}}
	got := TopRivals(rivalries, "Z", 0)
	if len(got) != 0 {
		t.Errorf("expected 0 matches for unknown commander, got %d", len(got))
	}
}

func TestLoadRivalries_MissingFile(t *testing.T) {
	dir := t.TempDir()
	got, err := LoadRivalries(dir)
	if err != nil {
		t.Errorf("missing file should not error, got: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}
