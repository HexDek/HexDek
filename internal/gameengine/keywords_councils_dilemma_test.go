package gameengine

import (
	"math/rand"
	"testing"
)

// ---------------------------------------------------------------------------
// Council's Dilemma tests — CR §701.20
// ---------------------------------------------------------------------------

func newDilemmaGame4P(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(4, rand.New(rand.NewSource(50)), nil)
}

func newDilemmaGame2P(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(2, rand.New(rand.NewSource(50)), nil)
}

// fixedVoter returns a CouncilsDilemmaVoter that hands back the index
// stored in `votes` keyed by seat. Defaults to abstain (-1) for seats
// not in the map.
func fixedVoter(votes map[int]int) CouncilsDilemmaVoter {
	return func(seat int, options []string) int {
		if idx, ok := votes[seat]; ok {
			return idx
		}
		return -1
	}
}

// ---------------------------------------------------------------------------
// (a) All 4 players vote the same option = 4 votes on that option
// ---------------------------------------------------------------------------

func TestCouncilsDilemma_AllPlayersSameOption(t *testing.T) {
	gs := newDilemmaGame4P(t)
	options := []string{"Pirate", "Treasure"}

	// Every seat votes index 0 (Pirate).
	voter := fixedVoter(map[int]int{0: 0, 1: 0, 2: 0, 3: 0})

	tally := TallyCouncilsDilemma(gs, 0, options, voter)
	if tally == nil {
		t.Fatal("TallyCouncilsDilemma returned nil")
	}
	if tally["Pirate"] != 4 {
		t.Fatalf("Pirate votes = %d, want 4 (all 4 players)", tally["Pirate"])
	}
	if tally["Treasure"] != 0 {
		t.Fatalf("Treasure votes = %d, want 0", tally["Treasure"])
	}
	// Both options must appear in the tally map.
	if _, ok := tally["Treasure"]; !ok {
		t.Fatal("zero-vote option must still be present in tally map")
	}
}

// ---------------------------------------------------------------------------
// (b) Split vote 2-2 distributes correctly
// ---------------------------------------------------------------------------

func TestCouncilsDilemma_SplitVoteDistributes(t *testing.T) {
	gs := newDilemmaGame4P(t)
	options := []string{"Pirate", "Treasure"}

	// 0 and 2 → Pirate; 1 and 3 → Treasure.
	voter := fixedVoter(map[int]int{0: 0, 1: 1, 2: 0, 3: 1})

	tally := TallyCouncilsDilemma(gs, 0, options, voter)
	if tally["Pirate"] != 2 {
		t.Fatalf("Pirate votes = %d, want 2", tally["Pirate"])
	}
	if tally["Treasure"] != 2 {
		t.Fatalf("Treasure votes = %d, want 2", tally["Treasure"])
	}
}

func TestCouncilsDilemma_ThreeWayVote(t *testing.T) {
	gs := newDilemmaGame4P(t)
	options := []string{"Draw", "Tokens", "Damage"}

	voter := fixedVoter(map[int]int{0: 0, 1: 1, 2: 2, 3: 1})
	tally := TallyCouncilsDilemma(gs, 0, options, voter)

	if tally["Draw"] != 1 {
		t.Fatalf("Draw = %d, want 1", tally["Draw"])
	}
	if tally["Tokens"] != 2 {
		t.Fatalf("Tokens = %d, want 2", tally["Tokens"])
	}
	if tally["Damage"] != 1 {
		t.Fatalf("Damage = %d, want 1", tally["Damage"])
	}
}

// ---------------------------------------------------------------------------
// (c) Effect callback fires once per non-zero option
// ---------------------------------------------------------------------------

func TestCouncilsDilemma_EffectFiresOncePerNonzeroOption(t *testing.T) {
	gs := newDilemmaGame4P(t)
	options := []string{"Pirate", "Treasure", "Boon"}

	// 0,1,3 vote Pirate; 2 votes Treasure; nobody votes Boon.
	voter := fixedVoter(map[int]int{0: 0, 1: 0, 2: 1, 3: 0})
	tally := TallyCouncilsDilemma(gs, 0, options, voter)

	type fire struct {
		option string
		votes  int
	}
	var fires []fire
	ApplyCouncilsDilemma(gs, options, tally, func(opt string, votes int) {
		fires = append(fires, fire{opt, votes})
	})

	if len(fires) != 2 {
		t.Fatalf("expected 2 effect fires (Pirate + Treasure), got %d", len(fires))
	}
	// Iteration order follows the options slice — Pirate first.
	if fires[0] != (fire{"Pirate", 3}) {
		t.Fatalf("fires[0] = %+v, want {Pirate, 3}", fires[0])
	}
	if fires[1] != (fire{"Treasure", 1}) {
		t.Fatalf("fires[1] = %+v, want {Treasure, 1}", fires[1])
	}
	// Boon received 0 votes — effect must NOT have fired for it.
	for _, f := range fires {
		if f.option == "Boon" {
			t.Fatal("Boon got 0 votes; effect callback must not fire")
		}
	}
}

func TestCouncilsDilemma_ZeroTotalTallyFiresNothing(t *testing.T) {
	gs := newDilemmaGame4P(t)
	options := []string{"A", "B"}
	// All abstain.
	voter := fixedVoter(map[int]int{})
	tally := TallyCouncilsDilemma(gs, 0, options, voter)

	fired := false
	ApplyCouncilsDilemma(gs, options, tally, func(opt string, votes int) {
		fired = true
	})
	if fired {
		t.Fatal("effect must not fire when every option got 0 votes")
	}
}

// ---------------------------------------------------------------------------
// (d) Vote order starts from controller
// ---------------------------------------------------------------------------

func TestCouncilsDilemma_VoteOrderStartsFromController(t *testing.T) {
	gs := newDilemmaGame4P(t)
	options := []string{"X", "Y"}

	var voteOrder []int
	voter := func(seat int, options []string) int {
		voteOrder = append(voteOrder, seat)
		return 0
	}
	// Controller = seat 2.
	TallyCouncilsDilemma(gs, 2, options, voter)

	want := []int{2, 3, 0, 1}
	if len(voteOrder) != len(want) {
		t.Fatalf("vote order length = %d, want %d", len(voteOrder), len(want))
	}
	for i := range want {
		if voteOrder[i] != want[i] {
			t.Fatalf("vote order[%d] = %d, want %d (full: %v want %v)",
				i, voteOrder[i], want[i], voteOrder, want)
		}
	}
}

func TestCouncilsDilemma_ControllerIsSeatZeroDefaultOrder(t *testing.T) {
	gs := newDilemmaGame4P(t)
	options := []string{"X", "Y"}

	var voteOrder []int
	voter := func(seat int, options []string) int {
		voteOrder = append(voteOrder, seat)
		return 0
	}
	TallyCouncilsDilemma(gs, 0, options, voter)
	want := []int{0, 1, 2, 3}
	if len(voteOrder) != 4 {
		t.Fatalf("vote order length = %d, want 4", len(voteOrder))
	}
	for i := range want {
		if voteOrder[i] != want[i] {
			t.Fatalf("vote order[%d] = %d, want %d", i, voteOrder[i], want[i])
		}
	}
}

// ---------------------------------------------------------------------------
// Eliminated seats are skipped (CR §800.4)
// ---------------------------------------------------------------------------

func TestCouncilsDilemma_LostSeatSkipped(t *testing.T) {
	gs := newDilemmaGame4P(t)
	// Seat 1 has been eliminated.
	gs.Seats[1].Lost = true

	options := []string{"A", "B"}

	var voted []int
	voter := func(seat int, _ []string) int {
		voted = append(voted, seat)
		return 0
	}
	tally := TallyCouncilsDilemma(gs, 0, options, voter)

	// Only 3 votes should have been cast (1 abstained because skipped).
	if len(voted) != 3 {
		t.Fatalf("expected 3 active voters, got %d (seats: %v)", len(voted), voted)
	}
	for _, s := range voted {
		if s == 1 {
			t.Fatal("eliminated seat must not be polled for a vote")
		}
	}
	if tally["A"] != 3 {
		t.Fatalf("A = %d, want 3 (lost seat skipped)", tally["A"])
	}
}

// ---------------------------------------------------------------------------
// Abstention via out-of-range callback index
// ---------------------------------------------------------------------------

func TestCouncilsDilemma_OutOfRangeIndexAbstains(t *testing.T) {
	gs := newDilemmaGame4P(t)
	options := []string{"A", "B"}

	// Seats 0,1 vote A (index 0); seats 2,3 abstain via index 99.
	voter := fixedVoter(map[int]int{0: 0, 1: 0, 2: 99, 3: -1})
	tally := TallyCouncilsDilemma(gs, 0, options, voter)

	if tally["A"] != 2 {
		t.Fatalf("A = %d, want 2 (the only valid votes)", tally["A"])
	}
	if tally["B"] != 0 {
		t.Fatalf("B = %d, want 0", tally["B"])
	}
}

// ---------------------------------------------------------------------------
// 2-player edge — Council's Dilemma still works at 2P
// ---------------------------------------------------------------------------

func TestCouncilsDilemma_TwoPlayerTable(t *testing.T) {
	gs := newDilemmaGame2P(t)
	options := []string{"A", "B"}
	voter := fixedVoter(map[int]int{0: 0, 1: 1})
	tally := TallyCouncilsDilemma(gs, 0, options, voter)
	if tally["A"] != 1 || tally["B"] != 1 {
		t.Fatalf("2P tally = %v, want A:1 B:1", tally)
	}
}

// ---------------------------------------------------------------------------
// Nil / invalid-input safety
// ---------------------------------------------------------------------------

func TestCouncilsDilemma_NilGameReturnsNil(t *testing.T) {
	if TallyCouncilsDilemma(nil, 0, []string{"A"}, fixedVoter(nil)) != nil {
		t.Fatal("nil game must return nil tally")
	}
}

func TestCouncilsDilemma_NilCallbackReturnsNil(t *testing.T) {
	gs := newDilemmaGame4P(t)
	if TallyCouncilsDilemma(gs, 0, []string{"A"}, nil) != nil {
		t.Fatal("nil callback must return nil tally")
	}
}

func TestCouncilsDilemma_EmptyOptionsReturnsNil(t *testing.T) {
	gs := newDilemmaGame4P(t)
	if TallyCouncilsDilemma(gs, 0, nil, fixedVoter(map[int]int{})) != nil {
		t.Fatal("empty options must return nil tally")
	}
}

func TestCouncilsDilemma_InvalidControllerReturnsNil(t *testing.T) {
	gs := newDilemmaGame4P(t)
	if TallyCouncilsDilemma(gs, -1, []string{"A"}, fixedVoter(nil)) != nil {
		t.Fatal("invalid controller seat must return nil tally")
	}
	if TallyCouncilsDilemma(gs, 99, []string{"A"}, fixedVoter(nil)) != nil {
		t.Fatal("out-of-range controller seat must return nil tally")
	}
}

func TestApplyCouncilsDilemma_NilSafe(t *testing.T) {
	gs := newDilemmaGame4P(t)
	options := []string{"A"}
	tally := map[string]int{"A": 1}
	// All nil-arg combinations must not panic.
	ApplyCouncilsDilemma(nil, options, tally, func(string, int) {})
	ApplyCouncilsDilemma(gs, options, nil, func(string, int) {})
	ApplyCouncilsDilemma(gs, options, tally, nil)
}
