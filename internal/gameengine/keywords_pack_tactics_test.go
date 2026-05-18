package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// Round-28 tests for Pack Tactics (CR §702.149). Mirrors Battalion's
// test fixtures, swapping the count threshold for the power threshold.

// ---------------------------------------------------------------------------
// Test card builders
// ---------------------------------------------------------------------------

func pt_makeGame(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(2, rand.New(rand.NewSource(28)), nil)
}

func pt_makePackTacticsCreature(seat int, name string, power, toughness int) *Permanent {
	c := &Card{
		Name:          name,
		Owner:         seat,
		Types:         []string{"creature"},
		BasePower:     power,
		BaseToughness: toughness,
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "pack tactics", Raw: "pack tactics"},
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

func pt_makePlainCreature(seat int, name string, power, toughness int) *Permanent {
	c := &Card{
		Name:          name,
		Owner:         seat,
		Types:         []string{"creature"},
		BasePower:     power,
		BaseToughness: toughness,
		AST:           &gameast.CardAST{Name: name},
	}
	return &Permanent{
		Card:       c,
		Controller: seat,
		Owner:      seat,
		Flags:      map[string]int{},
	}
}

// pt_setAttacking marks `p` as attacking the given defender seat,
// mirroring what declareAttackers does.
func pt_setAttacking(p *Permanent, defenderSeat int) {
	if p.Flags == nil {
		p.Flags = map[string]int{}
	}
	p.Flags[flagAttacking] = 1
	SetAttackerDefender(p, defenderSeat)
}

// pt_countTriggerEvents returns the number of pack_tactics_trigger
// events present in the log.
func pt_countTriggerEvents(gs *GameState) int {
	n := 0
	for _, ev := range gs.EventLog {
		if ev.Kind == "pack_tactics_trigger" {
			n++
		}
	}
	return n
}

// ---------------------------------------------------------------------------
// HasPackTactics detector smoke tests
// ---------------------------------------------------------------------------

func TestHasPackTactics_PositiveAndNegative(t *testing.T) {
	pkt := pt_makePackTacticsCreature(0, "Lorehold Apprentice", 2, 2)
	if !HasPackTactics(pkt.Card) {
		t.Fatal("HasPackTactics should detect the keyword")
	}
	plain := pt_makePlainCreature(0, "Goblin", 1, 1)
	if HasPackTactics(plain.Card) {
		t.Fatal("HasPackTactics should be false for plain creature")
	}
	if HasPackTactics(nil) {
		t.Fatal("HasPackTactics(nil) should be false")
	}
}

// ---------------------------------------------------------------------------
// (a) Attackers with total power 6+ triggers.
// ---------------------------------------------------------------------------

func TestPackTactics_TotalPower6Triggers(t *testing.T) {
	gs := pt_makeGame(t)
	pkt := pt_makePackTacticsCreature(0, "Lorehold Pack Leader", 3, 3)
	mate1 := pt_makePlainCreature(0, "Wolf 1", 2, 2)
	mate2 := pt_makePlainCreature(0, "Wolf 2", 1, 1)
	// Total power = 3 + 2 + 1 = 6 — exactly at threshold.
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, pkt, mate1, mate2)
	for _, p := range gs.Seats[0].Battlefield {
		pt_setAttacking(p, 1)
	}

	if !FirePackTacticsTriggers(gs, pkt, 1) {
		t.Fatal("expected pack tactics to fire at total power 6")
	}
	if pt_countTriggerEvents(gs) != 1 {
		t.Fatalf("expected 1 pack_tactics_trigger event, got %d", pt_countTriggerEvents(gs))
	}
}

// ---------------------------------------------------------------------------
// (b) Total power 5 does NOT trigger.
// ---------------------------------------------------------------------------

func TestPackTactics_TotalPower5DoesNotTrigger(t *testing.T) {
	gs := pt_makeGame(t)
	pkt := pt_makePackTacticsCreature(0, "Lorehold Apprentice", 2, 2)
	mate1 := pt_makePlainCreature(0, "Wolf 1", 2, 2)
	mate2 := pt_makePlainCreature(0, "Wolf 2", 1, 1)
	// Total power = 2 + 2 + 1 = 5 — just below threshold.
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, pkt, mate1, mate2)
	for _, p := range gs.Seats[0].Battlefield {
		pt_setAttacking(p, 1)
	}

	if FirePackTacticsTriggers(gs, pkt, 1) {
		t.Fatal("pack tactics should NOT fire at total power 5")
	}
	if pt_countTriggerEvents(gs) != 0 {
		t.Fatalf("expected 0 trigger events, got %d", pt_countTriggerEvents(gs))
	}
}

// ---------------------------------------------------------------------------
// (c) Opponent's attackers don't count.
// ---------------------------------------------------------------------------

func TestPackTactics_OpponentAttackersExcluded(t *testing.T) {
	gs := pt_makeGame(t)
	// Seat 0: pack tactics source + small mate (total power 4).
	pkt := pt_makePackTacticsCreature(0, "Pack Leader", 3, 3)
	mate := pt_makePlainCreature(0, "Cub", 1, 1)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, pkt, mate)
	for _, p := range gs.Seats[0].Battlefield {
		pt_setAttacking(p, 1)
	}

	// Seat 1: a giant attacking seat 0 (extra-combat / control swap
	// scenario). Their power must NOT count toward seat 0's total.
	enemy := pt_makePlainCreature(1, "Enemy Giant", 8, 8)
	gs.Seats[1].Battlefield = append(gs.Seats[1].Battlefield, enemy)
	pt_setAttacking(enemy, 0)

	if FirePackTacticsTriggers(gs, pkt, 1) {
		t.Fatal("opponent's attacking creature must not contribute to controller's total power")
	}
}

// ---------------------------------------------------------------------------
// (d) Buffs (Glorious Anthem +1/+1) count toward total power.
// ---------------------------------------------------------------------------

func TestPackTactics_BuffsCountTowardTotal(t *testing.T) {
	gs := pt_makeGame(t)
	pkt := pt_makePackTacticsCreature(0, "Lorehold Apprentice", 2, 2)
	mate := pt_makePlainCreature(0, "Squire", 2, 2)
	// Base total = 4. Need +2 from buffs to clear threshold of 6.
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, pkt, mate)
	for _, p := range gs.Seats[0].Battlefield {
		pt_setAttacking(p, 1)
	}

	// Confirm below-threshold without buffs.
	if FirePackTacticsTriggers(gs, pkt, 1) {
		t.Fatal("pre-buff total power 4 should not fire")
	}

	// Apply +1/+1 to each via Modifications (Glorious-Anthem-equivalent).
	for _, p := range gs.Seats[0].Battlefield {
		p.Modifications = append(p.Modifications, Modification{
			Power:     1,
			Toughness: 1,
			Duration:  "permanent",
			Timestamp: gs.NextTimestamp(),
		})
	}
	gs.InvalidateCharacteristicsCache()
	// Total is now 3+3 = 6.

	if !FirePackTacticsTriggers(gs, pkt, 1) {
		t.Fatalf("post-buff total power should fire; got powers pkt=%d mate=%d",
			pkt.Power(), mate.Power())
	}
}

// ---------------------------------------------------------------------------
// (e) Source itself must be attacking.
// ---------------------------------------------------------------------------

func TestPackTactics_SourceMustBeAttacking(t *testing.T) {
	gs := pt_makeGame(t)
	pkt := pt_makePackTacticsCreature(0, "Pack Leader", 3, 3)
	mate1 := pt_makePlainCreature(0, "Wolf 1", 3, 3)
	mate2 := pt_makePlainCreature(0, "Wolf 2", 2, 2)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, pkt, mate1, mate2)
	// Only mates attack — total attacker power is 5 (mate1+mate2). pkt
	// is NOT attacking and so doesn't contribute its own 3.
	pt_setAttacking(mate1, 1)
	pt_setAttacking(mate2, 1)

	if FirePackTacticsTriggers(gs, pkt, 1) {
		t.Fatal("pack tactics must not fire when its source isn't attacking")
	}

	// Now have the source attack too — total becomes 8, should fire.
	pt_setAttacking(pkt, 1)
	if !FirePackTacticsTriggers(gs, pkt, 1) {
		t.Fatal("with source attacking, total power 8 should fire")
	}
}

// ---------------------------------------------------------------------------
// Bonus: batch entry point fires one trigger per pack-tactics source.
// ---------------------------------------------------------------------------

func TestPackTactics_BatchHookFiresPerSource(t *testing.T) {
	gs := pt_makeGame(t)
	a := pt_makePackTacticsCreature(0, "Pack A", 3, 3)
	b := pt_makePackTacticsCreature(0, "Pack B", 3, 3)
	c := pt_makePlainCreature(0, "Filler", 1, 1)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, a, b, c)
	for _, p := range gs.Seats[0].Battlefield {
		pt_setAttacking(p, 1)
	}
	// Total power = 3+3+1 = 7. Two pack-tactics sources both attacking.

	FirePackTacticsForAttackers(gs, gs.Seats[0].Battlefield)

	if got := pt_countTriggerEvents(gs); got != 2 {
		t.Fatalf("batch hook should fire one trigger per pack-tactics source; got %d", got)
	}
}

// Bonus: nil-safety.
func TestPackTactics_NilSafe(t *testing.T) {
	if FirePackTacticsTriggers(nil, nil, 0) {
		t.Fatal("nil game/perm should be safe no-op")
	}
	gs := pt_makeGame(t)
	if FirePackTacticsTriggers(gs, nil, 0) {
		t.Fatal("nil attacker should be safe no-op")
	}
	FirePackTacticsForAttackers(nil, nil)
	FirePackTacticsForAttackers(gs, nil)
}
