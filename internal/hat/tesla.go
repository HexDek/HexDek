package hat

import "math"

// Tesla Causal Graphs — extract the causal pivot from a completed game.
//
// A causal pivot is the turn where the game's outcome was most strongly
// determined. Instead of labeling every training sample with just the
// final placement, we also label with pivot distance — giving the neural
// model signal about WHEN things changed, not just WHO won.
//
// The pivot is identified by finding the turn with the largest eval-score
// swing favoring the eventual winner relative to the field average.

// CausalPivot describes the single most game-deciding moment.
type CausalPivot struct {
	Turn       int     // turn number of the pivot
	WinnerSeat int     // seat that benefited most
	DeltaScore float64 // magnitude of the swing (0-1 scale)
}

// PivotLabel extends a training sample with causal pivot metadata.
type PivotLabel struct {
	PivotTurn     int     `json:"pivot_turn"`
	PivotDistance float64 `json:"pivot_distance"` // 0.0 = at pivot, 1.0 = far from pivot
	IsPostPivot   bool    `json:"is_post_pivot"`
}

// ExtractPivot finds the causal pivot from a sequence of per-seat eval
// snapshots. evalHistory[turn][seat] = eval score at that turn.
// winnerSeat is the seat that placed 1st.
func ExtractPivot(evalHistory map[int][]float64, winnerSeat int, gameTurns int) CausalPivot {
	if len(evalHistory) < 2 || winnerSeat < 0 {
		return CausalPivot{Turn: gameTurns / 2, WinnerSeat: winnerSeat}
	}

	// Collect sorted turn numbers.
	turns := make([]int, 0, len(evalHistory))
	for t := range evalHistory {
		turns = append(turns, t)
	}
	sortInts(turns)

	bestTurn := turns[len(turns)/2]
	bestDelta := 0.0

	for i := 1; i < len(turns); i++ {
		prev := evalHistory[turns[i-1]]
		curr := evalHistory[turns[i]]
		if winnerSeat >= len(prev) || winnerSeat >= len(curr) {
			continue
		}

		// Winner's score delta.
		winnerDelta := curr[winnerSeat] - prev[winnerSeat]

		// Field average delta (excluding winner).
		fieldDeltaSum := 0.0
		fieldCount := 0
		for s := 0; s < len(curr) && s < len(prev); s++ {
			if s == winnerSeat {
				continue
			}
			fieldDeltaSum += curr[s] - prev[s]
			fieldCount++
		}
		fieldAvgDelta := 0.0
		if fieldCount > 0 {
			fieldAvgDelta = fieldDeltaSum / float64(fieldCount)
		}

		// Relative swing: how much better the winner did vs the field.
		swing := winnerDelta - fieldAvgDelta
		if swing > bestDelta {
			bestDelta = swing
			bestTurn = turns[i]
		}
	}

	return CausalPivot{
		Turn:       bestTurn,
		WinnerSeat: winnerSeat,
		DeltaScore: bestDelta,
	}
}

// LabelSamplesWithPivot adds pivot distance labels to training samples.
// Samples near the pivot get lower distance (more weight during training);
// samples far from the pivot get higher distance.
func LabelSamplesWithPivot(samples []TrainingSample, pivot CausalPivot) []PivotLabel {
	labels := make([]PivotLabel, len(samples))
	for i, s := range samples {
		dist := math.Abs(float64(s.Turn-pivot.Turn)) / math.Max(float64(s.GameTurn), 1.0)
		if dist > 1.0 {
			dist = 1.0
		}
		labels[i] = PivotLabel{
			PivotTurn:     pivot.Turn,
			PivotDistance: dist,
			IsPostPivot:   s.Turn >= pivot.Turn,
		}
	}
	return labels
}

// PivotEnrichedSample is a training sample with causal pivot metadata,
// written to the JSONL file for the Python training script.
type PivotEnrichedSample struct {
	State         StateVector `json:"state"`
	Placement     float64     `json:"placement"`
	Turn          int         `json:"turn"`
	GameTurn      int         `json:"game_turn"`
	PivotTurn     int         `json:"pivot_turn"`
	PivotDistance float64     `json:"pivot_distance"`
	IsPostPivot   bool        `json:"is_post_pivot"`
}

// EnrichSamples merges training samples with pivot labels into
// pivot-enriched samples for export.
func EnrichSamples(samples []TrainingSample, labels []PivotLabel) []PivotEnrichedSample {
	out := make([]PivotEnrichedSample, len(samples))
	for i, s := range samples {
		out[i] = PivotEnrichedSample{
			State:     s.State,
			Placement: s.Placement,
			Turn:      s.Turn,
			GameTurn:  s.GameTurn,
		}
		if i < len(labels) {
			out[i].PivotTurn = labels[i].PivotTurn
			out[i].PivotDistance = labels[i].PivotDistance
			out[i].IsPostPivot = labels[i].IsPostPivot
		}
	}
	return out
}

// EvalSnapshotCollector captures per-seat eval scores at each snapshot
// turn during a game, feeding Tesla pivot extraction after game end.
type EvalSnapshotCollector struct {
	history map[int][]float64 // turn → [seat0_eval, seat1_eval, ...]
}

func NewEvalSnapshotCollector() *EvalSnapshotCollector {
	return &EvalSnapshotCollector{history: make(map[int][]float64)}
}

// Record stores the eval scores for all seats at the given turn.
func (c *EvalSnapshotCollector) Record(turn int, scores []float64) {
	c.history[turn] = append([]float64(nil), scores...)
}

// History returns the collected eval snapshots.
func (c *EvalSnapshotCollector) History() map[int][]float64 {
	return c.history
}

func sortInts(a []int) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j-1] > a[j]; j-- {
			a[j-1], a[j] = a[j], a[j-1]
		}
	}
}
