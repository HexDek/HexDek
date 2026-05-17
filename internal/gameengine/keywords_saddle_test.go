package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Saddle tests — CR §702.171
// ---------------------------------------------------------------------------

func newSaddleGame(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(2, rand.New(rand.NewSource(49)), nil)
}

// mountPerm builds a Mount permanent with "Saddle N" baked into the AST.
func mountPerm(gs *GameState, owner int, name string, saddleN int) *Permanent {
	card := &Card{
		Name:          name,
		Owner:         owner,
		Types:         []string{"creature", "mount"},
		TypeLine:      "Creature — Mount",
		BasePower:     2,
		BaseToughness: 2,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "saddle", Args: []interface{}{float64(saddleN)}},
			},
		},
	}
	perm := &Permanent{
		Card:       card,
		Controller: owner,
		Owner:      owner,
		Flags:      map[string]int{},
		Timestamp:  gs.NextTimestamp(),
	}
	gs.Seats[owner].Battlefield = append(gs.Seats[owner].Battlefield, perm)
	return perm
}

// creaturePerm builds a vanilla creature with the given power/toughness.
func creaturePerm(gs *GameState, owner int, name string, power int) *Permanent {
	card := &Card{
		Name:          name,
		Owner:         owner,
		Types:         []string{"creature"},
		BasePower:     power,
		BaseToughness: power,
		AST:           &gameast.CardAST{Name: name},
	}
	perm := &Permanent{
		Card:       card,
		Controller: owner,
		Owner:      owner,
		Flags:      map[string]int{},
		Timestamp:  gs.NextTimestamp(),
	}
	gs.Seats[owner].Battlefield = append(gs.Seats[owner].Battlefield, perm)
	return perm
}

// ---------------------------------------------------------------------------
// HasSaddle / SaddleCost / PermHasSaddle
// ---------------------------------------------------------------------------

func TestHasSaddle_Detects(t *testing.T) {
	gs := newSaddleGame(t)
	m := mountPerm(gs, 0, "Bristly Bill, Spine Sower", 3)
	if !HasSaddle(m.Card) {
		t.Fatal("HasSaddle should detect saddle keyword")
	}
	if !PermHasSaddle(m) {
		t.Fatal("PermHasSaddle should be true on a mount with the saddle keyword")
	}
	if SaddleCost(m.Card) != 3 {
		t.Fatalf("SaddleCost = %d, want 3", SaddleCost(m.Card))
	}
}

func TestHasSaddle_Negative(t *testing.T) {
	gs := newSaddleGame(t)
	c := creaturePerm(gs, 0, "Grizzly Bears", 2)
	if HasSaddle(c.Card) {
		t.Fatal("HasSaddle should be false for a vanilla creature")
	}
	if PermHasSaddle(c) {
		t.Fatal("PermHasSaddle should be false for a vanilla creature")
	}
	if SaddleCost(c.Card) != 0 {
		t.Fatalf("SaddleCost = %d, want 0", SaddleCost(c.Card))
	}
}

// ---------------------------------------------------------------------------
// (a) Sufficient power saddles + taps
// ---------------------------------------------------------------------------

func TestSaddleMount_SufficientPowerSucceeds(t *testing.T) {
	gs := newSaddleGame(t)
	mount := mountPerm(gs, 0, "Roxanne, Starfall Savant", 3)
	c1 := creaturePerm(gs, 0, "Goblin Brigand", 2)
	c2 := creaturePerm(gs, 0, "Goblin Brigand", 2)

	if !SaddleMount(gs, 0, mount, []*Permanent{c1, c2}) {
		t.Fatal("SaddleMount with total power 4 vs N=3 should succeed")
	}
	if !PermIsSaddled(mount) {
		t.Fatal("mount should report saddled after success")
	}
	if !c1.Tapped || !c2.Tapped {
		t.Fatal("all chosen tappers should be tapped")
	}
	// Mount itself was not tapped (saddle taps the tappers, not the mount).
	if mount.Tapped {
		t.Fatal("the mount itself should not be tapped by saddling")
	}
	// SaddlersThisTurn records the chosen creatures.
	if len(mount.SaddlersThisTurn) != 2 {
		t.Fatalf("SaddlersThisTurn = %d, want 2", len(mount.SaddlersThisTurn))
	}
}

func TestSaddleMount_ExactlyEnoughPowerSucceeds(t *testing.T) {
	gs := newSaddleGame(t)
	mount := mountPerm(gs, 0, "Mount", 3)
	c := creaturePerm(gs, 0, "Lhurgoyf", 3)
	if !SaddleMount(gs, 0, mount, []*Permanent{c}) {
		t.Fatal("exactly N power should succeed")
	}
	if !PermIsSaddled(mount) {
		t.Fatal("mount should be saddled")
	}
}

// ---------------------------------------------------------------------------
// (b) Insufficient power refused
// ---------------------------------------------------------------------------

func TestSaddleMount_InsufficientPowerRefused(t *testing.T) {
	gs := newSaddleGame(t)
	mount := mountPerm(gs, 0, "Mount", 4)
	c1 := creaturePerm(gs, 0, "Bird", 1)
	c2 := creaturePerm(gs, 0, "Bird", 1)

	if SaddleMount(gs, 0, mount, []*Permanent{c1, c2}) {
		t.Fatal("total power 2 vs N=4 should fail")
	}
	if PermIsSaddled(mount) {
		t.Fatal("mount must not be saddled on failure")
	}
	// Atomic: tappers must not be tapped on failure.
	if c1.Tapped || c2.Tapped {
		t.Fatal("failed SaddleMount must not tap any creature (atomic)")
	}
	if len(mount.SaddlersThisTurn) != 0 {
		t.Fatal("SaddlersThisTurn must remain empty on failure")
	}
}

func TestSaddleMount_EmptyTappersRefused(t *testing.T) {
	gs := newSaddleGame(t)
	mount := mountPerm(gs, 0, "Mount", 1)
	if SaddleMount(gs, 0, mount, nil) {
		t.Fatal("nil tappers (total power 0) should fail")
	}
	if PermIsSaddled(mount) {
		t.Fatal("mount must not be saddled when no tappers were chosen")
	}
}

// ---------------------------------------------------------------------------
// (c) Tapped creature can't saddle
// ---------------------------------------------------------------------------

func TestSaddleMount_TappedCreatureRefused(t *testing.T) {
	gs := newSaddleGame(t)
	mount := mountPerm(gs, 0, "Mount", 2)
	tappedAttacker := creaturePerm(gs, 0, "Already-Tapped Beast", 4)
	tappedAttacker.Tapped = true

	if SaddleMount(gs, 0, mount, []*Permanent{tappedAttacker}) {
		t.Fatal("a tapped creature cannot saddle (§702.171a)")
	}
	if PermIsSaddled(mount) {
		t.Fatal("mount must not be saddled when the tapper was already tapped")
	}
}

func TestSaddleMount_MountTappingItselfRefused(t *testing.T) {
	gs := newSaddleGame(t)
	mount := mountPerm(gs, 0, "Self-Saddler", 1)
	mount.Card.BasePower = 10 // could meet the cost if allowed

	if SaddleMount(gs, 0, mount, []*Permanent{mount}) {
		t.Fatal("a mount may not saddle itself (§702.171a 'other')")
	}
	if PermIsSaddled(mount) {
		t.Fatal("mount must not be saddled by self-tap attempt")
	}
	if mount.Tapped {
		t.Fatal("self-saddle attempt must not tap the mount")
	}
}

func TestSaddleMount_OpponentCreatureRefused(t *testing.T) {
	gs := newSaddleGame(t)
	mount := mountPerm(gs, 0, "Mount", 2)
	opp := creaturePerm(gs, 1, "Opponent's Beast", 4)

	if SaddleMount(gs, 0, mount, []*Permanent{opp}) {
		t.Fatal("a creature you don't control cannot saddle your mount")
	}
	if opp.Tapped {
		t.Fatal("failed saddle must not tap opponent's creature")
	}
}

func TestSaddleMount_NonCreatureRefused(t *testing.T) {
	gs := newSaddleGame(t)
	mount := mountPerm(gs, 0, "Mount", 1)
	// Artifact, not a creature.
	artifact := &Permanent{
		Card: &Card{
			Name:  "Sol Ring",
			Owner: 0,
			Types: []string{"artifact"},
			AST:   &gameast.CardAST{Name: "Sol Ring"},
		},
		Controller: 0,
		Owner:      0,
		Flags:      map[string]int{},
		Timestamp:  gs.NextTimestamp(),
	}
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, artifact)

	if SaddleMount(gs, 0, mount, []*Permanent{artifact}) {
		t.Fatal("a non-creature permanent cannot saddle (§702.171a creatures only)")
	}
}

func TestSaddleMount_DuplicateTapperRefused(t *testing.T) {
	gs := newSaddleGame(t)
	mount := mountPerm(gs, 0, "Mount", 3)
	c := creaturePerm(gs, 0, "Big Beast", 4) // alone would succeed
	// Passing the same pointer twice attempts to double-count its power
	// against the saddle threshold — must be rejected.
	if SaddleMount(gs, 0, mount, []*Permanent{c, c}) {
		t.Fatal("duplicate tapper must be refused (no double-counting)")
	}
	if c.Tapped {
		t.Fatal("rejected duplicate-tapper saddle must not tap the creature")
	}
}

// ---------------------------------------------------------------------------
// (d) Flag cleared at EOT
// ---------------------------------------------------------------------------

func TestSaddleMount_UnsaddleAtEOTClearsDesignation(t *testing.T) {
	gs := newSaddleGame(t)
	mount := mountPerm(gs, 0, "Mount", 2)
	c := creaturePerm(gs, 0, "Saddler", 3)

	if !SaddleMount(gs, 0, mount, []*Permanent{c}) {
		t.Fatal("setup: saddle should succeed")
	}
	if !PermIsSaddled(mount) {
		t.Fatal("setup: mount should be saddled")
	}
	if len(mount.SaddlersThisTurn) != 1 {
		t.Fatal("setup: SaddlersThisTurn should record the tapper")
	}

	UnsaddleAtEOT(gs)

	if PermIsSaddled(mount) {
		t.Fatal("§702.171b: saddled designation must wear off at EOT")
	}
	if _, ok := mount.Flags["saddled"]; ok {
		t.Fatal("flag should be deleted, not zeroed")
	}
	if len(mount.SaddlersThisTurn) != 0 {
		t.Fatal("SaddlersThisTurn should be cleared at EOT")
	}
	// Tapped state on the saddler does NOT reset at EOT — the tapper
	// stays tapped until its controller's next untap step. This test
	// guards against an accidental over-clear in UnsaddleAtEOT.
	if !c.Tapped {
		t.Fatal("UnsaddleAtEOT must not untap saddlers — that's the untap step's job")
	}
}

func TestUnsaddleAtEOT_NilSafeAndIdempotent(t *testing.T) {
	UnsaddleAtEOT(nil) // must not panic
	gs := newSaddleGame(t)
	UnsaddleAtEOT(gs) // empty board, must not panic
	mountPerm(gs, 0, "Mount", 1)
	UnsaddleAtEOT(gs) // saddleless mount, must not panic
	UnsaddleAtEOT(gs) // second call, must not panic (idempotent)
}

func TestUnsaddleAtEOT_PhasesPipelineClears(t *testing.T) {
	gs := newSaddleGame(t)
	mount := mountPerm(gs, 0, "Mount", 2)
	c := creaturePerm(gs, 0, "Saddler", 3)
	if !SaddleMount(gs, 0, mount, []*Permanent{c}) {
		t.Fatal("setup: saddle should succeed")
	}
	// The existing cleanup pass at end of turn (phases.go
	// ScanExpiredDurations on the cleanup step) should also drop the
	// designation, demonstrating that SaddleMount writes the same
	// state shape that the inline cleanup expects.
	ScanExpiredDurations(gs, "ending", "cleanup")
	if PermIsSaddled(mount) {
		t.Fatal("phases pipeline cleanup must drop saddled designation")
	}
}

// ---------------------------------------------------------------------------
// (e) PermIsSaddled query
// ---------------------------------------------------------------------------

func TestPermIsSaddled_DefaultsFalse(t *testing.T) {
	gs := newSaddleGame(t)
	mount := mountPerm(gs, 0, "Mount", 2)
	if PermIsSaddled(mount) {
		t.Fatal("a fresh mount must not be saddled")
	}
	if PermIsSaddled(nil) {
		t.Fatal("PermIsSaddled(nil) should be false")
	}
}

func TestPermIsSaddled_TrueOnlyAfterSuccessfulSaddle(t *testing.T) {
	gs := newSaddleGame(t)
	mount := mountPerm(gs, 0, "Mount", 5)
	c := creaturePerm(gs, 0, "Lone Bear", 2) // power 2 vs N=5
	SaddleMount(gs, 0, mount, []*Permanent{c})
	if PermIsSaddled(mount) {
		t.Fatal("PermIsSaddled must remain false after a failed saddle")
	}
	bigger := creaturePerm(gs, 0, "Big Bear", 5)
	SaddleMount(gs, 0, mount, []*Permanent{bigger})
	if !PermIsSaddled(mount) {
		t.Fatal("PermIsSaddled should flip to true after a successful saddle")
	}
}

// ---------------------------------------------------------------------------
// Saddle a mount missing the keyword (e.g. trying to saddle a vanilla)
// ---------------------------------------------------------------------------

func TestSaddleMount_NoKeywordRefused(t *testing.T) {
	gs := newSaddleGame(t)
	notAMount := creaturePerm(gs, 0, "Grizzly Bears", 2)
	tapper := creaturePerm(gs, 0, "Hill Giant", 3)

	if SaddleMount(gs, 0, notAMount, []*Permanent{tapper}) {
		t.Fatal("SaddleMount on a non-mount target must fail")
	}
	if tapper.Tapped {
		t.Fatal("failed saddle must not tap the chosen creature")
	}
}

// ---------------------------------------------------------------------------
// Nil safety
// ---------------------------------------------------------------------------

func TestSaddleMount_NilSafe(t *testing.T) {
	gs := newSaddleGame(t)
	mount := mountPerm(gs, 0, "Mount", 1)
	if SaddleMount(nil, 0, mount, nil) {
		t.Fatal("nil game should fail")
	}
	if SaddleMount(gs, 0, nil, nil) {
		t.Fatal("nil mount should fail")
	}
	if SaddleMount(gs, -1, mount, nil) {
		t.Fatal("invalid seat should fail")
	}
	if SaddleMount(gs, 0, mount, []*Permanent{nil}) {
		t.Fatal("nil tapper in list should fail")
	}
}
