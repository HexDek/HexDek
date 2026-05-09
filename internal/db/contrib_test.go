package db

import (
	"context"
	"testing"
)

func TestAwardCredits_FirstChunk(t *testing.T) {
	conn, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()
	ctx := context.Background()

	c, err := AwardCredits(ctx, conn, "alice", 400, 100, 5000, 3.0, 1700000000)
	if err != nil {
		t.Fatalf("award: %v", err)
	}
	if c.CreditsTotal != 400 {
		t.Errorf("credits: %d, want 400", c.CreditsTotal)
	}
	if c.ChunksCompleted != 1 {
		t.Errorf("chunks: %d, want 1", c.ChunksCompleted)
	}
	if c.GamesSimulated != 100 {
		t.Errorf("games: %d, want 100", c.GamesSimulated)
	}
	if c.Frozen {
		t.Error("first chunk should not freeze the account")
	}
	if c.LastZScore != 0 {
		t.Errorf("first chunk z-score should be 0 (no prior data), got %v", c.LastZScore)
	}
}

func TestAwardCredits_AnomalyFreezes(t *testing.T) {
	conn, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()
	ctx := context.Background()

	// Establish a baseline: 8 chunks at ~5000ms with low variance.
	// Welford needs a few samples to produce a stable σ.
	for i := 0; i < 8; i++ {
		elapsed := int64(5000 + i*10) // 5000..5070ms — tight range
		if _, err := AwardCredits(ctx, conn, "bob", 400, 100, elapsed, 3.0, int64(1700000000+i)); err != nil {
			t.Fatalf("baseline %d: %v", i, err)
		}
	}
	// Now hit it with a huge outlier: 200,000ms (far beyond 3σ of the
	// 50ms-spread baseline).
	c, err := AwardCredits(ctx, conn, "bob", 400, 100, 200_000, 3.0, 1700000099)
	if err != nil {
		t.Fatalf("outlier: %v", err)
	}
	if !c.Frozen {
		t.Errorf("outlier chunk should have tripped freeze; z=%v", c.LastZScore)
	}
	if c.LastZScore < 3.0 {
		t.Errorf("outlier z=%v, expected >= 3.0", c.LastZScore)
	}
}

func TestAwardCredits_FrozenAccountDoesNotAccrue(t *testing.T) {
	conn, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()
	ctx := context.Background()

	// First award — not frozen yet, accrues 400.
	if _, err := AwardCredits(ctx, conn, "carol", 400, 100, 5000, 3.0, 1700000000); err != nil {
		t.Fatal(err)
	}
	// Manually freeze.
	if _, err := conn.Exec(`UPDATE contributor_credits SET frozen = 1, frozen_reason = 'manual' WHERE owner = ?`, "carol"); err != nil {
		t.Fatal(err)
	}
	// Second award — should NOT add credits because account is frozen.
	c, err := AwardCredits(ctx, conn, "carol", 400, 100, 5100, 3.0, 1700000001)
	if err != nil {
		t.Fatal(err)
	}
	if c.CreditsTotal != 400 {
		t.Errorf("frozen account accrued credits: %d, want 400 (unchanged)", c.CreditsTotal)
	}
	// But chunks_completed and games_simulated should still tick — we
	// want the audit trail of all submissions even when not paying.
	if c.ChunksCompleted != 2 {
		t.Errorf("chunks_completed: %d, want 2", c.ChunksCompleted)
	}
}

func TestRejectChunk_IncrementsCounter(t *testing.T) {
	conn, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()
	ctx := context.Background()

	if err := RejectChunk(ctx, conn, "dave", 1700000000); err != nil {
		t.Fatal(err)
	}
	if err := RejectChunk(ctx, conn, "dave", 1700000001); err != nil {
		t.Fatal(err)
	}
	c, err := GetContributorCredits(ctx, conn, "dave")
	if err != nil {
		t.Fatal(err)
	}
	if c.ChunksRejected != 2 {
		t.Errorf("rejected: %d, want 2", c.ChunksRejected)
	}
	if c.CreditsTotal != 0 {
		t.Errorf("rejected chunks should not credit, got %d", c.CreditsTotal)
	}
}

func TestEnsureContributorRow_Idempotent(t *testing.T) {
	conn, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()
	ctx := context.Background()

	if err := EnsureContributorRow(ctx, conn, "eve", 1700000000); err != nil {
		t.Fatal(err)
	}
	// Calling twice should not error and should update last_active_at.
	if err := EnsureContributorRow(ctx, conn, "eve", 1700000050); err != nil {
		t.Fatal(err)
	}
	c, err := GetContributorCredits(ctx, conn, "eve")
	if err != nil {
		t.Fatal(err)
	}
	if c.FirstSeenAt != 1700000000 {
		t.Errorf("first_seen_at clobbered: %d", c.FirstSeenAt)
	}
	if c.LastActiveAt != 1700000050 {
		t.Errorf("last_active_at not refreshed: %d", c.LastActiveAt)
	}
}

func TestRecordAndFinalizeChunk(t *testing.T) {
	conn, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()
	ctx := context.Background()

	if err := RecordChunkAssignment(ctx, conn, "ck-1", "frank", 100, 4, 1700000000); err != nil {
		t.Fatal(err)
	}
	if err := FinalizeChunkRow(ctx, conn, "ck-1", 1, true, true, 4500, "abcd", 400, "", 1700000020); err != nil {
		t.Fatal(err)
	}
	row := conn.QueryRow(`SELECT accepted, spot_checked, spot_check_passed, elapsed_ms, outcome_hash, credits_awarded
                                  FROM contrib_chunk WHERE chunk_id = ?`, "ck-1")
	var accepted, sc, sp int
	var elapsed int64
	var hash string
	var credits int64
	if err := row.Scan(&accepted, &sc, &sp, &elapsed, &hash, &credits); err != nil {
		t.Fatal(err)
	}
	if accepted != 1 || sc != 1 || sp != 1 || elapsed != 4500 || hash != "abcd" || credits != 400 {
		t.Errorf("finalize round-trip failed: a=%d sc=%d sp=%d el=%d h=%q c=%d", accepted, sc, sp, elapsed, hash, credits)
	}
}
