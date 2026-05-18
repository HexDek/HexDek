package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Imprint tests — CR §702.40 (Mirrodin, 2003)
// ---------------------------------------------------------------------------

func newImprintGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(2040))
	return NewGameState(2, rng, nil)
}

// newImprintArtifactCard builds an artifact card whose oracle text
// references "imprint" — used to test HasImprint detection.
func newImprintArtifactCard(name, oracle string) *Card {
	return &Card{
		Name:  name,
		Types: []string{"artifact"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: oracle},
			},
		},
	}
}

func newPlainCardI2(name string) *Card {
	return &Card{
		Name:  name,
		Types: []string{"creature"},
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{},
		},
	}
}

// putImprintHost puts an artifact-shaped permanent on seat's
// battlefield with the given Timestamp (so target.Meta["imprinted_by"]
// can be asserted).
func putImprintHost(gs *GameState, seat int, name string, ts int) *Permanent {
	card := &Card{
		Name:  name,
		Types: []string{"artifact"},
		AST:   &gameast.CardAST{Name: name, Abilities: []gameast.Ability{}},
	}
	p := &Permanent{
		Card:       card,
		Controller: seat,
		Owner:      seat,
		Timestamp:  ts,
		Flags:      map[string]int{},
	}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, p)
	return p
}

// ---------------------------------------------------------------------------
// (e) HasImprint detector
// ---------------------------------------------------------------------------

func TestHasImprint_OracleText(t *testing.T) {
	c := newImprintArtifactCard("Isochron Scepter",
		"Imprint — When Isochron Scepter enters the battlefield, you may exile an instant card with mana value 2 or less from your hand.")
	if !HasImprint(c) {
		t.Fatal("HasImprint should detect oracle text containing \"imprint\"")
	}
}

func TestHasImprint_CaseInsensitive(t *testing.T) {
	c := newImprintArtifactCard("Caps Lock", "IMPRINT — This artifact does imprint stuff.")
	if !HasImprint(c) {
		t.Fatal("HasImprint should be case-insensitive")
	}
}

func TestHasImprint_KeywordAST(t *testing.T) {
	c := &Card{
		Name: "Chrome Mox",
		AST: &gameast.CardAST{
			Name:      "Chrome Mox",
			Abilities: []gameast.Ability{&gameast.Keyword{Name: "imprint"}},
		},
	}
	if !HasImprint(c) {
		t.Fatal("HasImprint should detect a direct Keyword AST node")
	}
}

func TestHasImprint_Negative(t *testing.T) {
	if HasImprint(newPlainCardI2("Plain Card")) {
		t.Fatal("HasImprint must be false on a vanilla card")
	}
	c := newImprintArtifactCard("Distractor",
		"This card has nothing to do with the mechanic, even if it says ETB.")
	if HasImprint(c) {
		t.Fatal("HasImprint must be false when oracle text doesn't contain the word")
	}
}

func TestHasImprint_Nil(t *testing.T) {
	if HasImprint(nil) {
		t.Fatal("HasImprint(nil) must be false")
	}
	if HasImprint(&Card{Name: "no-AST"}) {
		t.Fatal("HasImprint on AST-less card must be false")
	}
}

// ---------------------------------------------------------------------------
// (a) ApplyImprint exiles target + sets metadata
// ---------------------------------------------------------------------------

func TestApplyImprint_ExilesTargetFromHand(t *testing.T) {
	gs := newImprintGame(t)
	host := putImprintHost(gs, 0, "Isochron Scepter", 100)
	instant := &Card{
		Name:  "Lightning Bolt",
		Owner: 0,
		Types: []string{"instant"},
		CMC:   1,
		AST:   &gameast.CardAST{Name: "Lightning Bolt", Abilities: []gameast.Ability{}},
	}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, instant)

	got, err := ApplyImprint(gs, host, instant)
	if err != nil {
		t.Fatalf("ApplyImprint error: %v", err)
	}
	if got != instant {
		t.Errorf("ApplyImprint returned %v, want %v", got, instant)
	}

	// Card should be in exile, not in hand.
	stillInHand := false
	for _, c := range gs.Seats[0].Hand {
		if c == instant {
			stillInHand = true
		}
	}
	if stillInHand {
		t.Error("imprinted card should no longer be in hand")
	}
	inExile := false
	for _, c := range gs.Seats[0].Exile {
		if c == instant {
			inExile = true
		}
	}
	if !inExile {
		t.Error("imprinted card should be in exile")
	}

	// Metadata on both sides.
	if host.Flags["imprinted"] != 1 {
		t.Error("host.Flags[imprinted] should be 1")
	}
	if len(host.LinkedExile) != 1 || host.LinkedExile[0] != instant {
		t.Errorf("host.LinkedExile = %v, want [instant]", host.LinkedExile)
	}
	if instant.Meta == nil {
		t.Fatal("instant.Meta should be populated")
	}
	got2, _ := instant.Meta["imprinted_by"].(int)
	if got2 != 100 {
		t.Errorf("instant.Meta[imprinted_by] = %d, want 100 (host timestamp)", got2)
	}
	// imprint event logged.
	foundEvent := false
	for _, e := range gs.EventLog {
		if e.Kind == "imprint" && e.Source == "Isochron Scepter" {
			foundEvent = true
		}
	}
	if !foundEvent {
		t.Error("expected an imprint event in the log")
	}
}

func TestApplyImprint_ExilesFromLibrary(t *testing.T) {
	gs := newImprintGame(t)
	host := putImprintHost(gs, 0, "Mirror of Fate", 200)
	card := &Card{
		Name:  "Top of Library",
		Owner: 0,
		Types: []string{"sorcery"},
		AST:   &gameast.CardAST{Name: "Top of Library", Abilities: []gameast.Ability{}},
	}
	gs.Seats[0].Library = append(gs.Seats[0].Library, card)

	if _, err := ApplyImprint(gs, host, card); err != nil {
		t.Fatalf("ApplyImprint error: %v", err)
	}
	// Library should be empty now (1-card lib drained).
	for _, c := range gs.Seats[0].Library {
		if c == card {
			t.Error("imprinted card should no longer be in library")
		}
	}
	if len(host.LinkedExile) != 1 {
		t.Errorf("host.LinkedExile len = %d, want 1", len(host.LinkedExile))
	}
}

func TestApplyImprint_RejectsTargetNotFound(t *testing.T) {
	gs := newImprintGame(t)
	host := putImprintHost(gs, 0, "Host", 1)
	orphan := &Card{
		Name: "Orphan",
		AST:  &gameast.CardAST{Name: "Orphan", Abilities: []gameast.Ability{}},
	}
	// orphan is not in any of seat 0's zones.
	if _, err := ApplyImprint(gs, host, orphan); err == nil {
		t.Fatal("ApplyImprint should reject a target that isn't in the controller's zones")
	}
	if host.Flags != nil && host.Flags["imprinted"] == 1 {
		t.Error("host should not be flagged imprinted when ApplyImprint failed")
	}
}

func TestApplyImprint_NilSafe(t *testing.T) {
	if _, err := ApplyImprint(nil, nil, nil); err == nil {
		t.Error("expected error from nil inputs")
	}
	gs := newImprintGame(t)
	if _, err := ApplyImprint(gs, nil, &Card{}); err == nil {
		t.Error("expected error from nil perm")
	}
	if _, err := ApplyImprint(gs, &Permanent{Controller: 99}, &Card{}); err == nil {
		t.Error("expected error from invalid controller")
	}
}

// ---------------------------------------------------------------------------
// (b) GetImprintedCard returns the card
// ---------------------------------------------------------------------------

func TestGetImprintedCard_ReturnsImprintedCard(t *testing.T) {
	gs := newImprintGame(t)
	host := putImprintHost(gs, 0, "Scepter", 50)
	bolt := &Card{
		Name:  "Bolt",
		Owner: 0,
		Types: []string{"instant"},
		AST:   &gameast.CardAST{Name: "Bolt", Abilities: []gameast.Ability{}},
	}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, bolt)
	if _, err := ApplyImprint(gs, host, bolt); err != nil {
		t.Fatalf("ApplyImprint error: %v", err)
	}
	if got := GetImprintedCard(host); got != bolt {
		t.Errorf("GetImprintedCard = %v, want %v", got, bolt)
	}
}

func TestGetImprintedCard_NilWhenNotImprinted(t *testing.T) {
	host := &Permanent{Card: &Card{Name: "Vanilla"}}
	if got := GetImprintedCard(host); got != nil {
		t.Errorf("GetImprintedCard on non-imprinted host = %v, want nil", got)
	}
}

func TestGetImprintedCard_NilSafe(t *testing.T) {
	if GetImprintedCard(nil) != nil {
		t.Error("GetImprintedCard(nil) must be nil")
	}
}

func TestGetImprintedCards_MultipleImprints(t *testing.T) {
	gs := newImprintGame(t)
	host := putImprintHost(gs, 0, "MultiHost", 10)
	a := &Card{Name: "A", Owner: 0, AST: &gameast.CardAST{Name: "A"}}
	b := &Card{Name: "B", Owner: 0, AST: &gameast.CardAST{Name: "B"}}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, a, b)
	if _, err := ApplyImprint(gs, host, a); err != nil {
		t.Fatalf("first imprint error: %v", err)
	}
	if _, err := ApplyImprint(gs, host, b); err != nil {
		t.Fatalf("second imprint error: %v", err)
	}
	cards := GetImprintedCards(host)
	if len(cards) != 2 {
		t.Fatalf("GetImprintedCards len = %d, want 2", len(cards))
	}
	if cards[0] != a || cards[1] != b {
		t.Errorf("imprint order = [%v %v], want [a b]", cards[0], cards[1])
	}
	// GetImprintedCard returns first one (oldest).
	if got := GetImprintedCard(host); got != a {
		t.Errorf("GetImprintedCard returned %v, want oldest (a)", got)
	}
}

// ---------------------------------------------------------------------------
// (c) Isochron Scepter-style flow — read imprinted instant to copy it
// ---------------------------------------------------------------------------

// simulateIsochronCopy mirrors the printed Isochron Scepter ability:
// "{2}, {T}: You may copy the exiled card. If you do, you may cast the
// copy without paying its mana cost." We don't push to the real stack
// here — we just verify the surface lets a per_card handler READ the
// imprinted card and produce a copy.
func simulateIsochronCopy(host *Permanent) *Card {
	imprinted := GetImprintedCard(host)
	if imprinted == nil {
		return nil
	}
	copy := imprinted.DeepCopy()
	copy.IsCopy = true
	return copy
}

func TestImprint_IsochronScepterStyle_CopyImprintedInstant(t *testing.T) {
	gs := newImprintGame(t)
	scepter := putImprintHost(gs, 0, "Isochron Scepter", 1000)
	bolt := &Card{
		Name:  "Lightning Bolt",
		Owner: 0,
		Types: []string{"instant"},
		CMC:   1,
		AST:   &gameast.CardAST{Name: "Lightning Bolt", Abilities: []gameast.Ability{}},
	}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, bolt)
	if _, err := ApplyImprint(gs, scepter, bolt); err != nil {
		t.Fatalf("ApplyImprint error: %v", err)
	}

	// Activate Scepter — produces a copy of the imprinted Bolt.
	copy1 := simulateIsochronCopy(scepter)
	if copy1 == nil {
		t.Fatal("Scepter activation: should produce a copy of imprinted Bolt")
	}
	if !copy1.IsCopy {
		t.Error("copy should be flagged IsCopy")
	}
	if copy1.Name != "Lightning Bolt" {
		t.Errorf("copy.Name = %q, want \"Lightning Bolt\"", copy1.Name)
	}
	// Crucially, the ORIGINAL bolt is still imprinted — re-activating
	// produces another copy.
	copy2 := simulateIsochronCopy(scepter)
	if copy2 == nil || copy2 == copy1 {
		t.Error("re-activating Scepter should produce a NEW copy each time, not return the same pointer")
	}
	// And the original card is still in exile, still imprinted.
	if GetImprintedCard(scepter) != bolt {
		t.Error("the imprinted-card association should persist across activations")
	}
}

// ---------------------------------------------------------------------------
// (d) ReleaseImprint cleans both ends
// ---------------------------------------------------------------------------

func TestReleaseImprint_ClearsBothSides(t *testing.T) {
	gs := newImprintGame(t)
	host := putImprintHost(gs, 0, "Host", 42)
	card := &Card{
		Name:  "Target",
		Owner: 0,
		AST:   &gameast.CardAST{Name: "Target", Abilities: []gameast.Ability{}},
	}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)
	if _, err := ApplyImprint(gs, host, card); err != nil {
		t.Fatalf("ApplyImprint error: %v", err)
	}

	// Sanity: imprint state is live.
	if GetImprintedCard(host) != card {
		t.Fatal("setup failed: imprint not established")
	}

	// Release.
	ReleaseImprint(gs, host)

	// Host side cleaned.
	if host.Flags["imprinted"] != 0 {
		t.Errorf("host.Flags[imprinted] = %d, want 0 after release", host.Flags["imprinted"])
	}
	if len(host.LinkedExile) != 0 {
		t.Errorf("host.LinkedExile len = %d, want 0 after release", len(host.LinkedExile))
	}
	// Card side cleaned.
	if card.Meta != nil {
		if _, exists := card.Meta["imprinted_by"]; exists {
			t.Error("card.Meta[imprinted_by] should be removed after release")
		}
	}
	// GetImprintedCard returns nil.
	if GetImprintedCard(host) != nil {
		t.Error("GetImprintedCard should be nil after release")
	}
	// The card itself stays in exile (release dissolves the link, not the exile).
	stillInExile := false
	for _, c := range gs.Seats[0].Exile {
		if c == card {
			stillInExile = true
		}
	}
	if !stillInExile {
		t.Error("imprinted card should remain in exile after release (release dissolves the link only)")
	}
	// Released event logged.
	foundEvent := false
	for _, e := range gs.EventLog {
		if e.Kind == "imprint_released" && e.Source == "Host" {
			foundEvent = true
		}
	}
	if !foundEvent {
		t.Error("expected imprint_released event")
	}
}

func TestReleaseImprint_NoOpOnNonImprintedHost(t *testing.T) {
	gs := newImprintGame(t)
	host := putImprintHost(gs, 0, "Host", 1)
	// No imprint ever applied; release is a safe no-op.
	ReleaseImprint(gs, host)
	for _, e := range gs.EventLog {
		if e.Kind == "imprint_released" {
			t.Fatal("imprint_released event should NOT fire when no imprint state is active")
		}
	}
}

func TestReleaseImprint_NilSafe(t *testing.T) {
	ReleaseImprint(nil, nil)
	ReleaseImprint(nil, &Permanent{})
	gs := newImprintGame(t)
	ReleaseImprint(gs, nil)
}

func TestReleaseImprint_MultipleCardsAllCleared(t *testing.T) {
	gs := newImprintGame(t)
	host := putImprintHost(gs, 0, "Multi", 7)
	a := &Card{Name: "A", Owner: 0, AST: &gameast.CardAST{Name: "A"}}
	b := &Card{Name: "B", Owner: 0, AST: &gameast.CardAST{Name: "B"}}
	c := &Card{Name: "C", Owner: 0, AST: &gameast.CardAST{Name: "C"}}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, a, b, c)
	for _, card := range []*Card{a, b, c} {
		if _, err := ApplyImprint(gs, host, card); err != nil {
			t.Fatalf("ApplyImprint(%s) error: %v", card.Name, err)
		}
	}
	ReleaseImprint(gs, host)
	for _, card := range []*Card{a, b, c} {
		if card.Meta != nil {
			if _, ok := card.Meta["imprinted_by"]; ok {
				t.Errorf("card %s still has imprinted_by meta after release", card.Name)
			}
		}
	}
	if len(host.LinkedExile) != 0 {
		t.Errorf("LinkedExile should be empty after release; got %d entries", len(host.LinkedExile))
	}
}
