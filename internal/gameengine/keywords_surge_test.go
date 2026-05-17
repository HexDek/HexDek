package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Surge tests — CR §702.117
// ---------------------------------------------------------------------------

func newSurgeCard(name string, owner, cmc int, surgeArg string) *Card {
	return &Card{
		Name:  name,
		Owner: owner,
		Types: []string{"instant"},
		CMC:   cmc,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "surge", Args: []any{surgeArg}},
			},
		},
	}
}

func newSurgeGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(11))
	return NewGameState(2, rng, nil)
}

// ---------------------------------------------------------------------------
// (a) Surge cast succeeds when an ally spell was already cast this turn
// ---------------------------------------------------------------------------

func TestCastWithSurge_SucceedsAfterPriorSpell(t *testing.T) {
	gs := newSurgeGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	// Simulate that seat 0 has already cast a spell this turn — the
	// CanPaySurge predicate reads seat.Turn.SpellsCast directly.
	gs.Seats[0].Turn.SpellsCast = 1

	card := newSurgeCard("Reckless Bushwhacker", 0, 3, "{R}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	result, err := CastWithSurge(gs, 0, card, 1)
	if err != nil {
		t.Fatalf("CastWithSurge returned error: %v", err)
	}
	if result == nil {
		t.Fatal("CastWithSurge returned nil CostPaymentResult on success")
	}
	if len(gs.Stack) != 1 {
		t.Fatalf("expected 1 stack item after surge cast, got %d", len(gs.Stack))
	}
	// Card removed from hand.
	for _, c := range gs.Seats[0].Hand {
		if c == card {
			t.Fatal("surge-cast card should not still be in hand")
		}
	}
	// Per-turn flag set.
	if !SpellSurgedThisTurn(gs, 0) {
		t.Fatal("SpellSurgedThisTurn(0) should be true after a surge cast")
	}
}

// ---------------------------------------------------------------------------
// (b) Surge rejected when no prior spell was cast
// ---------------------------------------------------------------------------

func TestCastWithSurge_RejectedWithoutPriorSpell(t *testing.T) {
	gs := newSurgeGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	// gs.Seats[0].Turn.SpellsCast == 0 — no prior spell this turn.

	card := newSurgeCard("Reckless Bushwhacker", 0, 3, "{R}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	_, err := CastWithSurge(gs, 0, card, 1)
	if err == nil {
		t.Fatal("CastWithSurge should fail when no prior spell was cast this turn")
	}
	ce, ok := err.(*CastError)
	if !ok || ce.Reason != "surge_not_active" {
		t.Fatalf("expected CastError surge_not_active, got %v", err)
	}
	// State preserved: card still in hand, mana untouched, stack empty.
	found := false
	for _, c := range gs.Seats[0].Hand {
		if c == card {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("card should remain in hand after rejected surge cast")
	}
	if gs.Seats[0].ManaPool != 5 {
		t.Fatalf("mana should be untouched, got %d", gs.Seats[0].ManaPool)
	}
	if len(gs.Stack) != 0 {
		t.Fatalf("stack should be empty after rejected surge cast, got %d items", len(gs.Stack))
	}
	if SpellSurgedThisTurn(gs, 0) {
		t.Fatal("SpellSurgedThisTurn should remain false after a rejected surge cast")
	}
}

// ---------------------------------------------------------------------------
// (c) CostMeta stamped correctly on the stack item
// ---------------------------------------------------------------------------

func TestCastWithSurge_StampsCostMeta(t *testing.T) {
	gs := newSurgeGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	gs.Seats[0].Turn.SpellsCast = 1
	card := newSurgeCard("Reckless Bushwhacker", 0, 3, "{R}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	if _, err := CastWithSurge(gs, 0, card, 1); err != nil {
		t.Fatalf("CastWithSurge failed: %v", err)
	}
	if len(gs.Stack) != 1 {
		t.Fatalf("expected 1 stack item, got %d", len(gs.Stack))
	}
	item := gs.Stack[0]
	if item.Card != card {
		t.Fatal("stack item should reference the surge-cast card")
	}
	if item.Controller != 0 {
		t.Fatalf("Controller = %d, want 0", item.Controller)
	}
	if v, ok := item.CostMeta["surge_cast"]; !ok || v != true {
		t.Fatalf("CostMeta[\"surge_cast\"] = %v, want true", item.CostMeta["surge_cast"])
	}
	if v, ok := item.CostMeta["surge_cost"]; !ok || v != 1 {
		t.Fatalf("CostMeta[\"surge_cost\"] = %v, want 1", item.CostMeta["surge_cost"])
	}
	if !IsSurgeCast(item) {
		t.Fatal("IsSurgeCast should return true for a surge-cast stack item")
	}
	// Surge does not change resolution destination — neither
	// exile_on_resolve nor bought_back should be set.
	if ShouldExileOnResolve(item) {
		t.Fatal("surge cast must not set exile_on_resolve")
	}
	if ShouldReturnToHandOnResolve(item) {
		t.Fatal("surge cast must not set bought_back / return_to_hand")
	}
}

// ---------------------------------------------------------------------------
// (d) Full surge cost paid (alt cost, not the printed mana cost)
// ---------------------------------------------------------------------------

func TestCastWithSurge_PaysSurgeCostNotPrinted(t *testing.T) {
	gs := newSurgeGame(t)
	gs.Active = 0
	// Card has CMC 3 but surge cost 1. We seed only 4 mana so the
	// difference is visible — paying the printed cost would leave 1,
	// paying the surge cost should leave 3.
	gs.Seats[0].ManaPool = 4
	gs.Seats[0].Turn.SpellsCast = 1
	card := newSurgeCard("Reckless Bushwhacker", 0, 3, "{R}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	if _, err := CastWithSurge(gs, 0, card, 1); err != nil {
		t.Fatalf("CastWithSurge failed: %v", err)
	}
	if gs.Seats[0].ManaPool != 3 {
		t.Fatalf("expected 3 mana left (paid surge cost 1 of 4), got %d", gs.Seats[0].ManaPool)
	}
}

func TestCastWithSurge_InsufficientMana(t *testing.T) {
	gs := newSurgeGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 0
	gs.Seats[0].Turn.SpellsCast = 1
	card := newSurgeCard("Reckless Bushwhacker", 0, 3, "{R}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	_, err := CastWithSurge(gs, 0, card, 1)
	if err == nil {
		t.Fatal("CastWithSurge should fail with insufficient mana")
	}
	ce, ok := err.(*CastError)
	if !ok || ce.Reason != "insufficient_mana" {
		t.Fatalf("expected CastError insufficient_mana, got %v", err)
	}
	// Card stays in hand on failure.
	found := false
	for _, c := range gs.Seats[0].Hand {
		if c == card {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("card should remain in hand after failed surge cast")
	}
}

func TestCastWithSurge_NoKeyword(t *testing.T) {
	gs := newSurgeGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	gs.Seats[0].Turn.SpellsCast = 1
	card := &Card{
		Name:  "Plain Instant",
		Owner: 0,
		Types: []string{"instant"},
		AST:   &gameast.CardAST{Name: "Plain Instant"},
	}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)
	_, err := CastWithSurge(gs, 0, card, 1)
	if err == nil {
		t.Fatal("CastWithSurge should fail for a card without the surge keyword")
	}
	ce, ok := err.(*CastError)
	if !ok || ce.Reason != "no_surge_keyword" {
		t.Fatalf("expected CastError no_surge_keyword, got %v", err)
	}
}

func TestCastWithSurge_NotInHand(t *testing.T) {
	gs := newSurgeGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	gs.Seats[0].Turn.SpellsCast = 1
	card := newSurgeCard("Reckless Bushwhacker", 0, 3, "{R}")
	// Card NOT placed in hand.
	_, err := CastWithSurge(gs, 0, card, 1)
	if err == nil {
		t.Fatal("CastWithSurge should fail when card is not in hand")
	}
	ce, ok := err.(*CastError)
	if !ok || ce.Reason != "not_in_hand" {
		t.Fatalf("expected CastError not_in_hand, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// SurgeCost reader fallback — keyword arg parsing was already tested via
// keywords_combat_test.go indirectly; this exercises the wiring under the
// cast helper to make sure -1 routes to SurgeCost.
// ---------------------------------------------------------------------------

func TestCastWithSurge_NegativeCostFallsBackToPrintedSurgeCost(t *testing.T) {
	gs := newSurgeGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	gs.Seats[0].Turn.SpellsCast = 1
	// surge cost printed as numeric 2 → fallback path returns 2.
	card := &Card{
		Name:  "Numeric Surge",
		Owner: 0,
		Types: []string{"sorcery"},
		CMC:   5,
		AST: &gameast.CardAST{
			Name: "Numeric Surge",
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "surge", Args: []any{2}},
			},
		},
	}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	if _, err := CastWithSurge(gs, 0, card, -1); err != nil {
		t.Fatalf("CastWithSurge with -1 sentinel failed: %v", err)
	}
	if gs.Seats[0].ManaPool != 3 {
		t.Fatalf("expected 3 mana left (paid SurgeCost=2 of 5), got %d", gs.Seats[0].ManaPool)
	}
}
