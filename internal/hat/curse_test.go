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
	// 4-player Commander — survival-weighted scale.
	// 1st=1.0, 2nd=0.85, 3rd=0.45, 4th=0.0.
	if s := PlacementScore(1, 4); s != 1.0 {
		t.Errorf("1st of 4 = %f, want 1.0", s)
	}
	if s := PlacementScore(2, 4); s != 0.85 {
		t.Errorf("2nd of 4 = %f, want 0.85", s)
	}
	if s := PlacementScore(3, 4); s != 0.45 {
		t.Errorf("3rd of 4 = %f, want 0.45", s)
	}
	if s := PlacementScore(4, 4); s != 0.0 {
		t.Errorf("4th of 4 = %f, want 0.0", s)
	}

	// 2-player (binary fallback via linear formula).
	if s := PlacementScore(1, 2); s != 1.0 {
		t.Errorf("1st of 2 = %f, want 1.0", s)
	}
	if s := PlacementScore(2, 2); s != 0.0 {
		t.Errorf("2nd of 2 = %f, want 0.0", s)
	}

	// 3-player (linear fallback): 1.0 / 0.5 / 0.0
	if s := PlacementScore(2, 3); s != 0.5 {
		t.Errorf("2nd of 3 = %f, want 0.5", s)
	}

	// Invalid inputs return 0.5 (neutral).
	if s := PlacementScore(0, 4); s != 0.5 {
		t.Errorf("placement=0 should return 0.5, got %f", s)
	}
	if s := PlacementScore(1, 0); s != 0.5 {
		t.Errorf("totalPlayers=0 should return 0.5, got %f", s)
	}

	// Mean per-game score in 4-player (sanity — should be > 0.5 reflecting
	// the survival-reward shift).
	mean := (PlacementScore(1, 4) + PlacementScore(2, 4) + PlacementScore(3, 4) + PlacementScore(4, 4)) / 4.0
	if mean <= 0.5 {
		t.Errorf("mean 4-player score = %f, want > 0.5 (survival-weighted)", mean)
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

// TestNumCurseAxesMatchesTraitKeys is a guardrail: NumCurseAxes must equal
// the number of trait keys, the number of fields packed by axes(), and
// the number of mutate() calls in evolve(). If a future axis is added
// without updating one of these, this test catches the drift.
func TestNumCurseAxesMatchesTraitKeys(t *testing.T) {
	if NumCurseAxes != len(CurseTraitKeys) {
		t.Fatalf("NumCurseAxes = %d, len(CurseTraitKeys) = %d", NumCurseAxes, len(CurseTraitKeys))
	}
	var d CurseDNA
	if got := d.axes(); len(got) != NumCurseAxes {
		t.Fatalf("axes() returned %d entries, want %d", len(got), NumCurseAxes)
	}
}

func TestCurseTraitKeys_IncludesNewAxes(t *testing.T) {
	want := []string{
		"land_greed",
		"equipment_affinity",
		"graveyard_exploitation",
		"counterplay_timing",
		"token_pressure",
	}
	for _, k := range want {
		if !IsValidCurseTrait(k) {
			t.Errorf("expected %q to be a valid curse trait", k)
		}
	}
}

func TestInitPool_SeedsNewAxesInRange(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	pool := InitPool("test/expanded", rng)
	for i, dna := range pool.Population {
		check := []struct {
			name string
			v    float64
		}{
			{"LandGreed", dna.LandGreed},
			{"EquipmentAffinity", dna.EquipmentAffinity},
			{"GraveyardExploitation", dna.GraveyardExploitation},
			{"CounterplayTiming", dna.CounterplayTiming},
			{"TokenPressure", dna.TokenPressure},
		}
		for _, c := range check {
			if c.v < 0 || c.v > 1 {
				t.Errorf("pop[%d] %s = %f, want in [0,1]", i, c.name, c.v)
			}
		}
	}
}

func TestInitPoolWithConstraints_NewAxes(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	pool := InitPoolWithConstraints("t", rng, map[string]float64{
		"land_greed":             0.9,
		"equipment_affinity":     0.1,
		"graveyard_exploitation": 0.85,
		"counterplay_timing":     0.2,
		"token_pressure":         0.95,
	})
	for i, dna := range pool.Population {
		bands := []struct {
			name string
			v, t float64
		}{
			{"LandGreed", dna.LandGreed, 0.9},
			{"EquipmentAffinity", dna.EquipmentAffinity, 0.1},
			{"GraveyardExploitation", dna.GraveyardExploitation, 0.85},
			{"CounterplayTiming", dna.CounterplayTiming, 0.2},
			{"TokenPressure", dna.TokenPressure, 0.95},
		}
		for _, b := range bands {
			lo := b.t - curseConstraintBand - 1e-9
			hi := b.t + curseConstraintBand + 1e-9
			if lo < 0 {
				lo = 0
			}
			if hi > 1 {
				hi = 1
			}
			if b.v < lo || b.v > hi {
				t.Errorf("pop[%d] %s = %f, want within [%f,%f] (target %f)", i, b.name, b.v, lo, hi, b.t)
			}
		}
	}
}

func TestEvolve_MovesNewAxes(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	pool := InitPool("test/move", rng)
	// Pin every member's new axes to identical seeds, then force
	// evolution and confirm the killed-and-replaced slots actually got
	// fresh values from crossover+mutation.
	for i := range pool.Population {
		pool.Population[i].LandGreed = 0.5
		pool.Population[i].EquipmentAffinity = 0.5
		pool.Population[i].GraveyardExploitation = 0.5
		pool.Population[i].CounterplayTiming = 0.5
		pool.Population[i].TokenPressure = 0.5
		pool.Population[i].Fitness = float64(i+1) / float64(CursePopSize)
	}
	for i := 0; i < CurseEvolveAt; i++ {
		score := 0.0
		if i%2 == 0 {
			score = 1.0
		}
		pool.RecordResult(i%CursePopSize, score)
	}
	// At least one member should now differ from the pinned 0.5 on at
	// least one of the new axes (mutation shifted it).
	moved := false
	for _, dna := range pool.Population {
		if dna.LandGreed != 0.5 || dna.EquipmentAffinity != 0.5 ||
			dna.GraveyardExploitation != 0.5 || dna.CounterplayTiming != 0.5 ||
			dna.TokenPressure != 0.5 {
			moved = true
			break
		}
	}
	if !moved {
		t.Errorf("expected at least one new-axis value to move after evolution")
	}
}

func TestCurseAxisStats_PositiveCorrelation(t *testing.T) {
	var as CurseAxisStats
	// Axis 7 (LandGreed) is the win signal: high LandGreed → win.
	for i := 0; i < 50; i++ {
		var v [NumCurseAxes]float64
		outcome := 0.0
		if i%2 == 0 {
			v[7] = 1.0
			outcome = 1.0
		}
		// Other axes static (zero variance → zero correlation).
		as.RecordGame(v, outcome)
	}
	if as.N != 50 {
		t.Fatalf("N = %d, want 50", as.N)
	}
	r := as.Correlation(7)
	if r < 0.5 {
		t.Errorf("LandGreed correlation = %f, want > 0.5", r)
	}
	// Static axis 0 should produce 0 correlation (zero variance).
	if c0 := as.Correlation(0); c0 != 0 {
		t.Errorf("static axis correlation = %f, want 0", c0)
	}
	// MutationScale should narrow on the strongly positive axis.
	if s := as.MutationScale(7); s >= 1.0 {
		t.Errorf("MutationScale on positively-correlated axis = %f, want < 1.0", s)
	}
}

func TestCurseAxisStats_MinN(t *testing.T) {
	var as CurseAxisStats
	for i := 0; i < 5; i++ {
		var v [NumCurseAxes]float64
		v[0] = float64(i) / 5.0
		as.RecordGame(v, float64(i)/5.0)
	}
	if r := as.Correlation(0); r != 0 {
		t.Errorf("expected 0 correlation before minN, got %f", r)
	}
	if s := as.MutationScale(0); s != 1.0 {
		t.Errorf("expected MutationScale=1.0 before minN, got %f", s)
	}
}

func TestCurseRecordResult_FeedsAxisStats(t *testing.T) {
	rng := rand.New(rand.NewSource(123))
	pool := InitPool("test/feed", rng)
	for i := 0; i < 30; i++ {
		pool.RecordResult(i%CursePopSize, 0.6)
	}
	if pool.AxisStats.N != 30 {
		t.Errorf("AxisStats.N = %d, want 30", pool.AxisStats.N)
	}
	if pool.AxisStats.MeanScore == 0 {
		t.Errorf("AxisStats.MeanScore = 0, want non-zero after 30 records")
	}
}

func TestCurseCrossover_TakesGenesFromBothParents(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	pool := InitPool("xover", rng)
	a := CurseDNA{
		Aggression: 0.0, ComboPat: 0.0, ThreatParanoia: 0.0, ResourceGreed: 0.0,
		PoliticalMemory: 0.0, DrainAffinity: 0.0, ArtifactAffinity: 0.0,
		LandGreed: 0.0, EquipmentAffinity: 0.0, GraveyardExploitation: 0.0,
		CounterplayTiming: 0.0, TokenPressure: 0.0,
	}
	b := CurseDNA{
		Aggression: 1.0, ComboPat: 1.0, ThreatParanoia: 1.0, ResourceGreed: 1.0,
		PoliticalMemory: 1.0, DrainAffinity: 1.0, ArtifactAffinity: 1.0,
		LandGreed: 1.0, EquipmentAffinity: 1.0, GraveyardExploitation: 1.0,
		CounterplayTiming: 1.0, TokenPressure: 1.0,
	}
	// Many crossovers — across runs we expect both 0s and 1s on every gene.
	zerosFound := [NumCurseAxes]bool{}
	onesFound := [NumCurseAxes]bool{}
	for i := 0; i < 200; i++ {
		c := pool.crossover(&a, &b)
		axes := c.axes()
		for j := 0; j < NumCurseAxes; j++ {
			if axes[j] == 0.0 {
				zerosFound[j] = true
			}
			if axes[j] == 1.0 {
				onesFound[j] = true
			}
		}
	}
	for j := 0; j < NumCurseAxes; j++ {
		if !zerosFound[j] || !onesFound[j] {
			t.Errorf("axis %d: zeros=%v ones=%v, want both true", j, zerosFound[j], onesFound[j])
		}
	}
}

func TestSaveAndLoadPool_PreservesNewAxes(t *testing.T) {
	dir := t.TempDir()
	rng := rand.New(rand.NewSource(7))
	pool := InitPool("owner/expanded-deck", rng)
	pool.Population[0].LandGreed = 0.123
	pool.Population[0].EquipmentAffinity = 0.456
	pool.Population[0].GraveyardExploitation = 0.789
	pool.Population[0].CounterplayTiming = 0.321
	pool.Population[0].TokenPressure = 0.654

	if err := SavePool(dir, &pool); err != nil {
		t.Fatalf("SavePool: %v", err)
	}
	loaded, err := LoadPool(dir, "owner/expanded-deck", rng)
	if err != nil {
		t.Fatalf("LoadPool: %v", err)
	}
	d := loaded.Population[0]
	if d.LandGreed != 0.123 {
		t.Errorf("LandGreed = %f, want 0.123", d.LandGreed)
	}
	if d.EquipmentAffinity != 0.456 {
		t.Errorf("EquipmentAffinity = %f, want 0.456", d.EquipmentAffinity)
	}
	if d.GraveyardExploitation != 0.789 {
		t.Errorf("GraveyardExploitation = %f, want 0.789", d.GraveyardExploitation)
	}
	if d.CounterplayTiming != 0.321 {
		t.Errorf("CounterplayTiming = %f, want 0.321", d.CounterplayTiming)
	}
	if d.TokenPressure != 0.654 {
		t.Errorf("TokenPressure = %f, want 0.654", d.TokenPressure)
	}
}
