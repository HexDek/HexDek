package gameengine

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// CR §702.96 — Overload alt-cost mechanic.
//
// These tests cover:
//   - HasOverload / OverloadCost keyword detection
//   - CastOverload pays the alt cost and stamps CostMeta["overloaded"]=true
//   - PickTarget fans out single-target filters when the resolution
//     context is overloaded (gs.Flags["overload_active"] set)
//   - PickTarget does NOT fan out when overload is inactive
//   - Self / controller references are not rewritten
//   - The side-channel flag is correctly cleared after resolution

func newOverloadCard(name string, cost int) *Card {
	return &Card{
		Name: name,
		CMC:  2,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "overload", Args: []interface{}{float64(cost)}},
			},
		},
		Types: []string{"instant"},
	}
}

func TestOverload_CastOverload_AliasMatchesCastWithOverload(t *testing.T) {
	gs := newKWCombatGame(t)
	gs.Seats[0].ManaPool = 10
	card := newOverloadCard("Cyclonic Rift", 7)
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	if err := CastOverload(gs, 0, card); err != nil {
		t.Fatalf("CastOverload should succeed: %v", err)
	}
	if gs.Seats[0].ManaPool != 3 {
		t.Errorf("CastOverload should pay 7 mana, pool=%d", gs.Seats[0].ManaPool)
	}
	if len(gs.Stack) != 1 {
		t.Fatalf("spell should be on the stack, got %d", len(gs.Stack))
	}
	if !IsOverloaded(gs.Stack[0]) {
		t.Fatal("stack item should carry CostMeta[\"overloaded\"]=true")
	}
}

func TestOverload_CastOverload_InsufficientMana(t *testing.T) {
	gs := newKWCombatGame(t)
	gs.Seats[0].ManaPool = 3 // can't afford 7
	card := newOverloadCard("Cyclonic Rift", 7)
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)

	if err := CastOverload(gs, 0, card); err == nil {
		t.Fatal("CastOverload should fail when mana pool < overload cost")
	}
	if len(gs.Stack) != 0 {
		t.Errorf("failed cast must not leave a stack item, got %d", len(gs.Stack))
	}
}

func TestOverload_CastOverload_NoKeyword(t *testing.T) {
	gs := newKWCombatGame(t)
	gs.Seats[0].ManaPool = 10
	plain := &Card{Name: "Plain Spell", CMC: 1, AST: &gameast.CardAST{Name: "Plain Spell"}}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, plain)

	if err := CastOverload(gs, 0, plain); err == nil {
		t.Fatal("CastOverload on a card without overload should fail")
	}
}

// PickTarget with overload active must fan out single-target permanent
// filters to every matching battlefield permanent — CR §702.96 "change
// target to each".
func TestOverload_PickTarget_FansOutPermanents(t *testing.T) {
	gs := newKWCombatGame(t)
	addKWCombatBattlefield(gs, 0, "Goblin A", 1, 1, "creature")
	addKWCombatBattlefield(gs, 1, "Goblin B", 1, 1, "creature")
	addKWCombatBattlefield(gs, 1, "Goblin C", 2, 2, "creature")

	src := &Permanent{Card: &Card{Name: "Overloaded Spell"}, Controller: 0}
	f := gameast.Filter{Base: "creature", Targeted: true}

	// Without overload, single target only.
	if got := PickTarget(gs, src, f); len(got) != 1 {
		t.Fatalf("non-overloaded single-target filter should return 1 target, got %d", len(got))
	}

	// With overload active, fan out across every matching creature.
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	gs.Flags["overload_active"] = 1
	defer delete(gs.Flags, "overload_active")

	got := PickTarget(gs, src, f)
	if len(got) != 3 {
		t.Fatalf("overloaded creature target should fan out to 3 creatures, got %d", len(got))
	}
}

func TestOverload_PickTarget_FansOutPlayers(t *testing.T) {
	gs := newKWCombatGame4P(t)
	src := &Permanent{Card: &Card{Name: "Overloaded Ping"}, Controller: 0}
	f := gameast.Filter{Base: "opponent", Targeted: true}

	if got := PickTarget(gs, src, f); len(got) != 1 {
		t.Fatalf("non-overloaded opponent filter should return 1 target, got %d", len(got))
	}

	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	gs.Flags["overload_active"] = 1
	defer delete(gs.Flags, "overload_active")

	got := PickTarget(gs, src, f)
	if len(got) != 3 {
		t.Fatalf("overloaded opponent filter should fan out to 3 opponents, got %d", len(got))
	}
}

// "Self", "you", "controller", and equipped/enchanted self-references
// don't read as "target" on the printed card, so overload must leave
// them alone — CR §702.96 only rewrites the word "target".
func TestOverload_PickTarget_DoesNotExpandSelfReferences(t *testing.T) {
	gs := newKWCombatGame(t)
	addKWCombatBattlefield(gs, 0, "Other 1", 1, 1, "creature")
	addKWCombatBattlefield(gs, 1, "Other 2", 1, 1, "creature")
	src := addKWCombatBattlefield(gs, 0, "Source", 2, 2, "creature")

	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	gs.Flags["overload_active"] = 1
	defer delete(gs.Flags, "overload_active")

	selfTargets := PickTarget(gs, src, gameast.Filter{Base: "self"})
	if len(selfTargets) != 1 || selfTargets[0].Permanent != src {
		t.Fatalf("self filter must always resolve to the source even when overloaded, got %v", selfTargets)
	}

	youTargets := PickTarget(gs, src, gameast.Filter{Base: "you"})
	if len(youTargets) != 1 {
		t.Fatalf("you filter must not expand under overload, got %d", len(youTargets))
	}
}

// The side-channel flag must be set during resolution and cleared after.
func TestOverload_ResolutionFlag_SetAndClearedAroundResolve(t *testing.T) {
	gs := newKWCombatGame(t)

	overloadedItem := &StackItem{
		Controller: 0,
		Card:       newOverloadCard("Cyclonic Rift", 7),
		Kind:       "spell",
		CostMeta:   map[string]interface{}{"overloaded": true},
	}
	if IsOverloadActive(gs) {
		t.Fatal("overload flag should be unset at rest")
	}

	cleanup := beginOverloadResolution(gs, overloadedItem)
	if !IsOverloadActive(gs) {
		t.Fatal("beginOverloadResolution should set the side-channel flag")
	}
	cleanup()
	if IsOverloadActive(gs) {
		t.Fatal("cleanup should clear the side-channel flag")
	}

	plainItem := &StackItem{Controller: 0, Card: &Card{Name: "Plain"}, Kind: "spell"}
	cleanup = beginOverloadResolution(gs, plainItem)
	if IsOverloadActive(gs) {
		t.Fatal("non-overloaded item must not set the flag")
	}
	cleanup() // no-op, must not panic
}

// Nested-resolution safety: if a hook re-enters beginOverloadResolution
// (e.g. an overloaded spell triggers another overloaded effect), the
// outer scope's flag state must be restored on inner cleanup.
func TestOverload_ResolutionFlag_NestedRestore(t *testing.T) {
	gs := newKWCombatGame(t)
	outer := &StackItem{Controller: 0, Card: newOverloadCard("Outer", 5), CostMeta: map[string]interface{}{"overloaded": true}}
	inner := &StackItem{Controller: 0, Card: newOverloadCard("Inner", 3), CostMeta: map[string]interface{}{"overloaded": true}}

	endOuter := beginOverloadResolution(gs, outer)
	endInner := beginOverloadResolution(gs, inner)
	if !IsOverloadActive(gs) {
		t.Fatal("nested overload should keep flag set")
	}
	endInner()
	if !IsOverloadActive(gs) {
		t.Fatal("inner cleanup must NOT clear the flag while outer is still active")
	}
	endOuter()
	if IsOverloadActive(gs) {
		t.Fatal("outer cleanup should fully clear the flag")
	}
}

// End-to-end: an overloaded damage spell deals damage to every matching
// permanent when resolved through ResolveStackTop.
func TestOverload_EndToEnd_DamageFansOutOnResolve(t *testing.T) {
	gs := newKWCombatGame4P(t)
	a := addKWCombatBattlefield(gs, 1, "A", 1, 2, "creature")
	b := addKWCombatBattlefield(gs, 2, "B", 1, 2, "creature")
	c := addKWCombatBattlefield(gs, 3, "C", 1, 2, "creature")

	// Synthesize a "deal 1 damage to target creature" spell, cast it
	// for overload, and resolve. Every opponent creature should take 1.
	card := &Card{
		Name: "Electrickery",
		CMC:  1,
		AST: &gameast.CardAST{
			Name: "Electrickery",
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "overload", Args: []interface{}{float64(3)}},
			},
		},
		Types: []string{"instant"},
	}
	dmg := &gameast.Damage{
		Amount: gameast.NumberOrRef{IsInt: true, Int: 1},
		Target: gameast.Filter{Base: "creature", Targeted: true},
	}

	gs.Seats[0].ManaPool = 5
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, card)
	if err := CastOverload(gs, 0, card); err != nil {
		t.Fatalf("CastOverload: %v", err)
	}
	gs.Stack[len(gs.Stack)-1].Effect = dmg

	ResolveStackTop(gs)

	// Each opponent creature should have 1 marked damage.
	for _, p := range []*Permanent{a, b, c} {
		if p.MarkedDamage != 1 {
			t.Errorf("%s should have 1 marked damage under overload, got %d", p.Card.Name, p.MarkedDamage)
		}
	}

	// And the flag is cleared after resolution.
	if IsOverloadActive(gs) {
		t.Error("overload flag should be cleared after ResolveStackTop returns")
	}
}

// Sanity: when the same spell is NOT cast for overload, damage hits a
// single target — the alt-cost path is what triggers fan-out.
func TestOverload_EndToEnd_DamageStaysSingleTarget_WhenNotOverloaded(t *testing.T) {
	gs := newKWCombatGame4P(t)
	a := addKWCombatBattlefield(gs, 1, "A", 1, 2, "creature")
	b := addKWCombatBattlefield(gs, 2, "B", 1, 2, "creature")
	c := addKWCombatBattlefield(gs, 3, "C", 1, 2, "creature")

	card := &Card{Name: "Plain Ping", CMC: 1, AST: &gameast.CardAST{Name: "Plain Ping"}, Types: []string{"instant"}}
	dmg := &gameast.Damage{
		Amount: gameast.NumberOrRef{IsInt: true, Int: 1},
		Target: gameast.Filter{Base: "creature", Targeted: true},
	}

	gs.Stack = append(gs.Stack, &StackItem{
		Controller: 0,
		Card:       card,
		Kind:       "spell",
		Effect:     dmg,
	})
	ResolveStackTop(gs)

	total := a.MarkedDamage + b.MarkedDamage + c.MarkedDamage
	if total != 1 {
		t.Errorf("non-overloaded damage should hit exactly one creature, total marked=%d", total)
	}
}
