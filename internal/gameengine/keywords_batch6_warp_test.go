package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Warp tests — CR §702.185
// ---------------------------------------------------------------------------

func newWarpCard(name string, owner, cmc int) *Card {
	return &Card{
		Name:  name,
		Owner: owner,
		Types: []string{"creature"},
		CMC:   cmc,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "warp", Args: []any{"{1}{u}"}},
			},
		},
	}
}

func newWarpGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(42))
	return NewGameState(2, rng, nil)
}

// ---------------------------------------------------------------------------
// HasWarp
// ---------------------------------------------------------------------------

func TestHasWarp_Detects(t *testing.T) {
	card := newWarpCard("Starbreach Whale", 0, 5)
	if !HasWarp(card) {
		t.Fatal("HasWarp returned false for a card with warp keyword")
	}
}

func TestHasWarp_NegativeNoKeyword(t *testing.T) {
	card := &Card{
		Name:  "Plain Bear",
		Owner: 0,
		Types: []string{"creature"},
		AST: &gameast.CardAST{
			Name:      "Plain Bear",
			Abilities: []gameast.Ability{},
		},
	}
	if HasWarp(card) {
		t.Fatal("HasWarp returned true for a card with no warp keyword")
	}
}

func TestHasWarp_NilCard(t *testing.T) {
	if HasWarp(nil) {
		t.Fatal("HasWarp(nil) should return false")
	}
}

// ---------------------------------------------------------------------------
// CastWarp — happy path
// ---------------------------------------------------------------------------

func TestCastWarp_PaysCostAndPushesStack(t *testing.T) {
	gs := newWarpGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5

	card := newWarpCard("Starbreach Whale", 0, 5)
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	result, err := CastWarp(gs, 0, card, 2)
	if err != nil {
		t.Fatalf("CastWarp returned error: %v", err)
	}
	if result == nil {
		t.Fatal("CastWarp returned nil CostPaymentResult on success")
	}

	// Mana paid (warp cost was 2; pool was 5; expect 3 left).
	if gs.Seats[0].ManaPool != 3 {
		t.Fatalf("expected 3 mana left, got %d", gs.Seats[0].ManaPool)
	}
	// Card removed from hand.
	for _, c := range gs.Seats[0].Hand {
		if c == card {
			t.Fatal("warped card should no longer be in hand")
		}
	}
	// Stack item present with warped CostMeta.
	if len(gs.Stack) != 1 {
		t.Fatalf("expected 1 stack item, got %d", len(gs.Stack))
	}
	item := gs.Stack[0]
	if item.Card != card {
		t.Fatal("stack item Card should reference the cast card")
	}
	if item.Controller != 0 {
		t.Fatalf("stack item Controller=%d, want 0", item.Controller)
	}
	if v, ok := item.CostMeta["warped"]; !ok || v != true {
		t.Fatalf("stack item CostMeta[\"warped\"] = %v, want true", item.CostMeta["warped"])
	}
	if v, ok := item.CostMeta["warp_cost"]; !ok || v != 2 {
		t.Fatalf("stack item CostMeta[\"warp_cost\"] = %v, want 2", item.CostMeta["warp_cost"])
	}
	// Seat-level "spell was warped this turn" flag set.
	if !SpellWarpedThisTurn(gs, 0) {
		t.Fatal("SpellWarpedThisTurn(0) returned false after a warp cast")
	}
}

// ---------------------------------------------------------------------------
// CastWarp — failure paths
// ---------------------------------------------------------------------------

func TestCastWarp_NoKeyword(t *testing.T) {
	gs := newWarpGame(t)
	gs.Seats[0].ManaPool = 5
	card := &Card{Name: "Plain Bear", Owner: 0, Types: []string{"creature"},
		AST: &gameast.CardAST{Name: "Plain Bear", Abilities: []gameast.Ability{}}}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)
	if _, err := CastWarp(gs, 0, card, 2); err == nil {
		t.Fatal("CastWarp should fail for a card without warp")
	}
	// Card stays in hand, mana untouched.
	if gs.Seats[0].ManaPool != 5 {
		t.Fatalf("mana should be unchanged, got %d", gs.Seats[0].ManaPool)
	}
}

func TestCastWarp_NotInHand(t *testing.T) {
	gs := newWarpGame(t)
	gs.Seats[0].ManaPool = 5
	card := newWarpCard("Starbreach Whale", 0, 5)
	// Card NOT placed in hand.
	if _, err := CastWarp(gs, 0, card, 2); err == nil {
		t.Fatal("CastWarp should fail when card is not in hand")
	}
	if gs.Seats[0].ManaPool != 5 {
		t.Fatalf("mana should be unchanged, got %d", gs.Seats[0].ManaPool)
	}
}

func TestCastWarp_InsufficientMana(t *testing.T) {
	gs := newWarpGame(t)
	gs.Seats[0].ManaPool = 1
	card := newWarpCard("Starbreach Whale", 0, 5)
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)
	if _, err := CastWarp(gs, 0, card, 2); err == nil {
		t.Fatal("CastWarp should fail with insufficient mana")
	}
	// Card still in hand.
	found := false
	for _, c := range gs.Seats[0].Hand {
		if c == card {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("card should remain in hand after failed CastWarp")
	}
}

func TestCastWarp_InvalidSeat(t *testing.T) {
	gs := newWarpGame(t)
	card := newWarpCard("Starbreach Whale", 0, 5)
	if _, err := CastWarp(gs, 99, card, 2); err == nil {
		t.Fatal("CastWarp should fail on invalid seat index")
	}
}

// ---------------------------------------------------------------------------
// Cast-from-exile permission
// ---------------------------------------------------------------------------

func TestNewWarpCastFromExilePermission(t *testing.T) {
	perm := NewWarpCastFromExilePermission(2)
	if perm == nil {
		t.Fatal("permission should not be nil")
	}
	if perm.Zone != ZoneExile {
		t.Fatalf("permission.Zone = %q, want %q", perm.Zone, ZoneExile)
	}
	if perm.Keyword != "warp" {
		t.Fatalf("permission.Keyword = %q, want \"warp\"", perm.Keyword)
	}
	if perm.ManaCost != -1 {
		t.Fatalf("permission.ManaCost = %d, want -1 (use printed cost)", perm.ManaCost)
	}
	if perm.RequireController != 2 {
		t.Fatalf("permission.RequireController = %d, want 2", perm.RequireController)
	}
	if perm.ExileOnResolve {
		t.Fatal("permission.ExileOnResolve should be false (card was already in exile and now becomes a normal permanent)")
	}
	if perm.Duration != "" {
		t.Fatalf("permission.Duration = %q, want \"\" (permanent until cast)", perm.Duration)
	}
}

// ---------------------------------------------------------------------------
// Delayed exile trigger
// ---------------------------------------------------------------------------

func TestRegisterWarpExileTrigger_AppendsDelayedTrigger(t *testing.T) {
	gs := newWarpGame(t)
	card := newWarpCard("Starbreach Whale", 0, 5)
	perm := &Permanent{
		Card:       card,
		Owner:      0,
		Controller: 0,
		Timestamp:  gs.NextTimestamp(),
		Flags:      map[string]int{},
	}
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, perm)

	before := len(gs.DelayedTriggers)
	RegisterWarpExileTrigger(gs, perm)
	if len(gs.DelayedTriggers) != before+1 {
		t.Fatalf("expected 1 new delayed trigger, got %d (delta)", len(gs.DelayedTriggers)-before)
	}
	dt := gs.DelayedTriggers[len(gs.DelayedTriggers)-1]
	if dt.TriggerAt != "end_of_turn" {
		t.Fatalf("trigger.TriggerAt = %q, want \"end_of_turn\"", dt.TriggerAt)
	}
	if !dt.OneShot {
		t.Fatal("warp trigger should be OneShot=true")
	}
	if dt.SourceCardName != "Starbreach Whale" {
		t.Fatalf("trigger.SourceCardName = %q, want \"Starbreach Whale\"", dt.SourceCardName)
	}
	if dt.EffectFn == nil {
		t.Fatal("trigger.EffectFn should not be nil")
	}
}

func TestWarpExileTrigger_FiresAtEndStepAndExiles(t *testing.T) {
	gs := newWarpGame(t)
	card := newWarpCard("Starbreach Whale", 0, 5)
	perm := &Permanent{
		Card:       card,
		Owner:      0,
		Controller: 0,
		Timestamp:  gs.NextTimestamp(),
		Flags:      map[string]int{},
	}
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, perm)

	RegisterWarpExileTrigger(gs, perm)

	// Fire end-step delayed triggers.
	fired := FireDelayedTriggers(gs, "ending", "end")
	if fired != 1 {
		t.Fatalf("expected 1 trigger to fire, got %d", fired)
	}
	// Permanent removed from battlefield.
	for _, p := range gs.Seats[0].Battlefield {
		if p == perm {
			t.Fatal("permanent should be removed from battlefield after warp exile")
		}
	}
	// Card in owner's exile.
	foundInExile := false
	for _, c := range gs.Seats[0].Exile {
		if c == card {
			foundInExile = true
			break
		}
	}
	if !foundInExile {
		t.Fatal("card should be in owner's exile after warp delayed trigger fires")
	}
	// Cast-from-exile permission granted.
	if gs.ZoneCastGrants == nil {
		t.Fatal("ZoneCastGrants should be populated")
	}
	grant := gs.ZoneCastGrants[card]
	if grant == nil {
		t.Fatal("expected a ZoneCastPermission for the warp-exiled card")
	}
	if grant.Zone != ZoneExile {
		t.Fatalf("grant.Zone = %q, want %q", grant.Zone, ZoneExile)
	}
	if grant.RequireController != 0 {
		t.Fatalf("grant.RequireController = %d, want 0", grant.RequireController)
	}
	if grant.ManaCost != -1 {
		t.Fatalf("grant.ManaCost = %d, want -1 (printed cost)", grant.ManaCost)
	}
}

func TestWarpExileTrigger_PermanentAlreadyGone(t *testing.T) {
	gs := newWarpGame(t)
	card := newWarpCard("Starbreach Whale", 0, 5)
	perm := &Permanent{
		Card:       card,
		Owner:      0,
		Controller: 0,
		Timestamp:  gs.NextTimestamp(),
		Flags:      map[string]int{},
	}
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, perm)

	RegisterWarpExileTrigger(gs, perm)

	// Simulate the permanent being destroyed/bounced before end step.
	gs.Seats[0].Battlefield = nil

	// Trigger should still fire but be a no-op (logs warp_exile_skipped).
	fired := FireDelayedTriggers(gs, "ending", "end")
	if fired != 1 {
		t.Fatalf("expected trigger to still fire (got %d)", fired)
	}
	// No card added to exile.
	if len(gs.Seats[0].Exile) != 0 {
		t.Fatal("exile should remain empty when permanent has already left battlefield")
	}
	// No cast-from-exile grant.
	if gs.ZoneCastGrants != nil && gs.ZoneCastGrants[card] != nil {
		t.Fatal("no cast-from-exile permission should be granted when permanent is gone")
	}
}

// ---------------------------------------------------------------------------
// End-to-end: warp cast → resolve → end step → exile → cast-from-exile
// ---------------------------------------------------------------------------

func TestWarp_CastResolveEndStepExile_EndToEnd(t *testing.T) {
	gs := newWarpGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	card := newWarpCard("Starbreach Whale", 0, 5)
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	// 1. Cast for warp cost.
	if _, err := CastWarp(gs, 0, card, 2); err != nil {
		t.Fatalf("CastWarp failed: %v", err)
	}

	// 2. Resolve the spell — should enter the battlefield and register
	//    the warp delayed trigger via the stack.go hook.
	ResolveStackTop(gs)

	// Permanent on battlefield.
	var ourPerm *Permanent
	for _, p := range gs.Seats[0].Battlefield {
		if p != nil && p.Card == card {
			ourPerm = p
			break
		}
	}
	if ourPerm == nil {
		t.Fatal("warped permanent should be on battlefield after resolve")
	}

	// Delayed trigger registered.
	foundDT := false
	for _, dt := range gs.DelayedTriggers {
		if dt != nil && dt.SourceCardName == "Starbreach Whale" && dt.TriggerAt == "end_of_turn" {
			foundDT = true
			break
		}
	}
	if !foundDT {
		t.Fatal("expected warp delayed trigger to be registered after resolve")
	}

	// 3. Fire end step.
	FireDelayedTriggers(gs, "ending", "end")

	// Permanent gone, card in exile, grant present.
	for _, p := range gs.Seats[0].Battlefield {
		if p == ourPerm {
			t.Fatal("permanent should have been exiled at end of turn")
		}
	}
	foundInExile := false
	for _, c := range gs.Seats[0].Exile {
		if c == card {
			foundInExile = true
			break
		}
	}
	if !foundInExile {
		t.Fatal("card should be in exile after warp end-step trigger fires")
	}
	if gs.ZoneCastGrants[card] == nil {
		t.Fatal("cast-from-exile permission should be granted after warp exile")
	}
}

// ---------------------------------------------------------------------------
// Non-warp cast leaves the spell as a normal permanent (no delayed trigger)
// ---------------------------------------------------------------------------

func TestWarp_NormalCastDoesNotRegisterTrigger(t *testing.T) {
	gs := newWarpGame(t)
	gs.Active = 0
	card := newWarpCard("Starbreach Whale", 0, 5)
	// Push a non-warp stack item (no CostMeta["warped"]).
	gs.Stack = append(gs.Stack, &StackItem{
		Card:       card,
		Controller: 0,
		CastZone:   ZoneHand,
		// no CostMeta — this represents a normal hand cast.
	})

	ResolveStackTop(gs)

	for _, dt := range gs.DelayedTriggers {
		if dt != nil && dt.SourceCardName == "Starbreach Whale" {
			t.Fatal("normal cast should NOT register a warp delayed trigger")
		}
	}
}

// ---------------------------------------------------------------------------
// SpellWarpedThisTurn
// ---------------------------------------------------------------------------

func TestSpellWarpedThisTurn_FalseByDefault(t *testing.T) {
	gs := newWarpGame(t)
	if SpellWarpedThisTurn(gs, 0) {
		t.Fatal("SpellWarpedThisTurn should be false on a fresh game")
	}
}

func TestSpellWarpedThisTurn_TrueAfterCast(t *testing.T) {
	gs := newWarpGame(t)
	gs.Seats[0].ManaPool = 3
	card := newWarpCard("Starbreach Whale", 0, 5)
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)
	if _, err := CastWarp(gs, 0, card, 2); err != nil {
		t.Fatalf("CastWarp failed: %v", err)
	}
	if !SpellWarpedThisTurn(gs, 0) {
		t.Fatal("SpellWarpedThisTurn(0) should be true after a warp cast")
	}
	// Should be false for the other seat.
	if SpellWarpedThisTurn(gs, 1) {
		t.Fatal("SpellWarpedThisTurn(1) should be false — seat 1 didn't warp anything")
	}
}
