package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Test helpers — Round-33 gated-state batch
// ---------------------------------------------------------------------------

func newGatedBatchGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(331))
	return NewGameState(2, rng, nil)
}

// addCreatureWithPower drops a vanilla creature with explicit power on
// `seat`'s battlefield. Used for Coven / Ferocious / Raid setup.
func addCreatureWithPower(gs *GameState, seat int, name string, power int) *Permanent {
	c := &Card{
		Name: name, Owner: seat,
		BasePower: power, BaseToughness: power,
		Types: []string{"creature"},
		AST:   &gameast.CardAST{Name: name},
	}
	p := &Permanent{
		Card: c, Controller: seat, Owner: seat,
		Timestamp: gs.NextTimestamp(),
		Counters:  map[string]int{}, Flags: map[string]int{},
	}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, p)
	return p
}

// addGraveCard pushes a card with the given types onto `seat`'s
// graveyard. Used for Delirium setup.
func addGraveCard(gs *GameState, seat int, name string, types ...string) {
	gs.Seats[seat].Graveyard = append(gs.Seats[seat].Graveyard, &Card{
		Name:  name,
		Owner: seat,
		Types: append([]string(nil), types...),
	})
}

// keywordCard builds a card with a single AST keyword named `name`.
// Used to exercise the AST-keyword detection path.
func keywordCard(name, kw string) *Card {
	return &Card{
		Name: name,
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{&gameast.Keyword{Name: kw}},
		},
	}
}

// oracleCard builds a card whose AST has only a Static with raw text,
// for oracle-text detector coverage.
func oracleCard(name, raw string) *Card {
	return &Card{
		Name: name,
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{&gameast.Static{Raw: raw}},
		},
	}
}

// permWithKeyword builds a permanent on seat 0 carrying keyword `kw`,
// suitable for the rider integration tests.
func permWithKeyword(gs *GameState, seat int, name, kw string) *Permanent {
	c := keywordCard(name, kw)
	p := &Permanent{
		Card: c, Controller: seat, Owner: seat,
		Timestamp: gs.NextTimestamp(),
		Counters:  map[string]int{}, Flags: map[string]int{},
	}
	return p
}

// ===========================================================================
// Delirium (§702.151)
// ===========================================================================

func TestHasDelirium_Detects(t *testing.T) {
	if !HasDelirium(keywordCard("Mindwrack Demon", "delirium")) {
		t.Error("HasDelirium should detect AST keyword")
	}
	if !HasDelirium(oracleCard("Vessel of Nascency", "Delirium — Sacrifice this enchantment...")) {
		t.Error("HasDelirium should detect oracle text")
	}
	if HasDelirium(nil) {
		t.Error("HasDelirium(nil) should be false")
	}
	if HasDelirium(&Card{AST: &gameast.CardAST{}}) {
		t.Error("HasDelirium on empty AST should be false")
	}
}

func TestDeliriumActive(t *testing.T) {
	gs := newGatedBatchGame(t)
	// 0 types — inactive.
	if DeliriumActive(gs, 0) {
		t.Fatal("DeliriumActive with empty graveyard should be false")
	}
	// 3 distinct types — still inactive.
	addGraveCard(gs, 0, "Inst", "instant")
	addGraveCard(gs, 0, "Sorc", "sorcery")
	addGraveCard(gs, 0, "Crea", "creature")
	if DeliriumActive(gs, 0) {
		t.Fatal("DeliriumActive should be false at 3 types")
	}
	// 4 distinct types — active.
	addGraveCard(gs, 0, "Land", "land")
	if !DeliriumActive(gs, 0) {
		t.Fatal("DeliriumActive should flip true at 4 distinct types")
	}
	// 5+ distinct types — still active.
	addGraveCard(gs, 0, "Art", "artifact")
	if !DeliriumActive(gs, 0) {
		t.Fatal("DeliriumActive should stay true past 4 types")
	}
}

func TestDeliriumActive_FlipsAndPerSeatIsolation(t *testing.T) {
	gs := newGatedBatchGame(t)
	// Seat 0: 4 types — active.
	addGraveCard(gs, 0, "I", "instant")
	addGraveCard(gs, 0, "S", "sorcery")
	addGraveCard(gs, 0, "C", "creature")
	addGraveCard(gs, 0, "L", "land")
	if !DeliriumActive(gs, 0) {
		t.Fatal("seat 0 should be active at 4 types")
	}
	// Seat 1 with same graveyard contents would have to be set up
	// independently — confirm starting empty seat 1 is inactive.
	if DeliriumActive(gs, 1) {
		t.Fatal("seat 1's delirium must NOT see seat 0's graveyard")
	}
	// Wipe seat 0's graveyard via Tormod's-Crypt-style exile.
	gs.Seats[0].Graveyard = nil
	if DeliriumActive(gs, 0) {
		t.Fatal("DeliriumActive should flip back to false after graveyard wipe")
	}
	// Nil-safe.
	if DeliriumActive(nil, 0) {
		t.Fatal("DeliriumActive(nil, ...) should be false")
	}
	if DeliriumActive(gs, -1) || DeliriumActive(gs, 99) {
		t.Fatal("DeliriumActive with invalid seat should be false")
	}
}

// ===========================================================================
// Coven (§702.152)
// ===========================================================================

func TestHasCoven_Detects(t *testing.T) {
	if !HasCoven(keywordCard("Coven Beast", "coven")) {
		t.Error("HasCoven should detect AST keyword")
	}
	if !HasCoven(oracleCard("Coveted Lab", "Coven — Whenever you cast...")) {
		t.Error("HasCoven should detect oracle text")
	}
	if HasCoven(nil) {
		t.Error("HasCoven(nil) should be false")
	}
}

func TestCovenActive_RequiresThreeDistinctPowers(t *testing.T) {
	gs := newGatedBatchGame(t)
	// 0 creatures — inactive.
	if CovenActive(gs, 0) {
		t.Fatal("empty board: should be inactive")
	}
	// Three 2/2s — three creatures, but only ONE distinct power.
	addCreatureWithPower(gs, 0, "A", 2)
	addCreatureWithPower(gs, 0, "B", 2)
	addCreatureWithPower(gs, 0, "C", 2)
	if CovenActive(gs, 0) {
		t.Fatal("three creatures with identical powers should NOT enable coven")
	}
	// Bump one to 4-power — now powers are {2, 2, 4} → 2 distinct.
	addCreatureWithPower(gs, 0, "D", 4)
	if CovenActive(gs, 0) {
		t.Fatal("powers {2,2,2,4} = 2 distinct values, still not coven")
	}
	// Add a 3-power → {2,2,2,4,3} → 3 distinct, 5 creatures — active.
	addCreatureWithPower(gs, 0, "E", 3)
	if !CovenActive(gs, 0) {
		t.Fatal("powers {2,2,2,4,3} = 3 distinct values should enable coven")
	}
}

func TestCovenActive_FlipsWithBoardState(t *testing.T) {
	gs := newGatedBatchGame(t)
	addCreatureWithPower(gs, 0, "1/1", 1)
	addCreatureWithPower(gs, 0, "2/2", 2)
	addCreatureWithPower(gs, 0, "3/3", 3)
	if !CovenActive(gs, 0) {
		t.Fatal("setup: {1,2,3} should be coven-active")
	}
	// Remove the 3/3 → only {1,2} powers remain, coven flips off.
	gs.Seats[0].Battlefield = gs.Seats[0].Battlefield[:2]
	if CovenActive(gs, 0) {
		t.Fatal("coven should flip OFF when third distinct power leaves")
	}
}

func TestCovenActive_PerSeatIsolation(t *testing.T) {
	gs := newGatedBatchGame(t)
	addCreatureWithPower(gs, 0, "1", 1)
	addCreatureWithPower(gs, 0, "2", 2)
	addCreatureWithPower(gs, 0, "3", 3)
	if !CovenActive(gs, 0) {
		t.Fatal("setup: seat 0 should be active")
	}
	if CovenActive(gs, 1) {
		t.Fatal("seat 1's coven must NOT see seat 0's creatures")
	}
}

func TestCovenActive_NonCreatureAndPhasedSkipped(t *testing.T) {
	gs := newGatedBatchGame(t)
	addCreatureWithPower(gs, 0, "1", 1)
	addCreatureWithPower(gs, 0, "2", 2)
	// A 3-power "creature" that's phased out — must not count.
	phased := addCreatureWithPower(gs, 0, "Phased", 3)
	phased.PhasedOut = true
	if CovenActive(gs, 0) {
		t.Fatal("phased-out creature must not contribute to coven")
	}
	// Add a real 3-power creature — now active.
	addCreatureWithPower(gs, 0, "3", 3)
	if !CovenActive(gs, 0) {
		t.Fatal("real 3-power creature should enable coven")
	}
}

// ===========================================================================
// Ferocious (§702.135)
// ===========================================================================

func TestHasFerocious_Detects(t *testing.T) {
	if !HasFerocious(keywordCard("Crater's Claws", "ferocious")) {
		t.Error("HasFerocious should detect AST keyword")
	}
	if !HasFerocious(oracleCard("Boon Satyr", "Ferocious — If you control a creature with power 4...")) {
		t.Error("HasFerocious should detect oracle text")
	}
	if HasFerocious(nil) {
		t.Error("HasFerocious(nil) should be false")
	}
}

func TestFerociousActive_RequiresPowerFourPlus(t *testing.T) {
	gs := newGatedBatchGame(t)
	if FerociousActive(gs, 0) {
		t.Fatal("empty board: should be inactive")
	}
	// 3-power only — inactive.
	addCreatureWithPower(gs, 0, "3/3", 3)
	if FerociousActive(gs, 0) {
		t.Fatal("3-power creature should NOT enable ferocious")
	}
	// Add a 4-power — active.
	addCreatureWithPower(gs, 0, "4/4", 4)
	if !FerociousActive(gs, 0) {
		t.Fatal("4-power creature should enable ferocious")
	}
	// Higher than 4 also active.
	gs2 := newGatedBatchGame(t)
	addCreatureWithPower(gs2, 0, "7/7", 7)
	if !FerociousActive(gs2, 0) {
		t.Fatal("7-power creature should enable ferocious")
	}
}

func TestFerociousActive_FlipsWithCreatureLeaving(t *testing.T) {
	gs := newGatedBatchGame(t)
	addCreatureWithPower(gs, 0, "2/2", 2)
	bigBoy := addCreatureWithPower(gs, 0, "5/5", 5)
	if !FerociousActive(gs, 0) {
		t.Fatal("setup: 5-power should enable ferocious")
	}
	// Remove the 5/5 — should flip off (the 2/2 doesn't count).
	gs.Seats[0].Battlefield = gs.Seats[0].Battlefield[:1]
	if FerociousActive(gs, 0) {
		t.Fatal("ferocious should flip OFF when the 5-power creature leaves")
	}
	_ = bigBoy
}

func TestFerociousActive_PerSeatIsolation(t *testing.T) {
	gs := newGatedBatchGame(t)
	addCreatureWithPower(gs, 1, "5/5", 5) // opponent has the big boy
	if FerociousActive(gs, 0) {
		t.Fatal("seat 0 must NOT inherit seat 1's ferocious")
	}
	if !FerociousActive(gs, 1) {
		t.Fatal("seat 1 should be active")
	}
}

// ===========================================================================
// Raid (§702.128)
// ===========================================================================

func TestHasRaid_Detects(t *testing.T) {
	if !HasRaid(keywordCard("Mardu Raider", "raid")) {
		t.Error("HasRaid should detect AST keyword")
	}
	if !HasRaid(oracleCard("Raid Bombardment", "Raid — If you attacked with a creature this turn...")) {
		t.Error("HasRaid should detect oracle text")
	}
}

func TestRaidActive_FlipsOnAttackedFlag(t *testing.T) {
	gs := newGatedBatchGame(t)
	if RaidActive(gs, 0) {
		t.Fatal("RaidActive at turn start should be false")
	}
	// Combat declares attackers — flag flips.
	gs.Seats[0].Turn.Attacked = true
	if !RaidActive(gs, 0) {
		t.Fatal("RaidActive should be true after Turn.Attacked is set")
	}
}

func TestRaidActive_PerSeatIsolation(t *testing.T) {
	gs := newGatedBatchGame(t)
	gs.Seats[1].Turn.Attacked = true
	if RaidActive(gs, 0) {
		t.Fatal("seat 0's raid must NOT see seat 1's attack")
	}
	if !RaidActive(gs, 1) {
		t.Fatal("seat 1 should be raid-active")
	}
}

func TestRaidActive_NilOrInvalid(t *testing.T) {
	if RaidActive(nil, 0) {
		t.Fatal("RaidActive(nil, ...) should be false")
	}
	gs := newGatedBatchGame(t)
	if RaidActive(gs, -1) || RaidActive(gs, 99) {
		t.Fatal("RaidActive with invalid seat should be false")
	}
}

// ===========================================================================
// Revolt (§702.146)
// ===========================================================================

func TestHasRevolt_Detects(t *testing.T) {
	if !HasRevolt(keywordCard("Fatal Push", "revolt")) {
		t.Error("HasRevolt should detect AST keyword")
	}
	if !HasRevolt(oracleCard("Renegade Map", "Revolt — If a permanent you controlled left...")) {
		t.Error("HasRevolt should detect oracle text")
	}
}

func TestRevoltActive_FlipsOnPermanentsLeft(t *testing.T) {
	gs := newGatedBatchGame(t)
	if RevoltActive(gs, 0) {
		t.Fatal("RevoltActive at turn start should be false")
	}
	gs.Seats[0].Turn.PermanentsLeft = 1
	if !RevoltActive(gs, 0) {
		t.Fatal("RevoltActive should be true after PermanentsLeft increments")
	}
	gs.Seats[0].Turn.PermanentsLeft = 5
	if !RevoltActive(gs, 0) {
		t.Fatal("RevoltActive should stay true at higher counts")
	}
}

func TestRevoltActive_PerSeatIsolation(t *testing.T) {
	gs := newGatedBatchGame(t)
	gs.Seats[1].Turn.PermanentsLeft = 3
	if RevoltActive(gs, 0) {
		t.Fatal("seat 0's revolt must NOT see seat 1's LTBs")
	}
	if !RevoltActive(gs, 1) {
		t.Fatal("seat 1 should be revolt-active")
	}
}

// ===========================================================================
// resolveGatedRider integration — all five fire end-to-end
// ===========================================================================

func TestResolveGatedRider_DeliriumFires(t *testing.T) {
	gs := newGatedBatchGame(t)
	// 4 distinct grave types → delirium active.
	addGraveCard(gs, 0, "I", "instant")
	addGraveCard(gs, 0, "S", "sorcery")
	addGraveCard(gs, 0, "C", "creature")
	addGraveCard(gs, 0, "L", "land")

	src := permWithKeyword(gs, 0, "Test Delirium Card", "delirium")
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, src)

	resolveGatedRider(gs, src)

	if !findEventKind(gs, "delirium_rider") {
		t.Fatal("expected delirium_rider event from resolveGatedRider")
	}
}

func TestResolveGatedRider_CovenFires(t *testing.T) {
	gs := newGatedBatchGame(t)
	addCreatureWithPower(gs, 0, "1", 1)
	addCreatureWithPower(gs, 0, "2", 2)
	addCreatureWithPower(gs, 0, "3", 3)
	src := permWithKeyword(gs, 0, "Test Coven", "coven")
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, src)

	resolveGatedRider(gs, src)

	if !findEventKind(gs, "coven_rider") {
		t.Fatal("expected coven_rider event from resolveGatedRider")
	}
}

func TestResolveGatedRider_FerociousFires(t *testing.T) {
	gs := newGatedBatchGame(t)
	addCreatureWithPower(gs, 0, "5/5", 5)
	src := permWithKeyword(gs, 0, "Test Ferocious", "ferocious")
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, src)

	resolveGatedRider(gs, src)

	if !findEventKind(gs, "ferocious_rider") {
		t.Fatal("expected ferocious_rider event")
	}
}

func TestResolveGatedRider_RaidFires(t *testing.T) {
	gs := newGatedBatchGame(t)
	gs.Seats[0].Turn.Attacked = true
	src := permWithKeyword(gs, 0, "Test Raid", "raid")
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, src)

	resolveGatedRider(gs, src)

	if !findEventKind(gs, "raid_rider") {
		t.Fatal("expected raid_rider event")
	}
}

func TestResolveGatedRider_RevoltFires(t *testing.T) {
	gs := newGatedBatchGame(t)
	gs.Seats[0].Turn.PermanentsLeft = 1
	src := permWithKeyword(gs, 0, "Test Revolt", "revolt")
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, src)

	resolveGatedRider(gs, src)

	if !findEventKind(gs, "revolt_rider") {
		t.Fatal("expected revolt_rider event")
	}
}

func TestResolveGatedRider_InactiveGateSilent(t *testing.T) {
	// None of the gates are active — no rider events should fire.
	gs := newGatedBatchGame(t)
	src := permWithKeyword(gs, 0, "Dormant", "delirium")
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, src)
	resolveGatedRider(gs, src)
	for _, ev := range gs.EventLog {
		switch ev.Kind {
		case "delirium_rider", "coven_rider", "ferocious_rider",
			"raid_rider", "revolt_rider":
			t.Fatalf("no rider should fire when gates are inactive; got %s", ev.Kind)
		}
	}
}

// findEventKind returns true iff `gs.EventLog` contains at least one
// event with Kind==kind. Tiny local helper to keep assertions tight.
func findEventKind(gs *GameState, kind string) bool {
	for _, ev := range gs.EventLog {
		if ev.Kind == kind {
			return true
		}
	}
	return false
}

// ===========================================================================
// ApplyXxxRider nil safety + non-keyword cards no-op
// ===========================================================================

func TestApplyRiders_NilAndNonKeywordSafe(t *testing.T) {
	for _, fn := range []func(*GameState, *Permanent) bool{
		ApplyDeliriumRider, ApplyCovenRider, ApplyFerociousRider,
		ApplyRaidRider, ApplyRevoltRider,
	} {
		if fn(nil, nil) {
			t.Error("rider(nil, nil) should be false")
		}
	}
	gs := newGatedBatchGame(t)
	src := &Permanent{Card: &Card{Name: "Plain", AST: &gameast.CardAST{}}, Controller: 0, Owner: 0}
	for _, fn := range []func(*GameState, *Permanent) bool{
		ApplyDeliriumRider, ApplyCovenRider, ApplyFerociousRider,
		ApplyRaidRider, ApplyRevoltRider,
	} {
		if fn(gs, src) {
			t.Error("rider on a card without the keyword should not fire")
		}
	}
}
