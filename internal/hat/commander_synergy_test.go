package hat

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// synergySetup builds a 2-seat game with a named commander on the
// hat seat (1). The commander is placed either on the battlefield or
// in the command zone based on placement: "battlefield" or "command".
// Returns the configured hat (with strategy already wired) and the gs.
func synergySetup(t *testing.T, placement, cmdrName string, cmdrTypes []string, sp *StrategyProfile) (*YggdrasilHat, *gameengine.GameState) {
	t.Helper()
	gs := newTestGame(t, 2)
	gs.CommanderFormat = true

	cmdrCard := newTestCardMinimal(cmdrName, cmdrTypes, 4, nil)
	cmdrCard.BasePower, cmdrCard.BaseToughness = 3, 3
	gs.Seats[1].CommanderNames = []string{cmdrName}
	switch placement {
	case "battlefield":
		_ = newTestPermanent(gs.Seats[1], cmdrCard, 3, 3)
	case "command":
		gs.Seats[1].CommandZone = append(gs.Seats[1].CommandZone, cmdrCard)
	}

	h := NewYggdrasilHatWithNoise(sp, 0, 0)
	return h, gs
}

// TestCommanderSynergy_ValueEngineCardBoostedWhenCommanderOnField
// asserts a card listed as a ValueEngineKey scores higher when the
// commander is on the battlefield than when it's still in the
// command zone. We score the same card under both placements with
// identical strategy and compare.
func TestCommanderSynergy_ValueEngineCardBoostedWhenCommanderOnField(t *testing.T) {
	sp := &StrategyProfile{
		Archetype:       ArchetypeMidrange,
		ValueEngineKeys: []string{"Smothering Tithe"},
	}
	tithe := newTestCardMinimal("Smothering Tithe", []string{"enchantment"}, 4, nil)

	hOn, gsOn := synergySetup(t, "battlefield", "TestCmdr", []string{"creature", "legendary", "human", "wizard"}, sp)
	scoreOn := hOn.cardHeuristic(gsOn, 1, tithe)

	hOff, gsOff := synergySetup(t, "command", "TestCmdr", []string{"creature", "legendary", "human", "wizard"}, sp)
	scoreOff := hOff.cardHeuristic(gsOff, 1, tithe)

	if scoreOn-scoreOff < 0.20 {
		t.Errorf("expected ≥0.20 boost on ValueEngineKey when commander on field; on=%.3f off=%.3f delta=%.3f",
			scoreOn, scoreOff, scoreOn-scoreOff)
	}
}

// TestCommanderSynergy_NoBonusWhenCommanderInCommandZone — when the
// commander is sitting in the command zone, a generic non-ramp
// non-protection card should NOT be boosted (and gets the small
// "deploy commander first" penalty).
func TestCommanderSynergy_NoBonusWhenCommanderInCommandZone(t *testing.T) {
	sp := &StrategyProfile{
		Archetype:       ArchetypeMidrange,
		ValueEngineKeys: []string{}, // empty: scoring shouldn't get the +0.25 boost
	}
	// A vanilla 4-mana 4/4 creature — not ramp, not protection, not
	// in any synergy set.
	vanilla := newTestCardMinimal("Vanilla Beast", []string{"creature"}, 4, nil)
	vanilla.BasePower, vanilla.BaseToughness = 4, 4

	hOff, gsOff := synergySetup(t, "command", "TestCmdr", []string{"creature", "legendary", "human", "wizard"}, sp)
	scoreOff := hOff.cardHeuristic(gsOff, 1, vanilla)

	hAbsent, gsAbsent := synergySetup(t, "none", "TestCmdr", []string{"creature", "legendary", "human", "wizard"}, sp)
	scoreAbsent := hAbsent.cardHeuristic(gsAbsent, 1, vanilla)

	// With commander in command zone, score should be ≤ baseline (no
	// commander metadata participating). Strict less-than would over-
	// constrain since both setups still hit the Lovelace-themes path
	// the same way; ≤ is the right invariant.
	if scoreOff > scoreAbsent {
		t.Errorf("vanilla card should NOT be boosted when commander only in command zone; cmdZone=%.3f baseline=%.3f",
			scoreOff, scoreAbsent)
	}
}

// TestCommanderSynergy_ProtectionCardsBoostedWhenCommanderAbsent —
// when the commander is in the command zone, a counterspell (the
// canonical protection) scores higher than the same setup without
// the protection-bonus path firing.
func TestCommanderSynergy_ProtectionCardsBoostedWhenCommanderAbsent(t *testing.T) {
	sp := &StrategyProfile{
		Archetype:       ArchetypeMidrange,
		ValueEngineKeys: []string{},
	}
	// Counterspell is filtered OUT of cardHeuristic in some paths but
	// we score it directly here. Use a static-text counterspell so
	// CardHasCounterSpell returns true.
	counter := cardWithStaticText("Counterspell", []string{"instant"}, 2, "counter target spell")

	hOff, gsOff := synergySetup(t, "command", "TestCmdr", []string{"creature", "legendary", "human", "wizard"}, sp)
	scoreOff := hOff.cardHeuristic(gsOff, 1, counter)

	hOn, gsOn := synergySetup(t, "battlefield", "TestCmdr", []string{"creature", "legendary", "human", "wizard"}, sp)
	scoreOn := hOn.cardHeuristic(gsOn, 1, counter)

	// Protection-bonus only fires when the commander is in the
	// command zone (uncast). The counterspell should score strictly
	// higher in that state than when the commander is already on
	// the field (where the +0.15 protection bonus does NOT apply).
	if scoreOff <= scoreOn {
		t.Errorf("counterspell should be boosted as protection while commander is in command zone; cmdZone=%.3f onField=%.3f",
			scoreOff, scoreOn)
	}
}

// TestCommanderSynergy_TribalCreatureBonus — a creature sharing a
// subtype with the commander (both Wizards, etc.) gets the +0.10
// tribal nudge when the commander is on the battlefield.
func TestCommanderSynergy_TribalCreatureBonus(t *testing.T) {
	sp := &StrategyProfile{Archetype: ArchetypeTribal}
	wizard := newTestCardMinimal("Apprentice", []string{"creature", "wizard"}, 2, nil)
	wizard.BasePower, wizard.BaseToughness = 1, 2
	nonWizard := newTestCardMinimal("Bear", []string{"creature", "bear"}, 2, nil)
	nonWizard.BasePower, nonWizard.BaseToughness = 2, 2

	h, gs := synergySetup(t, "battlefield", "Big Wizard", []string{"creature", "legendary", "human", "wizard"}, sp)

	wizScore := h.cardHeuristic(gs, 1, wizard)
	bearScore := h.cardHeuristic(gs, 1, nonWizard)

	if wizScore-bearScore < 0.05 {
		t.Errorf("wizard sharing commander subtype should beat unrelated creature by ≥0.05; wizard=%.3f bear=%.3f delta=%.3f",
			wizScore, bearScore, wizScore-bearScore)
	}
}

// TestShouldCastCommander_StrategyEnablerLowersBuffer — a strategy
// profile with ≥3 ValueEngineKeys flags the commander as a
// strategy enabler: manaBuffer drops by 1 and maxTax rises by 2 so
// the hat is willing to recast through more tax. We check this
// behaviorally by finding a tax/mana point that flips between
// "won't cast" (without the enabler boost) and "will cast" (with).
func TestShouldCastCommander_StrategyEnablerLowersBuffer(t *testing.T) {
	enablerSP := &StrategyProfile{
		Archetype:       ArchetypeMidrange,
		ValueEngineKeys: []string{"A", "B", "C", "D"},
	}
	thinSP := &StrategyProfile{
		Archetype:       ArchetypeMidrange,
		ValueEngineKeys: []string{"A"}, // <3, no enabler bump
	}

	mkGame := func() *gameengine.GameState {
		gs := newTestGame(t, 2)
		gs.CommanderFormat = true
		gs.Turn = 5 // not late-game (>15 short-circuits to true)
		// Stuff seat 1's mana pool with enough lands to afford the
		// commander+tax budget under test.
		for i := 0; i < 8; i++ {
			land := newTestCardMinimal("Plains", []string{"land"}, 0, nil)
			lp := newTestPermanent(gs.Seats[1], land, 0, 0)
			_ = lp
		}
		gs.Seats[1].CommanderNames = []string{"Cmdr"}
		gs.Seats[1].CommandZone = append(gs.Seats[1].CommandZone,
			newTestCardMinimal("Cmdr", []string{"creature", "legendary"}, 4, nil))
		return gs
	}

	// At maxTax baseline of 6 for Midrange, picking tax=7 puts a thin
	// strategy at "won't cast unless avail ≥ 2*7+1 = 15", but the
	// enabler bump raises maxTax to 8 so tax=7 is just-affordable.
	tax := 7

	gsThin := mkGame()
	hThin := NewYggdrasilHatWithNoise(thinSP, 0, 0)
	gotThin := hThin.ShouldCastCommander(gsThin, 1, "Cmdr", tax)

	gsEnabler := mkGame()
	hEnabler := NewYggdrasilHatWithNoise(enablerSP, 0, 0)
	gotEnabler := hEnabler.ShouldCastCommander(gsEnabler, 1, "Cmdr", tax)

	if gotThin {
		t.Errorf("thin strategy (1 VE key) should refuse to cast at tax=%d under maxTax=6; got true", tax)
	}
	if !gotEnabler {
		t.Errorf("strategy-enabler (4 VE keys) should accept tax=%d under bumped maxTax=8; got false", tax)
	}
}
