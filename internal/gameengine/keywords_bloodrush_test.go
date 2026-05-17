package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Bloodrush tests — CR §702.99 (Gatecrash, 2013)
// ---------------------------------------------------------------------------

func newBloodrushGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(99))
	return NewGameState(2, rng, nil)
}

// newBloodrushCard builds a creature card with a bloodrush keyword.
// abilityWords may be "", "trample", or "trample,first strike" etc.
func newBloodrushCard(name string, manaCost string, p, t int, abilityWords string) *Card {
	args := []any{manaCost, float64(p), float64(t)}
	if abilityWords != "" {
		args = append(args, abilityWords)
	}
	return &Card{
		Name:          name,
		Types:         []string{"creature"},
		BasePower:     3,
		BaseToughness: 3,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "bloodrush", Args: args},
			},
		},
	}
}

func newAttackerOnBattlefield(gs *GameState, seat int, name string, basePower, baseToughness int) *Permanent {
	card := &Card{
		Name:          name,
		Types:         []string{"creature"},
		BasePower:     basePower,
		BaseToughness: baseToughness,
		Owner:         seat,
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{},
		},
	}
	p := &Permanent{Card: card, Controller: seat, Owner: seat}
	setPermFlag(p, flagAttacking, true)
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, p)
	return p
}

func newIdleCreatureOnBattlefield(gs *GameState, seat int, name string) *Permanent {
	card := &Card{
		Name:          name,
		Types:         []string{"creature"},
		BasePower:     2,
		BaseToughness: 2,
		Owner:         seat,
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{},
		},
	}
	p := &Permanent{Card: card, Controller: seat, Owner: seat}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, p)
	return p
}

// ---------------------------------------------------------------------------
// HasBloodrush / BloodrushCost / BloodrushPump
// ---------------------------------------------------------------------------

func TestHasBloodrush_Detects(t *testing.T) {
	card := newBloodrushCard("Slaughterhorn", "{1}{G}", 4, 4, "")
	if !HasBloodrush(card) {
		t.Fatal("HasBloodrush should be true")
	}
}

func TestHasBloodrush_Negative(t *testing.T) {
	card := &Card{
		Name:  "Plain Bear",
		Types: []string{"creature"},
		AST:   &gameast.CardAST{Name: "Plain Bear", Abilities: []gameast.Ability{}},
	}
	if HasBloodrush(card) {
		t.Fatal("HasBloodrush must be false on vanilla creature")
	}
}

func TestHasBloodrush_Nil(t *testing.T) {
	if HasBloodrush(nil) {
		t.Fatal("HasBloodrush(nil) must be false")
	}
}

func TestBloodrushCost_ParsesManaString(t *testing.T) {
	card := newBloodrushCard("Ghor-Clan Rampager", "{1}{R}{G}", 4, 4, "trample")
	if got := BloodrushCost(card); got != 3 {
		t.Fatalf("BloodrushCost = %d, want 3 ({1}{R}{G})", got)
	}
}

func TestBloodrushPump_ReadsPTAndAbility(t *testing.T) {
	card := newBloodrushCard("Ghor-Clan Rampager", "{1}{R}{G}", 4, 4, "trample")
	p, tgh, abs := BloodrushPump(card)
	if p != 4 || tgh != 4 {
		t.Fatalf("BloodrushPump pt = (%d,%d), want (4,4)", p, tgh)
	}
	if len(abs) != 1 || abs[0] != "trample" {
		t.Fatalf("BloodrushPump abilities = %v, want [trample]", abs)
	}
}

func TestBloodrushPump_MultipleAbilities(t *testing.T) {
	card := newBloodrushCard("Test Card", "{R}", 1, 1, "trample,first strike")
	_, _, abs := BloodrushPump(card)
	if len(abs) != 2 || abs[0] != "trample" || abs[1] != "first strike" {
		t.Fatalf("BloodrushPump abilities = %v, want [trample, first strike]", abs)
	}
}

func TestBloodrushPump_NoAbilities(t *testing.T) {
	card := newBloodrushCard("Slaughterhorn", "{1}{G}", 4, 4, "")
	_, _, abs := BloodrushPump(card)
	if len(abs) != 0 {
		t.Fatalf("BloodrushPump abilities = %v, want []", abs)
	}
}

// ---------------------------------------------------------------------------
// (a) ActivateBloodrush succeeds only on attacking target
// ---------------------------------------------------------------------------

func TestActivateBloodrush_RequiresAttackingTarget(t *testing.T) {
	gs := newBloodrushGame(t)
	gs.Seats[0].ManaPool = 5

	source := newBloodrushCard("Slaughterhorn", "{1}{G}", 4, 4, "")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, source)

	// Idle (non-attacking) target — should be rejected.
	idle := newIdleCreatureOnBattlefield(gs, 0, "Bear")
	if err := ActivateBloodrush(gs, 0, source, idle); err == nil {
		t.Fatal("ActivateBloodrush should reject non-attacking target")
	}
	// Card should still be in hand because activation failed.
	stillInHand := false
	for _, c := range gs.Seats[0].Hand {
		if c == source {
			stillInHand = true
		}
	}
	if !stillInHand {
		t.Fatal("source card should remain in hand after rejected activation")
	}
}

func TestActivateBloodrush_HappyPathOnAttacker(t *testing.T) {
	gs := newBloodrushGame(t)
	gs.Seats[0].ManaPool = 5

	source := newBloodrushCard("Slaughterhorn", "{1}{G}", 4, 4, "")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, source)

	attacker := newAttackerOnBattlefield(gs, 0, "Wolf Token", 1, 1)

	if err := ActivateBloodrush(gs, 0, source, attacker); err != nil {
		t.Fatalf("ActivateBloodrush error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// (b) Discards card from hand
// ---------------------------------------------------------------------------

func TestActivateBloodrush_DiscardsSourceFromHand(t *testing.T) {
	gs := newBloodrushGame(t)
	gs.Seats[0].ManaPool = 5

	source := newBloodrushCard("Slaughterhorn", "{1}{G}", 4, 4, "")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, source)
	attacker := newAttackerOnBattlefield(gs, 0, "Wolf Token", 1, 1)

	if err := ActivateBloodrush(gs, 0, source, attacker); err != nil {
		t.Fatalf("ActivateBloodrush error: %v", err)
	}

	for _, c := range gs.Seats[0].Hand {
		if c == source {
			t.Fatal("source card should have been removed from hand")
		}
	}
	inGraveyard := false
	for _, c := range gs.Seats[0].Graveyard {
		if c == source {
			inGraveyard = true
		}
	}
	if !inGraveyard {
		t.Fatal("source card should be in graveyard after bloodrush activation")
	}
}

// ---------------------------------------------------------------------------
// (c) Applies temp pump until end of turn
// ---------------------------------------------------------------------------

func TestActivateBloodrush_AppliesTempPump(t *testing.T) {
	gs := newBloodrushGame(t)
	gs.Seats[0].ManaPool = 5

	source := newBloodrushCard("Slaughterhorn", "{1}{G}", 4, 4, "")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, source)
	attacker := newAttackerOnBattlefield(gs, 0, "Wolf Token", 1, 1)

	preP, preT := attacker.Power(), attacker.Toughness()

	if err := ActivateBloodrush(gs, 0, source, attacker); err != nil {
		t.Fatalf("ActivateBloodrush error: %v", err)
	}

	// temp_power / temp_toughness flag stamps.
	if attacker.Flags["temp_power"] != 4 {
		t.Errorf("temp_power flag = %d, want 4", attacker.Flags["temp_power"])
	}
	if attacker.Flags["temp_toughness"] != 4 {
		t.Errorf("temp_toughness flag = %d, want 4", attacker.Flags["temp_toughness"])
	}
	// Power/Toughness reflect the buff for combat.
	if got := attacker.Power(); got != preP+4 {
		t.Errorf("Power after pump = %d, want %d", got, preP+4)
	}
	if got := attacker.Toughness(); got != preT+4 {
		t.Errorf("Toughness after pump = %d, want %d", got, preT+4)
	}
	// Modification carries until_end_of_turn duration.
	foundEOT := false
	for _, m := range attacker.Modifications {
		if m.Power == 4 && m.Toughness == 4 && m.Duration == "until_end_of_turn" {
			foundEOT = true
		}
	}
	if !foundEOT {
		t.Errorf("expected an EOT Modification(+4/+4); got %+v", attacker.Modifications)
	}
}

func TestActivateBloodrush_TempPumpClearedAtEOT(t *testing.T) {
	gs := newBloodrushGame(t)
	gs.Seats[0].ManaPool = 5

	source := newBloodrushCard("Slaughterhorn", "{1}{G}", 4, 4, "")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, source)
	attacker := newAttackerOnBattlefield(gs, 0, "Wolf Token", 1, 1)
	prePower := attacker.Power()

	if err := ActivateBloodrush(gs, 0, source, attacker); err != nil {
		t.Fatalf("ActivateBloodrush error: %v", err)
	}

	// Drive the EOT cleanup (§514.2 — until-EOT mods clear at cleanup step).
	ScanExpiredDurations(gs, "ending", "cleanup")

	if got := attacker.Power(); got != prePower {
		t.Errorf("Power after EOT cleanup = %d, want %d (pre-pump baseline)", got, prePower)
	}
	hasLingeringMod := false
	for _, m := range attacker.Modifications {
		if m.Duration == "until_end_of_turn" {
			hasLingeringMod = true
		}
	}
	if hasLingeringMod {
		t.Error("expected EOT modification to be cleared by cleanup step")
	}
}

// ---------------------------------------------------------------------------
// (d) Ability words (e.g. trample) layer on target
// ---------------------------------------------------------------------------

func TestActivateBloodrush_GrantsAbilityWord(t *testing.T) {
	gs := newBloodrushGame(t)
	gs.Seats[0].ManaPool = 5

	source := newBloodrushCard("Ghor-Clan Rampager", "{1}{R}{G}", 4, 4, "trample")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, source)
	attacker := newAttackerOnBattlefield(gs, 0, "Wolf Token", 1, 1)

	if attacker.HasKeyword("trample") {
		t.Fatal("attacker should not start with trample")
	}
	if err := ActivateBloodrush(gs, 0, source, attacker); err != nil {
		t.Fatalf("ActivateBloodrush error: %v", err)
	}
	if !attacker.HasKeyword("trample") {
		t.Errorf("attacker should have trample after bloodrush; granted=%v", attacker.GrantedAbilities)
	}
}

func TestActivateBloodrush_MultipleAbilityWordsLayer(t *testing.T) {
	gs := newBloodrushGame(t)
	gs.Seats[0].ManaPool = 5

	source := newBloodrushCard("Wild Hybrid", "{R}", 2, 2, "trample,haste")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, source)
	attacker := newAttackerOnBattlefield(gs, 0, "Wolf Token", 1, 1)

	if err := ActivateBloodrush(gs, 0, source, attacker); err != nil {
		t.Fatalf("ActivateBloodrush error: %v", err)
	}
	if !attacker.HasKeyword("trample") || !attacker.HasKeyword("haste") {
		t.Errorf("expected trample+haste after bloodrush; granted=%v", attacker.GrantedAbilities)
	}
}

// ---------------------------------------------------------------------------
// Event stamping: bloodrush_source carries the source card pointer
// ---------------------------------------------------------------------------

func TestActivateBloodrush_StampsSourceInEventDetails(t *testing.T) {
	gs := newBloodrushGame(t)
	gs.Seats[0].ManaPool = 5

	source := newBloodrushCard("Slaughterhorn", "{1}{G}", 4, 4, "")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, source)
	attacker := newAttackerOnBattlefield(gs, 0, "Wolf Token", 1, 1)

	if err := ActivateBloodrush(gs, 0, source, attacker); err != nil {
		t.Fatalf("ActivateBloodrush error: %v", err)
	}

	foundStamp := false
	for _, e := range gs.EventLog {
		if e.Kind != "bloodrush" {
			continue
		}
		if e.Details == nil {
			continue
		}
		if got, _ := e.Details["bloodrush_source"].(*Card); got == source {
			foundStamp = true
		}
	}
	if !foundStamp {
		t.Errorf("bloodrush event should stamp bloodrush_source = source card")
	}
}

// ---------------------------------------------------------------------------
// Rejection paths
// ---------------------------------------------------------------------------

func TestActivateBloodrush_RejectsCardNotInHand(t *testing.T) {
	gs := newBloodrushGame(t)
	gs.Seats[0].ManaPool = 5
	source := newBloodrushCard("Slaughterhorn", "{1}{G}", 4, 4, "")
	// Source NOT placed in hand.
	attacker := newAttackerOnBattlefield(gs, 0, "Wolf", 1, 1)
	if err := ActivateBloodrush(gs, 0, source, attacker); err == nil {
		t.Fatal("ActivateBloodrush should reject source that is not in hand")
	}
}

func TestActivateBloodrush_RejectsInsufficientMana(t *testing.T) {
	gs := newBloodrushGame(t)
	gs.Seats[0].ManaPool = 0
	source := newBloodrushCard("Slaughterhorn", "{1}{G}", 4, 4, "")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, source)
	attacker := newAttackerOnBattlefield(gs, 0, "Wolf", 1, 1)
	if err := ActivateBloodrush(gs, 0, source, attacker); err == nil {
		t.Fatal("ActivateBloodrush should reject when seat can't afford the cost")
	}
	if len(gs.Seats[0].Graveyard) != 0 {
		t.Error("source should not have been discarded on a rejected activation")
	}
}

func TestActivateBloodrush_RejectsNonCreatureTarget(t *testing.T) {
	gs := newBloodrushGame(t)
	gs.Seats[0].ManaPool = 5
	source := newBloodrushCard("Slaughterhorn", "{1}{G}", 4, 4, "")
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, source)

	enchantment := &Permanent{
		Card: &Card{
			Name:  "Honor of the Pure",
			Types: []string{"enchantment"},
			AST:   &gameast.CardAST{Name: "Honor of the Pure"},
		},
		Controller: 0,
	}
	setPermFlag(enchantment, flagAttacking, true) // can't actually happen, but verifies the creature gate
	if err := ActivateBloodrush(gs, 0, source, enchantment); err == nil {
		t.Fatal("ActivateBloodrush should reject non-creature target")
	}
}

func TestActivateBloodrush_RejectsMissingKeyword(t *testing.T) {
	gs := newBloodrushGame(t)
	gs.Seats[0].ManaPool = 5
	source := &Card{
		Name:  "Plain Bear",
		Types: []string{"creature"},
		AST:   &gameast.CardAST{Name: "Plain Bear", Abilities: []gameast.Ability{}},
	}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, source)
	attacker := newAttackerOnBattlefield(gs, 0, "Wolf", 1, 1)
	if err := ActivateBloodrush(gs, 0, source, attacker); err == nil {
		t.Fatal("ActivateBloodrush should reject card without bloodrush keyword")
	}
}
