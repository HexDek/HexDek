// Package anticheat is the contributor-level statistical audit layer for
// HexDek's anti-cheat pipeline. It is the per-contributor extension of the
// per-deck Phase-1 detector in internal/hat/anomaly.go.
//
// "Contributor" in this package is whoever submits game results — today
// that maps to the deck owner (one human running self-play locally). When
// the BOINC distributed-compute client lands a contributor will become a
// per-machine identifier; the schema does not change because the rolling
// counters are keyed by an opaque string.
//
// The auditor persists two tables:
//
//   contributor_stats — one row per contributor with running counters
//     for games, wins, sum-of-turns, sum-of-squared-turns, last activity.
//     Sum-of-squared-turns lets us compute sample variance without
//     re-reading every game (Welford-equivalent for fixed-precision use).
//
//   contributor_flags — one row per anomaly raised, with the triggering
//     metric, the contributor's value, the population mean+stddev seen at
//     flag time, the resulting z-score, a severity score (1..5 escalating
//     on repeat flags for the same metric), and review state
//     (resolved_at / resolved_by / resolved_note).
//
// The auditor is detection-only. Nothing in this package takes punitive
// action on a flag; downstream tooling (admin UI, replay verification)
// is responsible for second-screen review.
package anticheat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"
)

// MinGames — minimum lifetime games a contributor must have submitted
// before they're eligible for the population z-score check. Below this
// threshold the win-rate sample size is too small to be meaningful.
// Matches internal/hat/anomaly.go MinGames so the two layers see decks
// the same way.
const MinGames = 30

// ZThreshold — flag at >|3.0| standard deviations from the population
// mean. ~0.27% of samples under a normal distribution.
const ZThreshold = 3.0

// MaxSeverity — clamp on flag severity. After this many repeats on the
// same (contributor, metric) the level stops climbing; the count is
// still tracked via the row count itself.
const MaxSeverity = 5

// Metric names. Stored verbatim in contributor_flags.metric, so changing
// these is a schema migration.
const (
	MetricWinRate       = "win_rate"
	MetricAvgGameLength = "avg_game_length"
	MetricTurnVariance  = "turn_variance"
)

// schemaSQL is applied once per Open(). All tables and indexes are
// IF NOT EXISTS so the auditor coexists cleanly with internal/db.
const schemaSQL = `
CREATE TABLE IF NOT EXISTS contributor_stats (
  contributor_id   TEXT PRIMARY KEY,
  games            INTEGER NOT NULL DEFAULT 0,
  wins             INTEGER NOT NULL DEFAULT 0,
  total_turns      INTEGER NOT NULL DEFAULT 0,
  total_turns_sq   REAL    NOT NULL DEFAULT 0.0,
  last_game_at     INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS contributor_flags (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  contributor_id  TEXT    NOT NULL,
  metric          TEXT    NOT NULL,
  metric_value    REAL    NOT NULL,
  pop_mean        REAL    NOT NULL,
  pop_stddev      REAL    NOT NULL,
  z_score         REAL    NOT NULL,
  severity        INTEGER NOT NULL DEFAULT 1,
  detected_at     INTEGER NOT NULL,
  resolved_at     INTEGER,
  resolved_by     TEXT,
  resolved_note   TEXT
);

CREATE INDEX IF NOT EXISTS idx_contributor_flags_active
  ON contributor_flags (contributor_id, metric, resolved_at);

CREATE INDEX IF NOT EXISTS idx_contributor_flags_detected
  ON contributor_flags (detected_at);
`

// StatisticalAuditor runs the contributor-level z-score check on the
// SQLite store every time a game result lands. Safe for concurrent
// callers (relies on SQLite's own locking).
type StatisticalAuditor struct {
	db    *sql.DB
	now   func() time.Time // injectable for tests
	minGames int
	zThreshold float64
}

// NewStatisticalAuditor opens the auditor against an existing *sql.DB
// (typically the one shared with internal/db). It applies the schema
// idempotently, so calling NewStatisticalAuditor on the same DB twice
// is harmless.
func NewStatisticalAuditor(db *sql.DB) (*StatisticalAuditor, error) {
	if db == nil {
		return nil, errors.New("anticheat: nil *sql.DB")
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("anticheat: apply schema: %w", err)
	}
	return &StatisticalAuditor{
		db:         db,
		now:        time.Now,
		minGames:   MinGames,
		zThreshold: ZThreshold,
	}, nil
}

// SetThresholds overrides the eligibility and z thresholds. Used by
// tests to drive the audit with smaller populations; production code
// should leave the defaults alone.
func (a *StatisticalAuditor) SetThresholds(minGames int, z float64) {
	if minGames > 0 {
		a.minGames = minGames
	}
	if z > 0 {
		a.zThreshold = z
	}
}

// Game is the result of one game from one contributor. Turns is the
// total turn count (game length). Won is the contributor's outcome:
// true for a win, false for loss/draw. Both fields are necessary inputs
// to the audit dimensions.
type Game struct {
	ContributorID string
	Won           bool
	Turns         int
}

// RecordGame upserts the contributor's running stats and runs the
// audit. Returned flags are the new flags created by this call (zero
// or more — one game can flag on multiple dimensions). On error the
// stats may have been written but the audit was skipped.
func (a *StatisticalAuditor) RecordGame(ctx context.Context, g Game) ([]Flag, error) {
	if a == nil {
		return nil, errors.New("anticheat: nil auditor")
	}
	if g.ContributorID == "" {
		return nil, errors.New("anticheat: empty contributor_id")
	}
	if g.Turns < 0 {
		g.Turns = 0
	}

	// Upsert running counters. INSERT OR IGNORE creates the row, then
	// the UPDATE applies in one round-trip even on the first call.
	wonInt := 0
	if g.Won {
		wonInt = 1
	}
	turnsSq := float64(g.Turns) * float64(g.Turns)
	now := a.now().Unix()

	if _, err := a.db.ExecContext(ctx, `
		INSERT INTO contributor_stats (contributor_id, games, wins, total_turns, total_turns_sq, last_game_at)
		VALUES (?, 1, ?, ?, ?, ?)
		ON CONFLICT(contributor_id) DO UPDATE SET
		  games          = games + 1,
		  wins           = wins + excluded.wins,
		  total_turns    = total_turns + excluded.total_turns,
		  total_turns_sq = total_turns_sq + excluded.total_turns_sq,
		  last_game_at   = excluded.last_game_at
	`, g.ContributorID, wonInt, g.Turns, turnsSq, now); err != nil {
		return nil, fmt.Errorf("anticheat: upsert stats: %w", err)
	}

	return a.audit(ctx, g.ContributorID)
}

// ContributorStats is a read-only snapshot of one contributor's
// running counters.
type ContributorStats struct {
	ContributorID string
	Games         int
	Wins          int
	TotalTurns    int64
	TotalTurnsSq  float64
	LastGameAt    int64
}

// WinRate returns the lifetime win rate, or 0 when no games recorded.
func (s ContributorStats) WinRate() float64 {
	if s.Games == 0 {
		return 0
	}
	return float64(s.Wins) / float64(s.Games)
}

// AvgGameLength returns the mean turn count per game.
func (s ContributorStats) AvgGameLength() float64 {
	if s.Games == 0 {
		return 0
	}
	return float64(s.TotalTurns) / float64(s.Games)
}

// TurnVariance returns the sample variance (Bessel-corrected, n-1)
// of game length. Zero when fewer than 2 games are recorded.
func (s ContributorStats) TurnVariance() float64 {
	if s.Games < 2 {
		return 0
	}
	mean := s.AvgGameLength()
	// Var = (sum_sq - n * mean^2) / (n - 1).
	v := (s.TotalTurnsSq - float64(s.Games)*mean*mean) / float64(s.Games-1)
	if v < 0 {
		// Floating-point noise can drive a true-zero variance slightly
		// negative; clamp to 0 so downstream sqrt/stats stay sane.
		return 0
	}
	return v
}

// GetStats returns a single contributor's stats. ok=false when the
// contributor has not submitted any games yet.
func (a *StatisticalAuditor) GetStats(ctx context.Context, id string) (ContributorStats, bool, error) {
	row := a.db.QueryRowContext(ctx, `
		SELECT contributor_id, games, wins, total_turns, total_turns_sq, last_game_at
		FROM contributor_stats WHERE contributor_id = ?`, id)
	var s ContributorStats
	if err := row.Scan(&s.ContributorID, &s.Games, &s.Wins, &s.TotalTurns, &s.TotalTurnsSq, &s.LastGameAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ContributorStats{}, false, nil
		}
		return ContributorStats{}, false, err
	}
	return s, true, nil
}

// Flag mirrors a contributor_flags row.
type Flag struct {
	ID            int64
	ContributorID string
	Metric        string
	MetricValue   float64
	PopMean       float64
	PopStdDev     float64
	ZScore        float64
	Severity      int
	DetectedAt    time.Time
	ResolvedAt    *time.Time
	ResolvedBy    string
	ResolvedNote  string
}

// audit runs the population-level z-score check across every audit
// dimension for one contributor. New flags are written to
// contributor_flags and returned.
func (a *StatisticalAuditor) audit(ctx context.Context, id string) ([]Flag, error) {
	self, ok, err := a.GetStats(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("anticheat: load self stats: %w", err)
	}
	if !ok || self.Games < a.minGames {
		return nil, nil
	}

	dims := []audit{
		{metric: MetricWinRate, value: self.WinRate(), valueExpr: "(CAST(wins AS REAL) / games)"},
		{metric: MetricAvgGameLength, value: self.AvgGameLength(), valueExpr: "(CAST(total_turns AS REAL) / games)"},
		{
			metric: MetricTurnVariance, value: self.TurnVariance(),
			// Sample variance per row, excluding rows with games < 2.
			// Matches ContributorStats.TurnVariance exactly.
			valueExpr: "((total_turns_sq - games * (CAST(total_turns AS REAL) / games) * (CAST(total_turns AS REAL) / games)) / (games - 1))",
			minRowGames: 2,
		},
	}

	var flags []Flag
	for _, d := range dims {
		f, err := a.checkDimension(ctx, id, d)
		if err != nil {
			return flags, err
		}
		if f != nil {
			flags = append(flags, *f)
		}
	}
	return flags, nil
}

type audit struct {
	metric      string
	value       float64
	valueExpr   string // SQL expression yielding the metric for one contributor row
	minRowGames int    // minimum games for a row to enter the population (default minGames)
}

// checkDimension computes the population mean+stddev on the chosen
// metric over every eligible contributor *except* the one under test
// (leave-one-out — otherwise a single extreme outlier inflates the
// stddev and masks itself). Flags are written when |z| ≥ ZThreshold.
func (a *StatisticalAuditor) checkDimension(ctx context.Context, id string, d audit) (*Flag, error) {
	rowMin := d.minRowGames
	if rowMin < a.minGames {
		rowMin = a.minGames
	}

	// Pull every eligible peer's metric value in one query. We then
	// compute the mean+stddev in Go because SQLite's STDEV is an
	// extension we may not have, and a few hundred rows is trivial in
	// memory.
	rows, err := a.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT %s
		FROM contributor_stats
		WHERE games >= ? AND contributor_id != ?`, d.valueExpr),
		rowMin, id)
	if err != nil {
		return nil, fmt.Errorf("anticheat: load population for %s: %w", d.metric, err)
	}
	defer rows.Close()

	var values []float64
	for rows.Next() {
		var v float64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(values) < 2 {
		// Need at least 2 peers for a meaningful sample stddev.
		return nil, nil
	}

	mean, sd := meanStdDev(values)
	if sd <= 0 {
		// Every peer identical — nothing to compare against.
		return nil, nil
	}
	z := (d.value - mean) / sd
	if math.Abs(z) < a.zThreshold {
		return nil, nil
	}

	// Flag escalation: at most one active row per (contributor, metric).
	// If the contributor is still flagged on this metric from a prior
	// game, bump severity in place and refresh the snapshot fields —
	// this both matches the spec ("repeated flags escalate severity")
	// and prevents flag-flood when the audit runs on every game while
	// an extreme contributor's metric continues to drift far from the
	// population mean. Once an admin resolves a flag, the next flag on
	// the same metric starts fresh at severity 1.
	now := a.now().Unix()
	var existingID int64
	var existingSeverity int
	err = a.db.QueryRowContext(ctx, `
		SELECT id, severity FROM contributor_flags
		WHERE contributor_id = ? AND metric = ? AND resolved_at IS NULL
		LIMIT 1`,
		id, d.metric).Scan(&existingID, &existingSeverity)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		// First active flag for this metric — insert at severity 1.
		res, err := a.db.ExecContext(ctx, `
			INSERT INTO contributor_flags
			  (contributor_id, metric, metric_value, pop_mean, pop_stddev, z_score, severity, detected_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			id, d.metric, d.value, mean, sd, z, 1, now)
		if err != nil {
			return nil, fmt.Errorf("anticheat: insert flag: %w", err)
		}
		newID, _ := res.LastInsertId()
		return &Flag{
			ID:            newID,
			ContributorID: id,
			Metric:        d.metric,
			MetricValue:   d.value,
			PopMean:       mean,
			PopStdDev:     sd,
			ZScore:        z,
			Severity:      1,
			DetectedAt:    time.Unix(now, 0),
		}, nil
	case err != nil:
		return nil, fmt.Errorf("anticheat: lookup active flag: %w", err)
	}

	// Existing active flag — escalate severity (clamped) and refresh
	// the metric/z snapshot so reviewers see the current intensity.
	severity := existingSeverity + 1
	if severity > MaxSeverity {
		severity = MaxSeverity
	}
	if _, err := a.db.ExecContext(ctx, `
		UPDATE contributor_flags
		SET severity = ?, metric_value = ?, pop_mean = ?, pop_stddev = ?, z_score = ?, detected_at = ?
		WHERE id = ?`,
		severity, d.value, mean, sd, z, now, existingID); err != nil {
		return nil, fmt.Errorf("anticheat: escalate flag: %w", err)
	}
	return &Flag{
		ID:            existingID,
		ContributorID: id,
		Metric:        d.metric,
		MetricValue:   d.value,
		PopMean:       mean,
		PopStdDev:     sd,
		ZScore:        z,
		Severity:      severity,
		DetectedAt:    time.Unix(now, 0),
	}, nil
}

// ListFlags returns flags ordered by detected_at descending. When
// onlyActive is true only unresolved rows are included.
func (a *StatisticalAuditor) ListFlags(ctx context.Context, onlyActive bool, limit int) ([]Flag, error) {
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	q := `SELECT id, contributor_id, metric, metric_value, pop_mean, pop_stddev,
	             z_score, severity, detected_at, resolved_at, resolved_by, resolved_note
	      FROM contributor_flags`
	if onlyActive {
		q += ` WHERE resolved_at IS NULL`
	}
	q += ` ORDER BY detected_at DESC LIMIT ?`

	rows, err := a.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Flag
	for rows.Next() {
		var f Flag
		var detected int64
		var resolvedAt sql.NullInt64
		var resolvedBy, resolvedNote sql.NullString
		if err := rows.Scan(&f.ID, &f.ContributorID, &f.Metric, &f.MetricValue,
			&f.PopMean, &f.PopStdDev, &f.ZScore, &f.Severity, &detected,
			&resolvedAt, &resolvedBy, &resolvedNote); err != nil {
			return nil, err
		}
		f.DetectedAt = time.Unix(detected, 0)
		if resolvedAt.Valid {
			t := time.Unix(resolvedAt.Int64, 0)
			f.ResolvedAt = &t
		}
		if resolvedBy.Valid {
			f.ResolvedBy = resolvedBy.String
		}
		if resolvedNote.Valid {
			f.ResolvedNote = resolvedNote.String
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// ResolveFlag marks a single flag as reviewed. Returns sql.ErrNoRows
// when the id doesn't exist or the flag is already resolved (the second
// case keeps the audit history append-only — re-resolving would
// silently lose the previous reviewer's record).
func (a *StatisticalAuditor) ResolveFlag(ctx context.Context, id int64, by, note string) error {
	res, err := a.db.ExecContext(ctx, `
		UPDATE contributor_flags
		SET resolved_at = ?, resolved_by = ?, resolved_note = ?
		WHERE id = ? AND resolved_at IS NULL`,
		a.now().Unix(), by, note, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// meanStdDev returns the sample mean and Bessel-corrected stddev of
// the input. Caller guarantees len(xs) >= 2.
func meanStdDev(xs []float64) (mean, sd float64) {
	n := float64(len(xs))
	for _, v := range xs {
		mean += v
	}
	mean /= n
	var sumSq float64
	for _, v := range xs {
		d := v - mean
		sumSq += d * d
	}
	sd = math.Sqrt(sumSq / (n - 1))
	return mean, sd
}
