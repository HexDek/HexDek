package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// Tests for the §702.178 max-speed rider resolver hook + the
// SpeedDamageReporter wrapper (round 25).
//
// Required coverage:
//   (a) rider applies when speed=4
//   (b) skipped at speed<4
//   (c) once-per-turn AdvanceSpeed limit across multiple damage events
//   (d) HasMaxSpeedRider detects rider text

func newRiderGame(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(2, rand.New(rand.NewSource(23)), nil)
}

func newRiderCard(name, oracleText string) *Card {
	c := &Card{
		Name:          name,
		Types:         []string{"creature"},
		BasePower:     2,
		BaseToughness: 2,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: oracleText},
			},
		},
	}
	return c
}

func newRiderPerm(seat int, card *Card) *Permanent {
	card.Owner = seat
	return &Permanent{
		Card:       card,
		Controller: seat,
		Owner:      seat,
		Flags:      map[string]int{},
	}
}

// ---------------------------------------------------------------------------
// (d) HasMaxSpeedRider detects rider text.
// ---------------------------------------------------------------------------

func TestHasMaxSpeedRider_OracleTextEmDash(t *testing.T) {
	c := newRiderCard("Hotwire Pilot", "Flying. Max speed — Hotwire Pilot gets +2/+2.")
	if !HasMaxSpeedRider(c) {
		t.Fatal("HasMaxSpeedRider should detect em-dash rider in oracle text")
	}
}

func TestHasMaxSpeedRider_OracleTextAsciiHyphen(t *testing.T) {
	c := newRiderCard("Drift Racer", "Max speed - Draw a card.")
	if !HasMaxSpeedRider(c) {
		t.Fatal("HasMaxSpeedRider should detect ASCII-hyphen rider as a fallback")
	}
}

func TestHasMaxSpeedRider_NegativeOnPlainCard(t *testing.T) {
	c := newRiderCard("Plain Goblin", "Haste. Goblin Berserker enters tapped.")
	if HasMaxSpeedRider(c) {
		t.Fatal("HasMaxSpeedRider should not fire on a card without the rider line")
	}
	if HasMaxSpeedRider(nil) {
		t.Fatal("HasMaxSpeedRider(nil) should be false")
	}
}

func TestHasMaxSpeedRider_KeywordAST(t *testing.T) {
	c := &Card{
		Name: "Tagged Speedster",
		AST: &gameast.CardAST{
			Name: "Tagged Speedster",
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "max speed"},
			},
		},
	}
	if !HasMaxSpeedRider(c) {
		t.Fatal("HasMaxSpeedRider should detect a keyword-tagged rider even with no oracle text")
	}
}

// ---------------------------------------------------------------------------
// (a) Rider applies when speed=4.
// ---------------------------------------------------------------------------

func TestApplyMaxSpeedRider_FiresAtMaxSpeed(t *testing.T) {
	gs := newRiderGame(t)
	gs.Seats[0].Speed = MaxSpeedCap

	c := newRiderCard("Boosted Racer", "Max speed — Boosted Racer gets +3/+3.")
	p := newRiderPerm(0, c)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, p)

	if !ApplyMaxSpeedRider(gs, p) {
		t.Fatal("ApplyMaxSpeedRider should fire when controller is at MaxSpeedCap")
	}

	// Look for the max_speed_rider event in the log.
	found := false
	for _, ev := range gs.EventLog {
		if ev.Kind == "max_speed_rider" && ev.Seat == 0 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected a max_speed_rider event in the log")
	}
}

// ---------------------------------------------------------------------------
// (b) Skipped at speed<4.
// ---------------------------------------------------------------------------

func TestApplyMaxSpeedRider_SkipsBelowMaxSpeed(t *testing.T) {
	gs := newRiderGame(t)

	c := newRiderCard("Boosted Racer", "Max speed — Boosted Racer gets +3/+3.")
	p := newRiderPerm(0, c)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, p)

	for speed := 0; speed < MaxSpeedCap; speed++ {
		gs.Seats[0].Speed = speed
		if ApplyMaxSpeedRider(gs, p) {
			t.Fatalf("rider must not fire at speed=%d", speed)
		}
	}
}

func TestApplyMaxSpeedRider_SkipsWhenCardLacksRider(t *testing.T) {
	gs := newRiderGame(t)
	gs.Seats[0].Speed = MaxSpeedCap

	c := newRiderCard("Plain Racer", "Plain Racer has haste.")
	p := newRiderPerm(0, c)

	if ApplyMaxSpeedRider(gs, p) {
		t.Fatal("rider must not fire on a card without the rider line, even at max speed")
	}
}

// ---------------------------------------------------------------------------
// (c) Once-per-turn AdvanceSpeed limit across multiple damage events.
// ---------------------------------------------------------------------------

func TestSpeedDamageReporter_OnceperTurnAcrossMultipleEvents(t *testing.T) {
	gs := newRiderGame(t)

	// First event: should advance.
	if !SpeedDamageReporter(gs, 0) {
		t.Fatal("first damage event should advance speed (0 → 1)")
	}
	if SpeedOf(gs, 0) != 1 {
		t.Fatalf("speed after first event = %d, want 1", SpeedOf(gs, 0))
	}

	// Three more events same turn: must NOT advance further.
	for i := 0; i < 3; i++ {
		if SpeedDamageReporter(gs, 0) {
			t.Fatalf("damage event #%d should be no-op (once-per-turn gate)", i+2)
		}
	}
	if SpeedOf(gs, 0) != 1 {
		t.Fatalf("speed must stay at 1 across multiple events same turn; got %d", SpeedOf(gs, 0))
	}

	// Simulate next turn: reset gate and advance again.
	ResetSpeedAdvancedFlag(gs, 0)
	if !SpeedDamageReporter(gs, 0) {
		t.Fatal("after turn reset, next damage event should advance again")
	}
	if SpeedOf(gs, 0) != 2 {
		t.Fatalf("speed after next-turn advance = %d, want 2", SpeedOf(gs, 0))
	}
}

func TestSpeedDamageReporter_PerControllerIndependent(t *testing.T) {
	gs := newRiderGame(t)

	// Two different controllers can both advance on the same turn.
	if !SpeedDamageReporter(gs, 0) {
		t.Fatal("seat 0 damage event should advance speed")
	}
	if !SpeedDamageReporter(gs, 1) {
		t.Fatal("seat 1 damage event should advance speed (independent gate)")
	}
	if SpeedOf(gs, 0) != 1 || SpeedOf(gs, 1) != 1 {
		t.Fatalf("both seats should be at speed=1, got 0=%d, 1=%d",
			SpeedOf(gs, 0), SpeedOf(gs, 1))
	}
}

// ---------------------------------------------------------------------------
// Bonus: resolveSequence hook fires the rider exactly once per outer call.
// ---------------------------------------------------------------------------

func TestResolveSequenceHook_FiresRiderOnce(t *testing.T) {
	gs := newRiderGame(t)
	gs.Seats[0].Speed = MaxSpeedCap

	c := newRiderCard("Outer Racer", "Max speed — Outer Racer gets +1/+1.")
	p := newRiderPerm(0, c)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, p)

	// Empty outer sequence — the rider hook still needs to fire once at end.
	resolveSequence(gs, p, &gameast.Sequence{Items: nil})

	count := 0
	for _, ev := range gs.EventLog {
		if ev.Kind == "max_speed_rider" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 max_speed_rider event, got %d", count)
	}
}

func TestResolveSequenceHook_NestedSequenceFiresRiderOnce(t *testing.T) {
	gs := newRiderGame(t)
	gs.Seats[0].Speed = MaxSpeedCap

	c := newRiderCard("Nested Racer", "Max speed — Draw a card.")
	p := newRiderPerm(0, c)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, p)

	// Outer Sequence whose only item is another Sequence — the depth
	// guard should keep the rider from firing twice.
	inner := &gameast.Sequence{Items: []gameast.Effect{}}
	outer := &gameast.Sequence{Items: []gameast.Effect{inner}}
	resolveSequence(gs, p, outer)

	count := 0
	for _, ev := range gs.EventLog {
		if ev.Kind == "max_speed_rider" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("nested sequence should fire rider exactly once; got %d", count)
	}
}
