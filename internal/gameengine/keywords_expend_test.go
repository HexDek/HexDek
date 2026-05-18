package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Expend tests — CR §702.190 (Bloomburrow 2024)
// ---------------------------------------------------------------------------

func newExpendGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(2190))
	return NewGameState(2, rng, nil)
}

// newExpendCreature builds a creature card with one or more
// "expend N" keyword abilities.
func newExpendCreature(name string, owner int, thresholds ...int) *Card {
	abs := make([]gameast.Ability, 0, len(thresholds))
	for _, n := range thresholds {
		abs = append(abs, &gameast.Keyword{Name: "expend", Args: []any{float64(n)}})
	}
	return &Card{
		Name:          name,
		Owner:         owner,
		Types:         []string{"creature"},
		BasePower:     2,
		BaseToughness: 2,
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: abs,
		},
	}
}

func newPlainCreatureExp(name string, owner int) *Card {
	return &Card{
		Name:          name,
		Owner:         owner,
		Types:         []string{"creature"},
		BasePower:     2,
		BaseToughness: 2,
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{},
		},
	}
}

func putExpendPerm(gs *GameState, seat int, card *Card) *Permanent {
	p := &Permanent{Card: card, Controller: seat, Owner: seat}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, p)
	return p
}

func countExpendEvents(gs *GameState, kind string) int {
	n := 0
	for _, e := range gs.EventLog {
		if e.Kind == kind {
			n++
		}
	}
	return n
}

// installExpendCapture swaps TriggerHook with a capture stub that
// records every "expend" trigger fan-out for assertion.
type expendTriggerCapture struct {
	source     *Permanent
	controller int
	threshold  int
	total      int
}

func installExpendCapture(t *testing.T) (*[]expendTriggerCapture, func()) {
	t.Helper()
	prev := TriggerHook
	captured := []expendTriggerCapture{}
	TriggerHook = func(gs *GameState, event string, ctx map[string]interface{}) {
		if event != "expend" || ctx == nil {
			return
		}
		c := expendTriggerCapture{}
		c.source, _ = ctx["source"].(*Permanent)
		c.controller, _ = ctx["controller"].(int)
		c.threshold, _ = ctx["threshold"].(int)
		c.total, _ = ctx["total"].(int)
		captured = append(captured, c)
	}
	return &captured, func() { TriggerHook = prev }
}

// ---------------------------------------------------------------------------
// HasExpendTrigger / ExpendThresholds
// ---------------------------------------------------------------------------

func TestHasExpendTrigger_Detects(t *testing.T) {
	c := newExpendCreature("Expend4", 0, 4)
	if !HasExpendTrigger(c, 4) {
		t.Fatal("HasExpendTrigger(card, 4) should be true")
	}
	if HasExpendTrigger(c, 8) {
		t.Fatal("HasExpendTrigger(card, 8) should be false when only 4 is declared")
	}
}

func TestHasExpendTrigger_NilSafe(t *testing.T) {
	if HasExpendTrigger(nil, 4) {
		t.Fatal("HasExpendTrigger(nil, 4) must be false")
	}
	if HasExpendTrigger(newPlainCreatureExp("Plain", 0), 4) {
		t.Fatal("HasExpendTrigger on vanilla must be false")
	}
	if HasExpendTrigger(&Card{Name: "no-AST"}, 4) {
		t.Fatal("HasExpendTrigger on AST-less card must be false")
	}
	c := newExpendCreature("Zero", 0, 0)
	if HasExpendTrigger(c, 0) {
		t.Fatal("threshold 0 should be rejected by the constructor's positive-only filter")
	}
}

func TestExpendThresholds_MultiOrderDedup(t *testing.T) {
	c := newExpendCreature("Multi", 0, 4, 8, 4, 12) // duplicate 4 + ordered
	got := ExpendThresholds(c)
	want := []int{4, 8, 12}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("threshold[%d] = %d, want %d", i, got[i], w)
		}
	}
}

func TestExpendThresholds_Empty(t *testing.T) {
	if got := ExpendThresholds(newPlainCreatureExp("Plain", 0)); len(got) != 0 {
		t.Errorf("ExpendThresholds on vanilla = %v, want empty", got)
	}
	if got := ExpendThresholds(nil); len(got) != 0 {
		t.Errorf("ExpendThresholds(nil) = %v, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// (a) Spending {3}{G} = 4 mana counted
// ---------------------------------------------------------------------------

func TestTrackManaSpentThisTurn_AccumulatesAcrossPays(t *testing.T) {
	gs := newExpendGame(t)

	// Simulate paying {3}{G} for a 4-mana spell — generic 3 + green 1.
	// Engine code typically calls SpendMana once with the total cost,
	// but Expend tracks the AMOUNT, not the breakdown — we drive the
	// counter directly to mirror what cost-payment sites would do.
	if got := TrackManaSpentThisTurn(gs, 0, 3); got != 3 {
		t.Errorf("after 3 spent, total = %d, want 3", got)
	}
	if got := TrackManaSpentThisTurn(gs, 0, 1); got != 4 {
		t.Errorf("after 1 more spent, total = %d, want 4", got)
	}
	if gs.Seats[0].Turn.ManaSpent != 4 {
		t.Errorf("Turn.ManaSpent = %d, want 4 (verifies {3}{G} totals correctly)",
			gs.Seats[0].Turn.ManaSpent)
	}
	// "mana_spent" event fires once per call.
	if countExpendEvents(gs, "mana_spent") != 2 {
		t.Errorf("mana_spent event count = %d, want 2 (one per TrackManaSpentThisTurn call)",
			countExpendEvents(gs, "mana_spent"))
	}
}

func TestTrackManaSpentThisTurn_ZeroOrNegativeNoOp(t *testing.T) {
	gs := newExpendGame(t)
	if got := TrackManaSpentThisTurn(gs, 0, 0); got != 0 {
		t.Errorf("track(0) = %d, want 0", got)
	}
	if got := TrackManaSpentThisTurn(gs, 0, -5); got != 0 {
		t.Errorf("track(-5) = %d, want 0", got)
	}
	if gs.Seats[0].Turn.ManaSpent != 0 {
		t.Errorf("Turn.ManaSpent = %d, want 0 (zero/negative is a no-op)",
			gs.Seats[0].Turn.ManaSpent)
	}
}

// ---------------------------------------------------------------------------
// (b) Expend 4 trigger fires at exact threshold
// ---------------------------------------------------------------------------

func TestFireExpendTriggers_ExactThresholdFires(t *testing.T) {
	gs := newExpendGame(t)
	captured, restore := installExpendCapture(t)
	defer restore()

	src := putExpendPerm(gs, 0, newExpendCreature("Watcher 4", 0, 4))
	TrackManaSpentThisTurn(gs, 0, 4)

	if len(*captured) != 1 {
		t.Fatalf("captured count = %d, want 1", len(*captured))
	}
	c := (*captured)[0]
	if c.source != src {
		t.Errorf("ctx[source] mismatch")
	}
	if c.threshold != 4 {
		t.Errorf("ctx[threshold] = %d, want 4", c.threshold)
	}
	if c.total != 4 {
		t.Errorf("ctx[total] = %d, want 4", c.total)
	}
	if c.controller != 0 {
		t.Errorf("ctx[controller] = %d, want 0", c.controller)
	}
}

func TestFireExpendTriggers_BelowThresholdDoesNotFire(t *testing.T) {
	gs := newExpendGame(t)
	_, restore := installExpendCapture(t)
	defer restore()

	putExpendPerm(gs, 0, newExpendCreature("Watcher 4", 0, 4))
	TrackManaSpentThisTurn(gs, 0, 3)

	if countExpendEvents(gs, "expend_trigger") != 0 {
		t.Errorf("expend_trigger fired at 3 mana spent (threshold 4); want 0")
	}
}

func TestFireExpendTriggers_FiresOnceOnCrossing(t *testing.T) {
	gs := newExpendGame(t)
	_, restore := installExpendCapture(t)
	defer restore()

	putExpendPerm(gs, 0, newExpendCreature("Watcher 4", 0, 4))
	// Cross 4 by paying 5 — fires once.
	TrackManaSpentThisTurn(gs, 0, 5)
	// Spend another 3 — total goes 5 → 8; threshold 4 NOT re-fired.
	TrackManaSpentThisTurn(gs, 0, 3)

	if got := countExpendEvents(gs, "expend_trigger"); got != 1 {
		t.Errorf("expend_trigger count = %d, want 1 (threshold crosses once per turn)", got)
	}
}

// ---------------------------------------------------------------------------
// (c) Expend 4 + Expend 8 both fire (8 fires after 8 spent)
// ---------------------------------------------------------------------------

func TestFireExpendTriggers_MultipleThresholdsSameCard(t *testing.T) {
	gs := newExpendGame(t)
	captured, restore := installExpendCapture(t)
	defer restore()

	putExpendPerm(gs, 0, newExpendCreature("Watcher 4+8", 0, 4, 8))

	TrackManaSpentThisTurn(gs, 0, 4) // crosses 4
	if len(*captured) != 1 {
		t.Errorf("after 4 spent, fan-outs = %d, want 1", len(*captured))
	}
	TrackManaSpentThisTurn(gs, 0, 4) // total now 8 — crosses 8
	if len(*captured) != 2 {
		t.Fatalf("after 8 spent, fan-outs = %d, want 2", len(*captured))
	}
	if (*captured)[0].threshold != 4 || (*captured)[1].threshold != 8 {
		t.Errorf("threshold order = [%d, %d], want [4, 8]",
			(*captured)[0].threshold, (*captured)[1].threshold)
	}
}

func TestFireExpendTriggers_SingleLargeSpendCrossesMultiple(t *testing.T) {
	gs := newExpendGame(t)
	captured, restore := installExpendCapture(t)
	defer restore()

	putExpendPerm(gs, 0, newExpendCreature("Watcher 4+8", 0, 4, 8))
	// Spend 8 in one go — both thresholds (4 and 8) cross in this
	// single TrackManaSpentThisTurn call.
	TrackManaSpentThisTurn(gs, 0, 8)

	if len(*captured) != 2 {
		t.Fatalf("single 8-mana spend should cross both thresholds; got %d fan-outs", len(*captured))
	}
}

func TestFireExpendTriggers_TwoCardsAtDifferentThresholds(t *testing.T) {
	gs := newExpendGame(t)
	captured, restore := installExpendCapture(t)
	defer restore()

	a := putExpendPerm(gs, 0, newExpendCreature("Card-A (Expend 4)", 0, 4))
	b := putExpendPerm(gs, 0, newExpendCreature("Card-B (Expend 8)", 0, 8))

	TrackManaSpentThisTurn(gs, 0, 4)
	if len(*captured) != 1 || (*captured)[0].source != a {
		t.Errorf("after 4: expected single fan-out for A, got %d (%v)", len(*captured), *captured)
	}

	TrackManaSpentThisTurn(gs, 0, 4) // total = 8 → B fires
	if len(*captured) != 2 || (*captured)[1].source != b {
		t.Fatalf("after 8: expected second fan-out for B; got %d (%v)", len(*captured), *captured)
	}
}

// ---------------------------------------------------------------------------
// (d) Per-turn reset
// ---------------------------------------------------------------------------

func TestExpend_PerTurnReset(t *testing.T) {
	gs := newExpendGame(t)
	captured, restore := installExpendCapture(t)
	defer restore()

	putExpendPerm(gs, 0, newExpendCreature("Watcher 4", 0, 4))
	TrackManaSpentThisTurn(gs, 0, 5)
	if len(*captured) != 1 {
		t.Fatalf("setup: expected 1 fan-out, got %d", len(*captured))
	}

	// Simulate end-of-turn / new turn reset.
	gs.Seats[0].Turn.Reset()
	if gs.Seats[0].Turn.ManaSpent != 0 {
		t.Fatalf("Turn.Reset should zero ManaSpent; got %d", gs.Seats[0].Turn.ManaSpent)
	}

	// New turn: spend 4 again — Expend 4 should fire again (new crossing).
	TrackManaSpentThisTurn(gs, 0, 4)
	if len(*captured) != 2 {
		t.Errorf("new turn: expected Expend 4 to re-fire after reset; total fan-outs = %d, want 2",
			len(*captured))
	}
}

// ---------------------------------------------------------------------------
// (e) Per-seat isolation
// ---------------------------------------------------------------------------

func TestExpend_PerSeatIsolation(t *testing.T) {
	gs := newExpendGame(t)
	_, restore := installExpendCapture(t)
	defer restore()

	// Seat 0 owns the Expend 4 watcher.
	putExpendPerm(gs, 0, newExpendCreature("My Watcher", 0, 4))
	// Seat 1 also has an Expend 4 watcher.
	putExpendPerm(gs, 1, newExpendCreature("Opp Watcher", 1, 4))

	// Seat 0 spends 4 mana. ONLY their own watcher should fire.
	TrackManaSpentThisTurn(gs, 0, 4)
	if gs.Seats[0].Turn.ManaSpent != 4 {
		t.Errorf("seat 0 Turn.ManaSpent = %d, want 4", gs.Seats[0].Turn.ManaSpent)
	}
	if gs.Seats[1].Turn.ManaSpent != 0 {
		t.Errorf("seat 1 Turn.ManaSpent = %d, want 0 (per-seat isolation)",
			gs.Seats[1].Turn.ManaSpent)
	}
	// Verify only my watcher fired — the opponent watcher does not.
	myFires := 0
	oppFires := 0
	for _, e := range gs.EventLog {
		if e.Kind != "expend_trigger" {
			continue
		}
		switch e.Source {
		case "My Watcher":
			myFires++
		case "Opp Watcher":
			oppFires++
		}
	}
	if myFires != 1 {
		t.Errorf("my watcher fires = %d, want 1", myFires)
	}
	if oppFires != 0 {
		t.Errorf("opponent watcher fires = %d, want 0 (opponent's spend shouldn't ping their own watcher; mine shouldn't either)",
			oppFires)
	}

	// Now seat 1 spends 4. Their watcher fires; mine doesn't re-fire.
	TrackManaSpentThisTurn(gs, 1, 4)
	myFires = 0
	oppFires = 0
	for _, e := range gs.EventLog {
		if e.Kind != "expend_trigger" {
			continue
		}
		switch e.Source {
		case "My Watcher":
			myFires++
		case "Opp Watcher":
			oppFires++
		}
	}
	if myFires != 1 {
		t.Errorf("my watcher fires after seat 1 spends = %d, want still 1", myFires)
	}
	if oppFires != 1 {
		t.Errorf("opp watcher fires after seat 1 spends = %d, want 1", oppFires)
	}
}

// ---------------------------------------------------------------------------
// FireExpendTriggers direct API
// ---------------------------------------------------------------------------

func TestFireExpendTriggers_DirectPrevToNew(t *testing.T) {
	gs := newExpendGame(t)
	captured, restore := installExpendCapture(t)
	defer restore()

	putExpendPerm(gs, 0, newExpendCreature("Watcher", 0, 4, 8))
	// Caller passes prev=3, new=9 — both 4 and 8 should fire.
	FireExpendTriggers(gs, 0, 3, 9)
	if len(*captured) != 2 {
		t.Fatalf("direct fire: expected 2 fan-outs, got %d", len(*captured))
	}
	if (*captured)[0].threshold != 4 || (*captured)[1].threshold != 8 {
		t.Errorf("thresholds = [%d, %d], want [4, 8]",
			(*captured)[0].threshold, (*captured)[1].threshold)
	}
}

func TestFireExpendTriggers_PrevEqualsNewNoOp(t *testing.T) {
	gs := newExpendGame(t)
	_, restore := installExpendCapture(t)
	defer restore()
	putExpendPerm(gs, 0, newExpendCreature("Watcher", 0, 4))
	FireExpendTriggers(gs, 0, 5, 5)
	if countExpendEvents(gs, "expend_trigger") != 0 {
		t.Error("prev == new should be a no-op")
	}
}

func TestFireExpendTriggers_NegativePrevClampedToZero(t *testing.T) {
	gs := newExpendGame(t)
	captured, restore := installExpendCapture(t)
	defer restore()
	putExpendPerm(gs, 0, newExpendCreature("Watcher", 0, 4))
	FireExpendTriggers(gs, 0, -3, 4)
	if len(*captured) != 1 {
		t.Errorf("negative prev clamping: got %d fan-outs, want 1", len(*captured))
	}
}

func TestFireExpendTriggers_PhasedOutPermSkipped(t *testing.T) {
	gs := newExpendGame(t)
	_, restore := installExpendCapture(t)
	defer restore()

	p := putExpendPerm(gs, 0, newExpendCreature("Phased", 0, 4))
	p.PhasedOut = true

	TrackManaSpentThisTurn(gs, 0, 4)
	if countExpendEvents(gs, "expend_trigger") != 0 {
		t.Error("phased-out perm should not fire expend triggers (§702.26)")
	}
}

// ---------------------------------------------------------------------------
// Nil safety
// ---------------------------------------------------------------------------

func TestTrackManaSpentThisTurn_NilSafe(t *testing.T) {
	if got := TrackManaSpentThisTurn(nil, 0, 5); got != 0 {
		t.Errorf("nil gs = %d, want 0", got)
	}
	gs := newExpendGame(t)
	if got := TrackManaSpentThisTurn(gs, 99, 5); got != 0 {
		t.Errorf("invalid seat = %d, want 0", got)
	}
}

func TestFireExpendTriggers_NilSafe(t *testing.T) {
	FireExpendTriggers(nil, 0, 0, 4)
	gs := newExpendGame(t)
	FireExpendTriggers(gs, 99, 0, 4)
}
