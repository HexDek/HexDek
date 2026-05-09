package hat

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// dev/order-replacements — verifies the One Ring guard in
// YggdrasilHat.activationHeuristic. The activated ability adds a burden
// counter and draws N cards, then each upkeep drains N life. At low
// life with high burden counters, repeated activation is suicide.
//
// Spec: skip activation when life <= burdens*2.

func TestYggdrasil_OneRingHeuristic_HighLifeAllows(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	gs.Seats[0].Life = 30

	ringCard := newTestCardMinimal("The One Ring", []string{"artifact", "legendary"}, 4, nil)
	ring := newTestPermanent(gs.Seats[0], ringCard, 0, 0)
	ring.Counters = map[string]int{"burden": 2} // life threshold = 4 → we're way above

	opt := &gameengine.Activation{Permanent: ring, Ability: 0}
	score := h.activationHeuristic(gs, 0, opt)
	if score < 0 {
		t.Fatalf("at 30 life with 2 burdens (life > burdens*2) the heuristic must allow activation; got %f", score)
	}
}

func TestYggdrasil_OneRingHeuristic_LowLifeRefuses(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	gs.Seats[0].Life = 6 // burdens=3 → threshold=6, life <= 6 → refuse

	ringCard := newTestCardMinimal("The One Ring", []string{"artifact", "legendary"}, 4, nil)
	ring := newTestPermanent(gs.Seats[0], ringCard, 0, 0)
	ring.Counters = map[string]int{"burden": 3}

	opt := &gameengine.Activation{Permanent: ring, Ability: 0}
	score := h.activationHeuristic(gs, 0, opt)
	if score >= 0 {
		t.Fatalf("at life 6 with 3 burdens (life <= burdens*2) heuristic must refuse; got %f", score)
	}
}

func TestYggdrasil_OneRingHeuristic_FirstActivationAlwaysOK(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	gs.Seats[0].Life = 10

	ringCard := newTestCardMinimal("The One Ring", []string{"artifact", "legendary"}, 4, nil)
	ring := newTestPermanent(gs.Seats[0], ringCard, 0, 0)
	// Counters map left nil — first activation, 0 burdens.

	opt := &gameengine.Activation{Permanent: ring, Ability: 0}
	score := h.activationHeuristic(gs, 0, opt)
	if score < 0 {
		t.Fatalf("first activation at 10 life with 0 burdens must be allowed; got %f", score)
	}
}
