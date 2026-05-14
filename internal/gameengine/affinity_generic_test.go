package gameengine

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// makeAffinityCard constructs a minimal Card with a single Keyword
// ability whose Name + Raw text encode an "affinity for <type>" clause.
// Used by the table-driven tests below.
func makeAffinityCard(keywordName, raw string) *Card {
	return &Card{
		Types: []string{},
		AST: &gameast.CardAST{
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: keywordName, Raw: raw},
			},
		},
	}
}

func TestAffinityForType_Detection(t *testing.T) {
	cases := []struct {
		name       string
		keywordNm  string
		raw        string
		wantHas    bool
		wantType   string
	}{
		{
			name:      "explicit affinity for artifacts (keyword name form)",
			keywordNm: "affinity for artifacts",
			raw:       "Affinity for artifacts (this spell costs {1} less to cast for each artifact you control.)",
			wantHas:   true,
			wantType:  "artifact",
		},
		{
			name:      "affinity for humans — Riders of the Mark style",
			keywordNm: "affinity for humans",
			raw:       "Affinity for Humans",
			wantHas:   true,
			wantType:  "human",
		},
		{
			name:      "affinity for knights — hypothetical tribal",
			keywordNm: "affinity for knights",
			raw:       "Affinity for Knights",
			wantHas:   true,
			wantType:  "knight",
		},
		{
			name:      "generic keyword name with raw-text type",
			keywordNm: "affinity",
			raw:       "Affinity for dragons (this spell costs {1} less to cast for each Dragon you control.)",
			wantHas:   true,
			wantType:  "dragon",
		},
		{
			name:      "no affinity keyword",
			keywordNm: "flying",
			raw:       "Flying",
			wantHas:   false,
			wantType:  "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			card := makeAffinityCard(tc.keywordNm, tc.raw)
			has, typeStr := AffinityForType(card)
			if has != tc.wantHas {
				t.Errorf("AffinityForType.has: got %v, want %v", has, tc.wantHas)
			}
			if typeStr != tc.wantType {
				t.Errorf("AffinityForType.typeStr: got %q, want %q", typeStr, tc.wantType)
			}
		})
	}
}

func TestAffinityForType_NilSafety(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AffinityForType on nil card panicked: %v", r)
		}
	}()
	has, typeStr := AffinityForType(nil)
	if has || typeStr != "" {
		t.Errorf("nil card should return (false, \"\"), got (%v, %q)", has, typeStr)
	}
}

func TestCountPermanentsByType_Subtype(t *testing.T) {
	gs := &GameState{
		Seats: []*Seat{
			{
				Battlefield: []*Permanent{
					{Card: &Card{Types: []string{"Legendary", "Creature", "Human", "Knight"}}},
					{Card: &Card{Types: []string{"Creature", "Human", "Soldier"}}},
					{Card: &Card{Types: []string{"Creature", "Elf", "Druid"}}},
					{Card: &Card{Types: []string{"Artifact"}}},
				},
			},
		},
	}

	// Humans on the battlefield: the Human Knight + the Human Soldier
	got := CountPermanentsByType(gs, 0, "human")
	if got != 2 {
		t.Errorf("count humans: got %d, want 2", got)
	}

	// Knights: just the Human Knight
	got = CountPermanentsByType(gs, 0, "knight")
	if got != 1 {
		t.Errorf("count knights: got %d, want 1", got)
	}

	// Artifacts: just the standalone Artifact
	got = CountPermanentsByType(gs, 0, "artifact")
	if got != 1 {
		t.Errorf("count artifacts: got %d, want 1", got)
	}
}

func TestCountPermanentsByType_CompoundType(t *testing.T) {
	// "artifact creature" — compound type for Urza, Chief Artificer.
	// A permanent qualifies only if it's BOTH artifact AND creature.
	gs := &GameState{
		Seats: []*Seat{
			{
				Battlefield: []*Permanent{
					{Card: &Card{Types: []string{"Artifact", "Creature", "Construct"}}}, // counts
					{Card: &Card{Types: []string{"Artifact"}}},                            // not creature
					{Card: &Card{Types: []string{"Creature", "Elf"}}},                     // not artifact
					{Card: &Card{Types: []string{"Artifact", "Creature", "Golem"}}},       // counts
				},
			},
		},
	}
	got := CountPermanentsByType(gs, 0, "artifact creature")
	if got != 2 {
		t.Errorf("count artifact creatures: got %d, want 2", got)
	}
}

func TestCountPermanentsByType_NilSafety(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CountPermanentsByType nil-safety panicked: %v", r)
		}
	}()
	if got := CountPermanentsByType(nil, 0, "creature"); got != 0 {
		t.Errorf("nil gs should return 0, got %d", got)
	}
	if got := CountPermanentsByType(&GameState{}, -1, "creature"); got != 0 {
		t.Errorf("invalid seatIdx should return 0, got %d", got)
	}
}
