package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Tempting offer tests — CR §702.74
// ---------------------------------------------------------------------------

func newTemptingOfferGame(t *testing.T, seats int) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(99))
	gs := NewGameState(seats, rng, nil)
	// Seed every library with stub cards so resolveDraw has something
	// to take off the top.
	for i := range gs.Seats {
		for k := 0; k < 10; k++ {
			gs.Seats[i].Library = append(gs.Seats[i].Library, &Card{
				Name:  "Filler",
				Owner: i,
				Types: []string{"sorcery"},
				AST:   &gameast.CardAST{Name: "Filler"},
			})
		}
	}
	return gs
}

// drawOne is the canonical "draw one card with no target filter — defaults
// to source's controller" effect. ResolveEffect falls back to
// src.Controller when the filter is empty, which is the path Tempting
// offer relies on.
func drawOneEffect() gameast.Effect {
	return &gameast.Draw{Count: gameast.NumberOrRef{IsInt: true, Int: 1}}
}

// gainLifeOneEffect — same pattern as drawOneEffect, lets us assert
// per-seat life changes when the copy resolves for an opponent.
func gainLifeOneEffect() gameast.Effect {
	return &gameast.GainLife{Amount: gameast.NumberOrRef{IsInt: true, Int: 1}}
}

// ---------------------------------------------------------------------------
// (a) None accept → only the controller resolves a copy.
// ---------------------------------------------------------------------------

func TestTemptingOffer_NoneAccept_OnlyControllerResolves(t *testing.T) {
	gs := newTemptingOfferGame(t, 4)
	startHand := make([]int, len(gs.Seats))
	for i := range gs.Seats {
		startHand[i] = len(gs.Seats[i].Hand)
	}

	decline := func(seat int) bool { return false }
	recipients := ResolveTemptingOffer(gs, 2, drawOneEffect(), decline)

	if len(recipients) != 1 {
		t.Fatalf("recipients = %v, want exactly [2] (controller-only)", recipients)
	}
	if recipients[0] != 2 {
		t.Fatalf("recipients[0] = %d, want 2", recipients[0])
	}
	// Controller drew one; everyone else unchanged.
	for i, seat := range gs.Seats {
		want := startHand[i]
		if i == 2 {
			want++
		}
		if got := len(seat.Hand); got != want {
			t.Fatalf("seat %d hand size = %d, want %d", i, got, want)
		}
	}
}

func TestTemptingOffer_NilCallback_EquivalentToNoneAccept(t *testing.T) {
	gs := newTemptingOfferGame(t, 4)
	startHand := make([]int, len(gs.Seats))
	for i := range gs.Seats {
		startHand[i] = len(gs.Seats[i].Hand)
	}
	recipients := ResolveTemptingOffer(gs, 0, drawOneEffect(), nil)
	if len(recipients) != 1 || recipients[0] != 0 {
		t.Fatalf("nil callback should yield only-controller; got %v", recipients)
	}
	for i, seat := range gs.Seats {
		want := startHand[i]
		if i == 0 {
			want++
		}
		if got := len(seat.Hand); got != want {
			t.Fatalf("seat %d hand size = %d, want %d", i, got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// (b) All accept in a 4-seat game → 4 copies.
// ---------------------------------------------------------------------------

func TestTemptingOffer_AllAccept_FourCopies(t *testing.T) {
	gs := newTemptingOfferGame(t, 4)
	startHand := make([]int, len(gs.Seats))
	for i := range gs.Seats {
		startHand[i] = len(gs.Seats[i].Hand)
	}

	accept := func(seat int) bool { return true }
	recipients := ResolveTemptingOffer(gs, 1, drawOneEffect(), accept)

	if len(recipients) != 4 {
		t.Fatalf("recipients count = %d, want 4 (controller + 3 opponents)", len(recipients))
	}
	// Every seat drew exactly one card.
	for i, seat := range gs.Seats {
		if got := len(seat.Hand); got != startHand[i]+1 {
			t.Fatalf("seat %d hand size = %d, want %d", i, got, startHand[i]+1)
		}
	}
}

// ---------------------------------------------------------------------------
// (c) Accept order matches turn order from controller's left
// ---------------------------------------------------------------------------

func TestTemptingOffer_AcceptOrderMatchesTurnOrderFromLeft(t *testing.T) {
	gs := newTemptingOfferGame(t, 4)
	// All opponents accept; recipients should be controller first,
	// then opponents in APNAP order from controller's left.
	accept := func(seat int) bool { return true }

	for controllerSeat := 0; controllerSeat < 4; controllerSeat++ {
		recipients := ResolveTemptingOffer(gs, controllerSeat, drawOneEffect(), accept)

		if recipients[0] != controllerSeat {
			t.Fatalf("controller=%d: recipients[0] = %d, want %d (controller first)",
				controllerSeat, recipients[0], controllerSeat)
		}
		// LivingOpponents returns opponents in APNAP order — that's
		// the canonical "turn order from <seat>'s left."
		wantOpps := gs.LivingOpponents(controllerSeat)
		if len(recipients)-1 != len(wantOpps) {
			t.Fatalf("controller=%d: %d accepters, want %d",
				controllerSeat, len(recipients)-1, len(wantOpps))
		}
		for i, wantSeat := range wantOpps {
			if recipients[i+1] != wantSeat {
				t.Fatalf("controller=%d: recipients[%d] = %d, want %d (APNAP from controller's left)",
					controllerSeat, i+1, recipients[i+1], wantSeat)
			}
		}
	}
}

func TestTemptingOffer_MixedAccept_PreservesAPNAPSubset(t *testing.T) {
	gs := newTemptingOfferGame(t, 4)
	// Controller is seat 0; LivingOpponents(0) = [1, 2, 3].
	// Only seats 1 and 3 accept; seat 2 declines.
	accept := func(seat int) bool { return seat == 1 || seat == 3 }
	recipients := ResolveTemptingOffer(gs, 0, drawOneEffect(), accept)

	want := []int{0, 1, 3}
	if len(recipients) != len(want) {
		t.Fatalf("recipients = %v, want %v", recipients, want)
	}
	for i, w := range want {
		if recipients[i] != w {
			t.Fatalf("recipients[%d] = %d, want %d (controller, then APNAP-ordered accepters)",
				i, recipients[i], w)
		}
	}
}

// ---------------------------------------------------------------------------
// (d) Opponent copy targets THEIR resources (their library / life total).
// ---------------------------------------------------------------------------

func TestTemptingOffer_CopiesTargetEachRecipientsResources_Draw(t *testing.T) {
	gs := newTemptingOfferGame(t, 4)
	startHand := make([]int, len(gs.Seats))
	startLib := make([]int, len(gs.Seats))
	for i := range gs.Seats {
		startHand[i] = len(gs.Seats[i].Hand)
		startLib[i] = len(gs.Seats[i].Library)
	}

	// Controller = seat 2; opponents 3 and 0 accept; seat 1 declines.
	accept := func(seat int) bool { return seat == 3 || seat == 0 }
	ResolveTemptingOffer(gs, 2, drawOneEffect(), accept)

	// Recipients 2, 3, 0 each draw one — verify per-seat that the draw
	// came from THEIR own library, not the controller's.
	for _, seat := range []int{2, 3, 0} {
		if got := len(gs.Seats[seat].Hand); got != startHand[seat]+1 {
			t.Fatalf("seat %d hand = %d, want %d", seat, got, startHand[seat]+1)
		}
		if got := len(gs.Seats[seat].Library); got != startLib[seat]-1 {
			t.Fatalf("seat %d library = %d, want %d (drew from own library)",
				seat, got, startLib[seat]-1)
		}
	}
	// Seat 1 declined → no change to their hand or library.
	if got := len(gs.Seats[1].Hand); got != startHand[1] {
		t.Fatalf("seat 1 hand = %d, want %d (declined; no draw)", got, startHand[1])
	}
	if got := len(gs.Seats[1].Library); got != startLib[1] {
		t.Fatalf("seat 1 library = %d, want %d (declined; no draw)", got, startLib[1])
	}
}

func TestTemptingOffer_CopiesTargetEachRecipientsResources_GainLife(t *testing.T) {
	gs := newTemptingOfferGame(t, 4)
	startLife := make([]int, len(gs.Seats))
	for i, s := range gs.Seats {
		startLife[i] = s.Life
	}

	// Controller = seat 0; all opponents accept.
	accept := func(seat int) bool { return true }
	ResolveTemptingOffer(gs, 0, gainLifeOneEffect(), accept)

	// Every seat gained exactly 1 life — the gain landed on the
	// recipient's seat, not the controller's, because the synthetic
	// source's Controller field steers the default target.
	for i, s := range gs.Seats {
		if s.Life != startLife[i]+1 {
			t.Fatalf("seat %d life = %d, want %d (gain_life copy must apply to recipient)",
				i, s.Life, startLife[i]+1)
		}
	}
}

// ---------------------------------------------------------------------------
// Misc safety
// ---------------------------------------------------------------------------

func TestTemptingOffer_NilEffect(t *testing.T) {
	gs := newTemptingOfferGame(t, 2)
	if got := ResolveTemptingOffer(gs, 0, nil, nil); got != nil {
		t.Fatalf("nil base effect should return nil recipients, got %v", got)
	}
}

func TestTemptingOffer_InvalidController(t *testing.T) {
	gs := newTemptingOfferGame(t, 2)
	if got := ResolveTemptingOffer(gs, 99, drawOneEffect(), nil); got != nil {
		t.Fatalf("invalid controller should return nil recipients, got %v", got)
	}
}

func TestTemptingOffer_LostController(t *testing.T) {
	gs := newTemptingOfferGame(t, 2)
	gs.Seats[0].Lost = true
	if got := ResolveTemptingOffer(gs, 0, drawOneEffect(), nil); got != nil {
		t.Fatalf("Lost controller should return nil recipients, got %v", got)
	}
}

func TestTemptingOffer_SkipsLostOpponents(t *testing.T) {
	gs := newTemptingOfferGame(t, 4)
	gs.Seats[2].Lost = true // would-be opponent 2 is dead
	startHand := make([]int, len(gs.Seats))
	for i := range gs.Seats {
		startHand[i] = len(gs.Seats[i].Hand)
	}

	accept := func(seat int) bool { return true }
	recipients := ResolveTemptingOffer(gs, 0, drawOneEffect(), accept)

	// Controller (0) + living opponents (1, 3) — seat 2 skipped.
	want := []int{0, 1, 3}
	if len(recipients) != len(want) {
		t.Fatalf("recipients = %v, want %v (Lost seat must be skipped)", recipients, want)
	}
	for i, w := range want {
		if recipients[i] != w {
			t.Fatalf("recipients[%d] = %d, want %d", i, recipients[i], w)
		}
	}
	// Lost seat 2 did not draw.
	if got := len(gs.Seats[2].Hand); got != startHand[2] {
		t.Fatalf("Lost seat 2 hand = %d, want %d (no draw)", got, startHand[2])
	}
}
