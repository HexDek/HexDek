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

// TestScoreEquipTarget_CommanderStackingPriority: Voltron stacking should
// favor the commander much more aggressively than a non-commander, because
// the gameplan IS commander damage. With one equipment already attached,
// the commander should pull ahead of an equally-sized non-commander.
func TestScoreEquipTarget_CommanderStackingPriority(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.CommanderFormat = true
	gs.Seats[0].CommanderNames = []string{"Voltron Lord"}

	sword := newEquipment("Sword B", "equip {2}")
	equipPerm := newTestPermanent(gs.Seats[0], sword, 0, 0)

	beater := newTestPermanent(gs.Seats[0], newTestCardMinimal("Beater", []string{"creature"}, 3, nil), 3, 3)
	cmdr := newTestPermanent(gs.Seats[0], newTestCardMinimal("Voltron Lord", []string{"creature"}, 3, nil), 3, 3)

	prevBeater := newTestPermanent(gs.Seats[0], newEquipment("Prev1", "equip {2}"), 0, 0)
	prevBeater.AttachedTo = beater
	prevCmdr := newTestPermanent(gs.Seats[0], newEquipment("Prev2", "equip {2}"), 0, 0)
	prevCmdr.AttachedTo = cmdr

	sBeater := scoreEquipTarget(gs, 0, equipPerm, beater)
	sCmdr := scoreEquipTarget(gs, 0, equipPerm, cmdr)

	if sCmdr-sBeater < 10 {
		t.Errorf("commander voltron stacking should outpace non-commander stacking: cmdr=%d, beater=%d, diff=%d (want >=10)", sCmdr, sBeater, sCmdr-sBeater)
	}
}

// TestScoreEquipTarget_ConcentrationKicker: a creature with 2+ existing
// equipment ("build target") should gain extra concentration credit beyond
// the linear per-stack bonus, so the AI doesn't keep spreading after the
// pile is already worth committing to.
func TestScoreEquipTarget_ConcentrationKicker(t *testing.T) {
	gs := newTestGame(t, 2)

	sword := newEquipment("New Sword", "equip {2}")
	equipPerm := newTestPermanent(gs.Seats[0], sword, 0, 0)

	one := newTestPermanent(gs.Seats[0], newTestCardMinimal("OneStack", []string{"creature"}, 3, nil), 3, 3)
	two := newTestPermanent(gs.Seats[0], newTestCardMinimal("TwoStack", []string{"creature"}, 3, nil), 3, 3)

	e1 := newTestPermanent(gs.Seats[0], newEquipment("E1", "equip {2}"), 0, 0)
	e1.AttachedTo = one

	e2a := newTestPermanent(gs.Seats[0], newEquipment("E2a", "equip {2}"), 0, 0)
	e2a.AttachedTo = two
	e2b := newTestPermanent(gs.Seats[0], newEquipment("E2b", "equip {2}"), 0, 0)
	e2b.AttachedTo = two

	sOne := scoreEquipTarget(gs, 0, equipPerm, one)
	sTwo := scoreEquipTarget(gs, 0, equipPerm, two)

	// Linear-only would predict diff == 4 (one extra equipment * 4).
	// Concentration kicker should push the gap to at least 10.
	diff := sTwo - sOne
	if diff < 10 {
		t.Errorf("2-stacked target should beat 1-stacked target by concentration kicker: two=%d, one=%d, diff=%d (want >=10)", sTwo, sOne, diff)
	}
}

// TestScoreEquipTarget_SkullclampRecurrence: Skullclamp-style equipment
// (cheap equip + draw on death) should prefer disposable, low-toughness
// targets — the loop only fires when the equipped creature dies.
func TestScoreEquipTarget_SkullclampRecurrence(t *testing.T) {
	gs := newTestGame(t, 2)

	clamp := newEquipment("Skullclamp", "equip {1}\nequipped creature gets +1/-1.\nwhenever equipped creature dies, draw two cards")
	equipPerm := newTestPermanent(gs.Seats[0], clamp, 0, 0)

	beefy := newTestPermanent(gs.Seats[0], newTestCardMinimal("Beefy", []string{"creature"}, 4, nil), 4, 4)
	chump := newTestPermanent(gs.Seats[0], newTestCardMinimal("Chump", []string{"creature"}, 1, nil), 1, 1)

	sBeefy := scoreEquipTarget(gs, 0, equipPerm, beefy)
	sChump := scoreEquipTarget(gs, 0, equipPerm, chump)

	if sChump <= sBeefy {
		t.Errorf("Skullclamp should prefer 1-toughness target over a beater: chump=%d, beefy=%d", sChump, sBeefy)
	}
}

// TestScoreEquipTarget_SkullclampPrefersToken: tokens are the ideal
// Skullclamp fodder — they have no return-from-graveyard value.
func TestScoreEquipTarget_SkullclampPrefersToken(t *testing.T) {
	gs := newTestGame(t, 2)

	clamp := newEquipment("Skullclamp", "equip {1}\nwhenever equipped creature dies, draw two cards")
	equipPerm := newTestPermanent(gs.Seats[0], clamp, 0, 0)

	realCard := newTestPermanent(gs.Seats[0], newTestCardMinimal("Real", []string{"creature"}, 1, nil), 1, 1)
	tokenCard := newTestCardMinimal("Saproling", []string{"creature", "token"}, 0, nil)
	token := newTestPermanent(gs.Seats[0], tokenCard, 1, 1)

	sReal := scoreEquipTarget(gs, 0, equipPerm, realCard)
	sToken := scoreEquipTarget(gs, 0, equipPerm, token)

	if sToken <= sReal {
		t.Errorf("Skullclamp should prefer a token over a non-token of the same body: token=%d, real=%d", sToken, sReal)
	}
}

// TestScoreEquipTarget_ExpensiveEquipNotRecurrence: expensive equip costs
// disqualify the recurrence pattern — even with a death trigger, equip {5}
// is not a loop, so we should NOT prefer the chump over a 4/4 attacker.
func TestScoreEquipTarget_ExpensiveEquipNotRecurrence(t *testing.T) {
	gs := newTestGame(t, 2)

	bigEquip := newEquipment("Doom Anvil", "equip {5}\nwhenever equipped creature dies, you draw two cards")
	equipPerm := newTestPermanent(gs.Seats[0], bigEquip, 0, 0)

	beater := newTestPermanent(gs.Seats[0], newTestCardMinimal("Beater", []string{"creature"}, 4, nil), 4, 4)
	chump := newTestPermanent(gs.Seats[0], newTestCardMinimal("Chump", []string{"creature"}, 1, nil), 1, 1)

	sBeater := scoreEquipTarget(gs, 0, equipPerm, beater)
	sChump := scoreEquipTarget(gs, 0, equipPerm, chump)

	if sChump > sBeater {
		t.Errorf("expensive equip cost should not trigger recurrence preference: chump=%d, beater=%d (want beater >= chump)", sChump, sBeater)
	}
}

// TestScoreEquipTarget_SwordCycleConnectPayoff: Sword cycle (cheap equip
// + connect-payoff trigger) should heavily favor unblockable / evasive
// carriers because the trigger fires every turn.
func TestScoreEquipTarget_SwordCycleConnectPayoff(t *testing.T) {
	gs := newTestGame(t, 2)

	sff := newEquipment("Sword of Feast and Famine", "equip {2}\nwhenever equipped creature deals combat damage to a player, untap all lands you control")
	equipPerm := newTestPermanent(gs.Seats[0], sff, 0, 0)

	groundBig := newTestPermanent(gs.Seats[0], newTestCardMinimal("GroundBig", []string{"creature"}, 5, nil), 5, 5)
	flyer := newTestPermanent(gs.Seats[0], newTestCardMinimal("Flyer", []string{"creature"}, 2, nil), 2, 2)
	addKeyword(flyer, "flying")

	sGround := scoreEquipTarget(gs, 0, equipPerm, groundBig)
	sFlyer := scoreEquipTarget(gs, 0, equipPerm, flyer)

	if sFlyer <= sGround {
		t.Errorf("Sword cycle should prefer evasive 2/2 over ground 5/5: flyer=%d, ground=%d", sFlyer, sGround)
	}
}

// TestScoreEquipTarget_TrampleRiderHighPower: trample/lifelink/double-
// strike riders should get a synergy bonus on a high-power body where
// the keyword actually moves damage.
func TestScoreEquipTarget_TrampleRiderHighPower(t *testing.T) {
	gs := newTestGame(t, 2)

	trampleSword := newEquipment("Trample Maul", "equip {2}\nequipped creature has trample")
	equipPerm := newTestPermanent(gs.Seats[0], trampleSword, 0, 0)

	weak := newTestPermanent(gs.Seats[0], newTestCardMinimal("Weak", []string{"creature"}, 1, nil), 1, 1)
	beefy := newTestPermanent(gs.Seats[0], newTestCardMinimal("Beefy", []string{"creature"}, 5, nil), 5, 5)

	sWeak := scoreEquipTarget(gs, 0, equipPerm, weak)
	sBeefy := scoreEquipTarget(gs, 0, equipPerm, beefy)

	// Body alone gives beefy +14 (5*2+5 vs 1*2+1 = 15-3 = 12 raw),
	// the trample synergy adds +6 on top, so gap should be >= 16.
	gap := sBeefy - sWeak
	if gap < 16 {
		t.Errorf("trample rider on big body should get synergy bonus: beefy=%d, weak=%d, gap=%d (want >=16)", sBeefy, sWeak, gap)
	}
}

// TestScoreEquipTarget_ProtectionRiderCommander: hexproof/ward/protection
// riders get an extra commander bonus because protecting the commander
// from removal is uniquely valuable (commander tax, gameplan reliance).
func TestScoreEquipTarget_ProtectionRiderCommander(t *testing.T) {
	gs := newTestGame(t, 2)
	gs.CommanderFormat = true
	gs.Seats[0].CommanderNames = []string{"Targeted Lord"}

	hexCloak := newEquipment("Swiftfoot Boots", "equip {1}\nequipped creature has hexproof and haste")
	equipPerm := newTestPermanent(gs.Seats[0], hexCloak, 0, 0)

	plain := newEquipment("Plain Boots", "equip {1}\nequipped creature gets +1/+1")
	plainPerm := newTestPermanent(gs.Seats[0], plain, 0, 0)

	cmdr := newTestPermanent(gs.Seats[0], newTestCardMinimal("Targeted Lord", []string{"creature"}, 3, nil), 3, 3)

	sHex := scoreEquipTarget(gs, 0, equipPerm, cmdr)
	sPlain := scoreEquipTarget(gs, 0, plainPerm, cmdr)

	if sHex <= sPlain {
		t.Errorf("protection rider on commander should beat plain rider: hex=%d, plain=%d", sHex, sPlain)
	}
}

// TestYggdrasilChooseTarget_RecoversSourceEquipment: when an equipment's
// equip ability is on the stack, ChooseTarget should detect the source
// from gs.Stack and pass it into scoreEquipTarget so equipment-specific
// signals (recurrence, connect-payoff, riders) take effect.
//
// Setup: Skullclamp on stack, two friendly creatures (a 1-toughness
// chump and a 4/4 beater). Without source recovery, the 4/4 wins on
// raw P/T. With source recovery, the chump wins because Skullclamp
// triggers on death.
func TestYggdrasilChooseTarget_RecoversSourceEquipment(t *testing.T) {
	gs := newTestGame(t, 2)

	clamp := newEquipment("Skullclamp", "equip {1}\nwhenever equipped creature dies, draw two cards")
	clampPerm := newTestPermanent(gs.Seats[0], clamp, 0, 0)

	beater := newTestPermanent(gs.Seats[0], newTestCardMinimal("Beater", []string{"creature"}, 4, nil), 4, 4)
	chump := newTestPermanent(gs.Seats[0], newTestCardMinimal("Chump", []string{"creature"}, 1, nil), 1, 1)

	// Push the equip activation on the stack so ChooseTarget can recover
	// the source.
	gs.Stack = append(gs.Stack, &gameengine.StackItem{
		Controller: 0,
		Source:     clampPerm,
		Kind:       "activated",
	})

	hat := NewYggdrasilHat(nil, 0)
	legal := []gameengine.Target{
		{Kind: gameengine.TargetKindPermanent, Permanent: beater},
		{Kind: gameengine.TargetKindPermanent, Permanent: chump},
	}
	chosen := hat.ChooseTarget(gs, 0, gameast.Filter{Base: "creature", YouControl: true, Targeted: true}, legal)
	if chosen.Permanent != chump {
		var name string
		if chosen.Permanent != nil && chosen.Permanent.Card != nil {
			name = chosen.Permanent.Card.Name
		}
		t.Errorf("ChooseTarget should pick chump (Skullclamp recurrence target), got %q", name)
	}
}

// TestYggdrasilChooseTarget_FallbackSingleEquipment: when no stack item
// is present but exactly one equipment with an "equip" cost is on our
// battlefield, ChooseTarget falls back to using it as the source so
// equipment-specific scoring still applies.
func TestYggdrasilChooseTarget_FallbackSingleEquipment(t *testing.T) {
	gs := newTestGame(t, 2)

	clamp := newEquipment("Skullclamp", "equip {1}\nwhenever equipped creature dies, draw two cards")
	_ = newTestPermanent(gs.Seats[0], clamp, 0, 0)

	beater := newTestPermanent(gs.Seats[0], newTestCardMinimal("Beater", []string{"creature"}, 4, nil), 4, 4)
	chump := newTestPermanent(gs.Seats[0], newTestCardMinimal("Chump", []string{"creature"}, 1, nil), 1, 1)

	hat := NewYggdrasilHat(nil, 0)
	legal := []gameengine.Target{
		{Kind: gameengine.TargetKindPermanent, Permanent: beater},
		{Kind: gameengine.TargetKindPermanent, Permanent: chump},
	}
	chosen := hat.ChooseTarget(gs, 0, gameast.Filter{Base: "creature", YouControl: true, Targeted: true}, legal)
	if chosen.Permanent != chump {
		t.Errorf("fallback path: with only Skullclamp on field, should prefer chump")
	}
}

// TestEquipCostFromText covers the common shorthand forms.
func TestEquipCostFromText(t *testing.T) {
	cases := []struct {
		text string
		want int
	}{
		{"equip {1}", 1},
		{"equip {2}", 2},
		{"equip {0}", 0},
		{"equip {2}{w}", 3},
		{"equip {3}{r}{r}", 5},
		{"equip 4", 4},
		{"reach\nequip {1}", 1},
		{"no equip cost here", -1},
		{"equip", -1},
	}
	for _, c := range cases {
		got := equipCostFromText(c.text)
		if got != c.want {
			t.Errorf("equipCostFromText(%q) = %d, want %d", c.text, got, c.want)
		}
	}
}
