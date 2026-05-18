package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Battle tests — CR §310 / §704.5v
// ---------------------------------------------------------------------------

func newBattleGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(41))
	gs := NewGameState(2, rng, nil)
	gs.Active = 0
	gs.Step = "combat_damage"
	return gs
}

// newBattleCard models a printed battle. BaseToughness encodes the
// printed defense N — the stack.go ETB path stamps perm.Counters
// ["defense"] = card.BaseToughness for any cardHasType(card,"battle").
func newBattleCard(name string, owner, defense int) *Card {
	return &Card{
		Name:          name,
		Owner:         owner,
		Types:         []string{"battle", "siege"},
		BaseToughness: defense,
		AST:           &gameast.CardAST{Name: name},
	}
}

func newAttackerCard(name string, owner, power int) *Card {
	return &Card{
		Name:          name,
		Owner:         owner,
		Types:         []string{"creature"},
		BasePower:     power,
		BaseToughness: 3,
		AST:           &gameast.CardAST{Name: name},
	}
}

// castBattleToBattlefield drives the canonical ETB path so the
// defense-counter initializer in stack.go runs. The card is
// dropped into the seat's hand, then resolved as if cast.
func castBattleToBattlefield(t *testing.T, gs *GameState, seat int, card *Card) *Permanent {
	t.Helper()
	gs.Seats[seat].Hand = append(gs.Seats[seat].Hand, card)
	item := &StackItem{
		Card:       card,
		Controller: seat,
		CastZone:   ZoneHand,
	}
	resolvePermanentSpellETB(gs, item)
	// Find the permanent we just created.
	for _, p := range gs.Seats[seat].Battlefield {
		if p.Card == card {
			return p
		}
	}
	t.Fatalf("battle %q did not appear on seat %d's battlefield", card.DisplayName(), seat)
	return nil
}

// putAttacker drops a creature on seat's battlefield without going
// through the cast pipeline.
func putBattleAttacker(gs *GameState, seat int, card *Card) *Permanent {
	perm := &Permanent{
		Card:       card,
		Controller: seat,
		Owner:      seat,
		Timestamp:  gs.NextTimestamp(),
	}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, perm)
	return perm
}

// installBecomesDefeatedHook installs a TriggerHook that captures
// "becomes_defeated" events. Returns the captured-event pointer and
// a teardown closure.
func installBecomesDefeatedHook(t *testing.T) (*[]capturedTrigger, func()) {
	t.Helper()
	return installCapturingTriggerHook(t)
}

// ---------------------------------------------------------------------------
// IsBattle helper
// ---------------------------------------------------------------------------

func TestIsBattle_Detects(t *testing.T) {
	gs := newBattleGame(t)
	perm := putBattleAttacker(gs, 0, newBattleCard("Siege X", 0, 4))
	if !IsBattle(perm) {
		t.Fatal("IsBattle should be true for a battle permanent")
	}
	plain := putBattleAttacker(gs, 0, newAttackerCard("Bear", 0, 2))
	if IsBattle(plain) {
		t.Fatal("IsBattle should be false for a non-battle permanent")
	}
	if IsBattle(nil) {
		t.Fatal("IsBattle(nil) should be false")
	}
}

// ---------------------------------------------------------------------------
// (a) Battle ETBs with declared defense counter N
// ---------------------------------------------------------------------------

func TestBattle_ETBStampsDefenseCounters(t *testing.T) {
	gs := newBattleGame(t)
	card := newBattleCard("Invasion of Tarkir", 0, 4)
	perm := castBattleToBattlefield(t, gs, 0, card)
	if got := BattleDefenseCounters(perm); got != 4 {
		t.Fatalf("BattleDefenseCounters at ETB = %d, want 4 (printed BaseToughness)", got)
	}
}

// ---------------------------------------------------------------------------
// (e) BattleDefenseCounters query accurate
// ---------------------------------------------------------------------------

func TestBattleDefenseCounters_Queries(t *testing.T) {
	gs := newBattleGame(t)
	perm := putBattleAttacker(gs, 0, newBattleCard("Siege", 0, 3))
	// No counters yet (ETB skipped — putBattleAttacker bypasses ETB).
	if got := BattleDefenseCounters(perm); got != 0 {
		t.Fatalf("BattleDefenseCounters with no counters = %d, want 0", got)
	}
	AddDefenseCounters(gs, perm, 5)
	if got := BattleDefenseCounters(perm); got != 5 {
		t.Fatalf("BattleDefenseCounters after AddDefenseCounters(5) = %d, want 5", got)
	}
	RemoveDefenseCounters(gs, perm, 2)
	if got := BattleDefenseCounters(perm); got != 3 {
		t.Fatalf("BattleDefenseCounters after Remove(2) = %d, want 3", got)
	}
}

func TestAddDefenseCounters_NoOpForNonPositive(t *testing.T) {
	gs := newBattleGame(t)
	perm := putBattleAttacker(gs, 0, newBattleCard("Siege", 0, 3))
	AddDefenseCounters(gs, perm, 0)
	AddDefenseCounters(gs, perm, -2)
	if got := BattleDefenseCounters(perm); got != 0 {
		t.Fatalf("AddDefenseCounters with non-positive n should be no-op, got %d", got)
	}
}

func TestAddDefenseCounters_ClearsDefeatedLatch(t *testing.T) {
	gs := newBattleGame(t)
	perm := putBattleAttacker(gs, 0, newBattleCard("Siege", 0, 1))
	AddDefenseCounters(gs, perm, 1)
	RemoveDefenseCounters(gs, perm, 1)
	if !IsBattleDefeated(perm) {
		t.Fatal("setup: battle should be latched defeated after removing last counter")
	}
	AddDefenseCounters(gs, perm, 3)
	if IsBattleDefeated(perm) {
		t.Fatal("AddDefenseCounters should clear the battle_defeated latch when counters return to positive")
	}
}

// ---------------------------------------------------------------------------
// (c) Damage to battle removes counters  +
// (d) at 0 counters FireBattleZeroDefense fires
// ---------------------------------------------------------------------------

func TestRemoveDefenseCounters_DamageRemovesAndFiresAtZero(t *testing.T) {
	gs := newBattleGame(t)
	captured, restore := installBecomesDefeatedHook(t)
	defer restore()

	perm := putBattleAttacker(gs, 0, newBattleCard("Siege", 0, 3))
	AddDefenseCounters(gs, perm, 3)

	if removed := RemoveDefenseCounters(gs, perm, 2); removed != 2 {
		t.Fatalf("Remove(2) reported %d, want 2", removed)
	}
	if got := BattleDefenseCounters(perm); got != 1 {
		t.Fatalf("after Remove(2): %d counters, want 1", got)
	}
	if IsBattleDefeated(perm) {
		t.Fatal("battle should NOT be defeated while counters > 0")
	}
	// Removing more than remaining clamps at zero.
	if removed := RemoveDefenseCounters(gs, perm, 5); removed != 1 {
		t.Fatalf("Remove(5) when 1 left reported %d, want 1 (clamp)", removed)
	}
	if got := BattleDefenseCounters(perm); got != 0 {
		t.Fatalf("after Remove clamp: %d counters, want 0", got)
	}
	if !IsBattleDefeated(perm) {
		t.Fatal("battle should be latched defeated at 0 counters")
	}
	// becomes_defeated fired exactly once.
	defeats := 0
	for _, c := range *captured {
		if c.event == "becomes_defeated" {
			defeats++
			if c.ctx["source"] != perm {
				t.Fatal("becomes_defeated ctx.source mismatch")
			}
		}
	}
	if defeats != 1 {
		t.Fatalf("becomes_defeated fired %d times, want 1", defeats)
	}
}

func TestFireBattleZeroDefense_IdempotentLatch(t *testing.T) {
	gs := newBattleGame(t)
	captured, restore := installBecomesDefeatedHook(t)
	defer restore()

	perm := putBattleAttacker(gs, 0, newBattleCard("Siege", 0, 1))
	AddDefenseCounters(gs, perm, 1)
	RemoveDefenseCounters(gs, perm, 1) // 1 defeat fire
	FireBattleZeroDefense(gs, perm)    // should latch-skip
	FireBattleZeroDefense(gs, perm)    // should latch-skip

	defeats := 0
	for _, c := range *captured {
		if c.event == "becomes_defeated" {
			defeats++
		}
	}
	if defeats != 1 {
		t.Fatalf("becomes_defeated fired %d times, want 1 (latch should suppress repeats)", defeats)
	}
}

func TestFireBattleZeroDefense_NonBattleNoOp(t *testing.T) {
	gs := newBattleGame(t)
	captured, restore := installBecomesDefeatedHook(t)
	defer restore()

	plain := putBattleAttacker(gs, 0, newAttackerCard("Bear", 0, 2))
	FireBattleZeroDefense(gs, plain)
	for _, c := range *captured {
		if c.event == "becomes_defeated" {
			t.Fatal("becomes_defeated must not fire on a non-battle permanent")
		}
	}
}

// ---------------------------------------------------------------------------
// (b) Attack-battle declaration legal — combat-damage routing
// ---------------------------------------------------------------------------

func TestAttackerDefenderBattle_FlagPlumbing(t *testing.T) {
	gs := newBattleGame(t)
	attacker := putBattleAttacker(gs, 0, newAttackerCard("Soldier", 0, 2))
	battle := putBattleAttacker(gs, 1, newBattleCard("Siege", 1, 4))
	AddDefenseCounters(gs, battle, 4)

	SetAttackerDefenderBattle(attacker, battle)
	ts, ok := AttackerDefenderBattle(attacker)
	if !ok {
		t.Fatal("AttackerDefenderBattle should report a battle target after SetAttackerDefenderBattle")
	}
	if ts != battle.Timestamp {
		t.Fatalf("battle timestamp = %d, want %d", ts, battle.Timestamp)
	}
	if found, ok := LookupBattleByTimestamp(gs, ts); !ok || found != battle {
		t.Fatal("LookupBattleByTimestamp should resolve back to the original battle")
	}
	// Seat-style defender should be cleared so the damage step
	// routes to the battle path, not the player path.
	if _, ok := AttackerDefender(attacker); ok {
		t.Fatal("seat-style defender flag should be cleared when retargeting to a battle")
	}
}

func TestSetAttackerDefenderBattle_RejectsNonBattle(t *testing.T) {
	gs := newBattleGame(t)
	attacker := putBattleAttacker(gs, 0, newAttackerCard("Soldier", 0, 2))
	notABattle := putBattleAttacker(gs, 1, newAttackerCard("Wall", 1, 0))
	SetAttackerDefenderBattle(attacker, notABattle)
	if _, ok := AttackerDefenderBattle(attacker); ok {
		t.Fatal("SetAttackerDefenderBattle must reject non-battle targets")
	}
}

func TestCombatDamage_RoutesToBattleViaDealCombatDamageStep(t *testing.T) {
	gs := newBattleGame(t)
	captured, restore := installBecomesDefeatedHook(t)
	defer restore()

	// Seat-0 attacker, seat-1 battle (printed defense 4).
	attacker := putBattleAttacker(gs, 0, newAttackerCard("Big Hitter", 0, 5))
	battle := putBattleAttacker(gs, 1, newBattleCard("Siege", 1, 4))
	AddDefenseCounters(gs, battle, 4)

	SetAttackerDefenderBattle(attacker, battle)
	// Mark as attacking so DealCombatDamageStep iterates it.
	if attacker.Flags == nil {
		attacker.Flags = map[string]int{}
	}
	attacker.Flags[flagAttacking] = 1
	attacker.Flags[flagDeclaredAttacker] = 1

	// No blockers → damage routes through ApplyCombatDamageToBattle.
	DealCombatDamageStep(gs, []*Permanent{attacker}, nil, false)

	if got := BattleDefenseCounters(battle); got != 0 {
		t.Fatalf("battle counters after 5 damage to a 4-defense battle = %d, want 0 (clamp)", got)
	}
	if !IsBattleDefeated(battle) {
		t.Fatal("battle should be latched defeated after counters reach 0")
	}
	// Defender (seat 1) life total must NOT have been touched — the
	// damage went to the battle, not the player.
	if gs.Seats[1].Life != gs.Seats[1].StartingLife {
		t.Fatalf("defending seat life dropped from %d to %d — battle attack should not damage the player",
			gs.Seats[1].StartingLife, gs.Seats[1].Life)
	}
	// becomes_defeated fired once.
	defeats := 0
	for _, c := range *captured {
		if c.event == "becomes_defeated" {
			defeats++
		}
	}
	if defeats != 1 {
		t.Fatalf("becomes_defeated fired %d, want 1", defeats)
	}
}

func TestCombatDamage_PartialDamageRemovesPartialCounters(t *testing.T) {
	gs := newBattleGame(t)
	attacker := putBattleAttacker(gs, 0, newAttackerCard("Bear", 0, 2))
	battle := putBattleAttacker(gs, 1, newBattleCard("Siege", 1, 5))
	AddDefenseCounters(gs, battle, 5)
	SetAttackerDefenderBattle(attacker, battle)
	if attacker.Flags == nil {
		attacker.Flags = map[string]int{}
	}
	attacker.Flags[flagAttacking] = 1
	attacker.Flags[flagDeclaredAttacker] = 1

	DealCombatDamageStep(gs, []*Permanent{attacker}, nil, false)

	if got := BattleDefenseCounters(battle); got != 3 {
		t.Fatalf("battle counters after 2 damage = %d, want 3", got)
	}
	if IsBattleDefeated(battle) {
		t.Fatal("battle must not be defeated while counters > 0")
	}
}

// ---------------------------------------------------------------------------
// LookupBattleByTimestamp safety
// ---------------------------------------------------------------------------

func TestLookupBattleByTimestamp_NilSafe(t *testing.T) {
	if _, ok := LookupBattleByTimestamp(nil, 0); ok {
		t.Fatal("LookupBattleByTimestamp(nil) should return false")
	}
	gs := newBattleGame(t)
	if _, ok := LookupBattleByTimestamp(gs, -1); ok {
		t.Fatal("LookupBattleByTimestamp(negative ts) should return false")
	}
	if _, ok := LookupBattleByTimestamp(gs, 99999); ok {
		t.Fatal("LookupBattleByTimestamp(unknown ts) should return false")
	}
}

func TestApplyCombatDamageToBattle_NonBattleNoOp(t *testing.T) {
	gs := newBattleGame(t)
	attacker := putBattleAttacker(gs, 0, newAttackerCard("Bear", 0, 2))
	plain := putBattleAttacker(gs, 1, newAttackerCard("Wall", 1, 0))
	startLife := gs.Seats[1].Life
	ApplyCombatDamageToBattle(gs, attacker, 3, plain) // non-battle → no-op
	if startLife != gs.Seats[1].Life {
		t.Fatal("ApplyCombatDamageToBattle on a non-battle should not damage anyone")
	}
}
