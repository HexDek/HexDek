package huginn

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func tier3(pattern string, impact float64, obs int, examples ...string) LearnedInteraction {
	return LearnedInteraction{
		Pattern:          pattern,
		ExampleCards:     examples,
		ObservationCount: obs,
		AvgImpactScore:   impact,
		Tier:             TierConfirmed,
	}
}

func chainKeys(chains []FreyaChain) []string {
	keys := make([]string, len(chains))
	for i, c := range chains {
		keys[i] = canonicalChainKey(c.Cards)
	}
	sort.Strings(keys)
	return keys
}

func TestInferChains_NoChainsBelowTier3(t *testing.T) {
	li := LearnedInteraction{
		Pattern:      "produces mana → consumes mana",
		ExampleCards: []string{"A + B", "B + C"},
		Tier:         TierRecurring,
	}
	if got := InferChains([]LearnedInteraction{li}); len(got) != 0 {
		t.Fatalf("non-tier3 inputs must yield zero chains, got %d", len(got))
	}
}

func TestInferChains_ThreeCardChainViaSharedNode(t *testing.T) {
	interactions := []LearnedInteraction{
		tier3("produces mana → consumes mana", 8.0, 6, "A + B"),
		tier3("produces draw → consumes draw", 9.0, 7, "B + C"),
	}
	chains := InferChains(interactions)
	if len(chains) != 1 {
		t.Fatalf("expected 1 chain, got %d: %+v", len(chains), chains)
	}
	c := chains[0]
	if c.Length != 3 {
		t.Errorf("expected length 3, got %d", c.Length)
	}
	if len(c.Patterns) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(c.Patterns))
	}
	if c.MinConfidence != 6 {
		t.Errorf("expected min_confidence=6 (weakest segment), got %d", c.MinConfidence)
	}
	// (8+9)/2 = 8.5
	if c.AvgImpact < 8.49 || c.AvgImpact > 8.51 {
		t.Errorf("expected avg_impact ≈ 8.5, got %f", c.AvgImpact)
	}
	// Chain should travel through the shared card B.
	cards := strings.Join(c.Cards, ",")
	if cards != "A,B,C" && cards != "C,B,A" {
		t.Errorf("expected chain through B, got %v", c.Cards)
	}
}

func TestInferChains_DisconnectedPairsYieldNoChain(t *testing.T) {
	interactions := []LearnedInteraction{
		tier3("produces mana → consumes mana", 7.0, 5, "A + B"),
		tier3("produces draw → consumes draw", 7.0, 5, "C + D"),
	}
	if got := InferChains(interactions); len(got) != 0 {
		t.Fatalf("disconnected pairs must yield no chain, got %d: %+v", len(got), got)
	}
}

func TestInferChains_DedupesReversedTraversal(t *testing.T) {
	interactions := []LearnedInteraction{
		tier3("produces mana → consumes mana", 6.0, 5, "A + B"),
		tier3("produces draw → consumes draw", 6.0, 5, "B + C"),
		tier3("produces tokens → consumes tokens", 6.0, 5, "C + D"),
	}
	chains := InferChains(interactions)
	keys := chainKeys(chains)
	// Expected unique chains: A-B-C, B-C-D, A-B-C-D. No reversed dups.
	want := map[string]bool{"A→B→C": true, "B→C→D": true, "A→B→C→D": true}
	if len(keys) != len(want) {
		t.Fatalf("expected %d unique chains, got %d (%v)", len(want), len(keys), keys)
	}
	for _, k := range keys {
		if !want[k] {
			t.Errorf("unexpected chain key %q", k)
		}
	}
}

func TestInferChains_FiveCardChainAndCap(t *testing.T) {
	// A-B-C-D-E-F linear graph. Should emit chains of length 3,4,5 only.
	interactions := []LearnedInteraction{
		tier3("p1", 5.0, 5, "A + B"),
		tier3("p2", 5.0, 5, "B + C"),
		tier3("p3", 5.0, 5, "C + D"),
		tier3("p4", 5.0, 5, "D + E"),
		tier3("p5", 5.0, 5, "E + F"),
	}
	chains := InferChains(interactions)
	for _, c := range chains {
		if c.Length < chainMinLength || c.Length > chainMaxLength {
			t.Errorf("chain length %d outside [%d,%d]: %v", c.Length, chainMinLength, chainMaxLength, c.Cards)
		}
	}
	keys := chainKeys(chains)
	// Linear graph A-B-C-D-E-F: simple paths of len 3..5 along the line.
	// Length 3: ABC, BCD, CDE, DEF (4)
	// Length 4: ABCD, BCDE, CDEF (3)
	// Length 5: ABCDE, BCDEF (2)
	// Total = 9.
	if len(keys) != 9 {
		t.Fatalf("expected 9 chains for linear graph, got %d: %v", len(keys), keys)
	}
}

func TestInferChains_StrongestEdgeWinsForDuplicatePair(t *testing.T) {
	// Same A+B pair appears in two confirmed patterns. Only the strongest
	// edge should be reflected in chain segments and metadata.
	interactions := []LearnedInteraction{
		tier3("weak pattern", 3.0, 5, "A + B"),
		tier3("strong pattern", 12.0, 5, "A + B"),
		tier3("bridge", 8.0, 5, "B + C"),
	}
	chains := InferChains(interactions)
	if len(chains) != 1 {
		t.Fatalf("expected 1 chain, got %d", len(chains))
	}
	c := chains[0]
	// First segment should reflect strongest edge for A+B.
	// Chain orientation depends on canonicalization (A,B,C vs C,B,A), so
	// scan all patterns to find the A-B edge by elimination.
	foundStrong := false
	for _, p := range c.Patterns {
		if p == "strong pattern" {
			foundStrong = true
		}
		if p == "weak pattern" {
			t.Errorf("weak pattern should have been replaced by strong: %v", c.Patterns)
		}
	}
	if !foundStrong {
		t.Errorf("expected 'strong pattern' segment, got %v", c.Patterns)
	}
}

func TestExportTier3IncludesChains(t *testing.T) {
	dir := t.TempDir()
	interactions := []LearnedInteraction{
		tier3("produces mana → consumes mana", 7.5, 6, "A + B"),
		tier3("produces draw → consumes draw", 8.5, 7, "B + C"),
	}
	if err := exportTier3ForFreya(dir, interactions); err != nil {
		t.Fatalf("export: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, tier3FreyaFile))
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	var exp Tier3Export
	if err := json.Unmarshal(data, &exp); err != nil {
		t.Fatalf("parse export: %v", err)
	}
	if len(exp.Pairs) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(exp.Pairs))
	}
	if len(exp.Chains) != 1 {
		t.Errorf("expected 1 chain, got %d", len(exp.Chains))
	}

	// ReadTier3Export round-trip.
	roundTrip, err := ReadTier3Export(dir)
	if err != nil {
		t.Fatalf("read tier3 export: %v", err)
	}
	if len(roundTrip.Pairs) != 2 || len(roundTrip.Chains) != 1 {
		t.Errorf("round-trip mismatch: pairs=%d chains=%d", len(roundTrip.Pairs), len(roundTrip.Chains))
	}

	// Backward-compat helper still returns just pairs.
	pairs, err := ReadTier3ForFreya(dir)
	if err != nil {
		t.Fatalf("read pairs: %v", err)
	}
	if len(pairs) != 2 {
		t.Errorf("legacy ReadTier3ForFreya: expected 2 pairs, got %d", len(pairs))
	}
}

func TestReadTier3Export_LegacyArrayForm(t *testing.T) {
	dir := t.TempDir()
	legacy := []FreyaInteraction{
		{CardA: "A", CardB: "B", Pattern: "p", AvgImpact: 5.0, Confidence: 5},
	}
	data, _ := json.MarshalIndent(legacy, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, tier3FreyaFile), data, 0o644); err != nil {
		t.Fatalf("write legacy: %v", err)
	}
	exp, err := ReadTier3Export(dir)
	if err != nil {
		t.Fatalf("read legacy: %v", err)
	}
	if len(exp.Pairs) != 1 {
		t.Fatalf("legacy fallback should yield 1 pair, got %d", len(exp.Pairs))
	}
	if len(exp.Chains) != 0 {
		t.Errorf("legacy form has no chains, got %d", len(exp.Chains))
	}
}
