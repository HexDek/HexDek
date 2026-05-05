package tournament

import (
	"math"
	"testing"
)

func TestNormalizeRating_Defaults(t *testing.T) {
	// With no calibration data, uses defaults: floor=1200, ceiling=1800.
	var cal *Calibration

	tests := []struct {
		elo      float64
		expected float64
	}{
		{1200, 0},    // floor
		{1800, 100},  // ceiling
		{1500, 50},   // midpoint
		{1350, 25},   // quarter
		{1650, 75},   // three-quarters
		{1000, -33.33}, // below floor (unclamped)
		{2000, 133.33}, // above ceiling (unclamped)
	}

	for _, tt := range tests {
		got := NormalizeRating(tt.elo, cal)
		if math.Abs(got-tt.expected) > 0.1 {
			t.Errorf("NormalizeRating(%.0f, nil) = %.2f, want %.2f", tt.elo, got, tt.expected)
		}
	}
}

func TestNormalizeRating_WithCalibration(t *testing.T) {
	cal := &Calibration{
		Floor: &CalibrationFloor{
			DeckKeys: []string{"calibration/99lands_golos"},
			AvgELO:   1100,
		},
		Ceiling: &CalibrationCeiling{
			DeckKey: "moxfield/some_cedh_deck",
			ELO:     1900,
		},
	}

	// Span is 800 (1100 to 1900).
	tests := []struct {
		elo      float64
		expected float64
	}{
		{1100, 0},
		{1900, 100},
		{1500, 50},
		{1300, 25},
		{1700, 75},
	}

	for _, tt := range tests {
		got := NormalizeRating(tt.elo, cal)
		if math.Abs(got-tt.expected) > 0.1 {
			t.Errorf("NormalizeRating(%.0f, cal) = %.2f, want %.2f", tt.elo, got, tt.expected)
		}
	}
}

func TestNormalizeRatingClamped(t *testing.T) {
	var cal *Calibration

	// Below floor should clamp to 0.
	got := NormalizeRatingClamped(1000, cal)
	if got != 0 {
		t.Errorf("NormalizeRatingClamped(1000) = %.2f, want 0", got)
	}

	// Above ceiling should clamp to 100.
	got = NormalizeRatingClamped(2000, cal)
	if got != 100 {
		t.Errorf("NormalizeRatingClamped(2000) = %.2f, want 100", got)
	}

	// Within range should pass through.
	got = NormalizeRatingClamped(1500, cal)
	if math.Abs(got-50) > 0.1 {
		t.Errorf("NormalizeRatingClamped(1500) = %.2f, want 50", got)
	}
}

func TestSkillPercentile(t *testing.T) {
	var cal *Calibration // defaults: floor=1200, ceiling=1800

	tests := []struct {
		elo   float64
		grade string
	}{
		{1200, "F"},  // 0 percentile
		{1350, "D"},  // 25
		{1500, "C"},  // 50
		{1600, "B"},  // 66.7
		{1700, "A"},  // 83.3
		{1780, "S"},  // 96.7
		{1800, "S+"}, // 100
	}

	for _, tt := range tests {
		got := SkillPercentile(tt.elo, cal)
		if got != tt.grade {
			pct := NormalizeRatingClamped(tt.elo, cal)
			t.Errorf("SkillPercentile(%.0f) = %q (pct=%.1f), want %q", tt.elo, got, pct, tt.grade)
		}
	}
}

func TestLoadCeilingCalibration_Missing(t *testing.T) {
	// Non-existent file should return nil (graceful degradation).
	result := LoadCeilingCalibration("/tmp/nonexistent_ceiling_test.json")
	if result != nil {
		t.Error("expected nil for missing file")
	}
}
