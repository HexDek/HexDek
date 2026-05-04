package hat

import "testing"

func TestComposeNarrative_Basic(t *testing.T) {
	pivot := CausalPivot{Turn: 10, WinnerSeat: 0, DeltaScore: 0.4}
	events := []GameEvent{
		{Turn: 3, Seat: 0, Kind: "damage", Source: "Lightning Bolt", Amount: 3},
		{Turn: 10, Seat: 0, Kind: "boardwipe", Source: "Wrath of God"},
		{Turn: 14, Seat: 1, Kind: "player_lost"},
		{Turn: 16, Seat: 2, Kind: "player_lost"},
		{Turn: 18, Seat: 3, Kind: "player_lost"},
	}
	names := []string{"Atraxa", "Korvold", "Muldrotha", "Prossh"}

	n := ComposeNarrative(pivot, events, names, 0, 18)

	if n.Winner != "Atraxa" {
		t.Errorf("winner should be Atraxa, got %s", n.Winner)
	}
	if n.TotalTurns != 18 {
		t.Errorf("total turns should be 18, got %d", n.TotalTurns)
	}
	if len(n.Acts) != 3 {
		t.Fatalf("should have 3 acts, got %d", len(n.Acts))
	}
	if n.Acts[0].Name != "Setup" {
		t.Errorf("act 1 should be Setup, got %s", n.Acts[0].Name)
	}
	if n.Acts[1].Name != "Conflict" {
		t.Errorf("act 2 should be Conflict, got %s", n.Acts[1].Name)
	}
	if n.Acts[2].Name != "Resolution" {
		t.Errorf("act 3 should be Resolution, got %s", n.Acts[2].Name)
	}
	if n.Synopsis == "" {
		t.Error("synopsis should not be empty")
	}
}

func TestComposeNarrative_Highlights(t *testing.T) {
	pivot := CausalPivot{Turn: 8, WinnerSeat: 1, DeltaScore: 0.3}
	events := []GameEvent{
		{Turn: 2, Seat: 0, Kind: "damage", Source: "Goblin Guide", Amount: 2},
		{Turn: 5, Seat: 1, Kind: "combo_assembled", Source: "Thassa's Oracle"},
		{Turn: 8, Seat: 0, Kind: "player_lost"},
	}
	names := []string{"Krenko", "Thassa"}

	n := ComposeNarrative(pivot, events, names, 1, 10)

	kinds := make(map[string]bool)
	for _, h := range n.Highlights {
		kinds[h.Kind] = true
	}
	if !kinds["first_blood"] {
		t.Error("should have first_blood highlight")
	}
	if !kinds["pivot"] {
		t.Error("should have pivot highlight")
	}
	if !kinds["elimination"] {
		t.Error("should have elimination highlight")
	}
}

func TestComposeNarrative_ShortGame(t *testing.T) {
	pivot := CausalPivot{Turn: 2, WinnerSeat: 0, DeltaScore: 0.8}
	n := ComposeNarrative(pivot, nil, []string{"Flash Hulk", "Casual"}, 0, 3)

	if n.TotalTurns != 3 {
		t.Errorf("expected 3 turns, got %d", n.TotalTurns)
	}
	// Should not panic on short games.
	if n.Acts[0].StartTurn != 1 {
		t.Errorf("act 1 should start at turn 1, got %d", n.Acts[0].StartTurn)
	}
}

func TestSeatName_OutOfBounds(t *testing.T) {
	name := seatName([]string{"A", "B"}, 5)
	if name != "Seat 5" {
		t.Errorf("out-of-bounds should return 'Seat 5', got %s", name)
	}
}
