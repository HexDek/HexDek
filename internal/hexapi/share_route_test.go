package hexapi

import "testing"

func TestSlugToDeckName(t *testing.T) {
	tests := []struct {
		id    string
		owner string
		want  string
	}{
		{"voltron_uril_b4", "alice", "VOLTRON URIL"},
		{"banding_b4_soraya", "alice", "BANDING B4 SORAYA"},
		{"banding_b4_soraya", "soraya", "BANDING"},
		{"deck_abcd1234", "alice", "DECK"},
		{"plain_name", "alice", "PLAIN NAME"},
		{"", "alice", ""},
	}
	for _, tc := range tests {
		got := slugToDeckName(tc.id, tc.owner)
		if got != tc.want {
			t.Errorf("slugToDeckName(%q, %q) = %q, want %q", tc.id, tc.owner, got, tc.want)
		}
	}
}

func TestBuildShareTitle(t *testing.T) {
	tests := []struct {
		name string
		in   shareDeckMeta
		want string
	}{
		{"deck and commander differ",
			shareDeckMeta{DeckName: "VOLTRON URIL", Commander: "Uril, the Miststalker"},
			"VOLTRON URIL · Uril, the Miststalker"},
		{"deck name equals commander (case-insensitive)",
			shareDeckMeta{DeckName: "URIL, THE MISTSTALKER", Commander: "Uril, the Miststalker"},
			"URIL, THE MISTSTALKER"},
		{"missing deck name falls back to commander",
			shareDeckMeta{Commander: "Uril, the Miststalker"},
			"Uril, the Miststalker"},
		{"missing everything",
			shareDeckMeta{},
			"HEXDEK Deck"},
	}
	for _, tc := range tests {
		got := buildShareTitle(tc.in)
		if got != tc.want {
			t.Errorf("%s: buildShareTitle = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestBuildShareSummary(t *testing.T) {
	tests := []struct {
		name string
		in   shareDeckMeta
		want string
	}{
		{"full triplet",
			shareDeckMeta{Archetype: "voltron", Bracket: 4, Games: 200, Wins: 60, WinRate: 30.0},
			"Voltron · Bracket B4 · 30% WR · 200 games"},
		{"multi-word archetype",
			shareDeckMeta{Archetype: "counters_matter", Bracket: 3, Games: 50, Wins: 15, WinRate: 30.0},
			"Counters Matter · Bracket B3 · 30% WR · 50 games"},
		{"no games — drops WR",
			shareDeckMeta{Archetype: "control", Bracket: 5},
			"Control · Bracket B5"},
		{"no archetype, no bracket — fallback summary",
			shareDeckMeta{DeckName: "MY DECK", Commander: "Foo"},
			"MY DECK · Foo — Commander deck on HEXDEK."},
	}
	for _, tc := range tests {
		got := buildShareSummary(tc.in)
		if got != tc.want {
			t.Errorf("%s: buildShareSummary = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestBuildShareImageURL(t *testing.T) {
	tests := []struct {
		name string
		in   shareDeckMeta
		want string
	}{
		{"commander present",
			shareDeckMeta{Commander: "Uril, the Miststalker"},
			"https://hexdek.dev/api/card-art/Uril%2C%20the%20Miststalker"},
		{"DFC front face used",
			shareDeckMeta{Commander: "Esika, God of the Tree // The Prismatic Bridge"},
			"https://hexdek.dev/api/card-art/Esika%2C%20God%20of%20the%20Tree"},
		{"no commander — default",
			shareDeckMeta{},
			"https://hexdek.dev/og-default.png"},
	}
	for _, tc := range tests {
		got := buildShareImageURL(tc.in)
		if got != tc.want {
			t.Errorf("%s: buildShareImageURL = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestArchetypeLabel(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"voltron", "Voltron"},
		{"counters_matter", "Counters Matter"},
		{"", ""},
		{"ARTIFACTS", "Artifacts"},
	}
	for _, tc := range tests {
		got := archetypeLabel(tc.in)
		if got != tc.want {
			t.Errorf("archetypeLabel(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
