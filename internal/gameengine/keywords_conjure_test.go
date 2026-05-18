package gameengine

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Test helpers — Conjure
// ---------------------------------------------------------------------------

func newConjureGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(151))
	return NewGameState(2, rng, nil)
}

// installConjureLookup replaces ConjureCardLookup with a small
// in-memory factory and returns a restore function. The factory is
// case-insensitive on lookup; cards are constructed fresh on every
// call (so the test confirms ConjureCard's deep-copy contract by
// inspecting the returned card vs the prototype).
//
// Built-in catalog:
//   - "Lightning Bolt"  instant, CMC 1, R, mana cost "{R}", oracle says "deals 3 damage"
//   - "Goblin Token"    creature, CMC 1, R, 1/1 token-class
//   - "Grizzly Bears"   creature, CMC 2, G, 2/2 — carries an ETB-trigger
//                       observer so test (b) can verify ETB triggers fired
//                       on conjure-to-battlefield.
func installConjureLookup(t *testing.T) func() {
	t.Helper()
	prev := ConjureCardLookup
	ConjureCardLookup = func(name string) *Card {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "lightning bolt":
			return &Card{
				Name:           "Lightning Bolt",
				CMC:            1,
				Types:          []string{"instant"},
				Colors:         []string{"R"},
				ManaCostString: "{R}",
				AST: &gameast.CardAST{
					Name: "Lightning Bolt",
					Abilities: []gameast.Ability{
						&gameast.Static{Raw: "Lightning Bolt deals 3 damage to any target."},
					},
				},
			}
		case "grizzly bears":
			return &Card{
				Name:           "Grizzly Bears",
				CMC:            2,
				Types:          []string{"creature", "bear"},
				Colors:         []string{"G"},
				ManaCostString: "{1}{G}",
				BasePower:      2,
				BaseToughness:  2,
				AST: &gameast.CardAST{
					Name: "Grizzly Bears",
				},
			}
		case "goblin token":
			return &Card{
				Name:          "Goblin Token",
				CMC:           1,
				Types:         []string{"creature", "goblin"},
				Colors:        []string{"R"},
				BasePower:     1,
				BaseToughness: 1,
				AST:           &gameast.CardAST{Name: "Goblin Token"},
			}
		}
		return nil
	}
	return func() { ConjureCardLookup = prev }
}

// ===========================================================================
// (a) ConjureCard("Lightning Bolt", "hand") places a real card with {R} in hand
// ===========================================================================

func TestConjureCard_LightningBoltIntoHand(t *testing.T) {
	gs := newConjureGame(t)
	restore := installConjureLookup(t)
	defer restore()

	got := ConjureCard(gs, 0, "Lightning Bolt", ConjureZoneHand)
	if got == nil {
		t.Fatal("ConjureCard returned nil for a known card name")
	}

	// Card placed in hand.
	if len(gs.Seats[0].Hand) != 1 {
		t.Fatalf("expected 1 card in hand, got %d", len(gs.Seats[0].Hand))
	}
	if gs.Seats[0].Hand[0] != got {
		t.Errorf("card in hand should be the returned card")
	}

	// Real characteristics.
	if got.Name != "Lightning Bolt" {
		t.Errorf("Name: want \"Lightning Bolt\", got %q", got.Name)
	}
	if got.CMC != 1 {
		t.Errorf("CMC: want 1, got %d", got.CMC)
	}
	if got.ManaCostString != "{R}" {
		t.Errorf("ManaCostString: want \"{R}\", got %q", got.ManaCostString)
	}
	if len(got.Colors) != 1 || got.Colors[0] != "R" {
		t.Errorf("Colors: want [R], got %v", got.Colors)
	}
	if !cardHasType(got, "instant") {
		t.Errorf("conjured card should have type 'instant'; Types=%v", got.Types)
	}

	// Owner is the conjuring seat.
	if got.Owner != 0 {
		t.Errorf("Owner: want 0, got %d", got.Owner)
	}

	// Conjure event logged.
	sawEvent := false
	for _, ev := range gs.EventLog {
		if ev.Kind == "conjure" {
			if v, _ := ev.Details["name"].(string); v == "Lightning Bolt" {
				if z, _ := ev.Details["zone"].(string); z == ConjureZoneHand {
					sawEvent = true
					break
				}
			}
		}
	}
	if !sawEvent {
		t.Error("expected a conjure event for Lightning Bolt → hand")
	}
}

// ===========================================================================
// (b) Conjured permanent (battlefield zone) fires ETB triggers
// ===========================================================================

func TestConjureCard_ToBattlefieldFiresETBTriggers(t *testing.T) {
	gs := newConjureGame(t)
	restore := installConjureLookup(t)
	defer restore()

	// Use the existing TriggerHook seam to observe the
	// permanent_etb FireCardTrigger that FirePermanentETBTriggers
	// emits. If ETBs fire on conjure-to-battlefield, we'll see it
	// here.
	prev := TriggerHook
	defer func() { TriggerHook = prev }()
	var etbCount int
	var etbCard *Card
	TriggerHook = func(gs *GameState, event string, ctx map[string]interface{}) {
		if event != "permanent_etb" && event != "nonland_permanent_etb" {
			return
		}
		if event == "permanent_etb" {
			etbCount++
			if c, _ := ctx["card"].(*Card); c != nil {
				etbCard = c
			}
		}
	}

	bfCount := len(gs.Seats[0].Battlefield)
	got := ConjureCard(gs, 0, "Grizzly Bears", ConjureZoneBattlefield)
	if got == nil {
		t.Fatal("ConjureCard returned nil for Grizzly Bears → battlefield")
	}

	if len(gs.Seats[0].Battlefield) != bfCount+1 {
		t.Fatalf("expected 1 new permanent on battlefield; before=%d after=%d",
			bfCount, len(gs.Seats[0].Battlefield))
	}
	perm := gs.Seats[0].Battlefield[len(gs.Seats[0].Battlefield)-1]
	if perm.Card != got {
		t.Errorf("the new battlefield permanent should reference the returned card")
	}
	if perm.Controller != 0 || perm.Owner != 0 {
		t.Errorf("perm controller/owner: want 0/0, got %d/%d", perm.Controller, perm.Owner)
	}
	if !perm.SummoningSick {
		t.Error("conjured creature should enter summoning sick")
	}

	if etbCount != 1 {
		t.Errorf("expected exactly 1 permanent_etb trigger, got %d", etbCount)
	}
	if etbCard != got {
		t.Errorf("permanent_etb ctx.card should be the conjured card")
	}
}

// ===========================================================================
// (c) Meta["conjured"] = true on every conjured card
// ===========================================================================

func TestConjureCard_StampsMetaConjured(t *testing.T) {
	restore := installConjureLookup(t)
	defer restore()

	for _, zone := range []string{
		ConjureZoneHand,
		ConjureZoneLibrary,
		ConjureZoneGraveyard,
		ConjureZoneExile,
		ConjureZoneBattlefield,
	} {
		gsLocal := newConjureGame(t)
		got := ConjureCard(gsLocal, 0, "Lightning Bolt", zone)
		if got == nil {
			t.Errorf("zone=%q: ConjureCard returned nil unexpectedly", zone)
			continue
		}
		if !IsConjured(got) {
			t.Errorf("zone=%q: Meta[conjured]=true should be set", zone)
		}
		v, _ := got.Meta["conjured"].(bool)
		if !v {
			t.Errorf("zone=%q: raw Meta[conjured] should be the bool true", zone)
		}
	}
}

func TestIsConjured_NilAndDefaults(t *testing.T) {
	if IsConjured(nil) {
		t.Error("IsConjured(nil) should be false")
	}
	if IsConjured(&Card{}) {
		t.Error("IsConjured on a fresh non-conjured card should be false")
	}
	c := &Card{Meta: map[string]any{"conjured": false}}
	if IsConjured(c) {
		t.Error("IsConjured should be false when Meta[conjured]==false")
	}
}

// ===========================================================================
// (d) Unknown card name returns nil (graceful fallback)
// ===========================================================================

func TestConjureCard_UnknownNameReturnsNil(t *testing.T) {
	gs := newConjureGame(t)
	restore := installConjureLookup(t)
	defer restore()

	got := ConjureCard(gs, 0, "Does Not Exist", ConjureZoneHand)
	if got != nil {
		t.Fatalf("ConjureCard for unknown name should return nil; got %v", got)
	}

	// Side-effect guards: hand empty, no conjure event.
	if len(gs.Seats[0].Hand) != 0 {
		t.Errorf("hand should remain empty; got %d cards", len(gs.Seats[0].Hand))
	}
	for _, ev := range gs.EventLog {
		if ev.Kind == "conjure" {
			t.Fatalf("no conjure event should be logged for unknown name; got %+v", ev)
		}
	}
}

func TestConjureCard_NoLookupReturnsNil(t *testing.T) {
	prev := ConjureCardLookup
	ConjureCardLookup = nil
	defer func() { ConjureCardLookup = prev }()

	gs := newConjureGame(t)
	got := ConjureCard(gs, 0, "Lightning Bolt", ConjureZoneHand)
	if got != nil {
		t.Fatal("ConjureCard with no lookup wired should return nil")
	}
}

// ===========================================================================
// (e) Conjuring into library appends to top by default
// ===========================================================================

func TestConjureCard_LibraryPlacesOnTop(t *testing.T) {
	gs := newConjureGame(t)
	restore := installConjureLookup(t)
	defer restore()

	// Pre-seed the library with two known cards so we can confirm
	// the conjured card lands on TOP (index 0) and the existing
	// cards shift down.
	pre1 := &Card{Name: "Pre1", Owner: 0, Types: []string{"instant"}}
	pre2 := &Card{Name: "Pre2", Owner: 0, Types: []string{"sorcery"}}
	gs.Seats[0].Library = []*Card{pre1, pre2}

	got := ConjureCard(gs, 0, "Lightning Bolt", ConjureZoneLibrary)
	if got == nil {
		t.Fatal("ConjureCard returned nil")
	}
	lib := gs.Seats[0].Library
	if len(lib) != 3 {
		t.Fatalf("expected library length 3, got %d", len(lib))
	}
	if lib[0] != got {
		t.Errorf("conjured card should be at library[0] (top), got %v", lib[0])
	}
	if lib[1] != pre1 || lib[2] != pre2 {
		t.Errorf("pre-existing cards should shift down; got [%v, %v, %v]", lib[0], lib[1], lib[2])
	}
}

// ===========================================================================
// Independence: repeated conjures of the same name produce independent
// card objects (deep-copy contract).
// ===========================================================================

func TestConjureCard_DeepCopyIndependence(t *testing.T) {
	gs := newConjureGame(t)
	restore := installConjureLookup(t)
	defer restore()

	a := ConjureCard(gs, 0, "Lightning Bolt", ConjureZoneHand)
	b := ConjureCard(gs, 0, "Lightning Bolt", ConjureZoneHand)
	if a == nil || b == nil {
		t.Fatal("both conjures should succeed")
	}
	if a == b {
		t.Fatal("repeated conjures must return distinct card objects")
	}
	// Mutating one's Types slice must not affect the other.
	a.Types = append(a.Types, "sorcery")
	for _, ty := range b.Types {
		if ty == "sorcery" {
			t.Fatal("conjured cards must not share Types slices")
		}
	}
	// Mutating one's Meta must not affect the other.
	a.Meta["extra"] = "marker"
	if _, ok := b.Meta["extra"]; ok {
		t.Fatal("conjured cards must not share Meta maps")
	}
}

// ===========================================================================
// Owner re-homing across seats
// ===========================================================================

func TestConjureCard_OwnerSetToConjuringSeat(t *testing.T) {
	gs := newConjureGame(t)
	restore := installConjureLookup(t)
	defer restore()

	got := ConjureCard(gs, 1, "Lightning Bolt", ConjureZoneHand)
	if got == nil {
		t.Fatal("ConjureCard returned nil")
	}
	if got.Owner != 1 {
		t.Errorf("Owner should be the conjuring seat (1), got %d", got.Owner)
	}
	if len(gs.Seats[1].Hand) != 1 || gs.Seats[1].Hand[0] != got {
		t.Error("card should land in seat 1's hand")
	}
	if len(gs.Seats[0].Hand) != 0 {
		t.Errorf("seat 0's hand should remain empty, got %d", len(gs.Seats[0].Hand))
	}
}

// ===========================================================================
// Per-turn tracker
// ===========================================================================

func TestConjuredCardsThisTurn(t *testing.T) {
	gs := newConjureGame(t)
	restore := installConjureLookup(t)
	defer restore()

	if ConjuredCardsThisTurn(gs, 0) != 0 {
		t.Fatal("ConjuredCardsThisTurn should start at 0")
	}
	_ = ConjureCard(gs, 0, "Lightning Bolt", ConjureZoneHand)
	_ = ConjureCard(gs, 0, "Lightning Bolt", ConjureZoneExile)
	if got := ConjuredCardsThisTurn(gs, 0); got != 2 {
		t.Errorf("ConjuredCardsThisTurn(0): want 2, got %d", got)
	}
	if got := ConjuredCardsThisTurn(gs, 1); got != 0 {
		t.Errorf("ConjuredCardsThisTurn(1): want 0, got %d", got)
	}
}

// ===========================================================================
// Argument validation
// ===========================================================================

func TestConjureCard_RejectsInvalidArgs(t *testing.T) {
	gs := newConjureGame(t)
	restore := installConjureLookup(t)
	defer restore()

	cases := []struct {
		name string
		fn   func() *Card
	}{
		{"nil gs", func() *Card { return ConjureCard(nil, 0, "Lightning Bolt", ConjureZoneHand) }},
		{"invalid seat (negative)", func() *Card { return ConjureCard(gs, -1, "Lightning Bolt", ConjureZoneHand) }},
		{"invalid seat (out of range)", func() *Card { return ConjureCard(gs, 99, "Lightning Bolt", ConjureZoneHand) }},
		{"empty name", func() *Card { return ConjureCard(gs, 0, "", ConjureZoneHand) }},
		{"whitespace-only name", func() *Card { return ConjureCard(gs, 0, "   ", ConjureZoneHand) }},
		{"unknown zone", func() *Card { return ConjureCard(gs, 0, "Lightning Bolt", "stack") }},
	}
	for _, tc := range cases {
		if got := tc.fn(); got != nil {
			t.Errorf("%s: should return nil, got %v", tc.name, got)
		}
	}
}

// ===========================================================================
// Other zones (graveyard / exile) sanity
// ===========================================================================

func TestConjureCard_GraveyardAndExile(t *testing.T) {
	gs := newConjureGame(t)
	restore := installConjureLookup(t)
	defer restore()

	gravCard := ConjureCard(gs, 0, "Lightning Bolt", ConjureZoneGraveyard)
	if gravCard == nil {
		t.Fatal("conjure-to-graveyard returned nil")
	}
	if len(gs.Seats[0].Graveyard) != 1 || gs.Seats[0].Graveyard[0] != gravCard {
		t.Error("graveyard should contain the conjured card")
	}

	exileCard := ConjureCard(gs, 0, "Lightning Bolt", ConjureZoneExile)
	if exileCard == nil {
		t.Fatal("conjure-to-exile returned nil")
	}
	if len(gs.Seats[0].Exile) != 1 || gs.Seats[0].Exile[0] != exileCard {
		t.Error("exile should contain the conjured card")
	}
	// Independence between zones.
	if gravCard == exileCard {
		t.Error("two separate conjures should produce distinct cards")
	}
}
