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
	for i := 1; i < AmiiboPopSize; i++ {
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

	pool.RecordResult(0, true)
	if pool.Population[0].Fitness != 1.0 {
		t.Errorf("fitness after first win = %f, want 1.0", pool.Population[0].Fitness)
	}

	pool.RecordResult(0, false)
	if pool.Population[0].Fitness >= 1.0 {
		t.Errorf("fitness should decrease after loss, got %f", pool.Population[0].Fitness)
	}
}

func TestEvolution_TriggersAt100(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	pool := InitPool("test", rng)

	// Give population a clear fitness gradient so evolution has something to work with.
	for i := 0; i < AmiiboPopSize; i++ {
		pool.Population[i].Fitness = float64(i+1) / float64(AmiiboPopSize)
	}

	// Record 100 games to trigger evolution.
	for i := 0; i < AmiiboEvolveAt; i++ {
		pool.RecordResult(i%AmiiboPopSize, i%2 == 0)
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

	pools := make(map[string]*AmiiboPool)
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
	pools, err := LoadAllPools("/tmp/nonexistent-amiibo-dir-12345", rng)
	if err != nil {
		t.Fatalf("LoadAllPools nonexistent dir should not error: %v", err)
	}
	if len(pools) != 0 {
		t.Errorf("expected 0 pools, got %d", len(pools))
	}
}
