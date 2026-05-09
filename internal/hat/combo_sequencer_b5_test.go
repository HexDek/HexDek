package hat

// Regression tests for the B5 combo-execution ceiling. Three blind
// spots in the combo sequencer kept commander-engine and DFC decks
// stuck below 15% win rate even with all win-line pieces present:
//
//  1. CommandZone wasn't scanned, so a commander combo piece sitting
//     in the command zone was never recognized as "available" — the
//     sequencer reported "1/2 pieces" indefinitely for Kinnan +
//     Basalt Monolith and similar lines.
//  2. ZonesAccepted defaulted to {hand, battlefield, graveyard}, so
//     even if CommandZone WAS scanned, the line wouldn't accept a
//     command-zone source.
//  3. The zone index keyed strictly off c.Name. After an MDFC back-
//     face cast (Esika→Bridge) mutated Card.Name to the back-face
//     name, the Bridge permanent on the battlefield no longer matched
//     Freya's full "Front // Back" piece names.
//
// These tests exercise each fix end-to-end through ComboSequencer.Evaluate.

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// addCommandZone places a card in the seat's command zone. Mirrors
// what gameengine.SetupCommanderGame does for the commander.
func addCommandZone(seat *gameengine.Seat, c *gameengine.Card) {
	seat.CommandZone = append(seat.CommandZone, c)
}

// TestComboSequencer_RecognisesCommanderInCommandZone is the headline
// fix: with the commander as one of two combo pieces, the line must
// be reported as Executable when the other piece is in hand and there's
// enough mana to cast both.
func TestComboSequencer_RecognisesCommanderInCommandZone(t *testing.T) {
	gs := newTestGame(t, 2)
	seat := gs.Seats[0]

	// Kinnan (CMC 2) in command zone. Basalt Monolith (CMC 3) in hand.
	// 5 untapped lands cover the combined cost.
	kinnan := newTestCardMinimal("Kinnan, Bonder Prodigy", []string{"creature"}, 2, nil)
	addCommandZone(seat, kinnan)
	monolith := newTestCardMinimal("Basalt Monolith", []string{"artifact"}, 3, nil)
	seat.Hand = append(seat.Hand, monolith)
	for i := 0; i < 5; i++ {
		newTestPermanent(seat, newTestCardMinimal("Forest", []string{"land"}, 0, nil), 0, 0)
	}

	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{
				Pieces:    []string{"Basalt Monolith", "Kinnan, Bonder Prodigy"},
				Type:      "infinite",
				CastOrder: []string{"Basalt Monolith", "Kinnan, Bonder Prodigy"},
			},
		},
	}
	cs := NewComboSequencer(sp)
	result := cs.Evaluate(gs, 0)

	if !result.Executable {
		t.Fatalf("combo must be Executable with commander in command zone + other piece in hand; "+
			"PiecesFound=%d/%d", result.PiecesFound, result.PiecesTotal)
	}
	if result.PiecesFound != 2 {
		t.Errorf("expected 2 pieces found (commander + Monolith); got %d", result.PiecesFound)
	}
}

// TestComboSequencer_NextActionWalksCommandZonePiece — when the
// CastOrder puts the commander first and the other piece is already
// on the battlefield, NextAction must name the commander so
// ShouldCastCommander has the signal to fire.
func TestComboSequencer_NextActionWalksCommandZonePiece(t *testing.T) {
	gs := newTestGame(t, 2)
	seat := gs.Seats[0]

	kinnan := newTestCardMinimal("Kinnan, Bonder Prodigy", []string{"creature"}, 2, nil)
	addCommandZone(seat, kinnan)
	monolith := newTestCardMinimal("Basalt Monolith", []string{"artifact"}, 3, nil)
	newTestPermanent(seat, monolith, 0, 0) // already on battlefield
	for i := 0; i < 3; i++ {
		newTestPermanent(seat, newTestCardMinimal("Forest", []string{"land"}, 0, nil), 0, 0)
	}

	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{
				Pieces:    []string{"Basalt Monolith", "Kinnan, Bonder Prodigy"},
				Type:      "infinite",
				CastOrder: []string{"Basalt Monolith", "Kinnan, Bonder Prodigy"},
			},
		},
	}
	cs := NewComboSequencer(sp)
	result := cs.Evaluate(gs, 0)

	if !result.Executable {
		t.Fatalf("combo must be Executable; PiecesFound=%d", result.PiecesFound)
	}
	if result.NextAction != "Kinnan, Bonder Prodigy" {
		t.Errorf("NextAction should name the commander (since the other piece is on battlefield); got %q",
			result.NextAction)
	}
}

// TestComboSequencer_CommanderCostsIncludeTax — a piece in the command
// zone with non-zero tax must factor §903.8 (+2 per prior cast) into
// the executable check. Kinnan at tax 1 is 2+2 = 4 mana; with Monolith
// already on battlefield and only 3 mana floating, the combo should
// NOT be Executable yet.
func TestComboSequencer_CommanderCostsIncludeTax(t *testing.T) {
	gs := newTestGame(t, 2)
	seat := gs.Seats[0]

	kinnan := newTestCardMinimal("Kinnan, Bonder Prodigy", []string{"creature"}, 2, nil)
	addCommandZone(seat, kinnan)
	if seat.CommanderTax == nil {
		seat.CommanderTax = map[string]int{}
	}
	seat.CommanderTax["Kinnan, Bonder Prodigy"] = 1 // already cast once

	monolith := newTestCardMinimal("Basalt Monolith", []string{"artifact"}, 3, nil)
	newTestPermanent(seat, monolith, 0, 0)

	for i := 0; i < 3; i++ {
		newTestPermanent(seat, newTestCardMinimal("Forest", []string{"land"}, 0, nil), 0, 0)
	}

	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{Pieces: []string{"Basalt Monolith", "Kinnan, Bonder Prodigy"}, Type: "infinite"},
		},
	}
	cs := NewComboSequencer(sp)
	result := cs.Evaluate(gs, 0)

	if result.Executable {
		t.Errorf("combo must NOT be Executable with 3 mana but recast cost = 2 base + 2 tax = 4")
	}

	// Bump mana to 4: now executable.
	newTestPermanent(seat, newTestCardMinimal("Forest", []string{"land"}, 0, nil), 0, 0)
	result = cs.Evaluate(gs, 0)
	if !result.Executable {
		t.Errorf("combo must be Executable with 4 mana covering 2-base-CMC + 2-tax recast")
	}
}

// TestComboSequencer_RecognisesMDFCBackFaceOnBattlefield — Freya emits
// pieces with the canonical full DFC oracle name ("Esika, God of the
// Tree // The Prismatic Bridge"). After a back-face cast resolves,
// the on-battlefield permanent's Card.Name is mutated to just the
// back-face name. Without DFC alias indexing, the sequencer would
// fail to recognize the Bridge permanent as the combo piece.
func TestComboSequencer_RecognisesMDFCBackFaceOnBattlefield(t *testing.T) {
	gs := newTestGame(t, 2)
	seat := gs.Seats[0]

	// Bridge permanent — runtime Card.Name swapped to back face only,
	// BackFaceName preserved (the deckparser still sets it on the Card).
	bridge := &gameengine.Card{
		Name:         "The Prismatic Bridge",
		BackFaceName: "The Prismatic Bridge",
		BackFaceCMC:  6,
		Types:        []string{"legendary", "enchantment"},
	}
	newTestPermanent(seat, bridge, 0, 0)

	// Some support piece in hand.
	support := newTestCardMinimal("Tooth and Nail", []string{"sorcery"}, 9, nil)
	seat.Hand = append(seat.Hand, support)
	for i := 0; i < 9; i++ {
		newTestPermanent(seat, newTestCardMinimal("Forest", []string{"land"}, 0, nil), 0, 0)
	}

	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{
				// Freya names the piece with the canonical full DFC string.
				Pieces: []string{"Esika, God of the Tree // The Prismatic Bridge", "Tooth and Nail"},
				Type:   "determined",
			},
		},
	}
	cs := NewComboSequencer(sp)
	result := cs.Evaluate(gs, 0)

	if !result.Executable {
		t.Fatalf("Bridge on battlefield must match the full DFC piece name; PiecesFound=%d/%d",
			result.PiecesFound, result.PiecesTotal)
	}
}

// TestComboSequencer_RecognisesMDFCFullNameInHand — symmetric case:
// Freya pieces use the full "Front // Back" name, but the seat has the
// MDFC in hand pre-cast. The Card's runtime Name IS the full DFC
// string; this test guards the path where Freya emits a single-face
// piece name and we want to match the full-DFC-named card.
func TestComboSequencer_RecognisesMDFCFullNameInHand(t *testing.T) {
	gs := newTestGame(t, 2)
	seat := gs.Seats[0]

	// Card in hand with full DFC name.
	dfc := &gameengine.Card{
		Name:         "Esika, God of the Tree // The Prismatic Bridge",
		BackFaceName: "The Prismatic Bridge",
		BackFaceCMC:  6,
		Types:        []string{"legendary", "creature", "god", "cost:4"},
	}
	seat.Hand = append(seat.Hand, dfc)
	for i := 0; i < 4; i++ {
		newTestPermanent(seat, newTestCardMinimal("Forest", []string{"land"}, 0, nil), 0, 0)
	}

	// Freya keys the piece off the back face only — the alias index
	// must still match.
	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{Pieces: []string{"The Prismatic Bridge"}, Type: "finisher"},
		},
	}
	cs := NewComboSequencer(sp)
	result := cs.Evaluate(gs, 0)

	if result.PiecesFound != 1 {
		t.Errorf("MDFC in hand must match a single-face piece name via DFC alias; PiecesFound=%d",
			result.PiecesFound)
	}
}

// TestComboSequencer_BattlefieldOverridesHandForCostAccounting — when
// a piece appears in BOTH hand and battlefield (a copy was cast,
// another sits in hand), we must charge zero cast cost for that piece.
// The pre-fix loop walked zones in declaration order and would set
// needsCast=true for the hand copy even though battlefield had it
// resolved. Without this fix, decks running double-tap pieces (Spark
// Double, Sakashima copies) reported false negatives on Executable.
func TestComboSequencer_BattlefieldOverridesHandForCostAccounting(t *testing.T) {
	gs := newTestGame(t, 2)
	seat := gs.Seats[0]

	// Bond on battlefield AND another in hand.
	bond := newTestCardMinimal("Sanguine Bond", []string{"enchantment"}, 5, nil)
	newTestPermanent(seat, bond, 0, 0)
	bondCopy := newTestCardMinimal("Sanguine Bond", []string{"enchantment"}, 5, nil)
	seat.Hand = append(seat.Hand, bondCopy)

	// Blood in hand — needs 5 mana to cast.
	blood := newTestCardMinimal("Exquisite Blood", []string{"enchantment"}, 5, nil)
	seat.Hand = append(seat.Hand, blood)

	// 5 mana — only enough for Blood, not Blood + (incorrectly-charged) Bond.
	for i := 0; i < 5; i++ {
		newTestPermanent(seat, newTestCardMinimal("Swamp", []string{"land"}, 0, nil), 0, 0)
	}

	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{Pieces: []string{"Sanguine Bond", "Exquisite Blood"}, Type: "infinite"},
		},
	}
	cs := NewComboSequencer(sp)
	result := cs.Evaluate(gs, 0)

	if !result.Executable {
		t.Fatalf("combo must be Executable: Bond is on battlefield (no cast needed), "+
			"Blood (CMC 5) covered by 5 mana; got PiecesFound=%d", result.PiecesFound)
	}
}
