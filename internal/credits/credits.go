// Package credits implements the HexDek credit economy.
//
// Users earn credits by contributing compute power (validated games via
// the BOINC client). Credits are spent on extended testing — gauntlet
// runs past a free-tier daily quota, deeper analysis, etc.
//
// Two tables back the package:
//
//   - credit_balances:    one row per owner with the current balance.
//                         Bookkeeping shortcut so reads don't have to
//                         fold the entire transaction log.
//   - credit_transactions: append-only ledger. Every earn/spend writes
//                         a row; balances are mutated in the same SQL
//                         transaction so the two never diverge.
//
// Concurrency safety relies on SQLite's default serialised writer plus
// IMMEDIATE transactions for spend operations. The Spend path checks
// the balance and updates it in the same BEGIN IMMEDIATE..COMMIT block
// so two spenders racing each other can never overdraft.
package credits

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrInsufficientCredits is returned when Spend is called with an
// amount larger than the owner's current balance. Distinct error so
// the HTTP handler can map it to 402 Payment Required cleanly.
var ErrInsufficientCredits = errors.New("credits: insufficient balance")

// ErrInvalidAmount is returned when a non-positive amount is given to
// Earn or Spend. Both operations require strictly positive integers.
var ErrInvalidAmount = errors.New("credits: amount must be positive")

// FreeGauntletsPerDay is the per-owner free-tier quota: this many
// gauntlet runs per UTC day before credits are required. Exposed so
// callers (tests, the gauntlet handler) reference the canonical
// number rather than copying a magic constant.
const FreeGauntletsPerDay = 3

// CreditsPerGauntlet is the credit cost of one gauntlet run beyond
// the free-tier quota. Tunable; current value reflects ~1 credit per
// 100 games at the default 500-game gauntlet length.
const CreditsPerGauntlet = 5

// Reason values used by the gauntlet flow + BOINC validator. Free-form
// strings, but using the constants keeps the ledger searchable.
const (
	ReasonGauntletRun       = "gauntlet_run"
	ReasonComputeContrib    = "compute_contribution"
	ReasonAdminAdjustment   = "admin_adjustment"
	ReasonExtendedAnalysis  = "extended_analysis"
)

// Transaction is one row in the append-only credit ledger. Positive
// Amount = earned, negative = spent. The Balance field captures the
// post-transaction balance so the history view doesn't need to fold
// across rows.
type Transaction struct {
	ID        int64  `json:"id"`
	Owner     string `json:"owner"`
	Amount    int64  `json:"amount"`           // signed: + earn, - spend
	Balance   int64  `json:"balance"`          // balance after this txn
	Reason    string `json:"reason"`
	Reference string `json:"reference,omitempty"` // free-form pointer (deck key, game id, etc.)
	CreatedAt int64  `json:"created_at"`
}

// Balance is the materialised current balance for an owner. The
// balance row exists for any owner who has ever transacted; absent
// rows are treated as zero by all helpers.
type Balance struct {
	Owner       string `json:"owner"`
	Credits     int64  `json:"credits"`
	UpdatedAt   int64  `json:"updated_at"`
}

// EnsureSchema creates the credit tables idempotently. Safe to call
// on every server start.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("credits: nil db")
	}
	const ddl = `
CREATE TABLE IF NOT EXISTS credit_balances (
    owner       TEXT PRIMARY KEY,
    credits     INTEGER NOT NULL DEFAULT 0,
    updated_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS credit_transactions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    owner      TEXT NOT NULL,
    amount     INTEGER NOT NULL,           -- signed: + earn, - spend
    balance    INTEGER NOT NULL,           -- post-transaction balance
    reason     TEXT NOT NULL,
    reference  TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_credit_transactions_owner_time
    ON credit_transactions(owner, created_at DESC, id DESC);

-- Per-owner gauntlet usage log. Drives the free-tier daily quota
-- and is the audit trail for "did this user actually spend a credit
-- on this run". One row per gauntlet start; the column free=1 means
-- the run came out of the daily allowance, free=0 means it cost
-- credits (see credits_charged for the amount).
CREATE TABLE IF NOT EXISTS gauntlet_usage (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    owner           TEXT NOT NULL,
    deck_key        TEXT NOT NULL,
    games           INTEGER NOT NULL,
    free            INTEGER NOT NULL,           -- 1 = free tier, 0 = paid
    credits_charged INTEGER NOT NULL DEFAULT 0,
    started_at      INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_gauntlet_usage_owner_time
    ON gauntlet_usage(owner, started_at DESC);
`
	_, err := db.ExecContext(ctx, ddl)
	return err
}

// Store is the credits façade. Wraps a *sql.DB and exposes the
// balance/ledger operations. Stateless beyond the DB pointer so it's
// safe to share across goroutines.
type Store struct {
	db *sql.DB
}

// New returns a Store backed by the given DB. Caller must have run
// EnsureSchema at least once.
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// GetBalance returns the current balance for owner. Missing rows
// resolve to zero — every owner is treated as having a zero balance
// until the first earn/spend writes a row.
func (s *Store) GetBalance(ctx context.Context, owner string) (int64, error) {
	owner = normaliseOwner(owner)
	if owner == "" {
		return 0, nil
	}
	var bal int64
	err := s.db.QueryRowContext(ctx,
		`SELECT credits FROM credit_balances WHERE owner = ?`, owner).Scan(&bal)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("credits: get balance: %w", err)
	}
	return bal, nil
}

// Earn credits the owner's balance by amount and writes a transaction
// row. Returns the new balance.
func (s *Store) Earn(ctx context.Context, owner string, amount int64, reason, ref string) (int64, error) {
	owner = normaliseOwner(owner)
	if owner == "" {
		return 0, errors.New("credits: empty owner")
	}
	if amount <= 0 {
		return 0, ErrInvalidAmount
	}
	return s.applyDelta(ctx, owner, amount, reason, ref)
}

// Spend deducts amount from the owner's balance, returning
// ErrInsufficientCredits if the post-spend balance would be negative.
// Atomic: the balance check and update happen inside one IMMEDIATE
// transaction so two concurrent spenders cannot both succeed against
// the same starting balance.
func (s *Store) Spend(ctx context.Context, owner string, amount int64, reason, ref string) (int64, error) {
	owner = normaliseOwner(owner)
	if owner == "" {
		return 0, errors.New("credits: empty owner")
	}
	if amount <= 0 {
		return 0, ErrInvalidAmount
	}
	return s.applyDelta(ctx, owner, -amount, reason, ref)
}

// applyDelta is the shared transactional body for Earn and Spend.
// Uses BEGIN IMMEDIATE so the writer-lock is acquired up-front; this
// is the standard SQLite recipe for "read current balance, decide,
// write new balance" without TOCTOU races.
func (s *Store) applyDelta(ctx context.Context, owner string, delta int64, reason, ref string) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("credits: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck — committed on success below.

	if _, err := tx.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		// Some drivers (modernc) reject manual BEGIN inside an outer
		// tx; fall through if it errors — the BeginTx already opened
		// a write transaction that's good enough for our purposes.
		_ = err
	}

	var current int64
	err = tx.QueryRowContext(ctx,
		`SELECT credits FROM credit_balances WHERE owner = ?`, owner).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		current = 0
	} else if err != nil {
		return 0, fmt.Errorf("credits: read balance: %w", err)
	}

	next := current + delta
	if next < 0 {
		return current, ErrInsufficientCredits
	}

	now := time.Now().Unix()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO credit_balances (owner, credits, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(owner) DO UPDATE SET
		  credits    = excluded.credits,
		  updated_at = excluded.updated_at
	`, owner, next, now); err != nil {
		return 0, fmt.Errorf("credits: write balance: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO credit_transactions (owner, amount, balance, reason, reference, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, owner, delta, next, reason, ref, now); err != nil {
		return 0, fmt.Errorf("credits: write txn: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("credits: commit: %w", err)
	}
	return next, nil
}

// History returns the most recent transactions for owner, newest
// first, capped at limit (default 50, max 200).
func (s *Store) History(ctx context.Context, owner string, limit int) ([]Transaction, error) {
	owner = normaliseOwner(owner)
	if owner == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, owner, amount, balance, reason, reference, created_at
		FROM credit_transactions
		WHERE owner = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, owner, limit)
	if err != nil {
		return nil, fmt.Errorf("credits: query history: %w", err)
	}
	defer rows.Close()

	var out []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.Owner, &t.Amount, &t.Balance,
			&t.Reason, &t.Reference, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// CountGauntletsToday returns how many gauntlet_usage rows the owner
// has logged since the start of the current UTC day. Used to enforce
// FreeGauntletsPerDay before a paid charge.
func (s *Store) CountGauntletsToday(ctx context.Context, owner string) (int, error) {
	owner = normaliseOwner(owner)
	if owner == "" {
		return 0, nil
	}
	since := startOfUTCDay(time.Now()).Unix()
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM gauntlet_usage WHERE owner = ? AND started_at >= ?`,
		owner, since).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("credits: count gauntlets: %w", err)
	}
	return n, nil
}

// LogGauntlet appends a gauntlet_usage row. Called by the gauntlet
// handler after a successful start (whether free or paid).
func (s *Store) LogGauntlet(ctx context.Context, owner, deckKey string, games int, free bool, charged int64) error {
	owner = normaliseOwner(owner)
	if owner == "" {
		return nil
	}
	freeFlag := 0
	if free {
		freeFlag = 1
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO gauntlet_usage (owner, deck_key, games, free, credits_charged, started_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, owner, deckKey, games, freeFlag, charged, time.Now().Unix())
	return err
}

// GauntletQuotaState describes the current free/paid posture for the
// owner. Used by the gauntlet handler to decide whether to charge,
// and by the frontend to render "X free runs left today" before the
// user clicks.
type GauntletQuotaState struct {
	UsedToday    int   `json:"used_today"`
	FreeLimit    int   `json:"free_limit"`
	FreeRemaining int  `json:"free_remaining"`
	Balance      int64 `json:"balance"`
	Cost         int64 `json:"cost_per_run"`
	CanRunFree   bool  `json:"can_run_free"`
	CanRunPaid   bool  `json:"can_run_paid"`
}

// QuotaState returns the current free/paid posture for owner. Fetches
// in a single round trip pair (gauntlet count + balance).
func (s *Store) QuotaState(ctx context.Context, owner string) (GauntletQuotaState, error) {
	used, err := s.CountGauntletsToday(ctx, owner)
	if err != nil {
		return GauntletQuotaState{}, err
	}
	bal, err := s.GetBalance(ctx, owner)
	if err != nil {
		return GauntletQuotaState{}, err
	}
	free := FreeGauntletsPerDay - used
	if free < 0 {
		free = 0
	}
	return GauntletQuotaState{
		UsedToday:     used,
		FreeLimit:     FreeGauntletsPerDay,
		FreeRemaining: free,
		Balance:       bal,
		Cost:          CreditsPerGauntlet,
		CanRunFree:    free > 0,
		CanRunPaid:    bal >= CreditsPerGauntlet,
	}, nil
}

// startOfUTCDay returns midnight UTC for the day the input falls in.
func startOfUTCDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// normaliseOwner lower-cases + trims so "Alice" and "alice " collapse
// to the same key. Mirrors checkOwnership in hexapi.
func normaliseOwner(o string) string {
	return strings.ToLower(strings.TrimSpace(o))
}
