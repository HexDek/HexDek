package muninn

import (
	"testing"
)

func TestAutoArchiveViolation_AppendsAndParses(t *testing.T) {
	dir := t.TempDir()
	deckKeys := [4]string{"alpha", "beta", "gamma", "delta"}
	violations := []string{
		"[critical] 704.5a (seat 1): seat 1 has -3 life but is not marked lost",
		"[warning] zone_accounting (seat 2): seat 2 has 95 cards (expected ~100, diff=-5)",
	}

	if err := AutoArchiveViolation(dir, 1234, deckKeys, violations); err != nil {
		t.Fatalf("AutoArchiveViolation: %v", err)
	}

	got, err := ReadInvariantViolations(dir)
	if err != nil {
		t.Fatalf("ReadInvariantViolations: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 records, got %d", len(got))
	}

	if got[0].GameSeed != 1234 {
		t.Errorf("record 0: GameSeed=%d, want 1234", got[0].GameSeed)
	}
	if got[0].DeckKeys != deckKeys {
		t.Errorf("record 0: DeckKeys=%v, want %v", got[0].DeckKeys, deckKeys)
	}
	if got[0].ViolationType != "704.5a" {
		t.Errorf("record 0: ViolationType=%q, want %q", got[0].ViolationType, "704.5a")
	}
	if got[0].Message != violations[0] {
		t.Errorf("record 0: Message=%q, want %q", got[0].Message, violations[0])
	}
	if got[0].Timestamp == "" {
		t.Error("record 0: Timestamp is empty")
	}

	if got[1].ViolationType != "zone_accounting" {
		t.Errorf("record 1: ViolationType=%q, want %q", got[1].ViolationType, "zone_accounting")
	}
}

func TestAutoArchiveViolation_AppendsToExisting(t *testing.T) {
	dir := t.TempDir()
	deckA := [4]string{"a", "b", "c", "d"}
	deckB := [4]string{"e", "f", "g", "h"}

	if err := AutoArchiveViolation(dir, 1, deckA, []string{"[critical] 704.5f (seat 0): toughness 0"}); err != nil {
		t.Fatalf("first archive: %v", err)
	}
	if err := AutoArchiveViolation(dir, 2, deckB, []string{"[critical] 704.5c (seat 1): poison 12"}); err != nil {
		t.Fatalf("second archive: %v", err)
	}

	got, err := ReadInvariantViolations(dir)
	if err != nil {
		t.Fatalf("ReadInvariantViolations: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 records after two appends, got %d", len(got))
	}
	if got[0].GameSeed != 1 || got[1].GameSeed != 2 {
		t.Errorf("seed order wrong: %d, %d", got[0].GameSeed, got[1].GameSeed)
	}
}

func TestAutoArchiveViolation_EmptyIsNoop(t *testing.T) {
	dir := t.TempDir()
	deck := [4]string{}

	if err := AutoArchiveViolation(dir, 99, deck, nil); err != nil {
		t.Fatalf("nil slice: %v", err)
	}
	if err := AutoArchiveViolation(dir, 99, deck, []string{}); err != nil {
		t.Fatalf("empty slice: %v", err)
	}

	got, err := ReadInvariantViolations(dir)
	if err != nil {
		t.Fatalf("ReadInvariantViolations: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no records, got %d", len(got))
	}
}

func TestParseOracleViolation(t *testing.T) {
	cases := []struct {
		in       string
		wantType string
	}{
		{"[critical] 704.5a (seat 1): description here", "704.5a"},
		{"[warning] turn_bound (seat -1): runaway", "turn_bound"},
		{"[critical] 603.3b (seat -1): APNAP order", "603.3b"},
		{"plain string with no format", ""},
		{"[critical] mana_accounting: no seat tag", "mana_accounting"},
	}
	for _, tc := range cases {
		gotType, gotMsg := parseOracleViolation(tc.in)
		if gotType != tc.wantType {
			t.Errorf("parseOracleViolation(%q): type=%q, want %q", tc.in, gotType, tc.wantType)
		}
		if gotMsg != tc.in {
			t.Errorf("parseOracleViolation(%q): msg=%q, want full input", tc.in, gotMsg)
		}
	}
}
