package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Test helpers — Spectacle (CR §702.137)
// ---------------------------------------------------------------------------

func newSpectacleGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(19))
	return NewGameState(2, rng, nil)
}

// spectacleHandCard builds an instant with the spectacle keyword printed
// at `spectacleCost`, with `normalCMC` as the regular mana cost. The
// AST carries both the spectacle keyword (with numeric arg) and a
// throwaway "haste" keyword so post-cast inspection can confirm the
// card's printed identity if needed.
func spectacleHandCard(gs *GameState, seat int, name string, normalCMC, spectacleCost int) *Card {
	c := &Card{
		Name:  name,
		Owner: seat,
		CMC:   normalCMC,
		Types: []string{"instant"},
		Colors: []string{"R"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "spectacle", Args: []any{float64(spectacleCost)}},
			},
		},
	}
	gs.Seats[seat].Hand = append(gs.Seats[seat].Hand, c)
	return c
}

// markOpponentLostLife sets the opponent's Turn.LifeLost so
// CanPaySpectacle returns true. Mirrors what LoseLife / combat damage
// would have done earlier in the turn.
func markOpponentLostLife(gs *GameState, oppSeat, amount int) {
	gs.Seats[oppSeat].Turn.LifeLost += amount
}

// ===========================================================================
// (a) Succeeds when any opponent lost life this turn
// ===========================================================================

func TestCastWithSpectacle_SucceedsWhenOpponentLostLife(t *testing.T) {
	gs := newSpectacleGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 2
	card := spectacleHandCard(gs, 0, "Light Up the Stage", 4, 2)
	markOpponentLostLife(gs, 1, 3)

	res, err := CastWithSpectacle(gs, 0, card, 2)
	if err != nil {
		t.Fatalf("CastWithSpectacle: %v", err)
	}
	if res == nil {
		t.Fatal("CastWithSpectacle returned nil result on success")
	}

	// Card removed from hand and placed on the stack.
	if len(gs.Seats[0].Hand) != 0 {
		t.Errorf("card should be removed from hand, hand=%d", len(gs.Seats[0].Hand))
	}
	if len(gs.Stack) != 1 || gs.Stack[0].Card != card {
		t.Fatalf("stack should have the cast card on top, got %d items", len(gs.Stack))
	}

	// Per-turn flag set so cards that key off "if a spectacle spell was
	// cast this turn" can see it.
	if !SpellSpectacleThisTurn(gs, 0) {
		t.Error("SpellSpectacleThisTurn(seat 0) should be true after a spectacle cast")
	}
}

func TestCastWithSpectacle_SucceedsWithNonActiveOpponentLifeLoss(t *testing.T) {
	// 4-player variant — seat 0 casts, seat 2 (not the next opponent) is
	// the one who lost life. CanPaySpectacle iterates ALL opponents, so
	// this should still satisfy the condition.
	rng := rand.New(rand.NewSource(19))
	gs := NewGameState(4, rng, nil)
	gs.Active = 0
	gs.Seats[0].ManaPool = 1
	card := spectacleHandCard(gs, 0, "Skewer the Critics", 3, 1)
	markOpponentLostLife(gs, 2, 1)

	if _, err := CastWithSpectacle(gs, 0, card, 1); err != nil {
		t.Fatalf("CastWithSpectacle should accept any opponent's life loss, got %v", err)
	}
}

func TestCastWithSpectacle_DefaultsToPrintedSpectacleCostWhenNeg1(t *testing.T) {
	gs := newSpectacleGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 2
	card := spectacleHandCard(gs, 0, "Light Up the Stage", 4, 2)
	markOpponentLostLife(gs, 1, 1)

	// Pass -1 → helper should fall back to keywordArgCost("spectacle")=2.
	if _, err := CastWithSpectacle(gs, 0, card, -1); err != nil {
		t.Fatalf("CastWithSpectacle with default cost: %v", err)
	}
	if gs.Seats[0].ManaPool != 0 {
		t.Errorf("expected mana=0 after paying default spectacle cost {2}, got %d", gs.Seats[0].ManaPool)
	}
}

// ===========================================================================
// (b) Rejected when no opponent lost life
// ===========================================================================

func TestCastWithSpectacle_RejectedWhenNoOpponentLifeLoss(t *testing.T) {
	gs := newSpectacleGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	card := spectacleHandCard(gs, 0, "Light Up the Stage", 4, 2)
	// No opponent has lost life — CanPaySpectacle returns false.

	res, err := CastWithSpectacle(gs, 0, card, 2)
	if err == nil || res != nil {
		t.Fatal("CastWithSpectacle should reject when no opponent lost life this turn")
	}
	if ce, ok := err.(*CastError); !ok || ce.Reason != "spectacle_condition_unmet" {
		t.Errorf("expected CastError(spectacle_condition_unmet), got %T %v", err, err)
	}

	// Side effects guarded: card still in hand, mana untouched, stack empty.
	if len(gs.Seats[0].Hand) != 1 || gs.Seats[0].Hand[0] != card {
		t.Errorf("card should remain in hand after rejection, hand=%d", len(gs.Seats[0].Hand))
	}
	if gs.Seats[0].ManaPool != 5 {
		t.Errorf("mana pool should be untouched on rejection, got %d", gs.Seats[0].ManaPool)
	}
	if len(gs.Stack) != 0 {
		t.Errorf("stack should be empty on rejection, got %d", len(gs.Stack))
	}
	if SpellSpectacleThisTurn(gs, 0) {
		t.Error("per-turn flag must not be set when the cast was rejected")
	}
}

func TestCastWithSpectacle_RejectsCardWithoutKeyword(t *testing.T) {
	gs := newSpectacleGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	c := &Card{
		Name: "Plain Lightning", Owner: 0, Types: []string{"instant"},
		AST: &gameast.CardAST{Name: "Plain Lightning"},
	}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, c)
	markOpponentLostLife(gs, 1, 3)

	res, err := CastWithSpectacle(gs, 0, c, 2)
	if err == nil || res != nil {
		t.Fatal("CastWithSpectacle should reject a card without the spectacle keyword")
	}
	if ce, ok := err.(*CastError); !ok || ce.Reason != "no_spectacle_keyword" {
		t.Errorf("expected CastError(no_spectacle_keyword), got %T %v", err, err)
	}
}

func TestCastWithSpectacle_RejectsInsufficientMana(t *testing.T) {
	gs := newSpectacleGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 1
	card := spectacleHandCard(gs, 0, "Light Up the Stage", 4, 2)
	markOpponentLostLife(gs, 1, 3)

	res, err := CastWithSpectacle(gs, 0, card, 2)
	if err == nil || res != nil {
		t.Fatal("CastWithSpectacle should reject when mana < spectacle cost")
	}
	if ce, ok := err.(*CastError); !ok || ce.Reason != "insufficient_mana" {
		t.Errorf("expected CastError(insufficient_mana), got %T %v", err, err)
	}
}

// ===========================================================================
// (c) CostMeta stamped on the stack item
// ===========================================================================

func TestCastWithSpectacle_StampsCostMeta(t *testing.T) {
	gs := newSpectacleGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 2
	card := spectacleHandCard(gs, 0, "Light Up the Stage", 4, 2)
	markOpponentLostLife(gs, 1, 5)

	if _, err := CastWithSpectacle(gs, 0, card, 2); err != nil {
		t.Fatalf("cast: %v", err)
	}

	if len(gs.Stack) != 1 {
		t.Fatalf("stack should have 1 item, got %d", len(gs.Stack))
	}
	item := gs.Stack[0]
	if item.CostMeta == nil {
		t.Fatal("StackItem.CostMeta should not be nil")
	}
	if v, _ := item.CostMeta["spectacle_cast"].(bool); !v {
		t.Errorf("CostMeta[spectacle_cast] = %v, want true", item.CostMeta["spectacle_cast"])
	}
	if v, _ := item.CostMeta["alt_cost"].(string); v != "spectacle" {
		t.Errorf("CostMeta[alt_cost] = %v, want \"spectacle\"", item.CostMeta["alt_cost"])
	}
	if v, _ := item.CostMeta["spectacle_cost"].(int); v != 2 {
		t.Errorf("CostMeta[spectacle_cost] = %v, want 2", item.CostMeta["spectacle_cost"])
	}
	if item.CastZone != ZoneHand {
		t.Errorf("CastZone should be ZoneHand for a hand-cast spectacle, got %v", item.CastZone)
	}

	// Cast trail event with rule 702.137a.
	sawCast := false
	for _, ev := range gs.EventLog {
		if ev.Kind == "spectacle_cast" && ev.Amount == 2 {
			if rule, _ := ev.Details["rule"].(string); rule == "702.137a" {
				sawCast = true
				break
			}
		}
	}
	if !sawCast {
		t.Error("expected spectacle_cast event with rule 702.137a and amount 2")
	}
}

// ===========================================================================
// (d) Spectacle cost actually paid (not the regular mana cost)
// ===========================================================================

func TestCastWithSpectacle_PaysSpectacleCostNotPrintedCMC(t *testing.T) {
	// Printed CMC is 4 ({3}{R}). Spectacle cost is {R} = 1. With only 2
	// mana available, a normal cast would be impossible — but spectacle
	// must succeed and consume only the spectacle amount, leaving 1.
	gs := newSpectacleGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 2
	card := spectacleHandCard(gs, 0, "Light Up the Stage", 4, 1)
	markOpponentLostLife(gs, 1, 2)

	if _, err := CastWithSpectacle(gs, 0, card, 1); err != nil {
		t.Fatalf("CastWithSpectacle: %v", err)
	}

	// Paid 1, not 4 — proves we charged the spectacle cost.
	if gs.Seats[0].ManaPool != 1 {
		t.Fatalf("expected mana=1 (paid {1} spectacle, not {4} CMC), got %d", gs.Seats[0].ManaPool)
	}

	// pay_mana event should record reason=spectacle_cast (not a regular
	// cast), with the spectacle amount.
	sawPay := false
	for _, ev := range gs.EventLog {
		if ev.Kind != "pay_mana" {
			continue
		}
		reason, _ := ev.Details["reason"].(string)
		keyword, _ := ev.Details["keyword"].(string)
		if reason == "spectacle_cast" && keyword == "spectacle" && ev.Amount == 1 {
			sawPay = true
			break
		}
	}
	if !sawPay {
		t.Error("expected a pay_mana event tagged reason=spectacle_cast keyword=spectacle amount=1")
	}

	// And no event records a payment of the printed CMC (would be 4).
	for _, ev := range gs.EventLog {
		if ev.Kind == "pay_mana" && ev.Amount == 4 {
			t.Errorf("should not have paid the printed CMC of 4, found pay_mana event of amount 4")
		}
	}
}

func TestCastWithSpectacle_NilSafety(t *testing.T) {
	if _, err := CastWithSpectacle(nil, 0, nil, 0); err == nil {
		t.Fatal("CastWithSpectacle(nil...) should error")
	}
	gs := newSpectacleGame(t)
	if _, err := CastWithSpectacle(gs, -1, nil, 0); err == nil {
		t.Fatal("CastWithSpectacle with invalid seat should error")
	}
	if _, err := CastWithSpectacle(gs, 0, nil, 0); err == nil {
		t.Fatal("CastWithSpectacle with nil card should error")
	}
}

func TestSpellSpectacleThisTurn_FalseBeforeCast(t *testing.T) {
	gs := newSpectacleGame(t)
	if SpellSpectacleThisTurn(gs, 0) {
		t.Fatal("SpellSpectacleThisTurn should be false before any cast")
	}
}
