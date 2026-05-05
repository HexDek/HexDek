package userprofile

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestParseAcceptLanguage(t *testing.T) {
	cases := []struct {
		header, want string
	}{
		{"", ""},
		{"en-US,en;q=0.9", "US"},
		{"en_US,en;q=0.9", "US"},
		{"pl", "PL"},
		{"PL", "PL"},
		{"ja", "JP"},
		{"fr-CA;q=0.9,en;q=0.5", "CA"},
		{"en;q=0.5,fr-CA;q=0.9", "CA"}, // q-sort takes precedence over header order
		{"*", ""},
		{"xx-YY", "YY"},
		{"unknown", ""}, // unmapped language with no region
		{"pt-BR,pt;q=0.8,en;q=0.5", "BR"},
		{"de", "DE"},
		{"zh-CN,zh;q=0.9,en;q=0.7", "CN"},
	}
	for _, c := range cases {
		if got := ParseAcceptLanguage(c.header); got != c.want {
			t.Errorf("ParseAcceptLanguage(%q) = %q; want %q", c.header, got, c.want)
		}
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := EnsureSchema(context.Background(), db); err != nil {
		t.Fatalf("schema: %v", err)
	}
	return db
}

func TestSetCountryIfMissingFirstWriteWins(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	if err := SetCountryIfMissing(ctx, db, "alice", "US"); err != nil {
		t.Fatalf("first set: %v", err)
	}
	// Second call with a different country must NOT overwrite.
	if err := SetCountryIfMissing(ctx, db, "alice", "PL"); err != nil {
		t.Fatalf("second set: %v", err)
	}
	got, err := GetCountry(ctx, db, "alice")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "US" {
		t.Errorf("expected US (first-write-wins), got %q", got)
	}
}

func TestSetCountryForcesUpsert(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	_ = SetCountryIfMissing(ctx, db, "bob", "US")
	if err := SetCountry(ctx, db, "bob", "PL"); err != nil {
		t.Fatalf("force: %v", err)
	}
	got, _ := GetCountry(ctx, db, "bob")
	if got != "PL" {
		t.Errorf("expected PL after forced upsert, got %q", got)
	}
}

func TestBulkCountries(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	_ = SetCountryIfMissing(ctx, db, "alice", "US")
	_ = SetCountryIfMissing(ctx, db, "carol", "JP")

	got, err := BulkCountries(ctx, db, []string{"alice", "bob", "carol", "alice"})
	if err != nil {
		t.Fatalf("bulk: %v", err)
	}
	if got["alice"] != "US" || got["carol"] != "JP" {
		t.Errorf("missing entries: %+v", got)
	}
	if _, ok := got["bob"]; ok {
		t.Errorf("bob should be absent (no row), got %+v", got)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 entries, got %d (%+v)", len(got), got)
	}
}

func TestBulkCountriesEmpty(t *testing.T) {
	db := openTestDB(t)
	got, err := BulkCountries(context.Background(), db, nil)
	if err != nil {
		t.Fatalf("nil owners: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %+v", got)
	}
}

func TestLocaleMiddlewareStashesCountry(t *testing.T) {
	var seen string
	h := LocaleMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = CountryFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", "ja-JP,en;q=0.5")
	h.ServeHTTP(rec, req)
	if seen != "JP" {
		t.Errorf("expected JP from middleware, got %q", seen)
	}
}
