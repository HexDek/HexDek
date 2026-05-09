package hat

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
	"github.com/hexdek/hexdek/internal/gameengine"
)

// Tests for the board-aware mode-effect scorer + ShouldTriggerOptional.

// ---------------------------------------------------------------------
// destroy / exile — best-target scoring
// ---------------------------------------------------------------------

func TestScoreModeEffect_DestroyScoresHigherWithComboPiece(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)
	// Inject a combo-piece signal — comboPieceSet is normally
	// populated from Strategy.ComboPieces but the bare hat starts
	// empty, so we set it directly for the test.
	h.comboPieceSet = map[string]bool{"Thassa's Oracle": true}

	// Opponent only has a vanilla creature.
	bear := newTestCardMinimal("Grizzly Bears", []string{"creature"}, 2, nil)
	newTestPermanent(gs.Seats[1], bear, 2, 2)
	vanillaScore := h.scoreModeEffect(gs, 0, &gameast.Destroy{}, 0)

	// New game with the combo piece.
	gs2 := newTestGame(t, 2)
	thoracle := newTestCardMinimal("Thassa's Oracle", []string{"creature"}, 2, nil)
	newTestPermanent(gs2.Seats[1], thoracle, 1, 3)
	comboScore := h.scoreModeEffect(gs2, 0, &gameast.Destroy{}, 0)

	if comboScore <= vanillaScore {
		t.Fatalf("destroy should score higher when a combo piece is on the board; vanilla=%.2f combo=%.2f",
			vanillaScore, comboScore)
	}
	if comboScore < 0.85 {
		t.Fatalf("destroy on a combo piece should be ≥ 0.85; got %.2f", comboScore)
	}
}

func TestScoreModeEffect_DestroyScoresLowWhenNoTargets(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)
	// No opponent permanents.
	score := h.scoreModeEffect(gs, 0, &gameast.Destroy{}, 0)
	if score > 0.20 {
		t.Fatalf("destroy with no legal targets should score near-zero; got %.2f", score)
	}
}

// ---------------------------------------------------------------------
// draw — hand-size scaling
// ---------------------------------------------------------------------

func TestScoreModeEffect_DrawScalesWithHandSize(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)

	// Empty hand.
	gs.Seats[0].Hand = nil
	emptyScore := h.scoreModeEffect(gs, 0, &gameast.Draw{}, 0)

	// Full hand (7 cards).
	gs.Seats[0].Hand = make([]*gameengine.Card, 7)
	for i := range gs.Seats[0].Hand {
		gs.Seats[0].Hand[i] = newTestCardMinimal("X", []string{"sorcery"}, 1, nil)
	}
	fullScore := h.scoreModeEffect(gs, 0, &gameast.Draw{}, 0)

	if emptyScore <= fullScore {
		t.Fatalf("empty-hand draw should score higher than full-hand; empty=%.2f full=%.2f",
			emptyScore, fullScore)
	}
	if emptyScore < 0.80 {
		t.Fatalf("empty-hand draw should be ≥ 0.80; got %.2f", emptyScore)
	}
	if fullScore > 0.40 {
		t.Fatalf("full-hand draw should be ≤ 0.40; got %.2f", fullScore)
	}
}

// ---------------------------------------------------------------------
// gain_life — life-total scaling
// ---------------------------------------------------------------------

func TestScoreModeEffect_GainLifeScalesWithCurrentLife(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)

	gs.Seats[0].Life = 5
	lowScore := h.scoreModeEffect(gs, 0, &gameast.GainLife{}, 0)

	gs.Seats[0].Life = 40
	highScore := h.scoreModeEffect(gs, 0, &gameast.GainLife{}, 0)

	if lowScore <= highScore {
		t.Fatalf("low-life gain_life should score higher than full-life; low=%.2f high=%.2f",
			lowScore, highScore)
	}
	if lowScore < 0.80 {
		t.Fatalf("at 5 life, gain_life should be ≥ 0.80; got %.2f", lowScore)
	}
}

// ---------------------------------------------------------------------
// damage / lose_life — lethal awareness
// ---------------------------------------------------------------------

func TestScoreModeEffect_DamageScoresHigherWhenLethal(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)

	// Opp at full life (40) — 3 damage barely matters.
	gs.Seats[1].Life = 40
	lowScore := h.scoreModeEffect(gs, 0, &gameast.Damage{Amount: gameast.NumberOrRef{IsInt: true, Int: 3}}, 0)

	// Opp at 3 life — 3 damage is lethal.
	gs.Seats[1].Life = 3
	lethalScore := h.scoreModeEffect(gs, 0, &gameast.Damage{Amount: gameast.NumberOrRef{IsInt: true, Int: 3}}, 0)

	if lethalScore <= lowScore {
		t.Fatalf("damage should score higher when lethal; full-life=%.2f lethal=%.2f",
			lowScore, lethalScore)
	}
}

// ---------------------------------------------------------------------
// create_token — go-wider awareness
// ---------------------------------------------------------------------

func TestScoreModeEffect_CreateTokenScalesWithCreatureCount(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)

	// Empty board.
	emptyScore := h.scoreModeEffect(gs, 0, &gameast.CreateToken{}, 0)

	// Wide board.
	gs2 := newTestGame(t, 2)
	for i := 0; i < 6; i++ {
		c := newTestCardMinimal("Saproling", []string{"creature"}, 1, nil)
		newTestPermanent(gs2.Seats[0], c, 1, 1)
	}
	wideScore := h.scoreModeEffect(gs2, 0, &gameast.CreateToken{}, 0)

	if wideScore <= emptyScore {
		t.Fatalf("wide-board token-creation should score higher; empty=%.2f wide=%.2f",
			emptyScore, wideScore)
	}
}

// ---------------------------------------------------------------------
// counter_mod — payoff awareness
// ---------------------------------------------------------------------

func TestScoreModeEffect_CounterModScoresLowWithoutCreatures(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)
	// Empty board → no recipient.
	score := h.scoreModeEffect(gs, 0, &gameast.CounterMod{}, 0)
	if score > 0.20 {
		t.Fatalf("counter_mod with no creatures should score near-zero; got %.2f", score)
	}
}

func TestScoreModeEffect_CounterModScoresWithCreature(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)
	bear := newTestCardMinimal("Grizzly Bears", []string{"creature"}, 2, nil)
	newTestPermanent(gs.Seats[0], bear, 2, 2)
	score := h.scoreModeEffect(gs, 0, &gameast.CounterMod{}, 0)
	if score < 0.40 {
		t.Fatalf("counter_mod with a creature recipient should be ≥ 0.40; got %.2f", score)
	}
}

// ---------------------------------------------------------------------
// sacrifice — aristocrats payoff awareness
// ---------------------------------------------------------------------

func TestScoreModeEffect_SacrificeScoresLowWithoutPayoff(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)
	score := h.scoreModeEffect(gs, 0, &gameast.Sacrifice{}, 0)
	if score > 0.20 {
		t.Fatalf("sacrifice with no aristocrats payoff should be ≤ 0.20; got %.2f", score)
	}
}

func TestScoreModeEffect_SacrificeScoresWithPayoff(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)

	// Build a Blood Artist-like card with the oracle text embedded as
	// a Triggered ability — OracleTextLower walks AST.Abilities.
	bloodArtist := &gameengine.Card{
		Name:  "Blood Artist",
		Owner: 0,
		Types: []string{"creature"},
		AST: &gameast.CardAST{
			Name: "Blood Artist",
			Abilities: []gameast.Ability{
				&gameast.Triggered{
					Raw: "Whenever Blood Artist or another creature dies, target player loses 1 life and you gain 1 life.",
				},
			},
		},
	}
	newTestPermanent(gs.Seats[0], bloodArtist, 0, 1)

	score := h.scoreModeEffect(gs, 0, &gameast.Sacrifice{}, 0)
	if score < 0.55 {
		t.Fatalf("sacrifice with Blood Artist payoff should be ≥ 0.55; got %.2f", score)
	}
}

// ---------------------------------------------------------------------
// mill — opponent library size awareness
// ---------------------------------------------------------------------

func TestScoreModeEffect_MillScoresHigherAgainstSmallLibrary(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)

	// Big library.
	gs.Seats[1].Library = make([]*gameengine.Card, 50)
	for i := range gs.Seats[1].Library {
		gs.Seats[1].Library[i] = newTestCardMinimal("X", []string{"creature"}, 1, nil)
	}
	bigScore := h.scoreModeEffect(gs, 0, &gameast.Mill{}, 0)

	// Tiny library.
	gs.Seats[1].Library = gs.Seats[1].Library[:3]
	tinyScore := h.scoreModeEffect(gs, 0, &gameast.Mill{}, 0)

	if tinyScore <= bigScore {
		t.Fatalf("mill should score higher when opponent's library is small; big=%.2f tiny=%.2f",
			bigScore, tinyScore)
	}
}

// ---------------------------------------------------------------------
// bounce — CMC awareness
// ---------------------------------------------------------------------

func TestScoreModeEffect_BounceScoresWithHighCMCPermanent(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)

	// 1-CMC perm only.
	c1 := newTestCardMinimal("Memnite", []string{"creature"}, 0, nil)
	c1.CMC = 1
	newTestPermanent(gs.Seats[1], c1, 1, 1)
	lowScore := h.scoreModeEffect(gs, 0, &gameast.Bounce{}, 0)

	// 7-CMC bomb.
	gs2 := newTestGame(t, 2)
	c2 := newTestCardMinimal("Eldrazi", []string{"creature"}, 7, nil)
	c2.CMC = 7
	newTestPermanent(gs2.Seats[1], c2, 7, 7)
	highScore := h.scoreModeEffect(gs2, 0, &gameast.Bounce{}, 0)

	if highScore <= lowScore {
		t.Fatalf("bounce on 7-CMC bomb should score higher than 1-CMC; low=%.2f high=%.2f",
			lowScore, highScore)
	}
}

// ---------------------------------------------------------------------
// CurseDNA axes — boost/penalty wiring
// ---------------------------------------------------------------------

func TestScoreModeEffect_DNADrainAffinityBoostsSacrificeAndDamage(t *testing.T) {
	gs := newTestGame(t, 2)

	// Baseline (neutral DNA) sacrifice score.
	hNeutral := NewYggdrasilHat(nil, 0)
	hNeutral.DNA = &CurseDNA{DrainAffinity: 0.5}
	bloodArtist := &gameengine.Card{
		Name: "Blood Artist", Owner: 0, Types: []string{"creature"},
		AST: &gameast.CardAST{
			Name: "Blood Artist",
			Abilities: []gameast.Ability{
				&gameast.Triggered{
					Raw: "Whenever a creature dies, target player loses 1 life.",
				},
			},
		},
	}
	newTestPermanent(gs.Seats[0], bloodArtist, 0, 1)
	neutralScore := hNeutral.scoreModeEffect(gs, 0, &gameast.Sacrifice{}, 0)

	// High-drain DNA — should boost sacrifice.
	hDrain := NewYggdrasilHat(nil, 0)
	hDrain.DNA = &CurseDNA{DrainAffinity: 1.0}
	drainScore := hDrain.scoreModeEffect(gs, 0, &gameast.Sacrifice{}, 0)

	if drainScore <= neutralScore {
		t.Fatalf("DrainAffinity=1.0 should boost sacrifice score; neutral=%.2f drain=%.2f",
			neutralScore, drainScore)
	}
}

func TestScoreModeEffect_DNATokenPressureBoostsCreateToken(t *testing.T) {
	gs := newTestGame(t, 2)

	hNeutral := NewYggdrasilHat(nil, 0)
	hNeutral.DNA = &CurseDNA{TokenPressure: 0.5}
	neutralScore := hNeutral.scoreModeEffect(gs, 0, &gameast.CreateToken{}, 0)

	hToken := NewYggdrasilHat(nil, 0)
	hToken.DNA = &CurseDNA{TokenPressure: 1.0}
	tokenScore := hToken.scoreModeEffect(gs, 0, &gameast.CreateToken{}, 0)

	if tokenScore <= neutralScore {
		t.Fatalf("TokenPressure=1.0 should boost create_token score; neutral=%.2f token=%.2f",
			neutralScore, tokenScore)
	}
}

func TestScoreModeEffect_DNAResourceGreedBoostsDraw(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.Seats[0].Hand = make([]*gameengine.Card, 4) // mid-hand baseline

	hNeutral := NewYggdrasilHat(nil, 0)
	hNeutral.DNA = &CurseDNA{ResourceGreed: 0.5}
	neutralScore := hNeutral.scoreModeEffect(gs, 0, &gameast.Draw{}, 0)

	hGreedy := NewYggdrasilHat(nil, 0)
	hGreedy.DNA = &CurseDNA{ResourceGreed: 1.0}
	greedyScore := hGreedy.scoreModeEffect(gs, 0, &gameast.Draw{}, 0)

	if greedyScore <= neutralScore {
		t.Fatalf("ResourceGreed=1.0 should boost draw score; neutral=%.2f greedy=%.2f",
			neutralScore, greedyScore)
	}
}

func TestScoreModeEffect_DNAGraveyardExploitationBoostsReanimate(t *testing.T) {
	gs := newTestGame(t, 2)
	// Mid-size graveyard.
	gs.Seats[0].Graveyard = make([]*gameengine.Card, 4)
	for i := range gs.Seats[0].Graveyard {
		gs.Seats[0].Graveyard[i] = newTestCardMinimal("X", []string{"creature"}, 2, nil)
	}

	hNeutral := NewYggdrasilHat(nil, 0)
	hNeutral.DNA = &CurseDNA{GraveyardExploitation: 0.5}
	neutralScore := hNeutral.scoreModeEffect(gs, 0, &gameast.Reanimate{}, 0)

	hGY := NewYggdrasilHat(nil, 0)
	hGY.DNA = &CurseDNA{GraveyardExploitation: 1.0}
	gyScore := hGY.scoreModeEffect(gs, 0, &gameast.Reanimate{}, 0)

	if gyScore <= neutralScore {
		t.Fatalf("GraveyardExploitation=1.0 should boost reanimate score; neutral=%.2f gy=%.2f",
			neutralScore, gyScore)
	}
}

// ---------------------------------------------------------------------
// ChooseMode — board-aware mode selection
// ---------------------------------------------------------------------

func TestChooseMode_PicksDestroyOverDrawWhenComboPieceOnOpponent(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)
	h.comboPieceSet = map[string]bool{"Thassa's Oracle": true}
	// Crank the confidence threshold up so the top-N stochastic pick
	// sticks with the highest score instead of sampling within the
	// margin band.
	h.confidenceThreshold = 0.95
	gs.Seats[0].Hand = make([]*gameengine.Card, 7) // full hand → draw ~0.30
	for i := range gs.Seats[0].Hand {
		gs.Seats[0].Hand[i] = newTestCardMinimal("X", []string{"sorcery"}, 1, nil)
	}

	// Opponent has Thassa's Oracle (combo piece) → destroy ≈ 0.95.
	thoracle := newTestCardMinimal("Thassa's Oracle", []string{"creature"}, 2, nil)
	newTestPermanent(gs.Seats[1], thoracle, 1, 3)

	modes := []gameast.Effect{
		&gameast.Draw{},
		&gameast.Destroy{},
	}
	pick := h.ChooseMode(gs, 0, modes)
	if pick != 1 {
		drawScore := h.scoreModeEffect(gs, 0, modes[0], 0)
		destroyScore := h.scoreModeEffect(gs, 0, modes[1], 0)
		t.Fatalf("ChooseMode should pick destroy(idx=1) when opp has a combo piece; got %d (draw=%.2f destroy=%.2f)",
			pick, drawScore, destroyScore)
	}
}

// ---------------------------------------------------------------------
// ShouldTriggerOptional — net-positive gate
// ---------------------------------------------------------------------

func TestShouldTriggerOptional_AcceptsHighValueDraw(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)
	gs.Seats[0].Hand = nil // empty hand → high draw score

	if !h.ShouldTriggerOptional(gs, 0, &gameast.Draw{}) {
		t.Fatalf("optional draw with empty hand should be accepted")
	}
}

func TestShouldTriggerOptional_RejectsZeroValueSacrifice(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)
	// No aristocrats payoffs → sacrifice scores ~0.10.
	if h.ShouldTriggerOptional(gs, 0, &gameast.Sacrifice{}) {
		t.Fatalf("optional sacrifice with no payoff should be rejected")
	}
}

func TestShouldTriggerOptional_NilEffectFalse(t *testing.T) {
	gs := newTestGame(t, 2)
	h := NewYggdrasilHat(nil, 0)
	if h.ShouldTriggerOptional(gs, 0, nil) {
		t.Fatalf("nil effect should not trigger")
	}
}
