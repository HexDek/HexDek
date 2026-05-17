package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Demonstrate tests — CR §702.144 (Strixhaven, 2021)
// ---------------------------------------------------------------------------

func newDemonstrateGame(t *testing.T, seats int) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(2144))
	return NewGameState(seats, rng, nil)
}

func newDemonstrateSorcery(name string, owner int) *Card {
	return &Card{
		Name:  name,
		Owner: owner,
		Types: []string{"sorcery"},
		CMC:   3,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "demonstrate"},
			},
		},
	}
}

func newVanillaSorcery(name string, owner int) *Card {
	return &Card{
		Name:  name,
		Owner: owner,
		Types: []string{"sorcery"},
		CMC:   2,
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{},
		},
	}
}

// pushOriginalSpell mints a StackItem for `card` controlled by `seat`
// and pushes it onto the stack so ApplyDemonstrate can layer copies
// above it.
func pushOriginalSpell(gs *GameState, seat int, card *Card) *StackItem {
	item := &StackItem{
		Controller: seat,
		Card:       card,
		Kind:       "spell",
	}
	PushStackItem(gs, item)
	return item
}

// ---------------------------------------------------------------------------
// (d) HasDemonstrate detector
// ---------------------------------------------------------------------------

func TestHasDemonstrate_Positive(t *testing.T) {
	if !HasDemonstrate(newDemonstrateSorcery("Mage Duel", 0)) {
		t.Fatal("HasDemonstrate should be true on a demonstrate card")
	}
}

func TestHasDemonstrate_Negative(t *testing.T) {
	if HasDemonstrate(newVanillaSorcery("Plain", 0)) {
		t.Fatal("HasDemonstrate should be false on a vanilla card")
	}
}

func TestHasDemonstrate_Nil(t *testing.T) {
	if HasDemonstrate(nil) {
		t.Fatal("HasDemonstrate(nil) must be false")
	}
}

// ---------------------------------------------------------------------------
// (a) Opt-in mints 2 copies (controller + chosen opponent)
// ---------------------------------------------------------------------------

func TestApplyDemonstrate_OptInMintsTwoCopies(t *testing.T) {
	gs := newDemonstrateGame(t, 4)
	card := newDemonstrateSorcery("Expressive Iteration", 0)
	original := pushOriginalSpell(gs, 0, card)
	preStackLen := len(gs.Stack)

	count := ApplyDemonstrate(gs, original, 0,
		func() bool { return true },
		func() int { return 2 },
	)
	if count != 2 {
		t.Errorf("created = %d, want 2", count)
	}
	if got := len(gs.Stack) - preStackLen; got != 2 {
		t.Errorf("stack grew by %d, want 2", got)
	}

	// Inspect the two top items.
	top := gs.Stack[len(gs.Stack)-1]
	mid := gs.Stack[len(gs.Stack)-2]
	if !top.IsCopy || !mid.IsCopy {
		t.Errorf("both copies must have StackItem.IsCopy=true (top=%v mid=%v)",
			top.IsCopy, mid.IsCopy)
	}
	// Controllers: one should be 0 (controller), one should be 2 (opponent).
	controllers := map[int]int{}
	for _, it := range []*StackItem{top, mid} {
		controllers[it.Controller]++
	}
	if controllers[0] != 1 || controllers[2] != 1 {
		t.Errorf("expected one copy each for seat 0 and seat 2; got %v", controllers)
	}
}

// ---------------------------------------------------------------------------
// (b) Opt-out mints 0 copies
// ---------------------------------------------------------------------------

func TestApplyDemonstrate_OptOutMintsNoCopies(t *testing.T) {
	gs := newDemonstrateGame(t, 4)
	card := newDemonstrateSorcery("Mage Duel", 0)
	original := pushOriginalSpell(gs, 0, card)
	preLen := len(gs.Stack)

	count := ApplyDemonstrate(gs, original, 0,
		func() bool { return false },
		func() int { return 2 },
	)
	if count != 0 {
		t.Errorf("created = %d, want 0 on opt-out", count)
	}
	if got := len(gs.Stack); got != preLen {
		t.Errorf("stack changed (got %d, want %d) on opt-out", got, preLen)
	}
	// Decline event logged.
	foundDecline := false
	for _, e := range gs.EventLog {
		if e.Kind == "demonstrate_decline" {
			foundDecline = true
		}
	}
	if !foundDecline {
		t.Error("expected demonstrate_decline event on opt-out")
	}
}

func TestApplyDemonstrate_NilOptInTreatedAsOptOut(t *testing.T) {
	gs := newDemonstrateGame(t, 4)
	card := newDemonstrateSorcery("Mage Duel", 0)
	original := pushOriginalSpell(gs, 0, card)
	preLen := len(gs.Stack)

	count := ApplyDemonstrate(gs, original, 0, nil, func() int { return 2 })
	if count != 0 {
		t.Errorf("nil opt-in count = %d, want 0", count)
	}
	if len(gs.Stack) != preLen {
		t.Errorf("nil opt-in must not mint copies; stack grew")
	}
}

// ---------------------------------------------------------------------------
// (c) Opponent selection uses callback
// ---------------------------------------------------------------------------

func TestApplyDemonstrate_OpponentSelectionUsesCallback(t *testing.T) {
	gs := newDemonstrateGame(t, 4)
	card := newDemonstrateSorcery("Mage Duel", 0)
	original := pushOriginalSpell(gs, 0, card)

	// Test that the callback's returned seat is the one whose copy
	// appears on the stack. Use seat 3 as the chosen opponent.
	_ = ApplyDemonstrate(gs, original, 0,
		func() bool { return true },
		func() int { return 3 },
	)

	// The opponent copy should be controlled by seat 3.
	foundOpponentCopy := false
	for _, it := range gs.Stack {
		if !it.IsCopy {
			continue
		}
		if it.Controller != 3 {
			continue
		}
		if role, _ := it.CostMeta["demonstrate_role"].(string); role == "opponent" {
			foundOpponentCopy = true
		}
	}
	if !foundOpponentCopy {
		t.Error("expected opponent-role copy controlled by seat 3 (callback's choice)")
	}
}

func TestApplyDemonstrate_NilOpponentCallbackFallsBackToFirstLivingOpponent(t *testing.T) {
	gs := newDemonstrateGame(t, 4)
	card := newDemonstrateSorcery("Mage Duel", 0)
	original := pushOriginalSpell(gs, 0, card)

	_ = ApplyDemonstrate(gs, original, 0,
		func() bool { return true },
		nil,
	)

	// First living opponent of seat 0 in turn-order from left is seat 1.
	want := gs.LivingOpponents(0)[0]
	if want != 1 {
		t.Fatalf("test scaffold: expected first opponent of seat 0 to be 1, got %d", want)
	}
	foundOpponentCopy := false
	for _, it := range gs.Stack {
		if !it.IsCopy || it.Controller != want {
			continue
		}
		if role, _ := it.CostMeta["demonstrate_role"].(string); role == "opponent" {
			foundOpponentCopy = true
		}
	}
	if !foundOpponentCopy {
		t.Errorf("nil opponent callback: expected first living opponent (seat %d) to receive the copy", want)
	}
}

func TestApplyDemonstrate_InvalidOpponentChoiceFallsBack(t *testing.T) {
	gs := newDemonstrateGame(t, 4)
	card := newDemonstrateSorcery("Mage Duel", 0)
	original := pushOriginalSpell(gs, 0, card)

	// Callback returns the controller's own seat — invalid (must be an
	// opponent). Should fall back to first living opponent.
	_ = ApplyDemonstrate(gs, original, 0,
		func() bool { return true },
		func() int { return 0 },
	)

	fallback := gs.LivingOpponents(0)[0]
	foundFallback := false
	for _, it := range gs.Stack {
		if !it.IsCopy || it.Controller != fallback {
			continue
		}
		if role, _ := it.CostMeta["demonstrate_role"].(string); role == "opponent" {
			foundFallback = true
		}
	}
	if !foundFallback {
		t.Errorf("invalid opponent choice should fall back to first living opponent (seat %d)", fallback)
	}
}

// ---------------------------------------------------------------------------
// (e) Copies are CopyOf the original
// ---------------------------------------------------------------------------

func TestApplyDemonstrate_CopiesAreFlaggedAsCopyOfOriginal(t *testing.T) {
	gs := newDemonstrateGame(t, 4)
	card := newDemonstrateSorcery("Mage Duel", 0)
	original := pushOriginalSpell(gs, 0, card)

	_ = ApplyDemonstrate(gs, original, 0,
		func() bool { return true },
		func() int { return 2 },
	)

	copyCount := 0
	for _, it := range gs.Stack {
		if !it.IsCopy {
			continue
		}
		copyCount++
		// Each copy must be flagged on both StackItem.IsCopy and on
		// Card.IsCopy (the SBA / zone-accounting expectation per CR
		// §704.5e — the copy ceases to exist outside the stack).
		if it.Card == nil || !it.Card.IsCopy {
			t.Errorf("copy stack item missing Card.IsCopy=true (card=%+v)", it.Card)
		}
		// Each copy must reference the original spell's Card pointer via
		// CostMeta["demonstrate_source"] so replay can chain copy back to
		// the trigger source.
		if got, _ := it.CostMeta["demonstrate_source"].(*Card); got != card {
			t.Errorf("CostMeta[demonstrate_source] mismatch (got %v, want original card)", got)
		}
		// IsDemonstrateCopy predicate must return true.
		if !IsDemonstrateCopy(it) {
			t.Error("IsDemonstrateCopy should be true for a demonstrate-minted copy")
		}
		// Same name and Effect as the original.
		if it.Card.Name != card.Name {
			t.Errorf("copy name = %q, want %q", it.Card.Name, card.Name)
		}
		if it.Effect != original.Effect {
			t.Error("copy must share Effect pointer with the original spell")
		}
	}
	if copyCount != 2 {
		t.Errorf("found %d copies on stack, want 2", copyCount)
	}

	// Original spell underneath must NOT be flagged as a copy.
	if original.IsCopy {
		t.Error("original spell should not be flagged IsCopy")
	}
}

func TestApplyDemonstrate_CopiesPreserveTargets(t *testing.T) {
	gs := newDemonstrateGame(t, 4)
	card := newDemonstrateSorcery("Mage Duel", 0)
	original := pushOriginalSpell(gs, 0, card)
	// Stamp original with a synthetic target so we can verify the copy
	// inherits it.
	original.Targets = []Target{{Kind: TargetKindSeat, Seat: 1}}

	_ = ApplyDemonstrate(gs, original, 0,
		func() bool { return true },
		func() int { return 2 },
	)

	copies := 0
	for _, it := range gs.Stack {
		if !it.IsCopy {
			continue
		}
		copies++
		if len(it.Targets) != 1 || it.Targets[0].Seat != 1 ||
			it.Targets[0].Kind != TargetKindSeat {
			t.Errorf("copy targets = %v, want [{Kind:Seat Seat:1}]", it.Targets)
		}
		// Ensure the targets slice is its own backing array (no aliasing
		// to the original — mutating one must not affect the other).
		if len(it.Targets) > 0 && &it.Targets[0] == &original.Targets[0] {
			t.Error("copy targets slice aliases original; expected an independent copy")
		}
	}
	if copies != 2 {
		t.Errorf("copies = %d, want 2", copies)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestApplyDemonstrate_NoLivingOpponentsStillMintsControllerCopy(t *testing.T) {
	gs := newDemonstrateGame(t, 4)
	// Eliminate all opponents.
	for i := 1; i < 4; i++ {
		gs.Seats[i].Lost = true
	}
	card := newDemonstrateSorcery("Mage Duel", 0)
	original := pushOriginalSpell(gs, 0, card)
	preLen := len(gs.Stack)

	count := ApplyDemonstrate(gs, original, 0,
		func() bool { return true },
		func() int { return 1 },
	)
	if count != 1 {
		t.Errorf("created = %d, want 1 (controller copy only)", count)
	}
	if got := len(gs.Stack) - preLen; got != 1 {
		t.Errorf("stack grew by %d, want 1", got)
	}
	// Event marker present.
	foundNoOpp := false
	for _, e := range gs.EventLog {
		if e.Kind == "demonstrate_no_opponent" {
			foundNoOpp = true
		}
	}
	if !foundNoOpp {
		t.Error("expected demonstrate_no_opponent event when all opponents eliminated")
	}
}

func TestApplyDemonstrate_NilStackItem(t *testing.T) {
	gs := newDemonstrateGame(t, 4)
	if got := ApplyDemonstrate(gs, nil, 0, func() bool { return true }, nil); got != 0 {
		t.Errorf("nil spell count = %d, want 0", got)
	}
}

func TestApplyDemonstrate_InvalidControllerSeat(t *testing.T) {
	gs := newDemonstrateGame(t, 4)
	card := newDemonstrateSorcery("Mage Duel", 0)
	original := pushOriginalSpell(gs, 0, card)
	if got := ApplyDemonstrate(gs, original, 99, func() bool { return true }, nil); got != 0 {
		t.Errorf("invalid seat count = %d, want 0", got)
	}
}

func TestIsDemonstrateCopy_Negative(t *testing.T) {
	if IsDemonstrateCopy(nil) {
		t.Error("IsDemonstrateCopy(nil) must be false")
	}
	if IsDemonstrateCopy(&StackItem{}) {
		t.Error("IsDemonstrateCopy on plain item must be false")
	}
}

// ---------------------------------------------------------------------------
// Role tagging — controller copy first, opponent copy second
// ---------------------------------------------------------------------------

func TestApplyDemonstrate_ControllerCopyFirstOpponentSecond(t *testing.T) {
	gs := newDemonstrateGame(t, 4)
	card := newDemonstrateSorcery("Mage Duel", 0)
	original := pushOriginalSpell(gs, 0, card)
	preLen := len(gs.Stack)

	_ = ApplyDemonstrate(gs, original, 0,
		func() bool { return true },
		func() int { return 2 },
	)

	// gs.Stack[preLen] is the first copy pushed (controller); preLen+1 is
	// the second (opponent). PushStackItem appends.
	if len(gs.Stack) < preLen+2 {
		t.Fatalf("stack didn't grow enough: %d < %d", len(gs.Stack), preLen+2)
	}
	first := gs.Stack[preLen]
	second := gs.Stack[preLen+1]
	if role, _ := first.CostMeta["demonstrate_role"].(string); role != "controller" {
		t.Errorf("first pushed copy role = %q, want \"controller\"", role)
	}
	if role, _ := second.CostMeta["demonstrate_role"].(string); role != "opponent" {
		t.Errorf("second pushed copy role = %q, want \"opponent\"", role)
	}
}
