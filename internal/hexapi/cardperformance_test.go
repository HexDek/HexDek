package hexapi

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	hexdb "github.com/hexdek/hexdek/internal/db"
)

func TestCardPerformance_HandlerReturnsAggregate(t *testing.T) {
	tmp := t.TempDir()
	db, err := hexdb.Open(tmp + "/test.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Seed two games' worth of deltas for "Cyclonic Rift": one win on
	// turn 7 with 3 turns on the battlefield, one loss on turn 9 with 1
	// turn on the battlefield.
	if err := hexdb.BatchUpsertCardPerformance(context.Background(), db, []hexdb.CardPerformanceDelta{
		{CardName: "Cyclonic Rift", Win: 1, TurnPlayed: 7, BattlefieldTurns: 3},
	}); err != nil {
		t.Fatalf("seed 1: %v", err)
	}
	if err := hexdb.BatchUpsertCardPerformance(context.Background(), db, []hexdb.CardPerformanceDelta{
		{CardName: "Cyclonic Rift", Win: 0, TurnPlayed: 9, BattlefieldTurns: 1},
	}); err != nil {
		t.Fatalf("seed 2: %v", err)
	}

	h := &Handler{db: db}
	mux := http.NewServeMux()
	h.Register(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/cards/Cyclonic%20Rift/performance")
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

	if body["card_name"] != "Cyclonic Rift" {
		t.Errorf("card_name=%v", body["card_name"])
	}
	if body["games_included"].(float64) != 2 {
		t.Errorf("games_included=%v want 2", body["games_included"])
	}
	if body["wins_when_included"].(float64) != 1 {
		t.Errorf("wins_when_included=%v want 1", body["wins_when_included"])
	}
	if math.Abs(body["win_rate"].(float64)-0.5) > 1e-6 {
		t.Errorf("win_rate=%v want 0.5", body["win_rate"])
	}
	if math.Abs(body["avg_turn_played"].(float64)-8.0) > 1e-6 {
		t.Errorf("avg_turn_played=%v want 8", body["avg_turn_played"])
	}
	if math.Abs(body["avg_battlefield_time"].(float64)-2.0) > 1e-6 {
		t.Errorf("avg_battlefield_time=%v want 2", body["avg_battlefield_time"])
	}
}

func TestCardPerformance_HandlerUnknownCardReturnsZeros(t *testing.T) {
	db, err := hexdb.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	h := &Handler{db: db}
	mux := http.NewServeMux()
	h.Register(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/cards/Nothing/performance")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	body, _ := readJSON(resp.Body)
	if body["games_included"].(float64) != 0 {
		t.Errorf("games_included should be 0 for unknown card, got %v", body["games_included"])
	}
	if body["win_rate"].(float64) != 0 {
		t.Errorf("win_rate should be 0 for unknown card, got %v", body["win_rate"])
	}
}

func readJSON(r interface {
	Read(p []byte) (int, error)
}) (map[string]any, error) {
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	out := map[string]any{}
	if err := json.NewDecoder(strings.NewReader(string(buf[:n]))).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}
