package hat

import (
	"math"
	"math/rand"
	"testing"
)

// ---------------------------------------------------------------------------
// archetypeDistance
// ---------------------------------------------------------------------------

func TestArchetypeDistance_SameArchetypeSameColors(t *testing.T) {
	a := &StrategyProfile{
		Archetype:       "Combo",
		ColorDemand:     map[string]int{"U": 12, "B": 8},
		CommanderThemes: []string{"counterspells", "tutors"},
	}
	b := &StrategyProfile{
		Archetype:       "Combo",
		ColorDemand:     map[string]int{"U": 10, "B": 6},
		CommanderThemes: []string{"counterspells", "tutors"},
	}
	d := archetypeDistance(a, b)
	if d > 0.05 {
		t.Fatalf("identical archetype/colors/themes should be ~0; got %.3f", d)
	}
}

func TestArchetypeDistance_DifferentEverything(t *testing.T) {
	a := &StrategyProfile{
		Archetype:       "Aggro",
		ColorDemand:     map[string]int{"R": 18},
		CommanderThemes: []string{"haste", "burn"},
	}
	b := &StrategyProfile{
		Archetype:       "Control",
		ColorDemand:     map[string]int{"W": 12, "U": 10},
		CommanderThemes: []string{"counterspells", "draw"},
	}
	d := archetypeDistance(a, b)
	if d < 0.95 {
		t.Fatalf("disjoint archetype/colors/themes should be ~1.0; got %.3f", d)
	}
}

func TestArchetypeDistance_NilProfileMaxDistance(t *testing.T) {
	other := &StrategyProfile{Archetype: "Combo"}
	if d := archetypeDistance(nil, other); d != 1.0 {
		t.Fatalf("nil profile should return max distance 1.0; got %.3f", d)
	}
	if d := archetypeDistance(other, nil); d != 1.0 {
		t.Fatalf("nil profile should return max distance 1.0; got %.3f", d)
	}
}

func TestArchetypeDistance_PartialMatchInBetween(t *testing.T) {
	// Same archetype, partially overlapping colors, no theme overlap.
	a := &StrategyProfile{
		Archetype:       "Midrange",
		ColorDemand:     map[string]int{"G": 10, "W": 8},
		CommanderThemes: []string{"lifegain"},
	}
	b := &StrategyProfile{
		Archetype:       "Midrange",
		ColorDemand:     map[string]int{"G": 9, "U": 7},
		CommanderThemes: []string{"counters"},
	}
	d := archetypeDistance(a, b)
	// Expect: archetype 0 + color (1 - 1/3)*0.3 = 0.2 + theme 0.2 = ~0.4.
	if d < 0.30 || d > 0.50 {
		t.Fatalf("partial match should land in (0.30, 0.50); got %.3f", d)
	}
}

// ---------------------------------------------------------------------------
// InitPoolWithTransfer
// ---------------------------------------------------------------------------

func TestInitPoolWithTransfer_InheritsFromSameArchetype(t *testing.T) {
	rng := rand.New(rand.NewSource(7))

	// Donor pool: a high-fitness "Combo" / UB deck. Aggression deliberately
	// near 0 (combo decks don't rush combat).
	donor := buildEvolvedPool("donor-combo-ub", 4, 0.05, 0.95)
	differentDonor := buildEvolvedPool("donor-aggro-r", 2, 0.95, 0.05)

	pools := map[string]*CursePool{
		"donor-combo-ub": &donor,
		"donor-aggro-r":  &differentDonor,
	}
	strats := map[string]*StrategyProfile{
		"donor-combo-ub": {
			Archetype:       "Combo",
			ColorDemand:     map[string]int{"U": 12, "B": 8},
			CommanderThemes: []string{"counterspells", "tutors"},
			Bracket:         4,
		},
		"donor-aggro-r": {
			Archetype:       "Aggro",
			ColorDemand:     map[string]int{"R": 18},
			CommanderThemes: []string{"haste", "burn"},
			Bracket:         2,
		},
		"new-combo": {
			Archetype:       "Combo",
			ColorDemand:     map[string]int{"U": 11, "B": 7},
			CommanderThemes: []string{"counterspells", "tutors"},
			Bracket:         4,
		},
	}

	pool := InitPoolWithTransfer("new-combo", 4, rng, pools, strats)

	if pool.DeckKey != "new-combo" {
		t.Fatalf("deck key should round-trip; got %q", pool.DeckKey)
	}
	// Average aggression across the new pool should be near the donor's
	// 0.05 (with SD=0.10 noise; mean across 8 individuals is well within
	// 0.10 of the seed center).
	mean := meanAggression(&pool)
	if math.Abs(mean-0.05) > 0.15 {
		t.Fatalf("transferred aggression should track donor (~0.05); got mean=%.3f", mean)
	}
	mean = meanCombo(&pool)
	if math.Abs(mean-0.95) > 0.15 {
		t.Fatalf("transferred combo_patience should track donor (~0.95); got mean=%.3f", mean)
	}
	// Fitness must NOT be carried over — new deck has to prove itself.
	for _, dna := range pool.Population {
		if dna.Fitness != 0.25 {
			t.Errorf("transferred DNA should reset fitness to 0.25; got %.3f", dna.Fitness)
		}
	}
}

func TestInitPoolWithTransfer_FallsBackToRandomWhenNoMatch(t *testing.T) {
	rng := rand.New(rand.NewSource(11))

	// New deck has no existing pools to transfer from.
	strats := map[string]*StrategyProfile{
		"first-deck": {
			Archetype:       "Aggro",
			ColorDemand:     map[string]int{"R": 18},
			CommanderThemes: []string{"haste"},
			Bracket:         3,
		},
	}
	pool := InitPoolWithTransfer("first-deck", 3, rng, nil, strats)

	if pool.DeckKey != "first-deck" {
		t.Fatalf("deck key should round-trip even on fallback; got %q", pool.DeckKey)
	}
	if pool.Bracket != 3 {
		t.Fatalf("bracket should round-trip; got %d", pool.Bracket)
	}
	// Random init produces 8 individuals with values across (0,1) — the
	// spread between min and max should be substantial. Transfer with
	// SD=0.10 noise would compress that spread.
	min, max := 1.0, 0.0
	for _, dna := range pool.Population {
		if dna.Aggression < min {
			min = dna.Aggression
		}
		if dna.Aggression > max {
			max = dna.Aggression
		}
	}
	if max-min < 0.30 {
		t.Fatalf("random fallback should span a wide aggression range; got [%.3f, %.3f] = %.3f",
			min, max, max-min)
	}
}

func TestInitPoolWithTransfer_FallsBackWhenNewProfileNil(t *testing.T) {
	rng := rand.New(rand.NewSource(13))
	donor := buildEvolvedPool("evolved", 4, 0.10, 0.90)
	pools := map[string]*CursePool{"evolved": &donor}
	strats := map[string]*StrategyProfile{
		"evolved": {Archetype: "Combo", Bracket: 4},
		// new-deck intentionally absent → no profile → fall back.
	}
	pool := InitPoolWithTransfer("new-deck", 3, rng, pools, strats)

	if pool.DeckKey != "new-deck" {
		t.Fatalf("deck key should round-trip; got %q", pool.DeckKey)
	}
	// With no profile, no donor selected — random init produces wide spread.
	min, max := 1.0, 0.0
	for _, dna := range pool.Population {
		if dna.Aggression < min {
			min = dna.Aggression
		}
		if dna.Aggression > max {
			max = dna.Aggression
		}
	}
	if max-min < 0.30 {
		t.Fatalf("missing-profile should fall back to random; spread was %.3f", max-min)
	}
}

func TestInitPoolWithTransfer_AppliesWiderNoise(t *testing.T) {
	rng := rand.New(rand.NewSource(17))

	donor := buildEvolvedPool("donor", 3, 0.50, 0.50)
	pools := map[string]*CursePool{"donor": &donor}
	strats := map[string]*StrategyProfile{
		"donor": {
			Archetype:       "Midrange",
			ColorDemand:     map[string]int{"G": 14},
			CommanderThemes: []string{"lifegain"},
			Bracket:         3,
		},
		"twin": {
			Archetype:       "Midrange",
			ColorDemand:     map[string]int{"G": 14},
			CommanderThemes: []string{"lifegain"},
			Bracket:         3,
		},
	}
	pool := InitPoolWithTransfer("twin", 3, rng, pools, strats)

	// Sample SD of aggression across the 8-individual transferred pool
	// should track CurseTransferSD (~0.10), comfortably above the
	// per-generation mutation SD (CurseMutationSD = 0.05).
	sd := stddev(aggressionAxis(&pool))
	if sd <= CurseMutationSD {
		t.Fatalf("transfer noise SD should exceed CurseMutationSD; got %.4f vs %.4f",
			sd, CurseMutationSD)
	}
	if sd > 0.20 {
		t.Fatalf("transfer noise SD should not blow past CurseTransferSD by 2x; got %.4f", sd)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildEvolvedPool creates a CursePool that looks like it has played
// enough games to be a viable transfer donor: TotalGames > CurseEvolveAt
// and a single high-fitness individual with the requested aggression /
// combo_patience values.
func buildEvolvedPool(deckKey string, bracket int, aggression, combo float64) CursePool {
	p := CursePool{
		DeckKey:    deckKey,
		Bracket:    bracket,
		TotalGames: CurseEvolveAt + 1,
	}
	for i := range p.Population {
		p.Population[i] = CurseDNA{
			DeckKey:    deckKey,
			Aggression: aggression,
			ComboPat:   combo,
			Fitness:    0.30 + float64(i)*0.01, // monotonic, last is best
		}
	}
	return p
}

func meanAggression(p *CursePool) float64 {
	sum := 0.0
	for _, dna := range p.Population {
		sum += dna.Aggression
	}
	return sum / float64(len(p.Population))
}

func meanCombo(p *CursePool) float64 {
	sum := 0.0
	for _, dna := range p.Population {
		sum += dna.ComboPat
	}
	return sum / float64(len(p.Population))
}

func aggressionAxis(p *CursePool) []float64 {
	out := make([]float64, 0, len(p.Population))
	for _, dna := range p.Population {
		out = append(out, dna.Aggression)
	}
	return out
}

func stddev(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	mean := 0.0
	for _, v := range xs {
		mean += v
	}
	mean /= float64(len(xs))
	sq := 0.0
	for _, v := range xs {
		sq += (v - mean) * (v - mean)
	}
	return math.Sqrt(sq / float64(len(xs)-1))
}
