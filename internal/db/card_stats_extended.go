package db

import (
	"context"
	"database/sql"
)

// card_stats_extended — queries for the per-card analytics endpoints at
// /api/card-stats/{cardName} and /api/card-stats/{cardName}/by-commander.
//
// These aggregate data from card_stats (cross-commander deck-list win rate),
// card_win_stats (per-commander battlefield presence), and showmatch_elo
// (inclusion rate, bracket distribution).

// CardStatsOverview is the aggregated response for a single card: overall
// win rate, inclusion rate across known decks, top commanders, and bracket
// distribution.
type CardStatsOverview struct {
	CardName      string                  `json:"card_name"`
	WinRate       float64                 `json:"win_rate"`
	GamesPlayed   int                     `json:"games_played"`
	Wins          int                     `json:"wins"`
	Losses        int                     `json:"losses"`
	InclusionRate float64                 `json:"inclusion_rate"`
	DecksUsing    int                     `json:"decks_using"`
	TotalDecks    int                     `json:"total_decks"`
	TopCommanders []CardCommanderStat     `json:"top_commanders"`
	BracketDist   []CardBracketEntry      `json:"bracket_distribution"`
}

// CardCommanderStat is a row in the "top commanders using this card" list.
type CardCommanderStat struct {
	Commander string  `json:"commander"`
	Games     int     `json:"games"`
	Wins      int     `json:"wins"`
	WinRate   float64 `json:"win_rate"`
}

// CardBracketEntry counts how many decks at a given bracket include this card.
type CardBracketEntry struct {
	Bracket int `json:"bracket"`
	Count   int `json:"count"`
}

// LoadCardStatsByCommander returns per-commander breakdown for a given card
// name from the card_win_stats table. Sorted by games desc, limited.
func LoadCardStatsByCommander(ctx context.Context, sqlDB *sql.DB, cardName string, limit int) ([]CardCommanderStat, error) {
	if sqlDB == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	rows, err := sqlDB.QueryContext(ctx,
		`SELECT commander, games, wins
		 FROM card_win_stats
		 WHERE card_name = ? AND games >= 3
		 ORDER BY games DESC
		 LIMIT ?`, cardName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CardCommanderStat
	for rows.Next() {
		var s CardCommanderStat
		if err := rows.Scan(&s.Commander, &s.Games, &s.Wins); err != nil {
			return nil, err
		}
		if s.Games > 0 {
			s.WinRate = float64(s.Wins) / float64(s.Games)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// LoadCardInclusionRate calculates what fraction of tracked decks (in
// showmatch_elo) contain this card in their card_win_stats entries.
// Returns (decks_using, total_decks). This is an approximation: a deck
// "uses" the card if it appears in card_win_stats with the same
// commander that the deck runs. For decks that have never been in a game
// with this card on the battlefield, they won't appear — but for the
// engine analytics use case this is the correct denominator (decks that
// have actually played).
func LoadCardInclusionRate(ctx context.Context, sqlDB *sql.DB, cardName string) (decksUsing, totalDecks int, err error) {
	if sqlDB == nil {
		return 0, 0, nil
	}
	// Count distinct commanders that have played with this card.
	err = sqlDB.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT commander) FROM card_win_stats WHERE card_name = ?`,
		cardName).Scan(&decksUsing)
	if err != nil {
		return 0, 0, err
	}
	// Total distinct decks that have played at least one game.
	err = sqlDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM showmatch_elo WHERE games > 0`).Scan(&totalDecks)
	if err != nil {
		return 0, 0, err
	}
	return decksUsing, totalDecks, nil
}

// LoadCardBracketDistribution returns the bracket breakdown for decks
// that include this card. It joins card_win_stats (keyed card_name,
// commander) against showmatch_elo (which has a bracket column) via the
// commander name.
func LoadCardBracketDistribution(ctx context.Context, sqlDB *sql.DB, cardName string) ([]CardBracketEntry, error) {
	if sqlDB == nil {
		return nil, nil
	}
	rows, err := sqlDB.QueryContext(ctx,
		`SELECT e.bracket, COUNT(DISTINCT e.deck_key)
		 FROM card_win_stats cw
		 JOIN showmatch_elo e ON e.commander = cw.commander
		 WHERE cw.card_name = ? AND e.bracket > 0
		 GROUP BY e.bracket
		 ORDER BY e.bracket`, cardName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CardBracketEntry
	for rows.Next() {
		var entry CardBracketEntry
		if err := rows.Scan(&entry.Bracket, &entry.Count); err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}
