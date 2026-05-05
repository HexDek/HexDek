package hat

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CurseDNA holds the evolvable personality parameters for a single deck.
type CurseDNA struct {
	DeckKey         string  `json:"deck_key"`
	Generation      int     `json:"generation"`
	GamesPlayed     int     `json:"games_played"`
	Fitness         float64 `json:"fitness"`

	Aggression      float64 `json:"aggression"`        // attack threshold [0,1]
	ComboPat        float64 `json:"combo_patience"`     // how long to wait for full combo [0,1]
	ThreatParanoia  float64 `json:"threat_paranoia"`    // opponent threat weighting [0,1]
	ResourceGreed   float64 `json:"resource_greed"`     // card advantage vs tempo [0,1]
	PoliticalMemory float64 `json:"political_memory"`   // grudge/gratitude decay rate [0,1]
	DrainAffinity   float64 `json:"drain_affinity"`     // aristocrats drain engine weighting [0,1]
	ArtifactAffinity float64 `json:"artifact_affinity"` // artifact synergy weighting [0,1]
}

const (
	CursePopSize    = 8
	CurseEvolveAt   = 100  // games per deck before evolution step
	CurseMutationSD = 0.05 // gaussian mutation standard deviation
	CurseKillCount  = 2    // bottom N killed per evolution
	curseImmigrationInterval = 10 // every N generations, inject fresh random DNA
	dimStatsAlpha    = 0.04 // EMA decay for dimension stats (~25-game half-life)
	dimStatsMinN     = 20   // minimum games before applying corrections
)

// DimensionStats tracks EMA correlations between evaluator dimension scores
// and game outcomes. After enough games, WeightCorrections returns per-dimension
// multiplicative adjustments to bias the evaluator toward dimensions that
// actually predict wins for this deck.
type DimensionStats struct {
	MeanDim  [NumDimensions]float64 `json:"mean_dim"`
	MeanSqDim [NumDimensions]float64 `json:"mean_sq_dim"`
	CrossDim [NumDimensions]float64 `json:"cross_dim"`
	MeanScore float64               `json:"mean_score"`
	MeanSqScore float64             `json:"mean_sq_score"`
	N         int                    `json:"n"`
}

// RecordGame updates the running EMA statistics with one game's data.
func (ds *DimensionStats) RecordGame(dimMeans [NumDimensions]float64, outcome float64) {
	ds.N++
	alpha := dimStatsAlpha
	if ds.N <= 1 {
		for d := 0; d < NumDimensions; d++ {
			ds.MeanDim[d] = dimMeans[d]
			ds.MeanSqDim[d] = dimMeans[d] * dimMeans[d]
			ds.CrossDim[d] = dimMeans[d] * outcome
		}
		ds.MeanScore = outcome
		ds.MeanSqScore = outcome * outcome
		return
	}
	for d := 0; d < NumDimensions; d++ {
		ds.MeanDim[d] += alpha * (dimMeans[d] - ds.MeanDim[d])
		ds.MeanSqDim[d] += alpha * (dimMeans[d]*dimMeans[d] - ds.MeanSqDim[d])
		ds.CrossDim[d] += alpha * (dimMeans[d]*outcome - ds.CrossDim[d])
	}
	ds.MeanScore += alpha * (outcome - ds.MeanScore)
	ds.MeanSqScore += alpha * (outcome*outcome - ds.MeanSqScore)
}

// Correlation returns Pearson's r for dimension d versus game outcome.
func (ds *DimensionStats) Correlation(d int) float64 {
	if ds.N < dimStatsMinN || d < 0 || d >= NumDimensions {
		return 0
	}
	varD := ds.MeanSqDim[d] - ds.MeanDim[d]*ds.MeanDim[d]
	varO := ds.MeanSqScore - ds.MeanScore*ds.MeanScore
	if varD < 1e-10 || varO < 1e-10 {
		return 0
	}
	cov := ds.CrossDim[d] - ds.MeanDim[d]*ds.MeanScore
	r := cov / math.Sqrt(varD*varO)
	if r < -1 {
		r = -1
	}
	if r > 1 {
		r = 1
	}
	return r
}

// WeightCorrections returns per-dimension multiplicative corrections
// derived from outcome correlations. Positive correlation → boost weight,
// negative → suppress. Max swing: ±20%.
func (ds *DimensionStats) WeightCorrections() [NumDimensions]float64 {
	var corr [NumDimensions]float64
	for d := 0; d < NumDimensions; d++ {
		corr[d] = 1.0
		if ds.N >= dimStatsMinN {
			corr[d] = 1.0 + ds.Correlation(d)*0.20
		}
	}
	return corr
}

// CursePool maintains a population of DNA variants for a single deck.
type CursePool struct {
	DeckKey    string                      `json:"deck_key"`
	Population [CursePopSize]CurseDNA    `json:"population"`
	GameCount  int                         `json:"game_count"`
	TotalGames int                         `json:"total_games"`
	Bracket    int                         `json:"bracket"`    // power bracket (1-5) for fitness normalization
	GenCount   int                         `json:"gen_count"`  // total generations evolved
	DimStats   DimensionStats              `json:"dim_stats"`  // T3.2 outcome-correlated dimension learning
	rng        *rand.Rand // not serialized; caller injects via InitPool or SetRNG
}

// SetRNG assigns the random source used for selection and evolution.
// Must be called after deserializing a pool from JSON.
func (pool *CursePool) SetRNG(rng *rand.Rand) {
	pool.rng = rng
}

// SelectForGame picks a DNA variant using fitness-proportional selection.
func (pool *CursePool) SelectForGame() (dna *CurseDNA, idx int) {
	totalFitness := 0.0
	for i := range pool.Population {
		f := pool.Population[i].Fitness
		if f < 0.01 {
			f = 0.01 // floor to prevent starvation
		}
		totalFitness += f
	}
	roll := pool.rng.Float64() * totalFitness
	cumulative := 0.0
	for i := range pool.Population {
		f := pool.Population[i].Fitness
		if f < 0.01 {
			f = 0.01
		}
		cumulative += f
		if roll <= cumulative {
			return &pool.Population[i], i
		}
	}
	return &pool.Population[CursePopSize-1], CursePopSize - 1
}

// PlacementScore converts a 4-player finish position to a fitness value.
// 1st=1.0, 2nd=0.5, 3rd=0.2, 4th=0.0. Reduces noise vs binary win/loss.
func PlacementScore(placement, totalPlayers int) float64 {
	if totalPlayers <= 1 || placement <= 0 {
		return 0.5
	}
	return 1.0 - float64(placement-1)/float64(totalPlayers-1)
}

// RecordResult updates fitness after a game and possibly triggers evolution.
// score is a [0,1] fitness value — use PlacementScore() for 4-player games
// instead of binary 0/1 to reduce variance.
func (pool *CursePool) RecordResult(idx int, score float64) {
	dna := &pool.Population[idx]
	dna.GamesPlayed++
	pool.GameCount++
	pool.TotalGames++

	// Bracket-normalize: compare against expected performance for this bracket.
	normalized := score
	if pool.Bracket > 0 {
		expected := expectedBracketScore(pool.Bracket)
		if expected > 0 {
			normalized = score / expected
			if normalized > 2.0 {
				normalized = 2.0
			}
		}
	}

	// Rolling average fitness (EMA).
	if dna.GamesPlayed <= 1 {
		dna.Fitness = normalized
	} else {
		alpha := 2.0 / (float64(min(dna.GamesPlayed, 50)) + 1.0)
		dna.Fitness = dna.Fitness*(1-alpha) + normalized*alpha
	}

	if pool.GameCount >= CurseEvolveAt {
		pool.evolve()
	}
}

// expectedBracketScore returns the expected average placement score for a
// bracket tier. Derived from grinder data: B1 WR≈28.9%, B5 WR≈19.3%.
func expectedBracketScore(bracket int) float64 {
	switch bracket {
	case 1:
		return 0.58
	case 2:
		return 0.52
	case 3:
		return 0.48
	case 4:
		return 0.44
	case 5:
		return 0.39
	default:
		return 0.50
	}
}

// evolve runs one generation of genetic selection on the population.
// Sort by fitness, kill bottom CurseKillCount, clone top CurseKillCount
// into those slots, mutate clones, clamp, reset game counter.
// Every curseImmigrationInterval generations, inject fresh random DNA
// to prevent convergent collapse.
func (pool *CursePool) evolve() {
	pool.GenCount++

	// Build index sorted by fitness (ascending).
	indices := make([]int, CursePopSize)
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(a, b int) bool {
		return pool.Population[indices[a]].Fitness < pool.Population[indices[b]].Fitness
	})

	// Kill bottom, clone top into those slots.
	for k := 0; k < CurseKillCount; k++ {
		loserIdx := indices[k]
		donorIdx := indices[CursePopSize-1-k]

		clone := pool.Population[donorIdx]
		clone.GamesPlayed = 0
		clone.Generation++

		// Mutate each evolvable parameter.
		clone.Aggression = clampUnit(clone.Aggression + pool.rng.NormFloat64()*CurseMutationSD)
		clone.ComboPat = clampUnit(clone.ComboPat + pool.rng.NormFloat64()*CurseMutationSD)
		clone.ThreatParanoia = clampUnit(clone.ThreatParanoia + pool.rng.NormFloat64()*CurseMutationSD)
		clone.ResourceGreed = clampUnit(clone.ResourceGreed + pool.rng.NormFloat64()*CurseMutationSD)
		clone.PoliticalMemory = clampUnit(clone.PoliticalMemory + pool.rng.NormFloat64()*CurseMutationSD)
		clone.DrainAffinity = clampUnit(clone.DrainAffinity + pool.rng.NormFloat64()*CurseMutationSD)
		clone.ArtifactAffinity = clampUnit(clone.ArtifactAffinity + pool.rng.NormFloat64()*CurseMutationSD)

		pool.Population[loserIdx] = clone
	}

	// Immigration: every N generations, replace one mid-tier individual
	// with fresh random DNA. Prevents convergent collapse.
	if pool.GenCount%curseImmigrationInterval == 0 {
		immigrantIdx := indices[CurseKillCount] // lowest non-killed slot
		pool.Population[immigrantIdx] = CurseDNA{
			DeckKey:          pool.DeckKey,
			Aggression:       pool.rng.Float64(),
			ComboPat:         pool.rng.Float64(),
			ThreatParanoia:   pool.rng.Float64(),
			ResourceGreed:    pool.rng.Float64(),
			PoliticalMemory:  pool.rng.Float64(),
			DrainAffinity:    pool.rng.Float64(),
			ArtifactAffinity: pool.rng.Float64(),
			Fitness:          0.25,
		}
	}

	pool.GameCount = 0
}

// InitPool creates a fresh population with random parameters for a new deck.
func InitPool(deckKey string, rng *rand.Rand) CursePool {
	pool := CursePool{DeckKey: deckKey, rng: rng}
	for i := range pool.Population {
		pool.Population[i] = CurseDNA{
			DeckKey:          deckKey,
			Aggression:       rng.Float64(),
			ComboPat:         rng.Float64(),
			ThreatParanoia:   rng.Float64(),
			ResourceGreed:    rng.Float64(),
			PoliticalMemory:  rng.Float64(),
			DrainAffinity:    rng.Float64(),
			ArtifactAffinity: rng.Float64(),
			Fitness:          0.25, // baseline expected winrate in 4-player
		}
	}
	return pool
}

// InitPoolWithBracket creates a fresh population with bracket info for normalization.
func InitPoolWithBracket(deckKey string, bracket int, rng *rand.Rand) CursePool {
	pool := InitPool(deckKey, rng)
	pool.Bracket = bracket
	return pool
}

// clampUnit constrains v to [0.0, 1.0].
func clampUnit(v float64) float64 {
	if v < 0.0 {
		return 0.0
	}
	if v > 1.0 {
		return 1.0
	}
	return v
}

// sanitizeKey makes a deck key safe for use as a filename.
func sanitizeKey(key string) string {
	r := strings.NewReplacer("/", "_", "\\", "_", ":", "_", " ", "_")
	return r.Replace(key)
}

// SavePool writes a single pool to disk as JSON.
func SavePool(dir string, pool *CursePool) error {
	os.MkdirAll(dir, 0755)
	fname := filepath.Join(dir, sanitizeKey(pool.DeckKey)+".json")
	data, err := json.Marshal(pool)
	if err != nil {
		return err
	}
	return os.WriteFile(fname, data, 0644)
}

// LoadPool reads a single pool from disk and restores it.
func LoadPool(dir, deckKey string, rng *rand.Rand) (*CursePool, error) {
	fname := filepath.Join(dir, sanitizeKey(deckKey)+".json")
	data, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	var pool CursePool
	if err := json.Unmarshal(data, &pool); err != nil {
		return nil, err
	}
	pool.rng = rng
	return &pool, nil
}

// SaveAllPools writes every pool to disk.
func SaveAllPools(dir string, pools map[string]*CursePool) error {
	os.MkdirAll(dir, 0755)
	for _, pool := range pools {
		if err := SavePool(dir, pool); err != nil {
			return err
		}
	}
	return nil
}

// LoadAllPools reads all pool files from a directory.
func LoadAllPools(dir string, rng *rand.Rand) (map[string]*CursePool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*CursePool), nil
		}
		return nil, err
	}
	pools := make(map[string]*CursePool, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var pool CursePool
		if err := json.Unmarshal(data, &pool); err != nil {
			continue
		}
		pool.rng = rng
		pools[pool.DeckKey] = &pool
	}
	return pools, nil
}
