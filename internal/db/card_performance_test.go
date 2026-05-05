package db

import (
	"context"
	"math"
	"testing"
)

func TestCardPerformance_RunningMeansAndWinTracking(t *testing.T) {
	ctx := context.Background()
	db, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	// Game 1: Sol Ring on the winner's board. Played turn 1, 6 turns on bf.
	// Game 2: Sol Ring on a loser's board. Played turn 2, 4 turns on bf.
	// Game 3: Sol Ring drawn but never cast (TurnPlayed=0); winner.
	feed := func(deltas []CardPerformanceDelta) {
		if err := BatchUpsertCardPerformance(ctx, db, deltas); err != nil {
			t.Fatalf("upsert: %v", err)
		}
	}
	feed([]CardPerformanceDelta{{CardName: "Sol Ring", Win: 1, TurnPlayed: 1, BattlefieldTurns: 6}})
	feed([]CardPerformanceDelta{{CardName: "Sol Ring", Win: 0, TurnPlayed: 2, BattlefieldTurns: 4}})
	feed([]CardPerformanceDelta{{CardName: "Sol Ring", Win: 1}})

	got, err := LoadCardPerformance(ctx, db, "Sol Ring")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.GamesIncluded != 3 {
		t.Errorf("games_included = %d, want 3", got.GamesIncluded)
	}
	if got.WinsWhenIncluded != 2 {
		t.Errorf("wins_when_included = %d, want 2", got.WinsWhenIncluded)
	}
	// avg_turn_played = mean(1, 2) = 1.5 (third game had no turn data).
	if math.Abs(got.AvgTurnPlayed-1.5) > 1e-6 {
		t.Errorf("avg_turn_played = %v, want 1.5", got.AvgTurnPlayed)
	}
	// avg_battlefield_time = mean(6, 4) = 5.
	if math.Abs(got.AvgBattlefieldTime-5.0) > 1e-6 {
		t.Errorf("avg_battlefield_time = %v, want 5", got.AvgBattlefieldTime)
	}
}

func TestCardPerformance_LoadMissingReturnsZero(t *testing.T) {
	db, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	got, err := LoadCardPerformance(context.Background(), db, "Nonexistent")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.GamesIncluded != 0 || got.WinsWhenIncluded != 0 {
		t.Errorf("unknown card should produce zeros, got %+v", got)
	}
	if got.CardName != "Nonexistent" {
		t.Errorf("card_name should round-trip, got %q", got.CardName)
	}
}

func TestCardPerformance_EmptyBatchIsNoop(t *testing.T) {
	db, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	if err := BatchUpsertCardPerformance(context.Background(), db, nil); err != nil {
		t.Fatalf("nil batch: %v", err)
	}
	if err := BatchUpsertCardPerformance(context.Background(), db, []CardPerformanceDelta{}); err != nil {
		t.Fatalf("empty batch: %v", err)
	}
}
