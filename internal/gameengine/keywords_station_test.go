package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Station tests — CR §702.184 (Aetherdrift Spacecraft)
// ---------------------------------------------------------------------------

func newStationGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(42))
	return NewGameState(2, rng, nil)
}

func newSpacecraftCard(name string, threshold int) *Card {
	return &Card{
		Name:     name,
		Types:    []string{"artifact"},
		TypeLine: "Artifact — Spacecraft",
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "station", Args: []any{float64(threshold)}},
			},
		},
	}
}

func newCreatureCard(name string, power, toughness int) *Card {
	return &Card{
		Name:          name,
		Types:         []string{"creature"},
		BasePower:     power,
		BaseToughness: toughness,
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{},
		},
	}
}

// ---------------------------------------------------------------------------
// HasStation / StationThreshold
// ---------------------------------------------------------------------------

func TestHasStation_Detects(t *testing.T) {
	card := newSpacecraftCard("Surveyor's Scope", 5)
	if !HasStation(card) {
		t.Fatal("HasStation should return true for a station spacecraft")
	}
}

func TestHasStation_Negative(t *testing.T) {
	card := newCreatureCard("Grizzly Bears", 2, 2)
	if HasStation(card) {
		t.Fatal("HasStation should return false for a vanilla creature")
	}
}

func TestHasStation_Nil(t *testing.T) {
	if HasStation(nil) {
		t.Fatal("HasStation(nil) must be false")
	}
}

func TestStationThreshold_ParsesArg(t *testing.T) {
	card := newSpacecraftCard("Scrapheap Drifter", 6)
	n, ok := StationThreshold(card)
	if !ok || n != 6 {
		t.Fatalf("StationThreshold = (%d, %v), want (6, true)", n, ok)
	}
}

func TestStationThreshold_Missing(t *testing.T) {
	card := newCreatureCard("Grizzly Bears", 2, 2)
	if _, ok := StationThreshold(card); ok {
		t.Fatal("StationThreshold should return ok=false for non-station card")
	}
}

// ---------------------------------------------------------------------------
// IsSpacecraft
// ---------------------------------------------------------------------------

func TestIsSpacecraft_TypeLineMatch(t *testing.T) {
	card := newSpacecraftCard("Cruiser", 4)
	p := &Permanent{Card: card}
	if !IsSpacecraft(p) {
		t.Fatal("IsSpacecraft should detect 'Artifact — Spacecraft' typeline")
	}
}

func TestIsSpacecraft_Negative(t *testing.T) {
	p := &Permanent{Card: newCreatureCard("Grizzly Bears", 2, 2)}
	if IsSpacecraft(p) {
		t.Fatal("IsSpacecraft must reject a creature")
	}
}

// ---------------------------------------------------------------------------
// StationProgress / IsStationed
// ---------------------------------------------------------------------------

func TestStationProgress_ReadsChargeCounter(t *testing.T) {
	p := &Permanent{
		Card:     newSpacecraftCard("Drifter", 5),
		Counters: map[string]int{"charge": 3},
	}
	if got := StationProgress(p); got != 3 {
		t.Fatalf("StationProgress = %d, want 3", got)
	}
}

func TestIsStationed_BelowThreshold(t *testing.T) {
	p := &Permanent{
		Card:     newSpacecraftCard("Drifter", 5),
		Counters: map[string]int{"charge": 4},
	}
	if IsStationed(p) {
		t.Fatal("IsStationed should be false at 4 charges below threshold 5")
	}
}

func TestIsStationed_AtThreshold(t *testing.T) {
	p := &Permanent{
		Card:     newSpacecraftCard("Drifter", 5),
		Counters: map[string]int{"charge": 5},
	}
	if !IsStationed(p) {
		t.Fatal("IsStationed should be true at threshold")
	}
}

// ---------------------------------------------------------------------------
// ActivateStation — happy path
// ---------------------------------------------------------------------------

func TestActivateStation_TapsAndAddsCounters(t *testing.T) {
	gs := newStationGame(t)
	ship := &Permanent{
		Card:       newSpacecraftCard("Drifter", 5),
		Controller: 0,
	}
	pilot := &Permanent{
		Card:       newCreatureCard("Pilot", 3, 3),
		Controller: 0,
	}
	if err := ActivateStation(gs, 0, ship, pilot); err != nil {
		t.Fatalf("ActivateStation error: %v", err)
	}
	if !pilot.Tapped {
		t.Error("pilot should be tapped after stationing")
	}
	if got := StationProgress(ship); got != 3 {
		t.Errorf("StationProgress = %d, want 3 (pilot power)", got)
	}
	if IsStationed(ship) {
		t.Error("ship should not be stationed yet at 3/5")
	}
}

// ---------------------------------------------------------------------------
// ActivateStation — crosses threshold and fires becomes_stationed
// ---------------------------------------------------------------------------

func TestActivateStation_CrossesThreshold(t *testing.T) {
	gs := newStationGame(t)
	ship := &Permanent{
		Card:       newSpacecraftCard("Drifter", 5),
		Controller: 0,
		Counters:   map[string]int{"charge": 3},
	}
	pilot := &Permanent{
		Card:       newCreatureCard("Big Pilot", 4, 4),
		Controller: 0,
	}
	if err := ActivateStation(gs, 0, ship, pilot); err != nil {
		t.Fatalf("ActivateStation error: %v", err)
	}
	if !IsStationed(ship) {
		t.Errorf("ship should be stationed (counters=%d, threshold=5)", StationProgress(ship))
	}
	// Verify becomes_stationed event was logged exactly once.
	count := 0
	for _, e := range gs.EventLog {
		if e.Kind == "becomes_stationed" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("becomes_stationed event count = %d, want 1", count)
	}
}

// ---------------------------------------------------------------------------
// ActivateStation — rejection paths
// ---------------------------------------------------------------------------

func TestActivateStation_RejectsAlreadyTapped(t *testing.T) {
	gs := newStationGame(t)
	ship := &Permanent{Card: newSpacecraftCard("Drifter", 5), Controller: 0}
	pilot := &Permanent{
		Card:       newCreatureCard("Pilot", 3, 3),
		Controller: 0,
		Tapped:     true,
	}
	if err := ActivateStation(gs, 0, ship, pilot); err == nil {
		t.Fatal("ActivateStation should reject already-tapped contributor")
	}
}

func TestActivateStation_RejectsNonSpacecraft(t *testing.T) {
	gs := newStationGame(t)
	notShip := &Permanent{
		Card:       newCreatureCard("Just A Bear", 2, 2),
		Controller: 0,
	}
	// Force station keyword without spacecraft subtype.
	notShip.Card.AST.Abilities = append(notShip.Card.AST.Abilities,
		&gameast.Keyword{Name: "station", Args: []any{float64(3)}})
	pilot := &Permanent{Card: newCreatureCard("Pilot", 2, 2), Controller: 0}
	if err := ActivateStation(gs, 0, notShip, pilot); err == nil {
		t.Fatal("ActivateStation should reject non-spacecraft host")
	}
}

func TestActivateStation_RejectsWrongType(t *testing.T) {
	gs := newStationGame(t)
	ship := &Permanent{Card: newSpacecraftCard("Drifter", 5), Controller: 0}
	// An enchantment is neither artifact nor creature.
	contributor := &Permanent{
		Card: &Card{
			Name:  "Honor of the Pure",
			Types: []string{"enchantment"},
			AST:   &gameast.CardAST{Name: "Honor of the Pure"},
		},
		Controller: 0,
	}
	if err := ActivateStation(gs, 0, ship, contributor); err == nil {
		t.Fatal("ActivateStation should reject non-artifact non-creature contributor")
	}
}

func TestActivateStation_RejectsControllerMismatch(t *testing.T) {
	gs := newStationGame(t)
	ship := &Permanent{Card: newSpacecraftCard("Drifter", 5), Controller: 0}
	pilot := &Permanent{
		Card:       newCreatureCard("Opponent's Pilot", 3, 3),
		Controller: 1,
	}
	if err := ActivateStation(gs, 0, ship, pilot); err == nil {
		t.Fatal("ActivateStation should reject contributor controlled by another seat")
	}
}
