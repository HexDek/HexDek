package analytics

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

func TestExtractKillRecords_CommanderDamage(t *testing.T) {
	commanders := []string{"Alesha", "Bob", "Cora", "Dan"}
	events := []gameengine.Event{
		{Kind: "damage", Seat: 0, Target: 1, Source: "Alesha, Who Smiles at Death", Amount: 21,
			Details: map[string]interface{}{"commander": true, "combat": true}},
		{Kind: "seat_eliminated", Seat: 1,
			Details: map[string]interface{}{"reason": "21+ commander damage from Alesha", "turn": 7}},
	}
	records := ExtractKillRecords(events, 4, commanders, 0, "g1")
	if len(records) != 1 {
		t.Fatalf("expected 1 kill, got %d", len(records))
	}
	r := records[0]
	if r.KillerCommander != "Alesha" || r.VictimCommander != "Bob" {
		t.Errorf("wrong actors: %+v", r)
	}
	if r.Method != "commander_damage" {
		t.Errorf("method: got %q, want commander_damage", r.Method)
	}
	if r.LethalCard != "Alesha, Who Smiles at Death" {
		t.Errorf("lethal card: got %q", r.LethalCard)
	}
	if r.Turn != 7 || r.GameID != "g1" {
		t.Errorf("turn/gameid: got %d/%s", r.Turn, r.GameID)
	}
}

func TestExtractKillRecords_LifeTotalZero(t *testing.T) {
	commanders := []string{"A", "B"}
	events := []gameengine.Event{
		{Kind: "damage", Seat: 0, Target: 1, Source: "Lightning Bolt", Amount: 40,
			Details: map[string]interface{}{"combat": false}},
		{Kind: "seat_eliminated", Seat: 1,
			Details: map[string]interface{}{"reason": "life total 0"}},
	}
	records := ExtractKillRecords(events, 2, commanders, 0, "")
	if len(records) != 1 {
		t.Fatalf("expected 1, got %d", len(records))
	}
	if records[0].Method != "noncombat_damage" {
		t.Errorf("method: got %q", records[0].Method)
	}
	if records[0].LethalCard != "Lightning Bolt" {
		t.Errorf("lethal: got %q", records[0].LethalCard)
	}
}

func TestExtractKillRecords_Concession(t *testing.T) {
	commanders := []string{"A", "B"}
	events := []gameengine.Event{
		{Kind: "seat_eliminated", Seat: 1,
			Details: map[string]interface{}{"reason": "concession"}},
	}
	records := ExtractKillRecords(events, 2, commanders, 0, "")
	if len(records) != 0 {
		t.Errorf("concession should produce no kill record, got %d", len(records))
	}
}

func TestExtractKillRecords_Empty(t *testing.T) {
	if got := ExtractKillRecords(nil, 4, []string{"A", "B", "C", "D"}, 0, ""); len(got) != 0 {
		t.Errorf("empty events should return empty, got %d", len(got))
	}
	if got := ExtractKillRecords([]gameengine.Event{{Kind: "damage"}}, 4, nil, 0, ""); len(got) != 0 {
		t.Errorf("empty commanders should return empty, got %d", len(got))
	}
}

func TestPersistThreatGraph_CreateAndLoad(t *testing.T) {
	dir := t.TempDir()
	records := []KillRecord{
		{KillerCommander: "Alesha", VictimCommander: "Bob", Method: "combat_damage",
			LethalCard: "Lightning Bolt", Turn: 5, Timestamp: "2026-01-01T00:00:00Z"},
		{KillerCommander: "Alesha", VictimCommander: "Bob", Method: "commander_damage",
			LethalCard: "Alesha", Turn: 8, Timestamp: "2026-01-02T00:00:00Z"},
		{KillerCommander: "Bob", VictimCommander: "Cora", Method: "combat_damage",
			LethalCard: "Bob, the Banker", Turn: 6, Timestamp: "2026-01-03T00:00:00Z"},
	}

	if err := PersistThreatGraph(dir, records); err != nil {
		t.Fatalf("PersistThreatGraph: %v", err)
	}

	edges, err := LoadThreatGraph(dir)
	if err != nil {
		t.Fatalf("LoadThreatGraph: %v", err)
	}
	if len(edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(edges))
	}

	// Find Alesha→Bob.
	var ab *ThreatEdge
	for i := range edges {
		if edges[i].KillerCommander == "Alesha" && edges[i].VictimCommander == "Bob" {
			ab = &edges[i]
		}
	}
	if ab == nil {
		t.Fatalf("Alesha→Bob edge missing")
	}
	if ab.Kills != 2 {
		t.Errorf("kills: got %d, want 2", ab.Kills)
	}
	if ab.MethodBreakdown["combat_damage"] != 1 || ab.MethodBreakdown["commander_damage"] != 1 {
		t.Errorf("method breakdown: %+v", ab.MethodBreakdown)
	}
	if len(ab.TopLethalCards) != 2 {
		t.Errorf("top lethal cards: got %d, want 2", len(ab.TopLethalCards))
	}
}

func TestPersistThreatGraph_Merge(t *testing.T) {
	dir := t.TempDir()
	first := []KillRecord{
		{KillerCommander: "A", VictimCommander: "B", Method: "combat_damage", LethalCard: "Sword", Timestamp: "t1"},
	}
	if err := PersistThreatGraph(dir, first); err != nil {
		t.Fatalf("first persist: %v", err)
	}

	second := []KillRecord{
		{KillerCommander: "A", VictimCommander: "B", Method: "combat_damage", LethalCard: "Sword", Timestamp: "t2"},
		{KillerCommander: "A", VictimCommander: "B", Method: "commander_damage", LethalCard: "A", Timestamp: "t3"},
	}
	if err := PersistThreatGraph(dir, second); err != nil {
		t.Fatalf("second persist: %v", err)
	}

	edges, _ := LoadThreatGraph(dir)
	if len(edges) != 1 {
		t.Fatalf("expected 1 merged edge, got %d", len(edges))
	}
	if edges[0].Kills != 3 {
		t.Errorf("kills: got %d, want 3", edges[0].Kills)
	}
	if edges[0].MethodBreakdown["combat_damage"] != 2 {
		t.Errorf("combat_damage count: got %d, want 2", edges[0].MethodBreakdown["combat_damage"])
	}
	// Sword should not be duplicated.
	swordCount := 0
	for _, c := range edges[0].TopLethalCards {
		if c == "Sword" {
			swordCount++
		}
	}
	if swordCount != 1 {
		t.Errorf("Sword duplicated %d times", swordCount)
	}
	if edges[0].LastSeen != "t3" {
		t.Errorf("LastSeen: got %q, want t3", edges[0].LastSeen)
	}
}

func TestPersistThreatGraph_Empty(t *testing.T) {
	dir := t.TempDir()
	if err := PersistThreatGraph(dir, nil); err != nil {
		t.Errorf("empty records should be no-op, got: %v", err)
	}
}

func TestThreatSummaryFor(t *testing.T) {
	edges := []ThreatEdge{
		{KillerCommander: "Alesha", VictimCommander: "Bob", Kills: 5},
		{KillerCommander: "Alesha", VictimCommander: "Cora", Kills: 2},
		{KillerCommander: "Alesha", VictimCommander: "Dan", Kills: 3},
		{KillerCommander: "Bob", VictimCommander: "Alesha", Kills: 4},
		{KillerCommander: "Cora", VictimCommander: "Alesha", Kills: 1},
	}

	s := ThreatSummaryFor(edges, "Alesha", 2)
	if s.Commander != "Alesha" {
		t.Errorf("commander: got %q", s.Commander)
	}
	if s.TotalKills != 10 {
		t.Errorf("total kills: got %d, want 10", s.TotalKills)
	}
	if s.TotalDeaths != 5 {
		t.Errorf("total deaths: got %d, want 5", s.TotalDeaths)
	}
	if len(s.TopKills) != 2 {
		t.Fatalf("top kills should be capped at 2, got %d", len(s.TopKills))
	}
	if s.TopKills[0].VictimCommander != "Bob" {
		t.Errorf("top kill should be Bob (5 kills), got %s", s.TopKills[0].VictimCommander)
	}
	if s.TopKills[1].VictimCommander != "Dan" {
		t.Errorf("second top kill should be Dan (3 kills), got %s", s.TopKills[1].VictimCommander)
	}
	if len(s.TopDeaths) != 2 {
		t.Fatalf("top deaths got %d", len(s.TopDeaths))
	}
	if s.TopDeaths[0].KillerCommander != "Bob" {
		t.Errorf("top death should be Bob (4), got %s", s.TopDeaths[0].KillerCommander)
	}
}

func TestThreatSummaryFor_NoMatches(t *testing.T) {
	edges := []ThreatEdge{{KillerCommander: "X", VictimCommander: "Y", Kills: 3}}
	s := ThreatSummaryFor(edges, "Z", 5)
	if s.TotalKills != 0 || s.TotalDeaths != 0 {
		t.Errorf("unknown commander should have zero totals")
	}
	if len(s.TopKills) != 0 || len(s.TopDeaths) != 0 {
		t.Errorf("unknown commander should have empty top lists")
	}
}

func TestLoadThreatGraph_MissingFile(t *testing.T) {
	dir := t.TempDir()
	got, err := LoadThreatGraph(dir)
	if err != nil {
		t.Errorf("missing file should not error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}
