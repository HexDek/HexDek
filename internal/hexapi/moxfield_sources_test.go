package hexapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// Filename-convention scan of data/decks/moxfield* directories. Verifies
// the suffix-as-deck-ID extraction and Moxfield URL synthesis.
func TestHandleMoxfieldSources_FilenameConvention(t *testing.T) {
	decksDir := t.TempDir()
	moxDir := filepath.Join(decksDir, "moxfield")
	if err := os.MkdirAll(moxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, fn := range []string{
		"aang_and_katara_b2_puschel_o5qbswWq.txt",
		"abdel_adrian_gorion_s_ward_b4_badomen77_WK6MLBfq.txt",
		// Invalid (no underscore separator): should be skipped.
		"weirdfile.txt",
		// Too-short suffix: should be skipped.
		"foo_bar_a.txt",
	} {
		if err := os.WriteFile(filepath.Join(moxDir, fn), []byte("1 Card\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	h := &Handler{DecksDir: decksDir}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest("GET", "/api/imports/source/moxfield", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var got struct {
		Total   int `json:"total"`
		Sources []struct {
			Owner       string `json:"owner"`
			DeckKey     string `json:"deck_key"`
			MoxfieldID  string `json:"moxfield_id"`
			MoxfieldURL string `json:"moxfield_url"`
			Source      string `json:"source"`
		} `json:"sources"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Two valid filenames, two skipped.
	if got.Total != 2 {
		t.Fatalf("expected 2 sources, got %d (body=%s)", got.Total, rec.Body.String())
	}
	// Find the Aang entry and verify the URL synthesis.
	var aang *struct {
		Owner       string `json:"owner"`
		DeckKey     string `json:"deck_key"`
		MoxfieldID  string `json:"moxfield_id"`
		MoxfieldURL string `json:"moxfield_url"`
		Source      string `json:"source"`
	}
	for i, s := range got.Sources {
		if s.MoxfieldID == "o5qbswWq" {
			aang = &got.Sources[i]
			break
		}
	}
	if aang == nil {
		t.Fatalf("expected to find aang entry with moxfield_id=o5qbswWq; got %+v", got.Sources)
	}
	if aang.Owner != "moxfield" {
		t.Errorf("expected owner=moxfield, got %q", aang.Owner)
	}
	if aang.MoxfieldURL != "https://www.moxfield.com/decks/o5qbswWq" {
		t.Errorf("expected url=https://www.moxfield.com/decks/o5qbswWq, got %q", aang.MoxfieldURL)
	}
	if aang.Source != "filename_convention" {
		t.Errorf("expected source=filename_convention, got %q", aang.Source)
	}
}

// Limit parameter caps the response.
func TestHandleMoxfieldSources_LimitCap(t *testing.T) {
	decksDir := t.TempDir()
	moxDir := filepath.Join(decksDir, "moxfield")
	if err := os.MkdirAll(moxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Drop 5 valid entries.
	for i := 0; i < 5; i++ {
		fn := "deck_" + string(rune('a'+i)) + "_xyzabcd1.txt"
		if err := os.WriteFile(filepath.Join(moxDir, fn), []byte("1 Card\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	h := &Handler{DecksDir: decksDir}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest("GET", "/api/imports/source/moxfield?limit=3", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got struct {
		Total int `json:"total"`
	}
	json.Unmarshal(rec.Body.Bytes(), &got)
	if got.Total != 3 {
		t.Errorf("limit=3 should yield 3 entries, got %d", got.Total)
	}
}

// Empty case — no moxfield directories and no Showmatch SQLite — returns
// an empty list, not an error.
func TestHandleMoxfieldSources_EmptyState(t *testing.T) {
	h := &Handler{DecksDir: t.TempDir()}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest("GET", "/api/imports/source/moxfield", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got struct {
		Total   int           `json:"total"`
		Sources []interface{} `json:"sources"`
	}
	json.Unmarshal(rec.Body.Bytes(), &got)
	if got.Total != 0 {
		t.Errorf("expected 0 sources in empty state, got %d", got.Total)
	}
}
