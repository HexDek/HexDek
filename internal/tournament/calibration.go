package tournament

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/hexdek/hexdek/internal/hat"
)

// CalibrationCeiling is the loaded ceiling reference point (best B5 deck after
// extended Amiibo evolution). Written by cmd/hexdek-ceiling.
type CalibrationCeiling struct {
	DeckKey         string          `json:"deck_key"`
	Commander       string          `json:"commander"`
	Owner           string          `json:"owner"`
	Archetype       string          `json:"archetype"`
	Bracket         int             `json:"bracket"`
	ELO             float64         `json:"elo"`
	HexRating       float64         `json:"hex_rating"`
	WinRate         float64         `json:"win_rate"`
	GamesPlayed     int             `json:"games_played"`
	GauntletWins    int             `json:"gauntlet_wins"`
	GauntletGames   int             `json:"gauntlet_games"`
	GauntletWinRate float64         `json:"gauntlet_win_rate"`
	Generation      int             `json:"generation"`
	BestDNA         *hat.AmiiboDNA  `json:"best_dna"`
	PowerPercentile int             `json:"power_percentile"`
	GameplanSummary string          `json:"gameplan_summary"`
	CalibratedAt    time.Time       `json:"calibrated_at"`
}

// CalibrationFloor is the loaded floor reference (99-basic-land decks).
// The floor ELO is the average rating of the floor calibration decks.
type CalibrationFloor struct {
	DeckKeys []string `json:"deck_keys"`
	AvgELO   float64  `json:"avg_elo"`
}

// Calibration holds both floor and ceiling reference points for rating normalization.
type Calibration struct {
	Floor   *CalibrationFloor
	Ceiling *CalibrationCeiling
}

const (
	defaultCeilingPath = "data/calibration/ceiling.json"
	defaultFloorELO    = 1200.0 // Conservative default if no floor data exists
	defaultCeilingELO  = 1800.0 // Conservative default if no ceiling data exists
)

// LoadCeilingCalibration reads the ceiling calibration from disk.
// Returns nil if the file doesn't exist or can't be parsed (graceful degradation).
func LoadCeilingCalibration(path string) *CalibrationCeiling {
	if path == "" {
		path = defaultCeilingPath
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var ceiling CalibrationCeiling
	if err := json.Unmarshal(data, &ceiling); err != nil {
		return nil
	}
	return &ceiling
}

// LoadCalibration loads both floor and ceiling data. Provides sensible defaults
// if either is missing.
func LoadCalibration(ceilingPath string, floorELOs map[string]float64) *Calibration {
	cal := &Calibration{}

	// Load ceiling.
	cal.Ceiling = LoadCeilingCalibration(ceilingPath)

	// Build floor from known floor deck ELOs.
	if len(floorELOs) > 0 {
		floor := &CalibrationFloor{}
		total := 0.0
		for key, elo := range floorELOs {
			floor.DeckKeys = append(floor.DeckKeys, key)
			total += elo
		}
		if len(floor.DeckKeys) > 0 {
			floor.AvgELO = total / float64(len(floor.DeckKeys))
		}
		cal.Floor = floor
	}

	return cal
}

// NormalizeRating maps a raw ELO to a 0-100 scale where floor calibration
// decks = 0 and the ceiling deck = 100. Values can exceed 0-100 for decks
// outside the calibration range.
//
// Uses linear interpolation between floor and ceiling ELO anchors.
// If calibration data is not available, uses conservative defaults
// (floor=1200, ceiling=1800).
func NormalizeRating(elo float64, cal *Calibration) float64 {
	floorELO := defaultFloorELO
	ceilingELO := defaultCeilingELO

	if cal != nil {
		if cal.Floor != nil && cal.Floor.AvgELO > 0 {
			floorELO = cal.Floor.AvgELO
		}
		if cal.Ceiling != nil && cal.Ceiling.ELO > 0 {
			ceilingELO = cal.Ceiling.ELO
		}
	}

	// Prevent division by zero.
	span := ceilingELO - floorELO
	if span < 1 {
		span = 1
	}

	normalized := (elo - floorELO) / span * 100.0

	// Clamp to [0, 100] for display purposes (allow slight overshoot internally).
	return normalized
}

// NormalizeRatingClamped is like NormalizeRating but clamps output to [0, 100].
func NormalizeRatingClamped(elo float64, cal *Calibration) float64 {
	n := NormalizeRating(elo, cal)
	return math.Max(0, math.Min(100, n))
}

// SkillPercentile returns a human-friendly percentile string for leaderboard display.
func SkillPercentile(elo float64, cal *Calibration) string {
	pct := NormalizeRatingClamped(elo, cal)
	if pct >= 99 {
		return "S+"
	}
	if pct >= 90 {
		return "S"
	}
	if pct >= 80 {
		return "A"
	}
	if pct >= 65 {
		return "B"
	}
	if pct >= 45 {
		return "C"
	}
	if pct >= 25 {
		return "D"
	}
	return "F"
}

// FormatNormalized returns a formatted string like "73.2 (B)" for display.
func FormatNormalized(elo float64, cal *Calibration) string {
	pct := NormalizeRatingClamped(elo, cal)
	grade := SkillPercentile(elo, cal)
	return fmt.Sprintf("%.1f (%s)", pct, grade)
}
