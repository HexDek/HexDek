// Package db — gauntlet_runs persistence.
//
// One row per completed gauntlet captures the rating trajectory for the
// deck under test. Drives the deck-page ELO history chart so players see
// the calibration arc, not just the current rating snapshot.
package db

import (
	"context"
	"database/sql"
	"time"
)

// GauntletRunRecord is one persistent gauntlet-completion entry.
// Fields mirror the schema in schema.sql (gauntlet_runs).
type GauntletRunRecord struct {
	RunID      int64     `json:"run_id"`
	DeckKey    string    `json:"deck_key"`
	Commander  string    `json:"commander"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Games      int       `json:"games"`
	Wins       int       `json:"wins"`
	Losses     int       `json:"losses"`
	WinRate    float64   `json:"win_rate"`
	ELOStart   float64   `json:"elo_start"`
	ELOEnd     float64   `json:"elo_end"`
	ELODelta   float64   `json:"elo_delta"`
	AvgTurns   float64   `json:"avg_turns"`
	// Per-position counts for the deck under test (seat 0).
	// Index meaning: [1st, 2nd, 3rd, 4th].
	Place1st int `json:"place_1st"`
	Place2nd int `json:"place_2nd"`
	Place3rd int `json:"place_3rd"`
	Place4th int `json:"place_4th"`
}

// InsertGauntletRun writes a completed-gauntlet snapshot.
// Idempotent only by autoincrement run_id — caller is responsible for
// not double-inserting on retry. Safe to call from the gauntlet
// finalization path.
func InsertGauntletRun(ctx context.Context, sqlDB *sql.DB, rec GauntletRunRecord) error {
	if sqlDB == nil {
		return nil
	}
	_, err := sqlDB.ExecContext(ctx, `
		INSERT INTO gauntlet_runs (
			deck_key, commander, started_at, finished_at,
			games, wins, losses, win_rate,
			elo_start, elo_end, elo_delta, avg_turns,
			place_1st, place_2nd, place_3rd, place_4th
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.DeckKey, rec.Commander,
		rec.StartedAt.Unix(), rec.FinishedAt.Unix(),
		rec.Games, rec.Wins, rec.Losses, rec.WinRate,
		rec.ELOStart, rec.ELOEnd, rec.ELODelta, rec.AvgTurns,
		rec.Place1st, rec.Place2nd, rec.Place3rd, rec.Place4th,
	)
	return err
}

// LoadGauntletRuns returns the most recent N runs for a deck, newest
// first. Limit ≤ 0 returns all rows for the deck. Caller renders the
// rating trajectory by reversing this slice (chronological order).
func LoadGauntletRuns(ctx context.Context, sqlDB *sql.DB, deckKey string, limit int) ([]GauntletRunRecord, error) {
	if sqlDB == nil || deckKey == "" {
		return nil, nil
	}
	q := `SELECT run_id, deck_key, commander, started_at, finished_at,
	             games, wins, losses, win_rate,
	             elo_start, elo_end, elo_delta, avg_turns,
	             place_1st, place_2nd, place_3rd, place_4th
	      FROM gauntlet_runs WHERE deck_key = ?
	      ORDER BY finished_at DESC`
	args := []interface{}{deckKey}
	if limit > 0 {
		q += " LIMIT ?"
		args = append(args, limit)
	}
	rows, err := sqlDB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GauntletRunRecord
	for rows.Next() {
		var r GauntletRunRecord
		var started, finished int64
		if err := rows.Scan(
			&r.RunID, &r.DeckKey, &r.Commander, &started, &finished,
			&r.Games, &r.Wins, &r.Losses, &r.WinRate,
			&r.ELOStart, &r.ELOEnd, &r.ELODelta, &r.AvgTurns,
			&r.Place1st, &r.Place2nd, &r.Place3rd, &r.Place4th,
		); err != nil {
			return nil, err
		}
		r.StartedAt = time.Unix(started, 0)
		r.FinishedAt = time.Unix(finished, 0)
		out = append(out, r)
	}
	return out, rows.Err()
}
