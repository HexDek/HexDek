package hat

import (
	"testing"
)

// dev/phase-awareness — verifies the GamePhase detector + the three
// hooks wired into cardHeuristic, ChooseAttackers, and
// ShouldCastCommander.

// ---------------------------------------------------------------------------
// detectPhase
// ---------------------------------------------------------------------------

func TestDetectPhase_TurnSpine(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	// Give the seat a few cards in hand so the hand-exhaustion override
	// doesn't pull mid/late turns to Execute.
	for i := 0; i < 5; i++ {
		gs.Seats[0].Hand = append(gs.Seats[0].Hand, newTestCardMinimal("F", []string{"creature"}, 2, nil))
	}
	cases := []struct {
		turn int
		want GamePhase
	}{
		{1, PhaseDeploy},
		{4, PhaseDeploy},
		{5, PhaseDevelop},
		{9, PhaseDevelop},
		{10, PhaseExecute},
		{20, PhaseExecute},
	}
	for _, tc := range cases {
		gs.Turn = tc.turn
		got := h.detectPhase(gs, 0)
		if got != tc.want {
			t.Errorf("turn %d: want %v, got %v", tc.turn, tc.want, got)
		}
	}
}

func TestDetectPhase_RampedDeckBumpsOutOfDeploy(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	gs.Turn = 3
	gs.Seats[0].ManaPool = 9 // ramped to 9 by turn 3 → past Deploy

	if got := h.detectPhase(gs, 0); got != PhaseDevelop {
		t.Fatalf("expected ramped turn-3 deck to read as Develop; got %v", got)
	}
}

func TestDetectPhase_CommanderOutBumps(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	gs.CommanderFormat = true
	gs.Turn = 2
	gs.Seats[0].CommanderNames = []string{"Atraxa, Praetors' Voice"}
	atraxaCard := newTestCardMinimal("Atraxa, Praetors' Voice", []string{"creature", "legendary"}, 4, nil)
	_ = newTestPermanent(gs.Seats[0], atraxaCard, 4, 4)

	if got := h.detectPhase(gs, 0); got != PhaseDevelop {
		t.Fatalf("commander on battlefield should bump phase to at least Develop; got %v", got)
	}
}

func TestDetectPhase_HandExhaustionBumpsToExecute(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	gs.Turn = 7 // baseline Develop
	gs.Seats[0].Hand = nil

	if got := h.detectPhase(gs, 0); got != PhaseExecute {
		t.Fatalf("empty hand at turn 7 should bump to Execute; got %v", got)
	}
}

func TestDetectPhase_HighManaForcesExecute(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	gs.Turn = 6
	gs.Seats[0].ManaPool = 13

	if got := h.detectPhase(gs, 0); got != PhaseExecute {
		t.Fatalf("12+ mana should pull phase to Execute regardless of turn; got %v", got)
	}
}

func TestDetectPhase_BigBoardBumpsToExecute(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	gs.Turn = 7 // baseline Develop
	for i := 0; i < 9; i++ {
		c := newTestCardMinimal("Token", []string{"creature"}, 1, nil)
		newTestPermanent(gs.Seats[0], c, 1, 1)
	}
	if got := h.detectPhase(gs, 0); got != PhaseExecute {
		t.Fatalf("9+ creatures should bump phase to Execute; got %v", got)
	}
}

// ---------------------------------------------------------------------------
// cardHeuristic phase shaping
// ---------------------------------------------------------------------------

func TestCardHeuristic_DeployRewardsRamp(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	gs.Turn = 2
	gs.Seats[0].ManaPool = 2

	rampCard := newTestCardMinimal("Cultivate", []string{"sorcery"}, 3, nil)
	rampCard.TypeLine = "Sorcery"
	// Force category to Ramp via Strategy.CardRoles.
	h.Strategy = &StrategyProfile{CardRoles: map[string]string{"Cultivate": "Ramp"}}

	deployScore := h.cardHeuristic(gs, 0, rampCard)

	// Same card in Execute phase.
	gs.Turn = 12
	executeScore := h.cardHeuristic(gs, 0, rampCard)

	if deployScore <= executeScore {
		t.Fatalf("ramp should score higher in Deploy than Execute; deploy=%.3f execute=%.3f",
			deployScore, executeScore)
	}
}

func TestCardHeuristic_DeployPenalizesExpensiveThreat(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	gs.Turn = 2
	gs.Seats[0].ManaPool = 2

	bomb := newTestCardMinimal("Avenger of Zendikar", []string{"creature"}, 7, nil)
	bomb.TypeLine = "Creature — Plant Beast"
	h.Strategy = &StrategyProfile{CardRoles: map[string]string{"Avenger of Zendikar": "Threat"}}

	deployScore := h.cardHeuristic(gs, 0, bomb)
	gs.Turn = 12
	executeScore := h.cardHeuristic(gs, 0, bomb)

	if deployScore >= executeScore {
		t.Fatalf("expensive threat should score LOWER in Deploy than Execute; deploy=%.3f execute=%.3f",
			deployScore, executeScore)
	}
}

func TestCardHeuristic_ExecuteRewardsFinisher(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	gs.Turn = 12
	gs.Seats[0].ManaPool = 8

	finisher := newTestCardMinimal("Craterhoof Behemoth", []string{"creature"}, 8, nil)
	finisher.TypeLine = "Creature — Beast"
	h.Strategy = &StrategyProfile{
		CardRoles: map[string]string{"Craterhoof Behemoth": "Threat"},
	}
	// finisherSet is what isFinisher reads — set it directly to bypass
	// the Strategy ingestion path.
	h.finisherSet = map[string]bool{"Craterhoof Behemoth": true}

	executeScore := h.cardHeuristic(gs, 0, finisher)

	gs.Turn = 2
	gs.Seats[0].ManaPool = 1
	deployScore := h.cardHeuristic(gs, 0, finisher)

	if executeScore <= deployScore {
		t.Fatalf("finisher should score higher in Execute than Deploy; execute=%.3f deploy=%.3f",
			executeScore, deployScore)
	}
}

// ---------------------------------------------------------------------------
// ShouldCastCommander phase override
// ---------------------------------------------------------------------------

func TestShouldCastCommander_DeployAlwaysCasts(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	gs.CommanderFormat = true
	gs.Turn = 3
	gs.Seats[0].ManaPool = 6
	gs.Seats[0].CommanderNames = []string{"Krenko"}

	// High tax that would normally fail — Deploy should still return true.
	if !h.ShouldCastCommander(gs, 0, "Krenko", 4) {
		t.Fatalf("Deploy phase should always cast commander when affordable")
	}
}

func TestShouldCastCommander_ExecuteCastsAtHighTax(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	gs.CommanderFormat = true
	gs.Turn = 12 // baseline Execute
	gs.Seats[0].ManaPool = 10
	gs.Seats[0].CommanderNames = []string{"Krenko"}

	// Even at tax 10 (would normally exceed maxTax=8 default), Execute
	// returns true.
	if !h.ShouldCastCommander(gs, 0, "Krenko", 10) {
		t.Fatalf("Execute phase should cast commander at high tax when affordable")
	}
}

func TestShouldCastCommander_DevelopRespectsTaxCap(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	gs.CommanderFormat = true
	gs.Turn = 6 // baseline Develop
	gs.Seats[0].ManaPool = 8
	gs.Seats[0].CommanderNames = []string{"Krenko"}
	// Hand cards prevent the hand-exhaustion override from bumping
	// the phase to Execute (which would let any tax through).
	for i := 0; i < 4; i++ {
		gs.Seats[0].Hand = append(gs.Seats[0].Hand, newTestCardMinimal("F", []string{"creature"}, 2, nil))
	}

	// Default maxTax for Midrange archetype = 6; tax 12 should fail
	// because mana isn't enough for the doubled budget either.
	if h.ShouldCastCommander(gs, 0, "Krenko", 12) {
		t.Fatalf("Develop phase should NOT cast at tax 12 when mana doesn't cover")
	}
}
