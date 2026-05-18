package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// Round-34 tests for CastWithConvoke (CR §702.51).

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

func cv_makeGame(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(2, rand.New(rand.NewSource(51)), nil)
}

// cv_makeConvokeSpell builds an instant/sorcery card with the convoke
// keyword and the given converted mana cost.
func cv_makeConvokeSpell(name string, cmc int) *Card {
	return &Card{
		Name:  name,
		Types: []string{"sorcery"},
		CMC:   cmc,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "convoke", Raw: "convoke"},
			},
		},
	}
}

// cv_makeCreature builds a creature permanent on `seat`'s battlefield
// with the given colors. SummoningSick is set to mirror a freshly-cast
// creature; tests that need it untapped-and-fresh use the default.
func cv_makeCreature(seat int, name string, colors ...string) *Permanent {
	c := &Card{
		Name:          name,
		Owner:         seat,
		Types:         []string{"creature"},
		BasePower:     1,
		BaseToughness: 1,
		Colors:        append([]string(nil), colors...),
	}
	return &Permanent{
		Card:          c,
		Controller:    seat,
		Owner:         seat,
		SummoningSick: true,
		Flags:         map[string]int{},
	}
}

// cv_putOnBF adds creatures to seat 0's battlefield and returns them
// for convenient chaining.
func cv_putOnBF(gs *GameState, seat int, perms ...*Permanent) {
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, perms...)
}

func cv_castErrReason(err error) string {
	if ce, ok := err.(*CastError); ok {
		return ce.Reason
	}
	if err == nil {
		return ""
	}
	return err.Error()
}

// ---------------------------------------------------------------------------
// (a) 3 creatures tapped = 3 generic discount.
// ---------------------------------------------------------------------------

func TestCastWithConvoke_ThreeCreaturesGiveThreeDiscount(t *testing.T) {
	gs := cv_makeGame(t)
	spell := cv_makeConvokeSpell("Chord of Calling", 5)
	gs.Seats[0].Hand = []*Card{spell}
	gs.Seats[0].ManaPool = 2 // 5 cost - 3 from convoke = 2 net

	a := cv_makeCreature(0, "Llanowar Elves", "G")
	b := cv_makeCreature(0, "Birds of Paradise", "G")
	c := cv_makeCreature(0, "Mox Bot", "")
	cv_putOnBF(gs, 0, a, b, c)

	res, err := CastWithConvoke(gs, 0, spell, []*Permanent{a, b, c})
	if err != nil {
		t.Fatalf("expected success; got %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil CostPaymentResult")
	}
	if gs.Seats[0].ManaPool != 0 {
		t.Fatalf("mana pool = %d, want 0 (2 paid)", gs.Seats[0].ManaPool)
	}
	for _, p := range []*Permanent{a, b, c} {
		if !p.Tapped {
			t.Fatalf("creature %s should be tapped after convoke", p.Card.Name)
		}
	}
	if len(gs.Stack) != 1 {
		t.Fatalf("expected 1 stack item, got %d", len(gs.Stack))
	}
	if v, _ := gs.Stack[0].CostMeta["convoke_reduction"].(int); v != 3 {
		t.Fatalf("convoke_reduction = %v, want 3", gs.Stack[0].CostMeta["convoke_reduction"])
	}
}

// ---------------------------------------------------------------------------
// (b) Creature with G color contributes {G} portion.
// ---------------------------------------------------------------------------

func TestCastWithConvoke_GreenCreatureContributesGreen(t *testing.T) {
	gs := cv_makeGame(t)
	spell := cv_makeConvokeSpell("Chord of Calling", 3)
	gs.Seats[0].Hand = []*Card{spell}
	gs.Seats[0].ManaPool = 2

	g := cv_makeCreature(0, "Llanowar Elves", "G")
	r := cv_makeCreature(0, "Goblin", "R")
	cv_putOnBF(gs, 0, g, r)

	if _, err := CastWithConvoke(gs, 0, spell, []*Permanent{g, r}); err != nil {
		t.Fatalf("expected success; got %v", err)
	}
	if len(gs.Stack) != 1 {
		t.Fatalf("expected 1 stack item; got %d", len(gs.Stack))
	}
	colors, _ := gs.Stack[0].CostMeta["convoke_colors"].([]string)
	if len(colors) != 2 || colors[0] != "G" || colors[1] != "R" {
		t.Fatalf("convoke_colors = %v, want [G R]", colors)
	}
}

// ---------------------------------------------------------------------------
// (c) Already-tapped creature rejected.
// ---------------------------------------------------------------------------

func TestCastWithConvoke_RejectAlreadyTappedCreature(t *testing.T) {
	gs := cv_makeGame(t)
	spell := cv_makeConvokeSpell("Chord", 3)
	gs.Seats[0].Hand = []*Card{spell}
	gs.Seats[0].ManaPool = 3

	a := cv_makeCreature(0, "Fresh", "G")
	b := cv_makeCreature(0, "Spent", "G")
	b.Tapped = true
	cv_putOnBF(gs, 0, a, b)

	_, err := CastWithConvoke(gs, 0, spell, []*Permanent{a, b})
	if cv_castErrReason(err) != "convoke_creature_already_tapped" {
		t.Fatalf("expected convoke_creature_already_tapped; got %v", err)
	}
	// State must be unchanged on failure.
	if a.Tapped {
		t.Fatal("validation-pass failure must not tap any creatures")
	}
	if gs.Seats[0].ManaPool != 3 {
		t.Fatalf("mana pool changed on failure: %d", gs.Seats[0].ManaPool)
	}
	if len(gs.Stack) != 0 {
		t.Fatalf("no stack item should be pushed on failure; got %d", len(gs.Stack))
	}
}

// ---------------------------------------------------------------------------
// (d) Opponent's creature rejected.
// ---------------------------------------------------------------------------

func TestCastWithConvoke_RejectOpponentCreature(t *testing.T) {
	gs := cv_makeGame(t)
	spell := cv_makeConvokeSpell("Chord", 3)
	gs.Seats[0].Hand = []*Card{spell}
	gs.Seats[0].ManaPool = 3

	mine := cv_makeCreature(0, "Mine", "G")
	enemy := cv_makeCreature(1, "Enemy", "G")
	cv_putOnBF(gs, 0, mine)
	cv_putOnBF(gs, 1, enemy)

	_, err := CastWithConvoke(gs, 0, spell, []*Permanent{mine, enemy})
	if cv_castErrReason(err) != "convoke_creature_not_controlled" {
		t.Fatalf("expected convoke_creature_not_controlled; got %v", err)
	}
	if mine.Tapped {
		t.Fatal("no creature should be tapped on failure")
	}
}

// ---------------------------------------------------------------------------
// (e) CostMeta stamped with all expected keys.
// ---------------------------------------------------------------------------

func TestCastWithConvoke_StampsCostMeta(t *testing.T) {
	gs := cv_makeGame(t)
	spell := cv_makeConvokeSpell("Chord", 4)
	gs.Seats[0].Hand = []*Card{spell}
	gs.Seats[0].ManaPool = 2

	a := cv_makeCreature(0, "A", "G")
	b := cv_makeCreature(0, "B", "W")
	cv_putOnBF(gs, 0, a, b)

	if _, err := CastWithConvoke(gs, 0, spell, []*Permanent{a, b}); err != nil {
		t.Fatalf("expected success; got %v", err)
	}
	if len(gs.Stack) != 1 {
		t.Fatalf("expected 1 stack item; got %d", len(gs.Stack))
	}
	cm := gs.Stack[0].CostMeta

	if v, _ := cm["alt_cost"].(string); v != "convoke" {
		t.Errorf("alt_cost = %v, want convoke", cm["alt_cost"])
	}
	if v, _ := cm["convoke_creatures_used"].(int); v != 2 {
		t.Errorf("convoke_creatures_used = %v, want 2", cm["convoke_creatures_used"])
	}
	if v, _ := cm["convoke_reduction"].(int); v != 2 {
		t.Errorf("convoke_reduction = %v, want 2", cm["convoke_reduction"])
	}
	if v, _ := cm["convoke_net_cost"].(int); v != 2 {
		t.Errorf("convoke_net_cost = %v, want 2", cm["convoke_net_cost"])
	}
	colors, _ := cm["convoke_colors"].([]string)
	if len(colors) != 2 || colors[0] != "G" || colors[1] != "W" {
		t.Errorf("convoke_colors = %v, want [G W]", colors)
	}
}

// ---------------------------------------------------------------------------
// (f) Summoning-sick creature can still convoke (CR §702.51c).
// ---------------------------------------------------------------------------

func TestCastWithConvoke_SummoningSickCreatureAllowed(t *testing.T) {
	gs := cv_makeGame(t)
	spell := cv_makeConvokeSpell("Chord", 2)
	gs.Seats[0].Hand = []*Card{spell}
	gs.Seats[0].ManaPool = 1

	sick := cv_makeCreature(0, "Just Cast", "G")
	if !sick.SummoningSick {
		t.Fatal("test setup: creature should be summoning sick")
	}
	cv_putOnBF(gs, 0, sick)

	if _, err := CastWithConvoke(gs, 0, spell, []*Permanent{sick}); err != nil {
		t.Fatalf("summoning-sick creature must be allowed to convoke; got %v", err)
	}
	if !sick.Tapped {
		t.Fatal("summoning-sick creature should be tapped after successful convoke")
	}
}

// ---------------------------------------------------------------------------
// Bonus: cost reduction is capped at card CMC (no negative cast).
// ---------------------------------------------------------------------------

func TestCastWithConvoke_ReductionCappedAtCMC(t *testing.T) {
	gs := cv_makeGame(t)
	spell := cv_makeConvokeSpell("Cheap Spell", 1)
	gs.Seats[0].Hand = []*Card{spell}
	gs.Seats[0].ManaPool = 0

	a := cv_makeCreature(0, "A", "G")
	b := cv_makeCreature(0, "B", "G")
	c := cv_makeCreature(0, "C", "G")
	cv_putOnBF(gs, 0, a, b, c)

	if _, err := CastWithConvoke(gs, 0, spell, []*Permanent{a, b, c}); err != nil {
		t.Fatalf("expected success even with over-tap; got %v", err)
	}
	if v, _ := gs.Stack[0].CostMeta["convoke_reduction"].(int); v != 1 {
		t.Errorf("convoke_reduction should cap at CMC; got %d, want 1", v)
	}
	if v, _ := gs.Stack[0].CostMeta["convoke_net_cost"].(int); v != 0 {
		t.Errorf("convoke_net_cost should be 0; got %d", v)
	}
	// All three creatures are still tapped — over-tap is the caller's
	// choice, the engine doesn't refuse it.
	for _, p := range []*Permanent{a, b, c} {
		if !p.Tapped {
			t.Errorf("creature %s should be tapped even when over-tapping", p.Card.Name)
		}
	}
}

// Bonus: non-convoke card is rejected.
func TestCastWithConvoke_RejectNonConvokeCard(t *testing.T) {
	gs := cv_makeGame(t)
	plain := &Card{
		Name:  "Lightning Bolt",
		Types: []string{"instant"},
		CMC:   1,
		AST:   &gameast.CardAST{Name: "Lightning Bolt"},
	}
	gs.Seats[0].Hand = []*Card{plain}

	a := cv_makeCreature(0, "A", "G")
	cv_putOnBF(gs, 0, a)

	_, err := CastWithConvoke(gs, 0, plain, []*Permanent{a})
	if cv_castErrReason(err) != "no_convoke_keyword" {
		t.Fatalf("expected no_convoke_keyword; got %v", err)
	}
	if a.Tapped {
		t.Fatal("validation should reject before tapping")
	}
}

// Bonus: insufficient mana rejected after creature validation.
func TestCastWithConvoke_InsufficientManaRejected(t *testing.T) {
	gs := cv_makeGame(t)
	spell := cv_makeConvokeSpell("Big Spell", 6)
	gs.Seats[0].Hand = []*Card{spell}
	gs.Seats[0].ManaPool = 2 // need 5 after one convoke; only have 2

	a := cv_makeCreature(0, "A", "G")
	cv_putOnBF(gs, 0, a)

	_, err := CastWithConvoke(gs, 0, spell, []*Permanent{a})
	if cv_castErrReason(err) != "insufficient_mana" {
		t.Fatalf("expected insufficient_mana; got %v", err)
	}
	if a.Tapped {
		t.Fatal("no creature should be tapped on mana-rejection")
	}
}

// Bonus: duplicate creature in slice rejected.
func TestCastWithConvoke_RejectDuplicateCreature(t *testing.T) {
	gs := cv_makeGame(t)
	spell := cv_makeConvokeSpell("Chord", 3)
	gs.Seats[0].Hand = []*Card{spell}
	gs.Seats[0].ManaPool = 3

	a := cv_makeCreature(0, "A", "G")
	cv_putOnBF(gs, 0, a)

	_, err := CastWithConvoke(gs, 0, spell, []*Permanent{a, a})
	if cv_castErrReason(err) != "convoke_creature_duplicated" {
		t.Fatalf("expected convoke_creature_duplicated; got %v", err)
	}
}

// Bonus: nil-safety.
func TestCastWithConvoke_NilSafety(t *testing.T) {
	if _, err := CastWithConvoke(nil, 0, nil, nil); cv_castErrReason(err) != "nil_game" {
		t.Fatalf("nil game should be nil_game; got %v", err)
	}
	gs := cv_makeGame(t)
	if _, err := CastWithConvoke(gs, -1, nil, nil); cv_castErrReason(err) != "invalid_seat" {
		t.Fatalf("invalid seat; got %v", err)
	}
	if _, err := CastWithConvoke(gs, 0, nil, nil); cv_castErrReason(err) != "nil_card" {
		t.Fatalf("nil card; got %v", err)
	}
}
