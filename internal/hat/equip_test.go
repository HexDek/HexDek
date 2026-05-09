package hat

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
	"github.com/hexdek/hexdek/internal/gameengine"
)

func newEquipment(name string, oracleText string) *gameengine.Card {
	return &gameengine.Card{
		Name:  name,
		Types: []string{"artifact", "equipment"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: oracleText},
			},
		},
	}
}

func TestScoreEquipTarget_PrefersCommander(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.CommanderFormat = true
	gs.Seats[0].CommanderNames = []string{"Big Boss"}

	sword := newEquipment("Sword of Fire", "equip {2}")
	equipPerm := newTestPermanent(gs.Seats[0], sword, 0, 0)

	vanilla := newTestPermanent(gs.Seats[0], newTestCardMinimal("Vanilla", []string{"creature"}, 3, nil), 3, 3)
	commander := newTestPermanent(gs.Seats[0], newTestCardMinimal("Big Boss", []string{"creature"}, 4, nil), 4, 4)

	sVanilla := scoreEquipTarget(gs, 0, equipPerm, vanilla)
	sCommander := scoreEquipTarget(gs, 0, equipPerm, commander)

	if sCommander <= sVanilla {
		t.Errorf("commander (%d) should score higher than vanilla (%d)", sCommander, sVanilla)
	}
}

func TestScoreEquipTarget_PrefersEvasion(t *testing.T) {
	gs := newTestGame(t, 2)

	sword := newEquipment("Sword of Light", "equip {2}")
	equipPerm := newTestPermanent(gs.Seats[0], sword, 0, 0)

	ground := newTestPermanent(gs.Seats[0], newTestCardMinimal("Ground", []string{"creature"}, 3, nil), 3, 3)
	flyer := newTestPermanent(gs.Seats[0], newTestCardMinimal("Flyer", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(flyer, "flying")

	sGround := scoreEquipTarget(gs, 0, equipPerm, ground)
	sFlyer := scoreEquipTarget(gs, 0, equipPerm, flyer)

	if sFlyer <= sGround {
		t.Errorf("flyer (%d) should score higher than ground creature (%d)", sFlyer, sGround)
	}
}

func TestScoreEquipTarget_ConnectTriggerBoostedByEvasion(t *testing.T) {
	gs := newTestGame(t, 2)

	sff := newEquipment("Sword of Feast and Famine", "equip {2}\nwhenever equipped creature deals combat damage to a player, untap all lands")
	equipPerm := newTestPermanent(gs.Seats[0], sff, 0, 0)

	ground := newTestPermanent(gs.Seats[0], newTestCardMinimal("Ground", []string{"creature"}, 3, nil), 3, 3)
	flyer := newTestPermanent(gs.Seats[0], newTestCardMinimal("Flyer", []string{"creature"}, 3, nil), 3, 3)
	addKeyword(flyer, "flying")

	sGround := scoreEquipTarget(gs, 0, equipPerm, ground)
	sFlyer := scoreEquipTarget(gs, 0, equipPerm, flyer)

	diff := sFlyer - sGround
	if diff < 15 {
		t.Errorf("connect trigger on flyer should be much better than ground: flyer=%d, ground=%d, diff=%d (want >=15)", sFlyer, sGround, diff)
	}
}

func TestScoreEquipTarget_DeathTriggerOnCommander(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.CommanderFormat = true
	gs.Seats[0].CommanderNames = []string{"Voltron Leader"}

	skullclamp := newEquipment("Skullclamp", "equip {1}\nwhenever equipped creature dies, draw two cards")
	equipPerm := newTestPermanent(gs.Seats[0], skullclamp, 0, 0)

	normie := newTestPermanent(gs.Seats[0], newTestCardMinimal("Normie", []string{"creature"}, 2, nil), 2, 2)
	cmdr := newTestPermanent(gs.Seats[0], newTestCardMinimal("Voltron Leader", []string{"creature"}, 3, nil), 3, 3)

	sNormie := scoreEquipTarget(gs, 0, equipPerm, normie)
	sCmdr := scoreEquipTarget(gs, 0, equipPerm, cmdr)

	if sCmdr <= sNormie {
		t.Errorf("commander with death-trigger equip should score higher: cmdr=%d, normie=%d", sCmdr, sNormie)
	}
}

func TestScoreEquipTarget_EquipmentStacking(t *testing.T) {
	gs := newTestGame(t, 2)

	sword := newEquipment("Sword A", "equip {2}")
	equipPerm := newTestPermanent(gs.Seats[0], sword, 0, 0)

	bare := newTestPermanent(gs.Seats[0], newTestCardMinimal("Bare", []string{"creature"}, 3, nil), 3, 3)
	stacked := newTestPermanent(gs.Seats[0], newTestCardMinimal("Stacked", []string{"creature"}, 3, nil), 3, 3)

	prevEquip := newTestPermanent(gs.Seats[0], newEquipment("Prev Sword", "equip {2}"), 0, 0)
	prevEquip.AttachedTo = stacked

	sBare := scoreEquipTarget(gs, 0, equipPerm, bare)
	sStacked := scoreEquipTarget(gs, 0, equipPerm, stacked)

	if sStacked <= sBare {
		t.Errorf("creature with existing equipment should score higher for stacking: stacked=%d, bare=%d", sStacked, sBare)
	}
}

func TestScoreEquipTarget_IndestructibleCommanderBonus(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.CommanderFormat = true
	gs.Seats[0].CommanderNames = []string{"Immortal King"}

	resOrb := newEquipment("Resurrection Orb", "equip {2}\nequipped creature has indestructible")
	equipPerm := newTestPermanent(gs.Seats[0], resOrb, 0, 0)

	normie := newTestPermanent(gs.Seats[0], newTestCardMinimal("Normie", []string{"creature"}, 3, nil), 3, 3)
	cmdr := newTestPermanent(gs.Seats[0], newTestCardMinimal("Immortal King", []string{"creature"}, 4, nil), 4, 4)

	sNormie := scoreEquipTarget(gs, 0, equipPerm, normie)
	sCmdr := scoreEquipTarget(gs, 0, equipPerm, cmdr)

	if sCmdr-sNormie < 20 {
		t.Errorf("indestructible equipment on commander should have large bonus: cmdr=%d, normie=%d, diff=%d (want >=20)", sCmdr, sNormie, sCmdr-sNormie)
	}
}

func TestScoreEquipTarget_UnblockableWithConnect(t *testing.T) {
	gs := newTestGame(t, 2)

	sff := newEquipment("Sword of Feast and Famine", "equip {2}\nwhenever equipped creature deals combat damage to a player, untap all lands")
	equipPerm := newTestPermanent(gs.Seats[0], sff, 0, 0)

	normal := newTestPermanent(gs.Seats[0], newTestCardMinimal("Normal", []string{"creature"}, 3, nil), 3, 3)

	unblockable := newTestPermanent(gs.Seats[0], &gameengine.Card{
		Name:  "Invisible Stalker",
		Types: []string{"creature", "cost:2"},
		AST: &gameast.CardAST{
			Name: "Invisible Stalker",
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: "invisible stalker can't be blocked"},
			},
		},
		BasePower:     1,
		BaseToughness: 1,
	}, 1, 1)

	sNormal := scoreEquipTarget(gs, 0, equipPerm, normal)
	sUnblockable := scoreEquipTarget(gs, 0, equipPerm, unblockable)

	if sUnblockable <= sNormal {
		t.Errorf("unblockable with connect trigger should beat bigger ground creature: unblockable=%d, normal=%d", sUnblockable, sNormal)
	}
}

func TestScoreEquipTarget_NoTargets(t *testing.T) {
	gs := newTestGame(t, 2)

	sword := newEquipment("Sword", "equip {2}")
	_ = newTestPermanent(gs.Seats[0], sword, 0, 0)

	score := scoreEquipTarget(gs, 0, nil, nil)
	if score != 0 {
		t.Errorf("no creatures should score 0; got %d", score)
	}
}

func TestScoreEquipTarget_SummoningSickPenalty(t *testing.T) {
	gs := newTestGame(t, 2)

	sword := newEquipment("Sword", "equip {2}")
	equipPerm := newTestPermanent(gs.Seats[0], sword, 0, 0)

	ready := newTestPermanent(gs.Seats[0], newTestCardMinimal("Ready", []string{"creature"}, 3, nil), 3, 3)
	sick := newTestPermanent(gs.Seats[0], newTestCardMinimal("Sick", []string{"creature"}, 3, nil), 3, 3)
	sick.SummoningSick = true

	sReady := scoreEquipTarget(gs, 0, equipPerm, ready)
	sSick := scoreEquipTarget(gs, 0, equipPerm, sick)

	if sReady <= sSick {
		t.Errorf("non-summoning-sick should score higher: ready=%d, sick=%d", sReady, sSick)
	}
}
