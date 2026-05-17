package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// Round-26 tests for keywords_partner.go. The legacy partner_test.go
// already covers the cross-category legality matrix via the lower-
// level ValidatePartnerPair entry point; this file exercises the
// card-facing surface (HasFriendsForever, HasPartnerWith,
// PartnerWithTarget, CanBeCommandersTogether) and the §702.124g
// search-and-grab ETB trigger.

// ---------------------------------------------------------------------------
// Test card builders — kept distinct from partner_test.go's makePartnerCard
// to avoid coupling the new tests to legacy fixtures.
// ---------------------------------------------------------------------------

func kp_makeBarePartner(name string) *Card {
	return &Card{
		Name:          name,
		BasePower:     2,
		BaseToughness: 2,
		Types:         []string{"legendary", "creature"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "partner", Raw: "partner"},
			},
		},
	}
}

func kp_makePartnerWith(name, target string) *Card {
	return &Card{
		Name:          name,
		BasePower:     2,
		BaseToughness: 2,
		Types:         []string{"legendary", "creature"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "partner", Raw: "partner with " + target},
			},
		},
	}
}

func kp_makeFriendsForever(name string) *Card {
	return &Card{
		Name:          name,
		BasePower:     2,
		BaseToughness: 2,
		Types:         []string{"legendary", "creature"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "friends", Raw: "friends forever"},
			},
		},
	}
}

func kp_makePlainCreature(name string) *Card {
	return &Card{
		Name:          name,
		BasePower:     1,
		BaseToughness: 1,
		Types:         []string{"creature"},
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{},
		},
	}
}

// ---------------------------------------------------------------------------
// Predicates
// ---------------------------------------------------------------------------

func TestHasFriendsForever_Positive(t *testing.T) {
	c := kp_makeFriendsForever("Ardenn, Intrepid Archaeologist")
	if !HasFriendsForever(c) {
		t.Fatal("HasFriendsForever should be true for a Friends Forever card")
	}
}

func TestHasFriendsForever_NegativeOnPlainAndPartner(t *testing.T) {
	if HasFriendsForever(kp_makePlainCreature("Grizzly Bears")) {
		t.Fatal("plain creature should not report HasFriendsForever")
	}
	if HasFriendsForever(kp_makeBarePartner("Tymna")) {
		t.Fatal("bare-Partner card should not report HasFriendsForever")
	}
	if HasFriendsForever(nil) {
		t.Fatal("HasFriendsForever(nil) should be false")
	}
}

func TestHasPartnerWith_AndTarget(t *testing.T) {
	c := kp_makePartnerWith("Pir, Imaginative Rascal", "Toothy, Imaginary Friend")
	if !HasPartnerWith(c) {
		t.Fatal("HasPartnerWith should be true")
	}
	if got := PartnerWithTarget(c); got != "Toothy, Imaginary Friend" {
		t.Fatalf("PartnerWithTarget = %q, want %q", got, "Toothy, Imaginary Friend")
	}
	// Bare Partner has no name target.
	bare := kp_makeBarePartner("Kraum")
	if HasPartnerWith(bare) {
		t.Fatal("bare Partner should not report HasPartnerWith")
	}
	if PartnerWithTarget(bare) != "" {
		t.Fatal("PartnerWithTarget should be empty for bare Partner")
	}
}

// ---------------------------------------------------------------------------
// (a) Two friends-forever cards = valid pair.
// ---------------------------------------------------------------------------

func TestCanBeCommandersTogether_FriendsForeverPair(t *testing.T) {
	a := kp_makeFriendsForever("Ardenn")
	b := kp_makeFriendsForever("Rebbec")
	if !CanBeCommandersTogether(a, b) {
		t.Fatal("two Friends Forever cards should be a valid commander pair")
	}
}

// ---------------------------------------------------------------------------
// (b) Partner with X paired correctly only when both name-match.
// ---------------------------------------------------------------------------

func TestCanBeCommandersTogether_PartnerWithMatched(t *testing.T) {
	pir := kp_makePartnerWith("Pir, Imaginative Rascal", "Toothy, Imaginary Friend")
	toothy := kp_makePartnerWith("Toothy, Imaginary Friend", "Pir, Imaginative Rascal")
	if !CanBeCommandersTogether(pir, toothy) {
		t.Fatal("matched Partner-with pair should be legal")
	}
}

func TestCanBeCommandersTogether_PartnerWithMismatched(t *testing.T) {
	pir := kp_makePartnerWith("Pir, Imaginative Rascal", "Toothy, Imaginary Friend")
	wrong := kp_makePartnerWith("Wrong Partner", "Some Other Card")
	if CanBeCommandersTogether(pir, wrong) {
		t.Fatal("Partner-with names that don't cross-match must be illegal")
	}
}

// ---------------------------------------------------------------------------
// (c) ETB search trigger fires for partner-with.
// ---------------------------------------------------------------------------

func TestOnPartnerWithETB_FetchesNamedPartner(t *testing.T) {
	gs := NewGameState(2, rand.New(rand.NewSource(99)), nil)

	pir := kp_makePartnerWith("Pir, Imaginative Rascal", "Toothy, Imaginary Friend")
	toothy := kp_makePartnerWith("Toothy, Imaginary Friend", "Pir, Imaginative Rascal")
	filler := kp_makePlainCreature("Filler 1")
	filler2 := kp_makePlainCreature("Filler 2")

	gs.Seats[0].Library = []*Card{filler, toothy, filler2}
	for _, c := range gs.Seats[0].Library {
		c.Owner = 0
	}

	perm := &Permanent{
		Card:       pir,
		Controller: 0,
		Owner:      0,
		Flags:      map[string]int{},
	}
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, perm)

	found := OnPartnerWithETB(gs, perm)
	if found == nil {
		t.Fatal("OnPartnerWithETB should return the matched partner card")
	}
	if found.DisplayName() != "Toothy, Imaginary Friend" {
		t.Fatalf("fetched the wrong card: %q", found.DisplayName())
	}
	// Toothy should now be in hand.
	inHand := false
	for _, c := range gs.Seats[0].Hand {
		if c == toothy {
			inHand = true
			break
		}
	}
	if !inHand {
		t.Fatal("Toothy should be in the controller's hand after ETB search")
	}
	// And no longer in library.
	for _, c := range gs.Seats[0].Library {
		if c == toothy {
			t.Fatal("Toothy should be removed from the library after the fetch")
		}
	}
	// Search + reveal events should be in the log.
	sawSearch, sawReveal := false, false
	for _, ev := range gs.EventLog {
		switch ev.Kind {
		case "partner_with_search":
			sawSearch = true
		case "reveal":
			sawReveal = true
		}
	}
	if !sawSearch || !sawReveal {
		t.Fatalf("expected partner_with_search + reveal events; got search=%v reveal=%v",
			sawSearch, sawReveal)
	}
}

func TestOnPartnerWithETB_NoMatchInLibraryLogsButShuffles(t *testing.T) {
	gs := NewGameState(2, rand.New(rand.NewSource(42)), nil)

	pir := kp_makePartnerWith("Pir, Imaginative Rascal", "Toothy, Imaginary Friend")
	a := kp_makePlainCreature("Alpha")
	b := kp_makePlainCreature("Bravo")
	gs.Seats[0].Library = []*Card{a, b}

	perm := &Permanent{Card: pir, Controller: 0, Owner: 0, Flags: map[string]int{}}
	if got := OnPartnerWithETB(gs, perm); got != nil {
		t.Fatalf("OnPartnerWithETB should return nil when no partner present; got %v", got.DisplayName())
	}
	// no_match event should be present.
	sawNoMatch := false
	for _, ev := range gs.EventLog {
		if ev.Kind == "partner_with_no_match" {
			sawNoMatch = true
		}
	}
	if !sawNoMatch {
		t.Fatal("expected partner_with_no_match event when named partner absent")
	}
}

func TestOnPartnerWithETB_NoOpOnBarePartner(t *testing.T) {
	gs := NewGameState(2, rand.New(rand.NewSource(7)), nil)
	bare := kp_makeBarePartner("Kraum")
	perm := &Permanent{Card: bare, Controller: 0, Owner: 0, Flags: map[string]int{}}
	if got := OnPartnerWithETB(gs, perm); got != nil {
		t.Fatal("bare Partner has no name target — ETB hook should no-op")
	}
}

// ---------------------------------------------------------------------------
// (d) Standard Partner (no "with") does not need name match.
// ---------------------------------------------------------------------------

func TestCanBeCommandersTogether_BarePartnerPair(t *testing.T) {
	kraum := kp_makeBarePartner("Kraum, Ludevic's Opus")
	tymna := kp_makeBarePartner("Tymna the Weaver")
	if !CanBeCommandersTogether(kraum, tymna) {
		t.Fatal("two bare-Partner cards should pair regardless of names")
	}
}

// ---------------------------------------------------------------------------
// (e) Friends Forever + bare Partner is NOT a valid pair.
// ---------------------------------------------------------------------------

func TestCanBeCommandersTogether_FriendsForeverPlusBarePartnerIllegal(t *testing.T) {
	ff := kp_makeFriendsForever("Ardenn")
	bare := kp_makeBarePartner("Kraum")
	if CanBeCommandersTogether(ff, bare) {
		t.Fatal("Friends Forever + bare Partner must NOT pair (cross-category mix forbidden)")
	}
}

func TestCanBeCommandersTogether_FriendsForeverPlusPartnerWithIllegal(t *testing.T) {
	ff := kp_makeFriendsForever("Ardenn")
	pw := kp_makePartnerWith("Pir", "Toothy")
	if CanBeCommandersTogether(ff, pw) {
		t.Fatal("Friends Forever + Partner with X must NOT pair (cross-category mix forbidden)")
	}
}

// Bonus: nil safety + single-commander legality.
func TestCanBeCommandersTogether_NilSafe(t *testing.T) {
	if CanBeCommandersTogether(nil, kp_makeBarePartner("X")) {
		t.Fatal("nil partner should never legalize a pair")
	}
	if CanBeCommandersTogether(kp_makeBarePartner("X"), nil) {
		t.Fatal("nil partner should never legalize a pair")
	}
}

// Bonus: PartnerNameMatch is case+space tolerant.
func TestPartnerNameMatch_CaseInsensitive(t *testing.T) {
	if !PartnerNameMatch("Toothy, Imaginary Friend", " toothy, imaginary friend ") {
		t.Fatal("PartnerNameMatch should be case- and space-tolerant")
	}
	if PartnerNameMatch("Pir", "Toothy") {
		t.Fatal("PartnerNameMatch should reject different names")
	}
}
