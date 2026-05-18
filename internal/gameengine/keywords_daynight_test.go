package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Daybound / Nightbound tests — CR §702.149 / §702.150 + §726.2 / §730.2a
// ---------------------------------------------------------------------------

func newDayNightGame(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(2, rand.New(rand.NewSource(46)), nil)
}

// makeWerewolf builds a synthetic DFC permanent with daybound on the
// front face and nightbound on the back face. Parser-free so the test
// doesn't depend on the oracle corpus.
func makeWerewolf(gs *GameState, seat int, name string) *Permanent {
	front := &gameast.CardAST{
		Name: name + " (human)",
		Abilities: []gameast.Ability{
			&gameast.Keyword{Name: "daybound", Raw: "Daybound"},
		},
		FullyParsed: true,
	}
	back := &gameast.CardAST{
		Name: name + " (wolf)",
		Abilities: []gameast.Ability{
			&gameast.Keyword{Name: "nightbound", Raw: "Nightbound"},
		},
		FullyParsed: true,
	}
	card := &Card{
		AST:           front,
		Name:          name + " (human)",
		Owner:         seat,
		BasePower:     2,
		BaseToughness: 2,
		Types:         []string{"creature"},
		CMC:           2,
		TypeLine:      "Creature — Human Werewolf",
	}
	perm := &Permanent{
		Card:       card,
		Controller: seat,
		Owner:      seat,
		Flags:      map[string]int{},
	}
	perm.Timestamp = gs.NextTimestamp()
	InitDFCFaces(perm, front, back, name+" (human)", name+" (wolf)")
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, perm)
	return perm
}

// ---------------------------------------------------------------------------
// HasDaybound / HasNightbound / Perm*
// ---------------------------------------------------------------------------

func TestHasDaybound_Detects(t *testing.T) {
	gs := newDayNightGame(t)
	w := makeWerewolf(gs, 0, "Tovolar's Huntmaster")
	if !HasDaybound(w.Card) {
		t.Fatal("HasDaybound should detect daybound on front face")
	}
	if HasNightbound(w.Card) {
		t.Fatal("front-face card should not also report nightbound")
	}
	if !PermHasDaybound(w) {
		t.Fatal("PermHasDaybound should be true while front face is active")
	}
	if PermHasNightbound(w) {
		t.Fatal("PermHasNightbound should be false while front face is active")
	}
}

// ---------------------------------------------------------------------------
// (a) Game starts neither, becomes Day when first daybound ETBs
// ---------------------------------------------------------------------------

func TestDayNight_StartsNeither_BecomesDayOnFirstDayboundETB(t *testing.T) {
	gs := newDayNightGame(t)

	if !IsNeitherDayNorNight(gs) {
		t.Fatalf("§726.1 expected neither at game start, got %q", gs.DayNight)
	}
	if IsDay(gs) || IsNight(gs) {
		t.Fatal("IsDay/IsNight should be false at game start")
	}

	w := makeWerewolf(gs, 0, "Reckless Stormseeker")
	// Production ETB path runs FirePermanentETBTriggers, which calls
	// OnDayboundOrNightboundETB. Drive the dispatcher directly.
	FirePermanentETBTriggers(gs, w)

	if !IsDay(gs) {
		t.Fatalf("§726.2 expected day after first daybound ETB, got %q",
			gs.DayNight)
	}

	// Subsequent daybound ETBs do not re-trigger §726.2 (idempotent).
	before := len(gs.EventLog)
	FirePermanentETBTriggers(gs, makeWerewolf(gs, 1, "Tovolar's Magehunter"))
	for i := before; i < len(gs.EventLog); i++ {
		if gs.EventLog[i].Kind == "day_night_change" {
			t.Fatal("§726.2 should not re-fire once state is set")
		}
	}
}

func TestDayNight_VanillaETBDoesNotChangeState(t *testing.T) {
	gs := newDayNightGame(t)
	// Vanilla 1/1 creature ETB — no daybound/nightbound on either face.
	vanilla := &Permanent{
		Card: &Card{
			Name:          "Grizzly Bears",
			Owner:         0,
			Types:         []string{"creature"},
			BasePower:     2,
			BaseToughness: 2,
			AST: &gameast.CardAST{
				Name:      "Grizzly Bears",
				Abilities: []gameast.Ability{},
			},
		},
		Controller: 0,
		Owner:      0,
		Flags:      map[string]int{},
		Timestamp:  gs.NextTimestamp(),
	}
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, vanilla)

	FirePermanentETBTriggers(gs, vanilla)

	if !IsNeitherDayNorNight(gs) {
		t.Fatalf("§726.2 must not flip on a non-daybound ETB; got %q",
			gs.DayNight)
	}
}

// ---------------------------------------------------------------------------
// (b) Day→Night on 0-spell turn
// ---------------------------------------------------------------------------

func TestDayNight_DayToNight_OnZeroSpellTurn(t *testing.T) {
	gs := newDayNightGame(t)
	w := makeWerewolf(gs, 0, "Tovolar's Huntmaster")
	FirePermanentETBTriggers(gs, w)

	if !IsDay(gs) {
		t.Fatalf("precondition: expected day, got %q", gs.DayNight)
	}
	if w.Transformed {
		t.Fatal("precondition: werewolf should be front-face-up")
	}

	// Simulate: previous active player cast 0 spells last turn.
	gs.SpellsCastByActiveLastTurn = 0
	EvaluateDayNightAtTurnStart(gs)

	if !IsNight(gs) {
		t.Fatalf("§730.2a day+0casts expected night, got %q", gs.DayNight)
	}
	// Daybound creature should have transformed to its nightbound back face.
	if !w.Transformed {
		t.Fatal("§702.149 daybound creature should transform when state becomes night")
	}
	if !PermHasNightbound(w) {
		t.Fatal("active face after transform should carry nightbound")
	}
}

// ---------------------------------------------------------------------------
// (c) Night→Day on 2+ spell turn
// ---------------------------------------------------------------------------

func TestDayNight_NightToDay_OnTwoPlusSpellTurn(t *testing.T) {
	gs := newDayNightGame(t)
	w := makeWerewolf(gs, 0, "Tovolar's Huntmaster")
	FirePermanentETBTriggers(gs, w)

	// Force the world to night.
	gs.SpellsCastByActiveLastTurn = 0
	EvaluateDayNightAtTurnStart(gs)
	if !IsNight(gs) {
		t.Fatalf("setup: expected night, got %q", gs.DayNight)
	}
	if !w.Transformed {
		t.Fatal("setup: werewolf should be on its back face under night")
	}

	// Now simulate active player casting 2 spells last turn.
	gs.SpellsCastByActiveLastTurn = 2
	EvaluateDayNightAtTurnStart(gs)

	if !IsDay(gs) {
		t.Fatalf("§730.2a night+2casts expected day, got %q", gs.DayNight)
	}
	// Nightbound active face should have transformed back to its daybound
	// front face.
	if w.Transformed {
		t.Fatal("§702.150 nightbound creature should transform when state becomes day")
	}
	if !PermHasDaybound(w) {
		t.Fatal("active face after transform-back should carry daybound")
	}
}

// Night→Day requires 2 or more — exactly 1 spell does NOT flip.
func TestDayNight_NightHoldsOnOneSpell(t *testing.T) {
	gs := newDayNightGame(t)
	w := makeWerewolf(gs, 0, "Tovolar's Huntmaster")
	FirePermanentETBTriggers(gs, w)
	gs.SpellsCastByActiveLastTurn = 0
	EvaluateDayNightAtTurnStart(gs)
	if !IsNight(gs) {
		t.Fatalf("setup: expected night, got %q", gs.DayNight)
	}
	gs.SpellsCastByActiveLastTurn = 1
	EvaluateDayNightAtTurnStart(gs)
	if !IsNight(gs) {
		t.Fatalf("§730.2a night+1cast must stay night, got %q", gs.DayNight)
	}
	if !w.Transformed {
		t.Fatal("werewolf should remain on its back face while state stays night")
	}
}

// Day→Night requires 0 — 1+ spells leaves day alone.
func TestDayNight_DayHoldsOnAnySpellCast(t *testing.T) {
	gs := newDayNightGame(t)
	w := makeWerewolf(gs, 0, "Tovolar's Huntmaster")
	FirePermanentETBTriggers(gs, w)
	gs.SpellsCastByActiveLastTurn = 1
	EvaluateDayNightAtTurnStart(gs)
	if !IsDay(gs) {
		t.Fatalf("§730.2a day+1cast must stay day, got %q", gs.DayNight)
	}
	if w.Transformed {
		t.Fatal("werewolf should stay on its front face while state stays day")
	}
}

// ---------------------------------------------------------------------------
// (d) Daybound creatures auto-transform with day/night flips
// ---------------------------------------------------------------------------

func TestDayNight_AllWerewolvesTransformOnFlip(t *testing.T) {
	gs := newDayNightGame(t)
	// Multiple werewolves across both battlefields.
	a := makeWerewolf(gs, 0, "Tovolar's Huntmaster")
	b := makeWerewolf(gs, 0, "Reckless Stormseeker")
	c := makeWerewolf(gs, 1, "Arlinn the Pack's Hope")
	FirePermanentETBTriggers(gs, a)
	FirePermanentETBTriggers(gs, b)
	FirePermanentETBTriggers(gs, c)

	if !IsDay(gs) {
		t.Fatalf("precondition: day, got %q", gs.DayNight)
	}
	for _, w := range []*Permanent{a, b, c} {
		if w.Transformed {
			t.Fatalf("%s should start on its daybound front face",
				w.Card.DisplayName())
		}
	}

	// Day → Night: every daybound werewolf transforms.
	gs.SpellsCastByActiveLastTurn = 0
	EvaluateDayNightAtTurnStart(gs)
	if !IsNight(gs) {
		t.Fatalf("§730.2a expected night, got %q", gs.DayNight)
	}
	for _, w := range []*Permanent{a, b, c} {
		if !w.Transformed {
			t.Fatalf("%s should have transformed on day→night flip",
				w.Card.DisplayName())
		}
	}

	// Night → Day: every nightbound werewolf transforms back.
	gs.SpellsCastByActiveLastTurn = 2
	EvaluateDayNightAtTurnStart(gs)
	if !IsDay(gs) {
		t.Fatalf("§730.2a expected day, got %q", gs.DayNight)
	}
	for _, w := range []*Permanent{a, b, c} {
		if w.Transformed {
			t.Fatalf("%s should have transformed back on night→day flip",
				w.Card.DisplayName())
		}
	}
}

// SetDayNight to the same value must be a no-op and must NOT call
// ApplyDayboundNightboundTransforms (which would needlessly flip
// werewolves that are already on the right face).
func TestDayNight_SetSameStateIsNoOp(t *testing.T) {
	gs := newDayNightGame(t)
	w := makeWerewolf(gs, 0, "Tovolar's Huntmaster")
	FirePermanentETBTriggers(gs, w)
	beforeEvents := len(gs.EventLog)
	beforeTransformed := w.Transformed

	SetDayNight(gs, DayNightDay, "no-op-test", "726.1")

	if len(gs.EventLog) != beforeEvents {
		t.Fatal("SetDayNight to current state must not emit an event")
	}
	if w.Transformed != beforeTransformed {
		t.Fatal("SetDayNight to current state must not transform any werewolf")
	}
}

// ---------------------------------------------------------------------------
// Nil safety
// ---------------------------------------------------------------------------

func TestDayNight_NilSafe(t *testing.T) {
	if HasDaybound(nil) {
		t.Fatal("HasDaybound(nil) should be false")
	}
	if HasNightbound(nil) {
		t.Fatal("HasNightbound(nil) should be false")
	}
	if PermHasDaybound(nil) {
		t.Fatal("PermHasDaybound(nil) should be false")
	}
	if PermHasNightbound(nil) {
		t.Fatal("PermHasNightbound(nil) should be false")
	}
	if IsDay(nil) || IsNight(nil) {
		t.Fatal("IsDay/IsNight(nil) should be false")
	}
	if !IsNeitherDayNorNight(nil) {
		t.Fatal("IsNeitherDayNorNight(nil) should be true (no game = no state)")
	}
	OnDayboundOrNightboundETB(nil, nil) // must not panic
}
