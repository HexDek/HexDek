package db

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func newAnticheatTestDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := Open(t.TempDir() + "/anticheat.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

// seedShowmatchGame inserts one game + N seat rows so
// LoadGameForVerification + QuarantineRecentGames have something to
// chew on. Returns the game_id.
func seedShowmatchGame(t *testing.T, d *sql.DB, deckKeys []string, winner, turns int, finishedAt int64, rngSeed int64) int64 {
	t.Helper()
	ctx := context.Background()
	res, err := d.ExecContext(ctx,
		`INSERT INTO showmatch_game (started_at, finished_at, turns, winner, winner_name, end_reason, rng_seed)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		finishedAt-3600, finishedAt, turns, winner, "TEST", "last_seat_standing", rngSeed)
	if err != nil {
		t.Fatalf("insert showmatch_game: %v", err)
	}
	gid, _ := res.LastInsertId()
	for i, dk := range deckKeys {
		_, err := d.ExecContext(ctx,
			`INSERT INTO showmatch_game_seat (game_id, seat, commander, deck_key, life, hand_size, library_size, gy_size, bf_size, lost)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			gid, i, "Test Commander "+dk, dk, 0, 0, 0, 0, 0, 0)
		if err != nil {
			t.Fatalf("insert seat: %v", err)
		}
	}
	return gid
}

func TestEnqueueAndClaimVerification(t *testing.T) {
	d := newAnticheatTestDB(t)
	ctx := context.Background()

	id1, err := EnqueueVerification(ctx, d, VerificationEnqueueParams{
		GameID:        100,
		DeckKey:       "alice/aggro",
		RNGSeed:       42,
		NSeats:        4,
		DeckKeys:      []string{"alice/aggro", "bob/control", "carol/combo", "dave/midrange"},
		ClaimedWinner: 2,
		ClaimedTurns:  14,
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if id1 == 0 {
		t.Fatal("expected non-zero queue_id")
	}

	row, err := ClaimNextVerification(ctx, d)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if row == nil {
		t.Fatal("ClaimNextVerification returned nil for non-empty queue")
	}
	if row.Status != "running" {
		t.Errorf("expected status=running, got %q", row.Status)
	}
	if row.GameID != 100 || row.DeckKey != "alice/aggro" {
		t.Errorf("wrong row: %+v", row)
	}
	if !row.StartedAt.Valid || row.StartedAt.Int64 == 0 {
		t.Error("started_at not stamped on claim")
	}
	if len(row.DeckKeys) != 4 || row.DeckKeys[2] != "carol/combo" {
		t.Errorf("deck keys not roundtripped: %v", row.DeckKeys)
	}

	// Subsequent claim should return nil — only one row, already running.
	row2, err := ClaimNextVerification(ctx, d)
	if err != nil {
		t.Fatalf("second claim: %v", err)
	}
	if row2 != nil {
		t.Errorf("expected nil on second claim, got %+v", row2)
	}
}

func TestFinishVerification_TerminalStatus(t *testing.T) {
	d := newAnticheatTestDB(t)
	ctx := context.Background()

	id, _ := EnqueueVerification(ctx, d, VerificationEnqueueParams{
		GameID: 1, DeckKey: "x/y", RNGSeed: 1, NSeats: 4,
		DeckKeys: []string{"x/y", "a/b", "c/d", "e/f"}, ClaimedWinner: 0, ClaimedTurns: 5,
	})
	if _, err := ClaimNextVerification(ctx, d); err != nil {
		t.Fatalf("claim: %v", err)
	}
	if err := FinishVerification(ctx, d, id, "passed", 0, 5, "ok"); err != nil {
		t.Fatalf("finish: %v", err)
	}
	rows, err := ListVerifications(ctx, d, "passed", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 || rows[0].QueueID != id {
		t.Fatalf("expected one passed row, got %+v", rows)
	}

	// Bad status rejected.
	if err := FinishVerification(ctx, d, id, "bogus", 0, 0, ""); err == nil {
		t.Error("expected error on invalid terminal status")
	}
}

func TestVerificationStats_GroupsByStatus(t *testing.T) {
	d := newAnticheatTestDB(t)
	ctx := context.Background()

	// Insert 3 rows, transition each to its target status, THEN add 2
	// that stay pending. ClaimNextVerification always picks the oldest
	// pending row, so we drain transitions first to avoid grabbing a
	// row we wanted to leave pending.
	mkAndFinish := func(status string) {
		id, _ := EnqueueVerification(ctx, d, VerificationEnqueueParams{
			GameID: 1, DeckKey: "x/y", RNGSeed: 1, NSeats: 4,
			DeckKeys: []string{"x/y", "a/b", "c/d", "e/f"},
		})
		ClaimNextVerification(ctx, d)
		FinishVerification(ctx, d, id, status, 0, 0, "")
	}
	mkAndFinish("passed")
	mkAndFinish("failed")
	mkAndFinish("error")
	EnqueueVerification(ctx, d, VerificationEnqueueParams{
		GameID: 1, DeckKey: "x/y", RNGSeed: 1, NSeats: 4,
		DeckKeys: []string{"x/y", "a/b", "c/d", "e/f"},
	})
	EnqueueVerification(ctx, d, VerificationEnqueueParams{
		GameID: 1, DeckKey: "x/y", RNGSeed: 1, NSeats: 4,
		DeckKeys: []string{"x/y", "a/b", "c/d", "e/f"},
	})

	s, err := GetVerificationStats(ctx, d)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if s.Pending != 2 || s.Passed != 1 || s.Failed != 1 || s.Error != 1 {
		t.Errorf("wrong stats: %+v", s)
	}
}

func TestSanctions_InsertAndCount(t *testing.T) {
	d := newAnticheatTestDB(t)
	ctx := context.Background()

	if n, _ := CountOffenses(ctx, d, "alice/aggro"); n != 0 {
		t.Errorf("expected 0 offenses on fresh deck, got %d", n)
	}

	sid, err := InsertSanction(ctx, d, SanctionInsertParams{
		DeckKey:  "alice/aggro",
		Owner:    "alice",
		Severity: SeverityWarning,
		Reason:   "first offense",
	})
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if sid == 0 {
		t.Fatal("expected non-zero sanction_id")
	}

	if n, _ := CountOffenses(ctx, d, "alice/aggro"); n != 1 {
		t.Errorf("expected 1 offense, got %d", n)
	}

	// Bad severity.
	if _, err := InsertSanction(ctx, d, SanctionInsertParams{
		DeckKey: "x/y", Severity: "death-penalty",
	}); err == nil {
		t.Error("expected error on invalid severity")
	}
}

func TestActiveBan_FiltersWarningAndExpired(t *testing.T) {
	d := newAnticheatTestDB(t)
	ctx := context.Background()

	now := time.Now().Unix()
	// Warning — should NOT count as a ban.
	InsertSanction(ctx, d, SanctionInsertParams{
		DeckKey: "alice/aggro", Severity: SeverityWarning, IssuedAt: now,
	})
	if ban, _ := ActiveBan(ctx, d, "alice/aggro"); ban != nil {
		t.Errorf("warning leaked as active ban: %+v", ban)
	}
	// Expired temp ban — should NOT count.
	InsertSanction(ctx, d, SanctionInsertParams{
		DeckKey: "bob/control", Severity: SeverityTempBan, IssuedAt: now - 86400, ExpiresAt: now - 1,
	})
	if ban, _ := ActiveBan(ctx, d, "bob/control"); ban != nil {
		t.Errorf("expired temp ban still active: %+v", ban)
	}
	// Active temp ban — SHOULD count.
	InsertSanction(ctx, d, SanctionInsertParams{
		DeckKey: "carol/combo", Severity: SeverityTempBan, IssuedAt: now, ExpiresAt: now + 86400,
	})
	ban, err := ActiveBan(ctx, d, "carol/combo")
	if err != nil {
		t.Fatalf("active ban: %v", err)
	}
	if ban == nil || ban.Severity != SeverityTempBan {
		t.Errorf("expected active temp ban, got %+v", ban)
	}
	// Permanent ban — always active.
	InsertSanction(ctx, d, SanctionInsertParams{
		DeckKey: "mallory/cheater", Severity: SeverityPermanentBan, IssuedAt: now,
	})
	ban, _ = ActiveBan(ctx, d, "mallory/cheater")
	if ban == nil || ban.Severity != SeverityPermanentBan {
		t.Errorf("expected permanent ban, got %+v", ban)
	}
}

func TestQuarantineRecentGames_FlagsAffectedRows(t *testing.T) {
	d := newAnticheatTestDB(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Recent game with alice in seat 0.
	gid := seedShowmatchGame(t, d, []string{"alice/aggro", "bob/control", "carol/combo", "dave/midrange"}, 0, 14, now-3600, 100)
	// Old game (8 days ago) with alice — should NOT be touched.
	oldGid := seedShowmatchGame(t, d, []string{"alice/aggro", "x/y", "z/w", "p/q"}, 1, 22, now-int64((8*24*time.Hour).Seconds()), 200)
	// Game without alice — should NOT be touched.
	otherGid := seedShowmatchGame(t, d, []string{"u/v", "x/y", "z/w", "p/q"}, 2, 18, now-3600, 300)

	since := now - int64((7 * 24 * time.Hour).Seconds())
	n, err := QuarantineRecentGames(ctx, d, "alice/aggro", since)
	if err != nil {
		t.Fatalf("quarantine: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 game quarantined, got %d", n)
	}

	verified := func(g int64) int {
		var v int
		d.QueryRowContext(ctx, `SELECT verified FROM showmatch_game WHERE game_id=?`, g).Scan(&v)
		return v
	}
	if verified(gid) != -1 {
		t.Errorf("recent alice game not flagged: verified=%d", verified(gid))
	}
	if verified(oldGid) != 0 {
		t.Errorf("old alice game should not be flagged: verified=%d", verified(oldGid))
	}
	if verified(otherGid) != 0 {
		t.Errorf("non-alice game should not be flagged: verified=%d", verified(otherGid))
	}
}

func TestLoadGameForVerification_ReturnsSeatOrder(t *testing.T) {
	d := newAnticheatTestDB(t)
	ctx := context.Background()
	gid := seedShowmatchGame(t, d, []string{"alice/aggro", "bob/control", "carol/combo", "dave/midrange"}, 1, 18, time.Now().Unix(), 9999)
	g, err := LoadGameForVerification(ctx, d, gid)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if g.NSeats != 4 || g.Winner != 1 || g.Turns != 18 || g.RNGSeed != 9999 {
		t.Errorf("wrong header: %+v", g)
	}
	if len(g.DeckKeys) != 4 || g.DeckKeys[2] != "carol/combo" {
		t.Errorf("seat order broken: %v", g.DeckKeys)
	}
}
