package db

import (
	"context"
	"testing"
)

func newTestDB(t *testing.T) (ctx context.Context, sqlDB any, close func()) {
	t.Helper()
	dbPath := t.TempDir() + "/card_stats_test.db"
	d, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := EnsureCardStatsSchema(context.Background(), d); err != nil {
		t.Fatalf("schema: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return context.Background(), d, func() { d.Close() }
}

func TestCardStats_UpsertAccumulates(t *testing.T) {
	dbPath := t.TempDir() + "/cs.db"
	d, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()
	ctx := context.Background()
	if err := EnsureCardStatsSchema(ctx, d); err != nil {
		t.Fatalf("schema: %v", err)
	}

	// Game 1: Sol Ring in winner deck, Brainstorm in loser deck.
	if err := BatchUpsertCardStats(ctx, d, []CardStatDelta{
		{CardName: "Sol Ring", Win: 1},
		{CardName: "Brainstorm", Loss: 1},
	}); err != nil {
		t.Fatalf("upsert game 1: %v", err)
	}
	// Game 2: both cards in losing decks.
	if err := BatchUpsertCardStats(ctx, d, []CardStatDelta{
		{CardName: "Sol Ring", Loss: 1},
		{CardName: "Brainstorm", Loss: 1},
	}); err != nil {
		t.Fatalf("upsert game 2: %v", err)
	}
	// Game 3: Sol Ring wins again.
	if err := BatchUpsertCardStats(ctx, d, []CardStatDelta{
		{CardName: "Sol Ring", Win: 1},
	}); err != nil {
		t.Fatalf("upsert game 3: %v", err)
	}

	sol, err := LoadCardStat(ctx, d, "Sol Ring")
	if err != nil {
		t.Fatalf("load sol: %v", err)
	}
	if sol.Games != 3 || sol.Wins != 2 || sol.Losses != 1 {
		t.Errorf("Sol Ring: got %+v, want games=3 wins=2 losses=1", sol)
	}

	brain, err := LoadCardStat(ctx, d, "Brainstorm")
	if err != nil {
		t.Fatalf("load brain: %v", err)
	}
	if brain.Games != 2 || brain.Wins != 0 || brain.Losses != 2 {
		t.Errorf("Brainstorm: got %+v, want games=2 wins=0 losses=2", brain)
	}
}

func TestCardStats_NoRowsReturnsZeroStat(t *testing.T) {
	dbPath := t.TempDir() + "/cs.db"
	d, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()
	ctx := context.Background()
	if err := EnsureCardStatsSchema(ctx, d); err != nil {
		t.Fatalf("schema: %v", err)
	}
	stat, err := LoadCardStat(ctx, d, "Nonexistent Card")
	if err != nil {
		t.Errorf("expected nil error for missing card, got %v", err)
	}
	if stat.Games != 0 || stat.Wins != 0 || stat.Losses != 0 {
		t.Errorf("missing card should be zero-valued: %+v", stat)
	}
	if stat.CardName != "Nonexistent Card" {
		t.Errorf("card name: got %q, want %q", stat.CardName, "Nonexistent Card")
	}
}

func TestCardStats_EmptyBatchIsNoOp(t *testing.T) {
	dbPath := t.TempDir() + "/cs.db"
	d, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()
	ctx := context.Background()
	if err := EnsureCardStatsSchema(ctx, d); err != nil {
		t.Fatalf("schema: %v", err)
	}
	if err := BatchUpsertCardStats(ctx, d, nil); err != nil {
		t.Errorf("empty batch should not error: %v", err)
	}
	if err := BatchUpsertCardStats(ctx, d, []CardStatDelta{}); err != nil {
		t.Errorf("empty slice should not error: %v", err)
	}
}

func TestCardStats_BlankCardNameSkipped(t *testing.T) {
	dbPath := t.TempDir() + "/cs.db"
	d, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()
	ctx := context.Background()
	if err := EnsureCardStatsSchema(ctx, d); err != nil {
		t.Fatalf("schema: %v", err)
	}
	// Mixing a real card with a blank name should still record the real one.
	if err := BatchUpsertCardStats(ctx, d, []CardStatDelta{
		{CardName: "", Win: 1},
		{CardName: "Counterspell", Win: 1},
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	cs, _ := LoadCardStat(ctx, d, "Counterspell")
	if cs.Games != 1 || cs.Wins != 1 {
		t.Errorf("Counterspell: got %+v, want games=1 wins=1", cs)
	}
	blank, _ := LoadCardStat(ctx, d, "")
	if blank.Games != 0 {
		t.Errorf("blank-name card should not have been inserted: %+v", blank)
	}
}

// silence unused warning for the test helper above (kept as scaffolding
// for future tests that prefer the closure form).
var _ = newTestDB
