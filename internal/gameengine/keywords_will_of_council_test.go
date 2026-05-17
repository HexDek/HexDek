package gameengine

import (
	"math/rand"
	"testing"
)

// ---------------------------------------------------------------------------
// Will of the Council tests — CR §701.18 (Conspiracy 2014)
// ---------------------------------------------------------------------------

func newWotCGame(t *testing.T, seats int) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(1718))
	return NewGameState(seats, rng, nil)
}

// fixedVote returns a vote callback that always returns `choice`
// regardless of seat.
func fixedVote(choice int) WillOfCouncilVote {
	return func(seat int, opts [2]string) int { return choice }
}

// perSeatVote returns a callback that maps seat -> choice from a slice
// (length matches the number of seats voting). Out-of-range seats
// abstain (return 2).
func perSeatVote(votes map[int]int) WillOfCouncilVote {
	return func(seat int, opts [2]string) int {
		v, ok := votes[seat]
		if !ok {
			return 2
		}
		return v
	}
}

// ---------------------------------------------------------------------------
// (a) Majority A wins
// ---------------------------------------------------------------------------

func TestTallyWillOfCouncil_MajorityA(t *testing.T) {
	gs := newWotCGame(t, 4)
	// 3 seats vote A, 1 votes B.
	cb := perSeatVote(map[int]int{0: 0, 1: 0, 2: 1, 3: 0})

	winner, tally := TallyWillOfCouncil(gs, 0, "carnage", "homage", cb)
	if winner != "carnage" {
		t.Errorf("winner = %q, want \"carnage\"", winner)
	}
	if tally["carnage"] != 3 || tally["homage"] != 1 {
		t.Errorf("tally = %v, want carnage=3 homage=1", tally)
	}
}

// ---------------------------------------------------------------------------
// (b) Majority B wins
// ---------------------------------------------------------------------------

func TestTallyWillOfCouncil_MajorityB(t *testing.T) {
	gs := newWotCGame(t, 4)
	cb := perSeatVote(map[int]int{0: 1, 1: 1, 2: 0, 3: 1})

	winner, tally := TallyWillOfCouncil(gs, 0, "carnage", "homage", cb)
	if winner != "homage" {
		t.Errorf("winner = %q, want \"homage\"", winner)
	}
	if tally["carnage"] != 1 || tally["homage"] != 3 {
		t.Errorf("tally = %v, want carnage=1 homage=3", tally)
	}
}

// ---------------------------------------------------------------------------
// (c) Tie behavior — winner == "" so resolver can decide
// ---------------------------------------------------------------------------

func TestTallyWillOfCouncil_TieReturnsEmptyWinner(t *testing.T) {
	gs := newWotCGame(t, 4)
	cb := perSeatVote(map[int]int{0: 0, 1: 0, 2: 1, 3: 1})

	winner, tally := TallyWillOfCouncil(gs, 0, "carnage", "homage", cb)
	if winner != "" {
		t.Errorf("tied vote winner = %q, want \"\" (resolver decides)", winner)
	}
	if tally["carnage"] != 2 || tally["homage"] != 2 {
		t.Errorf("tally = %v, want 2-2", tally)
	}

	// Sanity: the result event should mark `tied=true` so a resolver
	// scanning the EventLog can branch on it.
	foundTied := false
	for _, e := range gs.EventLog {
		if e.Kind != "will_of_council_result" {
			continue
		}
		if b, _ := e.Details["tied"].(bool); b {
			foundTied = true
		}
	}
	if !foundTied {
		t.Error("expected will_of_council_result event with tied=true")
	}
}

// ---------------------------------------------------------------------------
// (d) 4-player vote order starting from controller
// ---------------------------------------------------------------------------

func TestTallyWillOfCouncil_FourPlayerOrderFromController(t *testing.T) {
	gs := newWotCGame(t, 4)
	// Controller is seat 2. Voting must proceed 2 -> 3 -> 0 -> 1 (APNAP
	// from seat 2's left).
	var visited []int
	cb := func(seat int, opts [2]string) int {
		visited = append(visited, seat)
		return 0
	}

	winner, _ := TallyWillOfCouncil(gs, 2, "A", "B", cb)
	if winner != "A" {
		t.Errorf("winner = %q, want \"A\"", winner)
	}
	want := []int{2, 3, 0, 1}
	if len(visited) != len(want) {
		t.Fatalf("voter order len = %d, want %d (got %v)", len(visited), len(want), visited)
	}
	for i, w := range want {
		if visited[i] != w {
			t.Errorf("voter[%d] = %d, want %d (full = %v)", i, visited[i], w, visited)
		}
	}
}

// ---------------------------------------------------------------------------
// (e) Tally count accurate (matches actual votes cast)
// ---------------------------------------------------------------------------

func TestTallyWillOfCouncil_TallyAccurate(t *testing.T) {
	gs := newWotCGame(t, 4)
	cb := perSeatVote(map[int]int{0: 0, 1: 1, 2: 0, 3: 1})

	_, tally := TallyWillOfCouncil(gs, 0, "A", "B", cb)
	if tally["A"] != 2 || tally["B"] != 2 {
		t.Errorf("tally = %v, want A=2 B=2", tally)
	}

	// Per-voter event log should have one entry per seat (4 total).
	voteEvents := 0
	for _, e := range gs.EventLog {
		if e.Kind == "will_of_council_vote" {
			voteEvents++
		}
	}
	if voteEvents != 4 {
		t.Errorf("will_of_council_vote event count = %d, want 4", voteEvents)
	}
}

// ---------------------------------------------------------------------------
// 2-player game — controller votes first
// ---------------------------------------------------------------------------

func TestTallyWillOfCouncil_TwoPlayerControllerFirst(t *testing.T) {
	gs := newWotCGame(t, 2)
	var visited []int
	cb := func(seat int, opts [2]string) int {
		visited = append(visited, seat)
		return 0
	}

	_, _ = TallyWillOfCouncil(gs, 1, "A", "B", cb)
	if len(visited) != 2 || visited[0] != 1 || visited[1] != 0 {
		t.Errorf("voter order = %v, want [1 0] (controller first)", visited)
	}
}

// ---------------------------------------------------------------------------
// Default callback (nil) → random vote via gs.Rng, but result is
// deterministic when gs.Rng is seeded.
// ---------------------------------------------------------------------------

func TestTallyWillOfCouncil_NilCallbackUsesRng(t *testing.T) {
	// Seed RNG explicitly so the test is deterministic.
	gs1 := NewGameState(4, rand.New(rand.NewSource(42)), nil)
	winner1, _ := TallyWillOfCouncil(gs1, 0, "A", "B", nil)

	gs2 := NewGameState(4, rand.New(rand.NewSource(42)), nil)
	winner2, _ := TallyWillOfCouncil(gs2, 0, "A", "B", nil)

	if winner1 != winner2 {
		t.Errorf("same-seed nil-callback runs disagree: %q vs %q", winner1, winner2)
	}

	// Different seeds should at least sometimes produce different
	// outcomes. With only 4 voters there's a small chance both seeds
	// produce identical sequences — try a couple if the first matches.
	saw := map[string]struct{}{winner1: {}}
	for seed := int64(1); seed < 50 && len(saw) < 2; seed++ {
		gs := NewGameState(4, rand.New(rand.NewSource(seed)), nil)
		w, _ := TallyWillOfCouncil(gs, 0, "A", "B", nil)
		saw[w] = struct{}{}
	}
	if len(saw) < 2 {
		t.Logf("warning: 50 seeds all produced the same winner — RNG distribution may be skewed (saw %v)", saw)
	}
}

// ---------------------------------------------------------------------------
// Lost (eliminated) opponents are skipped
// ---------------------------------------------------------------------------

func TestTallyWillOfCouncil_SkipsEliminatedSeats(t *testing.T) {
	gs := newWotCGame(t, 4)
	gs.Seats[2].Lost = true

	var visited []int
	cb := func(seat int, opts [2]string) int {
		visited = append(visited, seat)
		return 0
	}

	_, _ = TallyWillOfCouncil(gs, 0, "A", "B", cb)
	if len(visited) != 3 {
		t.Errorf("voter count = %d, want 3 (seat 2 eliminated)", len(visited))
	}
	for _, seat := range visited {
		if seat == 2 {
			t.Errorf("eliminated seat 2 should not be polled (visited=%v)", visited)
		}
	}
}

// ---------------------------------------------------------------------------
// Out-of-range vote = abstention, doesn't tally
// ---------------------------------------------------------------------------

func TestTallyWillOfCouncil_AbstentionDoesNotTally(t *testing.T) {
	gs := newWotCGame(t, 4)
	// Seats 0, 1 vote A. Seat 2 abstains (returns 5). Seat 3 votes B.
	cb := perSeatVote(map[int]int{0: 0, 1: 0, 2: 5, 3: 1})

	winner, tally := TallyWillOfCouncil(gs, 0, "A", "B", cb)
	if winner != "A" {
		t.Errorf("winner = %q, want \"A\"", winner)
	}
	if tally["A"] != 2 || tally["B"] != 1 {
		t.Errorf("tally = %v, want A=2 B=1 (seat 2 abstained)", tally)
	}
	// An abstention event must be logged.
	foundAbstain := false
	for _, e := range gs.EventLog {
		if e.Kind == "will_of_council_abstain" && e.Seat == 2 {
			foundAbstain = true
		}
	}
	if !foundAbstain {
		t.Error("expected will_of_council_abstain event for seat 2")
	}
}

// ---------------------------------------------------------------------------
// Nil game / invalid controller seat — safe defaults
// ---------------------------------------------------------------------------

func TestTallyWillOfCouncil_NilSafe(t *testing.T) {
	winner, tally := TallyWillOfCouncil(nil, 0, "A", "B", fixedVote(0))
	if winner != "" {
		t.Errorf("nil gs winner = %q, want \"\"", winner)
	}
	if tally["A"] != 0 || tally["B"] != 0 {
		t.Errorf("nil gs tally non-zero: %v", tally)
	}
}

func TestTallyWillOfCouncil_InvalidControllerSeat(t *testing.T) {
	gs := newWotCGame(t, 4)
	winner, tally := TallyWillOfCouncil(gs, 99, "A", "B", fixedVote(0))
	if winner != "" {
		t.Errorf("invalid seat winner = %q, want \"\"", winner)
	}
	if tally["A"] != 0 || tally["B"] != 0 {
		t.Errorf("invalid seat tally non-zero: %v", tally)
	}
}

// ---------------------------------------------------------------------------
// Per-vote event details carry the chosen option
// ---------------------------------------------------------------------------

func TestTallyWillOfCouncil_PerVoteEventCarriesChoice(t *testing.T) {
	gs := newWotCGame(t, 4)
	cb := perSeatVote(map[int]int{0: 0, 1: 1, 2: 0, 3: 1})

	_, _ = TallyWillOfCouncil(gs, 0, "carnage", "homage", cb)

	wantChoice := map[int]string{0: "carnage", 1: "homage", 2: "carnage", 3: "homage"}
	for _, e := range gs.EventLog {
		if e.Kind != "will_of_council_vote" {
			continue
		}
		expected, ok := wantChoice[e.Seat]
		if !ok {
			t.Errorf("unexpected vote event for seat %d", e.Seat)
			continue
		}
		got, _ := e.Details["choice"].(string)
		if got != expected {
			t.Errorf("seat %d choice in event = %q, want %q", e.Seat, got, expected)
		}
	}
}
