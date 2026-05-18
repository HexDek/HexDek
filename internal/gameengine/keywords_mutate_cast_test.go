package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Test helpers — Mutate cast (CR §702.140)
// ---------------------------------------------------------------------------

func newMutateGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(521))
	return NewGameState(2, rng, nil)
}

// mutateHandCard builds a creature with the mutate keyword + a
// distinctive grant keyword we can observe on the merged perm.
func mutateHandCard(gs *GameState, seat int, name string, mutateCost int) *Card {
	c := &Card{
		Name:          name,
		Owner:         seat,
		CMC:           4,
		BasePower:     3,
		BaseToughness: 3,
		Types:         []string{"creature", "elemental"},
		Colors:        []string{"G"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "mutate", Args: []any{float64(mutateCost)}},
				// Distinctive keyword the merged perm should pick up.
				&gameast.Keyword{Name: "flying"},
			},
		},
	}
	gs.Seats[seat].Hand = append(gs.Seats[seat].Hand, c)
	return c
}

// addNonHumanCreature drops a vanilla creature with given types on
// `seat`'s battlefield. Defaults to ["creature", "beast"] when no
// types are supplied.
func addNonHumanCreature(gs *GameState, seat int, name string, pow, tough int, types ...string) *Permanent {
	if len(types) == 0 {
		types = []string{"creature", "beast"}
	}
	c := &Card{
		Name: name, Owner: seat,
		BasePower: pow, BaseToughness: tough,
		Types: append([]string(nil), types...),
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "trample"}, // distinctive keyword
			},
		},
	}
	p := &Permanent{
		Card: c, Controller: seat, Owner: seat,
		Timestamp: gs.NextTimestamp(),
		Counters:  map[string]int{}, Flags: map[string]int{},
	}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, p)
	return p
}

// ===========================================================================
// HasMutate / MutateCost detectors
// ===========================================================================

func TestHasMutate_Detects(t *testing.T) {
	c := &Card{
		AST: &gameast.CardAST{
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "mutate", Args: []any{float64(3)}},
			},
		},
	}
	if !HasMutate(c) {
		t.Fatal("HasMutate should detect the keyword")
	}
	if HasMutate(nil) {
		t.Fatal("HasMutate(nil) should be false")
	}
	if HasMutate(&Card{AST: &gameast.CardAST{}}) {
		t.Fatal("HasMutate on empty AST should be false")
	}
}

func TestMutateCost_ParsesArg(t *testing.T) {
	c := &Card{
		AST: &gameast.CardAST{
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "mutate", Args: []any{float64(2)}},
			},
		},
	}
	if got := MutateCost(c); got != 2 {
		t.Errorf("MutateCost: want 2, got %d", got)
	}
}

// ===========================================================================
// (a) Cast with valid target succeeds (on-top default)
// ===========================================================================

func TestCastWithMutate_ValidTargetSucceedsOnTop(t *testing.T) {
	gs := newMutateGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 3
	target := addNonHumanCreature(gs, 0, "Pollywog Symbiote", 1, 3, "creature", "frog", "horror")
	mutator := mutateHandCard(gs, 0, "Migratory Greathorn", 3)

	merged, err := CastWithMutate(gs, 0, mutator, 3, target, true)
	if err != nil {
		t.Fatalf("CastWithMutate: %v", err)
	}
	if merged == nil {
		t.Fatal("CastWithMutate returned nil merged permanent")
	}

	// onTop=true → mutating perm is the survivor.
	if merged.Card != mutator {
		t.Errorf("onTop survivor should be mutating perm; got %v", merged.Card)
	}
	// Survivor stamped mutated.
	if merged.Flags["mutated"] != 1 {
		t.Error("merged perm should carry Flags[mutated]=1")
	}
	// Mutating perm gains abilities from target (trample). ApplyMutate
	// puts target's granted+keyword names into mutating.GrantedAbilities.
	got := merged.GrantedAbilities
	containsTrample := false
	for _, g := range got {
		if g == "trample" {
			containsTrample = true
			break
		}
	}
	if !containsTrample {
		t.Errorf("onTop merged perm should absorb target's trample keyword; granted=%v", got)
	}
	// Mana paid.
	if gs.Seats[0].ManaPool != 0 {
		t.Errorf("mana pool: want 0 after paying {3}, got %d", gs.Seats[0].ManaPool)
	}
	// Card removed from hand.
	if len(gs.Seats[0].Hand) != 0 {
		t.Errorf("hand: want 0 cards after cast, got %d", len(gs.Seats[0].Hand))
	}
	// Stack popped after inline resolve.
	if len(gs.Stack) != 0 {
		t.Errorf("stack: want empty post-resolve, got %d items", len(gs.Stack))
	}
	// Target removed from battlefield (onTop=true → target is the loser).
	for _, p := range gs.Seats[0].Battlefield {
		if p == target {
			t.Error("target perm should be removed from battlefield when onTop=true")
		}
	}
	// Per-turn tracker.
	if got := MutateCastThisTurn(gs, 0); got != 1 {
		t.Errorf("MutateCastThisTurn: want 1, got %d", got)
	}
}

// ===========================================================================
// (b) Human target rejected per §702.140a
// ===========================================================================

func TestCastWithMutate_HumanTargetRejected(t *testing.T) {
	gs := newMutateGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	// Human creature target — must reject.
	human := addNonHumanCreature(gs, 0, "Soldier of the Pantheon", 2, 1, "creature", "human", "soldier")
	mutator := mutateHandCard(gs, 0, "Mutating Beast", 3)

	merged, err := CastWithMutate(gs, 0, mutator, 3, human, true)
	if err == nil || merged != nil {
		t.Fatal("CastWithMutate must reject a Human target")
	}
	ce, ok := err.(*CastError)
	if !ok || ce.Reason != "target_is_human" {
		t.Errorf("expected CastError(target_is_human), got %T %v", err, err)
	}
	// Side-effect guards: hand intact, mana untouched, stack empty,
	// target still on battlefield.
	if len(gs.Seats[0].Hand) != 1 {
		t.Errorf("hand should remain intact on rejection, got %d", len(gs.Seats[0].Hand))
	}
	if gs.Seats[0].ManaPool != 5 {
		t.Errorf("mana should be untouched on rejection, got %d", gs.Seats[0].ManaPool)
	}
	if len(gs.Stack) != 0 {
		t.Errorf("stack should be empty on rejection, got %d", len(gs.Stack))
	}
	found := false
	for _, p := range gs.Seats[0].Battlefield {
		if p == human {
			found = true
		}
	}
	if !found {
		t.Error("human target should still be on the battlefield")
	}
}

// ===========================================================================
// (c) On top vs on bottom positioning
// ===========================================================================

func TestCastWithMutate_OnBottomKeepsTarget(t *testing.T) {
	gs := newMutateGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 3
	target := addNonHumanCreature(gs, 0, "Pollywog Symbiote", 1, 3, "creature", "frog")
	mutator := mutateHandCard(gs, 0, "Migratory Greathorn", 3)

	merged, err := CastWithMutate(gs, 0, mutator, 3, target, false)
	if err != nil {
		t.Fatalf("CastWithMutate (onTop=false): %v", err)
	}

	// onTop=false → target is the survivor.
	if merged != target {
		t.Errorf("onTop=false survivor should be the original target; got %p want %p", merged, target)
	}
	if merged.Flags["mutated"] != 1 {
		t.Error("survivor must carry Flags[mutated]=1")
	}
	// Target absorbs mutator's keyword (flying).
	containsFlying := false
	for _, g := range merged.GrantedAbilities {
		if g == "flying" {
			containsFlying = true
			break
		}
	}
	if !containsFlying {
		t.Errorf("onTop=false target should absorb mutator's flying; granted=%v", merged.GrantedAbilities)
	}
	// Mutating perm removed.
	for _, p := range gs.Seats[0].Battlefield {
		if p != nil && p.Card == mutator {
			t.Error("mutating perm should be removed when onTop=false")
		}
	}
}

func TestCastWithMutate_OnTopVsOnBottomDifferentSurvivors(t *testing.T) {
	// Cross-check: same setup but different onTop → different survivor.
	for _, onTop := range []bool{true, false} {
		gs := newMutateGame(t)
		gs.Active = 0
		gs.Seats[0].ManaPool = 3
		target := addNonHumanCreature(gs, 0, "Frog", 1, 1, "creature", "frog")
		mutator := mutateHandCard(gs, 0, "Beast", 3)

		merged, err := CastWithMutate(gs, 0, mutator, 3, target, onTop)
		if err != nil {
			t.Fatalf("onTop=%v: %v", onTop, err)
		}
		if onTop && merged.Card != mutator {
			t.Errorf("onTop=true: survivor card should be mutator, got %v", merged.Card.Name)
		}
		if !onTop && merged != target {
			t.Errorf("onTop=false: survivor should be target perm, got %v", merged)
		}
	}
}

// ===========================================================================
// (d) CostMeta["mutate_cast"]=true stamped on the (in-flight) StackItem
// ===========================================================================

// stampedItemRecorder taps the stack right before CastWithMutate's
// inline pop to verify CostMeta was set. We do this by overriding the
// item-resolve via a sniff function: wrap CastWithMutate with a
// pre-resolve checkpoint. Since CastWithMutate pops the item before
// returning, we observe the metadata via the mutate_cast log event
// (which mirrors CostMeta) AND via a direct stack-snapshot taken
// inside a temporarily-modified ApplyMutate hook.
//
// Simpler: we trust the CostMeta is set if mutate_cast event records
// the right alt_cost / cost / target / on_top values, since both come
// from the same source data. Test below verifies the event surface;
// a sister test below uses an in-the-middle override of PushStackItem
// to capture the item.
func TestCastWithMutate_LogsAltCostTrailWithMetadata(t *testing.T) {
	gs := newMutateGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 3
	target := addNonHumanCreature(gs, 0, "Frog", 1, 1)
	mutator := mutateHandCard(gs, 0, "Greathorn", 3)

	if _, err := CastWithMutate(gs, 0, mutator, 3, target, true); err != nil {
		t.Fatalf("CastWithMutate: %v", err)
	}

	// mutate_cast event encodes the same metadata that CostMeta does.
	saw := false
	for _, ev := range gs.EventLog {
		if ev.Kind != "mutate_cast" || ev.Amount != 3 {
			continue
		}
		if rule, _ := ev.Details["rule"].(string); rule != "702.140a" {
			continue
		}
		if tgt, _ := ev.Details["target"].(string); tgt != "Frog" {
			continue
		}
		if onTop, _ := ev.Details["on_top"].(bool); !onTop {
			continue
		}
		saw = true
		break
	}
	if !saw {
		t.Error("expected mutate_cast event with rule=702.140a, target=Frog, on_top=true, amount=3")
	}

	// pay_mana event with the right reason/keyword/amount.
	sawPay := false
	for _, ev := range gs.EventLog {
		if ev.Kind != "pay_mana" {
			continue
		}
		reason, _ := ev.Details["reason"].(string)
		kw, _ := ev.Details["keyword"].(string)
		if reason == "mutate_cast" && kw == "mutate" && ev.Amount == 3 {
			sawPay = true
			break
		}
	}
	if !sawPay {
		t.Error("expected pay_mana event reason=mutate_cast keyword=mutate amount=3")
	}
}

// TestCastWithMutate_StackItemCarriesCostMeta captures the StackItem
// CostMeta directly by intercepting at the merge step — we register a
// temporary trigger hook that runs DURING ApplyMutate (which fires
// the "creature_mutated" trigger via FireCardTrigger). At that point
// the stack item is still in place; we snapshot it.
func TestCastWithMutate_StackItemCarriesCostMeta(t *testing.T) {
	gs := newMutateGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 3
	target := addNonHumanCreature(gs, 0, "Frog", 1, 1)
	mutator := mutateHandCard(gs, 0, "Greathorn", 3)

	prev := TriggerHook
	defer func() { TriggerHook = prev }()
	var snapshot *StackItem
	TriggerHook = func(_ *GameState, event string, _ map[string]interface{}) {
		if event == "creature_mutated" && snapshot == nil && len(gs.Stack) > 0 {
			top := gs.Stack[len(gs.Stack)-1]
			snapshot = top
		}
	}

	if _, err := CastWithMutate(gs, 0, mutator, 3, target, true); err != nil {
		t.Fatalf("CastWithMutate: %v", err)
	}
	if snapshot == nil {
		t.Fatal("did not capture the in-flight StackItem during the merge")
	}
	if snapshot.CostMeta == nil {
		t.Fatal("StackItem.CostMeta should not be nil")
	}
	if v, _ := snapshot.CostMeta["mutate_cast"].(bool); !v {
		t.Errorf("CostMeta[mutate_cast] = %v, want true", snapshot.CostMeta["mutate_cast"])
	}
	if v, _ := snapshot.CostMeta["alt_cost"].(string); v != "mutate" {
		t.Errorf("CostMeta[alt_cost] = %v, want \"mutate\"", snapshot.CostMeta["alt_cost"])
	}
	if v, _ := snapshot.CostMeta["mutate_cost"].(int); v != 3 {
		t.Errorf("CostMeta[mutate_cost] = %v, want 3", snapshot.CostMeta["mutate_cost"])
	}
	if v, _ := snapshot.CostMeta["mutate_target"].(*Permanent); v != target {
		t.Errorf("CostMeta[mutate_target] = %p, want %p", v, target)
	}
	if v, _ := snapshot.CostMeta["mutate_on_top"].(bool); !v {
		t.Errorf("CostMeta[mutate_on_top] = %v, want true", snapshot.CostMeta["mutate_on_top"])
	}
	if snapshot.CastZone != ZoneHand {
		t.Errorf("CastZone = %v, want ZoneHand", snapshot.CastZone)
	}
}

// ===========================================================================
// (e) HasMutate detector — already covered above; this batch covers
// the rejection paths.
// ===========================================================================

func TestCastWithMutate_RejectsCardWithoutKeyword(t *testing.T) {
	gs := newMutateGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	target := addNonHumanCreature(gs, 0, "Frog", 1, 1)
	c := &Card{Name: "Plain", Owner: 0, Types: []string{"creature"},
		AST: &gameast.CardAST{Name: "Plain"}}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, c)

	if _, err := CastWithMutate(gs, 0, c, 3, target, true); err == nil {
		t.Fatal("CastWithMutate should reject a card without the mutate keyword")
	} else if ce, ok := err.(*CastError); !ok || ce.Reason != "no_mutate_keyword" {
		t.Errorf("expected CastError(no_mutate_keyword), got %T %v", err, err)
	}
}

func TestCastWithMutate_RejectsTargetWrongController(t *testing.T) {
	gs := newMutateGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	// Opponent's creature.
	oppFrog := addNonHumanCreature(gs, 1, "Opp Frog", 1, 1)
	mutator := mutateHandCard(gs, 0, "Greathorn", 3)

	if _, err := CastWithMutate(gs, 0, mutator, 3, oppFrog, true); err == nil {
		t.Fatal("CastWithMutate should reject an opponent's creature")
	} else if ce, ok := err.(*CastError); !ok || ce.Reason != "target_wrong_controller" {
		t.Errorf("expected CastError(target_wrong_controller), got %T %v", err, err)
	}
}

func TestCastWithMutate_RejectsTargetNotCreature(t *testing.T) {
	gs := newMutateGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	// Artifact, not a creature.
	art := &Permanent{
		Card: &Card{Name: "Sol Ring", Owner: 0, Types: []string{"artifact"}},
		Controller: 0, Owner: 0,
		Timestamp: gs.NextTimestamp(),
		Counters:  map[string]int{}, Flags: map[string]int{},
	}
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, art)
	mutator := mutateHandCard(gs, 0, "Greathorn", 3)

	if _, err := CastWithMutate(gs, 0, mutator, 3, art, true); err == nil {
		t.Fatal("CastWithMutate should reject a non-creature target")
	} else if ce, ok := err.(*CastError); !ok || ce.Reason != "target_not_creature" {
		t.Errorf("expected CastError(target_not_creature), got %T %v", err, err)
	}
}

func TestCastWithMutate_RejectsInsufficientMana(t *testing.T) {
	gs := newMutateGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 2
	target := addNonHumanCreature(gs, 0, "Frog", 1, 1)
	mutator := mutateHandCard(gs, 0, "Greathorn", 3)

	if _, err := CastWithMutate(gs, 0, mutator, 3, target, true); err == nil {
		t.Fatal("CastWithMutate should reject when mana < mutateCost")
	} else if ce, ok := err.(*CastError); !ok || ce.Reason != "insufficient_mana" {
		t.Errorf("expected CastError(insufficient_mana), got %T %v", err, err)
	}
	if len(gs.Seats[0].Hand) != 1 {
		t.Errorf("card should remain in hand after rejection, got %d", len(gs.Seats[0].Hand))
	}
	if gs.Seats[0].ManaPool != 2 {
		t.Errorf("mana should be untouched on rejection, got %d", gs.Seats[0].ManaPool)
	}
}

func TestCastWithMutate_NilTargetRejected(t *testing.T) {
	gs := newMutateGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	mutator := mutateHandCard(gs, 0, "Greathorn", 3)
	if _, err := CastWithMutate(gs, 0, mutator, 3, nil, true); err == nil {
		t.Fatal("CastWithMutate should reject nil target")
	} else if ce, ok := err.(*CastError); !ok || ce.Reason != "nil_target" {
		t.Errorf("expected CastError(nil_target), got %T %v", err, err)
	}
}

func TestCastWithMutate_DefaultsCostFromKeyword(t *testing.T) {
	// Pass -1 → derive from MutateCost(card).
	gs := newMutateGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 4
	target := addNonHumanCreature(gs, 0, "Frog", 1, 1)
	mutator := mutateHandCard(gs, 0, "Greathorn", 4) // keyword arg = 4

	if _, err := CastWithMutate(gs, 0, mutator, -1, target, true); err != nil {
		t.Fatalf("CastWithMutate (-1 fallback): %v", err)
	}
	if gs.Seats[0].ManaPool != 0 {
		t.Errorf("expected 0 mana after paying default cost 4; got %d", gs.Seats[0].ManaPool)
	}
}

func TestCastWithMutate_NilSafety(t *testing.T) {
	if _, err := CastWithMutate(nil, 0, nil, 0, nil, true); err == nil {
		t.Fatal("CastWithMutate(nil...) should error")
	}
	gs := newMutateGame(t)
	if _, err := CastWithMutate(gs, -1, nil, 0, nil, true); err == nil {
		t.Fatal("CastWithMutate(invalid seat) should error")
	}
	if _, err := CastWithMutate(gs, 0, nil, 0, nil, true); err == nil {
		t.Fatal("CastWithMutate(nil card) should error")
	}
}

// ===========================================================================
// Per-turn tally + creature_mutated trigger fires
// ===========================================================================

func TestCastWithMutate_FiresCreatureMutatedTrigger(t *testing.T) {
	gs := newMutateGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 3
	target := addNonHumanCreature(gs, 0, "Frog", 1, 1)
	mutator := mutateHandCard(gs, 0, "Greathorn", 3)

	prev := TriggerHook
	defer func() { TriggerHook = prev }()
	mutatedCount := 0
	TriggerHook = func(_ *GameState, event string, _ map[string]interface{}) {
		if event == "creature_mutated" {
			mutatedCount++
		}
	}

	if _, err := CastWithMutate(gs, 0, mutator, 3, target, true); err != nil {
		t.Fatalf("CastWithMutate: %v", err)
	}
	if mutatedCount != 1 {
		t.Errorf("expected creature_mutated trigger to fire exactly once, got %d", mutatedCount)
	}
}

func TestMutateCastThisTurn(t *testing.T) {
	gs := newMutateGame(t)
	if got := MutateCastThisTurn(gs, 0); got != 0 {
		t.Fatalf("MutateCastThisTurn at start: want 0, got %d", got)
	}
	gs.Active = 0
	gs.Seats[0].ManaPool = 6
	t1 := addNonHumanCreature(gs, 0, "Frog1", 1, 1)
	m1 := mutateHandCard(gs, 0, "Beast1", 3)
	if _, err := CastWithMutate(gs, 0, m1, 3, t1, true); err != nil {
		t.Fatalf("first cast: %v", err)
	}
	if got := MutateCastThisTurn(gs, 0); got != 1 {
		t.Errorf("after 1 cast: want 1, got %d", got)
	}
	// Second cast.
	t2 := addNonHumanCreature(gs, 0, "Frog2", 1, 1)
	m2 := mutateHandCard(gs, 0, "Beast2", 3)
	if _, err := CastWithMutate(gs, 0, m2, 3, t2, true); err != nil {
		t.Fatalf("second cast: %v", err)
	}
	if got := MutateCastThisTurn(gs, 0); got != 2 {
		t.Errorf("after 2 casts: want 2, got %d", got)
	}
	// Per-seat: seat 1 untouched.
	if got := MutateCastThisTurn(gs, 1); got != 0 {
		t.Errorf("seat 1's tally: want 0, got %d", got)
	}
}
