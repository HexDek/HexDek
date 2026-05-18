package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Test helpers — Threshold (CR §702.72)
// ---------------------------------------------------------------------------

func newThresholdGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(157))
	return NewGameState(2, rng, nil)
}

// fillGraveyard appends `n` filler cards to `seat`'s graveyard so
// ThresholdActive can be exercised with various sizes. Each filler
// is a fresh Card so threshold-aware code that walks the slice
// doesn't see aliasing artifacts.
func fillGraveyard(gs *GameState, seat, n int) {
	for i := 0; i < n; i++ {
		gs.Seats[seat].Graveyard = append(gs.Seats[seat].Graveyard, &Card{
			Name:  "Filler",
			Owner: seat,
			Types: []string{"instant"},
		})
	}
}

// ===========================================================================
// (a) ThresholdActive returns false for 0..6 cards in graveyard
// ===========================================================================

func TestThresholdActive_FalseAtZeroThroughSix(t *testing.T) {
	for n := 0; n <= 6; n++ {
		gs := newThresholdGame(t)
		fillGraveyard(gs, 0, n)
		if ThresholdActive(gs, 0) {
			t.Errorf("ThresholdActive should be false at graveyard size %d", n)
		}
	}
}

// ===========================================================================
// (b) ThresholdActive returns true at 7+
// ===========================================================================

func TestThresholdActive_TrueAtSevenAndAbove(t *testing.T) {
	for _, n := range []int{7, 8, 10, 25, 60} {
		gs := newThresholdGame(t)
		fillGraveyard(gs, 0, n)
		if !ThresholdActive(gs, 0) {
			t.Errorf("ThresholdActive should be true at graveyard size %d", n)
		}
	}
}

// ===========================================================================
// (c) HasThreshold detects 'threshold' in oracle / keyword
// ===========================================================================

func TestHasThreshold_DetectsKeyword(t *testing.T) {
	c := &Card{
		AST: &gameast.CardAST{
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "threshold"},
			},
		},
	}
	if !HasThreshold(c) {
		t.Fatal("HasThreshold should detect modern AST keyword tag")
	}
}

func TestHasThreshold_DetectsOracleTextEmDash(t *testing.T) {
	c := &Card{
		AST: &gameast.CardAST{
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: "Threshold — Werebear gets +3/+3 and has trample as long as seven or more cards are in your graveyard."},
			},
		},
	}
	if !HasThreshold(c) {
		t.Fatal("HasThreshold should detect via oracle-text \"threshold —\" prefix")
	}
}

func TestHasThreshold_DetectsAsciiHyphen(t *testing.T) {
	c := &Card{
		AST: &gameast.CardAST{
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: "Threshold - Add some bonus."},
			},
		},
	}
	if !HasThreshold(c) {
		t.Fatal("HasThreshold should detect ASCII-hyphen form")
	}
}

func TestHasThreshold_NegativeCases(t *testing.T) {
	if HasThreshold(nil) {
		t.Fatal("HasThreshold(nil) should be false")
	}
	if HasThreshold(&Card{AST: &gameast.CardAST{}}) {
		t.Fatal("HasThreshold should be false for an empty AST")
	}
	// Incidental use of the word "threshold" not as the keyword prefix
	// should NOT match.
	c := &Card{
		AST: &gameast.CardAST{
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: "This effect crosses a damage threshold of 5."},
			},
		},
	}
	if HasThreshold(c) {
		t.Fatal("HasThreshold should NOT match incidental uses of the word 'threshold'")
	}
}

// ===========================================================================
// (d) Flipping back to false when graveyard drops below 7
// ===========================================================================

func TestThresholdActive_FlipsOffWhenGraveyardShrinks(t *testing.T) {
	gs := newThresholdGame(t)

	// Start at 8 cards — active.
	fillGraveyard(gs, 0, 8)
	if !ThresholdActive(gs, 0) {
		t.Fatal("setup: should be active with 8 cards")
	}

	// Simulate Tormod's Crypt: exile (clear) the graveyard. Drops to 0.
	gs.Seats[0].Graveyard = nil
	if ThresholdActive(gs, 0) {
		t.Fatal("ThresholdActive should flip off when graveyard is wiped to 0")
	}

	// Simulate partial removal: 7 cards then pop two -> 5, should be off.
	fillGraveyard(gs, 0, 7)
	if !ThresholdActive(gs, 0) {
		t.Fatal("setup: should be active at 7")
	}
	gs.Seats[0].Graveyard = gs.Seats[0].Graveyard[:5]
	if ThresholdActive(gs, 0) {
		t.Fatal("ThresholdActive should flip off at 5 (below threshold)")
	}

	// And turn back on when we cross back to 7.
	fillGraveyard(gs, 0, 2)
	if !ThresholdActive(gs, 0) {
		t.Fatal("ThresholdActive should re-enable when graveyard climbs back to 7")
	}
}

// ===========================================================================
// (e) Per-seat independence
// ===========================================================================

func TestThresholdActive_PerSeatIndependent(t *testing.T) {
	gs := newThresholdGame(t)

	fillGraveyard(gs, 0, 9)
	// Seat 1's graveyard stays empty.

	if !ThresholdActive(gs, 0) {
		t.Error("seat 0 with 9 cards should be active")
	}
	if ThresholdActive(gs, 1) {
		t.Fatal("seat 0's full graveyard must NOT enable seat 1's threshold")
	}

	// Now fill seat 1's grave too.
	fillGraveyard(gs, 1, 7)
	if !ThresholdActive(gs, 1) {
		t.Error("seat 1 with 7 cards should now be active")
	}
	// Seat 0 still active independently.
	if !ThresholdActive(gs, 0) {
		t.Error("seat 0 should remain active independent of seat 1")
	}

	// Wipe seat 0's grave — seat 1 must stay active.
	gs.Seats[0].Graveyard = nil
	if ThresholdActive(gs, 0) {
		t.Error("seat 0 should be inactive after wipe")
	}
	if !ThresholdActive(gs, 1) {
		t.Fatal("seat 1's threshold must NOT depend on seat 0's graveyard")
	}
}

func TestThresholdActive_NilOrInvalidIsFalse(t *testing.T) {
	if ThresholdActive(nil, 0) {
		t.Fatal("ThresholdActive(nil, ...) should be false")
	}
	gs := newThresholdGame(t)
	if ThresholdActive(gs, -1) {
		t.Fatal("ThresholdActive with negative seat should be false")
	}
	if ThresholdActive(gs, 99) {
		t.Fatal("ThresholdActive with out-of-range seat should be false")
	}
}

// ===========================================================================
// ThresholdGraveCount constant value
// ===========================================================================

func TestThresholdGraveCount_IsSeven(t *testing.T) {
	if ThresholdGraveCount != 7 {
		t.Errorf("ThresholdGraveCount: want 7 (CR §702.72a), got %d", ThresholdGraveCount)
	}
}

// ===========================================================================
// ApplyThresholdRider integration with resolveSequence
// ===========================================================================

// thresholdRiderCard builds a card whose AST carries the threshold
// keyword PLUS a tagged threshold-rider Activated whose Effect we can
// observe firing. The cost-extra "threshold_rider" marker is the
// canonical tagging shape.
func thresholdRiderCard(name string, payload gameast.Effect) *Card {
	return &Card{
		Name: name,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "threshold"},
				&gameast.Activated{
					Cost:   gameast.Cost{Extra: []string{"threshold_rider"}},
					Effect: payload,
					Raw:    "Threshold — payload",
				},
			},
		},
	}
}

func TestApplyThresholdRider_FiresWhenActive(t *testing.T) {
	gs := newThresholdGame(t)
	gs.Seats[0].Life = 20
	fillGraveyard(gs, 0, 7) // active

	// Threshold rider draws 1 card. We pre-seat the library so the
	// draw is observable.
	gs.Seats[0].Library = []*Card{{Name: "DrawTarget", Owner: 0, Types: []string{"land"}}}
	card := thresholdRiderCard("Werebear", &gameast.Draw{
		Count:  *gameast.NumInt(1),
		Target: gameast.Filter{Base: "controller"},
	})
	src := &Permanent{
		Card: card, Controller: 0, Owner: 0,
		Timestamp: gs.NextTimestamp(),
		Counters:  map[string]int{}, Flags: map[string]int{},
	}

	handBefore := len(gs.Seats[0].Hand)
	fired := ApplyThresholdRider(gs, src)
	handAfter := len(gs.Seats[0].Hand)

	if !fired {
		t.Fatal("ApplyThresholdRider should fire when threshold is active and rider is present")
	}
	if handAfter-handBefore != 1 {
		t.Errorf("rider payload (draw 1) should run; drew %d", handAfter-handBefore)
	}

	// threshold_rider event logged.
	sawRider := false
	for _, ev := range gs.EventLog {
		if ev.Kind == "threshold_rider" {
			if rule, _ := ev.Details["rule"].(string); rule == "702.72" {
				sawRider = true
				break
			}
		}
	}
	if !sawRider {
		t.Error("expected a threshold_rider event with rule 702.72")
	}
}

func TestApplyThresholdRider_SilentWhenInactive(t *testing.T) {
	gs := newThresholdGame(t)
	gs.Seats[0].Library = []*Card{{Name: "X", Owner: 0}}
	// Empty graveyard — inactive.
	card := thresholdRiderCard("Werebear", &gameast.Draw{
		Count:  *gameast.NumInt(1),
		Target: gameast.Filter{Base: "controller"},
	})
	src := &Permanent{
		Card: card, Controller: 0, Owner: 0,
		Timestamp: gs.NextTimestamp(),
		Counters:  map[string]int{}, Flags: map[string]int{},
	}

	handBefore := len(gs.Seats[0].Hand)
	fired := ApplyThresholdRider(gs, src)
	handAfter := len(gs.Seats[0].Hand)

	if fired {
		t.Fatal("ApplyThresholdRider must not fire when ThresholdActive is false")
	}
	if handAfter != handBefore {
		t.Errorf("no draw should happen; before=%d after=%d", handBefore, handAfter)
	}
	for _, ev := range gs.EventLog {
		if ev.Kind == "threshold_rider" {
			t.Fatal("no threshold_rider event should be emitted when inactive")
		}
	}
}

func TestApplyThresholdRider_NoRiderCardNoOp(t *testing.T) {
	gs := newThresholdGame(t)
	fillGraveyard(gs, 0, 8) // active
	// Plain card — no threshold keyword, no rider.
	src := &Permanent{
		Card: &Card{
			Name: "Plain",
			AST:  &gameast.CardAST{Name: "Plain"},
		},
		Controller: 0, Owner: 0,
		Timestamp: gs.NextTimestamp(),
		Counters:  map[string]int{}, Flags: map[string]int{},
	}
	if ApplyThresholdRider(gs, src) {
		t.Fatal("ApplyThresholdRider on a card without threshold must not fire")
	}
}

func TestApplyThresholdRider_PendingEventWhenNoTaggedEffect(t *testing.T) {
	gs := newThresholdGame(t)
	fillGraveyard(gs, 0, 8)
	// HasThreshold true (oracle), but NO Activated tagged as the rider
	// payload — should log the pending event.
	card := &Card{
		Name: "Werebear",
		AST: &gameast.CardAST{
			Name: "Werebear",
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: "Threshold — Werebear gets +3/+3."},
			},
		},
	}
	src := &Permanent{
		Card: card, Controller: 0, Owner: 0,
		Timestamp: gs.NextTimestamp(),
		Counters:  map[string]int{}, Flags: map[string]int{},
	}

	fired := ApplyThresholdRider(gs, src)
	if !fired {
		t.Fatal("should fire even without tagged payload (logs pending event)")
	}
	sawPending := false
	for _, ev := range gs.EventLog {
		if ev.Kind == "threshold_rider_pending" {
			sawPending = true
			break
		}
	}
	if !sawPending {
		t.Error("expected a threshold_rider_pending event when AST lacks tagged payload")
	}
}

func TestApplyThresholdRider_NilSafe(t *testing.T) {
	if ApplyThresholdRider(nil, nil) {
		t.Fatal("ApplyThresholdRider(nil, nil) should be false")
	}
	gs := newThresholdGame(t)
	if ApplyThresholdRider(gs, nil) {
		t.Fatal("ApplyThresholdRider(gs, nil) should be false")
	}
	if ApplyThresholdRider(gs, &Permanent{}) {
		t.Fatal("ApplyThresholdRider with nil-Card perm should be false")
	}
}

// ===========================================================================
// resolveSequence wiring — outer-sequence rider fires once
// ===========================================================================

func TestResolveSequence_FiresThresholdRiderOnceForNestedSequences(t *testing.T) {
	// Verify the wiring in resolve.go: nested Sequence nodes don't
	// cause the rider to fire multiple times for one resolution.
	gs := newThresholdGame(t)
	fillGraveyard(gs, 0, 8)
	gs.Seats[0].Library = []*Card{
		{Name: "L1", Owner: 0}, {Name: "L2", Owner: 0}, {Name: "L3", Owner: 0},
	}

	card := thresholdRiderCard("Sequence Test", &gameast.Draw{
		Count:  *gameast.NumInt(1),
		Target: gameast.Filter{Base: "controller"},
	})
	src := &Permanent{
		Card: card, Controller: 0, Owner: 0,
		Timestamp: gs.NextTimestamp(),
		Counters:  map[string]int{}, Flags: map[string]int{},
	}

	// Build a nested Sequence: outer { inner { } }. The threshold
	// rider should fire ONCE for the outer resolution.
	inner := &gameast.Sequence{Items: []gameast.Effect{}}
	outer := &gameast.Sequence{Items: []gameast.Effect{inner}}

	resolveSequence(gs, src, outer)

	riderCount := 0
	for _, ev := range gs.EventLog {
		if ev.Kind == "threshold_rider" {
			riderCount++
		}
	}
	if riderCount != 1 {
		t.Errorf("threshold rider should fire exactly once for a nested-sequence resolution; got %d", riderCount)
	}
}
