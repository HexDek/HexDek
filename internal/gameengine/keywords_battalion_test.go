package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Battalion tests — CR §702.101 (Gatecrash 2013)
// ---------------------------------------------------------------------------

func newBattalionGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(101))
	return NewGameState(2, rng, nil)
}

// newBattalionCreature builds a creature card with the battalion
// keyword.
func newBattalionCreature(name string, owner, power, toughness int) *Card {
	return &Card{
		Name:          name,
		Owner:         owner,
		Types:         []string{"creature"},
		BasePower:     power,
		BaseToughness: toughness,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "battalion"},
			},
		},
	}
}

func newPlainBattalionCreature(name string, owner, power, toughness int) *Card {
	return &Card{
		Name:          name,
		Owner:         owner,
		Types:         []string{"creature"},
		BasePower:     power,
		BaseToughness: toughness,
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{},
		},
	}
}

// putAttackerOnBattlefield mints a Permanent for `card`, places it on
// `seat`'s battlefield, and flips the attacking flag.
func putAttackerOnBattlefield(gs *GameState, seat int, card *Card) *Permanent {
	p := &Permanent{Card: card, Controller: seat, Owner: seat}
	setPermFlag(p, flagAttacking, true)
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, p)
	return p
}

func putIdleOnBattlefield(gs *GameState, seat int, card *Card) *Permanent {
	p := &Permanent{Card: card, Controller: seat, Owner: seat}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, p)
	return p
}

// installCapturingTriggerHook swaps TriggerHook for a capture-only stub
// so the test can assert what FireCardTrigger forwarded. Returns a
// cleanup func and the captured-calls pointer.
type capturedTrigger struct {
	event string
	ctx   map[string]interface{}
}

func installCapturingTriggerHook(t *testing.T) (*[]capturedTrigger, func()) {
	t.Helper()
	original := TriggerHook
	captured := []capturedTrigger{}
	TriggerHook = func(gs *GameState, event string, ctx map[string]interface{}) {
		captured = append(captured, capturedTrigger{event: event, ctx: ctx})
	}
	return &captured, func() { TriggerHook = original }
}

// ---------------------------------------------------------------------------
// HasBattalion / PermanentHasBattalion
// ---------------------------------------------------------------------------

func TestHasBattalion_Detects(t *testing.T) {
	card := newBattalionCreature("Boros Elite", 0, 1, 1)
	if !HasBattalion(card) {
		t.Fatal("HasBattalion should be true for a battalion creature")
	}
}

func TestHasBattalion_Negative(t *testing.T) {
	card := newPlainBattalionCreature("Plain Soldier", 0, 1, 1)
	if HasBattalion(card) {
		t.Fatal("HasBattalion must be false for a card without the keyword")
	}
}

func TestHasBattalion_Nil(t *testing.T) {
	if HasBattalion(nil) {
		t.Fatal("HasBattalion(nil) must be false")
	}
}

func TestPermanentHasBattalion_FromGrant(t *testing.T) {
	card := newPlainBattalionCreature("Soldier", 0, 2, 2)
	p := &Permanent{Card: card, Controller: 0, GrantedAbilities: []string{"battalion"}}
	if !PermanentHasBattalion(p) {
		t.Fatal("PermanentHasBattalion should pick up granted-ability battalion")
	}
}

func TestPermanentHasBattalion_FromFlag(t *testing.T) {
	card := newPlainBattalionCreature("Soldier", 0, 2, 2)
	p := &Permanent{
		Card:       card,
		Controller: 0,
		Flags:      map[string]int{"kw:battalion": 1},
	}
	if !PermanentHasBattalion(p) {
		t.Fatal("PermanentHasBattalion should pick up kw:battalion flag")
	}
}

// ---------------------------------------------------------------------------
// (a) 3 attackers same controller = trigger fires
// ---------------------------------------------------------------------------

func TestFireBattalionTriggers_ThreeAttackersFire(t *testing.T) {
	gs := newBattalionGame(t)
	captured, restore := installCapturingTriggerHook(t)
	defer restore()

	boros := putAttackerOnBattlefield(gs, 0, newBattalionCreature("Boros Elite", 0, 1, 1))
	ally1 := putAttackerOnBattlefield(gs, 0, newPlainBattalionCreature("Ally One", 0, 2, 2))
	ally2 := putAttackerOnBattlefield(gs, 0, newPlainBattalionCreature("Ally Two", 0, 2, 2))

	FireBattalionTriggers(gs, 0, []*Permanent{boros, ally1, ally2})

	// Event log should contain a single battalion_trigger.
	count := 0
	for _, e := range gs.EventLog {
		if e.Kind == "battalion_trigger" {
			count++
			if e.Source != "Boros Elite" {
				t.Errorf("battalion_trigger source = %q, want \"Boros Elite\"", e.Source)
			}
			if e.Amount != 3 {
				t.Errorf("battalion_trigger Amount = %d, want 3", e.Amount)
			}
		}
	}
	if count != 1 {
		t.Errorf("battalion_trigger count = %d, want 1", count)
	}

	// TriggerHook should have received a battalion_triggered fan-out with
	// the right ctx.
	gotBattalion := 0
	for _, c := range *captured {
		if c.event != "battalion_triggered" {
			continue
		}
		gotBattalion++
		if src, _ := c.ctx["source"].(*Permanent); src != boros {
			t.Errorf("ctx[source] mismatch (got %v, want Boros Elite)", src)
		}
		if ctrl, _ := c.ctx["controller"].(int); ctrl != 0 {
			t.Errorf("ctx[controller] = %d, want 0", ctrl)
		}
		if cnt, _ := c.ctx["count"].(int); cnt != 3 {
			t.Errorf("ctx[count] = %d, want 3", cnt)
		}
	}
	if gotBattalion != 1 {
		t.Errorf("battalion_triggered fan-out count = %d, want 1", gotBattalion)
	}
}

// ---------------------------------------------------------------------------
// (b) 2 attackers = no trigger
// ---------------------------------------------------------------------------

func TestFireBattalionTriggers_TwoAttackersDoNotFire(t *testing.T) {
	gs := newBattalionGame(t)
	captured, restore := installCapturingTriggerHook(t)
	defer restore()

	boros := putAttackerOnBattlefield(gs, 0, newBattalionCreature("Boros Elite", 0, 1, 1))
	ally := putAttackerOnBattlefield(gs, 0, newPlainBattalionCreature("Ally", 0, 2, 2))

	FireBattalionTriggers(gs, 0, []*Permanent{boros, ally})

	for _, e := range gs.EventLog {
		if e.Kind == "battalion_trigger" {
			t.Fatalf("battalion_trigger should NOT fire with only 2 attackers; got %+v", e)
		}
	}
	for _, c := range *captured {
		if c.event == "battalion_triggered" {
			t.Fatal("battalion_triggered fan-out should not fire with 2 attackers")
		}
	}
}

// ---------------------------------------------------------------------------
// (c) Opponent's creatures don't count
// ---------------------------------------------------------------------------

func TestFireBattalionTriggers_OpponentAttackersDoNotCount(t *testing.T) {
	gs := newBattalionGame(t)
	_, restore := installCapturingTriggerHook(t)
	defer restore()

	boros := putAttackerOnBattlefield(gs, 0, newBattalionCreature("Boros Elite", 0, 1, 1))
	myAlly := putAttackerOnBattlefield(gs, 0, newPlainBattalionCreature("My Ally", 0, 2, 2))
	// Seat 1 has two attackers, but those are OPPONENT'S attackers — must
	// NOT count toward seat 0's battalion gate.
	opp1 := putAttackerOnBattlefield(gs, 1, newPlainBattalionCreature("Opp One", 1, 2, 2))
	opp2 := putAttackerOnBattlefield(gs, 1, newPlainBattalionCreature("Opp Two", 1, 2, 2))

	FireBattalionTriggers(gs, 0, []*Permanent{boros, myAlly, opp1, opp2})

	for _, e := range gs.EventLog {
		if e.Kind == "battalion_trigger" {
			t.Fatalf("battalion_trigger fired with only 2 same-controller attackers (boros + myAlly); got %+v", e)
		}
	}
}

// ---------------------------------------------------------------------------
// (d) Source must itself be attacking
// ---------------------------------------------------------------------------

func TestFireBattalionTriggers_SourceMustBeAttacking(t *testing.T) {
	gs := newBattalionGame(t)
	_, restore := installCapturingTriggerHook(t)
	defer restore()

	// Boros Elite is on the battlefield but NOT attacking (e.g. tapped,
	// or held back). Two other allies ARE attacking.
	boros := putIdleOnBattlefield(gs, 0, newBattalionCreature("Boros Elite", 0, 1, 1))
	a1 := putAttackerOnBattlefield(gs, 0, newPlainBattalionCreature("Ally One", 0, 2, 2))
	a2 := putAttackerOnBattlefield(gs, 0, newPlainBattalionCreature("Ally Two", 0, 2, 2))
	a3 := putAttackerOnBattlefield(gs, 0, newPlainBattalionCreature("Ally Three", 0, 2, 2))

	// Important: only attackers should be in the slice for the hook
	// caller. But test the guard: if a caller mistakenly includes a
	// non-attacker, FireBattalionTriggers must skip it.
	FireBattalionTriggers(gs, 0, []*Permanent{boros, a1, a2, a3})

	for _, e := range gs.EventLog {
		if e.Kind == "battalion_trigger" {
			t.Fatalf("battalion_trigger should not fire when source isn't attacking; got %+v", e)
		}
	}
}

// ---------------------------------------------------------------------------
// More than one battalion source — each fires independently
// ---------------------------------------------------------------------------

func TestFireBattalionTriggers_MultipleBattalionSourcesEachFire(t *testing.T) {
	gs := newBattalionGame(t)
	_, restore := installCapturingTriggerHook(t)
	defer restore()

	b1 := putAttackerOnBattlefield(gs, 0, newBattalionCreature("Boros Elite", 0, 1, 1))
	b2 := putAttackerOnBattlefield(gs, 0, newBattalionCreature("Wojek Halberdiers", 0, 2, 2))
	a := putAttackerOnBattlefield(gs, 0, newPlainBattalionCreature("Ally", 0, 1, 1))

	FireBattalionTriggers(gs, 0, []*Permanent{b1, b2, a})

	count := 0
	for _, e := range gs.EventLog {
		if e.Kind == "battalion_trigger" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("battalion_trigger count = %d, want 2 (one per battalion source)", count)
	}
}

// ---------------------------------------------------------------------------
// Non-creature attacker in the slice — guard skips it
// ---------------------------------------------------------------------------

func TestFireBattalionTriggers_NonCreaturesSkipped(t *testing.T) {
	gs := newBattalionGame(t)
	_, restore := installCapturingTriggerHook(t)
	defer restore()

	boros := putAttackerOnBattlefield(gs, 0, newBattalionCreature("Boros Elite", 0, 1, 1))
	a1 := putAttackerOnBattlefield(gs, 0, newPlainBattalionCreature("Ally", 0, 2, 2))
	// Non-creature in the slice (e.g. a planeswalker "attacking" via Garruk?).
	// Battalion's "creatures" clause means this must NOT count.
	planeswalker := putAttackerOnBattlefield(gs, 0, &Card{
		Name:  "Garruk",
		Types: []string{"planeswalker"},
		AST:   &gameast.CardAST{Name: "Garruk", Abilities: []gameast.Ability{}},
	})

	FireBattalionTriggers(gs, 0, []*Permanent{boros, a1, planeswalker})

	for _, e := range gs.EventLog {
		if e.Kind == "battalion_trigger" {
			t.Fatalf("battalion_trigger should not fire when only 2 creatures attack (planeswalker shouldn't count); got %+v", e)
		}
	}
}

// ---------------------------------------------------------------------------
// Nil / empty inputs
// ---------------------------------------------------------------------------

func TestFireBattalionTriggers_NilSafe(t *testing.T) {
	// Should not panic on nils.
	FireBattalionTriggers(nil, 0, nil)
	gs := newBattalionGame(t)
	FireBattalionTriggers(gs, 0, nil)
	FireBattalionTriggers(gs, 0, []*Permanent{nil, nil})
}
