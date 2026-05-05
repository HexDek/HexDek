package hexapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	hexdb "github.com/hexdek/hexdek/internal/db"
)

func TestPincerPageviewAndStitch(t *testing.T) {
	tmp := t.TempDir()
	db, err := hexdb.Open(tmp + "/test.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	h := &Handler{db: db}
	mux := http.NewServeMux()
	h.Register(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	post := func(path, body string) (int, string) {
		resp, err := http.Post(srv.URL+path, "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatalf("post %s: %v", path, err)
		}
		defer resp.Body.Close()
		buf := make([]byte, 4096)
		n, _ := resp.Body.Read(buf)
		return resp.StatusCode, strings.TrimSpace(string(buf[:n]))
	}

	const anon = "11111111-2222-3333-4444-555555555555"

	// Two pageviews from anon.
	if code, _ := post("/api/telemetry/pageview",
		`{"anon_id":"`+anon+`","path":"/decks","timestamp":1700000000000,"referrer":"r"}`); code != 204 {
		t.Errorf("pageview 1: status %d", code)
	}
	if code, _ := post("/api/telemetry/pageview",
		`{"anon_id":"`+anon+`","path":"/leaderboard","timestamp":1700000001000}`); code != 204 {
		t.Errorf("pageview 2: status %d", code)
	}
	// Bad anon_id.
	if code, _ := post("/api/telemetry/pageview", `{"anon_id":"not-a-uuid","path":"/x"}`); code != 400 {
		t.Errorf("bad anon: status %d (want 400)", code)
	}
	// Missing path.
	if code, _ := post("/api/telemetry/pageview", `{"anon_id":"`+anon+`"}`); code != 400 {
		t.Errorf("missing path: status %d (want 400)", code)
	}

	// Stitch.
	code, body := post("/api/telemetry/stitch", `{"anon_id":"`+anon+`","owner":"Alice"}`)
	if code != 200 {
		t.Fatalf("stitch: status %d  body=%s", code, body)
	}
	if !strings.Contains(strings.ReplaceAll(body, " ", ""), `"backfilled":2`) {
		t.Errorf("expected backfill 2, got %s", body)
	}

	var n, owned int
	db.QueryRow("SELECT COUNT(*) FROM pageviews").Scan(&n)
	db.QueryRow("SELECT COUNT(*) FROM pageviews WHERE owner = 'alice'").Scan(&owned)
	if n != 2 || owned != 2 {
		t.Errorf("rows: total=%d owned_by_alice=%d", n, owned)
	}

	var stitch int
	db.QueryRow("SELECT COUNT(*) FROM session_stitch").Scan(&stitch)
	if stitch != 1 {
		t.Errorf("session_stitch rows: %d", stitch)
	}

	// Re-stitch is idempotent.
	post("/api/telemetry/stitch", `{"anon_id":"`+anon+`","owner":"alice"}`)
	db.QueryRow("SELECT COUNT(*) FROM session_stitch").Scan(&stitch)
	if stitch != 1 {
		t.Errorf("after re-stitch session_stitch rows: %d", stitch)
	}
}
