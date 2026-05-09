package hat

import (
	"strings"
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// TestCommanderUrgency_StrategyEnabler — a strategy with ≥3 value
// engine keys (and no higher-tier signal) should score 0.7.
func TestCommanderUrgency_StrategyEnabler(t *testing.T) {
	sp := &StrategyProfile{
		Archetype:       ArchetypeMidrange,
		ValueEngineKeys: []string{"Smothering Tithe", "Rhystic Study", "Sylvan Library"},
	}
	h := NewYggdrasilHatWithNoise(sp, 0, 0)
	got := h.commanderUrgency("Some Generic Commander")
	if got != 0.7 {
		t.Fatalf("expected urgency 0.7 for ≥3 value engine keys, got %.2f", got)
	}
}

// TestCommanderUrgency_ComboCommander — when the commander name appears
// in any ComboPlan piece list, urgency is 0.9.
func TestCommanderUrgency_ComboCommander(t *testing.T) {
	sp := &StrategyProfile{
		Archetype: ArchetypeCombo,
		ComboPieces: []ComboPlan{
			{
				Pieces: []string{"Kiki-Jiki, Mirror Breaker", "Felidar Guardian"},
				Type:   "infinite",
			},
		},
	}
	h := NewYggdrasilHatWithNoise(sp, 0, 0)
	got := h.commanderUrgency("Kiki-Jiki, Mirror Breaker")
	if got != 0.9 {
		t.Fatalf("expected urgency 0.9 for commander in ComboPlan, got %.2f", got)
	}
	// Sanity: a non-combo-piece commander on the same strategy stays
	// below combo tier.
	other := h.commanderUrgency("Some Other Commander")
	if other >= 0.9 {
		t.Fatalf("non-combo-piece commander should not score combo-tier urgency, got %.2f", other)
	}
}

// TestCommanderUrgency_CommanderCentric — IsCommanderCentric trumps
// every other signal at 0.95.
func TestCommanderUrgency_CommanderCentric(t *testing.T) {
	sp := &StrategyProfile{
		Archetype:          ArchetypeMidrange,
		IsCommanderCentric: true,
	}
	h := NewYggdrasilHatWithNoise(sp, 0, 0)
	got := h.commanderUrgency("Uril, the Miststalker")
	if got != 0.95 {
		t.Fatalf("expected urgency 0.95 for IsCommanderCentric, got %.2f", got)
	}
}

// TestCommanderUrgency_Tribal — tribal strategy archetype yields 0.8
// (lord/anthem commander assumption).
func TestCommanderUrgency_Tribal(t *testing.T) {
	sp := &StrategyProfile{Archetype: ArchetypeTribal}
	h := NewYggdrasilHatWithNoise(sp, 0, 0)
	got := h.commanderUrgency("Krenko, Mob Boss")
	if got != 0.8 {
		t.Fatalf("expected urgency 0.8 for tribal archetype, got %.2f", got)
	}
}

// TestCommanderUrgency_Default — nil strategy and "nothing matches"
// strategy both fall through to 0.4.
func TestCommanderUrgency_Default(t *testing.T) {
	// Nil strategy.
	h1 := NewYggdrasilHatWithNoise(nil, 0, 0)
	if got := h1.commanderUrgency("Anything"); got != 0.4 {
		t.Fatalf("expected urgency 0.4 for nil strategy, got %.2f", got)
	}
	// Strategy with neither IsCommanderCentric, ComboPieces with the
	// commander name, Tribal archetype, nor ≥3 value engine keys.
	sp := &StrategyProfile{
		Archetype:       ArchetypeMidrange,
		ValueEngineKeys: []string{"Sol Ring"}, // only 1, below threshold
	}
	h2 := NewYggdrasilHatWithNoise(sp, 0, 0)
	if got := h2.commanderUrgency("Generic Commander"); got != 0.4 {
		t.Fatalf("expected urgency 0.4 for low-signal strategy, got %.2f", got)
	}
}

// makeInteractionTable builds a 4-seat game where opponents present
// high inferred interaction risk: hand cards, open mana, and known
// blue color identity. Returns the seat we cast as.
func makeInteractionTable(t *testing.T, h *YggdrasilHat) (*gameengine.GameState, int) {
	t.Helper()
	gs := newTestGame(t, 4)
	mySeat := 0

	// Give us comfortable mana for a 3-mana commander + 2 tax = 7
	// total cost at the higher end: avail = 7 puts us right at the
	// "tax*2+2" threshold where the interaction-risk gate would block.
	gs.Seats[mySeat].ManaPool = 7

	// Make all opponents look threatening: open mana + hand + blue.
	for i := 1; i < len(gs.Seats); i++ {
		s := gs.Seats[i]
		s.ManaPool = 4
		// Give them a hand so opponentHasInteraction's len(s.Hand)==0
		// short-circuit doesn't trigger.
		for j := 0; j < 5; j++ {
			s.Hand = append(s.Hand, newTestCardMinimal("Filler", []string{"creature"}, 1, nil))
		}
	}

	// Initialize hat tracking arrays via an ObserveEvent on a no-op event.
	h.ObserveEvent(gs, mySeat, &gameengine.Event{Kind: "turn_start", Seat: mySeat})

	// Force opponent[1] to look blue — drives prob >> 0.5 in
	// opponentHasInteraction.
	if len(h.opponentColors) > 1 {
		h.opponentColors[1]["U"] = true
		h.opponentColors[1]["B"] = true
	}
	return gs, mySeat
}

// TestShouldCastCommander_HighUrgencyIgnoresInteractionRisk — at urgency
// ≥ 0.6, the interaction-risk gate must NOT block the cast even when
// tableInteractionRisk would otherwise. Pin gs.Turn into the Develop
// phase so detectPhase doesn't auto-return on us.
func TestShouldCastCommander_HighUrgencyIgnoresInteractionRisk(t *testing.T) {
	sp := &StrategyProfile{
		Archetype:          ArchetypeMidrange,
		IsCommanderCentric: true, // urgency = 0.95
	}
	h := NewYggdrasilHatWithNoise(sp, 0, 0)
	gs, seat := makeInteractionTable(t, h)
	gs.Turn = 5 // PhaseDevelop — past Deploy auto-cast, before Execute auto-cast

	// Sanity-check the risk really is above threshold.
	if got := h.tableInteractionRisk(gs, seat); got <= 0.5 {
		t.Fatalf("test setup error: expected tableInteractionRisk > 0.5, got %.2f", got)
	}

	// Tax = 2 puts us in the "if tax >= 2" arm.
	if !h.ShouldCastCommander(gs, seat, "Uril, the Miststalker", 2) {
		t.Fatalf("high-urgency commander should cast through interaction risk")
	}
}

// TestShouldCastCommander_LowUrgencyRespectsInteractionRisk — same
// table state, low urgency, should refuse to cast. Pin gs.Turn into
// the Develop phase so detectPhase doesn't auto-return.
func TestShouldCastCommander_LowUrgencyRespectsInteractionRisk(t *testing.T) {
	// Default profile: no IsCommanderCentric, no combo/tribal/value
	// engines → urgency = 0.4.
	sp := &StrategyProfile{Archetype: ArchetypeMidrange}
	h := NewYggdrasilHatWithNoise(sp, 0, 0)
	gs, seat := makeInteractionTable(t, h)

	// Make sure we're below the "tax*2+2" comfort threshold AND
	// below the avail≥9 PhaseDevelop bump.
	gs.Seats[seat].Mana = nil
	gs.Seats[seat].ManaPool = 5 // tax=2 needs 6 to satisfy tax*2+2
	gs.Turn = 5                  // PhaseDevelop

	// Give us a non-trivial hand so the "hand <= 1 && turn >= 6"
	// Execute bump can never engage even if Turn drifts up later.
	for j := 0; j < 5; j++ {
		gs.Seats[seat].Hand = append(gs.Seats[seat].Hand,
			newTestCardMinimal("Filler", []string{"creature"}, 1, nil))
	}

	// Re-prime risk-tracking arrays after mana reset.
	h.ObserveEvent(gs, seat, &gameengine.Event{Kind: "turn_start", Seat: seat})
	if len(h.opponentColors) > 1 {
		h.opponentColors[1]["U"] = true
		h.opponentColors[1]["B"] = true
	}

	if got := h.tableInteractionRisk(gs, seat); got <= 0.5 {
		t.Fatalf("test setup error: expected tableInteractionRisk > 0.5, got %.2f", got)
	}

	if h.ShouldCastCommander(gs, seat, "Generic Commander", 2) {
		t.Fatalf("low-urgency commander should respect interaction risk and skip cast")
	}
}

// TestShouldCastCommander_LogsDecision — the hat should emit a
// CMDR-CAST line per decision so spectator replays can explain why a
// commander was or wasn't cast.
func TestShouldCastCommander_LogsDecision(t *testing.T) {
	sp := &StrategyProfile{
		Archetype:          ArchetypeMidrange,
		IsCommanderCentric: true,
	}
	h := NewYggdrasilHatWithNoise(sp, 0, 0)
	log := []string{}
	h.DecisionLog = &log

	gs := newTestGame(t, 2)
	gs.Seats[0].ManaPool = 5

	_ = h.ShouldCastCommander(gs, 0, "Uril, the Miststalker", 0)

	found := false
	for _, line := range log {
		if strings.Contains(line, "CMDR-CAST") && strings.Contains(line, "Uril, the Miststalker") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected a CMDR-CAST log line for the decision, got: %v", log)
	}
}
