package pincer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hexdek/hexdek/internal/db"
)

func newTestTracker(t *testing.T) *Tracker {
	t.Helper()
	// Use a tempfile DB rather than ":memory:" — SQLite's anonymous in-memory
	// mode gives each pool connection its own database, so schema applied on
	// one connection isn't visible on another.
	dbPath := t.TempDir() + "/pincer_test.db"
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	tr, err := New(database, Options{AdminToken: "secret-admin"})
	if err != nil {
		t.Fatalf("new tracker: %v", err)
	}
	return tr
}

func TestMiddlewareIssuesCookieOnFirstVisit(t *testing.T) {
	tr := newTestTracker(t)
	var seen string
	h := tr.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = FromContext(r.Context())
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	cookies := rr.Result().Cookies()
	var sid string
	for _, c := range cookies {
		if c.Name == CookieName {
			sid = c.Value
		}
	}
	if sid == "" {
		t.Fatal("no hexdek_session cookie set on first visit")
	}
	if !looksLikeUUID(sid) {
		t.Fatalf("cookie value %q is not a UUID", sid)
	}
	if seen != sid {
		t.Errorf("ctx session %q != cookie %q", seen, sid)
	}
}

func TestMiddlewareReusesExistingCookie(t *testing.T) {
	tr := newTestTracker(t)
	want := newUUID()
	h := tr.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("GET", "/decks", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: want})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	for _, c := range rr.Result().Cookies() {
		if c.Name == CookieName && c.Value != want {
			t.Errorf("middleware overwrote existing cookie: got %q, want %q", c.Value, want)
		}
	}
}

func TestRecordEventAndJourney(t *testing.T) {
	tr := newTestTracker(t)
	ctx := context.Background()
	sid := newUUID()
	// Seed the session row (Middleware does this; for unit test we do it directly).
	if _, err := tr.db.ExecContext(ctx,
		`INSERT INTO pincer_session (id, first_seen_at, last_seen_at) VALUES (?, ?, ?)`,
		sid, time.Now().Unix(), time.Now().Unix()); err != nil {
		t.Fatalf("seed: %v", err)
	}

	tr.recordEvent(sid, EventDeckViewed, "/decks/foo/bar", "")
	tr.recordEvent(sid, EventDeckImported, "/api/decks", `{"source":"moxfield"}`)
	if err := tr.Stitch(ctx, sid, "user-42"); err != nil {
		t.Fatalf("stitch: %v", err)
	}

	journey, err := tr.UserJourney(ctx, "user-42", 100)
	if err != nil {
		t.Fatalf("journey: %v", err)
	}
	if len(journey) < 3 {
		t.Fatalf("expected ≥3 events (deck_viewed, deck_imported, login); got %d", len(journey))
	}
	// Stitch records a login event — verify it's there.
	var foundLogin bool
	for _, e := range journey {
		if e.EventType == EventLogin {
			foundLogin = true
		}
	}
	if !foundLogin {
		t.Error("Stitch did not record a login event")
	}
}

func TestFunnelStats(t *testing.T) {
	tr := newTestTracker(t)
	ctx := context.Background()
	now := time.Now().Unix()

	// Three sessions: two anon, one authed.
	for i, uid := range []string{"", "", "user-1"} {
		sid := newUUID()
		var userArg any
		if uid != "" {
			userArg = uid
		}
		if _, err := tr.db.ExecContext(ctx,
			`INSERT INTO pincer_session (id, user_id, first_seen_at, last_seen_at) VALUES (?, ?, ?, ?)`,
			sid, userArg, now-int64(i), now); err != nil {
			t.Fatalf("seed session: %v", err)
		}
		tr.recordEvent(sid, EventDeckViewed, "/decks/x/y", "")
	}

	f, err := tr.FunnelStats(ctx, time.Hour)
	if err != nil {
		t.Fatalf("funnel: %v", err)
	}
	if f.TotalSessions != 3 {
		t.Errorf("total: got %d, want 3", f.TotalSessions)
	}
	if f.AnonSessions != 2 || f.AuthSessions != 1 {
		t.Errorf("anon/auth: got %d/%d, want 2/1", f.AnonSessions, f.AuthSessions)
	}
	if f.DeckViews != 3 {
		t.Errorf("deck_views: got %d, want 3", f.DeckViews)
	}
}

func TestAnalyticsEndpointAdminAuth(t *testing.T) {
	tr := newTestTracker(t)
	mux := http.NewServeMux()
	tr.Register(mux)

	cases := []struct {
		name       string
		header     string
		query      string
		wantStatus int
	}{
		{"no auth", "", "", http.StatusUnauthorized},
		{"wrong token", "Bearer nope", "", http.StatusUnauthorized},
		{"correct bearer", "Bearer secret-admin", "", http.StatusOK},
		{"correct query token", "", "?token=secret-admin", http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/analytics/sessions"+tc.query, nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)
			if rr.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d (body: %s)", rr.Code, tc.wantStatus, rr.Body.String())
			}
		})
	}
}

func TestStitchEndpoint(t *testing.T) {
	tr := newTestTracker(t)
	mux := http.NewServeMux()
	tr.Register(mux)
	handler := tr.Middleware(mux)

	// First request — middleware issues a cookie.
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, httptest.NewRequest("GET", "/", nil))
	var sid string
	for _, c := range rr1.Result().Cookies() {
		if c.Name == CookieName {
			sid = c.Value
		}
	}
	if sid == "" {
		t.Fatal("no cookie issued")
	}

	// Second request — stitch with the cookie attached.
	body := strings.NewReader(`{"user_id":"user-99"}`)
	req := httptest.NewRequest("POST", "/api/pincer/stitch", body)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: sid})
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusOK {
		t.Fatalf("stitch endpoint: got %d, want 200 (body: %s)", rr2.Code, rr2.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rr2.Body.Bytes(), &resp)
	if resp["user_id"] != "user-99" {
		t.Errorf("response missing user_id: %v", resp)
	}

	// Verify the row updated.
	var stored string
	if err := tr.db.QueryRow(`SELECT user_id FROM pincer_session WHERE id = ?`, sid).Scan(&stored); err != nil {
		t.Fatalf("query: %v", err)
	}
	if stored != "user-99" {
		t.Errorf("user_id not stitched: got %q, want %q", stored, "user-99")
	}
}

func TestClassifyRequest(t *testing.T) {
	cases := []struct {
		method, path, want string
	}{
		{"GET", "/decks/josh/gitrog", EventDeckViewed},
		{"GET", "/api/decks/josh/gitrog", EventDeckViewed},
		{"GET", "/decks", ""},        // listing, not viewing
		{"GET", "/api/live/game", EventGameWatched},
		{"POST", "/api/decks", EventDeckImported},
		{"POST", "/api/import/moxfield", EventDeckImported},
		{"GET", "/api/profile", ""},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		got := classifyRequest(req)
		if got != tc.want {
			t.Errorf("%s %s: got %q, want %q", tc.method, tc.path, got, tc.want)
		}
	}
}
