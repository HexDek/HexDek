package hexapi

import "testing"

func TestMatchScore_Tiers(t *testing.T) {
	cases := []struct {
		hay, needle string
		want        int
	}{
		{"krenko, mob boss", "krenko, mob boss", 100},
		{"krenko, mob boss", "krenko", 80},
		{"krenko, mob boss", "mob", 60},  // word-boundary after space
		{"krenko, mob boss", "boss", 60}, // word-boundary after space
		{"krenko, mob boss", "renk", 30}, // pure substring
		{"krenko, mob boss", "atraxa", 0},
		{"", "krenko", 0},
		{"krenko", "", 0},
	}
	for _, tc := range cases {
		if got := matchScore(tc.hay, tc.needle); got != tc.want {
			t.Errorf("matchScore(%q, %q) = %d, want %d", tc.hay, tc.needle, got, tc.want)
		}
	}
}

func TestMatchScore_Prefix_BeatsSubstring(t *testing.T) {
	prefix := matchScore("atraxa praetors voice", "atra")
	subs := matchScore("the atraxa fan club", "atra")
	if prefix <= subs {
		t.Errorf("prefix score (%d) must exceed mid-substring score (%d)",
			prefix, subs)
	}
}

func TestDedupCommanders_KeepsHighestScore(t *testing.T) {
	hits := []SearchResult{
		{Kind: "commander", Label: "Atraxa", Score: 30, Owner: "alice", ID: "deck-1"},
		{Kind: "commander", Label: "Atraxa", Score: 80, Owner: "bob", ID: "deck-2"},
		{Kind: "commander", Label: "Krenko", Score: 60, Owner: "alice", ID: "deck-3"},
	}
	dedupCommanders(&hits)
	if len(hits) != 2 {
		t.Fatalf("expected 2 deduped results, got %d", len(hits))
	}
	for _, h := range hits {
		if h.Label == "Atraxa" && h.Score != 80 {
			t.Errorf("Atraxa dedup should keep score=80, got %d", h.Score)
		}
	}
}

func TestTitleCaseCardName(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"krenko, mob boss", "Krenko, Mob Boss"},
		{"jace, the mind sculptor", "Jace, The Mind Sculptor"},
		{"sword of fire and ice", "Sword Of Fire And Ice"},
		{"birgi, god of storytelling", "Birgi, God Of Storytelling"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := titleCaseCardName(tc.in); got != tc.want {
			t.Errorf("titleCaseCardName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSortByScoreThenLabel(t *testing.T) {
	items := []SearchResult{
		{Label: "B Deck", Score: 50},
		{Label: "A Deck", Score: 80},
		{Label: "C Deck", Score: 80},
	}
	sortByScoreThenLabel(items)
	if items[0].Label != "A Deck" || items[1].Label != "C Deck" || items[2].Label != "B Deck" {
		t.Errorf("unexpected sort order: %+v", items)
	}
}
