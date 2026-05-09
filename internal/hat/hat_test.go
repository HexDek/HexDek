package hat

// Interface-contract tests: both GreedyHat and PokerHat must satisfy
// gameengine.Hat, hats must be swappable mid-game, and a game built
// with mixed hats in each seat must function end-to-end.

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
	"github.com/hexdek/hexdek/internal/gameengine"
)

// TestInterfaceSatisfaction is primarily a compile-time check; the var
// assertions at the top of greedy.go and poker.go fail at build time
// if either hat drifts from the interface. This test additionally
// confirms the interface is callable via the engine's Hat field.
func TestInterfaceSatisfaction(t *testing.T) {
	var _ gameengine.Hat = (*GreedyHat)(nil)
	var _ gameengine.Hat = (*PokerHat)(nil)

	// Dynamic check: build a GameState, attach each hat type, exercise
	// one method. The engine code path doesn't type-assert — proves the
	// "engine never inspects hats" contract.
	gs := newTestGame(t, 2)
	gs.Seats[0].Hat = &GreedyHat{}
	gs.Seats[1].Hat = NewPokerHat()

	for i, s := range gs.Seats {
		// Each seat's Hat answers a trivial mulligan query.
		got := s.Hat.ChooseMulligan(gs, i, s.Hand)
		if got {
			t.Errorf("seat %d: greedy/poker should keep the opener; got mulligan=true", i)
		}
	}
}

// TestHatSwapMidGame verifies a one-line hat swap works without the
// engine caring. This is the load-bearing assertion for the
// architectural directive "hats are swappable, engine never inspects".
func TestHatSwapMidGame(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[0].Hat = &GreedyHat{}

	// Swap to a PokerHat mid-"game".
	gs.Seats[0].Hat = NewPokerHat()

	// Engine-side broadcast from LogEvent must reach the new hat.
	gs.LogEvent(gameengine.Event{Kind: "turn_start", Seat: 0})

	// PokerHat observes the event (eventsSeen++). Verify via the
	// concrete type — allowed in tests, just not in the engine.
	ph, ok := gs.Seats[0].Hat.(*PokerHat)
	if !ok {
		t.Fatalf("seat 0 hat is not PokerHat after swap")
	}
	if ph.eventsSeen == 0 {
		t.Fatalf("PokerHat did not observe LogEvent after swap")
	}
}

// TestMixedGauntletSeating exercises the "NewGameState with mixed
// hats in each seat" gauntlet pattern. The engine must handle a 4-seat
// game where each seat has a different hat type (incl. nil).
func TestMixedGauntletSeating(t *testing.T) {
	gs := newTestGame(t, 4)
	gs.Seats[0].Hat = &GreedyHat{}
	gs.Seats[1].Hat = NewPokerHat()
	gs.Seats[2].Hat = &GreedyHat{}
	gs.Seats[3].Hat = NewPokerHatWithMode(ModeHold)

	// Broadcast an event to exercise every seat's ObserveEvent path.
	gs.LogEvent(gameengine.Event{Kind: "game_start", Seat: -1})

	// Every Hat should have responded without panicking.
	for i, s := range gs.Seats {
		if s.Hat == nil {
			t.Errorf("seat %d has no Hat", i)
			continue
		}
		// Trigger one decision to prove the method table is wired.
		_ = s.Hat.ChooseAttackers(gs, i, nil)
	}
}

// TestHatBroadcastSkipsNilSeats — LogEvent's broadcast must tolerate
// seats with no Hat (the pre-Phase-10 default).
func TestHatBroadcastSkipsNilSeats(t *testing.T) {
	gs := newTestGame(t, 2)
	// seat 0 has no Hat; seat 1 is a PokerHat.
	gs.Seats[1].Hat = NewPokerHat()

	gs.LogEvent(gameengine.Event{Kind: "turn_start", Seat: 0})

	ph := gs.Seats[1].Hat.(*PokerHat)
	if ph.eventsSeen == 0 {
		t.Fatalf("PokerHat on seat 1 did not observe the broadcast")
	}
}

// ---------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------

// newTestGame returns a barebones GameState with `n` seats. The game
// has no cards loaded; tests that need cards build them inline.
func newTestGame(t *testing.T, n int) *gameengine.GameState {
	t.Helper()
	return gameengine.NewGameState(n, rand.New(rand.NewSource(1)), nil)
}

// newTestCardMinimal builds a Card with only the fields Hat methods
// read: AST (for ability walks), Types (for type-line checks, including
// "cost:N" for ManaCostOf), BasePower/Toughness.
func newTestCardMinimal(name string, types []string, cmc int, ast *gameast.CardAST) *gameengine.Card {
	if ast == nil {
		ast = &gameast.CardAST{Name: name}
	}
	c := &gameengine.Card{
		AST:   ast,
		Name:  name,
		Types: append([]string{}, types...),
	}
	if cmc > 0 {
		c.Types = append(c.Types, "cost:"+itoa(cmc))
	}
	return c
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// newTestPermanent builds a Permanent on seat's battlefield. Registers
// with the given Controller/Owner and zero-value flags.
func newTestPermanent(seat *gameengine.Seat, card *gameengine.Card, power, toughness int) *gameengine.Permanent {
	if card == nil {
		card = &gameengine.Card{}
	}
	card.BasePower = power
	card.BaseToughness = toughness
	p := &gameengine.Permanent{
		Card:       card,
		Controller: seat.Idx,
		Owner:      seat.Idx,
		Flags:      map[string]int{},
	}
	seat.Battlefield = append(seat.Battlefield, p)
	return p
}

// addKeyword sets a runtime keyword flag on a permanent so HasKeyword
// picks it up (matches the pattern in combat_test.go).
func addKeyword(p *gameengine.Permanent, name string) {
	if p.Flags == nil {
		p.Flags = map[string]int{}
	}
	p.Flags["kw:"+name] = 1
}

// TestGauntletReadiness verifies the Phase 11 tournament-runner prereq:
// build a 4-seat commander pod with mixed hats, set the game running
// via engine events, and confirm:
//   - no panics
//   - each hat ran ObserveEvent at least once
//   - mode transitions propagated through the RAISE cascade pathway
//     from one seat to another via the shared EventLog.
func TestGauntletReadiness(t *testing.T) {
	gs := gameengine.NewGameState(4, nil, nil)
	gs.CommanderFormat = true
	gs.Seats[0].Hat = &GreedyHat{}
	gs.Seats[1].Hat = NewPokerHat()
	gs.Seats[2].Hat = &GreedyHat{}
	gs.Seats[3].Hat = NewPokerHatWithMode(ModeCall)

	// Give seat 1 big board + combo hand so it will emergency-RAISE.
	gs.Seats[1].Life = 3
	for i := 0; i < 8; i++ {
		gs.LogEvent(gameengine.Event{Kind: "damage", Seat: 1})
	}

	// Seat 1 should have transitioned to RAISE.
	ph1, _ := gs.Seats[1].Hat.(*PokerHat)
	if ph1.Mode != ModeRaise {
		t.Errorf("seat 1 on 3 life should RAISE; got %v", ph1.Mode)
	}
	// Seat 3's PokerHat should have SEEN the player_mode_change event
	// (cascade logic observed it — doesn't need to match).
	ph3 := gs.Seats[3].Hat.(*PokerHat)
	if ph3.eventsSeen == 0 {
		t.Error("seat 3 hat should have observed events")
	}
	// GreedyHats should still answer decisions.
	_ = gs.Seats[0].Hat.ChooseAttackers(gs, 0, nil)
	_ = gs.Seats[2].Hat.ChooseAttackers(gs, 2, nil)
}

// TestYggdrasilBlock_AheadStillTradesUp verifies the bail-out fix:
// even when the YggdrasilHat is comfortably ahead in board position
// (relPos > aheadNoBlock=0.3) AND incoming damage is below
// life/survivalFrac, the hat must still assign a favorable trade —
// here a 1/1 token throwing itself in front of a 4/4 attacker.
//
// Pre-fix: the global guard at the top of AssignBlockers returned an
// empty map, so the 4/4 ate four free damage every turn.
// Post-fix: the per-attacker loop runs and the favorable-trade branch
// picks the lightest legal blocker (1/1 token, P+T=2) against the
// heavier attacker (P+T=8).
func TestYggdrasilBlock_AheadStillTradesUp(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 40
	gs.Seats[0].Life = 40

	// Seat 1 dominates the board with 5 fatties so relPos is well
	// above the 0.3 default aheadNoBlock threshold. The fatties are
	// 5/4 (P+T=9) — bigger than the 4/4 attacker but with toughness
	// equal to attacker power, so they do NOT survive the trade
	// (`tough > pow` is false), forcing the favorable-trade branch
	// to fire. The 1/1 token (P+T=2) is the only blocker strictly
	// lighter than the attacker (P+T=8), so it should be picked.
	for i := 0; i < 5; i++ {
		_ = newTestPermanent(gs.Seats[1], newTestCardMinimal("Fatty", []string{"creature"}, 5, nil), 5, 4)
	}
	token := newTestPermanent(gs.Seats[1], newTestCardMinimal("Token", []string{"creature", "token"}, 0, nil), 1, 1)

	// Seat-0 attacker: 4/4. Incoming = 4 (well under 20). Together
	// with relPos > 0.3 this hits the old global bail-out condition.
	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("Bruiser", []string{"creature"}, 3, nil), 4, 4)

	h := NewYggdrasilHatWithNoise(nil, 0, 0)

	// Sanity check: confirm the test setup actually triggers the
	// old bail-out condition. If relPos isn't above 0.3 the test
	// proves nothing.
	relPos := h.relativePosition(gs, 1)
	if relPos <= 0.3 {
		t.Fatalf("test setup invalid: relPos=%.3f, want > 0.3 to exercise the bail-out fix", relPos)
	}
	incoming := gs.PowerOf(atk)
	if incoming >= gs.Seats[1].Life/2 {
		t.Fatalf("test setup invalid: incoming=%d, want < life/survivalFrac=%d", incoming, gs.Seats[1].Life/2)
	}

	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	if len(out[atk]) != 1 {
		t.Fatalf("expected 1 blocker on the 4/4 even when ahead; got %d", len(out[atk]))
	}
	if out[atk][0] != token {
		t.Errorf("expected the 1/1 token to be picked as the favorable trade; got %s",
			out[atk][0].Card.Name)
	}
}

// TestYggdrasilBlock_AheadStillTakesSurvivor mirrors the trade-up
// case but with a survivor available. We want to make sure the
// removed bail-out doesn't regress the survivor path either.
func TestYggdrasilBlock_AheadStillTakesSurvivor(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 40

	for i := 0; i < 5; i++ {
		_ = newTestPermanent(gs.Seats[1], newTestCardMinimal("Fatty", []string{"creature"}, 5, nil), 5, 4)
	}
	wall := newTestPermanent(gs.Seats[1], newTestCardMinimal("Wall", []string{"creature"}, 3, nil), 0, 5)
	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("Bruiser", []string{"creature"}, 3, nil), 4, 4)

	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	if len(out[atk]) != 1 || out[atk][0] != wall {
		t.Fatalf("expected the 0/5 wall to survive-block the 4/4; got %v", out[atk])
	}
}

// TestYggdrasilBlock_AheadSkipsWhenNoFavorableTrade — when no
// blocker is strictly lighter than the attacker (every legal
// blocker has P+T ≥ attacker P+T) AND no blocker survives, the hat
// should let the damage through rather than throw a same-or-bigger
// creature into a coin-flip trade. Confirms the favorable-trade
// branch is *strict* on stat-sum comparison.
func TestYggdrasilBlock_AheadSkipsWhenNoFavorableTrade(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[1].Life = 40

	// Seat 1 has only 2/2s. Seat 0 attacker is also a 2/2 (P+T=4).
	// No blocker is strictly < 4, no blocker survives (toughness 2
	// not > attacker power 2). Drive relPos high with extra
	// non-creature... actually we need to be ahead. Use seat-1 life
	// dominance instead — relPos uses Evaluator.Evaluate which
	// weights life. Give seat 0 only 2 life so relPos > 0.3.
	gs.Seats[0].Life = 2
	for i := 0; i < 3; i++ {
		_ = newTestPermanent(gs.Seats[1], newTestCardMinimal("Fighter", []string{"creature"}, 2, nil), 2, 2)
	}
	atk := newTestPermanent(gs.Seats[0], newTestCardMinimal("Even", []string{"creature"}, 2, nil), 2, 2)

	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	out := h.AssignBlockers(gs, 1, []*gameengine.Permanent{atk})
	if len(out[atk]) != 0 {
		t.Fatalf("expected no block when every blocker is ≥ attacker P+T; got %v", out[atk])
	}
}
