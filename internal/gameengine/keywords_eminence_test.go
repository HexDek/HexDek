package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Eminence tests — CR §702.107 (Commander 2017)
// ---------------------------------------------------------------------------

func newEminenceGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(2107))
	return NewGameState(4, rng, nil)
}

// newEminenceCardWithOracle builds a creature with the printed
// "Eminence —" oracle text.
func newEminenceCardWithOracle(name, oracle string) *Card {
	return &Card{
		Name:  name,
		Types: []string{"legendary", "creature"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: oracle},
			},
		},
	}
}

func newPlainEminenceCard(name string) *Card {
	return &Card{
		Name:  name,
		Types: []string{"creature"},
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{},
		},
	}
}

// putInCommandZone parks `card` in `seat`'s command zone.
func putInCommandZone(gs *GameState, seat int, card *Card) {
	gs.Seats[seat].CommandZone = append(gs.Seats[seat].CommandZone, card)
}

// putOnBattlefield parks a Permanent wrapping `card` on `seat`'s
// battlefield. Returns the perm.
func putOnBattlefieldE(gs *GameState, seat int, card *Card) *Permanent {
	p := &Permanent{Card: card, Controller: seat, Owner: seat}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, p)
	return p
}

// ---------------------------------------------------------------------------
// (d) HasEminence detector
// ---------------------------------------------------------------------------

func TestHasEminence_OraclePrefix(t *testing.T) {
	c := newEminenceCardWithOracle("The Ur-Dragon",
		"Eminence — As long as The Ur-Dragon is in the command zone or on the battlefield, other Dragon spells you cast cost {1} less to cast.")
	if !HasEminence(c) {
		t.Fatal("HasEminence should detect oracle text \"Eminence —\"")
	}
}

func TestHasEminence_CaseInsensitive(t *testing.T) {
	c := newEminenceCardWithOracle("Caps Lock", "EMINENCE — Big shouting.")
	if !HasEminence(c) {
		t.Fatal("HasEminence should be case-insensitive")
	}
}

func TestHasEminence_KeywordAST(t *testing.T) {
	c := &Card{
		Name: "Direct",
		AST: &gameast.CardAST{
			Name:      "Direct",
			Abilities: []gameast.Ability{&gameast.Keyword{Name: "eminence"}},
		},
	}
	if !HasEminence(c) {
		t.Fatal("HasEminence should detect a direct Keyword AST node")
	}
}

func TestHasEminence_Negative(t *testing.T) {
	if HasEminence(newPlainEminenceCard("Plain")) {
		t.Fatal("HasEminence must be false on a vanilla card")
	}
}

func TestHasEminence_Nil(t *testing.T) {
	if HasEminence(nil) {
		t.Fatal("HasEminence(nil) must be false")
	}
	if HasEminence(&Card{Name: "no-AST"}) {
		t.Fatal("HasEminence on AST-less card must be false")
	}
}

// ---------------------------------------------------------------------------
// (a) EminenceActive true when card in command zone
// ---------------------------------------------------------------------------

func TestEminenceActive_InCommandZone(t *testing.T) {
	gs := newEminenceGame(t)
	c := newEminenceCardWithOracle("The Ur-Dragon",
		"Eminence — Other Dragon spells you cast cost {1} less.")
	putInCommandZone(gs, 0, c)

	if !EminenceActive(gs, c) {
		t.Error("EminenceActive should be true for a card in any seat's command zone")
	}
	if got := EminenceZone(gs, c); got != "command_zone" {
		t.Errorf("EminenceZone = %q, want \"command_zone\"", got)
	}
	seat, ok := EminenceController(gs, c)
	if !ok || seat != 0 {
		t.Errorf("EminenceController = (%d, %v), want (0, true)", seat, ok)
	}
}

func TestEminenceActive_InAnyOpponentsCommandZone(t *testing.T) {
	// CR §702.107 — eminence cards function in THEIR OWNER's command
	// zone. The predicate honors any seat's CommandZone; EminenceController
	// returns the seat that holds it (the controller for "you" clauses).
	gs := newEminenceGame(t)
	c := newEminenceCardWithOracle("Their Eminence", "Eminence — Stuff.")
	putInCommandZone(gs, 2, c)

	if !EminenceActive(gs, c) {
		t.Error("EminenceActive should be true regardless of which seat holds the command-zone card")
	}
	if seat, _ := EminenceController(gs, c); seat != 2 {
		t.Errorf("EminenceController = %d, want 2 (the seat whose command zone holds the card)", seat)
	}
}

// ---------------------------------------------------------------------------
// (b) EminenceActive true when on battlefield
// ---------------------------------------------------------------------------

func TestEminenceActive_OnBattlefield(t *testing.T) {
	gs := newEminenceGame(t)
	c := newEminenceCardWithOracle("Ur-Dragon",
		"Eminence — Other Dragon spells you cast cost {1} less.")
	p := putOnBattlefieldE(gs, 1, c)

	if !EminenceActive(gs, c) {
		t.Error("EminenceActive should be true for a card on the battlefield")
	}
	if got := EminenceZone(gs, c); got != "battlefield" {
		t.Errorf("EminenceZone = %q, want \"battlefield\"", got)
	}
	seat, ok := EminenceController(gs, c)
	if !ok || seat != 1 {
		t.Errorf("EminenceController = (%d, %v), want (1, true)", seat, ok)
	}
	_ = p
}

func TestEminenceActive_PhasedOutDoesNotCount(t *testing.T) {
	gs := newEminenceGame(t)
	c := newEminenceCardWithOracle("Phased", "Eminence — Yawn.")
	p := putOnBattlefieldE(gs, 0, c)
	p.PhasedOut = true

	if EminenceActive(gs, c) {
		t.Error("EminenceActive should be false for a phased-out permanent (§702.26)")
	}
	if got := EminenceZone(gs, c); got != "" {
		t.Errorf("EminenceZone = %q, want \"\" for phased-out perm", got)
	}
}

// ---------------------------------------------------------------------------
// (c) EminenceActive false when in grave/exile/hand
// ---------------------------------------------------------------------------

func TestEminenceActive_InGraveyardFalse(t *testing.T) {
	gs := newEminenceGame(t)
	c := newEminenceCardWithOracle("Ur", "Eminence — X.")
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, c)

	if EminenceActive(gs, c) {
		t.Error("EminenceActive should be false in graveyard")
	}
}

func TestEminenceActive_InExileFalse(t *testing.T) {
	gs := newEminenceGame(t)
	c := newEminenceCardWithOracle("Ur", "Eminence — X.")
	gs.Seats[0].Exile = append(gs.Seats[0].Exile, c)

	if EminenceActive(gs, c) {
		t.Error("EminenceActive should be false in exile")
	}
}

func TestEminenceActive_InHandFalse(t *testing.T) {
	gs := newEminenceGame(t)
	c := newEminenceCardWithOracle("Ur", "Eminence — X.")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, c)

	if EminenceActive(gs, c) {
		t.Error("EminenceActive should be false in hand")
	}
}

func TestEminenceActive_InLibraryFalse(t *testing.T) {
	gs := newEminenceGame(t)
	c := newEminenceCardWithOracle("Ur", "Eminence — X.")
	gs.Seats[0].Library = append(gs.Seats[0].Library, c)

	if EminenceActive(gs, c) {
		t.Error("EminenceActive should be false in library")
	}
}

// ---------------------------------------------------------------------------
// EminenceController returns -1, false when card not in any zone
// ---------------------------------------------------------------------------

func TestEminenceController_NotPresentReturnsMinusOne(t *testing.T) {
	gs := newEminenceGame(t)
	c := newEminenceCardWithOracle("Floating", "Eminence — Phantom.")
	seat, ok := EminenceController(gs, c)
	if ok || seat != -1 {
		t.Errorf("EminenceController on absent card = (%d, %v), want (-1, false)", seat, ok)
	}
	if got := EminenceZone(gs, c); got != "" {
		t.Errorf("EminenceZone on absent card = %q, want \"\"", got)
	}
}

// ---------------------------------------------------------------------------
// (e) Ur-Dragon-style cost reduction applies from the command zone
// ---------------------------------------------------------------------------
//
// The existing ScanCostModifiers code in cost_modifiers.go ALREADY
// special-cases The Ur-Dragon's command-zone eminence discount. This
// test verifies that the engine's static-ability scan finds the
// discount via the same predicate path the new EminenceActive helper
// reads. We construct The Ur-Dragon in seat 0's command zone, queue
// up another Dragon spell cast by seat 0, and assert the scan emits
// a -1 cost modifier sourced to the eminence ability.

func newDragonSpell(name string, owner int) *Card {
	return &Card{
		Name:     name,
		Owner:    owner,
		Types:    []string{"creature", "dragon"},
		TypeLine: "Creature — Dragon",
		CMC:      6,
		AST:      &gameast.CardAST{Name: name, Abilities: []gameast.Ability{}},
	}
}

func TestEminence_UrDragonDiscountFromCommandZone(t *testing.T) {
	gs := newEminenceGame(t)

	urDragon := &Card{
		Name:     "The Ur-Dragon",
		Owner:    0,
		Types:    []string{"legendary", "creature", "dragon"},
		TypeLine: "Legendary Creature — Dragon Avatar",
		AST: &gameast.CardAST{
			Name: "The Ur-Dragon",
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: "Eminence — As long as The Ur-Dragon is in the command zone or on the battlefield, other Dragon spells you cast cost {1} less to cast."},
			},
		},
	}
	putInCommandZone(gs, 0, urDragon)

	// EminenceActive sanity.
	if !EminenceActive(gs, urDragon) {
		t.Fatal("Ur-Dragon should be eminence-active in command zone")
	}
	if !HasEminence(urDragon) {
		t.Fatal("HasEminence should detect Ur-Dragon's eminence ability")
	}

	// Cast scan: another Dragon for seat 0 should pick up the -1.
	dragon := newDragonSpell("Bladewing the Risen", 0)
	mods := ScanCostModifiers(gs, dragon, 0)

	foundReduction := false
	for _, m := range mods {
		if m.Kind == CostModReduction && m.Amount == 1 &&
			m.Source == "The Ur-Dragon (eminence, command zone)" {
			foundReduction = true
		}
	}
	if !foundReduction {
		t.Errorf("expected a -1 cost modifier from Ur-Dragon eminence; got %v", mods)
	}
}

func TestEminence_UrDragonDiscountAppliesOnBattlefieldToo(t *testing.T) {
	gs := newEminenceGame(t)

	urDragon := &Card{
		Name:     "The Ur-Dragon",
		Owner:    0,
		Types:    []string{"legendary", "creature", "dragon"},
		TypeLine: "Legendary Creature — Dragon Avatar",
		AST: &gameast.CardAST{
			Name: "The Ur-Dragon",
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: "Eminence — Other Dragon spells you cast cost {1} less to cast."},
			},
		},
	}
	// On the battlefield instead of command zone.
	putOnBattlefieldE(gs, 0, urDragon)

	if !EminenceActive(gs, urDragon) {
		t.Fatal("Ur-Dragon should be eminence-active on battlefield")
	}
	if got := EminenceZone(gs, urDragon); got != "battlefield" {
		t.Errorf("EminenceZone = %q, want \"battlefield\"", got)
	}

	// The battlefield branch of ScanCostModifiers ALSO emits a -1 for
	// other Dragons (the existing scan handles it via the per-perm
	// switch case on Card.Name). Verify a -1 still lands in the mods.
	dragon := newDragonSpell("Bogardan Hellkite", 0)
	mods := ScanCostModifiers(gs, dragon, 0)
	hasMinusOne := false
	for _, m := range mods {
		if m.Kind == CostModReduction && m.Amount == 1 {
			hasMinusOne = true
		}
	}
	if !hasMinusOne {
		t.Errorf("expected a -1 cost modifier from battlefield Ur-Dragon eminence; got %v", mods)
	}
}

// ---------------------------------------------------------------------------
// Nil-safety + multi-seat scanning
// ---------------------------------------------------------------------------

func TestEminenceActive_NilSafe(t *testing.T) {
	if EminenceActive(nil, nil) {
		t.Fatal("EminenceActive(nil, nil) must be false")
	}
	gs := newEminenceGame(t)
	if EminenceActive(gs, nil) {
		t.Fatal("EminenceActive with nil card must be false")
	}
	if EminenceActive(nil, &Card{Name: "X"}) {
		t.Fatal("EminenceActive with nil gs must be false")
	}
}

func TestEminenceController_NilSafe(t *testing.T) {
	if seat, ok := EminenceController(nil, nil); ok || seat != -1 {
		t.Errorf("EminenceController(nil, nil) = (%d, %v), want (-1, false)", seat, ok)
	}
}

func TestEminenceActive_FourSeatsAllScanned(t *testing.T) {
	gs := newEminenceGame(t)
	// Put the card in seat 3's command zone (last seat).
	c := newEminenceCardWithOracle("Last", "Eminence — Tail.")
	putInCommandZone(gs, 3, c)

	if !EminenceActive(gs, c) {
		t.Error("EminenceActive should scan all seats, not just seat 0")
	}
	if seat, _ := EminenceController(gs, c); seat != 3 {
		t.Errorf("EminenceController = %d, want 3", seat)
	}
}
