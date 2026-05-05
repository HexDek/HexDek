package db

import (
	"context"
	"database/sql"
	"fmt"
)

// MatchupRow is one (your-deck, opposing-commander) head-to-head entry,
// aggregated across every recorded showmatch_game where both seats were
// present. Win/loss is keyed off showmatch_game.winner: a row counts as
// a win when the winning seat == this deck's seat in that game. Draws
// (winner = -1) are counted in Games but not in Wins or Losses.
type MatchupRow struct {
	OpponentCommander string  `json:"opponent_commander"`
	Games             int     `json:"games"`
	Wins              int     `json:"wins"`
	Losses            int     `json:"losses"`
	Draws             int     `json:"draws"`
	WinRate           float64 `json:"win_rate"` // wins / games × 100, rounded to one decimal
}

// LoadDeckMatchups returns one MatchupRow per opposing commander the
// given deck has faced, sorted by Games desc then WinRate desc.
//
// `deckKey` is the canonical "owner/id" string the showmatch persistence
// layer writes to showmatch_game_seat.deck_key. Empty string returns no
// rows. Limit ≤ 0 → all opponents.
//
// Implementation is one self-join on showmatch_game_seat: each game the
// deck played emits N-1 (opponent) rows that we group on opponent
// commander to roll up win/loss counts in a single SQL pass.
func LoadDeckMatchups(ctx context.Context, db *sql.DB, deckKey string, limit int) ([]MatchupRow, error) {
	if deckKey == "" {
		return []MatchupRow{}, nil
	}
	q := `
		SELECT
		    op.commander                                         AS opp_commander,
		    COUNT(*)                                             AS games,
		    SUM(CASE WHEN g.winner = me.seat                    THEN 1 ELSE 0 END) AS wins,
		    SUM(CASE WHEN g.winner != me.seat AND g.winner >= 0 THEN 1 ELSE 0 END) AS losses,
		    SUM(CASE WHEN g.winner < 0                          THEN 1 ELSE 0 END) AS draws
		FROM showmatch_game_seat me
		JOIN showmatch_game     g  ON g.game_id  = me.game_id
		JOIN showmatch_game_seat op ON op.game_id = me.game_id AND op.seat != me.seat
		WHERE me.deck_key = ?
		GROUP BY op.commander
		ORDER BY games DESC, wins DESC`
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := db.QueryContext(ctx, q, deckKey)
	if err != nil {
		return nil, fmt.Errorf("matchups query: %w", err)
	}
	defer rows.Close()
	var out []MatchupRow
	for rows.Next() {
		var r MatchupRow
		if err := rows.Scan(&r.OpponentCommander, &r.Games, &r.Wins, &r.Losses, &r.Draws); err != nil {
			return nil, err
		}
		if r.Games > 0 {
			r.WinRate = float64(int(float64(r.Wins)/float64(r.Games)*1000+0.5)) / 10
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
