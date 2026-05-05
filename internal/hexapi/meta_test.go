package hexapi

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	hexdb "github.com/hexdek/hexdek/internal/db"
)

func TestColorsFromManaCost(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"{2}{W}{U}{B}", "WUB"},
		{"{G}{G}", "G"},
		{"{B}{R}{W}", "WBR"}, // canonical WUBRG order
		{"{2}", "C"},
		{"", "C"},
		{"{w/u}{r}", "WUR"}, // hybrid contributes both
	}
	for _, c := range cases {
		if got := colorsFromManaCost(c.in); got != c.want {
			t.Errorf("colorsFromManaCost(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestHandleMeta_AggregatesGameAndDeckData(t *testing.T) {
	tmp := t.TempDir()
	db, err := hexdb.Open(tmp + "/test.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	// Three games. Atraxa wins 2, Edgar wins 1.
	for i := 0; i < 3; i++ {
		winner := 0
		winnerName := "Atraxa, Praetors' Voice"
		if i == 2 {
			winner = 1
			winnerName = "Edgar Markov"
		}
		gameRec := hexdb.GameRecord{
			StartedAt:  time.Now().Add(-time.Hour).Unix(),
			FinishedAt: time.Now().Unix(),
			Turns:      10,
			Winner:     winner,
			WinnerName: winnerName,
			EndReason:  "last_seat_standing",
		}
		seats := []hexdb.GameSeatRecord{
			{Seat: 0, Commander: "Atraxa, Praetors' Voice", Life: 40, Lost: false},
			{Seat: 1, Commander: "Edgar Markov", Life: 0, Lost: true},
		}
		if _, err := hexdb.PersistGameTx(ctx, db, gameRec, seats); err != nil {
			t.Fatalf("persist: %v", err)
		}
	}

	// Card oracle for color identity lookup.
	if _, err := db.ExecContext(ctx, `
		INSERT INTO card_oracle (name, display_name, scryfall_id, mana_cost, cmc, type_line, oracle_text, image_uri_normal, image_uri_art, set_code, cached_at)
		VALUES (?, ?, '', ?, 0, '', '', '', '', '', 0)`,
		"atraxa, praetors' voice", "Atraxa, Praetors' Voice", "{G}{W}{U}{B}"); err != nil {
		t.Fatalf("oracle insert: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO card_oracle (name, display_name, scryfall_id, mana_cost, cmc, type_line, oracle_text, image_uri_normal, image_uri_art, set_code, cached_at)
		VALUES (?, ?, '', ?, 0, '', '', '', '', '', 0)`,
		"edgar markov", "Edgar Markov", "{R}{W}{B}"); err != nil {
		t.Fatalf("oracle insert: %v", err)
	}

	// Bracket data via showmatch_elo.
	for _, row := range []struct {
		key string
		br  int
	}{
		{"alice/atraxa", 4},
		{"bob/edgar", 3},
		{"carol/edgar2", 3},
	} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO showmatch_elo (deck_key, commander, owner, rating, hex_rating, games, wins, losses, delta, hex_delta, bracket, updated_at)
			VALUES (?, '', '', 1500, 0, 0, 0, 0, 0, 0, ?, 0)`, row.key, row.br); err != nil {
			t.Fatalf("elo insert: %v", err)
		}
	}

	// Synthetic Freya strategy.json files for archetype counts.
	decksDir := tmp + "/decks"
	for owner, archs := range map[string][]string{
		"alice": {"Combo", "Combo"},
		"bob":   {"Voltron"},
		"carol": {"Combo", "Stax"},
	} {
		ownerFreya := filepath.Join(decksDir, owner, "freya")
		if err := os.MkdirAll(ownerFreya, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		for i, a := range archs {
			path := filepath.Join(ownerFreya, owner+string(rune('a'+i))+".strategy.json")
			if err := os.WriteFile(path, []byte(`{"archetype":"`+a+`"}`), 0o644); err != nil {
				t.Fatalf("write freya: %v", err)
			}
		}
	}

	h := &Handler{db: db, DecksDir: decksDir}
	mux := http.NewServeMux()
	h.Register(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/meta")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body["total_games"].(float64) != 3 {
		t.Errorf("total_games = %v, want 3", body["total_games"])
	}

	cmdrs := body["top_commanders"].([]any)
	if len(cmdrs) != 2 {
		t.Fatalf("top_commanders len = %d, want 2", len(cmdrs))
	}
	first := cmdrs[0].(map[string]any)
	if first["commander"] != "Atraxa, Praetors' Voice" {
		t.Errorf("top commander = %v", first["commander"])
	}
	if first["games"].(float64) != 3 {
		t.Errorf("Atraxa games = %v, want 3", first["games"])
	}
	if first["wins"].(float64) != 2 {
		t.Errorf("Atraxa wins = %v, want 2", first["wins"])
	}
	if math.Abs(first["win_rate"].(float64)-2.0/3.0) > 1e-6 {
		t.Errorf("Atraxa win_rate = %v, want 0.6667", first["win_rate"])
	}

	colors := body["color_identity_winrates"].([]any)
	if len(colors) < 2 {
		t.Fatalf("expected at least 2 color identity buckets, got %d", len(colors))
	}
	// Atraxa = WUBG, Edgar = WBR. Both should appear.
	seen := map[string]bool{}
	for _, c := range colors {
		seen[c.(map[string]any)["colors"].(string)] = true
	}
	if !seen["WUBG"] || !seen["WBR"] {
		t.Errorf("expected WUBG and WBR buckets, got %v", seen)
	}

	br := body["bracket_distribution"].([]any)
	if len(br) != 2 {
		t.Fatalf("bracket_distribution len = %d, want 2 (B3 and B4)", len(br))
	}
	// b3 has 2 decks, b4 has 1; ORDER BY bracket ASC.
	if br[0].(map[string]any)["bracket"].(float64) != 3 ||
		br[0].(map[string]any)["decks"].(float64) != 2 {
		t.Errorf("first bracket row = %+v", br[0])
	}

	archs := body["top_archetypes"].([]any)
	if len(archs) != 3 {
		t.Fatalf("top_archetypes len = %d, want 3", len(archs))
	}
	combo := archs[0].(map[string]any)
	if combo["archetype"] != "Combo" || combo["decks"].(float64) != 3 {
		t.Errorf("top archetype = %+v, want Combo 3", combo)
	}
}
