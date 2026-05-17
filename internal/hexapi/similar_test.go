package hexapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeSig is a small constructor that takes the bits we care about for
// scoring tests so we don't have to write deck files for every case.
func makeSig(cmd, archetype, bracket string, colors []string, cards ...string) deckSig {
	s := deckSig{
		commanderCard: cmd,
		commanderName: cmd,
		archetype:     archetype,
		bracket:       bracket,
		colors:        colors,
		cards:         map[string]bool{},
	}
	for _, c := range cards {
		s.cards[strings.ToLower(c)] = true
	}
	return s
}

func TestComputeSimilarity_Jaccard(t *testing.T) {
	// Two decks of 4 cards each, 2 shared: union=6, jaccard=2/6≈0.333 → 20.0
	a := makeSig("Atraxa", "", "", nil, "Sol Ring", "Cyclonic Rift", "Counterspell", "Brainstorm")
	b := makeSig("Edgar Markov", "", "", nil, "Sol Ring", "Cyclonic Rift", "Vampiric Tutor", "Demonic Tutor")
	got := computeSimilarity(a, b)
	if got.shared != 2 {
		t.Errorf("shared = %d, want 2", got.shared)
	}
	// 2/6 = 33.33%, rounds to 33.
	if got.overlapPct != 33 {
		t.Errorf("overlapPct = %d, want 33", got.overlapPct)
	}
	// 60 * 0.3333 ≈ 20.0, no other bonuses.
	if got.score < 19.5 || got.score > 20.5 {
		t.Errorf("score = %.2f, want ~20", got.score)
	}
}

func TestComputeSimilarity_SameCommanderBonus(t *testing.T) {
	a := makeSig("Atraxa", "", "", nil, "Sol Ring")
	b := makeSig("Atraxa", "", "", nil, "Sol Ring")
	got := computeSimilarity(a, b)
	if !got.sameCommander {
		t.Fatal("expected sameCommander=true")
	}
	// jaccard=1.0 → 60, +25 commander = 85.
	if got.score < 84.5 || got.score > 85.5 {
		t.Errorf("score = %.2f, want ~85", got.score)
	}
}

func TestComputeSimilarity_ArchetypeBonus(t *testing.T) {
	a := makeSig("A", "combo", "", nil, "Sol Ring")
	b := makeSig("B", "combo", "", nil, "Mana Crypt")
	got := computeSimilarity(a, b)
	if !got.sameArchetype {
		t.Fatal("expected sameArchetype=true")
	}
	// No card overlap, but archetype match (+15) > floor (12), so kept.
	if got.dropped {
		t.Error("archetype-only match should not be dropped")
	}
	if got.score < 14.5 || got.score > 15.5 {
		t.Errorf("score = %.2f, want ~15", got.score)
	}
}

func TestComputeSimilarity_BracketAdjacency(t *testing.T) {
	cases := []struct {
		ab, bb string
		want   int // bracket bonus contribution
		dist   int
	}{
		{"3", "3", 10, 0},
		{"3", "4", 4, 1},
		{"2", "4", 0, 2},
		{"?", "3", 0, -1}, // unknown ⇒ no bonus, distance stays -1
	}
	for _, tc := range cases {
		a := makeSig("X", "", tc.ab, nil, "Sol Ring")
		b := makeSig("Y", "", tc.bb, nil, "Sol Ring")
		got := computeSimilarity(a, b)
		if got.bracketDistance != tc.dist {
			t.Errorf("bracket %s vs %s: distance = %d, want %d", tc.ab, tc.bb, got.bracketDistance, tc.dist)
		}
		// Card overlap is jaccard 1/1 = 60. Bracket bonus is extra.
		extra := got.score - 60.0
		if extra < float64(tc.want)-0.5 || extra > float64(tc.want)+0.5 {
			t.Errorf("bracket %s vs %s: bracket bonus = %.2f, want %d", tc.ab, tc.bb, extra, tc.want)
		}
	}
}

func TestComputeSimilarity_ColorIdentityBonus(t *testing.T) {
	a := makeSig("X", "", "", []string{"W", "U"}, "Sol Ring")
	b := makeSig("Y", "", "", []string{"U", "W"}, "Sol Ring") // order swapped
	got := computeSimilarity(a, b)
	if !got.sameColors {
		t.Fatal("expected sameColors=true (order should not matter)")
	}

	c := makeSig("Z", "", "", []string{"W"}, "Sol Ring")
	got = computeSimilarity(a, c)
	if got.sameColors {
		t.Error("mono-W and Azorius should not match")
	}
}

func TestComputeSimilarity_DropFloor(t *testing.T) {
	// Two unrelated 100-card decks, 1 shared card: jaccard ≈ 1/199,
	// score ≈ 0.3, no other bonuses → dropped.
	cardsA := make([]string, 100)
	cardsB := make([]string, 100)
	for i := 0; i < 100; i++ {
		cardsA[i] = "card_a_" + itoa(i)
		cardsB[i] = "card_b_" + itoa(i)
	}
	cardsA[0] = "Sol Ring"
	cardsB[0] = "Sol Ring"
	a := makeSig("Atraxa", "", "", nil, cardsA...)
	b := makeSig("Edgar", "", "", nil, cardsB...)
	got := computeSimilarity(a, b)
	if !got.dropped {
		t.Errorf("trivial overlap should be dropped, got score=%.2f", got.score)
	}
}

func TestComputeSimilarity_DropFloor_KeepsCommanderMatch(t *testing.T) {
	// Even with zero card overlap, sharing a commander keeps the result.
	a := makeSig("Atraxa", "", "", nil, "Sol Ring")
	b := makeSig("Atraxa", "", "", nil, "Brainstorm")
	got := computeSimilarity(a, b)
	if got.dropped {
		t.Error("same commander should never be dropped")
	}
	if !got.sameCommander {
		t.Error("sameCommander should be true")
	}
}

func TestComputeSimilarity_EmptyOther(t *testing.T) {
	a := makeSig("X", "", "", nil, "Sol Ring")
	b := makeSig("Y", "", "", nil) // no cards
	got := computeSimilarity(a, b)
	if got.shared != 0 || got.overlapPct != 0 {
		t.Errorf("empty other should yield 0/0, got %d/%d", got.shared, got.overlapPct)
	}
}

func TestParseBracket(t *testing.T) {
	cases := []struct {
		in   string
		want int
		ok   bool
	}{
		{"", 0, false},
		{"?", 0, false},
		{"0", 0, false},
		{"3", 3, true},
		{"5", 5, true},
		{"abc", 0, false},
	}
	for _, tc := range cases {
		n, ok := parseBracket(tc.in)
		if n != tc.want || ok != tc.ok {
			t.Errorf("parseBracket(%q) = (%d,%v), want (%d,%v)", tc.in, n, ok, tc.want, tc.ok)
		}
	}
}

func TestColorIdentityEqual(t *testing.T) {
	cases := []struct {
		a, b []string
		want bool
	}{
		{[]string{"W"}, []string{"W"}, true},
		{[]string{"W", "U"}, []string{"U", "W"}, true}, // order-independent
		{[]string{"w"}, []string{"W"}, true},           // case-insensitive
		{[]string{"W", "W"}, []string{"W"}, true},      // dedupe
		{[]string{"W"}, []string{"U"}, false},
		{[]string{"W", "U"}, []string{"W"}, false},
		{nil, nil, false}, // empty colorless vs empty: no signal, don't bonus
		{[]string{}, []string{}, false},
	}
	for _, tc := range cases {
		if got := colorIdentityEqual(tc.a, tc.b); got != tc.want {
			t.Errorf("colorIdentityEqual(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

// End-to-end: write a few decks + strategy.json sidecars, hit the handler,
// confirm ranking + new fields are correct.
func TestHandleSimilarDecks_Integration(t *testing.T) {
	dir := t.TempDir()

	// Owner 'alice', target deck: Atraxa, archetype=superfriends, bracket=3, WUBG.
	writeDeckFiles(t, dir, "alice", "atraxa_b3_target", "Atraxa, Praetors' Voice",
		"superfriends", 3, []string{"W", "U", "B", "G"},
		[]string{"Sol Ring", "Cyclonic Rift", "Doubling Season", "Teferi, Hero of Dominaria", "Narset, Parter of Veils"})

	// Twin deck: same commander, same archetype, same bracket, 4/5 cards shared.
	writeDeckFiles(t, dir, "alice", "atraxa_b3_twin", "Atraxa, Praetors' Voice",
		"superfriends", 3, []string{"W", "U", "B", "G"},
		[]string{"Sol Ring", "Cyclonic Rift", "Doubling Season", "Teferi, Hero of Dominaria", "Karn Liberated"})

	// Adjacent-bracket sibling: same commander/archetype but bracket=4.
	writeDeckFiles(t, dir, "bob", "atraxa_b4_sibling", "Atraxa, Praetors' Voice",
		"superfriends", 4, []string{"W", "U", "B", "G"},
		[]string{"Sol Ring", "Mana Crypt", "Teferi, Hero of Dominaria"})

	// Same-archetype, different commander, bracket=3, mono-U.
	writeDeckFiles(t, dir, "carol", "narset_b3_stranger", "Narset, Enlightened Master",
		"superfriends", 3, []string{"U"},
		[]string{"Doubling Season", "Counterspell"})

	// Unrelated deck — should drop below the floor and not appear.
	writeDeckFiles(t, dir, "dave", "krenko_b2_unrelated", "Krenko, Mob Boss",
		"aggro", 2, []string{"R"},
		[]string{"Goblin Guide", "Lightning Bolt"})

	h := &Handler{DecksDir: dir}
	req := httptest.NewRequest(http.MethodGet, "/api/decks/alice/atraxa_b3_target/similar?limit=10", nil)
	req.SetPathValue("owner", "alice")
	req.SetPathValue("id", "atraxa_b3_target")
	rr := httptest.NewRecorder()
	h.handleSimilarDecks(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}
	var got []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) < 3 {
		t.Fatalf("expected ≥3 results, got %d: %s", len(got), rr.Body.String())
	}

	// Twin (same owner) should be #1: same commander + archetype + bracket + colors +
	// 4 shared cards out of small union. Sibling (bob) should be #2.
	if got[0]["id"] != "atraxa_b3_twin" {
		t.Errorf("expected twin #1, got %v", got[0]["id"])
	}
	if got[1]["id"] != "atraxa_b4_sibling" {
		t.Errorf("expected sibling #2, got %v", got[1]["id"])
	}

	// Confirm the new fields are present and sensible on the top hit.
	top := got[0]
	mustBool(t, top, "same_commander", true)
	mustBool(t, top, "same_archetype", true)
	mustBool(t, top, "same_bracket", true)
	mustBool(t, top, "same_colors", true)
	if d, _ := top["bracket_distance"].(float64); d != 0 {
		t.Errorf("twin bracket_distance = %v, want 0", top["bracket_distance"])
	}
	if pct, _ := top["overlap_pct"].(float64); pct <= 0 || pct > 100 {
		t.Errorf("twin overlap_pct out of range: %v", top["overlap_pct"])
	}
	if sim, _ := top["similarity"].(float64); sim < 80 {
		t.Errorf("twin similarity should be high, got %v", top["similarity"])
	}

	// Sibling: bracket distance 1, bonus contributed.
	sib := got[1]
	if d, _ := sib["bracket_distance"].(float64); d != 1 {
		t.Errorf("sibling bracket_distance = %v, want 1", sib["bracket_distance"])
	}
	mustBool(t, sib, "same_bracket", false)

	// Unrelated krenko deck must not appear (floor drop).
	for _, row := range got {
		if row["id"] == "krenko_b2_unrelated" {
			t.Errorf("unrelated deck leaked into results: %v", row)
		}
	}
}

func TestHandleSimilarDecks_RejectsInvalidPath(t *testing.T) {
	h := &Handler{DecksDir: t.TempDir()}
	req := httptest.NewRequest(http.MethodGet, "/api/decks/../etc/similar", nil)
	req.SetPathValue("owner", "..")
	req.SetPathValue("id", "etc")
	rr := httptest.NewRecorder()
	h.handleSimilarDecks(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestHandleSimilarDecks_MissingDeck(t *testing.T) {
	h := &Handler{DecksDir: t.TempDir()}
	req := httptest.NewRequest(http.MethodGet, "/api/decks/alice/ghost/similar", nil)
	req.SetPathValue("owner", "alice")
	req.SetPathValue("id", "ghost")
	rr := httptest.NewRecorder()
	h.handleSimilarDecks(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

// writeDeckFiles writes a .txt deck and its Freya strategy.json sidecar.
func writeDeckFiles(t *testing.T, base, owner, id, commander, archetype string, bracket int, colors, cards []string) {
	t.Helper()
	ownerDir := filepath.Join(base, owner)
	freyaDir := filepath.Join(ownerDir, "freya")
	if err := os.MkdirAll(freyaDir, 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	b.WriteString("COMMANDER: " + commander + "\n")
	for _, c := range cards {
		b.WriteString("1 " + c + "\n")
	}
	if err := os.WriteFile(filepath.Join(ownerDir, id+".txt"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	strat := map[string]any{
		"archetype": archetype,
		"bracket":   bracket,
		"color_identity": map[string]any{
			"commander_colors": colors,
		},
	}
	data, _ := json.Marshal(strat)
	if err := os.WriteFile(filepath.Join(freyaDir, id+".strategy.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustBool(t *testing.T, row map[string]any, key string, want bool) {
	t.Helper()
	got, ok := row[key].(bool)
	if !ok {
		t.Errorf("%s not present or not bool: %v", key, row[key])
		return
	}
	if got != want {
		t.Errorf("%s = %v, want %v", key, got, want)
	}
}
