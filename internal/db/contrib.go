package db

import (
	"context"
	"database/sql"
	"errors"
	"math"
)

// ContributorCredits is the persisted state of a single contributor's
// distributed-compute account. Mirrors a row of contributor_credits.
type ContributorCredits struct {
	Owner            string
	CreditsTotal     int64
	ChunksCompleted  int64
	ChunksRejected   int64
	GamesSimulated   int64
	ElapsedMSN       int64
	ElapsedMSMean    float64
	ElapsedMSM2      float64
	LastZScore       float64
	Frozen           bool
	FrozenReason     string
	FirstSeenAt      int64
	LastActiveAt     int64
}

// GetContributorCredits returns the credits row for owner, or a zero
// row (with Owner set) if none exists yet.
func GetContributorCredits(ctx context.Context, database *sql.DB, owner string) (*ContributorCredits, error) {
	c := &ContributorCredits{Owner: owner}
	var frozen int64
	err := database.QueryRowContext(ctx, `
		SELECT credits_total, chunks_completed, chunks_rejected, games_simulated,
		       elapsed_ms_n, elapsed_ms_mean, elapsed_ms_m2, last_z_score,
		       frozen, frozen_reason, first_seen_at, last_active_at
		  FROM contributor_credits WHERE owner = ?`, owner,
	).Scan(
		&c.CreditsTotal, &c.ChunksCompleted, &c.ChunksRejected, &c.GamesSimulated,
		&c.ElapsedMSN, &c.ElapsedMSMean, &c.ElapsedMSM2, &c.LastZScore,
		&frozen, &c.FrozenReason, &c.FirstSeenAt, &c.LastActiveAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}
	c.Frozen = frozen != 0
	return c, nil
}

// EnsureContributorRow inserts a zero row for owner if none exists.
// Idempotent — safe to call on every connect.
func EnsureContributorRow(ctx context.Context, database *sql.DB, owner string, now int64) error {
	_, err := database.ExecContext(ctx, `
		INSERT INTO contributor_credits (owner, first_seen_at, last_active_at)
		VALUES (?, ?, ?)
		ON CONFLICT(owner) DO UPDATE SET last_active_at = excluded.last_active_at`,
		owner, now, now)
	return err
}

// AwardCredits commits an accepted chunk: updates running stats via
// Welford, increments credits/chunks/games counters, recomputes the
// last z-score, and freezes the account if the z exceeds threshold.
//
// All updates are in a single transaction so the credits/M2/last_z
// stay consistent.
func AwardCredits(ctx context.Context, database *sql.DB, owner string, credits, games int64, elapsedMS int64, freezeAtZ float64, now int64) (*ContributorCredits, error) {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Read existing row (creating one if missing).
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO contributor_credits (owner, first_seen_at, last_active_at)
		VALUES (?, ?, ?)
		ON CONFLICT(owner) DO NOTHING`, owner, now, now); err != nil {
		return nil, err
	}
	c := &ContributorCredits{Owner: owner}
	var frozen int64
	if err := tx.QueryRowContext(ctx, `
		SELECT credits_total, chunks_completed, chunks_rejected, games_simulated,
		       elapsed_ms_n, elapsed_ms_mean, elapsed_ms_m2,
		       frozen, frozen_reason, first_seen_at
		  FROM contributor_credits WHERE owner = ?`, owner,
	).Scan(
		&c.CreditsTotal, &c.ChunksCompleted, &c.ChunksRejected, &c.GamesSimulated,
		&c.ElapsedMSN, &c.ElapsedMSMean, &c.ElapsedMSM2,
		&frozen, &c.FrozenReason, &c.FirstSeenAt,
	); err != nil {
		return nil, err
	}
	c.Frozen = frozen != 0

	// Compute z-score against the prior distribution BEFORE folding the
	// new sample in — that way one anomalous chunk shows up at full
	// magnitude instead of being washed out by its own contribution.
	z := 0.0
	if c.ElapsedMSN >= 5 {
		variance := c.ElapsedMSM2 / float64(c.ElapsedMSN-1)
		if variance > 0 {
			std := math.Sqrt(variance)
			z = (float64(elapsedMS) - c.ElapsedMSMean) / std
		}
	}
	c.LastZScore = z

	// Welford update for the next round.
	c.ElapsedMSN++
	delta := float64(elapsedMS) - c.ElapsedMSMean
	c.ElapsedMSMean += delta / float64(c.ElapsedMSN)
	delta2 := float64(elapsedMS) - c.ElapsedMSMean
	c.ElapsedMSM2 += delta * delta2

	// Award.
	if !c.Frozen {
		c.CreditsTotal += credits
	}
	c.ChunksCompleted++
	c.GamesSimulated += games
	c.LastActiveAt = now

	// Trip freeze on 3-sigma slow OR fast outliers (suspiciously fast
	// implies fabricated results; suspiciously slow implies a bad
	// worker that we'd rather not pay).
	if !c.Frozen && math.Abs(z) >= freezeAtZ {
		c.Frozen = true
		c.FrozenReason = "anomaly z=" + ftoa(z)
		frozen = 1
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE contributor_credits SET
		  credits_total = ?, chunks_completed = ?, chunks_rejected = ?,
		  games_simulated = ?, elapsed_ms_n = ?, elapsed_ms_mean = ?, elapsed_ms_m2 = ?,
		  last_z_score = ?, frozen = ?, frozen_reason = ?, last_active_at = ?
		WHERE owner = ?`,
		c.CreditsTotal, c.ChunksCompleted, c.ChunksRejected,
		c.GamesSimulated, c.ElapsedMSN, c.ElapsedMSMean, c.ElapsedMSM2,
		c.LastZScore, frozen, c.FrozenReason, c.LastActiveAt,
		owner,
	); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return c, nil
}

// RejectChunk increments the rejected count for a contributor (used
// when spot-check or sig verification fails). No credits awarded.
func RejectChunk(ctx context.Context, database *sql.DB, owner string, now int64) error {
	if _, err := database.ExecContext(ctx, `
		INSERT INTO contributor_credits (owner, first_seen_at, last_active_at, chunks_rejected)
		VALUES (?, ?, ?, 1)
		ON CONFLICT(owner) DO UPDATE SET
		  chunks_rejected = chunks_rejected + 1,
		  last_active_at = excluded.last_active_at`,
		owner, now, now); err != nil {
		return err
	}
	return nil
}

// RecordChunkAssignment writes a pending chunk row at issuance.
func RecordChunkAssignment(ctx context.Context, database *sql.DB, chunkID, owner string, games, seats int, issuedAt int64) error {
	_, err := database.ExecContext(ctx, `
		INSERT INTO contrib_chunk (chunk_id, owner, issued_at, games_count, n_seats)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(chunk_id) DO NOTHING`,
		chunkID, owner, issuedAt, games, seats)
	return err
}

// FinalizeChunkRow updates a chunk row after it's been validated.
func FinalizeChunkRow(ctx context.Context, database *sql.DB, chunkID string, accepted int, spotChecked, spotPassed bool, elapsedMS int64, hash string, credits int64, reason string, now int64) error {
	sc := 0
	if spotChecked {
		sc = 1
	}
	sp := 0
	if spotPassed {
		sp = 1
	}
	_, err := database.ExecContext(ctx, `
		UPDATE contrib_chunk SET
		  returned_at = ?, accepted = ?, spot_checked = ?, spot_check_passed = ?,
		  elapsed_ms = ?, outcome_hash = ?, credits_awarded = ?, reason = ?
		WHERE chunk_id = ?`,
		now, accepted, sc, sp, elapsedMS, hash, credits, reason, chunkID,
	)
	return err
}

// ftoa formats a float without scientific notation for the
// frozen_reason text. Avoids importing strconv just for this.
func ftoa(f float64) string {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return "NaN"
	}
	// Round to 2 decimals.
	scaled := math.Round(f*100) / 100
	// Convert via fmt-equivalent — simple approach.
	return strconvFmtFloat(scaled)
}

// strconvFmtFloat is a thin wrapper to avoid an explicit strconv import
// at package top (kept symmetric with the rest of this file's style —
// helpers stay local).
func strconvFmtFloat(f float64) string {
	// Use scientific path for huge values, otherwise fixed.
	const epsilon = 1e-9
	if f == 0 || (math.Abs(f) < epsilon) {
		return "0.00"
	}
	sign := ""
	if f < 0 {
		sign = "-"
		f = -f
	}
	whole := int64(f)
	frac := int64(math.Round((f - float64(whole)) * 100))
	if frac == 100 {
		whole++
		frac = 0
	}
	wholePart := itoa64(whole)
	fracPart := "00"
	if frac >= 10 {
		fracPart = itoa64(frac)
	} else if frac > 0 {
		fracPart = "0" + itoa64(frac)
	}
	return sign + wholePart + "." + fracPart
}

func itoa64(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
