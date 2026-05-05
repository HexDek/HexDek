package hat

import (
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestMicroNet_Architecture(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	mn := NewMicroNet("test-deck", rng)

	if len(mn.Layers) != 4 {
		t.Fatalf("expected 4 layers, got %d", len(mn.Layers))
	}

	expected := []struct{ in, out int }{
		{MicroInputDim, MicroHidden1},
		{MicroHidden1, MicroHidden2},
		{MicroHidden2, MicroHidden3},
		{MicroHidden3, MicroOutputDim},
	}
	for i, exp := range expected {
		if len(mn.Layers[i].Weights) != exp.out {
			t.Errorf("layer %d: expected %d output neurons, got %d", i, exp.out, len(mn.Layers[i].Weights))
		}
		if len(mn.Layers[i].Weights[0]) != exp.in {
			t.Errorf("layer %d: expected %d input dim, got %d", i, exp.in, len(mn.Layers[i].Weights[0]))
		}
		if len(mn.Layers[i].Biases) != exp.out {
			t.Errorf("layer %d: expected %d biases, got %d", i, exp.out, len(mn.Layers[i].Biases))
		}
	}

	params := MicroParamCount(mn)
	if params < 5000 || params > 6500 {
		t.Errorf("expected ~5900 params, got %d", params)
	}
}

func TestMicroNet_ForwardPass(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	mn := NewMicroNet("test-deck", rng)

	var input [MicroInputDim]float64
	for i := range input {
		input[i] = rng.Float64()
	}

	output := mn.Forward(input)
	if output < 0 || output > 1 {
		t.Errorf("output should be in [0,1], got %f", output)
	}

	output2 := mn.Forward(input)
	if output != output2 {
		t.Errorf("deterministic forward pass: %f != %f", output, output2)
	}
}

func TestMicroNet_ForwardDifferentInputs(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	mn := NewMicroNet("test-deck", rng)

	var input1, input2 [MicroInputDim]float64
	for i := range input1 {
		input1[i] = 0.1
		input2[i] = 0.9
	}

	o1 := mn.Forward(input1)
	o2 := mn.Forward(input2)
	if o1 == o2 {
		t.Error("different inputs should produce different outputs")
	}
}

func TestMicroNet_TrainReducesLoss(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	mn := NewMicroNet("test-deck", rng)

	samples := make([]MicroSample, 100)
	for i := range samples {
		for j := range samples[i].Input {
			samples[i].Input[j] = rng.Float64()
		}
		samples[i].Placement = samples[i].Input[0]*0.5 + samples[i].Input[1]*0.3
		if samples[i].Placement > 1 {
			samples[i].Placement = 1
		}
	}

	var preLoss float64
	for _, s := range samples {
		pred := mn.Forward(s.Input)
		err := pred - s.Placement
		preLoss += err * err
	}
	preLoss /= float64(len(samples))

	mn.Train(samples, 100, 0.005)

	var postLoss float64
	for _, s := range samples {
		pred := mn.Forward(s.Input)
		err := pred - s.Placement
		postLoss += err * err
	}
	postLoss /= float64(len(samples))

	if postLoss >= preLoss {
		t.Errorf("training should reduce loss: pre=%.6f post=%.6f", preLoss, postLoss)
	}
	if mn.Generation != 1 {
		t.Errorf("expected generation 1 after training, got %d", mn.Generation)
	}
}

func TestMicroNet_SaveLoad(t *testing.T) {
	dir := t.TempDir()
	rng := rand.New(rand.NewSource(42))
	mn := NewMicroNet("test-deck", rng)

	var input [MicroInputDim]float64
	for i := range input {
		input[i] = 0.5
	}
	original := mn.Forward(input)

	if err := SaveMicroNet(dir, mn); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := LoadMicroNet(dir, "test-deck")
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	restored := loaded.Forward(input)
	if math.Abs(original-restored) > 1e-10 {
		t.Errorf("forward pass mismatch after save/load: %f vs %f", original, restored)
	}
}

func TestMicroCollector(t *testing.T) {
	mc := NewMicroCollector("test-deck")

	var input [MicroInputDim]float64
	for i := 0; i < 5; i++ {
		input[0] = float64(i)
		mc.Record(input)
	}

	if mc.Len() != 5 {
		t.Fatalf("expected 5 samples, got %d", mc.Len())
	}

	finalized := mc.Finalize(0.75)
	if len(finalized) != 5 {
		t.Fatalf("expected 5 finalized samples, got %d", len(finalized))
	}
	for _, s := range finalized {
		if s.Placement != 0.75 {
			t.Errorf("expected placement 0.75, got %f", s.Placement)
		}
	}

	if mc.Len() != 0 {
		t.Errorf("collector should be empty after finalize, got %d", mc.Len())
	}
}

func TestMicroTrainer_TrainOnThreshold(t *testing.T) {
	dir := t.TempDir()
	rng := rand.New(rand.NewSource(42))
	mt := NewMicroTrainer(dir, rng)

	samples := make([]MicroSample, microTrainThreshold)
	for i := range samples {
		for j := range samples[i].Input {
			samples[i].Input[j] = rng.Float64()
		}
		samples[i].Placement = rng.Float64()
	}

	mt.AddSamples("train-deck", samples)

	net := mt.GetNet("train-deck")
	if net == nil {
		t.Fatal("expected net to be created after threshold")
	}
	if net.Generation < 1 {
		t.Errorf("expected generation >= 1, got %d", net.Generation)
	}

	fname := filepath.Join(dir, "train-deck.json")
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		t.Error("expected micro-net file to be saved")
	}
}

func TestMicroTrainer_LoadAll(t *testing.T) {
	dir := t.TempDir()
	rng := rand.New(rand.NewSource(42))

	mn := NewMicroNet("saved-deck", rng)
	SaveMicroNet(dir, mn)

	mt := NewMicroTrainer(dir, rng)
	mt.LoadAll()

	if mt.NetCount() != 1 {
		t.Fatalf("expected 1 loaded net, got %d", mt.NetCount())
	}
	if mt.GetNet("saved-deck") == nil {
		t.Error("expected saved-deck net to be loaded")
	}
}

func TestMicroParamCount(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	mn := NewMicroNet("test", rng)
	count := MicroParamCount(mn)

	expected := MicroInputDim*MicroHidden1 + MicroHidden1 +
		MicroHidden1*MicroHidden2 + MicroHidden2 +
		MicroHidden2*MicroHidden3 + MicroHidden3 +
		MicroHidden3*MicroOutputDim + MicroOutputDim

	if count != expected {
		t.Errorf("expected %d params, got %d", expected, count)
	}
	t.Logf("MicroNet param count: %d", count)
}

func TestMicroNet_EmptyTrainNoOp(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	mn := NewMicroNet("test", rng)

	loss := mn.Train(nil, 10, 0.01)
	if loss != 0 {
		t.Errorf("expected 0 loss for empty training, got %f", loss)
	}
}
