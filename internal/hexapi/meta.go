package hexapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// handleMeta answers GET /api/meta with a metagame snapshot for the
// future Meta page: top commanders by games played, average win rate
// per commander color identity, the most popular archetypes, and the
// bracket distribution of the current deck pool.
//
// Sources:
//
//   - top_commanders          aggregated from showmatch_game_seat / showmatch_game
//   - color_identity_winrates aggregated from showmatch_game_seat joined to
//                             card_oracle.mana_cost (color identity = the set
//                             of {W,U,B,R,G} symbols in the commander's cost)
//   - top_archetypes          walked from data/decks/*/freya/*.strategy.json
//                             (archetype lives there, not in the DB)
//   - bracket_distribution    GROUP BY bracket on showmatch_elo
//
// The handler degrades gracefully — empty arrays / zero counters when
// a particular source isn't populated yet (no cached oracle, no Freya
// strategy files, etc.) rather than 500ing.
func (h *Handler) handleMeta(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sqlDB := h.cardStatsDB() // same selector as card-stats: showmatch DB if available, else h.db

	resp := map[string]any{
		"total_games":             0,
		"top_commanders":          []any{},
		"color_identity_winrates": []any{},
		"top_archetypes":          []any{},
		"bracket_distribution":    []any{},
	}

	if sqlDB != nil {
		if total, err := metaTotalGames(ctx, sqlDB); err == nil {
			resp["total_games"] = total
		}
		if top, err := metaTopCommanders(ctx, sqlDB, 10); err == nil {
			resp["top_commanders"] = top
		}
		if cs, err := metaColorIdentityWinrates(ctx, sqlDB); err == nil {
			resp["color_identity_winrates"] = cs
		}
		if br, err := metaBracketDistribution(ctx, sqlDB); err == nil {
			resp["bracket_distribution"] = br
		}
	}

	resp["top_archetypes"] = metaTopArchetypes(h.DecksDir, 10)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	_ = json.NewEncoder(w).Encode(resp)
}

type CommanderRow struct {
	Commander string  `json:"commander"`
	Games     int     `json:"games"`
	Wins      int     `json:"wins"`
	WinRate   float64 `json:"win_rate"`
}

func metaTotalGames(ctx context.Context, sqlDB *sql.DB) (int, error) {
	var n int
	err := sqlDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM showmatch_game`).Scan(&n)
	return n, err
}

func metaTopCommanders(ctx context.Context, sqlDB *sql.DB, limit int) ([]CommanderRow, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := sqlDB.QueryContext(ctx, `
		SELECT s.commander,
		       COUNT(*) AS games,
		       SUM(CASE WHEN g.winner = s.seat THEN 1 ELSE 0 END) AS wins
		FROM showmatch_game_seat s
		JOIN showmatch_game g ON g.game_id = s.game_id
		WHERE s.commander != ''
		GROUP BY s.commander
		ORDER BY games DESC, wins DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CommanderRow, 0, limit)
	for rows.Next() {
		var r CommanderRow
		var wins sql.NullInt64
		if err := rows.Scan(&r.Commander, &r.Games, &wins); err != nil {
			return out, err
		}
		r.Wins = int(wins.Int64)
		if r.Games > 0 {
			r.WinRate = float64(r.Wins) / float64(r.Games)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

type ColorIdentityRow struct {
	Colors  string  `json:"colors"`
	Games   int     `json:"games"`
	Wins    int     `json:"wins"`
	WinRate float64 `json:"win_rate"`
}

// colorsFromManaCost returns the canonical WUBRG-ordered identity from
// the symbol stream in a Scryfall mana_cost ("{2}{W}{U}{B}" → "WUB").
// Unknown / colorless costs return "C". Phyrexian and hybrid symbols
// like {W/U} contribute both colors.
func colorsFromManaCost(mc string) string {
	mc = strings.ToUpper(mc)
	seen := map[byte]bool{}
	for _, ch := range mc {
		switch byte(ch) {
		case 'W', 'U', 'B', 'R', 'G':
			seen[byte(ch)] = true
		}
	}
	out := []byte{}
	for _, c := range []byte{'W', 'U', 'B', 'R', 'G'} {
		if seen[c] {
			out = append(out, c)
		}
	}
	if len(out) == 0 {
		return "C"
	}
	return string(out)
}

func metaColorIdentityWinrates(ctx context.Context, sqlDB *sql.DB) ([]ColorIdentityRow, error) {
	// Pull (commander, won) per seat from the game results, then bucket
	// by color identity in Go — SQLite has no string-aggregation that
	// would let us derive colors from {W}{U}... in pure SQL.
	rows, err := sqlDB.QueryContext(ctx, `
		SELECT s.commander,
		       (CASE WHEN g.winner = s.seat THEN 1 ELSE 0 END) AS won
		FROM showmatch_game_seat s
		JOIN showmatch_game g ON g.game_id = s.game_id
		WHERE s.commander != ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bucket := map[string]*ColorIdentityRow{}
	for rows.Next() {
		var commander string
		var won int
		if err := rows.Scan(&commander, &won); err != nil {
			return nil, err
		}
		colors := commanderColors(ctx, sqlDB, commander)
		b, ok := bucket[colors]
		if !ok {
			b = &ColorIdentityRow{Colors: colors}
			bucket[colors] = b
		}
		b.Games++
		b.Wins += won
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]ColorIdentityRow, 0, len(bucket))
	for _, v := range bucket {
		if v.Games > 0 {
			v.WinRate = float64(v.Wins) / float64(v.Games)
		}
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Games != out[j].Games {
			return out[i].Games > out[j].Games
		}
		return out[i].Colors < out[j].Colors
	})
	return out, nil
}

// commanderColors looks up the cached mana_cost for a commander and
// derives its color identity. Falls back to "?" when the card isn't
// in card_oracle so the caller can still bucket games rather than
// drop them.
func commanderColors(ctx context.Context, sqlDB *sql.DB, commander string) string {
	var mc sql.NullString
	err := sqlDB.QueryRowContext(ctx,
		`SELECT mana_cost FROM card_oracle WHERE name = ?`,
		strings.ToLower(commander),
	).Scan(&mc)
	if err != nil || !mc.Valid {
		return "?"
	}
	return colorsFromManaCost(mc.String)
}

type BracketRow struct {
	Bracket int `json:"bracket"`
	Decks   int `json:"decks"`
}

func metaBracketDistribution(ctx context.Context, sqlDB *sql.DB) ([]BracketRow, error) {
	rows, err := sqlDB.QueryContext(ctx, `
		SELECT bracket, COUNT(*) FROM showmatch_elo
		GROUP BY bracket ORDER BY bracket`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []BracketRow{}
	for rows.Next() {
		var r BracketRow
		if err := rows.Scan(&r.Bracket, &r.Decks); err != nil {
			return out, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

type ArchetypeRow struct {
	Archetype string `json:"archetype"`
	Decks     int    `json:"decks"`
}

// metaTopArchetypes walks data/decks/*/freya/*.strategy.json and
// counts the archetype field. Returns up to limit entries sorted by
// deck count. Returns an empty slice (never nil) when DecksDir is
// unset or the walk fails — Meta page can render "no data" cleanly.
func metaTopArchetypes(decksDir string, limit int) []ArchetypeRow {
	if decksDir == "" {
		return []ArchetypeRow{}
	}
	if limit <= 0 {
		limit = 10
	}
	counts := map[string]int{}
	owners, err := os.ReadDir(decksDir)
	if err != nil {
		return []ArchetypeRow{}
	}
	for _, ownerEntry := range owners {
		if !ownerEntry.IsDir() {
			continue
		}
		freyaDir := filepath.Join(decksDir, ownerEntry.Name(), "freya")
		files, ferr := os.ReadDir(freyaDir)
		if ferr != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".strategy.json") {
				continue
			}
			data, rerr := os.ReadFile(filepath.Join(freyaDir, f.Name()))
			if rerr != nil {
				continue
			}
			var s struct {
				Archetype string `json:"archetype"`
			}
			if json.Unmarshal(data, &s) != nil {
				continue
			}
			a := strings.TrimSpace(s.Archetype)
			if a == "" {
				continue
			}
			counts[a]++
		}
	}
	out := make([]ArchetypeRow, 0, len(counts))
	for k, v := range counts {
		out = append(out, ArchetypeRow{Archetype: k, Decks: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Decks != out[j].Decks {
			return out[i].Decks > out[j].Decks
		}
		return out[i].Archetype < out[j].Archetype
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}
