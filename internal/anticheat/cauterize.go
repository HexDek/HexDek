package anticheat

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/hexdek/hexdek/internal/db"
)

// QuarantineWindow is how far back from "now" a sanctioned
// contributor's recent games get marked unverified. 7 days mirrors
// the showmatch ELO decay window — older games are too stale to
// matter for in-flight rankings.
const QuarantineWindow = 7 * 24 * time.Hour

// SanctionDecision is the result of a single cauterize call. The
// CauterizeService returns it to the caller (typically the
// VerificationWorker) so the failure detail can be threaded into the
// queue row's `detail` field.
type SanctionDecision struct {
	SanctionID    int64
	Severity      string
	OffenseNum    int
	Reason        string
	ExpiresAt     int64 // 0 = never expires (warning or permanent)
	GamesQuarantined int64
}

// CauterizeService applies the escalation policy:
//
//	1st offense → warning (no ban, contributor stays active)
//	2nd offense → 24-hour temp ban
//	3rd+ offense → permanent ban
//
// On every offense it ALSO marks the contributor's recent games
// (last QuarantineWindow) as unverified so downstream tooling can
// strip them from leaderboards / analytics.
type CauterizeService struct {
	db *sql.DB
}

func NewCauterizeService(database *sql.DB) *CauterizeService {
	return &CauterizeService{db: database}
}

// ApplyOnFailure issues the next-tier sanction for deckKey and
// quarantines its recent games. queueID is the verification_queue row
// that triggered this; pass 0 for sanctions issued out-of-band (e.g.
// admin-initiated).
//
// Idempotency: re-calling ApplyOnFailure for the same deckKey ALWAYS
// escalates by one tier — the worker is expected to call this at most
// once per failed verification.
func (c *CauterizeService) ApplyOnFailure(ctx context.Context, deckKey string, queueID int64, reason string) (SanctionDecision, error) {
	dec := SanctionDecision{}
	if strings.TrimSpace(deckKey) == "" {
		return dec, fmt.Errorf("cauterize: empty deck_key")
	}

	priorOffenses, err := db.CountOffenses(ctx, c.db, deckKey)
	if err != nil {
		return dec, fmt.Errorf("count offenses: %w", err)
	}
	dec.OffenseNum = priorOffenses + 1
	dec.Reason = reason

	now := time.Now().Unix()
	switch dec.OffenseNum {
	case 1:
		dec.Severity = db.SeverityWarning
		dec.ExpiresAt = 0
	case 2:
		dec.Severity = db.SeverityTempBan
		dec.ExpiresAt = now + int64((24 * time.Hour).Seconds())
	default:
		dec.Severity = db.SeverityPermanentBan
		dec.ExpiresAt = 0
	}

	owner := ownerFromDeckKey(deckKey)
	sid, err := db.InsertSanction(ctx, c.db, db.SanctionInsertParams{
		DeckKey:   deckKey,
		Owner:     owner,
		Severity:  dec.Severity,
		Reason:    reason,
		QueueID:   queueID,
		IssuedAt:  now,
		ExpiresAt: dec.ExpiresAt,
	})
	if err != nil {
		return dec, fmt.Errorf("insert sanction: %w", err)
	}
	dec.SanctionID = sid

	since := now - int64(QuarantineWindow.Seconds())
	n, err := db.QuarantineRecentGames(ctx, c.db, deckKey, since)
	if err != nil {
		// Sanction is already inserted; report the partial state via
		// error, the caller can decide whether to retry.
		return dec, fmt.Errorf("quarantine: %w", err)
	}
	dec.GamesQuarantined = n
	return dec, nil
}

// IsBanned reports whether deckKey currently has an active temp_ban
// or permanent_ban. Warnings do not count as bans (by design — the
// 1st offense is meant to inform, not block).
func (c *CauterizeService) IsBanned(ctx context.Context, deckKey string) (bool, *db.SanctionRow, error) {
	row, err := db.ActiveBan(ctx, c.db, deckKey)
	if err != nil {
		return false, nil, err
	}
	return row != nil, row, nil
}

// ownerFromDeckKey extracts the "owner" half of the canonical
// "owner/name" deck key. Mirrors deckparser conventions used
// elsewhere in the repo.
func ownerFromDeckKey(key string) string {
	if i := strings.Index(key, "/"); i > 0 {
		return key[:i]
	}
	return ""
}
