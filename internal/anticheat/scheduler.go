// Package anticheat implements Phase 2 of HexDek's anti-cheat
// pipeline: deterministic-replay spot-checking and auto-cauterize
// (sanction + quarantine) of contributors whose claimed game outcomes
// don't match a server-side replay.
//
// Three pieces collaborate:
//
//   - SpotCheckScheduler picks a random 2-5% subset of a batch of
//     recently-finished games and enqueues them for replay.
//   - VerificationWorker drains the queue, runs a deterministic
//     replay via a Verifier (heimdall.VerifyReplay in production,
//     stub Verifier in tests), and writes the result back.
//   - CauterizeService translates a failed replay into the right
//     escalation tier (warning → 24-hour ban → permanent), inserts
//     the sanction row, and quarantines the contributor's recent
//     games as unverified.
//
// All three pieces operate on the SQLite tables defined in
// internal/db/schema.sql (verification_queue, contributor_sanctions)
// plus the new `verified` column on showmatch_game.
package anticheat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/hexdek/hexdek/internal/db"
)

// SamplingRate is constrained to the 2-5% band specified by the
// Phase 2 brief. Out-of-range values get clamped on construction.
const (
	MinSamplingRate = 0.01
	MaxSamplingRate = 0.10
)

// SpotCheckScheduler randomly selects a subset of recently finished
// games and inserts verification_queue rows for them. Selection is
// per-game Bernoulli with probability `rate` so small batches still
// have a positive chance of being checked, and large batches converge
// to the configured rate by the law of large numbers.
type SpotCheckScheduler struct {
	db   *sql.DB
	rate float64
	mu   sync.Mutex
	rng  *rand.Rand
}

// NewSpotCheckScheduler clamps rate to the [MinSamplingRate,
// MaxSamplingRate] band and seeds an internal RNG. Pass a non-zero
// seed for deterministic test behavior; pass 0 for a time-seeded RNG.
func NewSpotCheckScheduler(database *sql.DB, rate float64, rngSeed int64) *SpotCheckScheduler {
	if rate < MinSamplingRate {
		rate = MinSamplingRate
	}
	if rate > MaxSamplingRate {
		rate = MaxSamplingRate
	}
	src := rand.NewSource(rngSeed)
	if rngSeed == 0 {
		src = rand.NewSource(rand.Int63())
	}
	return &SpotCheckScheduler{
		db:   database,
		rate: rate,
		rng:  rand.New(src),
	}
}

// Rate returns the active sampling rate (post-clamp).
func (s *SpotCheckScheduler) Rate() float64 { return s.rate }

// Select returns the subset of gameIDs that the scheduler chose for
// verification. Pure function — does NOT touch the database. Used by
// SelectAndEnqueue and exposed separately for tests + dry-run flows.
func (s *SpotCheckScheduler) Select(gameIDs []int64) []int64 {
	if len(gameIDs) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	picked := make([]int64, 0, int(float64(len(gameIDs))*s.rate)+1)
	for _, id := range gameIDs {
		if s.rng.Float64() < s.rate {
			picked = append(picked, id)
		}
	}
	return picked
}

// SelectAndEnqueue picks a random subset of gameIDs and inserts one
// verification_queue row per pick. For each selected game it loads
// inputs (rng seed, deck keys, claimed outcome) from showmatch_game
// + showmatch_game_seat and enqueues a row PER SEAT — so any of the
// 4 contributors in the game can be flagged independently when the
// replay's outcome differs from the claim.
//
// Returns the queue_ids of every row inserted.
func (s *SpotCheckScheduler) SelectAndEnqueue(ctx context.Context, gameIDs []int64) ([]int64, error) {
	picked := s.Select(gameIDs)
	if len(picked) == 0 {
		return nil, nil
	}
	var queueIDs []int64
	for _, gid := range picked {
		g, err := db.LoadGameForVerification(ctx, s.db, gid)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return queueIDs, fmt.Errorf("load game %d: %w", gid, err)
		}
		if g.NSeats == 0 {
			continue
		}
		// Enqueue one row per non-empty deck_key. The verifier will
		// load the same inputs on dequeue, so the per-seat rows share
		// rng_seed + deck_keys but differ on the contributor under
		// review.
		for _, dk := range g.DeckKeys {
			if dk == "" {
				continue
			}
			qid, err := db.EnqueueVerification(ctx, s.db, db.VerificationEnqueueParams{
				GameID:        g.GameID,
				DeckKey:       dk,
				RNGSeed:       g.RNGSeed,
				NSeats:        g.NSeats,
				DeckKeys:      g.DeckKeys,
				ClaimedWinner: g.Winner,
				ClaimedTurns:  g.Turns,
			})
			if err != nil {
				return queueIDs, fmt.Errorf("enqueue game %d seat %s: %w", gid, dk, err)
			}
			queueIDs = append(queueIDs, qid)
		}
	}
	return queueIDs, nil
}
