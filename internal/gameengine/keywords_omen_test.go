package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Omen tests — CR §702.176
// ---------------------------------------------------------------------------

func newOmenCard(name string, owner, cmc int, omenArg string) *Card {
	return &Card{
		Name:  name,
		Owner: owner,
		Types: []string{"sorcery"},
		CMC:   cmc,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "omen", Args: []any{omenArg}},
			},
		},
	}
}

func newOmenGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(7))
	return NewGameState(2, rng, nil)
}

// ---------------------------------------------------------------------------
// HasOmen
// ---------------------------------------------------------------------------

func TestHasOmen_Detects(t *testing.T) {
	card := newOmenCard("Augur's Sight", 0, 3, "{1}{U}")
	if !HasOmen(card) {
		t.Fatal("HasOmen returned false for a card with omen")
	}
}

func TestHasOmen_NegativeNoKeyword(t *testing.T) {
	card := &Card{
		Name:  "Lightning Bolt",
		Owner: 0,
		Types: []string{"instant"},
		AST:   &gameast.CardAST{Name: "Lightning Bolt"},
	}
	if HasOmen(card) {
		t.Fatal("HasOmen returned true for a card with no omen keyword")
	}
}

func TestHasOmen_NilCard(t *testing.T) {
	if HasOmen(nil) {
		t.Fatal("HasOmen(nil) should return false")
	}
}

// ---------------------------------------------------------------------------
// OmenCost — mana-string parsing
// ---------------------------------------------------------------------------

func TestOmenCost_ParsesManaString(t *testing.T) {
	card := newOmenCard("Augur's Sight", 0, 3, "{1}{U}")
	if got := OmenCost(card); got != 2 {
		t.Fatalf("OmenCost = %d, want 2 (for {1}{U})", got)
	}
}

func TestOmenCost_NumericArg(t *testing.T) {
	card := &Card{
		Name:  "Numeric Omen",
		Owner: 0,
		Types: []string{"sorcery"},
		CMC:   4,
		AST: &gameast.CardAST{
			Name: "Numeric Omen",
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "omen", Args: []any{3}},
			},
		},
	}
	if got := OmenCost(card); got != 3 {
		t.Fatalf("OmenCost = %d, want 3 (numeric arg)", got)
	}
}

func TestOmenCost_NoArgsFallsBackToCMC(t *testing.T) {
	card := &Card{
		Name:  "Bare Omen",
		Owner: 0,
		Types: []string{"sorcery"},
		CMC:   4,
		AST: &gameast.CardAST{
			Name: "Bare Omen",
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "omen"},
			},
		},
	}
	if got := OmenCost(card); got != 4 {
		t.Fatalf("OmenCost = %d, want 4 (fallback to CMC when no args)", got)
	}
}

// ---------------------------------------------------------------------------
// (a) CastOmen — pays cost, stamps CostMeta, registers exile-cast grant
// ---------------------------------------------------------------------------

func TestCastOmen_PaysAndStampsCostMeta(t *testing.T) {
	gs := newOmenGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	card := newOmenCard("Augur's Sight", 0, 3, "{1}{U}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	result, err := CastOmen(gs, 0, card, 2)
	if err != nil {
		t.Fatalf("CastOmen returned error: %v", err)
	}
	if result == nil {
		t.Fatal("CastOmen returned nil CostPaymentResult on success")
	}

	// Mana paid (omen cost was 2; pool was 5; expect 3 left).
	if gs.Seats[0].ManaPool != 3 {
		t.Fatalf("expected 3 mana left, got %d", gs.Seats[0].ManaPool)
	}
	// Card removed from hand.
	for _, c := range gs.Seats[0].Hand {
		if c == card {
			t.Fatal("omen-cast card should not be in hand any more")
		}
	}
	// Stack item present with the expected CostMeta stamps.
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
	if v, ok := item.CostMeta["omen"]; !ok || v != true {
		t.Fatalf("CostMeta[\"omen\"] = %v, want true", item.CostMeta["omen"])
	}
	if v, ok := item.CostMeta["omen_cost"]; !ok || v != 2 {
		t.Fatalf("CostMeta[\"omen_cost\"] = %v, want 2", item.CostMeta["omen_cost"])
	}
	if v, ok := item.CostMeta["zone_cast_keyword"]; !ok || v != "omen" {
		t.Fatalf("CostMeta[\"zone_cast_keyword\"] = %v, want \"omen\"", item.CostMeta["zone_cast_keyword"])
	}
	if !ShouldExileOnResolve(item) {
		t.Fatal("omen cast should stamp exile_on_resolve=true so stack.go routes to exile")
	}

	// Seat-level "spell omen-cast this turn" flag set.
	if !SpellOmenCastThisTurn(gs, 0) {
		t.Fatal("SpellOmenCastThisTurn(0) returned false after an omen cast")
	}

	// The cast-from-exile grant is registered at cast time and keyed on
	// the card pointer so it survives the stack→exile move on resolve.
	grant := GetZoneCastGrant(gs, card)
	if grant == nil {
		t.Fatal("CastOmen should register a ZoneCastPermission for the omen face")
	}
	if grant.Zone != ZoneExile {
		t.Fatalf("grant.Zone = %q, want %q", grant.Zone, ZoneExile)
	}
	if grant.Keyword != "omen" {
		t.Fatalf("grant.Keyword = %q, want \"omen\"", grant.Keyword)
	}
	if grant.RequireController != 0 {
		t.Fatalf("grant.RequireController = %d, want 0 (card.Owner)", grant.RequireController)
	}
	if grant.ManaCost != -1 {
		t.Fatalf("grant.ManaCost = %d, want -1 (use card's printed mana cost)", grant.ManaCost)
	}
}

// ---------------------------------------------------------------------------
// CastOmen — failure paths
// ---------------------------------------------------------------------------

func TestCastOmen_NoKeyword(t *testing.T) {
	gs := newOmenGame(t)
	gs.Seats[0].ManaPool = 5
	card := &Card{
		Name:  "Plain Sorcery",
		Owner: 0,
		Types: []string{"sorcery"},
		AST:   &gameast.CardAST{Name: "Plain Sorcery"},
	}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)
	if _, err := CastOmen(gs, 0, card, 2); err == nil {
		t.Fatal("CastOmen should fail for a card without the omen keyword")
	}
	if gs.Seats[0].ManaPool != 5 {
		t.Fatalf("mana should be unchanged on failure, got %d", gs.Seats[0].ManaPool)
	}
}

func TestCastOmen_NotInHand(t *testing.T) {
	gs := newOmenGame(t)
	gs.Seats[0].ManaPool = 5
	card := newOmenCard("Augur's Sight", 0, 3, "{1}{U}")
	// Card NOT placed in hand.
	if _, err := CastOmen(gs, 0, card, 2); err == nil {
		t.Fatal("CastOmen should fail when card is not in hand")
	}
	if gs.Seats[0].ManaPool != 5 {
		t.Fatalf("mana should be unchanged on failure, got %d", gs.Seats[0].ManaPool)
	}
}

func TestCastOmen_InsufficientMana(t *testing.T) {
	gs := newOmenGame(t)
	gs.Seats[0].ManaPool = 1
	card := newOmenCard("Augur's Sight", 0, 3, "{1}{U}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)
	if _, err := CastOmen(gs, 0, card, 2); err == nil {
		t.Fatal("CastOmen should fail with insufficient mana")
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
		t.Fatal("card should remain in hand after a failed omen cast")
	}
}

// ---------------------------------------------------------------------------
// (b) Resolve goes to exile, not graveyard
// ---------------------------------------------------------------------------

func TestOmen_ResolveExilesInsteadOfGraveyard(t *testing.T) {
	gs := newOmenGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	card := newOmenCard("Augur's Sight", 0, 3, "{1}{U}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	if _, err := CastOmen(gs, 0, card, 2); err != nil {
		t.Fatalf("CastOmen failed: %v", err)
	}

	// Resolve — stack.go's ShouldExileOnResolve branch routes to exile.
	ResolveStackTop(gs)

	// Must NOT be in graveyard (the §702.176 "exile it" clause replaces
	// the default §608.2g graveyard destination).
	for _, c := range gs.Seats[0].Graveyard {
		if c == card {
			t.Fatal("omen card should be in exile after resolution, not graveyard (§702.176)")
		}
	}
	// Must be in exile (its owner's exile, per the implicit owner
	// destination of "exile it").
	foundInExile := false
	for _, c := range gs.Seats[0].Exile {
		if c == card {
			foundInExile = true
			break
		}
	}
	if !foundInExile {
		t.Fatal("omen card should be in owner's exile after resolution")
	}

	// The cast-from-exile grant must still be present — it activates
	// only now that the card has landed in exile.
	if grant := GetZoneCastGrant(gs, card); grant == nil {
		t.Fatal("post-resolve omen grant should still be registered")
	}
}

// ---------------------------------------------------------------------------
// (c) Omen face castable from exile via the registered ZoneCastPermission
// ---------------------------------------------------------------------------

func TestOmen_Face2_CastableFromExileViaGrant(t *testing.T) {
	gs := newOmenGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	// CMC 3 — the second cast uses the printed mana cost via ManaCost=-1.
	card := newOmenCard("Augur's Sight", 0, 3, "{1}{U}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	// First cast — pay omen cost (2), card resolves to exile.
	if _, err := CastOmen(gs, 0, card, 2); err != nil {
		t.Fatalf("CastOmen failed: %v", err)
	}
	ResolveStackTop(gs)

	// Sanity: card in exile.
	inExile := false
	for _, c := range gs.Seats[0].Exile {
		if c == card {
			inExile = true
			break
		}
	}
	if !inExile {
		t.Fatal("setup: omen card should be in exile after resolution")
	}

	// Refill mana so the second cast can pay the printed mana cost (CMC 3).
	gs.Seats[0].ManaPool = 3
	gs.Stack = gs.Stack[:0]

	// CanCastFromZone should return the omen grant for ZoneExile.
	grant := GetZoneCastGrant(gs, card)
	if grant == nil {
		t.Fatal("ZoneCastPermission for omen should be present pre-cast")
	}
	perms := []*ZoneCastPermission{grant}
	got := CanCastFromZone(gs, 0, card, ZoneExile, perms)
	if got == nil {
		t.Fatal("CanCastFromZone should return the omen grant for ZoneExile")
	}
	if got.Keyword != "omen" {
		t.Fatalf("matched grant.Keyword = %q, want \"omen\"", got.Keyword)
	}

	// Drive the second cast through CastFromZone. CastFromZone runs the
	// full cast pipeline (cast triggers, push, priority, DrainStack), so
	// the spell will resolve before this call returns.
	if _, err := CastFromZone(gs, 0, card, ZoneExile, grant, nil); err != nil {
		t.Fatalf("CastFromZone for omen face2 failed: %v", err)
	}
	// Mana paid was the card's CMC (3) since ManaCost=-1 instructs
	// CastFromZone to use the printed mana cost.
	if gs.Seats[0].ManaPool != 0 {
		t.Fatalf("expected 0 mana left after paying CMC=3, got %d", gs.Seats[0].ManaPool)
	}
	// Card removed from exile.
	for _, c := range gs.Seats[0].Exile {
		if c == card {
			t.Fatal("card should no longer be in exile after the omen-face cast")
		}
	}
	// The second cast must NOT exile-on-resolve: the §702.176 omen
	// replacement applies only to the omen-cost cast, not the later
	// from-exile cast. After DrainStack the omen face has resolved
	// and the card has gone to its owner's graveyard (CR §608.2g).
	inGraveyard := false
	for _, c := range gs.Seats[0].Graveyard {
		if c == card {
			inGraveyard = true
			break
		}
	}
	if !inGraveyard {
		t.Fatal("face2 cast should resolve to graveyard, not exile (omen replacement does not re-apply to the from-exile cast)")
	}
}

// ---------------------------------------------------------------------------
// NewOmenCastFromExilePermission — shape sanity
// ---------------------------------------------------------------------------

func TestNewOmenCastFromExilePermission_Shape(t *testing.T) {
	p := NewOmenCastFromExilePermission(2)
	if p == nil {
		t.Fatal("NewOmenCastFromExilePermission returned nil")
	}
	if p.Zone != ZoneExile {
		t.Fatalf("Zone = %q, want %q", p.Zone, ZoneExile)
	}
	if p.Keyword != "omen" {
		t.Fatalf("Keyword = %q, want \"omen\"", p.Keyword)
	}
	if p.ManaCost != -1 {
		t.Fatalf("ManaCost = %d, want -1", p.ManaCost)
	}
	if p.RequireController != 2 {
		t.Fatalf("RequireController = %d, want 2 (owner)", p.RequireController)
	}
	if p.Duration != "" {
		t.Fatalf("Duration = %q, want \"\" (permanent until cast)", p.Duration)
	}
	if p.ExileOnResolve {
		t.Fatal("face2 grant should NOT set ExileOnResolve — second cast resolves normally")
	}
}
