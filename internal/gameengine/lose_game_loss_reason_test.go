package gameengine

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// resolveLoseGame — LossReason invariant (Phage the Untouchable + generics)
// ---------------------------------------------------------------------------
//
// Regression cover for the R36 goldilocks audit (docs/goldilocks-r36-report.md):
//
//   [lose_game] TurnStructure: active seat 0 is Lost but life is 20
//   with no LossReason
//
// The previous resolveLoseGame implementation flipped seat.Lost = true
// without populating seat.LossReason, so any TurnStructure invariant
// running between the lose-game effect and a life-state SBA pass would
// flag the inconsistency. These tests pin the fix:
//   - LossReason is always set when Lost is set
//   - Phage-style effects (named card) tag with the source card name
//   - Generic effects (no source) fall back to "card_effect"
//   - Platinum Angel-style replacements cancel the loss
//   - the TurnStructure invariant passes post-resolution

func newLoseGameTest(t *testing.T, seats int) *GameState {
	t.Helper()
	return NewGameState(seats, rand.New(rand.NewSource(7)), nil)
}

// putSourcePerm mints a battlefield permanent that we can pass as the
// lose-game effect's source (mimics Phage's ETB context where her own
// permanent is the source emitting the LoseGame effect).
func putSourcePerm(gs *GameState, seat int, name string) *Permanent {
	card := &Card{
		Name:          name,
		Owner:         seat,
		Types:         []string{"creature", "legendary"},
		BasePower:     4,
		BaseToughness: 4,
	}
	p := &Permanent{
		Card:       card,
		Controller: seat,
		Owner:      seat,
		Timestamp:  gs.NextTimestamp(),
		Counters:   map[string]int{},
		Flags:      map[string]int{},
	}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, p)
	return p
}

// loseGameForController builds the LoseGame effect Phage's ETB
// generates: "you lose the game" → target is the controller.
func loseGameForController() *gameast.LoseGame {
	return &gameast.LoseGame{
		Target: gameast.Filter{
			Base:       "player",
			Quantifier: "you",
			Targeted:   false,
		},
	}
}

// ---------------------------------------------------------------------------
// (a) Phage causes loss with LossReason set
// ---------------------------------------------------------------------------

func TestResolveLoseGame_PhageStampsLossReasonWithSourceName(t *testing.T) {
	gs := newLoseGameTest(t, 2)
	gs.Active = 0
	phage := putSourcePerm(gs, 0, "Phage the Untouchable")

	// Force the target picker to resolve to controller seat 0. The Filter
	// {Base:"player", Quantifier:"you"} doesn't always pick controller in
	// the PickTarget defaults — drive the effect with an explicit Self-
	// targeted shape to ensure seat 0 is hit. We construct the LoseGame
	// node and pass `phage` as the source: PickTarget on Base="player"
	// without Targeted defaults to the source's controller (seat 0).
	resolveLoseGame(gs, phage, loseGameForController())

	s := gs.Seats[0]
	if !s.Lost {
		t.Fatal("seat 0 should be Lost after Phage's lose-game effect")
	}
	if s.LossReason == "" {
		t.Errorf("LossReason must not be empty (R36 invariant); got %q", s.LossReason)
	}
	// Tag should carry Phage's name so post-mortems can attribute.
	if !strings.Contains(s.LossReason, "Phage the Untouchable") {
		t.Errorf("LossReason should contain source name; got %q", s.LossReason)
	}
	if !strings.HasPrefix(s.LossReason, "card_effect") {
		t.Errorf("LossReason should be prefixed with \"card_effect\"; got %q", s.LossReason)
	}
}

// ---------------------------------------------------------------------------
// (b) Other lose-game effects (no source) also tag
// ---------------------------------------------------------------------------

func TestResolveLoseGame_GenericTagWithoutSource(t *testing.T) {
	gs := newLoseGameTest(t, 2)
	gs.Active = 1
	// Build an unnamed source for the lose-game effect — sourceName()
	// returns "" when src is nil. We pass a permanent with no Card to
	// hit the no-name fallback explicitly.
	srcless := &Permanent{
		Controller: 1,
		Owner:      1,
	}
	gs.Seats[1].Battlefield = append(gs.Seats[1].Battlefield, srcless)

	resolveLoseGame(gs, srcless, loseGameForController())

	s := gs.Seats[1]
	if !s.Lost {
		t.Fatal("seat 1 should be Lost after lose-game effect")
	}
	if s.LossReason != "card_effect" {
		t.Errorf("LossReason fallback = %q, want \"card_effect\"", s.LossReason)
	}
}

func TestResolveLoseGame_LosesGameEvent_HasReasonInDetails(t *testing.T) {
	gs := newLoseGameTest(t, 2)
	gs.Active = 0
	src := putSourcePerm(gs, 0, "Door to Nothingness")
	resolveLoseGame(gs, src, loseGameForController())

	found := false
	for _, e := range gs.EventLog {
		if e.Kind != "lose_game" {
			continue
		}
		if r, _ := e.Details["reason"].(string); strings.Contains(r, "Door to Nothingness") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected lose_game event with reason containing source name; log = %+v",
			gs.EventLog)
	}
}

// ---------------------------------------------------------------------------
// (c) Post-loss TurnStructure invariant passes
// ---------------------------------------------------------------------------

func TestResolveLoseGame_PostLossInvariantPasses(t *testing.T) {
	gs := newLoseGameTest(t, 2)
	gs.Active = 0
	gs.Phase = "precombat_main"
	gs.Step = "main"
	// Seat 0's life is the default starting value (>0) — this is the
	// exact pre-bug scenario that the R36 audit caught.
	startLife := gs.Seats[0].Life
	if startLife <= 0 {
		t.Fatalf("test scaffold: seat 0 should start with positive life; got %d", startLife)
	}

	phage := putSourcePerm(gs, 0, "Phage the Untouchable")
	resolveLoseGame(gs, phage, loseGameForController())

	// Life unchanged.
	if gs.Seats[0].Life != startLife {
		t.Errorf("life changed unexpectedly: %d → %d", startLife, gs.Seats[0].Life)
	}
	// Lost is true.
	if !gs.Seats[0].Lost {
		t.Fatal("seat 0 should be Lost")
	}
	// LossReason set.
	if gs.Seats[0].LossReason == "" {
		t.Fatal("LossReason should be set")
	}
	// TurnStructure invariant — the exact one that fired in R36 —
	// must now PASS.
	if err := checkTurnStructure(gs); err != nil {
		t.Errorf("TurnStructure invariant should pass after lose-game; got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Phage on a non-active seat — Lost still set, LossReason populated
// ---------------------------------------------------------------------------

func TestResolveLoseGame_NonActiveSeatTaggedToo(t *testing.T) {
	gs := newLoseGameTest(t, 2)
	gs.Active = 0
	// Source on seat 1 (not active player). Phage-style "you lose"
	// targets the controller, which is seat 1.
	src := putSourcePerm(gs, 1, "Phage the Untouchable")
	resolveLoseGame(gs, src, loseGameForController())

	if !gs.Seats[1].Lost {
		t.Fatal("seat 1 should be Lost")
	}
	if gs.Seats[1].LossReason == "" {
		t.Errorf("LossReason on non-active seat 1 should also be populated; got %q",
			gs.Seats[1].LossReason)
	}
	// Active seat (0) untouched.
	if gs.Seats[0].Lost {
		t.Error("active seat 0 should not be Lost — Phage's effect targets the controller, not the active player")
	}
}

// ---------------------------------------------------------------------------
// CR §614 — would_lose_game replacement (Platinum-Angel style) cancels
// ---------------------------------------------------------------------------

func TestResolveLoseGame_PlatinumAngelStyleReplacementCancels(t *testing.T) {
	gs := newLoseGameTest(t, 2)
	gs.Active = 0
	phage := putSourcePerm(gs, 0, "Phage the Untouchable")

	// Register a synthetic would_lose_game replacement: cancel any
	// would_lose_game event targeting seat 0 (Platinum Angel proxy).
	gs.RegisterReplacement(&ReplacementEffect{
		EventType:      "would_lose_game",
		HandlerID:      "platinum_angel_test",
		ControllerSeat: 0,
		Applies: func(gs *GameState, ev *ReplEvent) bool {
			return ev.TargetSeat == 0
		},
		ApplyFn: func(gs *GameState, ev *ReplEvent) {
			ev.Cancelled = true
		},
	})

	resolveLoseGame(gs, phage, loseGameForController())

	if gs.Seats[0].Lost {
		t.Error("seat 0 should NOT be Lost when a would_lose_game replacement cancels")
	}
	if gs.Seats[0].LossReason != "" {
		t.Errorf("LossReason should remain empty when replacement cancels; got %q",
			gs.Seats[0].LossReason)
	}
	// "lose_game_replaced" event should be logged.
	found := false
	for _, e := range gs.EventLog {
		if e.Kind == "lose_game_replaced" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a lose_game_replaced event; log = %+v", gs.EventLog)
	}
}

// ---------------------------------------------------------------------------
// LossReason set BEFORE Lost (invariant safety under concurrent reads)
// ---------------------------------------------------------------------------
//
// The fix stamps LossReason first, then Lost. We can't directly test
// ordering without instrumentation, but we can pin the invariant
// contract: after the call returns, BOTH fields are consistent.

func TestResolveLoseGame_LostAndLossReasonAreConsistent(t *testing.T) {
	gs := newLoseGameTest(t, 4)
	for i := 0; i < 4; i++ {
		gs.Active = i
		src := putSourcePerm(gs, i, "Multi-Seat Loss Source")
		resolveLoseGame(gs, src, loseGameForController())
		s := gs.Seats[i]
		if s.Lost && s.LossReason == "" {
			t.Errorf("seat %d: Lost=true but LossReason empty (the R36 invariant violation)", i)
		}
		if !s.Lost && s.LossReason != "" {
			t.Errorf("seat %d: LossReason set without Lost (inverse inconsistency); got %q", i, s.LossReason)
		}
	}
}

// ---------------------------------------------------------------------------
// Nil safety
// ---------------------------------------------------------------------------

func TestResolveLoseGame_NilSafe(t *testing.T) {
	// nil gs is a safe no-op.
	resolveLoseGame(nil, nil, loseGameForController())
	// nil source + nil effect targets — exercise the empty-target branch.
	gs := newLoseGameTest(t, 2)
	resolveLoseGame(gs, nil, &gameast.LoseGame{
		Target: gameast.Filter{Base: "nothing"},
	})
	// No seat should be Lost.
	for i, s := range gs.Seats {
		if s != nil && s.Lost {
			t.Errorf("seat %d should not be Lost on nil-source no-target effect", i)
		}
	}
}
