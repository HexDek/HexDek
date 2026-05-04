package hat

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestEncodeState_Dimensions(t *testing.T) {
	gs := newTestGame(t, 4)
	for i := range gs.Seats {
		gs.Seats[i].Life = 40
		gs.Seats[i].StartingLife = 40
	}
	gs.Turn = 10

	v := EncodeState(gs, 0)
	if len(v) != StateVectorDim {
		t.Fatalf("expected %d dimensions, got %d", StateVectorDim, len(v))
	}

	// Life should be normalized: 40/40 = 1.0.
	if v[0] != 1.0 {
		t.Errorf("seat 0 life should be 1.0, got %.3f", v[0])
	}
	// Turn 10/80 = 0.125.
	gBase := maxSeats * seatFeatures
	if math.Abs(v[gBase]-0.125) > 0.001 {
		t.Errorf("turn should be ~0.125, got %.3f", v[gBase])
	}
	// Game stage at turn 10 = mid = 0.5.
	if v[gBase+2] != 0.5 {
		t.Errorf("game stage should be 0.5 (mid), got %.3f", v[gBase+2])
	}
}

func TestEncodeState_PerspectiveRotation(t *testing.T) {
	gs := newTestGame(t, 4)
	gs.Seats[0].Life = 10
	gs.Seats[1].Life = 20
	gs.Seats[2].Life = 30
	gs.Seats[3].Life = 40
	for i := range gs.Seats {
		gs.Seats[i].StartingLife = 40
	}

	// From seat 2's perspective, slot 0 should have seat 2's life (30).
	v := EncodeState(gs, 2)
	expected := 30.0 / 40.0
	if math.Abs(v[0]-expected) > 0.001 {
		t.Errorf("perspective seat life should be %.3f (30/40), got %.3f", expected, v[0])
	}
	// Slot 1 should be seat 3's life (40/40 = 1.0).
	if math.Abs(v[seatFeatures]-1.0) > 0.001 {
		t.Errorf("slot 1 life should be 1.0 (40/40), got %.3f", v[seatFeatures])
	}
}

func TestEncodeState_BattlefieldCounts(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[0].Life = 40
	gs.Seats[0].StartingLife = 40
	gs.Seats[1].Life = 40
	gs.Seats[1].StartingLife = 40

	// Add 3 creatures to seat 0.
	for i := 0; i < 3; i++ {
		c := newTestCardMinimal("Bear "+itoa(i), []string{"creature"}, 2, nil)
		newTestPermanent(gs.Seats[0], c, 2, 2)
	}

	v := EncodeState(gs, 0)
	// creature_count at index 8: 3/15 = 0.2
	expected := 3.0 / 15.0
	if math.Abs(v[8]-expected) > 0.001 {
		t.Errorf("creature count should be %.3f, got %.3f", expected, v[8])
	}
	// total_power at index 12: 6/40 = 0.15
	powExpected := 6.0 / 40.0
	if math.Abs(v[12]-powExpected) > 0.001 {
		t.Errorf("total power should be %.3f, got %.3f", powExpected, v[12])
	}
}

func TestEncodeState_NilSafe(t *testing.T) {
	v := EncodeState(nil, 0)
	for i, f := range v {
		if f != 0 {
			t.Errorf("nil gs should produce zero vector, got %.3f at index %d", f, i)
		}
	}
}

func TestNeuralEvaluator_Forward(t *testing.T) {
	// Tiny 2-layer model: 92 → 2 → 1
	w1 := make([][]float64, 2)
	for i := range w1 {
		w1[i] = make([]float64, StateVectorDim)
	}
	w1[0][0] = 5.0              // hidden 0 responds to own life
	w1[1][seatFeatures] = 5.0   // hidden 1 responds to opponent life

	ne := &NeuralEvaluator{
		Layers: []NeuralLayer{
			{Weights: w1, Biases: []float64{0.0, 0.0}},
			{Weights: [][]float64{{1.0, -1.0}}, Biases: []float64{0.0}},
		},
	}

	// Scenario: high own life, low opponent life → positive score.
	var v StateVector
	v[0] = 1.0
	v[seatFeatures] = 0.25

	score := ne.Evaluate(v)
	if score <= 0.5 {
		t.Errorf("high life vs low opponent should score > 0.5, got %.3f", score)
	}

	// Reverse: low own life, high opponent life → lower score.
	var v2 StateVector
	v2[0] = 0.25
	v2[seatFeatures] = 1.0

	score2 := ne.Evaluate(v2)
	if score2 >= score {
		t.Errorf("reversed scenario should score lower: %.3f >= %.3f", score2, score)
	}
}

func TestNeuralEvaluator_MultiLayer(t *testing.T) {
	// 3-layer model: 92 → 4 → 2 → 1
	w1 := make([][]float64, 4)
	for i := range w1 {
		w1[i] = make([]float64, StateVectorDim)
		w1[i][i%StateVectorDim] = 3.0
	}
	w2 := [][]float64{
		{1.0, 0.5, -0.5, -1.0},
		{-1.0, 1.0, 0.5, -0.5},
	}
	w3 := [][]float64{
		{2.0, -1.0},
	}

	ne := &NeuralEvaluator{
		Layers: []NeuralLayer{
			{Weights: w1, Biases: []float64{0, 0, 0, 0}},
			{Weights: w2, Biases: []float64{0, 0}},
			{Weights: w3, Biases: []float64{0}},
		},
		Arch: "mlp",
	}

	var v StateVector
	v[0] = 0.8
	v[1] = 0.3

	score := ne.Evaluate(v)
	if score <= 0 || score >= 1 {
		t.Errorf("multi-layer score should be in (0,1), got %.6f", score)
	}
	if math.IsNaN(score) {
		t.Fatal("multi-layer produced NaN")
	}
}

func TestTrainingCollector_SnapshotAndFinalize(t *testing.T) {
	gs := newTestGame(t, 4)
	for i := range gs.Seats {
		gs.Seats[i].Life = 40
		gs.Seats[i].StartingLife = 40
	}

	tc := NewTrainingCollector(5) // snapshot every 5 turns

	// Turn 1 always gets a snapshot.
	gs.Turn = 1
	tc.Snapshot(gs, 0)
	// Turn 5 gets a snapshot.
	gs.Turn = 5
	tc.Snapshot(gs, 0)
	// Turn 7 does NOT.
	gs.Turn = 7
	tc.Snapshot(gs, 0)
	// Turn 10 gets one.
	gs.Turn = 10
	tc.Snapshot(gs, 0)

	samples := tc.Finalize(0.75, 12)
	if len(samples) != 3 {
		t.Fatalf("expected 3 samples (turns 1, 5, 10), got %d", len(samples))
	}
	for _, s := range samples {
		if s.Placement != 0.75 {
			t.Errorf("placement should be 0.75, got %.3f", s.Placement)
		}
		if s.GameTurn != 12 {
			t.Errorf("game_turn should be 12, got %d", s.GameTurn)
		}
	}
}

func TestAppendSamples_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "samples.jsonl")

	gs := newTestGame(t, 4)
	for i := range gs.Seats {
		gs.Seats[i].Life = 40
		gs.Seats[i].StartingLife = 40
	}
	_ = rand.New(rand.NewSource(42))

	samples := []TrainingSample{
		{State: EncodeState(gs, 0), Placement: 1.0, Turn: 5, GameTurn: 20},
		{State: EncodeState(gs, 1), Placement: 0.33, Turn: 10, GameTurn: 20},
	}

	if err := AppendSamples(path, samples); err != nil {
		t.Fatalf("AppendSamples: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("file should not be empty")
	}
}

func TestNeuralEvaluator_EmptyModel(t *testing.T) {
	ne := &NeuralEvaluator{}
	var v StateVector
	score := ne.Evaluate(v)
	if score != 0.5 {
		t.Errorf("empty model should return 0.5, got %.3f", score)
	}
}

func TestLoadNeuralEvaluator_LegacyFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "legacy.json")

	// Write old "mlp-1h" format.
	w1 := make([][]float64, 2)
	for i := range w1 {
		w1[i] = make([]float64, StateVectorDim)
		w1[i][0] = float64(i + 1)
	}
	legacy := map[string]any{
		"arch": "mlp-1h",
		"w1":   w1,
		"b1":   []float64{0.1, 0.2},
		"w2":   []float64{1.0, -1.0},
		"b2":   0.5,
	}
	data, _ := json.Marshal(legacy)
	os.WriteFile(path, data, 0644)

	ne, err := LoadNeuralEvaluator(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(ne.Layers) != 2 {
		t.Fatalf("legacy should produce 2 layers, got %d", len(ne.Layers))
	}
	if len(ne.Layers[0].Weights) != 2 {
		t.Errorf("layer 0 should have 2 rows, got %d", len(ne.Layers[0].Weights))
	}
	if len(ne.Layers[1].Weights) != 1 {
		t.Errorf("output layer should have 1 row, got %d", len(ne.Layers[1].Weights))
	}

	var v StateVector
	v[0] = 0.5
	score := ne.Evaluate(v)
	if math.IsNaN(score) || score <= 0 || score >= 1 {
		t.Errorf("legacy model score should be in (0,1), got %.6f", score)
	}
}
