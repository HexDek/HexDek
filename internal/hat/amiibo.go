package hat

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// AmiiboDNA holds the evolvable personality parameters for a single deck.
type AmiiboDNA struct {
	DeckKey         string  `json:"deck_key"`
	Generation      int     `json:"generation"`
	GamesPlayed     int     `json:"games_played"`
	Fitness         float64 `json:"fitness"`

	Aggression      float64 `json:"aggression"`        // attack threshold [0,1]
	ComboPat        float64 `json:"combo_patience"`     // how long to wait for full combo [0,1]
	ThreatParanoia  float64 `json:"threat_paranoia"`    // opponent threat weighting [0,1]
	ResourceGreed   float64 `json:"resource_greed"`     // card advantage vs tempo [0,1]
	PoliticalMemory float64 `json:"political_memory"`   // grudge/gratitude decay rate [0,1]
}

const (
	AmiiboPopSize    = 8
	AmiiboEvolveAt   = 100  // games per deck before evolution step
	AmiiboMutationSD = 0.05 // gaussian mutation standard deviation
	AmiiboKillCount  = 2    // bottom N killed per evolution
)

// AmiiboPool maintains a population of DNA variants for a single deck.
type AmiiboPool struct {
	DeckKey    string                      `json:"deck_key"`
	Population [AmiiboPopSize]AmiiboDNA    `json:"population"`
	GameCount  int                         `json:"game_count"`
	rng        *rand.Rand // not serialized; caller injects via InitPool or SetRNG
}

// SetRNG assigns the random source used for selection and evolution.
// Must be called after deserializing a pool from JSON.
func (pool *AmiiboPool) SetRNG(rng *rand.Rand) {
	pool.rng = rng
}

// SelectForGame picks a DNA variant using fitness-proportional selection.
func (pool *AmiiboPool) SelectForGame() (dna *AmiiboDNA, idx int) {
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
	return &pool.Population[AmiiboPopSize-1], AmiiboPopSize - 1
}

// RecordResult updates fitness after a game and possibly triggers evolution.
func (pool *AmiiboPool) RecordResult(idx int, won bool) {
	dna := &pool.Population[idx]
	dna.GamesPlayed++
	pool.GameCount++

	// Rolling average fitness (EMA).
	result := 0.0
	if won {
		result = 1.0
	}
	if dna.GamesPlayed <= 1 {
		dna.Fitness = result
	} else {
		alpha := 2.0 / (float64(min(dna.GamesPlayed, 50)) + 1.0)
		dna.Fitness = dna.Fitness*(1-alpha) + result*alpha
	}

	if pool.GameCount >= AmiiboEvolveAt {
		pool.evolve()
	}
}

// evolve runs one generation of genetic selection on the population.
// Sort by fitness, kill bottom AmiiboKillCount, clone top AmiiboKillCount
// into those slots, mutate clones, clamp, reset game counter.
func (pool *AmiiboPool) evolve() {
	// Build index sorted by fitness (ascending).
	indices := make([]int, AmiiboPopSize)
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(a, b int) bool {
		return pool.Population[indices[a]].Fitness < pool.Population[indices[b]].Fitness
	})

	// Kill bottom, clone top into those slots.
	for k := 0; k < AmiiboKillCount; k++ {
		loserIdx := indices[k]
		donorIdx := indices[AmiiboPopSize-1-k]

		clone := pool.Population[donorIdx]
		clone.GamesPlayed = 0
		clone.Generation++

		// Mutate each evolvable parameter.
		clone.Aggression = clampUnit(clone.Aggression + pool.rng.NormFloat64()*AmiiboMutationSD)
		clone.ComboPat = clampUnit(clone.ComboPat + pool.rng.NormFloat64()*AmiiboMutationSD)
		clone.ThreatParanoia = clampUnit(clone.ThreatParanoia + pool.rng.NormFloat64()*AmiiboMutationSD)
		clone.ResourceGreed = clampUnit(clone.ResourceGreed + pool.rng.NormFloat64()*AmiiboMutationSD)
		clone.PoliticalMemory = clampUnit(clone.PoliticalMemory + pool.rng.NormFloat64()*AmiiboMutationSD)

		pool.Population[loserIdx] = clone
	}

	pool.GameCount = 0
}

// InitPool creates a fresh population with random parameters for a new deck.
func InitPool(deckKey string, rng *rand.Rand) AmiiboPool {
	pool := AmiiboPool{DeckKey: deckKey, rng: rng}
	for i := range pool.Population {
		pool.Population[i] = AmiiboDNA{
			DeckKey:         deckKey,
			Aggression:      rng.Float64(),
			ComboPat:        rng.Float64(),
			ThreatParanoia:  rng.Float64(),
			ResourceGreed:   rng.Float64(),
			PoliticalMemory: rng.Float64(),
			Fitness:         0.25, // baseline expected winrate in 4-player
		}
	}
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
func SavePool(dir string, pool *AmiiboPool) error {
	os.MkdirAll(dir, 0755)
	fname := filepath.Join(dir, sanitizeKey(pool.DeckKey)+".json")
	data, err := json.Marshal(pool)
	if err != nil {
		return err
	}
	return os.WriteFile(fname, data, 0644)
}

// LoadPool reads a single pool from disk and restores it.
func LoadPool(dir, deckKey string, rng *rand.Rand) (*AmiiboPool, error) {
	fname := filepath.Join(dir, sanitizeKey(deckKey)+".json")
	data, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	var pool AmiiboPool
	if err := json.Unmarshal(data, &pool); err != nil {
		return nil, err
	}
	pool.rng = rng
	return &pool, nil
}

// SaveAllPools writes every pool to disk.
func SaveAllPools(dir string, pools map[string]*AmiiboPool) error {
	os.MkdirAll(dir, 0755)
	for _, pool := range pools {
		if err := SavePool(dir, pool); err != nil {
			return err
		}
	}
	return nil
}

// LoadAllPools reads all pool files from a directory.
func LoadAllPools(dir string, rng *rand.Rand) (map[string]*AmiiboPool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*AmiiboPool), nil
		}
		return nil, err
	}
	pools := make(map[string]*AmiiboPool, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var pool AmiiboPool
		if err := json.Unmarshal(data, &pool); err != nil {
			continue
		}
		pool.rng = rng
		pools[pool.DeckKey] = &pool
	}
	return pools, nil
}
