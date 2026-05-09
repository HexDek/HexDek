package anticheat

import (
	"context"
	"database/sql"

	"github.com/hexdek/hexdek/internal/hat"
)

// StatisticalAuditor is the persistence-aware wrapper around
// hat.AnomalyDetector that the showmatch loop references on every
// post-game tick. It is the gateway between the (in-memory) Phase-0
// anomaly scaffolding in internal/hat/anomaly.go and the (on-disk)
// Phase 2 spot-check + cauterize pipeline in this package.
//
// The implementation is intentionally minimal: it forwards records
// to hat.DefaultAnomalyDetector so the existing detection logic
// keeps working, and surfaces flags in a shape the showmatch loop
// already expects. Persistence (DB-backed flag history) is a follow-
// up — Phase 2 itself only needs the in-memory detector to keep
// firing so spot-check ordering can later be biased toward flagged
// contributors.
type StatisticalAuditor struct {
	db       *sql.DB
	detector *hat.AnomalyDetector
}

// NewStatisticalAuditor returns an auditor backed by db (for future
// persistence) and a fresh in-memory hat.AnomalyDetector.
func NewStatisticalAuditor(database *sql.DB) (*StatisticalAuditor, error) {
	return &StatisticalAuditor{
		db:       database,
		detector: hat.NewAnomalyDetector(),
	}, nil
}

// Game is the per-record payload from the showmatch post-game hook.
// One Game per seat per finished match.
type Game struct {
	ContributorID string
	Won           bool
	Turns         int
}

// Flag is the post-record signal returned to the caller. Downstream
// hexapi logging accesses each field directly, so renames here are a
// breaking change for showmatch.go.
type Flag struct {
	ContributorID string
	Metric        string
	MetricValue   float64
	ZScore        float64
	Severity      int
	PopMean       float64
	PopStdDev     float64
}

// RecordGame forwards to the in-memory detector and lifts any raised
// AnomalyFlag into the auditor's Flag shape. Returns nil flags when
// the game didn't trigger anything (steady-state). The ctx is
// reserved for future DB writes; today's path is in-memory only.
func (a *StatisticalAuditor) RecordGame(ctx context.Context, g Game) ([]Flag, error) {
	if a == nil || g.ContributorID == "" {
		return nil, nil
	}
	flag := a.detector.Record(g.ContributorID, g.Won)
	if flag == nil {
		return nil, nil
	}
	return []Flag{{
		ContributorID: flag.DeckID,
		Metric:        "win_rate",
		MetricValue:   flag.WinRate,
		ZScore:        flag.ZScore,
		Severity:      severityFromZ(flag.ZScore),
		PopMean:       flag.PopMean,
		PopStdDev:     flag.PopStdDev,
	}}, nil
}

// Detector exposes the underlying hat detector so admin handlers
// (RegisterAdminAnomalies) can list flagged contributors without
// re-implementing the snapshot logic.
func (a *StatisticalAuditor) Detector() *hat.AnomalyDetector {
	if a == nil {
		return nil
	}
	return a.detector
}

func severityFromZ(z float64) int {
	if z < 0 {
		z = -z
	}
	switch {
	case z >= 5.0:
		return 3
	case z >= 4.0:
		return 2
	default:
		return 1
	}
}
