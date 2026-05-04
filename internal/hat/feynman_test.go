package hat

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

func newFeynmanGame(t *testing.T, nSeats int) *gameengine.GameState {
	t.Helper()
	gs := newTestGame(t, nSeats)
	for i := range gs.Seats {
		gs.Seats[i].Life = 40
		gs.Seats[i].StartingLife = 40
		// Give each seat a 100-card library to satisfy zone accounting.
		for j := 0; j < 100; j++ {
			gs.Seats[i].Library = append(gs.Seats[i].Library,
				&gameengine.Card{Name: "Card", Owner: i})
		}
	}
	return gs
}

func TestFeynman_CleanGame(t *testing.T) {
	gs := newFeynmanGame(t, 4)
	// Normal end: 3 seats lost.
	gs.Seats[1].Lost = true
	gs.Seats[2].Lost = true
	gs.Seats[3].Lost = true
	gs.Turn = 15

	result := CheckGame(gs)
	if !result.Clean() {
		t.Errorf("clean game should have no violations, got %d: %v",
			len(result.Violations), result.Violations)
	}
	if result.Checked < 5 {
		t.Errorf("should check at least 5 invariants, checked %d", result.Checked)
	}
}

func TestFeynman_LifeSBA(t *testing.T) {
	gs := newFeynmanGame(t, 2)
	gs.Seats[0].Life = -5
	gs.Seats[0].Lost = false // BUG: should be lost
	gs.Seats[1].Lost = false
	gs.Turn = 10

	result := CheckGame(gs)
	found := false
	for _, v := range result.Violations {
		if v.Rule == "704.5a" {
			found = true
		}
	}
	if !found {
		t.Error("expected 704.5a violation for seat 0 with negative life")
	}
}

func TestFeynman_PoisonSBA(t *testing.T) {
	gs := newFeynmanGame(t, 2)
	gs.Seats[0].PoisonCounters = 10
	gs.Seats[0].Lost = false // BUG
	gs.Seats[1].Lost = false
	gs.Turn = 8

	result := CheckGame(gs)
	found := false
	for _, v := range result.Violations {
		if v.Rule == "704.5c" {
			found = true
		}
	}
	if !found {
		t.Error("expected 704.5c violation for 10 poison")
	}
}

func TestFeynman_TurnBounds(t *testing.T) {
	gs := newFeynmanGame(t, 2)
	gs.Turn = 250
	gs.Seats[1].Lost = true

	result := CheckGame(gs)
	found := false
	for _, v := range result.Violations {
		if v.Rule == "turn_bound" {
			found = true
		}
	}
	if !found {
		t.Error("expected turn_bound violation for 250-turn game")
	}
}

func TestFeynman_NoWinner(t *testing.T) {
	gs := newFeynmanGame(t, 4)
	// Nobody lost — game ended prematurely.
	gs.Turn = 10

	result := CheckGame(gs)
	found := false
	for _, v := range result.Violations {
		if v.Rule == "game_end" {
			found = true
		}
	}
	if !found {
		t.Error("expected game_end violation when nobody lost")
	}
}

func TestFeynman_FormatViolations(t *testing.T) {
	violations := []OracleViolation{
		{Rule: "704.5a", Description: "test", Seat: 0, Severity: "critical"},
	}
	s := FormatViolations(violations)
	if s == "" {
		t.Error("format should produce non-empty string")
	}
	clean := FormatViolations(nil)
	if clean == "" {
		t.Error("clean format should produce non-empty string")
	}
}
