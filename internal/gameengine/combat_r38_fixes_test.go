package gameengine

// R38 combat-audit follow-up tests.
//
// One test cluster per fix landed on dev/combat-r38-fix:
//
//   1. Banding redistribution is invoked from DealCombatDamageStep so a
//      band of blockers takes its total damage redistributed across the
//      band (saving one) instead of evenly killing both.
//   2. attackerHasProtectionFrom now consults type-based protection
//      ("protection from creatures" / "from artifacts"), preventing the
//      creature/artifact blocker from blocking, and preventing damage
//      from the same source through applyCombatDamageToCreature.
//   3. DeclareAttackers enforces goad's must-attack-if-able + cannot-
//      attack-goader rules regardless of Hat compliance.

import (
	"math/rand"
	"testing"
)

type r38BlockerHat struct {
	GreedyHatStub
	blockers []*Permanent
}

func (h *r38BlockerHat) AssignBlockers(gs *GameState, seatIdx int, attackers []*Permanent) map[*Permanent][]*Permanent {
	out := map[*Permanent][]*Permanent{}
	for _, a := range attackers {
		out[a] = nil
	}
	if len(attackers) > 0 {
		out[attackers[0]] = h.blockers
	}
	return out
}

func TestCombat_R38_BandingRedistribution_BlockerBandSavesOne(t *testing.T) {
	gs := NewGameState(2, rand.New(rand.NewSource(1)), nil)
	gs.Active = 0

	atk := addCreature(gs, 0, "Ogre", 4, 4)
	blk1 := addCreature(gs, 1, "Banded A", 1, 2, "banding")
	blk2 := addCreature(gs, 1, "Banded B", 1, 2, "banding")

	gs.Seats[1].Hat = &r38BlockerHat{blockers: []*Permanent{blk1, blk2}}

	CombatPhase(gs)
	StateBasedActions(gs)

	if !alive(gs, atk) {
		t.Errorf("attacker should survive (took 1+1=2 of 4 toughness): marked=%d", atk.MarkedDamage)
	}
	survivors := 0
	for _, p := range []*Permanent{blk1, blk2} {
		if alive(gs, p) {
			survivors++
		}
	}
	if survivors != 1 {
		t.Errorf("banding redistribution should save exactly one of two banded 1/2s, got %d survivors", survivors)
	}
}

func TestCombat_R38_BandingRedistribution_SoloBanderSkipped(t *testing.T) {
	gs := NewGameState(2, rand.New(rand.NewSource(1)), nil)
	gs.Active = 0

	atk := addCreature(gs, 0, "Ogre", 3, 3)
	blk := addCreature(gs, 1, "Lone Bander", 1, 4, "banding")
	gs.Seats[1].Hat = &r38BlockerHat{blockers: []*Permanent{blk}}

	CombatPhase(gs)

	if !alive(gs, atk) {
		t.Errorf("attacker should survive 1 dmg from lone blocker")
	}
	if blk.MarkedDamage != 3 {
		t.Errorf("solo bander should retain its 3 marked damage, got %d", blk.MarkedDamage)
	}
}

func TestCombat_R38_ProtectionFromCreatures_BlockerCantBlock(t *testing.T) {
	gs := newCombatGame(t)

	atk := addCreature(gs, 0, "Sky Soldier", 2, 2)
	atk.Flags["prot_type:creature"] = 1

	blk := addCreature(gs, 1, "Big Vanilla", 5, 5)

	if CanBlockGS(gs, atk, blk) {
		t.Fatalf("creature blocker should not be a legal blocker for prot-from-creatures attacker")
	}

	CombatPhase(gs)
	if gs.Seats[1].Life != 18 {
		t.Errorf("attacker should hit defender unblocked: want life 18, got %d", gs.Seats[1].Life)
	}
}

func TestCombat_R38_ProtectionFromArtifacts_PreventsBlockAndDamage(t *testing.T) {
	gs := newCombatGame(t)

	atk := addCreature(gs, 0, "Steel Hekkaton", 4, 4)
	atk.Card.Types = append(atk.Card.Types, "artifact")

	def := addCreature(gs, 1, "Mirran Crusader", 2, 2)
	def.Flags["prot_type:artifact"] = 1

	if CanBlockGS(gs, atk, def) {
		t.Errorf("prot-from-artifacts creature should not be a legal blocker for an artifact attacker")
	}

	applyCombatDamageToCreature(gs, atk, 4, def)
	if def.MarkedDamage != 0 {
		t.Errorf("prot-from-artifacts creature should take 0 damage from artifact source, got marked=%d",
			def.MarkedDamage)
	}
}

type r38GoadHat struct {
	GreedyHatStub
	gotDefenders []int
}

func (h *r38GoadHat) ChooseAttackers(gs *GameState, seatIdx int, legal []*Permanent) []*Permanent {
	return nil
}
func (h *r38GoadHat) ChooseAttackTarget(gs *GameState, seatIdx int, attacker *Permanent, legalDefenders []int) int {
	h.gotDefenders = append([]int{}, legalDefenders...)
	if len(legalDefenders) == 0 {
		return -1
	}
	return legalDefenders[0]
}

func TestCombat_R38_Goad_EngineForcesAttackerWhenHatSkips(t *testing.T) {
	gs := NewGameState(3, rand.New(rand.NewSource(1)), nil)
	gs.Active = 1
	gs.Turn = 5

	goaded := addCreature(gs, 1, "Forced Marcher", 2, 2)
	GoadCreature(gs, 0, goaded)

	if !IsGoaded(goaded, gs.Turn) {
		t.Fatalf("precondition: IsGoaded false at turn %d", gs.Turn)
	}
	if !MustAttackIfAble(gs, goaded) {
		t.Fatalf("precondition: MustAttackIfAble false on controller's own turn")
	}

	hat := &r38GoadHat{}
	gs.Seats[1].Hat = hat

	declared := DeclareAttackers(gs, 1)

	if len(declared) != 1 || declared[0] != goaded {
		t.Fatalf("expected the goaded creature force-added, got %d attackers", len(declared))
	}
	for _, d := range hat.gotDefenders {
		if d == 0 {
			t.Errorf("defender pool included goader seat 0; want it filtered out: %v", hat.gotDefenders)
		}
	}
	if def, ok := AttackerDefender(goaded); !ok || def != 2 {
		t.Errorf("goaded creature should attack seat 2 (not goader seat 0): def=%d ok=%v", def, ok)
	}
}

func TestCombat_R38_Goad_AttacksGoaderWhenOnlyOpponent(t *testing.T) {
	gs := NewGameState(2, rand.New(rand.NewSource(1)), nil)
	gs.Active = 1
	gs.Turn = 5

	goaded := addCreature(gs, 1, "Cornered Soldier", 2, 2)
	GoadCreature(gs, 0, goaded)

	hat := &r38GoadHat{}
	gs.Seats[1].Hat = hat

	declared := DeclareAttackers(gs, 1)
	if len(declared) != 1 || declared[0] != goaded {
		t.Fatalf("expected goaded creature force-attacked, got %d declared", len(declared))
	}
	if def, ok := AttackerDefender(goaded); !ok || def != 0 {
		t.Errorf("goader as only opp: should attack them; got def=%d ok=%v", def, ok)
	}
	if len(hat.gotDefenders) != 1 || hat.gotDefenders[0] != 0 {
		t.Errorf("if-able escape: defender pool should be [0], got %v", hat.gotDefenders)
	}
}
