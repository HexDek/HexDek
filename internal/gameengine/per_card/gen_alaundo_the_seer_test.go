package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// TestAlaundo_Activate_DrawsAndSuspends verifies the activated ability
// draws one and exiles the highest-CMC permanent from hand.
func TestAlaundo_Activate_DrawsAndSuspends(t *testing.T) {
	gs := newGame(t, 2)

	// Library: 1 card to draw.
	gs.Seats[0].Library = append(gs.Seats[0].Library,
		&gameengine.Card{Name: "Drawn", Owner: 0, Types: []string{"creature", "cost:2"}},
	)
	// Hand: a low-CMC creature and a high-CMC permanent. Handler should
	// suspend the higher one.
	low := &gameengine.Card{Name: "LowCMC", Owner: 0, Types: []string{"creature", "cost:2"}}
	high := &gameengine.Card{Name: "HighCMC", Owner: 0, Types: []string{"creature", "cost:7"}}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, low, high)

	alaundo := addPerm(gs, 0, "Alaundo the Seer", "creature", "legendary")
	gameengine.InvokeActivatedHook(gs, alaundo, 0, nil)

	// Hand should still hold the low-CMC card; high-CMC moved to exile.
	if len(gs.Seats[0].Hand) != 2 {
		t.Errorf("expected 2 cards in hand (drawn + low), got %d", len(gs.Seats[0].Hand))
	}
	foundHighInExile := false
	for _, c := range gs.Seats[0].Exile {
		if c != nil && c.Name == "HighCMC" {
			foundHighInExile = true
		}
	}
	if !foundHighInExile {
		t.Errorf("expected HighCMC in exile, got exile size=%d", len(gs.Seats[0].Exile))
	}
}

// TestAlaundo_Activate_ReleasesAfterThreeTaps verifies the suspended
// release fires every alaundoReleaseInterval-th activation.
func TestAlaundo_Activate_ReleasesAfterThreeTaps(t *testing.T) {
	gs := newGame(t, 2)

	// Stack the library and hand with enough fodder for three
	// activations: 3 draws, 3 hand fatties.
	for i := 0; i < 3; i++ {
		gs.Seats[0].Library = append(gs.Seats[0].Library,
			&gameengine.Card{Name: "L", Owner: 0, Types: []string{"sorcery"}},
		)
		gs.Seats[0].Hand = append(gs.Seats[0].Hand,
			&gameengine.Card{Name: "Fattie", Owner: 0, Types: []string{"creature", "cost:6"}},
		)
	}

	alaundo := addPerm(gs, 0, "Alaundo the Seer", "creature", "legendary")
	bfBefore := len(gs.Seats[0].Battlefield)

	for i := 0; i < 3; i++ {
		gameengine.InvokeActivatedHook(gs, alaundo, 0, nil)
	}

	// After 3 activations one Fattie should have been released to bf.
	released := 0
	for _, p := range gs.Seats[0].Battlefield {
		if p == nil || p == alaundo {
			continue
		}
		if p.Card != nil && p.Card.Name == "Fattie" {
			released++
		}
	}
	if released < 1 {
		t.Errorf("expected at least 1 Fattie released to battlefield after 3 activations, bfBefore=%d bfAfter=%d", bfBefore, len(gs.Seats[0].Battlefield))
	}
}

