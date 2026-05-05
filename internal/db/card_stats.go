package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// card_stats — card-name-keyed (cross-commander) win/loss aggregate.
//
// Companion to the existing card_win_stats, which is keyed
// (card_name, commander) and counts only cards on the battlefield at
// game end. card_stats is keyed by card_name alone and tracks deck-list
// participation: every card in a winning deck gets +1 game and +1 win,
// every card in a losing deck gets +1 game and +1 loss — regardless of
// whether the card ever made it into play. This is what powers the
// per-card analytics endpoint at /api/cards/{name}/stats.
//
// avg_turn_played is reserved on the schema but currently stays at 0 —
// per-card "turn it was first played" requires runtime instrumentation
// we don't yet collect. It's exposed in the response for forward
// compatibility; handlers should treat 0 as "not measured."

const cardStatsSchema = `
CREATE TABLE IF NOT EXISTS card_stats (
    card_name       TEXT PRIMARY KEY,
    wins            INTEGER NOT NULL DEFAULT 0,
    losses          INTEGER NOT NULL DEFAULT 0,
    games           INTEGER NOT NULL DEFAULT 0,
    avg_turn_played REAL    NOT NULL DEFAULT 0,
    updated_at      INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_card_stats_winrate ON card_stats(wins, games);
`

// EnsureCardStatsSchema creates the card_stats table idempotently.
// Call once at startup before the showmatch persistGame path runs.
func EnsureCardStatsSchema(ctx context.Context, sqlDB *sql.DB) error {
	if sqlDB == nil {
		return errors.New("db: nil database for card_stats schema")
	}
	_, err := sqlDB.ExecContext(ctx, cardStatsSchema)
	return err
}

// CardStat is the in-memory view of one row.
type CardStat struct {
	CardName      string
	Wins          int
	Losses        int
	Games         int
	AvgTurnPlayed float64
}

// CardStatDelta is what callers feed into BatchUpsertCardStats — the
// per-game increment, not the cumulative count.
type CardStatDelta struct {
	CardName string
	Win      int // 1 if the card was in the winning deck this game, else 0
	Loss     int // 1 if the card was in a losing deck this game, else 0
}

// BatchUpsertCardStats applies a list of per-card deltas in one
// transaction. Each row contributes +1 to games, plus its win/loss
// increment. Idempotent at the row level (UPSERT), but the caller is
// responsible for not double-feeding the same game.
func BatchUpsertCardStats(ctx context.Context, sqlDB *sql.DB, deltas []CardStatDelta) error {
	if len(deltas) == 0 {
		return nil
	}
	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO card_stats (card_name, wins, losses, games, avg_turn_played, updated_at)
		 VALUES (?, ?, ?, 1, 0, ?)
		 ON CONFLICT(card_name) DO UPDATE SET
		   wins = wins + excluded.wins,
		   losses = losses + excluded.losses,
		   games = games + 1,
		   updated_at = excluded.updated_at`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	now := time.Now().Unix()
	for _, d := range deltas {
		if d.CardName == "" {
			continue
		}
		if _, err := stmt.ExecContext(ctx, d.CardName, d.Win, d.Loss, now); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// LoadCardStat returns the aggregate row for a single card name. Returns
// a zero-value CardStat with Games=0 when the card has no recorded plays
// (rather than sql.ErrNoRows) so callers can render "no data yet"
// without branching on the error.
func LoadCardStat(ctx context.Context, sqlDB *sql.DB, cardName string) (CardStat, error) {
	var s CardStat
	s.CardName = cardName
	err := sqlDB.QueryRowContext(ctx,
		`SELECT card_name, wins, losses, games, avg_turn_played
		 FROM card_stats WHERE card_name = ?`, cardName,
	).Scan(&s.CardName, &s.Wins, &s.Losses, &s.Games, &s.AvgTurnPlayed)
	if errors.Is(err, sql.ErrNoRows) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	return s, nil
}
