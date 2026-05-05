package hat

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHarvestHighFitness_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	cfg := DistillationConfig{
		CurseDir:         dir,
		TrainingDir:       filepath.Join(dir, "training"),
		FitnessQuartile:   0.25,
		MinPoolsForSeed:   5,
		MinGensForHarvest: 3,
		CycleInterval:     time.Minute,
		ReseedThreshold:   0.35,
	}
	rng := rand.New(rand.NewSource(42))
	dm := NewDistillationManager(cfg, rng)

	manifest, err := dm.HarvestHighFitness(nil)
	if err != nil {
		t.Fatalf("HarvestHighFitness on empty dir: %v", err)
	}
	if manifest == nil {
		t.Fatal("expected non-nil manifest")
	}
	if len(manifest.HarvestedDNA) != 0 {
		t.Errorf("expected 0 harvested DNA, got %d", len(manifest.HarvestedDNA))
	}
}

func TestHarvestHighFitness_FiltersLowGen(t *testing.T) {
	dir := t.TempDir()
	rng := rand.New(rand.NewSource(42))

	// Create a pool with only 1 generation (below threshold of 3).
	pool := InitPool("low-gen-deck", rng)
	pool.GenCount = 1
	for i := range pool.Population {
		pool.Population[i].Fitness = 0.9
		pool.Population[i].GamesPlayed = 50
	}
	SavePool(dir, &pool)

	cfg := DistillationConfig{
		CurseDir:         dir,
		TrainingDir:       filepath.Join(dir, "training"),
		FitnessQuartile:   0.25,
		MinPoolsForSeed:   5,
		MinGensForHarvest: 3,
		CycleInterval:     time.Minute,
		ReseedThreshold:   0.35,
	}
	dm := NewDistillationManager(cfg, rng)

	manifest, err := dm.HarvestHighFitness(nil)
	if err != nil {
		t.Fatalf("HarvestHighFitness: %v", err)
	}
	if len(manifest.HarvestedDNA) != 0 {
		t.Errorf("expected 0 harvested (pool too young), got %d", len(manifest.HarvestedDNA))
	}
}

func TestHarvestHighFitness_SelectsTopQuartile(t *testing.T) {
	dir := t.TempDir()
	rng := rand.New(rand.NewSource(42))

	// Create 4 pools with different fitness levels.
	for i, key := range []string{"deck-a", "deck-b", "deck-c", "deck-d"} {
		pool := InitPool(key, rng)
		pool.GenCount = 5
		for j := range pool.Population {
			// Spread fitness: deck-a has lowest, deck-d has highest.
			pool.Population[j].Fitness = float64(i*CursePopSize+j+1) / float64(4*CursePopSize)
			pool.Population[j].GamesPlayed = 50
		}
		SavePool(dir, &pool)
	}

	cfg := DistillationConfig{
		CurseDir:         dir,
		TrainingDir:       filepath.Join(dir, "training"),
		FitnessQuartile:   0.25,
		MinPoolsForSeed:   2,
		MinGensForHarvest: 3,
		CycleInterval:     time.Minute,
		ReseedThreshold:   0.35,
	}
	dm := NewDistillationManager(cfg, rng)

	manifest, err := dm.HarvestHighFitness(nil)
	if err != nil {
		t.Fatalf("HarvestHighFitness: %v", err)
	}

	totalDNA := 4 * CursePopSize
	expectedHarvest := int(float64(totalDNA) * 0.25)
	// Allow +-1 due to ceiling.
	if len(manifest.HarvestedDNA) < expectedHarvest-1 || len(manifest.HarvestedDNA) > expectedHarvest+1 {
		t.Errorf("expected ~%d harvested DNA, got %d", expectedHarvest, len(manifest.HarvestedDNA))
	}

	// All harvested should have fitness > median.
	for _, h := range manifest.HarvestedDNA {
		if h.DNA.Fitness < 0.5 {
			t.Errorf("harvested DNA has low fitness: %f", h.DNA.Fitness)
		}
	}
}

func TestHarvestHighFitness_WithStrategyLookup(t *testing.T) {
	dir := t.TempDir()
	rng := rand.New(rand.NewSource(42))

	pool := InitPool("combo-deck", rng)
	pool.GenCount = 5
	for i := range pool.Population {
		pool.Population[i].Fitness = 0.8
		pool.Population[i].GamesPlayed = 50
	}
	SavePool(dir, &pool)

	cfg := DistillationConfig{
		CurseDir:         dir,
		TrainingDir:       filepath.Join(dir, "training"),
		FitnessQuartile:   0.50,
		MinPoolsForSeed:   1,
		MinGensForHarvest: 3,
		CycleInterval:     time.Minute,
		ReseedThreshold:   0.35,
	}
	dm := NewDistillationManager(cfg, rng)

	lookup := func(deckKey string) *StrategyProfile {
		if deckKey == "combo-deck" {
			return &StrategyProfile{
				Archetype: ArchetypeCombo,
				Bracket:   4,
			}
		}
		return nil
	}

	manifest, err := dm.HarvestHighFitness(lookup)
	if err != nil {
		t.Fatalf("HarvestHighFitness: %v", err)
	}

	for _, h := range manifest.HarvestedDNA {
		if h.Archetype != ArchetypeCombo {
			t.Errorf("expected archetype %q, got %q", ArchetypeCombo, h.Archetype)
		}
		if h.Bracket != 4 {
			t.Errorf("expected bracket 4, got %d", h.Bracket)
		}
	}
}

func TestDNAParamsFromCurse(t *testing.T) {
	dna := CurseDNA{
		Aggression:       0.1,
		ComboPat:         0.2,
		ThreatParanoia:   0.3,
		ResourceGreed:    0.4,
		PoliticalMemory:  0.5,
		DrainAffinity:    0.6,
		ArtifactAffinity: 0.7,
	}
	params := DNAParamsFromCurse(&dna)
	expected := [7]float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7}
	if params != expected {
		t.Errorf("DNAParamsFromCurse = %v, want %v", params, expected)
	}
}

func TestFitnessToWeight(t *testing.T) {
	tests := []struct {
		fitness float64
		wantMin float64
		wantMax float64
	}{
		{0.0, 0.49, 0.51},
		{0.5, 1.20, 1.30},
		{1.0, 1.99, 2.01},
		{-0.5, 0.49, 0.51}, // clamp
		{2.0, 1.99, 2.01},  // clamp
	}
	for _, tt := range tests {
		w := FitnessToWeight(tt.fitness)
		if w < tt.wantMin || w > tt.wantMax {
			t.Errorf("FitnessToWeight(%f) = %f, want [%f, %f]", tt.fitness, w, tt.wantMin, tt.wantMax)
		}
	}
}

func TestEnrichWithDNA(t *testing.T) {
	dna := CurseDNA{
		Aggression:       0.8,
		ComboPat:         0.3,
		ThreatParanoia:   0.6,
		ResourceGreed:    0.4,
		PoliticalMemory:  0.2,
		DrainAffinity:    0.7,
		ArtifactAffinity: 0.5,
		Fitness:          0.75,
	}

	samples := []PivotEnrichedSample{
		{
			State:     StateVector{},
			Placement: 1.0,
			Turn:      5,
			GameTurn:  20,
		},
		{
			State:     StateVector{},
			Placement: 0.5,
			Turn:      10,
			GameTurn:  20,
		},
	}

	enriched := EnrichWithDNA(samples, &dna)
	if len(enriched) != 2 {
		t.Fatalf("expected 2 enriched samples, got %d", len(enriched))
	}

	expectedWeight := FitnessToWeight(0.75)
	for _, s := range enriched {
		if s.FitnessWeight != expectedWeight {
			t.Errorf("FitnessWeight = %f, want %f", s.FitnessWeight, expectedWeight)
		}
		if s.DNAParams[0] != 0.8 {
			t.Errorf("DNAParams[0] = %f, want 0.8", s.DNAParams[0])
		}
	}
}

func TestSeedDNAFromManifest_NoManifest(t *testing.T) {
	cfg := DefaultDistillationConfig(t.TempDir())
	rng := rand.New(rand.NewSource(42))
	dm := NewDistillationManager(cfg, rng)

	params := dm.SeedDNAFromManifest(ArchetypeCombo, 3)
	for i, p := range params {
		if p != 0.5 {
			t.Errorf("params[%d] = %f, want 0.5 (no manifest fallback)", i, p)
		}
	}
}

func TestSeedDNAFromManifest_ArchetypeMatch(t *testing.T) {
	cfg := DefaultDistillationConfig(t.TempDir())
	rng := rand.New(rand.NewSource(42))
	dm := NewDistillationManager(cfg, rng)

	// Inject a manifest with known DNA.
	dm.mu.Lock()
	dm.manifest = &DistillationManifest{
		HarvestedDNA: []HighFitnessDNA{
			{
				DNA: CurseDNA{
					Aggression: 0.2, ComboPat: 0.9, ThreatParanoia: 0.3,
					ResourceGreed: 0.8, PoliticalMemory: 0.1, DrainAffinity: 0.4,
					ArtifactAffinity: 0.5,
				},
				Archetype: ArchetypeCombo,
				Bracket:   4,
			},
			{
				DNA: CurseDNA{
					Aggression: 0.4, ComboPat: 0.7, ThreatParanoia: 0.5,
					ResourceGreed: 0.6, PoliticalMemory: 0.3, DrainAffinity: 0.2,
					ArtifactAffinity: 0.3,
				},
				Archetype: ArchetypeCombo,
				Bracket:   4,
			},
			{
				DNA: CurseDNA{
					Aggression: 0.9, ComboPat: 0.1, ThreatParanoia: 0.8,
					ResourceGreed: 0.2, PoliticalMemory: 0.7, DrainAffinity: 0.1,
					ArtifactAffinity: 0.1,
				},
				Archetype: ArchetypeAggro,
				Bracket:   2,
			},
		},
	}
	dm.mu.Unlock()

	// Query for combo archetype — should get average of the two combo entries.
	params := dm.SeedDNAFromManifest(ArchetypeCombo, 4)
	expectedAggression := (0.2 + 0.4) / 2.0
	if abs(params[0]-expectedAggression) > 0.01 {
		t.Errorf("params[0] (aggression) = %f, want %f", params[0], expectedAggression)
	}
	expectedComboPat := (0.9 + 0.7) / 2.0
	if abs(params[1]-expectedComboPat) > 0.01 {
		t.Errorf("params[1] (combo_pat) = %f, want %f", params[1], expectedComboPat)
	}
}

func TestSeedDNAFromManifest_BracketFallback(t *testing.T) {
	cfg := DefaultDistillationConfig(t.TempDir())
	rng := rand.New(rand.NewSource(42))
	dm := NewDistillationManager(cfg, rng)

	dm.mu.Lock()
	dm.manifest = &DistillationManifest{
		HarvestedDNA: []HighFitnessDNA{
			{
				DNA:       CurseDNA{Aggression: 0.7, ComboPat: 0.3},
				Archetype: ArchetypeAggro,
				Bracket:   3,
			},
		},
	}
	dm.mu.Unlock()

	// Query for control (no match) at bracket 3 (match).
	params := dm.SeedDNAFromManifest(ArchetypeControl, 3)
	if abs(params[0]-0.7) > 0.01 {
		t.Errorf("bracket fallback: params[0] = %f, want 0.7", params[0])
	}
}

func TestInitPoolSeeded(t *testing.T) {
	cfg := DefaultDistillationConfig(t.TempDir())
	rng := rand.New(rand.NewSource(42))
	dm := NewDistillationManager(cfg, rng)

	dm.mu.Lock()
	dm.manifest = &DistillationManifest{
		HarvestedDNA: []HighFitnessDNA{
			{
				DNA: CurseDNA{
					Aggression: 0.3, ComboPat: 0.8, ThreatParanoia: 0.5,
					ResourceGreed: 0.7, PoliticalMemory: 0.4, DrainAffinity: 0.6,
					ArtifactAffinity: 0.2,
				},
				Archetype: ArchetypeCombo,
				Bracket:   4,
			},
		},
	}
	dm.mu.Unlock()

	pool := dm.InitPoolSeeded("new-deck", ArchetypeCombo, 4, rng)

	if pool.DeckKey != "new-deck" {
		t.Errorf("pool deck key = %q, want 'new-deck'", pool.DeckKey)
	}
	if pool.Bracket != 4 {
		t.Errorf("pool bracket = %d, want 4", pool.Bracket)
	}

	// All population members should be near the centroid (within noise).
	for i, dna := range pool.Population {
		if dna.Aggression < 0.0 || dna.Aggression > 1.0 {
			t.Errorf("pop[%d] aggression out of range: %f", i, dna.Aggression)
		}
		// Should be roughly near 0.3 (centroid) ± 0.1 noise.
		if abs(dna.Aggression-0.3) > 0.4 {
			t.Errorf("pop[%d] aggression too far from centroid: %f (want ~0.3)", i, dna.Aggression)
		}
	}
}

func TestSaveAndLoadDistillationManifest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")

	original := &DistillationManifest{
		HarvestTime: time.Now().Truncate(time.Second),
		TotalPools:  10,
		MeanFitness: 0.72,
		TopFitness:  0.95,
		HarvestedDNA: []HighFitnessDNA{
			{
				DNA:       CurseDNA{Aggression: 0.5, Fitness: 0.9},
				Archetype: ArchetypeMidrange,
				Bracket:   3,
				GenCount:  7,
			},
		},
	}

	if err := SaveDistillationManifest(path, original); err != nil {
		t.Fatalf("SaveDistillationManifest: %v", err)
	}

	loaded, err := LoadDistillationManifest(path)
	if err != nil {
		t.Fatalf("LoadDistillationManifest: %v", err)
	}

	if loaded.TotalPools != 10 {
		t.Errorf("loaded TotalPools = %d, want 10", loaded.TotalPools)
	}
	if loaded.MeanFitness != 0.72 {
		t.Errorf("loaded MeanFitness = %f, want 0.72", loaded.MeanFitness)
	}
	if len(loaded.HarvestedDNA) != 1 {
		t.Fatalf("loaded HarvestedDNA length = %d, want 1", len(loaded.HarvestedDNA))
	}
	if loaded.HarvestedDNA[0].Archetype != ArchetypeMidrange {
		t.Errorf("loaded archetype = %q, want %q", loaded.HarvestedDNA[0].Archetype, ArchetypeMidrange)
	}
}

func TestTryLoadManifest(t *testing.T) {
	dir := t.TempDir()
	trainingDir := filepath.Join(dir, "training")
	os.MkdirAll(trainingDir, 0755)

	cfg := DistillationConfig{
		CurseDir:       filepath.Join(dir, "curse"),
		TrainingDir:     trainingDir,
		FitnessQuartile: 0.25,
		CycleInterval:   time.Minute,
	}
	rng := rand.New(rand.NewSource(42))
	dm := NewDistillationManager(cfg, rng)

	// No manifest exists yet.
	if dm.TryLoadManifest() {
		t.Error("TryLoadManifest should return false when no file exists")
	}

	// Save a manifest.
	manifest := &DistillationManifest{
		TotalPools: 5,
		HarvestedDNA: []HighFitnessDNA{
			{DNA: CurseDNA{Aggression: 0.6}, Archetype: ArchetypeAggro},
		},
	}
	path := filepath.Join(trainingDir, "distillation_manifest.json")
	SaveDistillationManifest(path, manifest)

	// Now it should load.
	if !dm.TryLoadManifest() {
		t.Error("TryLoadManifest should return true after saving manifest")
	}
	if dm.ManifestSize() != 1 {
		t.Errorf("ManifestSize = %d, want 1", dm.ManifestSize())
	}
}

func TestTryCycle_RespectsInterval(t *testing.T) {
	dir := t.TempDir()
	cfg := DistillationConfig{
		CurseDir:         filepath.Join(dir, "curse"),
		TrainingDir:       filepath.Join(dir, "training"),
		FitnessQuartile:   0.25,
		MinPoolsForSeed:   1,
		MinGensForHarvest: 3,
		CycleInterval:     time.Hour, // long interval
		ReseedThreshold:   0.35,
	}
	os.MkdirAll(cfg.CurseDir, 0755)
	os.MkdirAll(cfg.TrainingDir, 0755)

	rng := rand.New(rand.NewSource(42))
	dm := NewDistillationManager(cfg, rng)

	// First cycle should trigger.
	triggered := dm.TryCycle(nil)
	if !triggered {
		t.Error("first TryCycle should trigger")
	}

	// Wait for goroutine to finish.
	for dm.IsRunning() {
		// spin
	}

	// Second cycle should NOT trigger (within interval).
	triggered = dm.TryCycle(nil)
	if triggered {
		t.Error("second TryCycle should not trigger within interval")
	}
}

func TestAppendDNAEnrichedSamples(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "enriched.jsonl")

	samples := []DNAEnrichedSample{
		{
			PivotEnrichedSample: PivotEnrichedSample{
				State:     StateVector{},
				Placement: 1.0,
				Turn:      5,
				GameTurn:  20,
			},
			DNAParams:     [7]float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7},
			FitnessWeight: 1.5,
		},
	}

	err := AppendDNAEnrichedSamples(path, samples)
	if err != nil {
		t.Fatalf("AppendDNAEnrichedSamples: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() == 0 {
		t.Error("expected non-empty file")
	}
}

func TestReseedUnderperformers(t *testing.T) {
	dir := t.TempDir()
	curseDir := filepath.Join(dir, "curse")
	os.MkdirAll(curseDir, 0755)

	rng := rand.New(rand.NewSource(42))

	// Create an underperforming pool.
	pool := InitPool("bad-deck", rng)
	pool.GenCount = 5
	for i := range pool.Population {
		pool.Population[i].Fitness = 0.1 // very low
		pool.Population[i].GamesPlayed = 50
	}
	SavePool(curseDir, &pool)

	cfg := DistillationConfig{
		CurseDir:         curseDir,
		TrainingDir:       filepath.Join(dir, "training"),
		FitnessQuartile:   0.25,
		MinPoolsForSeed:   1,
		MinGensForHarvest: 3,
		CycleInterval:     time.Minute,
		ReseedThreshold:   0.35,
	}
	dm := NewDistillationManager(cfg, rng)

	// Set manifest with high-fitness reference data.
	dm.mu.Lock()
	dm.manifest = &DistillationManifest{
		HarvestedDNA: []HighFitnessDNA{
			{
				DNA: CurseDNA{
					Aggression: 0.7, ComboPat: 0.6, ThreatParanoia: 0.5,
					ResourceGreed: 0.4, PoliticalMemory: 0.3, DrainAffinity: 0.8,
					ArtifactAffinity: 0.9,
				},
				Archetype: ArchetypeMidrange,
				Bracket:   3,
			},
		},
	}
	dm.mu.Unlock()

	reseeded := dm.reseedUnderperformers(nil)
	if reseeded != 1 {
		t.Errorf("expected 1 pool reseeded, got %d", reseeded)
	}

	// Reload pool and verify some members were replaced.
	loaded, err := LoadPool(curseDir, "bad-deck", rng)
	if err != nil {
		t.Fatalf("LoadPool: %v", err)
	}

	// Bottom half should have been reset to fitness 0.25 (fresh).
	resetCount := 0
	for _, dna := range loaded.Population {
		if dna.Fitness == 0.25 {
			resetCount++
		}
	}
	if resetCount < CursePopSize/2 {
		t.Errorf("expected at least %d reset members, got %d", CursePopSize/2, resetCount)
	}
}

func TestDistillationManager_Status(t *testing.T) {
	cfg := DefaultDistillationConfig(t.TempDir())
	rng := rand.New(rand.NewSource(42))
	dm := NewDistillationManager(cfg, rng)

	status := dm.Status()
	if status == "" {
		t.Error("expected non-empty status")
	}
	if !contains(status, "running=false") {
		t.Errorf("expected 'running=false' in status, got %q", status)
	}
	if !contains(status, "no manifest") {
		t.Errorf("expected 'no manifest' in status, got %q", status)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
