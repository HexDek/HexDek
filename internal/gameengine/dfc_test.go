package gameengine

// Tests for CR §712 DFC + §726 Day/Night + §702.144/145 Daybound/
// Nightbound. Mirrors scripts/test_day_night.py (Python-side parity).

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// syntheticDaybound builds a DFC Permanent: front face daybound 2/2,
// back face nightbound 3/3. Parser-free — we hand-forge the ASTs so
// the test doesn't depend on the oracle corpus.
func syntheticDaybound(gs *GameState, seatIdx int) *Permanent {
	front := &gameast.CardAST{
		Name: "Test Wolf Human",
		Abilities: []gameast.Ability{
			&gameast.Keyword{Name: "daybound", Raw: "Daybound"},
		},
		FullyParsed: true,
	}
	back := &gameast.CardAST{
		Name: "Test Wolf Beast",
		Abilities: []gameast.Ability{
			&gameast.Keyword{Name: "nightbound", Raw: "Nightbound"},
		},
		FullyParsed: true,
	}
	card := &Card{
		AST:           front,
		Name:          "Test Wolf Human",
		Owner:         seatIdx,
		BasePower:     2,
		BaseToughness: 2,
		Types:         []string{"creature"},
		CMC:           2,
		TypeLine:      "Creature — Human Werewolf",
	}
	perm := &Permanent{
		Card:          card,
		Controller:    seatIdx,
		Owner:         seatIdx,
		Tapped:        false,
		SummoningSick: false,
	}
	perm.Timestamp = gs.NextTimestamp()
	InitDFCFaces(perm, front, back, "Test Wolf Human", "Test Wolf Beast")
	gs.Seats[seatIdx].Battlefield = append(gs.Seats[seatIdx].Battlefield, perm)
	return perm
}

func TestDayNightInitialState(t *testing.T) {
	gs := NewGameState(2, rand.New(rand.NewSource(0)), nil)
	if gs.DayNight != DayNightNeither {
		t.Fatalf("§726.2 expected 'neither', got %q", gs.DayNight)
	}
}

func TestDayNight_becomes_day_on_daybound_etb(t *testing.T) {
	gs := NewGameState(2, rand.New(rand.NewSource(0)), nil)
	syntheticDaybound(gs, 0)
	// Manually trigger — in production the runtime's ETB path calls
	// MaybeBecomeDay.
	MaybeBecomeDay(gs, "test_etb")
	if gs.DayNight != DayNightDay {
		t.Fatalf("§726.2 expected 'day' after daybound ETB, got %q",
			gs.DayNight)
	}
	// Verify the event was emitted.
	found := false
	for _, e := range gs.EventLog {
		if e.Kind == "day_night_change" &&
			e.Details["to_state"] == DayNightDay {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("§726.2 day_night_change event missing from log")
	}
}

func TestDayNight_3a_day_to_night(t *testing.T) {
	gs := NewGameState(2, rand.New(rand.NewSource(0)), nil)
	p := syntheticDaybound(gs, 0)
	MaybeBecomeDay(gs, "test_etb")
	if gs.DayNight != DayNightDay {
		t.Fatalf("precondition: state should be day, got %q", gs.DayNight)
	}
	if p.Transformed {
		t.Fatalf("precondition: werewolf should be front-face-up")
	}
	// Simulate: previous active player cast 0 spells.
	gs.SpellsCastByActiveLastTurn = 0
	EvaluateDayNightAtTurnStart(gs)
	if gs.DayNight != DayNightNight {
		t.Fatalf("§726.3a day+0casts expected night, got %q",
			gs.DayNight)
	}
	if !p.Transformed {
		t.Fatalf("§702.144 daybound werewolf should transform on state→night")
	}
	if p.Card.Name != "Test Wolf Beast" {
		t.Fatalf("after transform expected back-face name, got %q",
			p.Card.Name)
	}
}

func TestDayNight_3a_night_to_day(t *testing.T) {
	gs := NewGameState(2, rand.New(rand.NewSource(0)), nil)
	p := syntheticDaybound(gs, 0)
	MaybeBecomeDay(gs, "test_etb")
	// Flip to night
	gs.SpellsCastByActiveLastTurn = 0
	EvaluateDayNightAtTurnStart(gs)
	if gs.DayNight != DayNightNight || !p.Transformed {
		t.Fatalf("precondition: night + transformed, got %q / %v",
			gs.DayNight, p.Transformed)
	}
	// Active player cast 2+ → back to day.
	gs.SpellsCastByActiveLastTurn = 2
	EvaluateDayNightAtTurnStart(gs)
	if gs.DayNight != DayNightDay {
		t.Fatalf("§726.3a night+2casts expected day, got %q", gs.DayNight)
	}
	if p.Transformed {
		t.Fatalf("§702.145 nightbound werewolf should transform back")
	}
	if p.Card.Name != "Test Wolf Human" {
		t.Fatalf("after retransform expected front-face name, got %q",
			p.Card.Name)
	}
}

func TestDayNight_3a_no_transition(t *testing.T) {
	gs := NewGameState(2, rand.New(rand.NewSource(0)), nil)
	syntheticDaybound(gs, 0)
	MaybeBecomeDay(gs, "test_etb")
	// Day + 1 cast → stays day
	gs.SpellsCastByActiveLastTurn = 1
	EvaluateDayNightAtTurnStart(gs)
	if gs.DayNight != DayNightDay {
		t.Fatalf("day+1cast should stay day, got %q", gs.DayNight)
	}
	// Force to night manually
	SetDayNight(gs, DayNightNight, "force", "test")
	// Night + 1 cast → stays night
	gs.SpellsCastByActiveLastTurn = 1
	EvaluateDayNightAtTurnStart(gs)
	if gs.DayNight != DayNightNight {
		t.Fatalf("night+1cast should stay night, got %q", gs.DayNight)
	}
}

func TestTransform_preserves_counters_and_timestamp(t *testing.T) {
	gs := NewGameState(2, rand.New(rand.NewSource(0)), nil)
	p := syntheticDaybound(gs, 0)
	p.Counters = map[string]int{"+1/+1": 3}
	tsBefore := p.Timestamp
	MaybeBecomeDay(gs, "test_etb")
	gs.SpellsCastByActiveLastTurn = 0
	EvaluateDayNightAtTurnStart(gs)
	// §712.3 — counters preserved.
	if p.Counters["+1/+1"] != 3 {
		t.Fatalf("§712.3 counters lost after transform: got %v", p.Counters)
	}
	// §712.8 — timestamp refreshed.
	if p.Timestamp <= tsBefore {
		t.Fatalf("§712.8 timestamp should advance on transform: %d -> %d",
			tsBefore, p.Timestamp)
	}
}

func TestTransform_nondfc_noop(t *testing.T) {
	gs := NewGameState(2, rand.New(rand.NewSource(0)), nil)
	ast := &gameast.CardAST{Name: "Plain Goblin", FullyParsed: true}
	card := &Card{AST: ast, Name: "Plain Goblin", Owner: 0,
		BasePower: 1, BaseToughness: 1, Types: []string{"creature"}}
	perm := &Permanent{Card: card, Controller: 0, Owner: 0}
	perm.Timestamp = gs.NextTimestamp()
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, perm)
	before := perm.Card.Name
	ok := TransformPermanent(gs, perm, "test")
	if ok {
		t.Fatalf("non-DFC TransformPermanent should return false")
	}
	if perm.Card.Name != before {
		t.Fatalf("non-DFC card name changed: %q -> %q", before, perm.Card.Name)
	}
}

func TestDFCCommanderNameMatching(t *testing.T) {
	// Full oracle DFC name.
	full := "Ral, Monsoon Mage // Ral, Leyline Prodigy"
	if got := DFCFrontFaceName(full); got != "Ral, Monsoon Mage" {
		t.Fatalf("DFCFrontFaceName(%q) = %q, want 'Ral, Monsoon Mage'",
			full, got)
	}
	// Single-slash fallback (Ashling-style decklist).
	if got := DFCFrontFaceName("Ashling, Rekindled / Ashling, Rimebound"); got != "Ashling, Rekindled" {
		t.Fatalf("single-slash DFC front-face parse failed: %q", got)
	}
	// Non-DFC: name as-is.
	if got := DFCFrontFaceName("Lightning Bolt"); got != "Lightning Bolt" {
		t.Fatalf("non-DFC name should pass through: %q", got)
	}
	// DFCCardMatchesName: front-face name matches full oracle name.
	card := &Card{Name: full}
	if !DFCCardMatchesName(card, "Ral, Monsoon Mage") {
		t.Fatalf("DFCCardMatchesName front-face match failed")
	}
	if !DFCCardMatchesName(card, "Ral, Leyline Prodigy") {
		t.Fatalf("DFCCardMatchesName back-face match failed")
	}
	if !DFCCardMatchesName(card, full) {
		t.Fatalf("DFCCardMatchesName full-name match failed")
	}
	if DFCCardMatchesName(card, "Jace, the Mind Sculptor") {
		t.Fatalf("DFCCardMatchesName should NOT match an unrelated card")
	}
}

func TestIsCommanderCard_DFC(t *testing.T) {
	gs := NewGameState(2, rand.New(rand.NewSource(0)), nil)
	gs.CommanderFormat = true
	full := "Ral, Monsoon Mage // Ral, Leyline Prodigy"
	gs.Seats[0].CommanderNames = []string{full}
	// A card whose .Name is the front face only should still be
	// recognized as the commander.
	frontOnly := &Card{Name: "Ral, Monsoon Mage"}
	if !IsCommanderCard(gs, 0, frontOnly) {
		t.Fatalf("IsCommanderCard should accept front-face name")
	}
	// Unrelated card
	other := &Card{Name: "Teferi, Time Raveler"}
	if IsCommanderCard(gs, 0, other) {
		t.Fatalf("IsCommanderCard should reject unrelated card")
	}
	// Full-name card
	fullCard := &Card{Name: full}
	if !IsCommanderCard(gs, 0, fullCard) {
		t.Fatalf("IsCommanderCard should accept full name")
	}
}

// -----------------------------------------------------------------------------
// StripAdventureHalfTypes (§715 / §709 — Adventures + split-type leak)
// -----------------------------------------------------------------------------

func TestStripAdventureHalfTypes_AdventureCreature(t *testing.T) {
	c := &Card{
		Name:     "Adventurous Eater // Have a Bite",
		Types:    []string{"creature", "human", "warlock", "//", "sorcery"},
		TypeLine: "creature — human warlock // sorcery",
	}
	if !StripAdventureHalfTypes(c) {
		t.Fatalf("expected strip to mutate; Types=%v", c.Types)
	}
	if !equalStrSliceDFC(c.Types, []string{"creature", "human", "warlock"}) {
		t.Errorf("Types = %v, want [creature human warlock]", c.Types)
	}
	if c.TypeLine != "creature — human warlock" {
		t.Errorf("TypeLine = %q, want %q", c.TypeLine, "creature — human warlock")
	}
	for _, ty := range c.Types {
		if ty == "instant" || ty == "sorcery" || ty == "//" {
			t.Errorf("post-strip Types must not contain %q; got %v", ty, c.Types)
		}
	}
}

func TestStripAdventureHalfTypes_VirtueOfKnowledgeShape(t *testing.T) {
	// "Virtue of Knowledge // Vantress Visions" — Enchantment // Instant.
	c := &Card{
		Name:     "Virtue of Knowledge // Vantress Visions",
		Types:    []string{"enchantment", "//", "instant"},
		TypeLine: "enchantment // instant",
	}
	if !StripAdventureHalfTypes(c) {
		t.Fatalf("expected strip to fire on Virtue/Vantress shape")
	}
	if !equalStrSliceDFC(c.Types, []string{"enchantment"}) {
		t.Errorf("Types = %v, want [enchantment]", c.Types)
	}
	if c.TypeLine != "enchantment" {
		t.Errorf("TypeLine = %q, want %q", c.TypeLine, "enchantment")
	}
}

func TestStripAdventureHalfTypes_NoOpWhenNoSplit(t *testing.T) {
	c := &Card{
		Name:     "Grizzly Bears",
		Types:    []string{"creature", "bear"},
		TypeLine: "creature — bear",
	}
	before := append([]string(nil), c.Types...)
	if StripAdventureHalfTypes(c) {
		t.Fatalf("expected no-op for vanilla creature")
	}
	if !equalStrSliceDFC(c.Types, before) {
		t.Errorf("Types mutated unexpectedly: %v", c.Types)
	}
}

func TestStripAdventureHalfTypes_NoOpForToken(t *testing.T) {
	c := &Card{
		Types: []string{"token", "creature", "//", "sorcery"},
	}
	if StripAdventureHalfTypes(c) {
		t.Fatalf("token strip should be a no-op")
	}
}

func TestStripAdventureHalfTypes_NoOpWhenFrontHasNoPermanentType(t *testing.T) {
	c := &Card{
		Types:    []string{"instant", "//", "sorcery"},
		TypeLine: "instant // sorcery",
	}
	if StripAdventureHalfTypes(c) {
		t.Fatalf("strip should refuse when front has no permanent type; Types=%v", c.Types)
	}
}

func TestStripAdventureHalfTypes_Idempotent(t *testing.T) {
	c := &Card{
		Types:    []string{"creature", "elf", "//", "instant"},
		TypeLine: "creature — elf // instant",
	}
	if !StripAdventureHalfTypes(c) {
		t.Fatalf("first call should mutate")
	}
	if StripAdventureHalfTypes(c) {
		t.Fatalf("second call should be a no-op")
	}
	if !equalStrSliceDFC(c.Types, []string{"creature", "elf"}) {
		t.Errorf("Types = %v, want [creature elf]", c.Types)
	}
}

func TestEnsureBattlefieldFrontFace_AdventureSurvivesMDFCNoOp(t *testing.T) {
	// Adventure card (not an MDFC) — SwapToBackFace is a no-op (gated by
	// IsMDFC), and the stripper handles it.
	c := &Card{
		Name:     "Adventurous Eater // Have a Bite",
		Types:    []string{"creature", "human", "warlock", "//", "sorcery"},
		TypeLine: "creature — human warlock // sorcery",
	}
	EnsureBattlefieldFrontFace(c)
	if !equalStrSliceDFC(c.Types, []string{"creature", "human", "warlock"}) {
		t.Errorf("post-call Types = %v, want adventure-stripped", c.Types)
	}
}

func TestEnsureBattlefieldFrontFace_MDFCSwapWinsOverStrip(t *testing.T) {
	// Back-face-land MDFC. SwapToBackFace fires first and replaces Types
	// wholesale with BackFaceTypes; the stripper then runs and is a
	// no-op because no "//" remains.
	c := &Card{
		Name:             "Fell the Profane // Fell Mire",
		Types:            []string{"sorcery", "//", "land", "swamp"},
		TypeLine:         "sorcery // land — swamp",
		BackFaceName:     "Fell Mire",
		BackFaceTypes:    []string{"land", "swamp"},
		BackFaceTypeLine: "land — swamp",
	}
	EnsureBattlefieldFrontFace(c)
	if !equalStrSliceDFC(c.Types, []string{"land", "swamp"}) {
		t.Errorf("post-call Types = %v, want [land swamp] from back-face swap", c.Types)
	}
	if c.Name != "Fell Mire" {
		t.Errorf("Name = %q, want %q", c.Name, "Fell Mire")
	}
}

func equalStrSliceDFC(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
