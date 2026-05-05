// hexdek-heimdall-scanner reports per-card health from the showmatch DB.
//
// It joins per-card win/board-presence counters (card_win_stats) against
// commander-aggregate ELO records (showmatch_elo) to surface two signals:
//
//   1. wr_delta  = win_rate_when_present - commander_overall_win_rate
//                  Cards with wr_delta < -0.05 are flagged: their presence
//                  in a list correlates with the deck winning *less* than
//                  the deck's baseline (broken handler / dead-card signal).
//
//   2. cast_to_outcome_delta = on_board_at_win_rate - win_rate_when_present
//                  Negative values mean the card is in winning lists but
//                  rarely on the board at the win — i.e. it gets cast and
//                  the position deteriorates (gets removed, sacrificed for
//                  no win, or never resolves into outcome).
//
// Output: data/heimdall/card_health.json, ranked by severity (most
// negative wr_delta first).
//
// Note: the project schema's per-card table is `card_win_stats` (the
// `showmatch_card_stats` name in reset-elo is tolerated-if-missing legacy
// referring to the same conceptual table). This tool reads card_win_stats.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	_ "modernc.org/sqlite"
)

type CardHealth struct {
	CardName             string  `json:"card_name"`
	Commander            string  `json:"commander"`
	Games                int     `json:"games"`
	Wins                 int     `json:"wins"`
	OnBoardAtWin         int     `json:"on_board_at_win"`
	WinRateWhenPresent   float64 `json:"win_rate_when_present"`
	OverallWinRate       float64 `json:"overall_win_rate"`
	WRDelta              float64 `json:"wr_delta"`
	OnBoardAtWinRate     float64 `json:"on_board_at_win_rate"`
	CastToOutcomeDelta   float64 `json:"cast_to_outcome_delta"`
	BrokenHandlerSuspect bool    `json:"broken_handler_suspect"`
}

type Report struct {
	Source        string       `json:"source"`
	CommanderRows int          `json:"commander_rows"`
	CardRows      int          `json:"card_rows"`
	MinGames      int          `json:"min_games"`
	BrokenCount   int          `json:"broken_count"`
	Cards         []CardHealth `json:"cards"`
}

func main() {
	dbPath := flag.String("db", "data/hexdek.db", "path to SQLite DB")
	out := flag.String("out", "data/heimdall/card_health.json", "output JSON path")
	minGames := flag.Int("min-games", 5, "minimum games before a card is considered")
	threshold := flag.Float64("threshold", 0.05, "wr_delta below -threshold flags broken_handler_suspect")
	flag.Parse()

	if _, err := os.Stat(*dbPath); err != nil {
		log.Fatalf("db not found at %s: %v", *dbPath, err)
	}

	db, err := sql.Open("sqlite", *dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	commanderWR, cmdRows, err := loadCommanderWinRates(ctx, db)
	if err != nil {
		log.Fatalf("load commander win rates: %v", err)
	}

	cards, err := loadCardStats(ctx, db, *minGames)
	if err != nil {
		log.Fatalf("load card stats: %v", err)
	}

	report := Report{
		Source:        *dbPath,
		CommanderRows: cmdRows,
		CardRows:      len(cards),
		MinGames:      *minGames,
		Cards:         make([]CardHealth, 0, len(cards)),
	}

	for _, c := range cards {
		if c.Games <= 0 {
			continue
		}
		wrPresent := float64(c.Wins) / float64(c.Games)
		overall := commanderWR[c.Commander]
		var onBoardRate float64
		if c.Wins > 0 {
			onBoardRate = float64(c.OnBoardAtWin) / float64(c.Wins)
		}
		h := CardHealth{
			CardName:           c.CardName,
			Commander:          c.Commander,
			Games:              c.Games,
			Wins:               c.Wins,
			OnBoardAtWin:       c.OnBoardAtWin,
			WinRateWhenPresent: wrPresent,
			OverallWinRate:     overall,
			WRDelta:            wrPresent - overall,
			OnBoardAtWinRate:   onBoardRate,
			CastToOutcomeDelta: onBoardRate - wrPresent,
		}
		if h.WRDelta < -(*threshold) {
			h.BrokenHandlerSuspect = true
			report.BrokenCount++
		}
		report.Cards = append(report.Cards, h)
	}

	sort.Slice(report.Cards, func(i, j int) bool {
		return report.Cards[i].WRDelta < report.Cards[j].WRDelta
	})

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		log.Fatalf("mkdir output: %v", err)
	}
	f, err := os.Create(*out)
	if err != nil {
		log.Fatalf("create output: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		log.Fatalf("encode report: %v", err)
	}

	fmt.Printf("scanned %d cards across %d commanders; flagged %d broken-handler suspects (wr_delta < -%.2f)\n",
		len(report.Cards), report.CommanderRows, report.BrokenCount, *threshold)
	fmt.Printf("wrote %s\n", *out)
}

func loadCommanderWinRates(ctx context.Context, db *sql.DB) (map[string]float64, int, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT commander, SUM(games) AS g, SUM(wins) AS w
		 FROM showmatch_elo
		 WHERE commander != ''
		 GROUP BY commander`)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make(map[string]float64)
	count := 0
	for rows.Next() {
		var commander string
		var games, wins sql.NullInt64
		if err := rows.Scan(&commander, &games, &wins); err != nil {
			return nil, 0, err
		}
		count++
		if !games.Valid || games.Int64 <= 0 {
			out[commander] = 0
			continue
		}
		out[commander] = float64(wins.Int64) / float64(games.Int64)
	}
	return out, count, rows.Err()
}

type cardRow struct {
	CardName     string
	Commander    string
	Games        int
	Wins         int
	OnBoardAtWin int
}

func loadCardStats(ctx context.Context, db *sql.DB, minGames int) ([]cardRow, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT card_name, commander, games, wins, on_board_at_win
		 FROM card_win_stats
		 WHERE games >= ?`, minGames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []cardRow
	for rows.Next() {
		var r cardRow
		if err := rows.Scan(&r.CardName, &r.Commander, &r.Games, &r.Wins, &r.OnBoardAtWin); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
