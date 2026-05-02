package hat

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------
// NewComboSequencer construction
// ---------------------------------------------------------------------

func TestComboSequencer_NilProfile(t *testing.T) {
	cs := NewComboSequencer(nil)
	if cs != nil {
		t.Fatal("nil profile should return nil sequencer")
	}
}

func TestComboSequencer_EmptyPieces(t *testing.T) {
	sp := &StrategyProfile{ComboPieces: []ComboPlan{}}
	cs := NewComboSequencer(sp)
	if cs != nil {
		t.Fatal("empty ComboPieces should return nil sequencer")
	}
}

func TestComboSequencer_BuildsLines(t *testing.T) {
	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{Pieces: []string{"Thassa's Oracle", "Demonic Consultation"}, Type: "infinite", CastOrder: []string{"Demonic Consultation", "Thassa's Oracle"}},
			{Pieces: []string{"Sanguine Bond", "Exquisite Blood"}, Type: "infinite"},
		},
	}
	cs := NewComboSequencer(sp)
	if cs == nil {
		t.Fatal("expected non-nil sequencer")
	}
	if len(cs.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(cs.Lines))
	}
	// First line should use explicit CastOrder.
	if cs.Lines[0].SequenceOrder[0] != "Demonic Consultation" {
		t.Fatalf("expected CastOrder override; got %s", cs.Lines[0].SequenceOrder[0])
	}
	// Second line should fall back to Pieces order.
	if cs.Lines[1].SequenceOrder[0] != "Sanguine Bond" {
		t.Fatalf("expected Pieces fallback; got %s", cs.Lines[1].SequenceOrder[0])
	}
	// Infinite combos should need protection.
	if !cs.Lines[0].NeedsProtection {
		t.Fatal("infinite combo should need protection")
	}
}

// ---------------------------------------------------------------------
// Evaluate: executable combo
// ---------------------------------------------------------------------

func TestComboSequencer_Executable_BothInHand(t *testing.T) {
	gs := newTestGame(t, 2)
	seat := gs.Seats[0]

	// Put both combo pieces in hand.
	oracle := newTestCardMinimal("Thassa's Oracle", []string{"creature"}, 2, nil)
	consult := newTestCardMinimal("Demonic Consultation", []string{"instant"}, 1, nil)
	seat.Hand = append(seat.Hand, oracle, consult)

	// Give enough mana: 3 untapped lands.
	for i := 0; i < 3; i++ {
		land := newTestCardMinimal("Island", []string{"land"}, 0, nil)
		newTestPermanent(seat, land, 0, 0)
	}

	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{
				Pieces:    []string{"Thassa's Oracle", "Demonic Consultation"},
				Type:      "infinite",
				CastOrder: []string{"Demonic Consultation", "Thassa's Oracle"},
			},
		},
	}
	cs := NewComboSequencer(sp)
	result := cs.Evaluate(gs, 0)

	if !result.Executable {
		t.Fatal("combo should be executable with both pieces in hand and enough mana")
	}
	if result.NextAction != "Demonic Consultation" {
		t.Fatalf("next action should be Demonic Consultation (cast first); got %q", result.NextAction)
	}
	if result.PiecesFound != 2 {
		t.Fatalf("expected 2 pieces found; got %d", result.PiecesFound)
	}
	if result.PiecesTotal != 2 {
		t.Fatalf("expected 2 pieces total; got %d", result.PiecesTotal)
	}
}

func TestComboSequencer_Executable_OneOnBattlefield(t *testing.T) {
	gs := newTestGame(t, 2)
	seat := gs.Seats[0]

	// Bond already on battlefield.
	bond := newTestCardMinimal("Sanguine Bond", []string{"enchantment"}, 5, nil)
	newTestPermanent(seat, bond, 0, 0)

	// Blood in hand.
	blood := newTestCardMinimal("Exquisite Blood", []string{"enchantment"}, 5, nil)
	seat.Hand = append(seat.Hand, blood)

	// 5 untapped lands for the remaining piece.
	for i := 0; i < 5; i++ {
		land := newTestCardMinimal("Swamp", []string{"land"}, 0, nil)
		newTestPermanent(seat, land, 0, 0)
	}

	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{Pieces: []string{"Sanguine Bond", "Exquisite Blood"}, Type: "infinite"},
		},
	}
	cs := NewComboSequencer(sp)
	result := cs.Evaluate(gs, 0)

	if !result.Executable {
		t.Fatal("combo should be executable with one piece on battlefield and one in hand")
	}
	if result.NextAction != "Exquisite Blood" {
		t.Fatalf("next action should be Exquisite Blood; got %q", result.NextAction)
	}
}

// ---------------------------------------------------------------------
// Evaluate: not enough mana
// ---------------------------------------------------------------------

func TestComboSequencer_NotExecutable_InsufficientMana(t *testing.T) {
	gs := newTestGame(t, 2)
	seat := gs.Seats[0]

	oracle := newTestCardMinimal("Thassa's Oracle", []string{"creature"}, 2, nil)
	consult := newTestCardMinimal("Demonic Consultation", []string{"instant"}, 1, nil)
	seat.Hand = append(seat.Hand, oracle, consult)

	// Only 1 land — not enough for CMC 2 + CMC 1 = 3.
	land := newTestCardMinimal("Island", []string{"land"}, 0, nil)
	newTestPermanent(seat, land, 0, 0)

	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{Pieces: []string{"Thassa's Oracle", "Demonic Consultation"}, Type: "infinite"},
		},
	}
	cs := NewComboSequencer(sp)
	result := cs.Evaluate(gs, 0)

	if result.Executable {
		t.Fatal("combo should NOT be executable with insufficient mana")
	}
	if result.PiecesFound != 2 {
		t.Fatalf("pieces should still be found even with insufficient mana; got %d", result.PiecesFound)
	}
}

// ---------------------------------------------------------------------
// Evaluate: assembling (missing 1 piece + tutor in hand)
// ---------------------------------------------------------------------

func TestComboSequencer_Assembling_OnePieceMissingWithTutor(t *testing.T) {
	gs := newTestGame(t, 2)
	seat := gs.Seats[0]

	// Only one combo piece in hand.
	oracle := newTestCardMinimal("Thassa's Oracle", []string{"creature"}, 2, nil)
	seat.Hand = append(seat.Hand, oracle)

	// A tutor in hand (has "search your library" in oracle text via AST).
	tutorAST := &gameast.CardAST{
		Name: "Demonic Tutor",
		Abilities: []gameast.Ability{
			&gameast.Activated{
				Raw:    "Search your library for a card, put it into your hand, then shuffle.",
				Effect: &gameast.Tutor{Destination: "hand", Query: gameast.Filter{Base: "card"}},
			},
		},
	}
	tutor := newTestCardMinimal("Demonic Tutor", []string{"sorcery"}, 2, tutorAST)
	seat.Hand = append(seat.Hand, tutor)

	// Some lands.
	for i := 0; i < 4; i++ {
		land := newTestCardMinimal("Swamp", []string{"land"}, 0, nil)
		newTestPermanent(seat, land, 0, 0)
	}

	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{Pieces: []string{"Thassa's Oracle", "Demonic Consultation"}, Type: "infinite"},
		},
	}
	cs := NewComboSequencer(sp)
	result := cs.Evaluate(gs, 0)

	if result.Executable {
		t.Fatal("combo should NOT be executable (missing Demonic Consultation)")
	}
	if !result.Assembling {
		t.Fatal("should be assembling: 1 piece missing + tutor in hand")
	}
	if result.MissingPiece != "Demonic Consultation" {
		t.Fatalf("missing piece should be Demonic Consultation; got %q", result.MissingPiece)
	}
}

// ---------------------------------------------------------------------
// Evaluate: missing piece, no tutor = neither executable nor assembling
// ---------------------------------------------------------------------

func TestComboSequencer_NotAssembling_NoTutor(t *testing.T) {
	gs := newTestGame(t, 2)
	seat := gs.Seats[0]

	oracle := newTestCardMinimal("Thassa's Oracle", []string{"creature"}, 2, nil)
	seat.Hand = append(seat.Hand, oracle)

	// No tutor, just a random card.
	filler := newTestCardMinimal("Lightning Bolt", []string{"instant"}, 1, nil)
	seat.Hand = append(seat.Hand, filler)

	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{Pieces: []string{"Thassa's Oracle", "Demonic Consultation"}, Type: "infinite"},
		},
	}
	cs := NewComboSequencer(sp)
	result := cs.Evaluate(gs, 0)

	if result.Executable {
		t.Fatal("should not be executable")
	}
	if result.Assembling {
		t.Fatal("should not be assembling without a tutor")
	}
	if result.PiecesFound != 1 {
		t.Fatalf("expected 1 piece found; got %d", result.PiecesFound)
	}
}

// ---------------------------------------------------------------------
// Evaluate: 3-card combo
// ---------------------------------------------------------------------

func TestComboSequencer_ThreeCardCombo(t *testing.T) {
	gs := newTestGame(t, 2)
	seat := gs.Seats[0]

	// Isochron Scepter + Dramatic Reversal + a mana rock on battlefield.
	scepter := newTestCardMinimal("Isochron Scepter", []string{"artifact"}, 2, nil)
	reversal := newTestCardMinimal("Dramatic Reversal", []string{"instant"}, 2, nil)
	rock := newTestCardMinimal("Sol Ring", []string{"artifact"}, 1, nil)

	seat.Hand = append(seat.Hand, scepter, reversal)
	newTestPermanent(seat, rock, 0, 0)

	// Enough lands.
	for i := 0; i < 4; i++ {
		land := newTestCardMinimal("Island", []string{"land"}, 0, nil)
		newTestPermanent(seat, land, 0, 0)
	}

	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{
				Pieces:    []string{"Isochron Scepter", "Dramatic Reversal", "Sol Ring"},
				Type:      "infinite",
				CastOrder: []string{"Isochron Scepter", "Dramatic Reversal"},
			},
		},
	}
	cs := NewComboSequencer(sp)
	result := cs.Evaluate(gs, 0)

	if !result.Executable {
		t.Fatal("3-card combo should be executable")
	}
	if result.PiecesFound != 3 {
		t.Fatalf("expected 3 pieces found; got %d", result.PiecesFound)
	}
	if result.NextAction != "Isochron Scepter" {
		t.Fatalf("next action should be Isochron Scepter; got %q", result.NextAction)
	}
}

// ---------------------------------------------------------------------
// Evaluate: best line selection (multiple combos)
// ---------------------------------------------------------------------

func TestComboSequencer_BestLineSelection(t *testing.T) {
	gs := newTestGame(t, 2)
	seat := gs.Seats[0]

	// Combo A: 2-card, both in hand (executable).
	bond := newTestCardMinimal("Sanguine Bond", []string{"enchantment"}, 5, nil)
	blood := newTestCardMinimal("Exquisite Blood", []string{"enchantment"}, 5, nil)
	seat.Hand = append(seat.Hand, bond, blood)

	// Combo B: 2-card, only 1 piece (not executable).
	oracle := newTestCardMinimal("Thassa's Oracle", []string{"creature"}, 2, nil)
	seat.Hand = append(seat.Hand, oracle)

	// Enough mana for combo A.
	for i := 0; i < 10; i++ {
		land := newTestCardMinimal("Swamp", []string{"land"}, 0, nil)
		newTestPermanent(seat, land, 0, 0)
	}

	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{Pieces: []string{"Thassa's Oracle", "Demonic Consultation"}, Type: "infinite"},
			{Pieces: []string{"Sanguine Bond", "Exquisite Blood"}, Type: "infinite"},
		},
	}
	cs := NewComboSequencer(sp)
	result := cs.Evaluate(gs, 0)

	if !result.Executable {
		t.Fatal("should be executable (combo B is complete)")
	}
	if result.BestLine == nil {
		t.Fatal("BestLine should not be nil")
	}
	// Best line should be the Sanguine Bond combo (executable beats partial).
	if result.BestLine.PiecesNeeded[0] != "Sanguine Bond" {
		t.Fatalf("best line should be Sanguine Bond combo; got %v", result.BestLine.PiecesNeeded)
	}
}

// ---------------------------------------------------------------------
// Evaluate: piece in graveyard counts
// ---------------------------------------------------------------------

func TestComboSequencer_PieceInGraveyard(t *testing.T) {
	gs := newTestGame(t, 2)
	seat := gs.Seats[0]

	oracle := newTestCardMinimal("Thassa's Oracle", []string{"creature"}, 2, nil)
	seat.Hand = append(seat.Hand, oracle)

	// Demonic Consultation in graveyard.
	consult := newTestCardMinimal("Demonic Consultation", []string{"instant"}, 1, nil)
	seat.Graveyard = append(seat.Graveyard, consult)

	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{Pieces: []string{"Thassa's Oracle", "Demonic Consultation"}, Type: "infinite"},
		},
	}
	cs := NewComboSequencer(sp)
	result := cs.Evaluate(gs, 0)

	// Both pieces found (hand + graveyard).
	if result.PiecesFound != 2 {
		t.Fatalf("expected 2 pieces found (hand + graveyard); got %d", result.PiecesFound)
	}
}

// ---------------------------------------------------------------------
// Evaluate: nil sequencer returns empty assessment
// ---------------------------------------------------------------------

func TestComboSequencer_NilEvaluate(t *testing.T) {
	gs := newTestGame(t, 2)
	var cs *ComboSequencer
	result := cs.Evaluate(gs, 0)
	if result.Executable || result.Assembling || result.BestLine != nil {
		t.Fatal("nil sequencer should return empty assessment")
	}
}

// ---------------------------------------------------------------------
// Evaluate: all pieces on battlefield (activated combo)
// ---------------------------------------------------------------------

func TestComboSequencer_AllOnBattlefield(t *testing.T) {
	gs := newTestGame(t, 2)
	seat := gs.Seats[0]

	// Both pieces already on battlefield.
	bond := newTestCardMinimal("Sanguine Bond", []string{"enchantment"}, 5, nil)
	blood := newTestCardMinimal("Exquisite Blood", []string{"enchantment"}, 5, nil)
	newTestPermanent(seat, bond, 0, 0)
	newTestPermanent(seat, blood, 0, 0)

	sp := &StrategyProfile{
		ComboPieces: []ComboPlan{
			{Pieces: []string{"Sanguine Bond", "Exquisite Blood"}, Type: "infinite"},
		},
	}
	cs := NewComboSequencer(sp)
	result := cs.Evaluate(gs, 0)

	if !result.Executable {
		t.Fatal("combo with all pieces on battlefield should be executable")
	}
	// NextAction should be first in sequence since nothing needs casting.
	if result.NextAction != "Sanguine Bond" {
		t.Fatalf("next action for all-on-battlefield should be first sequence piece; got %q", result.NextAction)
	}
}
