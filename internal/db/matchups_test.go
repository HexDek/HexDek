package db

import (
	"context"
	"database/sql"
	"testing"
)

func mustExec(t *testing.T, d *sql.DB, q string, args ...any) {
	t.Helper()
	if _, err := d.Exec(q, args...); err != nil {
		t.Fatalf("exec %q: %v", q, err)
	}
}

// TestLoadDeckMatchups seeds a 2-player game, a 4-player pod, and a
// draw, then verifies the aggregated matchup row counts and win-rate
// math.
func TestLoadDeckMatchups(t *testing.T) {
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()
	ctx := context.Background()

	// Game 1: me (seat 0) vs atraxa (seat 1). I win.
	mustExec(t, d, `INSERT INTO showmatch_game(game_id, started_at, finished_at, turns, winner, winner_name, end_reason)
	                VALUES (1, 0, 0, 10, 0, 'me', 'concession')`)
	mustExec(t, d, `INSERT INTO showmatch_game_seat(game_id, seat, commander, deck_key, life, lost)
	                VALUES (1, 0, 'Brudiclad, Telchor Engineer', 'josh/brud', 40, 0),
	                       (1, 1, 'Atraxa, Praetors'' Voice',     'jane/atra', 0,  1)`)

	// Game 2: me vs atraxa. atraxa wins.
	mustExec(t, d, `INSERT INTO showmatch_game(game_id, started_at, finished_at, turns, winner, winner_name, end_reason)
	                VALUES (2, 0, 0, 12, 1, 'atraxa', 'damage')`)
	mustExec(t, d, `INSERT INTO showmatch_game_seat(game_id, seat, commander, deck_key, life, lost)
	                VALUES (2, 0, 'Brudiclad, Telchor Engineer', 'josh/brud', 0,  1),
	                       (2, 1, 'Atraxa, Praetors'' Voice',     'jane/atra', 12, 0)`)

	// Game 3: 4-player pod, me + atraxa + krenko + sisay. krenko wins.
	mustExec(t, d, `INSERT INTO showmatch_game(game_id, started_at, finished_at, turns, winner, winner_name, end_reason)
	                VALUES (3, 0, 0, 8, 2, 'krenko', 'damage')`)
	mustExec(t, d, `INSERT INTO showmatch_game_seat(game_id, seat, commander, deck_key, life, lost)
	                VALUES (3, 0, 'Brudiclad, Telchor Engineer', 'josh/brud', 0,  1),
	                       (3, 1, 'Atraxa, Praetors'' Voice',     'jane/atra', 0,  1),
	                       (3, 2, 'Krenko, Mob Boss',             'tom/kren',  18, 0),
	                       (3, 3, 'Sisay, Weatherlight Captain',  'kim/sisay', 0,  1)`)

	// Game 4: a draw vs sisay. winner=-1.
	mustExec(t, d, `INSERT INTO showmatch_game(game_id, started_at, finished_at, turns, winner, winner_name, end_reason)
	                VALUES (4, 0, 0, 50, -1, 'DRAW', 'turn_limit')`)
	mustExec(t, d, `INSERT INTO showmatch_game_seat(game_id, seat, commander, deck_key, life, lost)
	                VALUES (4, 0, 'Brudiclad, Telchor Engineer', 'josh/brud', 1, 0),
	                       (4, 1, 'Sisay, Weatherlight Captain',  'kim/sisay', 1, 0)`)

	rows, err := LoadDeckMatchups(ctx, d, "josh/brud", 0)
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	by := map[string]MatchupRow{}
	for _, r := range rows {
		by[r.OpponentCommander] = r
	}

	atra, ok := by["Atraxa, Praetors' Voice"]
	if !ok {
		t.Fatalf("expected Atraxa row")
	}
	// 3 games (2 head-to-head + 1 in the 4p pod), 1 win, 2 losses, 0 draws.
	if atra.Games != 3 || atra.Wins != 1 || atra.Losses != 2 || atra.Draws != 0 {
		t.Fatalf("atraxa unexpected: %+v", atra)
	}
	if atra.WinRate != 33.3 {
		t.Fatalf("atraxa win_rate: got %v want 33.3", atra.WinRate)
	}

	kren := by["Krenko, Mob Boss"]
	if kren.Games != 1 || kren.Wins != 0 || kren.Losses != 1 {
		t.Fatalf("krenko unexpected: %+v", kren)
	}

	sisay := by["Sisay, Weatherlight Captain"]
	// 4p pod: krenko won, so vs sisay this is a loss. + draw game = 2 games, 0 wins, 1 loss, 1 draw.
	if sisay.Games != 2 || sisay.Wins != 0 || sisay.Losses != 1 || sisay.Draws != 1 {
		t.Fatalf("sisay unexpected: %+v", sisay)
	}

	// Sort: atraxa (3 games) first.
	if rows[0].OpponentCommander != "Atraxa, Praetors' Voice" {
		t.Fatalf("expected atraxa first, got %q", rows[0].OpponentCommander)
	}

	// Empty deck key returns empty.
	empty, _ := LoadDeckMatchups(ctx, d, "", 0)
	if len(empty) != 0 {
		t.Fatalf("empty deck_key should return no rows, got %d", len(empty))
	}

	// Limit clamps the row count.
	limited, _ := LoadDeckMatchups(ctx, d, "josh/brud", 1)
	if len(limited) != 1 {
		t.Fatalf("limit=1 should return 1 row, got %d", len(limited))
	}
}
