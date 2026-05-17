package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Behold registry tests — CR §701.4
// ---------------------------------------------------------------------------

func newBeholdGame(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(2, rand.New(rand.NewSource(48)), nil)
}

// dragonCard is a hand-card with the Dragon creature subtype baked into
// its Types slice. cardHasType (used by CardHasBeholdQuality) matches
// case-insensitively against the slice.
func dragonCard(name string, owner int) *Card {
	return &Card{
		Name:     name,
		Owner:    owner,
		Types:    []string{"creature", "dragon"},
		TypeLine: "Creature — Dragon",
		CMC:      5,
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{},
		},
	}
}

func squirrelPerm(gs *GameState, owner int, name string) *Permanent {
	card := &Card{
		Name:     name,
		Owner:    owner,
		Types:    []string{"creature", "squirrel"},
		TypeLine: "Creature — Squirrel",
		AST:      &gameast.CardAST{Name: name},
	}
	perm := &Permanent{
		Card:       card,
		Controller: owner,
		Owner:      owner,
		Flags:      map[string]int{},
		Timestamp:  gs.NextTimestamp(),
	}
	gs.Seats[owner].Battlefield = append(gs.Seats[owner].Battlefield, perm)
	return perm
}

// ---------------------------------------------------------------------------
// (a) Reveal-from-hand path records quality
// ---------------------------------------------------------------------------

func TestBehold_RevealFromHandRecordsQuality(t *testing.T) {
	gs := newBeholdGame(t)
	d := dragonCard("Shivan Dragon", 0)
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, d)

	if HasBeheld(gs, 0, "dragon") {
		t.Fatal("registry must start empty")
	}

	ok := BeholdRevealFromHand(gs, 0, "Dragon", "Mabel, Heir to Cragflame", d)
	if !ok {
		t.Fatal("BeholdRevealFromHand should succeed for a Dragon in hand")
	}
	if !HasBeheld(gs, 0, "dragon") {
		t.Fatal("HasBeheld(dragon) should be true after the reveal")
	}
	// Case-insensitive: query with the printed casing works.
	if !HasBeheld(gs, 0, "Dragon") {
		t.Fatal("HasBeheld should be case-insensitive")
	}
	if BeheldCount(gs, 0, "dragon") != 1 {
		t.Fatalf("BeheldCount = %d, want 1", BeheldCount(gs, 0, "dragon"))
	}
	// Card stays in hand — Behold does not move it.
	if len(gs.Seats[0].Hand) != 1 || gs.Seats[0].Hand[0] != d {
		t.Fatal("revealed card must remain in hand")
	}
}

func TestBehold_RevealRejectsWrongQuality(t *testing.T) {
	gs := newBeholdGame(t)
	d := dragonCard("Shivan Dragon", 0)
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, d)

	if BeholdRevealFromHand(gs, 0, "Squirrel", "Some Card", d) {
		t.Fatal("revealing a Dragon must not satisfy 'behold a Squirrel'")
	}
	if HasBeheld(gs, 0, "squirrel") {
		t.Fatal("no behold should be recorded on a quality mismatch")
	}
}

func TestBehold_RevealRejectsCardNotInHand(t *testing.T) {
	gs := newBeholdGame(t)
	d := dragonCard("Shivan Dragon", 0)
	// Not in hand — sitting in the graveyard for example.
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, d)

	if BeholdRevealFromHand(gs, 0, "Dragon", "Source", d) {
		t.Fatal("reveal path requires the card to be in hand")
	}
	if HasBeheld(gs, 0, "dragon") {
		t.Fatal("no behold should be recorded when the reveal is illegal")
	}
}

// ---------------------------------------------------------------------------
// (b) Choose-from-battlefield path records quality
// ---------------------------------------------------------------------------

func TestBehold_ChoosePermanentRecordsQuality(t *testing.T) {
	gs := newBeholdGame(t)
	s := squirrelPerm(gs, 0, "Chatterfang Pup")

	ok := BeholdChoosePermanent(gs, 0, "Squirrel", "Chatterfang, Squirrel General", s)
	if !ok {
		t.Fatal("BeholdChoosePermanent should succeed for a Squirrel you control")
	}
	if !HasBeheld(gs, 0, "squirrel") {
		t.Fatal("HasBeheld(squirrel) should be true after the choose")
	}
	if BeheldCount(gs, 0, "squirrel") != 1 {
		t.Fatalf("count = %d, want 1", BeheldCount(gs, 0, "squirrel"))
	}
	// Permanent stays on the battlefield untouched.
	if len(gs.Seats[0].Battlefield) != 1 || gs.Seats[0].Battlefield[0] != s {
		t.Fatal("chosen permanent must remain on the battlefield")
	}
}

func TestBehold_ChooseRejectsOpponentsPermanent(t *testing.T) {
	gs := newBeholdGame(t)
	// Seat 1's Squirrel — seat 0 may not choose it for their behold.
	s := squirrelPerm(gs, 1, "Opponent's Squirrel")

	if BeholdChoosePermanent(gs, 0, "Squirrel", "Source", s) {
		t.Fatal("choose path must reject a permanent the player does not control")
	}
	if HasBeheld(gs, 0, "squirrel") {
		t.Fatal("no behold should be recorded on an illegal choose")
	}
}

func TestBehold_ChooseRejectsWrongQuality(t *testing.T) {
	gs := newBeholdGame(t)
	s := squirrelPerm(gs, 0, "Chatterfang Pup")

	if BeholdChoosePermanent(gs, 0, "Dragon", "Source", s) {
		t.Fatal("a Squirrel must not satisfy 'behold a Dragon'")
	}
}

// ---------------------------------------------------------------------------
// (c) HasBeheld queries
// ---------------------------------------------------------------------------

func TestBehold_HasBeheld_PerSeatIsolation(t *testing.T) {
	gs := newBeholdGame(t)
	Behold(gs, 0, "Dragon", "src")
	if !HasBeheld(gs, 0, "Dragon") {
		t.Fatal("seat 0 should register the behold")
	}
	// Seat 1's registry is independent.
	if HasBeheld(gs, 1, "Dragon") {
		t.Fatal("seat 1 must not see seat 0's behold")
	}
	if HasBeheld(gs, 0, "Squirrel") {
		t.Fatal("seat 0 has not beheld a Squirrel")
	}
}

func TestBehold_BeheldCount_IncrementsOnRepeat(t *testing.T) {
	gs := newBeholdGame(t)
	Behold(gs, 0, "Dragon", "src1")
	Behold(gs, 0, "Dragon", "src2")
	Behold(gs, 0, "Dragon", "src3")
	if got := BeheldCount(gs, 0, "Dragon"); got != 3 {
		t.Fatalf("BeheldCount after 3 beholds = %d, want 3", got)
	}
	// Returned count from Behold is the new value after recording.
	if n := Behold(gs, 0, "Dragon", "src4"); n != 4 {
		t.Fatalf("Behold should return 4 on the 4th invocation, got %d", n)
	}
}

func TestBehold_BeheldCount_QuiescentDefault(t *testing.T) {
	gs := newBeholdGame(t)
	if BeheldCount(gs, 0, "Dragon") != 0 {
		t.Fatal("fresh registry should report 0 for any quality")
	}
	if HasBeheld(gs, 0, "Dragon") {
		t.Fatal("fresh registry should report nothing beheld")
	}
}

// ---------------------------------------------------------------------------
// (d) Per-card "when you behold X" trigger fires
// ---------------------------------------------------------------------------

func TestBehold_FiresBeheldTriggerHook(t *testing.T) {
	gs := newBeholdGame(t)

	prev := TriggerHook
	defer func() { TriggerHook = prev }()

	type fire struct {
		seat    int
		quality string
		source  string
		count   int
	}
	var observed []fire
	TriggerHook = func(gs *GameState, ev string, ctx map[string]interface{}) {
		if ev != "beheld" {
			return
		}
		seat, _ := ctx["seat"].(int)
		q, _ := ctx["quality"].(string)
		src, _ := ctx["source"].(string)
		count, _ := ctx["behold_count"].(int)
		observed = append(observed, fire{seat, q, src, count})
	}

	Behold(gs, 0, "Dragon", "Mabel, Heir to Cragflame")
	Behold(gs, 0, "Dragon", "Mabel, Heir to Cragflame")
	Behold(gs, 1, "Squirrel", "Chatterfang")

	if len(observed) != 3 {
		t.Fatalf("expected 3 beheld trigger fires, got %d", len(observed))
	}
	// First fire: seat 0, Dragon, count 1.
	if observed[0] != (fire{0, "dragon", "Mabel, Heir to Cragflame", 1}) {
		t.Fatalf("fire[0] = %+v, want seat=0 quality=dragon count=1", observed[0])
	}
	// Second fire on the same quality — count is now 2.
	if observed[1].count != 2 {
		t.Fatalf("second behold should fire with count=2, got %d", observed[1].count)
	}
	// Third fire is seat 1's Squirrel, count 1 (independent registry).
	if observed[2] != (fire{1, "squirrel", "Chatterfang", 1}) {
		t.Fatalf("fire[2] = %+v, want seat=1 quality=squirrel count=1", observed[2])
	}
}

// ---------------------------------------------------------------------------
// (e) Turn-end (turn-start of next turn) clears the registry
// ---------------------------------------------------------------------------

func TestBehold_UntapAllClearsRegistry(t *testing.T) {
	gs := newBeholdGame(t)
	Behold(gs, 0, "Dragon", "src")
	Behold(gs, 1, "Squirrel", "src")

	if !HasBeheld(gs, 0, "Dragon") || !HasBeheld(gs, 1, "Squirrel") {
		t.Fatal("setup: both seats should have beholds recorded")
	}

	// Start-of-turn untap is the canonical reset hook for behold.
	UntapAll(gs, 0)

	if HasBeheld(gs, 0, "Dragon") {
		t.Fatal("§701.4 'this turn' window must close at UntapAll")
	}
	if HasBeheld(gs, 1, "Squirrel") {
		t.Fatal("behold registry reset must clear ALL seats, not just the active one")
	}
	if BeheldCount(gs, 0, "Dragon") != 0 {
		t.Fatal("BeheldCount should be 0 after reset")
	}
}

func TestBehold_ClearBeholdRegistryDirect(t *testing.T) {
	gs := newBeholdGame(t)
	Behold(gs, 0, "Dragon", "src")
	ClearBeholdRegistry(gs)
	if HasBeheld(gs, 0, "Dragon") {
		t.Fatal("ClearBeholdRegistry must drop all entries")
	}
	// Subsequent Behold calls re-initialize the registry cleanly.
	Behold(gs, 0, "Dragon", "src")
	if BeheldCount(gs, 0, "Dragon") != 1 {
		t.Fatal("Behold after reset should start counting from 1 again")
	}
}

// ---------------------------------------------------------------------------
// Quality matchers — type-line fallback (Elder Dragon, etc.)
// ---------------------------------------------------------------------------

func TestBehold_RevealMatchesTypeLineSubtype(t *testing.T) {
	gs := newBeholdGame(t)
	// Card with "Dragon" only in TypeLine, not split into Types[].
	c := &Card{
		Name:     "Bladewing the Risen",
		Owner:    0,
		Types:    []string{"creature"},
		TypeLine: "Legendary Creature — Zombie Dragon",
		AST:      &gameast.CardAST{Name: "Bladewing the Risen"},
	}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, c)

	if !BeholdRevealFromHand(gs, 0, "Dragon", "Source", c) {
		t.Fatal("TypeLine substring should let 'Dragon' match a Zombie Dragon")
	}
}

// ---------------------------------------------------------------------------
// Nil safety + edge cases
// ---------------------------------------------------------------------------

func TestBehold_NilSafe(t *testing.T) {
	if Behold(nil, 0, "Dragon", "src") != 0 {
		t.Fatal("Behold(nil game) should return 0")
	}
	if BeholdRevealFromHand(nil, 0, "Dragon", "src", &Card{}) {
		t.Fatal("BeholdRevealFromHand(nil game) should be false")
	}
	if BeholdChoosePermanent(nil, 0, "Dragon", "src", &Permanent{}) {
		t.Fatal("BeholdChoosePermanent(nil game) should be false")
	}
	if HasBeheld(nil, 0, "Dragon") {
		t.Fatal("HasBeheld(nil game) should be false")
	}
	if BeheldCount(nil, 0, "Dragon") != 0 {
		t.Fatal("BeheldCount(nil game) should be 0")
	}
	ClearBeholdRegistry(nil) // must not panic

	gs := newBeholdGame(t)
	if Behold(gs, -1, "Dragon", "src") != 0 {
		t.Fatal("Behold with invalid seat should return 0")
	}
	if Behold(gs, 0, "", "src") != 0 {
		t.Fatal("Behold with empty quality should be a no-op")
	}
	if Behold(gs, 0, "  ", "src") != 0 {
		t.Fatal("Behold with whitespace quality should be a no-op")
	}
}
