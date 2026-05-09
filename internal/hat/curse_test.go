package hat

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestInitPool(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	pool := InitPool("test/deck", rng)

	if pool.DeckKey != "test/deck" {
		t.Errorf("expected deck key 'test/deck', got %q", pool.DeckKey)
	}
	for i, dna := range pool.Population {
		if dna.Fitness != 0.25 {
			t.Errorf("pop[%d] fitness = %f, want 0.25", i, dna.Fitness)
		}
		if dna.Aggression < 0 || dna.Aggression > 1 {
			t.Errorf("pop[%d] Aggression out of range: %f", i, dna.Aggression)
		}
	}
}

func TestSelectForGame(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	pool := InitPool("test", rng)

	pool.Population[0].Fitness = 0.9
	for i := 1; i < CursePopSize; i++ {
		pool.Population[i].Fitness = 0.01
	}

	hitZero := 0
	for i := 0; i < 100; i++ {
		_, idx := pool.SelectForGame()
		if idx == 0 {
			hitZero++
		}
	}
	if hitZero < 50 {
		t.Errorf("expected high-fitness member selected >50%%, got %d%%", hitZero)
	}
}

func TestRecordResult_UpdatesFitness(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	pool := InitPool("test", rng)

	pool.RecordResult(0, 1.0) // 1st place
	if pool.Population[0].Fitness != 1.0 {
		t.Errorf("fitness after first win = %f, want 1.0", pool.Population[0].Fitness)
	}

	pool.RecordResult(0, 0.0) // 4th place
	if pool.Population[0].Fitness >= 1.0 {
		t.Errorf("fitness should decrease after loss, got %f", pool.Population[0].Fitness)
	}
}

func TestPlacementScore(t *testing.T) {
	if s := PlacementScore(1, 4); s != 1.0 {
		t.Errorf("1st of 4 = %f, want 1.0", s)
	}
	if s := PlacementScore(4, 4); s != 0.0 {
		t.Errorf("4th of 4 = %f, want 0.0", s)
	}
	if s := PlacementScore(2, 4); s < 0.3 || s > 0.7 {
		t.Errorf("2nd of 4 = %f, want ~0.67", s)
	}
}

func TestEvolution_TriggersAt100(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	pool := InitPool("test", rng)

	// Give population a clear fitness gradient so evolution has something to work with.
	for i := 0; i < CursePopSize; i++ {
		pool.Population[i].Fitness = float64(i+1) / float64(CursePopSize)
	}

	// Record 100 games to trigger evolution.
	for i := 0; i < CurseEvolveAt; i++ {
		score := 0.0
		if i%2 == 0 {
			score = 1.0
		}
		pool.RecordResult(i%CursePopSize, score)
	}

	if pool.GameCount != 0 {
		t.Errorf("GameCount should reset after evolution, got %d", pool.GameCount)
	}
}

func TestClampUnit(t *testing.T) {
	if v := clampUnit(-0.5); v != 0.0 {
		t.Errorf("clampUnit(-0.5) = %f, want 0.0", v)
	}
	if v := clampUnit(1.5); v != 1.0 {
		t.Errorf("clampUnit(1.5) = %f, want 1.0", v)
	}
	if v := clampUnit(0.5); v != 0.5 {
		t.Errorf("clampUnit(0.5) = %f, want 0.5", v)
	}
}

func TestSaveAndLoadPool(t *testing.T) {
	dir := t.TempDir()
	rng := rand.New(rand.NewSource(42))
	pool := InitPool("owner/cool-deck", rng)

	pool.Population[0].Fitness = 0.75
	pool.Population[0].Aggression = 0.3
	pool.GameCount = 50

	if err := SavePool(dir, &pool); err != nil {
		t.Fatalf("SavePool: %v", err)
	}

	fname := filepath.Join(dir, "owner_cool-deck.json")
	if _, err := os.Stat(fname); err != nil {
		t.Fatalf("expected file %s, got error: %v", fname, err)
	}

	loaded, err := LoadPool(dir, "owner/cool-deck", rng)
	if err != nil {
		t.Fatalf("LoadPool: %v", err)
	}

	if loaded.DeckKey != "owner/cool-deck" {
		t.Errorf("loaded deck key = %q, want 'owner/cool-deck'", loaded.DeckKey)
	}
	if loaded.GameCount != 50 {
		t.Errorf("loaded game count = %d, want 50", loaded.GameCount)
	}
	if loaded.Population[0].Fitness != 0.75 {
		t.Errorf("loaded fitness = %f, want 0.75", loaded.Population[0].Fitness)
	}
	if loaded.Population[0].Aggression != 0.3 {
		t.Errorf("loaded aggression = %f, want 0.3", loaded.Population[0].Aggression)
	}
}

func TestSaveAndLoadAllPools(t *testing.T) {
	dir := t.TempDir()
	rng := rand.New(rand.NewSource(42))

	pools := make(map[string]*CursePool)
	for _, key := range []string{"deck-a", "deck-b", "deck-c"} {
		p := InitPool(key, rng)
		pools[key] = &p
	}

	if err := SaveAllPools(dir, pools); err != nil {
		t.Fatalf("SaveAllPools: %v", err)
	}

	loaded, err := LoadAllPools(dir, rng)
	if err != nil {
		t.Fatalf("LoadAllPools: %v", err)
	}
	if len(loaded) != 3 {
		t.Fatalf("expected 3 loaded pools, got %d", len(loaded))
	}
	for _, key := range []string{"deck-a", "deck-b", "deck-c"} {
		if _, ok := loaded[key]; !ok {
			t.Errorf("missing pool for %q", key)
		}
	}
}

func TestLoadAllPools_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	rng := rand.New(rand.NewSource(42))

	pools, err := LoadAllPools(dir, rng)
	if err != nil {
		t.Fatalf("LoadAllPools empty dir: %v", err)
	}
	if len(pools) != 0 {
		t.Errorf("expected 0 pools from empty dir, got %d", len(pools))
	}
}

func TestLoadAllPools_NonexistentDir(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	pools, err := LoadAllPools("/tmp/nonexistent-curse-dir-12345", rng)
	if err != nil {
		t.Fatalf("LoadAllPools nonexistent dir should not error: %v", err)
	}
	if len(pools) != 0 {
		t.Errorf("expected 0 pools, got %d", len(pools))
	}
}

func TestDimensionStats_RecordAndCorrelation(t *testing.T) {
	var ds DimensionStats

	for i := 0; i < 50; i++ {
		var dims [NumDimensions]float64
		outcome := 0.0
		if i%2 == 0 {
			dims[0] = 1.0 // high board presence on wins
			outcome = 1.0
		} else {
			dims[0] = 0.0
			outcome = 0.0
		}
		dims[1] = 0.5 // card advantage always neutral
		ds.RecordGame(dims, outcome)
	}

	if ds.N != 50 {
		t.Errorf("expected N=50, got %d", ds.N)
	}

	corrBoard := ds.Correlation(0)
	if corrBoard < 0.5 {
		t.Errorf("expected strong positive correlation for dim 0, got %f", corrBoard)
	}

	corrCards := ds.Correlation(1)
	if corrCards > 0.3 || corrCards < -0.3 {
		t.Errorf("expected near-zero correlation for dim 1, got %f", corrCards)
	}

	corrections := ds.WeightCorrections()
	if corrections[0] <= 1.0 {
		t.Errorf("expected positive correction for dim 0, got %f", corrections[0])
	}
}

func TestCurseClampRange(t *testing.T) {
	if v := clampRange(0.5, 0.2, 0.4); v != 0.4 {
		t.Errorf("clampRange(0.5, 0.2, 0.4) = %f, want 0.4", v)
	}
	if v := clampRange(0.1, 0.2, 0.4); v != 0.2 {
		t.Errorf("clampRange(0.1, 0.2, 0.4) = %f, want 0.2", v)
	}
	if v := clampRange(0.3, 0.2, 0.4); v != 0.3 {
		t.Errorf("clampRange(0.3, 0.2, 0.4) = %f, want 0.3", v)
	}
}

func TestCurseInitPoolWithConstraints_SeedsLockedTraits(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	constraints := map[string]float64{
		"aggression":       0.9,
		"artifact_affinity": 0.1,
	}
	pool := InitPoolWithConstraints("test/deck", rng, constraints)

	for i, dna := range pool.Population {
		if dna.Aggression < 0.8 || dna.Aggression > 1.0 {
			t.Errorf("pop[%d] Aggression = %f, want within [0.8,1.0] (target 0.9 ±0.1)", i, dna.Aggression)
		}
		if dna.ArtifactAffinity < 0.0 || dna.ArtifactAffinity > 0.2 {
			t.Errorf("pop[%d] ArtifactAffinity = %f, want within [0.0,0.2] (target 0.1 ±0.1)", i, dna.ArtifactAffinity)
		}
	}
	if pool.Constraints["aggression"] != 0.9 {
		t.Errorf("pool.Constraints[aggression] = %f, want 0.9", pool.Constraints["aggression"])
	}
}

func TestCurseInitPoolWithConstraints_IgnoresUnknownKeys(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	pool := InitPoolWithConstraints("test", rng, map[string]float64{
		"bogus_trait": 0.5,
		"aggression":  0.5,
	})
	if _, ok := pool.Constraints["bogus_trait"]; ok {
		t.Errorf("expected bogus_trait to be filtered out")
	}
	if _, ok := pool.Constraints["aggression"]; !ok {
		t.Errorf("expected aggression to survive")
	}
}

func TestCurseEvolution_RespectsConstraints(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	pool := InitPoolWithConstraints("test", rng, map[string]float64{
		"aggression": 0.85,
	})

	// Force a fitness gradient and run enough games to trigger several
	// evolution + immigration cycles.
	for gen := 0; gen < curseImmigrationInterval+1; gen++ {
		for i := 0; i < CursePopSize; i++ {
			pool.Population[i].Fitness = float64(i+1) / float64(CursePopSize)
		}
		for i := 0; i < CurseEvolveAt; i++ {
			score := 0.0
			if i%2 == 0 {
				score = 1.0
			}
			pool.RecordResult(i%CursePopSize, score)
		}
	}

	for i, dna := range pool.Population {
		if dna.Aggression < 0.75-1e-9 || dna.Aggression > 0.95+1e-9 {
			t.Errorf("pop[%d] Aggression = %f after evolution, want within [0.75,0.95]", i, dna.Aggression)
		}
	}
}

func TestCurseApplyConstraintsToAll(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	pool := InitPool("test", rng)
	for i := range pool.Population {
		pool.Population[i].Aggression = 0.05
	}
	pool.Constraints = map[string]float64{"aggression": 0.7}
	pool.ApplyConstraintsToAll()
	for i, dna := range pool.Population {
		if dna.Aggression < 0.6-1e-9 || dna.Aggression > 0.8+1e-9 {
			t.Errorf("pop[%d] Aggression = %f, want within [0.6,0.8]", i, dna.Aggression)
		}
	}
}

func TestCurseIsValidTrait(t *testing.T) {
	for _, k := range CurseTraitKeys {
		if !IsValidCurseTrait(k) {
			t.Errorf("expected %q to be valid", k)
		}
	}
	if IsValidCurseTrait("not_a_trait") {
		t.Errorf("expected 'not_a_trait' to be invalid")
	}
}

func TestCurseSavePoolPersistsConstraints(t *testing.T) {
	dir := t.TempDir()
	rng := rand.New(rand.NewSource(7))
	pool := InitPoolWithConstraints("owner/deck", rng, map[string]float64{
		"aggression":     0.9,
		"combo_patience": 0.2,
	})
	if err := SavePool(dir, &pool); err != nil {
		t.Fatalf("SavePool: %v", err)
	}
	loaded, err := LoadPool(dir, "owner/deck", rng)
	if err != nil {
		t.Fatalf("LoadPool: %v", err)
	}
	if loaded.Constraints["aggression"] != 0.9 {
		t.Errorf("loaded aggression constraint = %f, want 0.9", loaded.Constraints["aggression"])
	}
	if loaded.Constraints["combo_patience"] != 0.2 {
		t.Errorf("loaded combo_patience constraint = %f, want 0.2", loaded.Constraints["combo_patience"])
	}
}

func TestDimensionStats_MinN(t *testing.T) {
	var ds DimensionStats
	for i := 0; i < 5; i++ {
		var dims [NumDimensions]float64
		ds.RecordGame(dims, 1.0)
	}
	corr := ds.Correlation(0)
	if corr != 0 {
		t.Errorf("expected 0 correlation before minN, got %f", corr)
	}
	corrections := ds.WeightCorrections()
	if corrections[0] != 1.0 {
		t.Errorf("expected no correction before minN, got %f", corrections[0])
	}
}
