package hat

import (
	"encoding/json"
	"math"
	"os"
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

const (
	seatFeatures   = 22
	globalFeatures = 4
	maxSeats       = 4
	StateVectorDim = seatFeatures*maxSeats + globalFeatures // 92
)

// StateVector is a fixed-size numeric encoding of a Commander game state
// from a specific seat's perspective. All values are normalized to
// approximately [0, 1]. Seat 0 in the vector is always the perspective
// seat; opponents are rotated into slots 1-3.
type StateVector [StateVectorDim]float64

// EncodeState converts a GameState into a StateVector from the given
// seat's perspective. The encoding is designed for neural network input:
// all features are normalized, perspective-invariant (seat 0 = self),
// and information-complete (no hidden info — the model sees what the
// hat sees).
func EncodeState(gs *gameengine.GameState, perspectiveSeat int) StateVector {
	var v StateVector
	if gs == nil {
		return v
	}

	nSeats := len(gs.Seats)
	if nSeats == 0 || perspectiveSeat < 0 || perspectiveSeat >= nSeats {
		return v
	}

	for slot := 0; slot < maxSeats && slot < nSeats; slot++ {
		// Rotate so perspective seat is always slot 0.
		actualSeat := (perspectiveSeat + slot) % nSeats
		s := gs.Seats[actualSeat]
		if s == nil {
			continue
		}
		base := slot * seatFeatures
		v[base+0] = clampNorm(float64(s.Life), 40.0)
		v[base+1] = clampNorm(float64(s.PoisonCounters), 10.0)
		v[base+2] = maxCmdrDamageTaken(s)
		v[base+3] = clampNorm(float64(len(s.Hand)), 10.0)
		v[base+4] = clampNorm(float64(len(s.Library)), 99.0)
		v[base+5] = clampNorm(float64(len(s.Graveyard)), 40.0)
		v[base+6] = clampNorm(float64(len(s.Exile)), 20.0)

		lands, creatures, artifacts, enchantments, planeswalkers := 0, 0, 0, 0, 0
		totalPow, totalTough := 0, 0
		cmdrOnBF := false
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil {
				continue
			}
			tl := strings.ToLower(p.Card.TypeLine)
			if strings.Contains(tl, "land") {
				lands++
			}
			if p.IsCreature() {
				creatures++
				totalPow += gs.PowerOf(p)
				totalTough += gs.ToughnessOf(p)
			}
			if strings.Contains(tl, "artifact") {
				artifacts++
			}
			if strings.Contains(tl, "enchantment") {
				enchantments++
			}
			if strings.Contains(tl, "planeswalker") {
				planeswalkers++
			}
			for _, cn := range s.CommanderNames {
				if strings.EqualFold(p.Card.DisplayName(), cn) {
					cmdrOnBF = true
				}
			}
		}

		v[base+7] = clampNorm(float64(lands), 12.0)
		v[base+8] = clampNorm(float64(creatures), 15.0)
		v[base+9] = clampNorm(float64(artifacts), 10.0)
		v[base+10] = clampNorm(float64(enchantments), 8.0)
		v[base+11] = clampNorm(float64(planeswalkers), 3.0)
		v[base+12] = clampNorm(float64(totalPow), 40.0)
		v[base+13] = clampNorm(float64(totalTough), 40.0)
		v[base+14] = clampNorm(float64(gameengine.AvailableManaEstimate(gs, s)), 15.0)

		if cmdrOnBF {
			v[base+15] = 1.0
		}
		maxTax := 0
		if s.CommanderCastCounts != nil {
			for _, t := range s.CommanderCastCounts {
				if t > maxTax {
					maxTax = t
				}
			}
		}
		v[base+16] = clampNorm(float64(maxTax)*2.0, 12.0)
		v[base+17] = clampNorm(float64(s.SpellsCastThisTurn), 5.0)

		hasCounter, hasRemoval := false, false
		for _, c := range s.Hand {
			if c == nil {
				continue
			}
			if slot == 0 { // only perspective seat's hand is known
				if gameengine.CardHasCounterSpell(c) {
					hasCounter = true
				}
				ot := gameengine.OracleTextLower(c)
				if strings.Contains(ot, "destroy") || strings.Contains(ot, "exile target") {
					hasRemoval = true
				}
			}
		}
		if hasCounter {
			v[base+18] = 1.0
		}
		if hasRemoval {
			v[base+19] = 1.0
		}
		if s.Lost {
			v[base+20] = 1.0
		}
		if gs.Active == actualSeat {
			v[base+21] = 1.0
		}
	}

	// Global features.
	gBase := maxSeats * seatFeatures
	v[gBase+0] = clampNorm(float64(gs.Turn), 80.0)
	v[gBase+1] = clampNorm(float64(len(gs.Stack)), 5.0)
	// Game stage: early (0-5) = 0.0, mid (6-15) = 0.5, late (16+) = 1.0.
	switch {
	case gs.Turn <= 5:
		v[gBase+2] = 0.0
	case gs.Turn <= 15:
		v[gBase+2] = 0.5
	default:
		v[gBase+2] = 1.0
	}
	// Living opponents count (helps model understand remaining competition).
	living := 0
	for i, s := range gs.Seats {
		if s != nil && !s.Lost && i != perspectiveSeat {
			living++
		}
	}
	v[gBase+3] = clampNorm(float64(living), 3.0)

	return v
}

func clampNorm(val, maxVal float64) float64 {
	if maxVal <= 0 {
		return 0
	}
	n := val / maxVal
	if n < 0 {
		return 0
	}
	if n > 1 {
		return 1
	}
	return n
}

func maxCmdrDamageTaken(s *gameengine.Seat) float64 {
	if s.CommanderDamage == nil {
		return 0
	}
	maxDmg := 0
	for _, cmdrMap := range s.CommanderDamage {
		for _, dmg := range cmdrMap {
			if dmg > maxDmg {
				maxDmg = dmg
			}
		}
	}
	return clampNorm(float64(maxDmg), 21.0)
}

// TrainingSample is one (state, outcome) pair for supervised learning.
// The model learns to predict Placement from StateVector.
type TrainingSample struct {
	State     StateVector `json:"state"`
	Placement float64     `json:"placement"` // 1.0 = 1st place, 0.0 = last
	Turn      int         `json:"turn"`      // turn at which state was captured
	GameTurn  int         `json:"game_turn"` // total turns in the game
}

// TrainingCollector accumulates (state, outcome) samples during grinder
// games. It snapshots the state at regular intervals during the game,
// then stamps all snapshots with the final placement when the game ends.
type TrainingCollector struct {
	interval int // snapshot every N turns
	buf      []pendingSample
}

type pendingSample struct {
	state StateVector
	turn  int
}

func NewTrainingCollector(snapshotInterval int) *TrainingCollector {
	if snapshotInterval <= 0 {
		snapshotInterval = 5
	}
	return &TrainingCollector{
		interval: snapshotInterval,
		buf:      make([]pendingSample, 0, 16),
	}
}

// Snapshot captures the current state if this turn is a snapshot turn.
func (tc *TrainingCollector) Snapshot(gs *gameengine.GameState, perspectiveSeat int) {
	if gs.Turn%tc.interval != 0 && gs.Turn != 1 {
		return
	}
	tc.buf = append(tc.buf, pendingSample{
		state: EncodeState(gs, perspectiveSeat),
		turn:  gs.Turn,
	})
}

// Finalize stamps all buffered snapshots with the game outcome and
// returns the training samples. Call after the game ends.
func (tc *TrainingCollector) Finalize(placement float64, gameTurns int) []TrainingSample {
	samples := make([]TrainingSample, len(tc.buf))
	for i, ps := range tc.buf {
		samples[i] = TrainingSample{
			State:     ps.state,
			Placement: placement,
			Turn:      ps.turn,
			GameTurn:  gameTurns,
		}
	}
	tc.buf = tc.buf[:0]
	return samples
}

// AppendSamples writes training samples to a JSONL file (append mode).
func AppendSamples(path string, samples []TrainingSample) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, s := range samples {
		if err := enc.Encode(s); err != nil {
			return err
		}
	}
	return nil
}

// NeuralLayer is one dense layer: output = activation(Weights @ input + Biases).
type NeuralLayer struct {
	Weights [][]float64 // [out_dim][in_dim]
	Biases  []float64   // [out_dim]
}

// NeuralEvaluator wraps a trained model for use as a position evaluator.
// Supports N hidden layers. Inference is pure Go matrix math — no external deps.
type NeuralEvaluator struct {
	Layers []NeuralLayer
	Arch   string
}

// Evaluate runs forward inference on the state vector and returns a
// score in [0, 1] (estimated win probability).
func (ne *NeuralEvaluator) Evaluate(v StateVector) float64 {
	if len(ne.Layers) == 0 {
		return 0.5
	}

	current := v[:]
	for i, layer := range ne.Layers {
		outDim := len(layer.Biases)
		next := make([]float64, outDim)
		for j := 0; j < outDim; j++ {
			sum := layer.Biases[j]
			for k, w := range layer.Weights[j] {
				if k < len(current) {
					sum += w * current[k]
				}
			}
			if i < len(ne.Layers)-1 {
				next[j] = relu(sum)
			} else {
				next[j] = sum
			}
		}
		current = next
	}

	if len(current) > 0 {
		return sigmoid(current[0])
	}
	return 0.5
}

func relu(x float64) float64 {
	if x > 0 {
		return x
	}
	return 0
}

func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// JSON serialization — supports both legacy "mlp-1h" and new "mlp" formats.
type neuralLayerJSON struct {
	Weights [][]float64 `json:"weights"`
	Biases  []float64   `json:"biases"`
}

type neuralModelJSON struct {
	Arch   string            `json:"arch"`
	Layers []neuralLayerJSON `json:"layers,omitempty"`
	// Legacy single-layer fields (backward compat with "mlp-1h").
	W1 [][]float64 `json:"w1,omitempty"`
	B1 []float64   `json:"b1,omitempty"`
	W2 []float64   `json:"w2,omitempty"`
	B2 float64     `json:"b2,omitempty"`
}

func LoadNeuralEvaluator(path string) (*NeuralEvaluator, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m neuralModelJSON
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	ne := &NeuralEvaluator{Arch: m.Arch}

	if len(m.Layers) > 0 {
		ne.Layers = make([]NeuralLayer, len(m.Layers))
		for i, l := range m.Layers {
			ne.Layers[i] = NeuralLayer{Weights: l.Weights, Biases: l.Biases}
		}
	} else if len(m.W1) > 0 {
		// Legacy "mlp-1h": convert W1/B1 + W2/B2 into 2-layer model.
		ne.Layers = []NeuralLayer{
			{Weights: m.W1, Biases: m.B1},
			{Weights: [][]float64{m.W2}, Biases: []float64{m.B2}},
		}
	}

	return ne, nil
}

// TryLoadNeuralEvaluator attempts to load a trained model from the
// given path. Returns nil if the file doesn't exist or can't be parsed
// — the caller should treat nil as "no neural model available."
func TryLoadNeuralEvaluator(path string) *NeuralEvaluator {
	ne, err := LoadNeuralEvaluator(path)
	if err != nil {
		return nil
	}
	if len(ne.Layers) == 0 {
		return nil
	}
	return ne
}
