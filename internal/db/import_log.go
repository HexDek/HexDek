package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ImportLogEntry struct {
	ID         int64  `json:"id"`
	Owner      string `json:"owner"`
	DeckKey    string `json:"deck_key"`
	DeckName   string `json:"deck_name"`
	Commander  string `json:"commander"`
	Source     string `json:"source"`
	SourceURL  string `json:"source_url,omitempty"`
	CardCount  int    `json:"card_count"`
	ImportedAt int64  `json:"imported_at"`
}

func InsertImportLog(ctx context.Context, db *sql.DB, e ImportLogEntry) (int64, error) {
	if e.ImportedAt == 0 {
		e.ImportedAt = time.Now().Unix()
	}
	res, err := db.ExecContext(ctx,
		`INSERT INTO import_log (owner, deck_key, deck_name, commander, source, source_url, card_count, imported_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		e.Owner, e.DeckKey, e.DeckName, e.Commander, e.Source, e.SourceURL, e.CardCount, e.ImportedAt)
	if err != nil {
		return 0, fmt.Errorf("insert import_log: %w", err)
	}
	return res.LastInsertId()
}

func ListImportLogs(ctx context.Context, db *sql.DB, owner string, limit int) ([]ImportLogEntry, error) {
	if limit <= 0 || limit > 200 {
		limit = 25
	}
	rows, err := db.QueryContext(ctx,
		`SELECT id, owner, deck_key, deck_name, commander, source, source_url, card_count, imported_at
		 FROM import_log
		 WHERE owner = ?
		 ORDER BY imported_at DESC, id DESC
		 LIMIT ?`,
		owner, limit)
	if err != nil {
		return nil, fmt.Errorf("query import_log: %w", err)
	}
	defer rows.Close()

	var out []ImportLogEntry
	for rows.Next() {
		var e ImportLogEntry
		if err := rows.Scan(&e.ID, &e.Owner, &e.DeckKey, &e.DeckName, &e.Commander, &e.Source, &e.SourceURL, &e.CardCount, &e.ImportedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
