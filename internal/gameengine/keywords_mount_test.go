package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// Round-35 tests for Mount subtype + saddle integration.

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

func mt_makeGame(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(2, rand.New(rand.NewSource(35)), nil)
}

// mt_makeMount builds a Mount permanent on `seat`'s battlefield with
// the given Saddle N cost. Types match the printed OTJ shape
// ("Creature — Beast Mount"). Power 0 because Mounts are typically
// small and we want explicit tapper power to drive saddle cost in
// tests.
func mt_makeMount(seat int, name string, saddleCost int) *Permanent {
	c := &Card{
		Name:          name,
		Owner:         seat,
		Types:         []string{"creature", "beast", "mount"},
		BasePower:     0,
		BaseToughness: 4,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "saddle", Args: []any{saddleCost}},
			},
		},
	}
	return &Permanent{
		Card:       c,
		Controller: seat,
		Owner:      seat,
		Flags:      map[string]int{},
	}
}

// mt_makeSaddleNonMount builds a Saddle-bearing permanent that is NOT
// a Mount (test (e) — a non-mount with the Saddle ability). The card
// has the "saddle" keyword but lacks the "mount" subtype.
func mt_makeSaddleNonMount(seat int, name string, saddleCost int) *Permanent {
	c := &Card{
		Name:          name,
		Owner:         seat,
		Types:         []string{"creature", "human"}, // no "mount"
		BasePower:     0,
		BaseToughness: 4,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "saddle", Args: []any{saddleCost}},
			},
		},
	}
	return &Permanent{
		Card:       c,
		Controller: seat,
		Owner:      seat,
		Flags:      map[string]int{},
	}
}

// mt_makeRider builds an untapped creature with the given power for
// tapping into a saddle activation.
func mt_makeRider(seat int, name string, power int) *Permanent {
	c := &Card{
		Name:          name,
		Owner:         seat,
		Types:         []string{"creature", "human"},
		BasePower:     power,
		BaseToughness: 2,
	}
	return &Permanent{
		Card:       c,
		Controller: seat,
		Owner:      seat,
		Flags:      map[string]int{},
	}
}

func mt_makeNonMountCreature(seat int, name string) *Permanent {
	c := &Card{
		Name:          name,
		Owner:         seat,
		Types:         []string{"creature", "human"},
		BasePower:     1,
		BaseToughness: 1,
	}
	return &Permanent{Card: c, Controller: seat, Owner: seat, Flags: map[string]int{}}
}

func mt_countEvents(gs *GameState, kind string) int {
	n := 0
	for _, ev := range gs.EventLog {
		if ev.Kind == kind {
			n++
		}
	}
	return n
}

// ---------------------------------------------------------------------------
// (a) IsMount true for Mount subtype.
// ---------------------------------------------------------------------------

func TestIsMount_PositiveForMountSubtype(t *testing.T) {
	mount := mt_makeMount(0, "Slickshot Show-Off", 2)
	if !IsMount(mount.Card) {
		t.Fatal("IsMount should be true for a card with the mount subtype")
	}
	if !PermIsMount(mount) {
		t.Fatal("PermIsMount should be true for a Mount permanent")
	}
}

func TestIsMount_TypeLineFallback(t *testing.T) {
	c := &Card{
		Name:     "Mount via TypeLine",
		Types:    []string{"creature", "beast"}, // no mount in Types
		TypeLine: "Creature — Beast Mount",
		AST:      &gameast.CardAST{Name: "Mount via TypeLine"},
	}
	if !IsMount(c) {
		t.Fatal("IsMount should fall back to TypeLine when Types omits the subtype")
	}
}

// ---------------------------------------------------------------------------
// (b) IsMount false for non-mount.
// ---------------------------------------------------------------------------

func TestIsMount_NegativeForNonMount(t *testing.T) {
	plain := mt_makeNonMountCreature(0, "Human Rider")
	if IsMount(plain.Card) {
		t.Fatal("IsMount must be false for a non-mount creature")
	}
	if PermIsMount(plain) {
		t.Fatal("PermIsMount must be false for a non-mount permanent")
	}
	if IsMount(nil) {
		t.Fatal("IsMount(nil) must be false")
	}
	if PermIsMount(nil) {
		t.Fatal("PermIsMount(nil) must be false")
	}
}

// ---------------------------------------------------------------------------
// (c) CountMountsControlled accurate.
// ---------------------------------------------------------------------------

func TestCountMountsControlled_Accurate(t *testing.T) {
	gs := mt_makeGame(t)

	// Seat 0: 2 mounts, 1 plain creature.
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield,
		mt_makeMount(0, "Mount A", 2),
		mt_makeMount(0, "Mount B", 3),
		mt_makeNonMountCreature(0, "Plain Goblin"),
	)
	// Seat 1: 1 mount.
	gs.Seats[1].Battlefield = append(gs.Seats[1].Battlefield,
		mt_makeMount(1, "Enemy Mount", 1),
	)

	if got := CountMountsControlled(gs, 0); got != 2 {
		t.Fatalf("seat 0 mount count = %d, want 2", got)
	}
	if got := CountMountsControlled(gs, 1); got != 1 {
		t.Fatalf("seat 1 mount count = %d, want 1", got)
	}
	if got := CountMountsControlled(gs, -1); got != 0 {
		t.Fatalf("invalid seat should return 0; got %d", got)
	}
	if got := CountMountsControlled(nil, 0); got != 0 {
		t.Fatalf("nil game should return 0; got %d", got)
	}
}

func TestCountMountsControlled_ExcludesPhasedOut(t *testing.T) {
	gs := mt_makeGame(t)
	m1 := mt_makeMount(0, "Active Mount", 2)
	m2 := mt_makeMount(0, "Phased Mount", 2)
	m2.PhasedOut = true
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, m1, m2)

	if got := CountMountsControlled(gs, 0); got != 1 {
		t.Fatalf("phased-out mount should not count; got %d, want 1", got)
	}
}

// ---------------------------------------------------------------------------
// (d) Mount-saddled trigger fires from SaddleMount.
// ---------------------------------------------------------------------------

func TestSaddleMount_FiresMountSaddledTrigger(t *testing.T) {
	gs := mt_makeGame(t)
	mount := mt_makeMount(0, "Stage-Coach Mount", 2)
	rider := mt_makeRider(0, "Big Rider", 3)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, mount, rider)

	if !SaddleMount(gs, 0, mount, []*Permanent{rider}) {
		t.Fatal("SaddleMount should succeed (rider power 3 ≥ saddle cost 2)")
	}
	if mt_countEvents(gs, "mount_saddled") != 1 {
		t.Fatalf("expected 1 mount_saddled event from SaddleMount; got %d",
			mt_countEvents(gs, "mount_saddled"))
	}
	if mount.Flags["saddled"] != 1 {
		t.Fatal("mount should be saddled after successful SaddleMount")
	}
}

func TestSaddleMount_NoTriggerOnFailure(t *testing.T) {
	gs := mt_makeGame(t)
	mount := mt_makeMount(0, "Stage-Coach Mount", 4) // expensive saddle
	weak := mt_makeRider(0, "Weak Rider", 1)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, mount, weak)

	if SaddleMount(gs, 0, mount, []*Permanent{weak}) {
		t.Fatal("SaddleMount should fail when rider power < cost")
	}
	if mt_countEvents(gs, "mount_saddled") != 0 {
		t.Fatalf("no mount_saddled event should fire on failed saddle; got %d",
			mt_countEvents(gs, "mount_saddled"))
	}
}

func TestActivateSaddle_FiresMountSaddledTrigger(t *testing.T) {
	gs := mt_makeGame(t)
	mount := mt_makeMount(0, "Greedy Saddle Mount", 3)
	rider := mt_makeRider(0, "Big Rider", 4)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, mount, rider)

	// ActivateSaddle takes the cost directly (not from the keyword).
	if !ActivateSaddle(gs, mount, 3) {
		t.Fatal("ActivateSaddle should succeed")
	}
	if mt_countEvents(gs, "mount_saddled") != 1 {
		t.Fatalf("expected 1 mount_saddled event from ActivateSaddle; got %d",
			mt_countEvents(gs, "mount_saddled"))
	}
}

// ---------------------------------------------------------------------------
// (e) Does NOT fire for non-mount saddled (e.g. a non-mount that
// somehow has the Saddle ability).
// ---------------------------------------------------------------------------

func TestSaddleMount_NoTriggerForNonMount(t *testing.T) {
	gs := mt_makeGame(t)
	// "Mount" that's actually a human — has Saddle but not the subtype.
	notAMount := mt_makeSaddleNonMount(0, "Fake Mount", 2)
	rider := mt_makeRider(0, "Big Rider", 3)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, notAMount, rider)

	if !SaddleMount(gs, 0, notAMount, []*Permanent{rider}) {
		t.Fatal("SaddleMount should succeed even on non-Mount (cost paid)")
	}
	// Saddled flag is set on the carrier (because the cost was paid),
	// but the Mount-saddled trigger should NOT fire because the carrier
	// isn't a Mount.
	if notAMount.Flags["saddled"] != 1 {
		t.Fatal("saddled flag should still be set on the non-mount")
	}
	if mt_countEvents(gs, "mount_saddled") != 0 {
		t.Fatalf("mount_saddled event must NOT fire for non-Mount carrier; got %d",
			mt_countEvents(gs, "mount_saddled"))
	}
}

func TestActivateSaddle_NoTriggerForNonMount(t *testing.T) {
	gs := mt_makeGame(t)
	notAMount := mt_makeSaddleNonMount(0, "Fake Mount", 2)
	rider := mt_makeRider(0, "Big Rider", 4)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, notAMount, rider)

	if !ActivateSaddle(gs, notAMount, 2) {
		t.Fatal("ActivateSaddle should succeed for non-mount with Saddle ability")
	}
	if mt_countEvents(gs, "mount_saddled") != 0 {
		t.Fatalf("ActivateSaddle on non-mount must not publish mount_saddled; got %d",
			mt_countEvents(gs, "mount_saddled"))
	}
}

// ---------------------------------------------------------------------------
// Bonus: nil-safety on FireMountSaddledTriggers.
// ---------------------------------------------------------------------------

func TestFireMountSaddledTriggers_NilSafe(t *testing.T) {
	FireMountSaddledTriggers(nil, nil)
	gs := mt_makeGame(t)
	FireMountSaddledTriggers(gs, nil)
	// A nil-card perm shouldn't crash either.
	FireMountSaddledTriggers(gs, &Permanent{Controller: 0})
}
