package hexapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hexdek/hexdek/internal/db"
)

func newAdminTestHandler(t *testing.T) (*AdminAnticheatHandler, *http.ServeMux) {
	t.Helper()
	d, err := db.Open(t.TempDir() + "/admin_anticheat.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { d.Close() })

	h := &AdminAnticheatHandler{
		DB:         d,
		AdminToken: "test-admin-token",
	}
	mux := http.NewServeMux()
	h.Register(mux)
	return h, mux
}

func TestAdminVerifications_RequiresToken(t *testing.T) {
	_, mux := newAdminTestHandler(t)
	req := httptest.NewRequest("GET", "/api/admin/verifications", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAdminVerifications_RejectsWrongToken(t *testing.T) {
	_, mux := newAdminTestHandler(t)
	req := httptest.NewRequest("GET", "/api/admin/verifications", nil)
	req.Header.Set("Authorization", "Bearer not-the-real-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAdminVerifications_AcceptsCorrectToken(t *testing.T) {
	h, mux := newAdminTestHandler(t)

	// Seed one row.
	_, err := db.EnqueueVerification(context.Background(), h.DB, db.VerificationEnqueueParams{
		GameID:        100,
		DeckKey:       "alice/aggro",
		RNGSeed:       42,
		NSeats:        4,
		DeckKeys:      []string{"alice/aggro", "bob/control", "carol/combo", "dave/midrange"},
		ClaimedWinner: 1,
		ClaimedTurns:  18,
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/admin/verifications", nil)
	req.Header.Set("Authorization", "Bearer test-admin-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var body struct {
		Stats db.VerificationStats   `json:"stats"`
		Rows  []verificationViewRow  `json:"rows"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Stats.Pending != 1 {
		t.Errorf("expected pending=1, got %+v", body.Stats)
	}
	if len(body.Rows) != 1 || body.Rows[0].DeckKey != "alice/aggro" {
		t.Errorf("rows shape wrong: %+v", body.Rows)
	}
}

func TestAdminVerifications_FiltersByStatus(t *testing.T) {
	h, mux := newAdminTestHandler(t)

	// One pending, one passed.
	id, _ := db.EnqueueVerification(context.Background(), h.DB, db.VerificationEnqueueParams{
		GameID: 1, DeckKey: "alice/aggro", NSeats: 4,
		DeckKeys: []string{"alice/aggro", "x", "y", "z"},
	})
	db.ClaimNextVerification(context.Background(), h.DB)
	db.FinishVerification(context.Background(), h.DB, id, "passed", 0, 5, "")
	db.EnqueueVerification(context.Background(), h.DB, db.VerificationEnqueueParams{
		GameID: 2, DeckKey: "alice/aggro", NSeats: 4,
		DeckKeys: []string{"alice/aggro", "x", "y", "z"},
	})

	req := httptest.NewRequest("GET", "/api/admin/verifications?status=passed", nil)
	req.Header.Set("Authorization", "Bearer test-admin-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var body struct {
		Rows []verificationViewRow `json:"rows"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Rows) != 1 || body.Rows[0].Status != "passed" {
		t.Errorf("status filter not applied: %+v", body.Rows)
	}
}

func TestAdminSanctions_RequiresToken(t *testing.T) {
	_, mux := newAdminTestHandler(t)
	req := httptest.NewRequest("GET", "/api/admin/sanctions", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAdminSanctions_ListAll(t *testing.T) {
	h, mux := newAdminTestHandler(t)

	now := time.Now().Unix()
	db.InsertSanction(context.Background(), h.DB, db.SanctionInsertParams{
		DeckKey: "alice/aggro", Severity: db.SeverityWarning, IssuedAt: now, Reason: "first",
	})
	db.InsertSanction(context.Background(), h.DB, db.SanctionInsertParams{
		DeckKey: "alice/aggro", Severity: db.SeverityTempBan, IssuedAt: now, ExpiresAt: now + 86400, Reason: "second",
	})
	db.InsertSanction(context.Background(), h.DB, db.SanctionInsertParams{
		DeckKey: "mallory/cheater", Severity: db.SeverityPermanentBan, IssuedAt: now, Reason: "egregious",
	})

	req := httptest.NewRequest("GET", "/api/admin/sanctions", nil)
	req.Header.Set("Authorization", "Bearer test-admin-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body struct {
		Count int            `json:"count"`
		Rows  []sanctionView `json:"rows"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Count != 3 {
		t.Errorf("expected 3 rows, got %d", body.Count)
	}
}

func TestAdminSanctions_ActiveOnlyHidesWarnings(t *testing.T) {
	h, mux := newAdminTestHandler(t)

	now := time.Now().Unix()
	db.InsertSanction(context.Background(), h.DB, db.SanctionInsertParams{
		DeckKey: "alice/aggro", Severity: db.SeverityWarning, IssuedAt: now,
	})
	db.InsertSanction(context.Background(), h.DB, db.SanctionInsertParams{
		DeckKey: "bob/control", Severity: db.SeverityTempBan, IssuedAt: now, ExpiresAt: now + 86400,
	})

	req := httptest.NewRequest("GET", "/api/admin/sanctions?active=1", nil)
	req.Header.Set("Authorization", "Bearer test-admin-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	var body struct {
		Count int            `json:"count"`
		Rows  []sanctionView `json:"rows"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Count != 1 || body.Rows[0].DeckKey != "bob/control" {
		t.Errorf("active=1 not filtering: %+v", body.Rows)
	}
}

func TestAdminSanctions_PerDeckList(t *testing.T) {
	h, mux := newAdminTestHandler(t)

	now := time.Now().Unix()
	db.InsertSanction(context.Background(), h.DB, db.SanctionInsertParams{
		DeckKey: "alice/aggro", Severity: db.SeverityWarning, IssuedAt: now - 100,
	})
	db.InsertSanction(context.Background(), h.DB, db.SanctionInsertParams{
		DeckKey: "alice/aggro", Severity: db.SeverityTempBan, IssuedAt: now, ExpiresAt: now + 86400,
	})
	db.InsertSanction(context.Background(), h.DB, db.SanctionInsertParams{
		DeckKey: "bob/control", Severity: db.SeverityWarning, IssuedAt: now,
	})

	req := httptest.NewRequest("GET", "/api/admin/sanctions?deck_key=alice/aggro", nil)
	req.Header.Set("Authorization", "Bearer test-admin-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var body struct {
		Rows []sanctionView `json:"rows"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Rows) != 2 {
		t.Errorf("expected 2 rows for alice, got %d: %+v", len(body.Rows), body.Rows)
	}
	for _, r := range body.Rows {
		if r.DeckKey != "alice/aggro" {
			t.Errorf("non-alice row leaked: %+v", r)
		}
	}
}

func TestAdminAnticheatStats(t *testing.T) {
	h, mux := newAdminTestHandler(t)

	now := time.Now().Unix()
	db.InsertSanction(context.Background(), h.DB, db.SanctionInsertParams{
		DeckKey: "bob/control", Severity: db.SeverityTempBan, IssuedAt: now, ExpiresAt: now + 86400,
	})
	db.EnqueueVerification(context.Background(), h.DB, db.VerificationEnqueueParams{
		GameID: 1, DeckKey: "x/y", NSeats: 4,
		DeckKeys: []string{"x/y", "a", "b", "c"},
	})

	req := httptest.NewRequest("GET", "/api/admin/anticheat/stats", nil)
	req.Header.Set("Authorization", "Bearer test-admin-token")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body struct {
		Queue      db.VerificationStats `json:"queue"`
		ActiveBans int                  `json:"active_bans"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Queue.Pending != 1 {
		t.Errorf("expected queue.pending=1, got %+v", body.Queue)
	}
	if body.ActiveBans != 1 {
		t.Errorf("expected active_bans=1, got %d", body.ActiveBans)
	}
}

func TestAdmin_FailClosedWhenNoTokenConfigured(t *testing.T) {
	d, err := db.Open(t.TempDir() + "/admin_no_token.db")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { d.Close() })

	h := &AdminAnticheatHandler{DB: d, AdminToken: ""}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest("GET", "/api/admin/verifications", nil)
	req.Header.Set("Authorization", "Bearer anything")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with empty AdminToken, got %d", w.Code)
	}
}
