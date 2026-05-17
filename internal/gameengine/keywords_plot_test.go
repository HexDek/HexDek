package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Plot tests — CR §702.172
// ---------------------------------------------------------------------------

func newPlotCard(name string, owner, cmc int, plotArg string) *Card {
	return &Card{
		Name:  name,
		Owner: owner,
		Types: []string{"sorcery"},
		CMC:   cmc,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "plot", Args: []any{plotArg}},
			},
		},
	}
}

func newPlotGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(13))
	gs := NewGameState(2, rng, nil)
	gs.Active = 0
	gs.Step = "precombat_main"
	return gs
}

// ---------------------------------------------------------------------------
// HasPlot / PlotCost — quick keyword-reader sanity
// ---------------------------------------------------------------------------

func TestHasPlot_Detects(t *testing.T) {
	card := newPlotCard("Slickshot Show-Off", 0, 3, "{2}")
	if !HasPlot(card) {
		t.Fatal("HasPlot returned false for a card with plot")
	}
}

func TestPlotCost_ParsesArg(t *testing.T) {
	card := newPlotCard("Slickshot Show-Off", 0, 3, "{2}")
	if got := PlotCost(card); got != 2 {
		t.Fatalf("PlotCost = %d, want 2 (from {2})", got)
	}
}

// ---------------------------------------------------------------------------
// (a) ActivatePlot pays the plot cost and exiles the card
// ---------------------------------------------------------------------------

func TestActivatePlot_PaysCostAndExiles(t *testing.T) {
	gs := newPlotGame(t)
	gs.Turn = 3
	gs.Seats[0].ManaPool = 5
	card := newPlotCard("Slickshot Show-Off", 0, 3, "{2}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	result, err := ActivatePlot(gs, 0, card, 2)
	if err != nil {
		t.Fatalf("ActivatePlot returned error: %v", err)
	}
	if result == nil {
		t.Fatal("ActivatePlot returned nil CostPaymentResult on success")
	}
	if gs.Seats[0].ManaPool != 3 {
		t.Fatalf("mana left = %d, want 3 (paid plot 2 of 5)", gs.Seats[0].ManaPool)
	}
	// Card moved hand → exile.
	for _, c := range gs.Seats[0].Hand {
		if c == card {
			t.Fatal("plotted card should not still be in hand")
		}
	}
	inExile := false
	for _, c := range gs.Seats[0].Exile {
		if c == card {
			inExile = true
			break
		}
	}
	if !inExile {
		t.Fatal("plotted card should be in owner's exile")
	}
	// PlotMeta stamped.
	meta := GetPlotMeta(gs, card)
	if meta == nil {
		t.Fatal("expected PlotMeta to be stamped after ActivatePlot")
	}
	if meta.Seat != 0 {
		t.Fatalf("meta.Seat = %d, want 0", meta.Seat)
	}
	if meta.Turn != 3 {
		t.Fatalf("meta.Turn = %d, want 3 (current turn at activation)", meta.Turn)
	}
	if meta.ExiledAt != 3 {
		t.Fatalf("meta.ExiledAt = %d, want 3", meta.ExiledAt)
	}
}

// ---------------------------------------------------------------------------
// (e) ZoneCastPermission registered with Zone=exile, ManaCost=0
// ---------------------------------------------------------------------------

func TestActivatePlot_RegistersZoneCastPermission(t *testing.T) {
	gs := newPlotGame(t)
	gs.Turn = 4
	gs.Seats[0].ManaPool = 5
	card := newPlotCard("Slickshot Show-Off", 0, 3, "{2}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	if _, err := ActivatePlot(gs, 0, card, 2); err != nil {
		t.Fatalf("ActivatePlot failed: %v", err)
	}

	grant := GetZoneCastGrant(gs, card)
	if grant == nil {
		t.Fatal("expected a ZoneCastPermission registered for the plotted card")
	}
	if grant.Zone != ZoneExile {
		t.Fatalf("grant.Zone = %q, want %q", grant.Zone, ZoneExile)
	}
	if grant.Keyword != "plot" {
		t.Fatalf("grant.Keyword = %q, want \"plot\"", grant.Keyword)
	}
	if grant.ManaCost != 0 {
		t.Fatalf("grant.ManaCost = %d, want 0 (plot cast is free per §702.172b)", grant.ManaCost)
	}
	if grant.RequireController != 0 {
		t.Fatalf("grant.RequireController = %d, want 0 (owner)", grant.RequireController)
	}
	if grant.GrantTurn != 4 {
		t.Fatalf("grant.GrantTurn = %d, want 4 (matches activation turn)", grant.GrantTurn)
	}
}

// ---------------------------------------------------------------------------
// (b) Cannot cast same turn — sorcery speed + turn gate
// ---------------------------------------------------------------------------

func TestCastPlot_RejectedSameTurnAsActivation(t *testing.T) {
	gs := newPlotGame(t)
	gs.Turn = 3
	gs.Seats[0].ManaPool = 5
	card := newPlotCard("Slickshot Show-Off", 0, 3, "{2}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	if _, err := ActivatePlot(gs, 0, card, 2); err != nil {
		t.Fatalf("ActivatePlot failed: %v", err)
	}
	// Same turn — gs.Turn still 3 — must be rejected.
	_, err := CastPlot(gs, 0, card)
	if err == nil {
		t.Fatal("CastPlot on the same turn as ActivatePlot should be rejected (§702.172b)")
	}
	ce, ok := err.(*CastError)
	if !ok || ce.Reason != "same_turn_as_plot" {
		t.Fatalf("expected CastError same_turn_as_plot, got %v", err)
	}
	// Card stays in exile, grant + meta intact.
	inExile := false
	for _, c := range gs.Seats[0].Exile {
		if c == card {
			inExile = true
		}
	}
	if !inExile {
		t.Fatal("card should remain in exile after rejected CastPlot")
	}
	if GetPlotMeta(gs, card) == nil {
		t.Fatal("PlotMeta should remain after rejected CastPlot")
	}
}

func TestCastPlot_RejectedAtInstantSpeed(t *testing.T) {
	gs := newPlotGame(t)
	gs.Turn = 3
	gs.Seats[0].ManaPool = 5
	card := newPlotCard("Slickshot Show-Off", 0, 3, "{2}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)
	if _, err := ActivatePlot(gs, 0, card, 2); err != nil {
		t.Fatalf("ActivatePlot failed: %v", err)
	}
	// Advance to a later turn but step out of a main phase.
	gs.Turn = 4
	gs.Step = "combat_declare_attackers"

	_, err := CastPlot(gs, 0, card)
	if err == nil {
		t.Fatal("CastPlot during combat (non-main phase) should be rejected (§702.172b 'as a sorcery')")
	}
	ce, ok := err.(*CastError)
	if !ok || ce.Reason != "sorcery_speed_only" {
		t.Fatalf("expected CastError sorcery_speed_only, got %v", err)
	}
}

func TestActivatePlot_RejectedAtInstantSpeed(t *testing.T) {
	gs := newPlotGame(t)
	gs.Step = "combat_declare_attackers"
	gs.Seats[0].ManaPool = 5
	card := newPlotCard("Slickshot Show-Off", 0, 3, "{2}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	_, err := ActivatePlot(gs, 0, card, 2)
	if err == nil {
		t.Fatal("ActivatePlot off-main-phase should be rejected (§702.172a 'Activate only as a sorcery')")
	}
	ce, ok := err.(*CastError)
	if !ok || ce.Reason != "sorcery_speed_only" {
		t.Fatalf("expected CastError sorcery_speed_only, got %v", err)
	}
	// State preserved.
	if gs.Seats[0].ManaPool != 5 {
		t.Fatalf("mana should be untouched, got %d", gs.Seats[0].ManaPool)
	}
	for _, c := range gs.Seats[0].Exile {
		if c == card {
			t.Fatal("card should NOT be in exile after rejected ActivatePlot")
		}
	}
}

// ---------------------------------------------------------------------------
// (c) Cast on later turn at {0} mana cost  +  (d) CostMeta stamped
// ---------------------------------------------------------------------------

func TestCastPlot_LaterTurnFreeCastAndCostMeta(t *testing.T) {
	gs := newPlotGame(t)
	gs.Turn = 3
	gs.Seats[0].ManaPool = 5
	card := newPlotCard("Slickshot Show-Off", 0, 3, "{2}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	if _, err := ActivatePlot(gs, 0, card, 2); err != nil {
		t.Fatalf("ActivatePlot failed: %v", err)
	}
	manaAfterActivate := gs.Seats[0].ManaPool

	// Advance one turn. Sorcery timing satisfied (active, main phase,
	// stack empty).
	gs.Turn = 4

	_, err := CastPlot(gs, 0, card)
	if err != nil {
		t.Fatalf("CastPlot failed: %v", err)
	}
	// Plot cast must NOT spend mana on the cast itself — §702.172b
	// "rather than its mana cost." The card's CMC was 3 but no mana
	// should be deducted; only the plot activation already spent 2.
	if gs.Seats[0].ManaPool != manaAfterActivate {
		t.Fatalf("plot cast should not deduct mana; pool = %d, want %d",
			gs.Seats[0].ManaPool, manaAfterActivate)
	}
	// Card removed from exile.
	for _, c := range gs.Seats[0].Exile {
		if c == card {
			t.Fatal("card should be off exile after CastPlot")
		}
	}
	// Stack item present with the plot_cast stamps.
	if len(gs.Stack) != 1 {
		t.Fatalf("expected 1 stack item after CastPlot, got %d", len(gs.Stack))
	}
	item := gs.Stack[0]
	if item.Card != card {
		t.Fatal("stack item should reference the plot-cast card")
	}
	if item.Controller != 0 {
		t.Fatalf("Controller = %d, want 0", item.Controller)
	}
	if item.CastZone != ZoneExile {
		t.Fatalf("CastZone = %q, want %q", item.CastZone, ZoneExile)
	}
	if v, ok := item.CostMeta["plot_cast"]; !ok || v != true {
		t.Fatalf("CostMeta[\"plot_cast\"] = %v, want true", item.CostMeta["plot_cast"])
	}
	if v, ok := item.CostMeta["zone_cast_keyword"]; !ok || v != "plot" {
		t.Fatalf("CostMeta[\"zone_cast_keyword\"] = %v, want \"plot\"", item.CostMeta["zone_cast_keyword"])
	}
	if !IsPlotCast(item) {
		t.Fatal("IsPlotCast should return true for the plot-cast stack item")
	}
	if !SpellPlotCastThisTurn(gs, 0) {
		t.Fatal("SpellPlotCastThisTurn(0) should be true after a plot cast")
	}
	// PlotMeta + grant consumed — plot is single-use.
	if GetPlotMeta(gs, card) != nil {
		t.Fatal("PlotMeta should be cleared after CastPlot consumes the plot")
	}
	if GetZoneCastGrant(gs, card) != nil {
		t.Fatal("ZoneCastPermission should be removed after CastPlot consumes the plot")
	}
}

// ---------------------------------------------------------------------------
// (e) ZoneCastPermission usable via the generic zone-cast pipeline
// ---------------------------------------------------------------------------

func TestPlot_ZoneCastPermissionDrivesExileCast(t *testing.T) {
	gs := newPlotGame(t)
	gs.Turn = 3
	gs.Seats[0].ManaPool = 5
	card := newPlotCard("Slickshot Show-Off", 0, 3, "{2}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)
	if _, err := ActivatePlot(gs, 0, card, 2); err != nil {
		t.Fatalf("ActivatePlot failed: %v", err)
	}
	gs.Turn = 4

	grant := GetZoneCastGrant(gs, card)
	if grant == nil {
		t.Fatal("expected plot grant to be present pre-cast")
	}
	perms := []*ZoneCastPermission{grant}

	// CanCastFromZone should match the plot grant for ZoneExile with
	// 0 mana required.
	gs.Seats[0].ManaPool = 0
	matched := CanCastFromZone(gs, 0, card, ZoneExile, perms)
	if matched == nil {
		t.Fatal("CanCastFromZone should match the plot grant for ZoneExile at 0 mana")
	}
	if matched.Keyword != "plot" {
		t.Fatalf("matched.Keyword = %q, want \"plot\"", matched.Keyword)
	}
}

// ---------------------------------------------------------------------------
// IsPlotCastEligible — predicate sanity
// ---------------------------------------------------------------------------

func TestIsPlotCastEligible_TurnGate(t *testing.T) {
	gs := newPlotGame(t)
	gs.Turn = 3
	gs.Seats[0].ManaPool = 5
	card := newPlotCard("Slickshot Show-Off", 0, 3, "{2}")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)
	if _, err := ActivatePlot(gs, 0, card, 2); err != nil {
		t.Fatalf("ActivatePlot failed: %v", err)
	}
	if IsPlotCastEligible(gs, 0, card) {
		t.Fatal("plot should NOT be cast-eligible on the same turn as activation")
	}
	gs.Turn = 4
	if !IsPlotCastEligible(gs, 0, card) {
		t.Fatal("plot SHOULD be cast-eligible on a later turn for the activating seat")
	}
	if IsPlotCastEligible(gs, 1, card) {
		t.Fatal("plot should NOT be cast-eligible for the wrong seat")
	}
}
