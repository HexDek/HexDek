package hexapi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeckContainsCard_TextDeck(t *testing.T) {
	dir := t.TempDir()
	deck := filepath.Join(dir, "test.txt")
	contents := `COMMANDER: Krenko, Mob Boss
1 Sol Ring
1 Dramatic Reversal (CMR)
20 Mountain
1 Goblin Chieftain
`
	if err := os.WriteFile(deck, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		needle string
		want   bool
	}{
		{"sol ring", true},
		{"sol", true},               // substring within name
		{"dramatic reversal", true}, // set-code suffix stripped
		{"mountain", true},
		{"krenko", true},               // commander matched
		{"krenko, mob boss", true},     // full commander
		{"island", false},              // absent
		{"goblin chieftain", true},     // last line
		{"GOBLIN", true},               // case-insensitive caller (already lowered)
	}
	for _, tc := range cases {
		needle := tc.needle
		// API normalises before calling; mirror that.
		needle = lowerASCII(needle)
		if got := deckContainsCard(deck, needle); got != tc.want {
			t.Errorf("deckContainsCard(_, %q) = %v, want %v", tc.needle, got, tc.want)
		}
	}
}

func TestDeckContainsCard_EmptyNeedle(t *testing.T) {
	if deckContainsCard("/tmp/anything", "") {
		t.Fatal("empty needle must return false")
	}
}

func TestDeckContainsCard_MissingFile(t *testing.T) {
	if deckContainsCard(filepath.Join(t.TempDir(), "nope.txt"), "anything") {
		t.Fatal("missing file must return false")
	}
}

// lowerASCII is a local helper so the test does not pull in strings.
func lowerASCII(s string) string {
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		out[i] = c
	}
	return string(out)
}
