package tournament

import (
	"path/filepath"
	"time"

	"github.com/hexdek/hexdek/internal/achievements"
)

// ownerFromDeckPath extracts the owner name from a deck file path
// following the data/decks/{owner}/{deck}.txt layout the tournament
// runner expects. Returns empty string for empty input.
func ownerFromDeckPath(p string) string {
	if p == "" {
		return ""
	}
	return filepath.Base(filepath.Dir(p))
}

// awardAchievements feeds a single GameOutcome to the achievements
// tracker. owners is a slice parallel to commander/deck index so
// SeatStats.CommanderIdx maps to the right owner. No-op when tracker
// is nil, no PostGameStats are present, or no owners can be resolved.
func awardAchievements(tracker *achievements.Tracker, o GameOutcome, owners []string) {
	if tracker == nil || len(o.PostGameStats) == 0 || len(owners) == 0 {
		return
	}
	seats := make([]achievements.SeatOutcome, 0, len(o.PostGameStats))
	for _, ss := range o.PostGameStats {
		if ss.CommanderIdx < 0 || ss.CommanderIdx >= len(owners) {
			continue
		}
		owner := owners[ss.CommanderIdx]
		if owner == "" {
			continue
		}
		seats = append(seats, achievements.SeatOutcome{
			Owner:     owner,
			Won:       ss.Won,
			FinalLife: ss.FinalLife,
		})
	}
	if len(seats) == 0 {
		return
	}
	_ = tracker.OnGameComplete(achievements.GameOutcome{
		Turns:      o.Turns,
		Seats:      seats,
		FinishedAt: time.Now(),
	})
}
