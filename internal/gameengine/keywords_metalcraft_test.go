package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Metalcraft tests — CR §702.97 (Scars of Mirrodin, 2010)
// ---------------------------------------------------------------------------

func newMetalcraftGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(2097))
	return NewGameState(2, rng, nil)
}

// putArtifactPerm mints a vanilla artifact permanent onto seat's
// battlefield. The `extraTypes` arg lets a test add subtypes ("creature",
// "land", "equipment", token-related tags) so we can verify §702.97a's
// "artifact-creature counts" claim.
func putArtifactPerm(gs *GameState, seat int, name string, extraTypes ...string) *Permanent {
	types := append([]string{"artifact"}, extraTypes...)
	card := &Card{
		Name:  name,
		Owner: seat,
		Types: types,
		AST:   &gameast.CardAST{Name: name, Abilities: []gameast.Ability{}},
	}
	p := &Permanent{Card: card, Controller: seat, Owner: seat}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, p)
	return p
}

func putNonArtifactCreaturePerm(gs *GameState, seat int, name string) *Permanent {
	card := &Card{
		Name:          name,
		Owner:         seat,
		Types:         []string{"creature"},
		BasePower:     2,
		BaseToughness: 2,
		AST:           &gameast.CardAST{Name: name, Abilities: []gameast.Ability{}},
	}
	p := &Permanent{Card: card, Controller: seat, Owner: seat}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, p)
	return p
}

func newCardWithStaticTextMC(name, text string) *Card {
	return &Card{
		Name:  name,
		Types: []string{"creature"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: text},
			},
		},
	}
}

func newPlainNonMetalCard(name string) *Card {
	return &Card{
		Name:  name,
		Types: []string{"creature"},
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{},
		},
	}
}

// ---------------------------------------------------------------------------
// (e) HasMetalcraft detects "metalcraft" in oracle text
// ---------------------------------------------------------------------------

func TestHasMetalcraft_OracleText(t *testing.T) {
	card := newCardWithStaticTextMC("Auriok Sunchaser",
		"Metalcraft — Auriok Sunchaser gets +2/+2 and has flying as long as you control three or more artifacts.")
	if !HasMetalcraft(card) {
		t.Fatal("HasMetalcraft should detect metalcraft in oracle text")
	}
}

func TestHasMetalcraft_CaseInsensitive(t *testing.T) {
	card := newCardWithStaticTextMC("Mox", "METALCRAFT enables this ability.")
	if !HasMetalcraft(card) {
		t.Fatal("HasMetalcraft should be case-insensitive")
	}
}

func TestHasMetalcraft_KeywordAST(t *testing.T) {
	card := &Card{
		Name: "Direct Keyword",
		AST: &gameast.CardAST{
			Name: "Direct Keyword",
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "metalcraft"},
			},
		},
	}
	if !HasMetalcraft(card) {
		t.Fatal("HasMetalcraft should detect a direct Keyword AST node")
	}
}

func TestHasMetalcraft_Negative(t *testing.T) {
	if HasMetalcraft(newPlainNonMetalCard("Plain Card")) {
		t.Fatal("HasMetalcraft should be false on a vanilla card")
	}
	// Mentions "artifact" but not "metalcraft" — must not match.
	c := newCardWithStaticTextMC("Artifact Lover",
		"As long as you control three or more artifacts, this creature gets +2/+2.")
	if HasMetalcraft(c) {
		t.Fatal("HasMetalcraft must require the literal word, not just an artifact-threshold rider")
	}
}

func TestHasMetalcraft_Nil(t *testing.T) {
	if HasMetalcraft(nil) {
		t.Fatal("HasMetalcraft(nil) must be false")
	}
	if HasMetalcraft(&Card{Name: "no-AST"}) {
		t.Fatal("HasMetalcraft on an AST-less card must be false")
	}
}

// ---------------------------------------------------------------------------
// (a) MetalcraftActive false at 0–2 artifacts
// ---------------------------------------------------------------------------

func TestMetalcraftActive_ZeroArtifacts(t *testing.T) {
	gs := newMetalcraftGame(t)
	if MetalcraftActive(gs, 0) {
		t.Fatal("MetalcraftActive should be false with 0 artifacts")
	}
}

func TestMetalcraftActive_OneArtifact(t *testing.T) {
	gs := newMetalcraftGame(t)
	putArtifactPerm(gs, 0, "Sol Ring")
	if MetalcraftActive(gs, 0) {
		t.Fatal("MetalcraftActive should be false with only 1 artifact")
	}
}

func TestMetalcraftActive_TwoArtifacts(t *testing.T) {
	gs := newMetalcraftGame(t)
	putArtifactPerm(gs, 0, "Sol Ring")
	putArtifactPerm(gs, 0, "Mox Opal")
	if MetalcraftActive(gs, 0) {
		t.Fatal("MetalcraftActive should be false with only 2 artifacts (threshold is 3)")
	}
}

// ---------------------------------------------------------------------------
// (b) MetalcraftActive true at 3+ artifacts
// ---------------------------------------------------------------------------

func TestMetalcraftActive_ThreeArtifactsThreshold(t *testing.T) {
	gs := newMetalcraftGame(t)
	putArtifactPerm(gs, 0, "Sol Ring")
	putArtifactPerm(gs, 0, "Mox Opal")
	putArtifactPerm(gs, 0, "Mana Vault")
	if !MetalcraftActive(gs, 0) {
		t.Fatal("MetalcraftActive should be true exactly at 3 artifacts")
	}
}

func TestMetalcraftActive_FourArtifacts(t *testing.T) {
	gs := newMetalcraftGame(t)
	for _, name := range []string{"Sol Ring", "Mox Opal", "Mana Vault", "Grim Monolith"} {
		putArtifactPerm(gs, 0, name)
	}
	if !MetalcraftActive(gs, 0) {
		t.Fatal("MetalcraftActive should be true at 4 artifacts")
	}
}

// ---------------------------------------------------------------------------
// (c) Artifact-creature counts toward the count (per CR §702.97a)
// ---------------------------------------------------------------------------

func TestMetalcraftActive_ArtifactCreatureCounts(t *testing.T) {
	gs := newMetalcraftGame(t)
	putArtifactPerm(gs, 0, "Sol Ring")
	putArtifactPerm(gs, 0, "Bonesplitter", "equipment")
	// Phyrexian Revoker — artifact creature.
	putArtifactPerm(gs, 0, "Phyrexian Revoker", "creature")

	if !MetalcraftActive(gs, 0) {
		t.Fatal("MetalcraftActive should be true: artifact creatures count toward the 3-artifact threshold")
	}
	if got := ArtifactCount(gs, 0); got != 3 {
		t.Errorf("ArtifactCount = %d, want 3", got)
	}
}

func TestMetalcraftActive_ArtifactLandCounts(t *testing.T) {
	gs := newMetalcraftGame(t)
	putArtifactPerm(gs, 0, "Inkmoth Nexus", "land")
	putArtifactPerm(gs, 0, "Tree of Tales", "land")
	putArtifactPerm(gs, 0, "Seat of the Synod", "land")
	if !MetalcraftActive(gs, 0) {
		t.Fatal("MetalcraftActive should be true: artifact lands count toward the threshold")
	}
}

func TestMetalcraftActive_TokenArtifactsCount(t *testing.T) {
	gs := newMetalcraftGame(t)
	putArtifactPerm(gs, 0, "Treasure", "treasure", "token")
	putArtifactPerm(gs, 0, "Clue", "clue", "token")
	putArtifactPerm(gs, 0, "Food", "food", "token")
	if !MetalcraftActive(gs, 0) {
		t.Fatal("MetalcraftActive should count Treasure / Clue / Food token artifacts")
	}
}

func TestMetalcraftActive_NonArtifactCreaturesDoNotCount(t *testing.T) {
	gs := newMetalcraftGame(t)
	// 3 vanilla creatures, no artifacts.
	for _, n := range []string{"Bear", "Wolf", "Lion"} {
		putNonArtifactCreaturePerm(gs, 0, n)
	}
	if MetalcraftActive(gs, 0) {
		t.Fatal("MetalcraftActive must be false when seat controls only non-artifact creatures")
	}
}

// ---------------------------------------------------------------------------
// (d) Opponent's artifacts don't count
// ---------------------------------------------------------------------------

func TestMetalcraftActive_OpponentArtifactsExcluded(t *testing.T) {
	gs := newMetalcraftGame(t)
	// Seat 1 has 5 artifacts.
	for _, n := range []string{"A1", "A2", "A3", "A4", "A5"} {
		putArtifactPerm(gs, 1, n)
	}
	// Seat 0 has zero.
	if MetalcraftActive(gs, 0) {
		t.Fatal("MetalcraftActive(0) should be false when only opponent (seat 1) has artifacts")
	}
}

func TestMetalcraftActive_MixedSeatsScopedCorrectly(t *testing.T) {
	gs := newMetalcraftGame(t)
	// Seat 0: 2 artifacts (below threshold).
	putArtifactPerm(gs, 0, "Sol Ring")
	putArtifactPerm(gs, 0, "Mox Opal")
	// Seat 1: 3 artifacts (above threshold for seat 1 only).
	for _, n := range []string{"X1", "X2", "X3"} {
		putArtifactPerm(gs, 1, n)
	}
	if MetalcraftActive(gs, 0) {
		t.Error("MetalcraftActive(0) should be false (seat 0 has only 2 artifacts)")
	}
	if !MetalcraftActive(gs, 1) {
		t.Error("MetalcraftActive(1) should be true (seat 1 has 3 artifacts)")
	}
}

// ---------------------------------------------------------------------------
// Phased-out artifacts don't count
// ---------------------------------------------------------------------------

func TestMetalcraftActive_PhasedOutDoesNotCount(t *testing.T) {
	gs := newMetalcraftGame(t)
	a := putArtifactPerm(gs, 0, "Sol Ring")
	b := putArtifactPerm(gs, 0, "Mox Opal")
	c := putArtifactPerm(gs, 0, "Mana Vault")
	c.PhasedOut = true

	if MetalcraftActive(gs, 0) {
		t.Errorf("MetalcraftActive should be false when one of the artifacts is phased out (got %d, want 2)",
			ArtifactCount(gs, 0))
	}
	_ = a
	_ = b
}

// ---------------------------------------------------------------------------
// Nil-safety
// ---------------------------------------------------------------------------

func TestMetalcraftActive_NilSafe(t *testing.T) {
	if MetalcraftActive(nil, 0) {
		t.Fatal("nil gs must be false")
	}
	gs := newMetalcraftGame(t)
	if MetalcraftActive(gs, 99) {
		t.Fatal("invalid seat must be false")
	}
	if MetalcraftActive(gs, -1) {
		t.Fatal("negative seat must be false")
	}
}

func TestArtifactCount_NilSafe(t *testing.T) {
	if got := ArtifactCount(nil, 0); got != 0 {
		t.Errorf("nil gs count = %d, want 0", got)
	}
	gs := newMetalcraftGame(t)
	if got := ArtifactCount(gs, 99); got != 0 {
		t.Errorf("invalid seat count = %d, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// Legacy CheckMetalcraft alias still works
// ---------------------------------------------------------------------------

func TestCheckMetalcraft_AliasBacksMetalcraftActive(t *testing.T) {
	gs := newMetalcraftGame(t)
	if CheckMetalcraft(gs, 0) {
		t.Fatal("legacy CheckMetalcraft should be false at 0 artifacts")
	}
	for _, n := range []string{"A1", "A2", "A3"} {
		putArtifactPerm(gs, 0, n)
	}
	if !CheckMetalcraft(gs, 0) {
		t.Fatal("legacy CheckMetalcraft should be true at 3 artifacts (alias to MetalcraftActive)")
	}
	// Sanity: alias and canonical agree across all valid seats.
	for seat := 0; seat < len(gs.Seats); seat++ {
		if MetalcraftActive(gs, seat) != CheckMetalcraft(gs, seat) {
			t.Errorf("alias mismatch at seat %d", seat)
		}
	}
}
