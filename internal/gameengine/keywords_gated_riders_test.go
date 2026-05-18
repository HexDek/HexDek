package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// Round-31 integration tests for resolveGatedRider. Exercises the
// resolveSequence hook end-to-end: a Sequence-typed spell resolves, the
// outer-call gate fires resolveGatedRider, and each of Threshold /
// Metalcraft / Hellbent fires (or stays silent) depending on game
// state. Each case asserts on the rider's event-log signature so any
// future regression in the dispatch order or active-check is loud.

// ---------------------------------------------------------------------------
// Test fixtures
// ---------------------------------------------------------------------------

func gr_makeGame(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(2, rand.New(rand.NewSource(31)), nil)
}

// gr_makeRiderCard builds a Card whose AST carries a single Static
// ability with the given raw text. OracleTextLower reconstructs from
// that for the rider detectors, so the printed rider line lives in
// `oracleText`.
func gr_makeRiderCard(name, oracleText string) *Card {
	return &Card{
		Name:          name,
		Types:         []string{"sorcery"},
		BasePower:     0,
		BaseToughness: 0,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: oracleText},
			},
		},
	}
}

func gr_makePerm(seat int, card *Card) *Permanent {
	card.Owner = seat
	return &Permanent{
		Card:       card,
		Controller: seat,
		Owner:      seat,
		Flags:      map[string]int{},
	}
}

// gr_filler builds a generic card we can shovel into a graveyard or
// hand to drive *Active predicates above/below threshold.
func gr_filler(name string, ctype string) *Card {
	return &Card{Name: name, Types: []string{ctype}}
}

// gr_artifactPerm builds an artifact permanent for Metalcraft counts.
func gr_artifactPerm(seat int, name string) *Permanent {
	c := &Card{Name: name, Owner: seat, Types: []string{"artifact"}}
	return &Permanent{
		Card:       c,
		Controller: seat,
		Owner:      seat,
		Flags:      map[string]int{},
	}
}

// gr_countEvents returns the number of event-log entries with the
// given Kind.
func gr_countEvents(gs *GameState, kind string) int {
	n := 0
	for _, ev := range gs.EventLog {
		if ev.Kind == kind {
			n++
		}
	}
	return n
}

// gr_runSpell wraps resolveSequence on an empty outer Sequence, which
// is enough to trigger the rider hooks (the hooks fire AFTER the body,
// so an empty body still drives them on the outer unwind).
func gr_runSpell(gs *GameState, src *Permanent) {
	resolveSequence(gs, src, &gameast.Sequence{Items: nil})
}

// ---------------------------------------------------------------------------
// (a) Threshold rider: active → fires; inactive → silent.
// ---------------------------------------------------------------------------

func TestGatedRider_ThresholdActiveFires(t *testing.T) {
	gs := gr_makeGame(t)
	// 7+ cards in graveyard so ThresholdActive is true.
	for i := 0; i < 7; i++ {
		gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, gr_filler("g", "creature"))
	}
	src := gr_makePerm(0, gr_makeRiderCard("Hidden Stockpile",
		"Threshold — Draw a card."))
	gr_runSpell(gs, src)

	if gr_countEvents(gs, "threshold_rider") != 1 {
		t.Fatalf("expected 1 threshold_rider event when active; got %d",
			gr_countEvents(gs, "threshold_rider"))
	}
}

func TestGatedRider_ThresholdInactiveSilent(t *testing.T) {
	gs := gr_makeGame(t)
	// Only 3 cards in graveyard — below 7-card threshold.
	for i := 0; i < 3; i++ {
		gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, gr_filler("g", "creature"))
	}
	src := gr_makePerm(0, gr_makeRiderCard("Hidden Stockpile",
		"Threshold — Draw a card."))
	gr_runSpell(gs, src)

	if gr_countEvents(gs, "threshold_rider") != 0 {
		t.Fatalf("expected 0 threshold_rider events when inactive; got %d",
			gr_countEvents(gs, "threshold_rider"))
	}
}

// ---------------------------------------------------------------------------
// (b) Metalcraft rider: active → fires; inactive → silent.
// ---------------------------------------------------------------------------

func TestGatedRider_MetalcraftActiveFires(t *testing.T) {
	gs := gr_makeGame(t)
	// 3 artifacts on battlefield → MetalcraftActive true.
	for i := 0; i < 3; i++ {
		gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, gr_artifactPerm(0, "Mox"))
	}
	src := gr_makePerm(0, gr_makeRiderCard("Etched Champion",
		"Metalcraft — Etched Champion has protection from all colors."))
	gr_runSpell(gs, src)

	if gr_countEvents(gs, "metalcraft_rider") != 1 {
		t.Fatalf("expected 1 metalcraft_rider event when active; got %d",
			gr_countEvents(gs, "metalcraft_rider"))
	}
}

func TestGatedRider_MetalcraftInactiveSilent(t *testing.T) {
	gs := gr_makeGame(t)
	// Only 2 artifacts — below threshold.
	for i := 0; i < 2; i++ {
		gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, gr_artifactPerm(0, "Mox"))
	}
	src := gr_makePerm(0, gr_makeRiderCard("Etched Champion",
		"Metalcraft — Etched Champion has protection from all colors."))
	gr_runSpell(gs, src)

	if gr_countEvents(gs, "metalcraft_rider") != 0 {
		t.Fatalf("expected 0 metalcraft_rider events when inactive; got %d",
			gr_countEvents(gs, "metalcraft_rider"))
	}
}

// ---------------------------------------------------------------------------
// (c) Hellbent rider: active → fires; inactive → silent.
// ---------------------------------------------------------------------------

func TestGatedRider_HellbentActiveFires(t *testing.T) {
	gs := gr_makeGame(t)
	// Hand is empty → HellbentActive true.
	gs.Seats[0].Hand = nil
	src := gr_makePerm(0, gr_makeRiderCard("Demonfire",
		"Hellbent — Demonfire can't be countered."))
	gr_runSpell(gs, src)

	if gr_countEvents(gs, "hellbent_rider") != 1 {
		t.Fatalf("expected 1 hellbent_rider event when active; got %d",
			gr_countEvents(gs, "hellbent_rider"))
	}
}

func TestGatedRider_HellbentInactiveSilent(t *testing.T) {
	gs := gr_makeGame(t)
	// Hand has cards → HellbentActive false.
	gs.Seats[0].Hand = []*Card{gr_filler("Hand 1", "instant"), gr_filler("Hand 2", "instant")}
	src := gr_makePerm(0, gr_makeRiderCard("Demonfire",
		"Hellbent — Demonfire can't be countered."))
	gr_runSpell(gs, src)

	if gr_countEvents(gs, "hellbent_rider") != 0 {
		t.Fatalf("expected 0 hellbent_rider events when inactive; got %d",
			gr_countEvents(gs, "hellbent_rider"))
	}
}

// ---------------------------------------------------------------------------
// (d) All three gates independent — a Threshold-only spell shouldn't
// fire Metalcraft / Hellbent events even when those states are active.
// ---------------------------------------------------------------------------

func TestGatedRider_IndependentGates_ThresholdOnly(t *testing.T) {
	gs := gr_makeGame(t)
	// Make ALL three gating conditions true at once.
	for i := 0; i < 8; i++ {
		gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, gr_filler("g", "creature"))
	}
	for i := 0; i < 3; i++ {
		gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, gr_artifactPerm(0, "Mox"))
	}
	gs.Seats[0].Hand = nil // hellbent active

	// Card prints ONLY a threshold rider — no metalcraft, no hellbent
	// keywords in its oracle text.
	src := gr_makePerm(0, gr_makeRiderCard("Plain Threshold Spell",
		"Threshold — Draw two cards instead."))
	gr_runSpell(gs, src)

	if gr_countEvents(gs, "threshold_rider") != 1 {
		t.Fatalf("threshold rider should fire; got %d events",
			gr_countEvents(gs, "threshold_rider"))
	}
	if gr_countEvents(gs, "metalcraft_rider") != 0 {
		t.Fatalf("metalcraft rider must NOT fire on a threshold-only card; got %d",
			gr_countEvents(gs, "metalcraft_rider"))
	}
	if gr_countEvents(gs, "hellbent_rider") != 0 {
		t.Fatalf("hellbent rider must NOT fire on a threshold-only card; got %d",
			gr_countEvents(gs, "hellbent_rider"))
	}
}

func TestGatedRider_IndependentGates_MetalcraftOnly(t *testing.T) {
	gs := gr_makeGame(t)
	// All three active again.
	for i := 0; i < 8; i++ {
		gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, gr_filler("g", "creature"))
	}
	for i := 0; i < 3; i++ {
		gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, gr_artifactPerm(0, "Mox"))
	}
	gs.Seats[0].Hand = nil

	src := gr_makePerm(0, gr_makeRiderCard("Plain Metalcraft Spell",
		"Metalcraft — Gain 5 life."))
	gr_runSpell(gs, src)

	if gr_countEvents(gs, "metalcraft_rider") != 1 {
		t.Fatalf("metalcraft rider should fire; got %d",
			gr_countEvents(gs, "metalcraft_rider"))
	}
	if gr_countEvents(gs, "threshold_rider") != 0 {
		t.Fatalf("threshold rider must NOT fire on metalcraft-only card; got %d",
			gr_countEvents(gs, "threshold_rider"))
	}
	if gr_countEvents(gs, "hellbent_rider") != 0 {
		t.Fatalf("hellbent rider must NOT fire on metalcraft-only card; got %d",
			gr_countEvents(gs, "hellbent_rider"))
	}
}

func TestGatedRider_IndependentGates_HellbentOnly(t *testing.T) {
	gs := gr_makeGame(t)
	for i := 0; i < 8; i++ {
		gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, gr_filler("g", "creature"))
	}
	for i := 0; i < 3; i++ {
		gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, gr_artifactPerm(0, "Mox"))
	}
	gs.Seats[0].Hand = nil

	src := gr_makePerm(0, gr_makeRiderCard("Plain Hellbent Spell",
		"Hellbent — Discard target creature card from opponent's hand."))
	gr_runSpell(gs, src)

	if gr_countEvents(gs, "hellbent_rider") != 1 {
		t.Fatalf("hellbent rider should fire; got %d",
			gr_countEvents(gs, "hellbent_rider"))
	}
	if gr_countEvents(gs, "threshold_rider") != 0 {
		t.Fatalf("threshold rider must NOT fire on hellbent-only card; got %d",
			gr_countEvents(gs, "threshold_rider"))
	}
	if gr_countEvents(gs, "metalcraft_rider") != 0 {
		t.Fatalf("metalcraft rider must NOT fire on hellbent-only card; got %d",
			gr_countEvents(gs, "metalcraft_rider"))
	}
}

// ---------------------------------------------------------------------------
// Bonus: a card carrying ALL three riders fires all three when their
// gates are active simultaneously. Demonstrates that resolveGatedRider
// doesn't short-circuit between dispatches.
// ---------------------------------------------------------------------------

func TestGatedRider_AllThreeOnOneCard(t *testing.T) {
	gs := gr_makeGame(t)
	for i := 0; i < 8; i++ {
		gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, gr_filler("g", "creature"))
	}
	for i := 0; i < 3; i++ {
		gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, gr_artifactPerm(0, "Mox"))
	}
	gs.Seats[0].Hand = nil

	src := gr_makePerm(0, gr_makeRiderCard("Triple Rider Spell",
		"Threshold — Draw a card. Metalcraft — Gain 3 life. Hellbent — Deal 2 damage."))
	gr_runSpell(gs, src)

	for _, kind := range []string{"threshold_rider", "metalcraft_rider", "hellbent_rider"} {
		if gr_countEvents(gs, kind) != 1 {
			t.Fatalf("%s should fire exactly once on a triple-rider card; got %d",
				kind, gr_countEvents(gs, kind))
		}
	}
}

// Bonus: nested Sequence doesn't double-fire any rider.
func TestGatedRider_NestedSequenceDoesNotDoubleFire(t *testing.T) {
	gs := gr_makeGame(t)
	for i := 0; i < 8; i++ {
		gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, gr_filler("g", "creature"))
	}
	src := gr_makePerm(0, gr_makeRiderCard("Nested Threshold",
		"Threshold — Draw a card."))

	inner := &gameast.Sequence{Items: []gameast.Effect{}}
	outer := &gameast.Sequence{Items: []gameast.Effect{inner}}
	resolveSequence(gs, src, outer)

	if got := gr_countEvents(gs, "threshold_rider"); got != 1 {
		t.Fatalf("nested sequence should fire threshold rider exactly once; got %d", got)
	}
}
