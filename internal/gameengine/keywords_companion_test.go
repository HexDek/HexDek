package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Companion tests — CR §702.139
// ---------------------------------------------------------------------------

func newCompanionGame(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(2, rand.New(rand.NewSource(52)), nil)
}

func companionCard(name string) *Card {
	return &Card{
		Name:  name,
		Owner: 0,
		Types: []string{"creature", "legendary"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "companion"},
			},
		},
	}
}

func nonCompanionCard(name string) *Card {
	return &Card{
		Name:  name,
		Owner: 0,
		Types: []string{"creature"},
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{},
		},
	}
}

// ---------------------------------------------------------------------------
// (a) HasCompanion detect
// ---------------------------------------------------------------------------

func TestHasCompanion_Detects(t *testing.T) {
	if !HasCompanion(companionCard("Jegantha, the Wellspring")) {
		t.Fatal("HasCompanion should detect companion keyword")
	}
}

func TestHasCompanion_Negative(t *testing.T) {
	if HasCompanion(nonCompanionCard("Grizzly Bears")) {
		t.Fatal("HasCompanion should be false for non-companion")
	}
	if HasCompanion(nil) {
		t.Fatal("HasCompanion(nil) should be false")
	}
}

func TestCompanionRestriction_KnownCompanionsHaveText(t *testing.T) {
	names := []string{
		"Lurrus of the Dream-Den",
		"Yorion, Sky Nomad",
		"Jegantha, the Wellspring",
		"Kaheera, the Orphanguard",
		"Keruga, the Macrosage",
	}
	for _, n := range names {
		if got := CompanionRestriction(companionCard(n)); got == "" {
			t.Fatalf("%s should have a printed restriction descriptor", n)
		}
	}
}

func TestCompanionRestriction_NonCompanionIsEmpty(t *testing.T) {
	if got := CompanionRestriction(nonCompanionCard("Grizzly Bears")); got != "" {
		t.Fatalf("non-companion descriptor = %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// (b) Pre-game exile sets CompanionExiled
// ---------------------------------------------------------------------------

func TestPreGameExileCompanion_SetsField(t *testing.T) {
	gs := newCompanionGame(t)
	jegantha := companionCard("Jegantha, the Wellspring")

	if SeatCompanionExiled(gs.Seats[0]) != nil {
		t.Fatal("companion should be nil at game start")
	}

	PreGameExileCompanion(gs, 0, jegantha)

	if SeatCompanionExiled(gs.Seats[0]) != jegantha {
		t.Fatal("PreGameExileCompanion should set the companion pointer")
	}
	if SeatCompanionUsed(gs.Seats[0]) {
		t.Fatal("CompanionUsed should remain false right after declaration")
	}
}

func TestPreGameExileCompanion_LogsDeclaration(t *testing.T) {
	gs := newCompanionGame(t)
	jegantha := companionCard("Jegantha, the Wellspring")
	before := len(gs.EventLog)
	PreGameExileCompanion(gs, 0, jegantha)
	found := false
	for i := before; i < len(gs.EventLog); i++ {
		if gs.EventLog[i].Kind == "companion_declared" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("PreGameExileCompanion should emit companion_declared event")
	}
}

func TestPreGameExileCompanion_NilSafe(t *testing.T) {
	gs := newCompanionGame(t)
	PreGameExileCompanion(nil, 0, companionCard("Jegantha"))
	PreGameExileCompanion(gs, -1, companionCard("Jegantha"))
	PreGameExileCompanion(gs, 0, nil)
	if SeatCompanionExiled(gs.Seats[0]) != nil {
		t.Fatal("nil-arg paths must not set the companion field")
	}
}

// ---------------------------------------------------------------------------
// (c) PayCompanionCost moves to hand + flips used
// ---------------------------------------------------------------------------

func TestPayCompanionCost_MovesToHandAndFlipsUsed(t *testing.T) {
	gs := newCompanionGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	jegantha := companionCard("Jegantha, the Wellspring")
	PreGameExileCompanion(gs, 0, jegantha)

	if err := PayCompanionCost(gs, 0, jegantha); err != nil {
		t.Fatalf("PayCompanionCost returned error: %v", err)
	}
	// Mana paid (3 of 5 → 2 left).
	if gs.Seats[0].ManaPool != 2 {
		t.Fatalf("mana pool = %d, want 2", gs.Seats[0].ManaPool)
	}
	// Companion now in hand.
	foundInHand := false
	for _, c := range gs.Seats[0].Hand {
		if c == jegantha {
			foundInHand = true
			break
		}
	}
	if !foundInHand {
		t.Fatal("companion should be in hand after PayCompanionCost")
	}
	// CompanionUsed flag set.
	if !SeatCompanionUsed(gs.Seats[0]) {
		t.Fatal("SeatCompanionUsed should be true after activation")
	}
}

// ---------------------------------------------------------------------------
// (d) Used-companion can't be re-paid
// ---------------------------------------------------------------------------

func TestPayCompanionCost_OncePerGame(t *testing.T) {
	gs := newCompanionGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 10
	jegantha := companionCard("Jegantha, the Wellspring")
	PreGameExileCompanion(gs, 0, jegantha)

	if err := PayCompanionCost(gs, 0, jegantha); err != nil {
		t.Fatalf("first PayCompanionCost failed: %v", err)
	}
	// Second call must fail.
	if err := PayCompanionCost(gs, 0, jegantha); err == nil {
		t.Fatal("§702.139b: companion ability is once per game; second call must fail")
	}
	// Mana pool should reflect only ONE payment (10 - 3 = 7).
	if gs.Seats[0].ManaPool != 7 {
		t.Fatalf("mana pool = %d, want 7 (only one payment)", gs.Seats[0].ManaPool)
	}
}

func TestPayCompanionCost_RejectsTimingViolations(t *testing.T) {
	gs := newCompanionGame(t)
	gs.Seats[0].ManaPool = 5
	jegantha := companionCard("Jegantha")
	PreGameExileCompanion(gs, 0, jegantha)

	// Active is 1, not 0.
	gs.Active = 1
	if err := PayCompanionCost(gs, 0, jegantha); err == nil {
		t.Fatal("must reject when it's not the seat's turn")
	}

	// Now active, but stack non-empty.
	gs.Active = 0
	gs.Stack = append(gs.Stack, &StackItem{})
	if err := PayCompanionCost(gs, 0, jegantha); err == nil {
		t.Fatal("must reject when stack is non-empty (sorcery speed only)")
	}
}

func TestPayCompanionCost_RejectsInsufficientMana(t *testing.T) {
	gs := newCompanionGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 2
	jegantha := companionCard("Jegantha")
	PreGameExileCompanion(gs, 0, jegantha)
	if err := PayCompanionCost(gs, 0, jegantha); err == nil {
		t.Fatal("must reject when mana pool < 3")
	}
}

func TestPayCompanionCost_RejectsCardMismatch(t *testing.T) {
	gs := newCompanionGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	jegantha := companionCard("Jegantha")
	yorion := companionCard("Yorion, Sky Nomad")
	PreGameExileCompanion(gs, 0, jegantha)

	// Caller asks to pay the cost for Yorion, but Jegantha is the
	// declared companion — must fail.
	if err := PayCompanionCost(gs, 0, yorion); err == nil {
		t.Fatal("must reject a card-mismatch payment")
	}
	if SeatCompanionUsed(gs.Seats[0]) {
		t.Fatal("rejected mismatch must not flip CompanionUsed")
	}
}

func TestPayCompanionCost_NoCompanionDeclared(t *testing.T) {
	gs := newCompanionGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	if err := PayCompanionCost(gs, 0, nil); err == nil {
		t.Fatal("must reject when no companion has been declared")
	}
}

// ---------------------------------------------------------------------------
// (e) Restriction validation — Jegantha + Yorion examples
// ---------------------------------------------------------------------------

func TestValidateCompanionDeck_Jegantha_AcceptsDistinctPips(t *testing.T) {
	jegantha := companionCard("Jegantha, the Wellspring")
	// Deck with no repeated colored pips per card.
	deck := []*Card{
		{Name: "Mox Diamond", ManaCostString: "", Types: []string{"artifact"}},
		{Name: "Lightning Bolt", ManaCostString: "{R}", Types: []string{"instant"}},
		{Name: "Watery Grave", ManaCostString: "", Types: []string{"land"}},
		{Name: "Cryptic Command", ManaCostString: "{1}{U}{U}{U}", Types: []string{"instant"}}, // 3x U
	}
	// Cryptic Command has 3 blue pips — Jegantha rejects this deck.
	if ValidateCompanionDeck(jegantha, deck) {
		t.Fatal("Jegantha must reject a deck containing a card with repeated colored pips")
	}
}

func TestValidateCompanionDeck_Jegantha_AllowsSinglePips(t *testing.T) {
	jegantha := companionCard("Jegantha, the Wellspring")
	deck := []*Card{
		{Name: "Lightning Bolt", ManaCostString: "{R}", Types: []string{"instant"}},
		{Name: "Mana Drain", ManaCostString: "{U}{U}", Types: []string{"instant"}}, // repeat — should reject
	}
	if ValidateCompanionDeck(jegantha, deck) {
		t.Fatal("Jegantha must reject {U}{U}")
	}

	cleanDeck := []*Card{
		{Name: "Lightning Bolt", ManaCostString: "{R}", Types: []string{"instant"}},
		{Name: "Bant Charm", ManaCostString: "{G}{W}{U}", Types: []string{"instant"}},
		{Name: "Sol Ring", ManaCostString: "{1}", Types: []string{"artifact"}},
		{Name: "Vivid Crag", ManaCostString: "", Types: []string{"land"}},
	}
	if !ValidateCompanionDeck(jegantha, cleanDeck) {
		t.Fatal("Jegantha must accept a deck with no repeat colored pips")
	}
}

func TestValidateCompanionDeck_Yorion_RequiresEightyCards(t *testing.T) {
	yorion := companionCard("Yorion, Sky Nomad")

	// 60-card deck — too small.
	smallDeck := make([]*Card, 60)
	for i := range smallDeck {
		smallDeck[i] = &Card{Name: "Filler", Types: []string{"creature"}}
	}
	if ValidateCompanionDeck(yorion, smallDeck) {
		t.Fatal("Yorion must reject a 60-card deck")
	}

	// 80-card deck — accept.
	bigDeck := make([]*Card, 80)
	for i := range bigDeck {
		bigDeck[i] = &Card{Name: "Filler", Types: []string{"creature"}}
	}
	if !ValidateCompanionDeck(yorion, bigDeck) {
		t.Fatal("Yorion must accept an 80-card deck")
	}

	// 100-card deck (Commander) — accept.
	commDeck := make([]*Card, 100)
	for i := range commDeck {
		commDeck[i] = &Card{Name: "Filler", Types: []string{"creature"}}
	}
	if !ValidateCompanionDeck(yorion, commDeck) {
		t.Fatal("Yorion must accept a 100-card commander deck")
	}
}

func TestValidateCompanionDeck_Lurrus_RejectsHighCMCPermanents(t *testing.T) {
	lurrus := companionCard("Lurrus of the Dream-Den")
	deck := []*Card{
		{Name: "Snapcaster Mage", CMC: 2, Types: []string{"creature"}},
		{Name: "Lightning Bolt", CMC: 1, Types: []string{"instant"}}, // instant CMC>2 OK (not permanent)
		{Name: "Counterspell", CMC: 2, Types: []string{"instant"}},
		{Name: "Sol Ring", CMC: 1, Types: []string{"artifact"}},
	}
	if !ValidateCompanionDeck(lurrus, deck) {
		t.Fatal("Lurrus should accept a deck of <=2 CMC permanents + cheap instants")
	}
	// Add a 3 CMC permanent — must reject.
	deck = append(deck, &Card{Name: "Watchwolf", CMC: 2, Types: []string{"creature"}})
	deck = append(deck, &Card{Name: "Goyf", CMC: 3, Types: []string{"creature"}})
	if ValidateCompanionDeck(lurrus, deck) {
		t.Fatal("Lurrus must reject a deck containing a 3+ CMC permanent")
	}
}

func TestValidateCompanionDeck_UnknownCompanionAcceptsAnyDeck(t *testing.T) {
	unknown := companionCard("Unwired Companion Card")
	deck := []*Card{{Name: "Anything", CMC: 7, Types: []string{"creature"}}}
	if !ValidateCompanionDeck(unknown, deck) {
		t.Fatal("unwired companion validators should default to true")
	}
}

func TestValidateCompanionDeck_NilCardRejects(t *testing.T) {
	if ValidateCompanionDeck(nil, nil) {
		t.Fatal("nil companion card should not validate")
	}
}

// ---------------------------------------------------------------------------
// manaCostHasRepeatColoredPip — unit
// ---------------------------------------------------------------------------

func TestManaCostHasRepeatColoredPip(t *testing.T) {
	cases := []struct {
		cost string
		want bool
	}{
		{"{R}", false},
		{"{R}{R}", true},
		{"{1}{W}{U}", false},
		{"{2}{W}{W}", true},
		{"", false},
		{"{X}{X}", false},     // X isn't a colored pip
		{"{1}{2}{3}", false},  // generic only
		{"{G}{W}{U}{B}{R}", false},
		{"{B}{B}{R}", true},
	}
	for _, c := range cases {
		if got := manaCostHasRepeatColoredPip(c.cost); got != c.want {
			t.Errorf("manaCostHasRepeatColoredPip(%q) = %v, want %v", c.cost, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Accessor + nil safety
// ---------------------------------------------------------------------------

func TestSeatAccessors_NilSafe(t *testing.T) {
	if SeatCompanionExiled(nil) != nil {
		t.Fatal("SeatCompanionExiled(nil) should return nil")
	}
	if SeatCompanionUsed(nil) {
		t.Fatal("SeatCompanionUsed(nil) should return false")
	}
}

func TestPayCompanionCost_NilSafe(t *testing.T) {
	gs := newCompanionGame(t)
	if err := PayCompanionCost(nil, 0, nil); err == nil {
		t.Fatal("nil game should return error")
	}
	if err := PayCompanionCost(gs, -1, nil); err == nil {
		t.Fatal("invalid seat should return error")
	}
	if err := PayCompanionCost(gs, 99, nil); err == nil {
		t.Fatal("out-of-range seat should return error")
	}
}
