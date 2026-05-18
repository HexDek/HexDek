package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Test helpers — Sweep (CR §702.32)
// ---------------------------------------------------------------------------

func newSweepGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(541))
	return NewGameState(2, rng, nil)
}

// addBasicLand drops a basic land of `landType` ("plains", "mountain",
// etc.) onto `seat`'s battlefield. Card.Types interleaves the "basic"
// supertype with "land" and the subtype, matching the runtime Card
// shape the rest of the engine assumes.
func addBasicLand(gs *GameState, seat int, landType string, name string) *Permanent {
	c := &Card{
		Name:  name,
		Owner: seat,
		Types: []string{"basic", "land", landType},
	}
	p := &Permanent{
		Card: c, Controller: seat, Owner: seat,
		Timestamp: gs.NextTimestamp(),
		Counters:  map[string]int{}, Flags: map[string]int{},
	}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, p)
	return p
}

// addNonBasicLand drops a non-basic land (lacks the "basic" supertype)
// of `landType` on the battlefield.
func addNonBasicLand(gs *GameState, seat int, landType string, name string) *Permanent {
	c := &Card{
		Name:  name,
		Owner: seat,
		Types: []string{"land", landType}, // no "basic"
	}
	p := &Permanent{
		Card: c, Controller: seat, Owner: seat,
		Timestamp: gs.NextTimestamp(),
		Counters:  map[string]int{}, Flags: map[string]int{},
	}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, p)
	return p
}

// sweepHandCard builds a sorcery with the sweep keyword tagged for
// `landType`.
func sweepHandCard(gs *GameState, seat int, name, landType string) *Card {
	c := &Card{
		Name:  name,
		Owner: seat,
		CMC:   3,
		Types: []string{"sorcery"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "sweep", Args: []any{landType}},
			},
		},
	}
	gs.Seats[seat].Hand = append(gs.Seats[seat].Hand, c)
	return c
}

// ===========================================================================
// HasSweep detector
// ===========================================================================

func TestHasSweep_DetectsKeywordWithMatchingArg(t *testing.T) {
	c := &Card{
		AST: &gameast.CardAST{
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "sweep", Args: []any{"plains"}},
			},
		},
	}
	if !HasSweep(c, "plains") {
		t.Fatal("HasSweep should match keyword arg 'plains' for landType 'plains'")
	}
	if !HasSweep(c, "PLAINS") {
		t.Fatal("HasSweep should be case-insensitive")
	}
	if HasSweep(c, "mountain") {
		t.Fatal("HasSweep should reject mismatched land type")
	}
}

func TestHasSweep_DetectsOracleText(t *testing.T) {
	c := &Card{
		Name: "Eerie Procession",
		AST: &gameast.CardAST{
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: "Sweep — Return any number of Plains you control to your hand. Search your library for a card..."},
			},
		},
	}
	if !HasSweep(c, "plains") {
		t.Fatal("HasSweep should detect oracle text 'Sweep —' with matching land type")
	}
	if HasSweep(c, "mountain") {
		t.Fatal("HasSweep should NOT match wrong land type via oracle text")
	}
}

func TestHasSweep_NegativeCases(t *testing.T) {
	if HasSweep(nil, "plains") {
		t.Fatal("HasSweep(nil, ...) should be false")
	}
	if HasSweep(&Card{}, "") {
		t.Fatal("HasSweep with empty landType should be false")
	}
	if HasSweep(&Card{AST: &gameast.CardAST{}}, "plains") {
		t.Fatal("HasSweep on empty AST should be false")
	}
	// Card whose flavor text mentions "Plains" but lacks the "sweep"
	// prefix — should NOT match.
	c := &Card{
		AST: &gameast.CardAST{
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: "The plains stretched out under the dying sun."},
			},
		},
	}
	if HasSweep(c, "plains") {
		t.Fatal("HasSweep must require both 'sweep' prefix AND landType")
	}
}

func TestHasSweepKeyword_AnyClause(t *testing.T) {
	c := &Card{
		AST: &gameast.CardAST{
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "sweep", Args: []any{"mountain"}},
			},
		},
	}
	if !HasSweepKeyword(c) {
		t.Fatal("HasSweepKeyword should match any sweep keyword")
	}
	if HasSweepKeyword(nil) {
		t.Fatal("HasSweepKeyword(nil) should be false")
	}
	if HasSweepKeyword(&Card{AST: &gameast.CardAST{}}) {
		t.Fatal("HasSweepKeyword on empty AST should be false")
	}
}

// ===========================================================================
// (a) Cast with 3 Plains returns succeeds + stamps count=3
// ===========================================================================

func TestCastWithSweep_ThreePlainsSucceeds(t *testing.T) {
	gs := newSweepGame(t)
	gs.Active = 0
	p1 := addBasicLand(gs, 0, "plains", "Plains 1")
	p2 := addBasicLand(gs, 0, "plains", "Plains 2")
	p3 := addBasicLand(gs, 0, "plains", "Plains 3")
	spell := sweepHandCard(gs, 0, "Eerie Procession", "plains")

	item, err := CastWithSweep(gs, 0, spell, "plains", []*Permanent{p1, p2, p3})
	if err != nil {
		t.Fatalf("CastWithSweep: %v", err)
	}
	if item == nil {
		t.Fatal("CastWithSweep returned nil item on success")
	}

	// CostMeta stamped with count=3 and land type=plains.
	if v, _ := item.CostMeta["sweep_count"].(int); v != 3 {
		t.Errorf("CostMeta[sweep_count] = %v, want 3", item.CostMeta["sweep_count"])
	}
	if v, _ := item.CostMeta["sweep_land_type"].(string); v != "plains" {
		t.Errorf("CostMeta[sweep_land_type] = %v, want \"plains\"", item.CostMeta["sweep_land_type"])
	}
	if v, _ := item.CostMeta["sweep"].(bool); !v {
		t.Errorf("CostMeta[sweep] = %v, want true", item.CostMeta["sweep"])
	}
	if v, _ := item.CostMeta["alt_cost"].(string); v != "sweep" {
		t.Errorf("CostMeta[alt_cost] = %v, want \"sweep\"", item.CostMeta["alt_cost"])
	}
	if item.CastZone != ZoneHand {
		t.Errorf("CastZone = %v, want ZoneHand", item.CastZone)
	}

	// All 3 plains removed from battlefield + back in hand.
	for _, p := range []*Permanent{p1, p2, p3} {
		for _, bp := range gs.Seats[0].Battlefield {
			if bp == p {
				t.Errorf("plains %s should be removed from battlefield", p.Card.Name)
			}
		}
	}
	if len(gs.Seats[0].Hand) != 3 {
		t.Errorf("hand should have the 3 returned plains, got %d cards", len(gs.Seats[0].Hand))
	}
	// And the spell itself is on the stack, not in hand.
	if len(gs.Stack) != 1 || gs.Stack[0] != item {
		t.Errorf("stack should have the spell on top, got %d items", len(gs.Stack))
	}
	for _, c := range gs.Seats[0].Hand {
		if c == spell {
			t.Error("spell should be removed from hand after cast")
		}
	}
	// sweep_cast event logged with amount=3.
	saw := false
	for _, ev := range gs.EventLog {
		if ev.Kind == "sweep_cast" && ev.Amount == 3 {
			if rule, _ := ev.Details["rule"].(string); rule == "702.32a" {
				saw = true
				break
			}
		}
	}
	if !saw {
		t.Error("expected sweep_cast event with rule=702.32a and amount=3")
	}
	// Per-turn tracker.
	if got := SpellSweepThisTurn(gs, 0); got != 1 {
		t.Errorf("SpellSweepThisTurn: want 1, got %d", got)
	}
}

// ===========================================================================
// (b) Non-basic land rejected
// ===========================================================================

func TestCastWithSweep_NonBasicLandRejected(t *testing.T) {
	gs := newSweepGame(t)
	gs.Active = 0
	plains := addBasicLand(gs, 0, "plains", "Basic Plains")
	shock := addNonBasicLand(gs, 0, "plains", "Hallowed Fountain") // typed plains but not basic
	spell := sweepHandCard(gs, 0, "Eerie Procession", "plains")

	item, err := CastWithSweep(gs, 0, spell, "plains", []*Permanent{plains, shock})
	if err == nil || item != nil {
		t.Fatal("CastWithSweep must reject a non-basic land in the return list")
	}
	if ce, ok := err.(*CastError); !ok || ce.Reason != "target_not_basic" {
		t.Errorf("expected CastError(target_not_basic), got %T %v", err, err)
	}

	// Side-effect guards: NEITHER land returned, spell still in hand,
	// no stack push, no event.
	for _, p := range []*Permanent{plains, shock} {
		found := false
		for _, bp := range gs.Seats[0].Battlefield {
			if bp == p {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s should remain on battlefield after rejection", p.Card.Name)
		}
	}
	if len(gs.Seats[0].Hand) != 1 || gs.Seats[0].Hand[0] != spell {
		t.Errorf("spell should remain in hand on rejection; hand=%d", len(gs.Seats[0].Hand))
	}
	if len(gs.Stack) != 0 {
		t.Errorf("stack should be empty on rejection, got %d", len(gs.Stack))
	}
	for _, ev := range gs.EventLog {
		if ev.Kind == "sweep_cast" {
			t.Fatal("no sweep_cast event should fire on rejection")
		}
	}
}

// ===========================================================================
// (c) Wrong land type rejected (Mountain on Plains-typed sweep)
// ===========================================================================

func TestCastWithSweep_WrongLandTypeRejected(t *testing.T) {
	gs := newSweepGame(t)
	gs.Active = 0
	plains := addBasicLand(gs, 0, "plains", "Plains")
	mountain := addBasicLand(gs, 0, "mountain", "Mountain")
	spell := sweepHandCard(gs, 0, "Eerie Procession", "plains")

	// Try to return a Mountain when sweep is for Plains.
	item, err := CastWithSweep(gs, 0, spell, "plains", []*Permanent{plains, mountain})
	if err == nil || item != nil {
		t.Fatal("CastWithSweep must reject a Mountain in a Plains-sweep return list")
	}
	if ce, ok := err.(*CastError); !ok || ce.Reason != "target_wrong_land_type" {
		t.Errorf("expected CastError(target_wrong_land_type), got %T %v", err, err)
	}

	// Side effects guarded.
	for _, p := range []*Permanent{plains, mountain} {
		found := false
		for _, bp := range gs.Seats[0].Battlefield {
			if bp == p {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s should remain on battlefield", p.Card.Name)
		}
	}
	if len(gs.Seats[0].Hand) != 1 {
		t.Errorf("spell should remain in hand on rejection, got %d", len(gs.Seats[0].Hand))
	}
}

// ===========================================================================
// (d) Opponent's land rejected
// ===========================================================================

func TestCastWithSweep_OpponentLandRejected(t *testing.T) {
	gs := newSweepGame(t)
	gs.Active = 0
	// Seat 0 has one Plains; seat 1 has another Plains.
	mine := addBasicLand(gs, 0, "plains", "My Plains")
	theirs := addBasicLand(gs, 1, "plains", "Opp Plains")
	spell := sweepHandCard(gs, 0, "Eerie Procession", "plains")

	item, err := CastWithSweep(gs, 0, spell, "plains", []*Permanent{mine, theirs})
	if err == nil || item != nil {
		t.Fatal("CastWithSweep must reject an opponent's land in the return list")
	}
	if ce, ok := err.(*CastError); !ok || ce.Reason != "target_not_controller" {
		t.Errorf("expected CastError(target_not_controller), got %T %v", err, err)
	}

	// Side effects: BOTH plains still on their respective battlefields,
	// spell still in hand.
	for _, p := range []*Permanent{mine, theirs} {
		found := false
		for _, bp := range gs.Seats[p.Controller].Battlefield {
			if bp == p {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s should remain on seat %d's battlefield", p.Card.Name, p.Controller)
		}
	}
	if len(gs.Seats[0].Hand) != 1 {
		t.Errorf("caster's spell should remain in hand on rejection")
	}
}

// ===========================================================================
// (e) sweep_count=0 valid (no returns, no scale)
// ===========================================================================

func TestCastWithSweep_ZeroReturnsValid(t *testing.T) {
	gs := newSweepGame(t)
	gs.Active = 0
	plains := addBasicLand(gs, 0, "plains", "Plains")
	spell := sweepHandCard(gs, 0, "Eerie Procession", "plains")

	item, err := CastWithSweep(gs, 0, spell, "plains", nil)
	if err != nil {
		t.Fatalf("CastWithSweep with empty return slice should succeed (any number includes 0); got %v", err)
	}
	if item == nil {
		t.Fatal("CastWithSweep returned nil item")
	}
	if v, _ := item.CostMeta["sweep_count"].(int); v != 0 {
		t.Errorf("CostMeta[sweep_count]: want 0, got %v", item.CostMeta["sweep_count"])
	}
	if v, _ := item.CostMeta["sweep_land_type"].(string); v != "plains" {
		t.Errorf("CostMeta[sweep_land_type]: want plains, got %v", item.CostMeta["sweep_land_type"])
	}
	// Plains untouched.
	found := false
	for _, bp := range gs.Seats[0].Battlefield {
		if bp == plains {
			found = true
			break
		}
	}
	if !found {
		t.Error("plains should remain on battlefield when no returns chosen")
	}
	// Spell still pushed on stack (not in hand).
	if len(gs.Stack) != 1 || gs.Stack[0] != item {
		t.Errorf("expected 1 stack item, got %d", len(gs.Stack))
	}
	// Event with amount=0.
	saw := false
	for _, ev := range gs.EventLog {
		if ev.Kind == "sweep_cast" && ev.Amount == 0 {
			saw = true
			break
		}
	}
	if !saw {
		t.Error("expected sweep_cast event with amount=0")
	}
}

// ===========================================================================
// Rejection paths + extras
// ===========================================================================

func TestCastWithSweep_NoSweepKeywordRejected(t *testing.T) {
	gs := newSweepGame(t)
	gs.Active = 0
	c := &Card{Name: "Plain", Owner: 0, Types: []string{"sorcery"},
		AST: &gameast.CardAST{Name: "Plain"}}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, c)

	if _, err := CastWithSweep(gs, 0, c, "plains", nil); err == nil {
		t.Fatal("CastWithSweep must reject a card without the sweep keyword")
	} else if ce, ok := err.(*CastError); !ok || ce.Reason != "no_sweep_keyword" {
		t.Errorf("expected CastError(no_sweep_keyword), got %T %v", err, err)
	}
}

func TestCastWithSweep_NilTargetInListRejected(t *testing.T) {
	gs := newSweepGame(t)
	gs.Active = 0
	plains := addBasicLand(gs, 0, "plains", "Plains")
	spell := sweepHandCard(gs, 0, "Eerie Procession", "plains")
	if _, err := CastWithSweep(gs, 0, spell, "plains", []*Permanent{plains, nil}); err == nil {
		t.Fatal("CastWithSweep must reject nil entries in the return list")
	} else if ce, ok := err.(*CastError); !ok || ce.Reason != "nil_target_perm" {
		t.Errorf("expected CastError(nil_target_perm), got %T %v", err, err)
	}
}

func TestCastWithSweep_NotInHandRejected(t *testing.T) {
	gs := newSweepGame(t)
	gs.Active = 0
	plains := addBasicLand(gs, 0, "plains", "Plains")
	// Card has sweep keyword but is NOT in seat 0's hand.
	c := &Card{
		Name: "Floating Sweep", Owner: 0,
		Types: []string{"sorcery"},
		AST: &gameast.CardAST{
			Name: "Floating Sweep",
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "sweep", Args: []any{"plains"}},
			},
		},
	}
	if _, err := CastWithSweep(gs, 0, c, "plains", []*Permanent{plains}); err == nil {
		t.Fatal("CastWithSweep must reject a card not in the caster's hand")
	} else if ce, ok := err.(*CastError); !ok || ce.Reason != "not_in_hand" {
		t.Errorf("expected CastError(not_in_hand), got %T %v", err, err)
	}
	// Side effect: plains MUST still be on battlefield (we abort
	// before any return-to-hand fires).
	found := false
	for _, bp := range gs.Seats[0].Battlefield {
		if bp == plains {
			found = true
			break
		}
	}
	if !found {
		t.Error("plains should remain on battlefield when cast is rejected")
	}
}

func TestCastWithSweep_NilSafety(t *testing.T) {
	if _, err := CastWithSweep(nil, 0, nil, "plains", nil); err == nil {
		t.Fatal("CastWithSweep(nil, ...) should error")
	}
	gs := newSweepGame(t)
	if _, err := CastWithSweep(gs, -1, nil, "plains", nil); err == nil {
		t.Fatal("invalid seat should error")
	}
	if _, err := CastWithSweep(gs, 0, nil, "plains", nil); err == nil {
		t.Fatal("nil card should error")
	}
}

// ===========================================================================
// Per-turn tracker
// ===========================================================================

func TestSpellSweepThisTurn_IncrementsPerCast(t *testing.T) {
	gs := newSweepGame(t)
	if got := SpellSweepThisTurn(gs, 0); got != 0 {
		t.Fatalf("starts at 0; got %d", got)
	}
	// First cast.
	plains := addBasicLand(gs, 0, "plains", "P1")
	s1 := sweepHandCard(gs, 0, "Sweep1", "plains")
	gs.Active = 0
	if _, err := CastWithSweep(gs, 0, s1, "plains", []*Permanent{plains}); err != nil {
		t.Fatalf("first cast: %v", err)
	}
	if got := SpellSweepThisTurn(gs, 0); got != 1 {
		t.Errorf("after 1 cast: want 1, got %d", got)
	}
	// Second cast.
	s2 := sweepHandCard(gs, 0, "Sweep2", "plains")
	if _, err := CastWithSweep(gs, 0, s2, "plains", nil); err != nil {
		t.Fatalf("second cast: %v", err)
	}
	if got := SpellSweepThisTurn(gs, 0); got != 2 {
		t.Errorf("after 2 casts: want 2, got %d", got)
	}
	// Per-seat isolation.
	if got := SpellSweepThisTurn(gs, 1); got != 0 {
		t.Errorf("seat 1 tally: want 0, got %d", got)
	}
}
