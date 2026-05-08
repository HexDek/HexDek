package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// TestCreatePermanent_DedupOnDoubleCall reproduces the Feynman #3 hand-bloat
// bug: MoveCard("graveyard","battlefield") places the card, then
// enterBattlefieldWithETB calls createPermanent again for the same card.
// Before the dedup guard, this produced two Permanent wrappers pointing to
// the same *Card, leading to phantom zone drift when they left the battlefield
// via different paths.
func TestCreatePermanent_DedupOnDoubleCall(t *testing.T) {
	gs := newGame(t, 2)
	card := addCard(gs, 0, "Sheoldred", "creature")
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, card)

	// Simulate the double-call pattern from reanimation handlers:
	// Step 1: MoveCard places the card on the battlefield.
	result := gameengine.MoveCard(gs, card, 0, "graveyard", "battlefield", "reanimate")
	if result.FinalZone != "battlefield" {
		t.Fatalf("MoveCard returned %q, want battlefield", result.FinalZone)
	}
	if len(gs.Seats[0].Battlefield) != 1 {
		t.Fatalf("after MoveCard: battlefield=%d, want 1", len(gs.Seats[0].Battlefield))
	}

	// Step 2: enterBattlefieldWithETB (calls createPermanent internally).
	// Before the fix, this would create a SECOND permanent.
	perm := enterBattlefieldWithETB(gs, 0, card, false)
	if perm == nil {
		t.Fatal("enterBattlefieldWithETB returned nil")
	}

	// The dedup guard should return the existing permanent, not create a new one.
	if len(gs.Seats[0].Battlefield) != 1 {
		t.Errorf("dedup failed: battlefield has %d permanents, want 1", len(gs.Seats[0].Battlefield))
	}
	if gs.Seats[0].Battlefield[0].Card != card {
		t.Error("permanent wraps wrong card")
	}
}

// TestCreatePermanent_NoDedupForDifferentCards confirms that the dedup guard
// does NOT block distinct cards from entering the battlefield.
func TestCreatePermanent_NoDedupForDifferentCards(t *testing.T) {
	gs := newGame(t, 2)
	card1 := addCard(gs, 0, "Sheoldred", "creature")
	card2 := addCard(gs, 0, "Grave Titan", "creature")

	perm1 := enterBattlefieldWithETB(gs, 0, card1, false)
	perm2 := enterBattlefieldWithETB(gs, 0, card2, false)

	if perm1 == nil || perm2 == nil {
		t.Fatal("expected both permanents to be created")
	}
	if len(gs.Seats[0].Battlefield) != 2 {
		t.Errorf("two distinct cards should produce 2 permanents, got %d", len(gs.Seats[0].Battlefield))
	}
}

// TestCreatePermanent_CrossSeatAllowed verifies that the dedup guard only
// checks the TARGET seat. A card on seat 1's battlefield should NOT block
// createPermanent on seat 0 — this is the control-change pattern (Etali,
// Bribery, etc.) where MoveCard places on owner's seat then ETB fires
// under the controller.
func TestCreatePermanent_CrossSeatAllowed(t *testing.T) {
	gs := newGame(t, 2)
	card := addCard(gs, 0, "Stolen Creature", "creature")

	// Place on seat 1's battlefield directly (simulating MoveCard to owner).
	gs.Seats[1].Battlefield = append(gs.Seats[1].Battlefield, &gameengine.Permanent{
		Card:       card,
		Controller: 1,
		Owner:      0,
		Counters:   map[string]int{},
		Flags:      map[string]int{},
	})

	// Seat 0 tries createPermanent — should create a new perm (not dedup)
	// because the card is on a DIFFERENT seat.
	perm := createPermanent(gs, 0, card, false)
	if perm == nil {
		t.Fatal("cross-seat createPermanent should not be blocked by dedup")
	}
	if perm.Controller != 0 {
		t.Errorf("new perm should have controller=0, got %d", perm.Controller)
	}
	if len(gs.Seats[0].Battlefield) != 1 {
		t.Errorf("seat 0 should have 1 perm, got %d", len(gs.Seats[0].Battlefield))
	}
}
