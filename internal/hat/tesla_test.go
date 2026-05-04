package hat

import (
	"math"
	"testing"
)

func TestExtractPivot_BasicSwing(t *testing.T) {
	// Winner (seat 0) has a big eval jump at turn 10.
	history := map[int][]float64{
		1:  {0.25, 0.25, 0.25, 0.25},
		5:  {0.30, 0.25, 0.23, 0.22},
		10: {0.70, 0.15, 0.10, 0.05}, // big swing for seat 0
		15: {0.80, 0.10, 0.05, 0.05},
		20: {0.90, 0.05, 0.03, 0.02},
	}
	pivot := ExtractPivot(history, 0, 20)
	if pivot.Turn != 10 {
		t.Errorf("pivot should be at turn 10, got %d", pivot.Turn)
	}
	if pivot.DeltaScore <= 0 {
		t.Error("pivot delta should be positive")
	}
}

func TestExtractPivot_TwoSamples(t *testing.T) {
	history := map[int][]float64{
		1:  {0.5, 0.5},
		10: {0.9, 0.1},
	}
	pivot := ExtractPivot(history, 0, 10)
	if pivot.Turn != 10 {
		t.Errorf("expected pivot at turn 10, got %d", pivot.Turn)
	}
}

func TestExtractPivot_SingleSample(t *testing.T) {
	history := map[int][]float64{
		5: {0.5, 0.5},
	}
	pivot := ExtractPivot(history, 0, 10)
	// With <2 samples, defaults to midpoint.
	if pivot.Turn != 5 {
		t.Errorf("expected midpoint pivot, got %d", pivot.Turn)
	}
}

func TestLabelSamplesWithPivot(t *testing.T) {
	samples := []TrainingSample{
		{Turn: 1, GameTurn: 20},
		{Turn: 10, GameTurn: 20},
		{Turn: 15, GameTurn: 20},
		{Turn: 20, GameTurn: 20},
	}
	pivot := CausalPivot{Turn: 10, WinnerSeat: 0, DeltaScore: 0.5}
	labels := LabelSamplesWithPivot(samples, pivot)

	if len(labels) != 4 {
		t.Fatalf("expected 4 labels, got %d", len(labels))
	}
	// Turn 10 is at the pivot: distance should be 0.
	if labels[1].PivotDistance != 0.0 {
		t.Errorf("at-pivot distance should be 0, got %f", labels[1].PivotDistance)
	}
	if labels[1].IsPostPivot != true {
		t.Error("turn 10 should be post-pivot (inclusive)")
	}
	// Turn 1 is 9 turns from pivot in a 20-turn game: 9/20 = 0.45.
	expected := 9.0 / 20.0
	if math.Abs(labels[0].PivotDistance-expected) > 0.01 {
		t.Errorf("turn 1 distance should be ~%.2f, got %.2f", expected, labels[0].PivotDistance)
	}
	if labels[0].IsPostPivot {
		t.Error("turn 1 should be pre-pivot")
	}
}

func TestEnrichSamples(t *testing.T) {
	samples := []TrainingSample{
		{Placement: 1.0, Turn: 5, GameTurn: 20},
	}
	labels := []PivotLabel{
		{PivotTurn: 10, PivotDistance: 0.25, IsPostPivot: false},
	}
	enriched := EnrichSamples(samples, labels)
	if len(enriched) != 1 {
		t.Fatal("expected 1 enriched sample")
	}
	if enriched[0].PivotTurn != 10 {
		t.Errorf("pivot turn should be 10, got %d", enriched[0].PivotTurn)
	}
	if enriched[0].Placement != 1.0 {
		t.Errorf("placement should be preserved, got %f", enriched[0].Placement)
	}
}

func TestEvalSnapshotCollector(t *testing.T) {
	c := NewEvalSnapshotCollector()
	c.Record(1, []float64{0.5, 0.5})
	c.Record(5, []float64{0.6, 0.4})

	h := c.History()
	if len(h) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(h))
	}
	if h[5][0] != 0.6 {
		t.Errorf("expected 0.6 at turn 5 seat 0, got %f", h[5][0])
	}
}
