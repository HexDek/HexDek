package credits

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	// File-backed + WAL + busy_timeout mirrors the production
	// hexdek-server DSN (internal/db/sqlite.go) so the concurrency
	// tests actually exercise the locking model real callers see.
	dsn := filepath.Join(dir, "credits.db") +
		"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(4)
	t.Cleanup(func() { db.Close() })
	if err := EnsureSchema(context.Background(), db); err != nil {
		t.Fatalf("schema: %v", err)
	}
	return New(db)
}

func TestEarnAndBalance(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if bal, _ := s.GetBalance(ctx, "alice"); bal != 0 {
		t.Errorf("starting balance = %d, want 0", bal)
	}

	bal, err := s.Earn(ctx, "alice", 50, ReasonComputeContrib, "game:1234")
	if err != nil {
		t.Fatal(err)
	}
	if bal != 50 {
		t.Errorf("after earn: bal=%d, want 50", bal)
	}

	bal, err = s.Earn(ctx, "alice", 25, ReasonComputeContrib, "game:1235")
	if err != nil {
		t.Fatal(err)
	}
	if bal != 75 {
		t.Errorf("cumulative earn: bal=%d, want 75", bal)
	}
}

func TestSpend(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if _, err := s.Earn(ctx, "alice", 100, ReasonComputeContrib, ""); err != nil {
		t.Fatal(err)
	}
	bal, err := s.Spend(ctx, "alice", 30, ReasonGauntletRun, "alice/krenko")
	if err != nil {
		t.Fatal(err)
	}
	if bal != 70 {
		t.Errorf("after spend: bal=%d, want 70", bal)
	}
}

func TestSpendInsufficient(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if _, err := s.Earn(ctx, "alice", 5, ReasonComputeContrib, ""); err != nil {
		t.Fatal(err)
	}

	_, err := s.Spend(ctx, "alice", 50, ReasonGauntletRun, "alice/krenko")
	if err == nil {
		t.Fatal("expected ErrInsufficientCredits")
	}
	if err != ErrInsufficientCredits {
		t.Errorf("err = %v, want ErrInsufficientCredits", err)
	}

	// Balance must be unchanged after a rejected spend.
	if bal, _ := s.GetBalance(ctx, "alice"); bal != 5 {
		t.Errorf("balance after rejected spend = %d, want 5", bal)
	}
}

func TestSpendNegativeRejected(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Spend(context.Background(), "alice", -10, "x", ""); err != ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
	if _, err := s.Spend(context.Background(), "alice", 0, "x", ""); err != ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount on zero, got %v", err)
	}
}

func TestEarnNegativeRejected(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Earn(context.Background(), "alice", -1, "x", ""); err != ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestHistoryOrdering(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		if _, err := s.Earn(ctx, "alice", 10, ReasonComputeContrib, "g"); err != nil {
			t.Fatal(err)
		}
		// Ensure created_at differs at the second granularity. Without
		// this the test depends on AUTOINCREMENT id ordering, which we
		// also do — the ORDER BY uses id as tiebreaker.
		time.Sleep(2 * time.Millisecond)
	}
	if _, err := s.Spend(ctx, "alice", 5, ReasonGauntletRun, "alice/k"); err != nil {
		t.Fatal(err)
	}

	txns, err := s.History(ctx, "alice", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(txns) != 6 {
		t.Fatalf("history len=%d, want 6", len(txns))
	}
	if txns[0].Amount != -5 {
		t.Errorf("most recent txn should be the spend; got amount=%d", txns[0].Amount)
	}
	if txns[0].Balance != 45 {
		t.Errorf("most recent balance = %d, want 45", txns[0].Balance)
	}
}

// TestSpendIsConcurrencySafe spawns N goroutines all trying to spend
// from the same balance. Only the spends that fit should succeed; the
// final balance must equal initial - successful_spends * cost. A naive
// "read then write" would let two spenders both see the same balance
// and both succeed, yielding a negative balance.
func TestSpendIsConcurrencySafe(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	const initial = 100
	const cost = 10
	const goroutines = 30 // 30 spends × 10 = 300 > 100, so most must fail

	if _, err := s.Earn(ctx, "alice", initial, ReasonComputeContrib, ""); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	var success int64
	var insufficient int64
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.Spend(ctx, "alice", cost, ReasonGauntletRun, "concurrent")
			if err == nil {
				atomic.AddInt64(&success, 1)
			} else if err == ErrInsufficientCredits {
				atomic.AddInt64(&insufficient, 1)
			} else {
				t.Errorf("unexpected err: %v", err)
			}
		}()
	}
	wg.Wait()

	if int64(initial)-success*cost < 0 {
		t.Fatalf("overspent: initial=%d, successes=%d × cost=%d", initial, success, cost)
	}
	if success != initial/cost {
		t.Errorf("expected exactly %d successful spends, got %d (insufficient=%d)",
			initial/cost, success, insufficient)
	}
	finalBal, _ := s.GetBalance(ctx, "alice")
	wantBal := int64(initial) - success*cost
	if finalBal != wantBal {
		t.Errorf("final balance = %d, want %d", finalBal, wantBal)
	}

	// Audit: the ledger must hold (1 earn + N successes + M failures-NOT-recorded)
	// and the latest balance row must equal the latest transaction's balance field.
	txns, _ := s.History(ctx, "alice", 200)
	if int64(len(txns)) != 1+success {
		t.Errorf("ledger row count = %d, want %d (1 earn + %d successful spends)",
			len(txns), 1+success, success)
	}
	if len(txns) > 0 && txns[0].Balance != finalBal {
		t.Errorf("ledger head balance %d != live balance %d",
			txns[0].Balance, finalBal)
	}
}

func TestQuotaState(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	q, err := s.QuotaState(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if q.UsedToday != 0 {
		t.Errorf("UsedToday = %d, want 0", q.UsedToday)
	}
	if q.FreeRemaining != FreeGauntletsPerDay {
		t.Errorf("FreeRemaining = %d, want %d", q.FreeRemaining, FreeGauntletsPerDay)
	}
	if !q.CanRunFree {
		t.Error("CanRunFree should be true with empty usage log")
	}
	if q.CanRunPaid {
		t.Error("CanRunPaid should be false with zero balance")
	}

	// Burn the free quota.
	for i := 0; i < FreeGauntletsPerDay; i++ {
		if err := s.LogGauntlet(ctx, "alice", "alice/krenko", 500, true, 0); err != nil {
			t.Fatal(err)
		}
	}
	q, _ = s.QuotaState(ctx, "alice")
	if q.CanRunFree {
		t.Error("CanRunFree should be false after burning the quota")
	}
	if q.FreeRemaining != 0 {
		t.Errorf("FreeRemaining = %d, want 0", q.FreeRemaining)
	}

	// Top up so paid is available.
	if _, err := s.Earn(ctx, "alice", 50, ReasonComputeContrib, ""); err != nil {
		t.Fatal(err)
	}
	q, _ = s.QuotaState(ctx, "alice")
	if !q.CanRunPaid {
		t.Error("CanRunPaid should be true with sufficient balance")
	}
}

// ---------------------------------------------------------------------------
// HTTP handler tests
// ---------------------------------------------------------------------------

func TestHandlerBalanceRequiresAuth(t *testing.T) {
	s := newTestStore(t)
	mux := http.NewServeMux()
	s.Register(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/credits", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("no auth: code=%d, want 401", rec.Code)
	}
}

func TestHandlerBalanceReturnsJSON(t *testing.T) {
	s := newTestStore(t)
	mux := http.NewServeMux()
	s.Register(mux)
	if _, err := s.Earn(context.Background(), "alice", 42, ReasonComputeContrib, ""); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/api/credits", nil)
	req.Header.Set("X-HexDek-Owner", "alice")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	var bal Balance
	if err := json.Unmarshal(rec.Body.Bytes(), &bal); err != nil {
		t.Fatal(err)
	}
	if bal.Credits != 42 || bal.Owner != "alice" {
		t.Errorf("got %+v, want credits=42 owner=alice", bal)
	}
}

func TestHandlerSpendInsufficientReturns402(t *testing.T) {
	s := newTestStore(t)
	mux := http.NewServeMux()
	s.Register(mux)

	body := strings.NewReader(`{"amount": 50, "reason": "extended_analysis"}`)
	req := httptest.NewRequest("POST", "/api/credits/spend", body)
	req.Header.Set("X-HexDek-Owner", "alice")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusPaymentRequired {
		t.Errorf("insufficient: code=%d, want 402", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "insufficient_credits") {
		t.Errorf("body missing error code: %s", rec.Body.String())
	}
}

func TestHandlerSpendHappyPath(t *testing.T) {
	s := newTestStore(t)
	mux := http.NewServeMux()
	s.Register(mux)
	if _, err := s.Earn(context.Background(), "alice", 100, ReasonComputeContrib, ""); err != nil {
		t.Fatal(err)
	}

	body := strings.NewReader(`{"amount": 30, "reason": "extended_analysis", "reference": "alice/krenko"}`)
	req := httptest.NewRequest("POST", "/api/credits/spend", body)
	req.Header.Set("X-HexDek-Owner", "alice")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	var bal Balance
	_ = json.Unmarshal(rec.Body.Bytes(), &bal)
	if bal.Credits != 70 {
		t.Errorf("post-spend balance = %d, want 70", bal.Credits)
	}
}

func TestHandlerHistoryReturnsArray(t *testing.T) {
	s := newTestStore(t)
	mux := http.NewServeMux()
	s.Register(mux)
	ctx := context.Background()
	if _, err := s.Earn(ctx, "alice", 10, ReasonComputeContrib, "g1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Spend(ctx, "alice", 5, ReasonGauntletRun, "alice/k"); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/api/credits/history", nil)
	req.Header.Set("X-HexDek-Owner", "alice")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Owner        string        `json:"owner"`
		Transactions []Transaction `json:"transactions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Transactions) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(resp.Transactions))
	}
}

func TestHandlerSpendValidatesInput(t *testing.T) {
	s := newTestStore(t)
	mux := http.NewServeMux()
	s.Register(mux)

	cases := []struct {
		body string
		want int
	}{
		{`{"amount": 0, "reason": "x"}`, 400},
		{`{"amount": -5, "reason": "x"}`, 400},
		{`{"amount": 10, "reason": ""}`, 400},
		{`not json`, 400},
	}
	for _, tc := range cases {
		req := httptest.NewRequest("POST", "/api/credits/spend", strings.NewReader(tc.body))
		req.Header.Set("X-HexDek-Owner", "alice")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Errorf("body %q → code=%d, want %d", tc.body, rec.Code, tc.want)
		}
	}
}
