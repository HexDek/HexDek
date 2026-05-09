package hexapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// setupCloneTest builds a Handler with an in-memory SQLite, an alice
// source deck on disk, and a registered router. The caller invokes
// requests against the returned mux as bob (the would-be cloner).
func setupCloneTest(t *testing.T) (*Handler, http.Handler, string) {
	t.Helper()
	tmp := t.TempDir()
	decksDir := filepath.Join(tmp, "decks")
	if err := os.MkdirAll(filepath.Join(decksDir, "alice"), 0o755); err != nil {
		t.Fatal(err)
	}
	deckBody := "COMMANDER: Krenko, Mob Boss\n1 Sol Ring\n20 Mountain\n1 Goblin Chieftain\n"
	if err := os.WriteFile(
		filepath.Join(decksDir, "alice", "krenko.txt"),
		[]byte(deckBody), 0o644); err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite", filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := EnsureDeckMetaSchema(context.Background(), db); err != nil {
		t.Fatalf("schema deck_meta: %v", err)
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS import_log (
		    id           INTEGER PRIMARY KEY AUTOINCREMENT,
		    owner        TEXT NOT NULL,
		    deck_key     TEXT NOT NULL,
		    deck_name    TEXT NOT NULL DEFAULT '',
		    commander    TEXT NOT NULL DEFAULT '',
		    source       TEXT NOT NULL,
		    source_url   TEXT NOT NULL DEFAULT '',
		    card_count   INTEGER NOT NULL DEFAULT 0,
		    imported_at  INTEGER NOT NULL
		);`); err != nil {
		t.Fatalf("schema import_log: %v", err)
	}

	h := &Handler{DecksDir: decksDir}
	h.SetDB(db)
	mux := http.NewServeMux()
	h.Register(mux)
	return h, mux, decksDir
}

func TestCloneDeck_HappyPath(t *testing.T) {
	h, mux, decksDir := setupCloneTest(t)

	req := httptest.NewRequest("POST", "/api/decks/alice/krenko/clone", nil)
	req.Header.Set("X-HexDek-Owner", "bob")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("clone: code=%d body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["owner"] != "bob" {
		t.Errorf("owner = %v, want bob", resp["owner"])
	}
	if resp["cloned_from"] != "alice/krenko" {
		t.Errorf("cloned_from = %v, want alice/krenko", resp["cloned_from"])
	}
	id, _ := resp["id"].(string)
	if id == "" {
		t.Fatal("response missing id")
	}

	// File landed under bob's owner dir, not alice's.
	wantDeck := filepath.Join(decksDir, "bob", id+".txt")
	if _, err := os.Stat(wantDeck); err != nil {
		t.Fatalf("clone file not at %s: %v", wantDeck, err)
	}

	// cloned_from is persisted and surfaces on GET.
	getReq := httptest.NewRequest("GET", "/api/decks/bob/"+id, nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)
	if getRec.Code != 200 {
		t.Fatalf("get: code=%d body=%s", getRec.Code, getRec.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(getRec.Body.Bytes(), &got); err != nil {
		t.Fatalf("get decode: %v", err)
	}
	if got["cloned_from"] != "alice/krenko" {
		t.Errorf("GET cloned_from = %v, want alice/krenko", got["cloned_from"])
	}

	// clone_log row exists for the bob owner.
	var n int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM clone_log WHERE owner = ?`, "bob").
		Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("clone_log rows = %d, want 1", n)
	}
}

func TestCloneDeck_Unauthenticated(t *testing.T) {
	_, mux, _ := setupCloneTest(t)

	req := httptest.NewRequest("POST", "/api/decks/alice/krenko/clone", nil)
	// No X-HexDek-Owner header.
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("unauth clone: code=%d, want 401", rec.Code)
	}
}

func TestCloneDeck_RejectsSelfClone(t *testing.T) {
	_, mux, _ := setupCloneTest(t)

	req := httptest.NewRequest("POST", "/api/decks/alice/krenko/clone", nil)
	req.Header.Set("X-HexDek-Owner", "alice")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("self-clone: code=%d, want 400", rec.Code)
	}
}

func TestCloneDeck_DeckNotFound(t *testing.T) {
	_, mux, _ := setupCloneTest(t)

	req := httptest.NewRequest("POST", "/api/decks/alice/ghost/clone", nil)
	req.Header.Set("X-HexDek-Owner", "bob")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("missing src: code=%d, want 404", rec.Code)
	}
}

func TestCloneDeck_RateLimit(t *testing.T) {
	h, mux, _ := setupCloneTest(t)

	// Pre-load clone_log with CloneRateLimit recent rows for bob.
	now := time.Now().Unix()
	for i := 0; i < CloneRateLimit; i++ {
		if _, err := h.db.Exec(
			`INSERT INTO clone_log (owner, src_key, dst_key, cloned_at) VALUES (?, ?, ?, ?)`,
			"bob", "alice/krenko", "bob/k_clone", now); err != nil {
			t.Fatal(err)
		}
	}

	req := httptest.NewRequest("POST", "/api/decks/alice/krenko/clone", nil)
	req.Header.Set("X-HexDek-Owner", "bob")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("rate-limited: code=%d body=%s, want 429",
			rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("rate-limited response missing Retry-After header")
	}
}

func TestCloneDeck_NameCollisionAppendsSuffix(t *testing.T) {
	_, mux, decksDir := setupCloneTest(t)
	if err := os.MkdirAll(filepath.Join(decksDir, "bob"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Pre-create the natural clone destination so the handler has to
	// pick krenko_clone2.
	if err := os.WriteFile(
		filepath.Join(decksDir, "bob", "krenko_clone.txt"),
		[]byte("placeholder"), 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/api/decks/alice/krenko/clone", nil)
	req.Header.Set("X-HexDek-Owner", "bob")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("clone: code=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	id, _ := resp["id"].(string)
	if id == "krenko_clone" {
		t.Errorf("collision: clone reused existing id %q", id)
	}
}
