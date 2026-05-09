package hat

import (
	"math/rand"
	"strings"
)

// CurseTransferSD is the gaussian noise applied to each trait when
// seeding a new pool from a transferred CurseDNA. Wider than the
// per-generation mutation SD (CurseMutationSD = 0.05) so the new pool
// still explores meaningful neighborhood — without it the population
// would be 8 near-identical clones of the donor.
const CurseTransferSD = 0.10

// InitPoolWithTransfer creates a fresh population for a new deck,
// seeded from the closest-archetype existing pool's top-fitness DNA
// instead of pure random. With ~1300 evolved pools in production, a
// new deck inherits useful personality axes from a similar deck and
// converges in ~10 games rather than the ~100 games random seeding
// requires.
//
// existingPools may be nil or empty (first deck ever, or no strategy
// data available); in those cases this falls back to InitPoolWithBracket
// so the caller gets a valid pool either way.
//
// strategyMap should be the showmatch's deck-key → *StrategyProfile
// table. The new deck's profile (looked up by deckKey) is compared
// against every existing pool's deck profile to find the closest match.
func InitPoolWithTransfer(
	deckKey string,
	bracket int,
	rng *rand.Rand,
	existingPools map[string]*CursePool,
	strategyMap map[string]*StrategyProfile,
) CursePool {
	newProfile := strategyMap[deckKey]
	donor := pickTransferDonor(deckKey, bracket, newProfile, existingPools, strategyMap)
	if donor == nil {
		return InitPoolWithBracket(deckKey, bracket, rng)
	}
	return seedPoolFromDNA(deckKey, bracket, rng, donor)
}

// pickTransferDonor selects the highest-fitness CurseDNA from the
// existing pool whose deck is closest in archetype to the new deck.
// Returns nil when no usable donor exists (no profile available, no
// other pools, etc.).
func pickTransferDonor(
	newKey string,
	newBracket int,
	newProfile *StrategyProfile,
	existingPools map[string]*CursePool,
	strategyMap map[string]*StrategyProfile,
) *CurseDNA {
	if newProfile == nil || len(existingPools) == 0 {
		return nil
	}
	bestPool := (*CursePool)(nil)
	bestDist := 1.01 // strictly > the max archetypeDistance can return
	for key, pool := range existingPools {
		if pool == nil || key == newKey {
			continue
		}
		// Skip pools that haven't played enough games to have meaningful
		// fitness — transferring random noise from a fresh pool defeats
		// the purpose. CurseEvolveAt is the per-generation threshold;
		// we accept anything that's seen at least one full evolution.
		if pool.TotalGames < CurseEvolveAt {
			continue
		}
		otherProfile := strategyMap[key]
		if otherProfile == nil {
			continue
		}
		d := archetypeDistance(newProfile, otherProfile)
		// Bracket is a weak tiebreaker — prefer the same bracket when
		// the archetype distance is essentially tied.
		if pool.Bracket != newBracket {
			d += 0.05
		}
		if d < bestDist {
			bestDist = d
			bestPool = pool
		}
	}
	if bestPool == nil {
		return nil
	}
	// Pick the highest-fitness DNA in the donor pool.
	best := &bestPool.Population[0]
	for i := 1; i < len(bestPool.Population); i++ {
		if bestPool.Population[i].Fitness > best.Fitness {
			best = &bestPool.Population[i]
		}
	}
	return best
}

// seedPoolFromDNA builds a CursePool whose 8 individuals are the donor
// DNA's trait vector plus per-trait gaussian noise (SD=CurseTransferSD).
// All values are clamped into [0,1].
func seedPoolFromDNA(deckKey string, bracket int, rng *rand.Rand, donor *CurseDNA) CursePool {
	pool := CursePool{DeckKey: deckKey, Bracket: bracket, rng: rng}
	jitter := func(v float64) float64 {
		return clampUnit(v + rng.NormFloat64()*CurseTransferSD)
	}
	for i := range pool.Population {
		pool.Population[i] = CurseDNA{
			DeckKey:               deckKey,
			Aggression:            jitter(donor.Aggression),
			ComboPat:              jitter(donor.ComboPat),
			ThreatParanoia:        jitter(donor.ThreatParanoia),
			ResourceGreed:         jitter(donor.ResourceGreed),
			PoliticalMemory:       jitter(donor.PoliticalMemory),
			DrainAffinity:         jitter(donor.DrainAffinity),
			ArtifactAffinity:      jitter(donor.ArtifactAffinity),
			LandGreed:             jitter(donor.LandGreed),
			EquipmentAffinity:     jitter(donor.EquipmentAffinity),
			GraveyardExploitation: jitter(donor.GraveyardExploitation),
			CounterplayTiming:     jitter(donor.CounterplayTiming),
			TokenPressure:         jitter(donor.TokenPressure),
			// Fitness intentionally seeded at the baseline (NOT carried
			// over): the new deck still has to prove itself in real games.
			Fitness: 0.25,
		}
	}
	return pool
}

// archetypeDistance returns 0.0 when two strategy profiles are
// effectively identical and 1.0 when they share nothing. Composition:
//
//	archetype mismatch        — weight 0.5
//	color identity disjoint   — weight 0.3 (1 − jaccard over color demand)
//	commander themes disjoint — weight 0.2 (1 − jaccard over themes)
//
// Either profile being nil yields 1.0 (no signal — treat as max
// distance so the caller falls back to random init).
func archetypeDistance(a, b *StrategyProfile) float64 {
	if a == nil || b == nil {
		return 1.0
	}
	d := 0.0
	if !sameArchetype(a.Archetype, b.Archetype) {
		d += 0.5
	}
	d += 0.3 * (1.0 - colorJaccard(a.ColorDemand, b.ColorDemand))
	d += 0.2 * (1.0 - themeJaccard(a.CommanderThemes, b.CommanderThemes))
	if d < 0 {
		return 0
	}
	if d > 1 {
		return 1
	}
	return d
}

// sameArchetype compares archetype labels case-insensitively. Empty
// labels never match (signal is absent on both sides).
func sameArchetype(a, b string) bool {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	if a == "" || b == "" {
		return false
	}
	return a == b
}

// colorJaccard returns the Jaccard similarity of two color identity
// sets, derived from non-zero ColorDemand keys. 1.0 means identical
// color identity, 0.0 means completely disjoint. Both maps empty
// returns 1.0 (colorless decks are also "the same color identity").
func colorJaccard(a, b map[string]int) float64 {
	aset := nonZeroColorKeys(a)
	bset := nonZeroColorKeys(b)
	return jaccard(aset, bset)
}

// themeJaccard returns the Jaccard similarity over commander themes,
// case-insensitive.
func themeJaccard(a, b []string) float64 {
	aset := lowerSet(a)
	bset := lowerSet(b)
	return jaccard(aset, bset)
}

// jaccard returns |A ∩ B| / |A ∪ B|. Both empty sets yield 1.0 —
// "no theme data on either side" should not penalize the match.
func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	inter := 0
	for k := range a {
		if _, ok := b[k]; ok {
			inter++
		}
	}
	union := len(a) + len(b) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

func nonZeroColorKeys(m map[string]int) map[string]struct{} {
	out := map[string]struct{}{}
	for k, v := range m {
		if v <= 0 {
			continue
		}
		out[strings.ToUpper(strings.TrimSpace(k))] = struct{}{}
	}
	return out
}

func lowerSet(s []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, v := range s {
		k := strings.ToLower(strings.TrimSpace(v))
		if k == "" {
			continue
		}
		out[k] = struct{}{}
	}
	return out
}
