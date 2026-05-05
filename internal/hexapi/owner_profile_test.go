package hexapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestParseAcceptLanguageCountry(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"en-US,en;q=0.9", "US"},
		{"en-US", "US"},
		{"fr-CA;q=0.8,fr;q=0.5", "CA"},
		{"ja-JP", "JP"},
		{"zh-Hant-TW", "TW"}, // script subtag in middle, region last
		{"es-419", ""},       // numeric UN M.49 region — not a 2-letter code
		{"en", ""},           // language only
		{"*", ""},
		{"", ""},
		{"en-us;q=1", "US"}, // lowercase region → uppercased
	}
	for _, tc := range cases {
		if got := parseAcceptLanguageCountry(tc.in); got != tc.want {
			t.Errorf("parseAcceptLanguageCountry(%q) = %q, want %q",
				tc.in, got, tc.want)
		}
	}
}

func TestHandleOwnerProfile_StoredPrefWinsOverAcceptLanguage(t *testing.T) {
	dir := t.TempDir()
	decksDir := filepath.Join(dir, "decks")
	profilesDir := filepath.Join(dir, "profiles")
	if err := os.MkdirAll(decksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(profilesDir, "alice.json"),
		[]byte(`{"country":"jp"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	h := &Handler{DecksDir: decksDir}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest("GET", "/api/profile/alice", nil)
	req.Header.Set("Accept-Language", "fr-FR,fr;q=0.9")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var got OwnerProfile
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Country != "JP" {
		t.Errorf("stored 'jp' should win over Accept-Language FR; got %q", got.Country)
	}
	if got.Owner != "alice" {
		t.Errorf("Owner = %q, want alice", got.Owner)
	}
}

func TestHandleOwnerProfile_AcceptLanguageFallback(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "decks"), 0o755); err != nil {
		t.Fatal(err)
	}

	h := &Handler{DecksDir: filepath.Join(dir, "decks")}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest("GET", "/api/profile/bob", nil)
	req.Header.Set("Accept-Language", "es-MX")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var got OwnerProfile
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Country != "MX" {
		t.Errorf("expected MX from Accept-Language; got %q", got.Country)
	}
}

func TestHandleOwnerProfilesBatch(t *testing.T) {
	dir := t.TempDir()
	decksDir := filepath.Join(dir, "decks")
	profilesDir := filepath.Join(dir, "profiles")
	for _, d := range []string{decksDir, profilesDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(profilesDir, "alice.json"),
		[]byte(`{"country":"JP"}`), 0o644)
	os.WriteFile(filepath.Join(profilesDir, "carol.json"),
		[]byte(`{"country":"BR"}`), 0o644)

	h := &Handler{DecksDir: decksDir}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest("GET", "/api/profiles?owners=alice,bob,carol", nil)
	req.Header.Set("Accept-Language", "de-DE")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["alice"] != "JP" {
		t.Errorf("alice: stored JP expected; got %q", got["alice"])
	}
	if got["carol"] != "BR" {
		t.Errorf("carol: stored BR expected; got %q", got["carol"])
	}
	if got["bob"] != "DE" {
		t.Errorf("bob: Accept-Language DE fallback expected; got %q", got["bob"])
	}
}

func TestHandleOwnerProfilesBatch_DedupAndCap(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "decks"), 0o755); err != nil {
		t.Fatal(err)
	}
	h := &Handler{DecksDir: filepath.Join(dir, "decks")}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest("GET", "/api/profiles?owners=alice,alice,alice", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var got map[string]string
	json.Unmarshal(rec.Body.Bytes(), &got)
	if len(got) != 1 {
		t.Errorf("dedup: expected 1 entry, got %d (%v)", len(got), got)
	}
}
