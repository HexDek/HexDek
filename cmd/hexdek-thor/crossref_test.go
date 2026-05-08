package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// normalizeCrossRef tests
// ---------------------------------------------------------------------------

func TestNormalizeCrossRef(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Lightning Bolt", "lightning bolt"},
		{"Niv-Mizzet, the Firemind", "nivmizzet the firemind"},
		{"  Jace's  Erasure  ", "jaces erasure"},
		{"Nicol Bolas, the Ravager // Nicol Bolas, the Arisen", "nicol bolas the ravager // nicol bolas the arisen"},
		{"Who // What // When // Where // Why", "who // what // when // where // why"},
		{"", ""},
		{"Sol Ring", "sol ring"},
		{"Korvold, Fae-Cursed King", "korvold faecursed king"},
		{"Anje Falkenrath", "anje falkenrath"},
		// Unicode punctuation.
		{"Teysa, Orzhov Scion", "teysa orzhov scion"},
		{"Riku’s Reflection", "rikus reflection"}, // right single quotation mark
	}
	for _, tt := range tests {
		got := normalizeCrossRef(tt.input)
		if got != tt.want {
			t.Errorf("normalizeCrossRef(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Mock Muninn data helpers
// ---------------------------------------------------------------------------

func writeMockMuninn(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	// parser_gaps.json
	gaps := []map[string]interface{}{
		{
			"snippet":    "Chainer, Nightmare Adept",
			"count":      3,
			"first_seen": "2026-05-01T00:00:00Z",
			"last_seen":  "2026-05-01T00:00:00Z",
		},
		{
			"snippet":    "Tiamat",
			"count":      13,
			"first_seen": "2026-05-01T00:00:00Z",
			"last_seen":  "2026-05-01T00:00:00Z",
		},
		{
			"snippet":    "Smirking Spelljacker",
			"count":      4,
			"first_seen": "2026-05-01T00:00:00Z",
			"last_seen":  "2026-05-01T00:00:00Z",
		},
	}
	writeJSON(t, filepath.Join(dir, "parser_gaps.json"), gaps)

	// dead_triggers.json
	triggers := []map[string]interface{}{
		{
			"trigger_name": "triggered_ability",
			"card_name":    "The One Ring",
			"count":        58,
			"games_seen":   2,
			"last_seen":    "2026-05-01T00:00:00Z",
		},
		{
			"trigger_name": "triggered_ability",
			"card_name":    "Sylvan Library",
			"count":        64,
			"games_seen":   4,
			"last_seen":    "2026-05-01T00:00:00Z",
		},
	}
	writeJSON(t, filepath.Join(dir, "dead_triggers.json"), triggers)

	// crashes.json — empty (no per_card references in stack trace)
	writeJSON(t, filepath.Join(dir, "crashes.json"), []interface{}{})

	// invariant_violations.json — present but not used for card-level matching
	writeJSON(t, filepath.Join(dir, "invariant_violations.json"), []interface{}{})
}

func writeJSON(t *testing.T, path string, data interface{}) {
	t.Helper()
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// RunCrossRef tests
// ---------------------------------------------------------------------------

func TestRunCrossRef_BasicSetOperations(t *testing.T) {
	dir := t.TempDir()
	muninnDir := filepath.Join(dir, "muninn")
	writeMockMuninn(t, muninnDir)

	// Thor failures: some overlap with Muninn, some not.
	thorFailures := []failure{
		// Matches Muninn: "Tiamat" (parser_gap)
		{CardName: "Tiamat", Interaction: "destroy", Invariant: "zone_accounting", Message: "zone mismatch"},
		// Matches Muninn: "The One Ring" (dead_trigger)
		{CardName: "The One Ring", Interaction: "phase_upkeep", Panicked: true, PanicMsg: "nil pointer"},
		// Thor-only: "Lightning Bolt" (not in Muninn)
		{CardName: "Lightning Bolt", Interaction: "counter_mod_plus", Invariant: "life_sanity", Message: "life < 0"},
		// Thor-only: "Sol Ring" (not in Muninn)
		{CardName: "Sol Ring", Interaction: "sacrifice", Invariant: "zone_accounting", Message: "card vanished"},
	}

	result, err := RunCrossRef(thorFailures, muninnDir)
	if err != nil {
		t.Fatal(err)
	}

	// True Positives: Tiamat, The One Ring
	if len(result.TruePositives) != 2 {
		t.Errorf("TruePositives: got %d, want 2", len(result.TruePositives))
		for _, tp := range result.TruePositives {
			t.Logf("  TP: %s", tp.CardName)
		}
	}

	// False Negatives: Chainer, Smirking Spelljacker, Sylvan Library (in Muninn, not Thor)
	if len(result.FalseNegatives) != 3 {
		t.Errorf("FalseNegatives: got %d, want 3", len(result.FalseNegatives))
		for _, fn := range result.FalseNegatives {
			t.Logf("  FN: %s", fn.CardName)
		}
	}

	// False Positives: Lightning Bolt, Sol Ring (in Thor, not Muninn)
	if len(result.FalsePositives) != 2 {
		t.Errorf("FalsePositives: got %d, want 2", len(result.FalsePositives))
		for _, fp := range result.FalsePositives {
			t.Logf("  FP: %s", fp.CardName)
		}
	}

	// Totals.
	if result.ThorTotal != 4 {
		t.Errorf("ThorTotal: got %d, want 4", result.ThorTotal)
	}
	if result.MuninnTotal != 5 {
		t.Errorf("MuninnTotal: got %d, want 5", result.MuninnTotal)
	}
}

func TestRunCrossRef_NormalizationMatches(t *testing.T) {
	dir := t.TempDir()
	muninnDir := filepath.Join(dir, "muninn")
	if err := os.MkdirAll(muninnDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Muninn has "Korvold, Fae-Cursed King" — note comma and hyphen.
	gaps := []map[string]interface{}{
		{
			"snippet":    "Korvold, Fae-Cursed King",
			"count":      1,
			"first_seen": "2026-05-01T00:00:00Z",
			"last_seen":  "2026-05-01T00:00:00Z",
		},
	}
	writeJSON(t, filepath.Join(muninnDir, "parser_gaps.json"), gaps)
	writeJSON(t, filepath.Join(muninnDir, "dead_triggers.json"), []interface{}{})
	writeJSON(t, filepath.Join(muninnDir, "crashes.json"), []interface{}{})

	// Thor has same card but slightly different formatting.
	thorFailures := []failure{
		{CardName: "Korvold, Fae-Cursed King", Interaction: "destroy", Message: "boom"},
	}

	result, err := RunCrossRef(thorFailures, muninnDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.TruePositives) != 1 {
		t.Errorf("expected 1 true positive, got %d", len(result.TruePositives))
	}
	if len(result.FalseNegatives) != 0 {
		t.Errorf("expected 0 false negatives, got %d", len(result.FalseNegatives))
	}
	if len(result.FalsePositives) != 0 {
		t.Errorf("expected 0 false positives, got %d", len(result.FalsePositives))
	}
}

func TestRunCrossRef_EmptyInputs(t *testing.T) {
	dir := t.TempDir()
	muninnDir := filepath.Join(dir, "muninn")
	if err := os.MkdirAll(muninnDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Empty Muninn files.
	writeJSON(t, filepath.Join(muninnDir, "parser_gaps.json"), []interface{}{})
	writeJSON(t, filepath.Join(muninnDir, "dead_triggers.json"), []interface{}{})
	writeJSON(t, filepath.Join(muninnDir, "crashes.json"), []interface{}{})

	result, err := RunCrossRef(nil, muninnDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.TruePositives) != 0 || len(result.FalseNegatives) != 0 || len(result.FalsePositives) != 0 {
		t.Errorf("expected all empty, got TP=%d FN=%d FP=%d",
			len(result.TruePositives), len(result.FalseNegatives), len(result.FalsePositives))
	}
}

func TestRunCrossRef_MissingMuninnDir(t *testing.T) {
	// Muninn dir doesn't exist — ReadParserGaps etc. return empty, no error.
	result, err := RunCrossRef([]failure{
		{CardName: "Sol Ring", Interaction: "tap"},
	}, filepath.Join(t.TempDir(), "nonexistent"))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.FalsePositives) != 1 {
		t.Errorf("expected 1 false positive, got %d", len(result.FalsePositives))
	}
}

func TestRunCrossRef_DuplicateThorFailures(t *testing.T) {
	dir := t.TempDir()
	muninnDir := filepath.Join(dir, "muninn")
	if err := os.MkdirAll(muninnDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(muninnDir, "parser_gaps.json"), []interface{}{})
	writeJSON(t, filepath.Join(muninnDir, "dead_triggers.json"), []interface{}{})
	writeJSON(t, filepath.Join(muninnDir, "crashes.json"), []interface{}{})

	// Same card with multiple Thor failures — should consolidate to one entry.
	thorFailures := []failure{
		{CardName: "Sol Ring", Interaction: "destroy", Invariant: "zone", Message: "fail1"},
		{CardName: "Sol Ring", Interaction: "exile", Invariant: "zone", Message: "fail2"},
		{CardName: "Sol Ring", Interaction: "bounce", Panicked: true, PanicMsg: "panic!"},
	}

	result, err := RunCrossRef(thorFailures, muninnDir)
	if err != nil {
		t.Fatal(err)
	}

	if result.ThorTotal != 1 {
		t.Errorf("ThorTotal: got %d, want 1 (should deduplicate)", result.ThorTotal)
	}
	if len(result.FalsePositives) != 1 {
		t.Errorf("FalsePositives: got %d, want 1", len(result.FalsePositives))
	}
	// The issue should contain all three failure descriptions.
	if len(result.FalsePositives) > 0 {
		issue := result.FalsePositives[0].ThorIssue
		if !strings.Contains(issue, "destroy") || !strings.Contains(issue, "exile") || !strings.Contains(issue, "PANIC") {
			t.Errorf("issue should contain all failures: %s", issue)
		}
	}
}

// ---------------------------------------------------------------------------
// Report writing test
// ---------------------------------------------------------------------------

func TestWriteCrossRefReport(t *testing.T) {
	result := &CrossRefResult{
		TruePositives: []CrossRefEntry{
			{CardName: "Tiamat", MuninnIssue: "parser_gap (count=13)", ThorIssue: "zone_accounting@destroy: zone mismatch"},
		},
		FalseNegatives: []CrossRefEntry{
			{CardName: "Chainer, Nightmare Adept", MuninnIssue: "parser_gap (count=3)"},
		},
		FalsePositives: []CrossRefEntry{
			{CardName: "Lightning Bolt", ThorIssue: "life_sanity@counter_mod_plus: life < 0"},
		},
		ThorTotal:   2,
		MuninnTotal: 2,
	}

	path := filepath.Join(t.TempDir(), "report.md")
	if err := WriteCrossRefReport(path, result); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Verify structure.
	checks := []string{
		"Thor vs Muninn Cross-Reference Report",
		"True Positives (1)",
		"False Negatives (1)",
		"False Positives (1)",
		"Tiamat",
		"Chainer",
		"Lightning Bolt",
		"Thor recall:",
		"Thor precision:",
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("report missing %q", want)
		}
	}
}

func TestWriteCrossRefReport_Empty(t *testing.T) {
	result := &CrossRefResult{}

	path := filepath.Join(t.TempDir(), "report.md")
	if err := WriteCrossRefReport(path, result); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "No overlap found") {
		t.Error("empty report should say no overlap")
	}
	if !strings.Contains(content, "Thor catches everything") {
		t.Error("empty report should note Thor catches everything")
	}
	if !strings.Contains(content, "All Thor failures are confirmed") {
		t.Error("empty report should note all confirmed")
	}
}

// ---------------------------------------------------------------------------
// Helper function tests
// ---------------------------------------------------------------------------

func TestTruncate(t *testing.T) {
	if got := truncate("hello world", 5); got != "hello..." {
		t.Errorf("truncate short: got %q", got)
	}
	if got := truncate("hi", 5); got != "hi" {
		t.Errorf("truncate no-op: got %q", got)
	}
	if got := truncate("", 5); got != "" {
		t.Errorf("truncate empty: got %q", got)
	}
}

func TestEscMd(t *testing.T) {
	if got := escMd("a | b"); got != "a \\| b" {
		t.Errorf("escMd pipe: got %q", got)
	}
	if got := escMd("line1\nline2"); got != "line1 line2" {
		t.Errorf("escMd newline: got %q", got)
	}
}
