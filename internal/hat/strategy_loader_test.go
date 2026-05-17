package hat

import "testing"

// Freya only serializes 8 of the hat's 20 eval-weight dimensions. The
// loader must merge them into the full archetype profile so the other 12
// dimensions (StaxLockProgress, DrainEngine, ArtifactSynergy, ...) keep
// their archetype-appropriate values rather than being zeroed out.
func TestBuildFromStrategyJSON_SparseWeightsMergeWithArchetypeDefaults(t *testing.T) {
	sj := &strategyFileJSON{
		Archetype: ArchetypeStax,
		Weights: &freyaEvalWeights{
			BoardPresence:     0.7,
			CardAdvantage:     1.2,
			ManaAdvantage:     1.0,
			LifeResource:      0.5,
			ComboProximity:    0.3,
			ThreatExposure:    1.5,
			CommanderProgress: 0.8,
			GraveyardValue:    0.4,
		},
	}

	sp := buildFromStrategyJSON(sj)
	if sp == nil || sp.Weights == nil {
		t.Fatal("expected weights to be populated")
	}

	if sp.Weights.BoardPresence != 0.7 {
		t.Errorf("BoardPresence: got %.2f, want 0.7 (from Freya)", sp.Weights.BoardPresence)
	}
	if sp.Weights.ThreatExposure != 1.5 {
		t.Errorf("ThreatExposure: got %.2f, want 1.5 (from Freya)", sp.Weights.ThreatExposure)
	}

	stax := DefaultWeightsForArchetype(ArchetypeStax)
	if sp.Weights.StaxLockProgress != stax.StaxLockProgress {
		t.Errorf("StaxLockProgress zero-filled instead of inherited from stax default: got %.2f, want %.2f",
			sp.Weights.StaxLockProgress, stax.StaxLockProgress)
	}
	if sp.Weights.StackInteraction != stax.StackInteraction {
		t.Errorf("StackInteraction: got %.2f, want %.2f (from stax default)",
			sp.Weights.StackInteraction, stax.StackInteraction)
	}
	if sp.Weights.OpponentGraveyardThreat != stax.OpponentGraveyardThreat {
		t.Errorf("OpponentGraveyardThreat: got %.2f, want %.2f (from stax default)",
			sp.Weights.OpponentGraveyardThreat, stax.OpponentGraveyardThreat)
	}
}

// Lifegain decks in the moxfield corpus previously fell back to midrange
// weights because the hat had no lifegain profile. Confirm Freya's sparse
// `life_resource: 1.8` plus the new lifegain default fills in DrainEngine
// from the lifegain profile, not zero.
func TestBuildFromStrategyJSON_LifegainGetsLifegainDimensions(t *testing.T) {
	sj := &strategyFileJSON{
		Archetype: ArchetypeLifegain,
		Weights: &freyaEvalWeights{
			BoardPresence:     0.9,
			CardAdvantage:     0.7,
			ManaAdvantage:     0.5,
			LifeResource:      1.8,
			ComboProximity:    0.8,
			ThreatExposure:    0.6,
			CommanderProgress: 0.6,
			GraveyardValue:    0.3,
		},
	}

	sp := buildFromStrategyJSON(sj)
	if sp == nil || sp.Weights == nil {
		t.Fatal("expected weights to be populated")
	}

	life := DefaultWeightsForArchetype(ArchetypeLifegain)
	if sp.Weights.DrainEngine != life.DrainEngine {
		t.Errorf("DrainEngine: got %.2f, want %.2f (lifegain default)",
			sp.Weights.DrainEngine, life.DrainEngine)
	}
	if sp.Weights.LifeResource != 1.8 {
		t.Errorf("LifeResource: got %.2f, want 1.8 (from Freya)", sp.Weights.LifeResource)
	}
}
