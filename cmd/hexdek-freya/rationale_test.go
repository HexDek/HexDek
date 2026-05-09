package main

import (
	"strings"
	"testing"
)

func TestBuildValueChainRationale_KnownEngine(t *testing.T) {
	chain := &ValueChain{
		Name:        "Aristocrats Engine",
		Depth:       3,
		BridgeCards: []string{"Blood Artist"},
		Steps: []ValueChainStep{
			{Label: "GENERATE", Cards: []string{"Bitterblossom", "Ophiomancer"}},
			{Label: "SACRIFICE", Cards: []string{"Viscera Seer"}},
			{Label: "DRAIN", Cards: []string{"Blood Artist", "Zulaport Cutthroat"}},
		},
	}
	r := buildValueChainRationale(chain)
	if r == nil {
		t.Fatal("rationale is nil")
	}
	if r.Trigger == "" {
		t.Error("expected trigger text for known engine")
	}
	if !strings.Contains(strings.ToLower(r.HowItWorks), "sacrifice") &&
		!strings.Contains(strings.ToLower(r.HowItWorks), "death") {
		t.Errorf("aristocrats how_it_works should mention sacrifice/death, got %q", r.HowItWorks)
	}
	if len(r.KeyPieces) == 0 {
		t.Error("expected at least one key piece")
	}
	if r.KeyPieces[0] != "Blood Artist" {
		t.Errorf("expected bridge card first in KeyPieces, got %v", r.KeyPieces)
	}
}

func TestBuildValueChainRationale_UnknownEngine(t *testing.T) {
	chain := &ValueChain{
		Name:  "Made Up Engine",
		Depth: 2,
		Steps: []ValueChainStep{
			{Label: "A", Cards: []string{"Card A"}},
			{Label: "B", Cards: []string{"Card B"}},
		},
	}
	r := buildValueChainRationale(chain)
	if r == nil || r.Trigger == "" || r.HowItWorks == "" {
		t.Fatal("rationale fields should fall back to a generic description for unknown engine names")
	}
	if len(r.KeyPieces) == 0 {
		t.Error("expected key pieces from steps when bridges are empty")
	}
}

func TestBuildComboWinLineRationale_SplitsDescriptionTags(t *testing.T) {
	wl := &WinLine{
		Pieces: []string{"Thassa's Oracle", "Demonic Consultation"},
		Type:   "infinite",
		Desc:   "Cast Consultation naming a card not in deck, exile library. Oracle ETB with empty library = win. | OUTLETS IN DECK: Demonic Consultation",
		TutorPaths: []TutorChain{
			{Tutor: "Vampiric Tutor", Finds: "Thassa's Oracle"},
			{Tutor: "Demonic Tutor", Finds: "Thassa's Oracle"},
		},
	}
	r := buildComboWinLineRationale(wl)
	if r == nil {
		t.Fatal("rationale nil")
	}
	if len(r.Forms) != 2 {
		t.Errorf("expected 2 forms, got %d", len(r.Forms))
	}
	if len(r.Resolves) == 0 || !strings.Contains(r.Resolves[0], "Consultation") {
		t.Errorf("first resolve should be the mechanism sentence, got %v", r.Resolves)
	}
	// The OUTLETS tag should land in conditions, not be discarded.
	foundOutlets := false
	for _, c := range r.Conditions {
		if strings.Contains(c, "OUTLETS") {
			foundOutlets = true
		}
	}
	if !foundOutlets {
		t.Error("expected OUTLETS IN DECK to appear in conditions")
	}
	// Tutor coverage summary should mention the two tutors that find Oracle.
	tutorCoverageFound := false
	for _, c := range r.Conditions {
		if strings.Contains(c, "Thassa's Oracle") && strings.Contains(c, "Vampiric Tutor") {
			tutorCoverageFound = true
		}
	}
	if !tutorCoverageFound {
		t.Errorf("expected per-piece tutor coverage line, got conditions=%v", r.Conditions)
	}
	// The second piece has no tutor in this test — should be flagged.
	naturalDrawFlag := false
	for _, c := range r.Conditions {
		if strings.Contains(c, "Demonic Consultation") && strings.Contains(c, "must draw naturally") {
			naturalDrawFlag = true
		}
	}
	if !naturalDrawFlag {
		t.Errorf("expected natural-draw flag for piece without tutor, got conditions=%v", r.Conditions)
	}
}

func TestBuildComboWinLineRationale_NoTutors(t *testing.T) {
	wl := &WinLine{
		Pieces: []string{"Card A", "Card B"},
		Type:   "infinite",
		Desc:   "Some loop description.",
	}
	r := buildComboWinLineRationale(wl)
	noTutorFlag := false
	for _, c := range r.Conditions {
		if strings.Contains(c, "No tutor support") {
			noTutorFlag = true
		}
	}
	if !noTutorFlag {
		t.Errorf("expected no-tutor warning when win line has no tutor paths, got %v", r.Conditions)
	}
}

func TestAltWinconResolves(t *testing.T) {
	cases := []struct {
		ot       string
		contains string
	}{
		{"target player wins the game", "wins the game"},
		{"each opponent loses the game", "lose the game"},
		{"each opponent loses 1 life for each", "drains"},
		{"do unrelated stuff", "alternate win condition"},
	}
	for _, c := range cases {
		got := altWinconResolves("Test Card", c.ot)
		if !strings.Contains(got, c.contains) {
			t.Errorf("oracle %q: resolve text should mention %q, got %q", c.ot, c.contains, got)
		}
	}
}

func TestSuggestCuttableSwaps_ReflectsDeckGaps(t *testing.T) {
	dp := &DeckProfile{
		PrimaryArchetype: "Combo",
		PrimaryWinLine:   "Thassa's Oracle + Consultation",
		WinLineCount:     1,
	}
	report := &FreyaReport{
		Stats: &DeckStatistics{
			RampCount:       6, // below 10
			DrawSourceCount: 12,
		},
	}
	swaps := suggestCuttableSwaps(dp, report)
	hasRamp := false
	hasArchetype := false
	hasInteraction := false
	for _, s := range swaps {
		if strings.Contains(s, "ramp") {
			hasRamp = true
		}
		if strings.Contains(s, "Combo") {
			hasArchetype = true
		}
		if strings.Contains(s, "interaction") {
			hasInteraction = true
		}
	}
	if !hasRamp {
		t.Error("expected ramp suggestion when deck has <10 ramp sources")
	}
	if !hasArchetype {
		t.Error("expected archetype-flavoured suggestion")
	}
	if !hasInteraction {
		t.Error("expected interaction suggestion in default tail")
	}

	// When ramp/draw are healthy, those suggestions should not appear.
	report.Stats.RampCount = 12
	report.Stats.DrawSourceCount = 12
	swaps = suggestCuttableSwaps(dp, report)
	for _, s := range swaps {
		if strings.Contains(s, "ramp piece") {
			t.Errorf("did not expect ramp suggestion when deck has enough ramp, got %v", swaps)
		}
	}
}

func TestPluralS(t *testing.T) {
	if pluralS(1) != "" {
		t.Errorf("pluralS(1) should be empty string")
	}
	if pluralS(0) != "s" || pluralS(2) != "s" {
		t.Errorf("pluralS(non-1) should be 's'")
	}
}
