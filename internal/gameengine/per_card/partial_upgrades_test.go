package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// dev/partial-handler-upgrades — tests for the secondary abilities
// added on top of partial-coverage handlers.

// ---------------------------------------------------------------------
// Omnath, Locus of Creation
// ---------------------------------------------------------------------

func TestOmnath_ETBDrawsCardOnly(t *testing.T) {
	gs := newGame(t, 2)
	beforeLife := gs.Seats[0].Life
	addLibrary(gs, 0, "Plains")
	omnath := addPerm(gs, 0, "Omnath, Locus of Creation", "creature")

	omnathLocusOfCreationETB(gs, omnath)

	if len(gs.Seats[0].Hand) != 1 {
		t.Fatalf("Omnath ETB should draw 1 card; hand=%d", len(gs.Seats[0].Hand))
	}
	if gs.Seats[0].Life != beforeLife {
		t.Fatalf("Omnath ETB should NOT gain life (that's landfall); life=%d", gs.Seats[0].Life)
	}
}

func TestOmnath_LandfallProgression(t *testing.T) {
	gs := newGame(t, 4)
	gs.Turn = 1
	omnath := addPerm(gs, 0, "Omnath, Locus of Creation", "creature")
	omnath.Flags["omnath_landfall_turn"] = gs.Turn
	omnath.Flags["omnath_landfall_count"] = 0

	land := addPerm(gs, 0, "Plains", "land", "basic", "plains")
	beforeLife := gs.Seats[0].Life
	beforePool := gs.Seats[0].ManaPool
	beforeOppLife := gs.Seats[1].Life

	// 1st landfall — gain 4 life.
	omnathLandfall(gs, omnath, map[string]interface{}{
		"controller_seat": 0,
		"perm":            land,
	})
	if gs.Seats[0].Life != beforeLife+4 {
		t.Fatalf("first landfall should gain 4 life; before=%d after=%d", beforeLife, gs.Seats[0].Life)
	}

	// 2nd landfall — add RGWU.
	omnathLandfall(gs, omnath, map[string]interface{}{
		"controller_seat": 0,
		"perm":            land,
	})
	if gs.Seats[0].ManaPool < beforePool+4 {
		t.Fatalf("second landfall should add 4 mana; before=%d after=%d", beforePool, gs.Seats[0].ManaPool)
	}

	// 3rd landfall — 4 damage to each opponent.
	omnathLandfall(gs, omnath, map[string]interface{}{
		"controller_seat": 0,
		"perm":            land,
	})
	if gs.Seats[1].Life != beforeOppLife-4 {
		t.Fatalf("third landfall should deal 4 to opp1; before=%d after=%d", beforeOppLife, gs.Seats[1].Life)
	}
	if gs.Seats[2].Life != beforeOppLife-4 {
		t.Fatalf("third landfall should deal 4 to opp2; got %d", gs.Seats[2].Life)
	}
}

func TestOmnath_LandfallResetsAcrossTurns(t *testing.T) {
	gs := newGame(t, 2)
	gs.Turn = 1
	omnath := addPerm(gs, 0, "Omnath, Locus of Creation", "creature")
	omnath.Flags["omnath_landfall_turn"] = gs.Turn
	land := addPerm(gs, 0, "Plains", "land")
	for i := 0; i < 3; i++ {
		omnathLandfall(gs, omnath, map[string]interface{}{
			"controller_seat": 0,
			"perm":            land,
		})
	}
	// Advance turn — first landfall on the new turn should be "1st" again.
	gs.Turn = 2
	beforeLife := gs.Seats[0].Life
	omnathLandfall(gs, omnath, map[string]interface{}{
		"controller_seat": 0,
		"perm":            land,
	})
	if gs.Seats[0].Life != beforeLife+4 {
		t.Fatalf("new-turn first landfall should gain 4 life; got %d (before %d)", gs.Seats[0].Life, beforeLife)
	}
}

// ---------------------------------------------------------------------
// Marchesa, the Black Rose — dethrone grant
// ---------------------------------------------------------------------

func TestMarchesa_GrantsDethroneToOtherCreatures(t *testing.T) {
	gs := newGame(t, 2)
	marchesa := addPerm(gs, 0, "Marchesa, the Black Rose", "creature")
	other := addPerm(gs, 0, "Goblin Guide", "creature")
	notMine := addPerm(gs, 1, "Goblin Guide", "creature")

	marchesaGrantDethrone(gs, marchesa)

	// Continuous effects apply via the layer pipeline; the chars-resolver
	// runs on demand. We verify by computing characteristics for each
	// permanent and reading the keyword set off the resolved chars.
	otherChars := gameengine.GetEffectiveCharacteristics(gs, other)
	mineChars := gameengine.GetEffectiveCharacteristics(gs, notMine)

	hasDethrone := func(chars *gameengine.Characteristics) bool {
		if chars == nil {
			return false
		}
		for _, k := range chars.Keywords {
			if k == "dethrone" {
				return true
			}
		}
		return false
	}
	if !hasDethrone(otherChars) {
		t.Fatalf("Marchesa should grant dethrone to other creatures we control; keywords=%v", otherChars.Keywords)
	}
	if hasDethrone(mineChars) {
		t.Fatalf("Marchesa should NOT grant dethrone to opponents' creatures")
	}
}

// ---------------------------------------------------------------------
// Kalamax, the Stormsire — real copy on stack
// ---------------------------------------------------------------------

func TestKalamax_RealCopyPushedToStack(t *testing.T) {
	gs := newGame(t, 2)
	kalamax := addPerm(gs, 0, "Kalamax, the Stormsire", "creature")
	kalamax.Tapped = true

	// Put a real instant on the stack so Kalamax has something to copy.
	bolt := &gameengine.Card{Name: "Lightning Bolt", Owner: 0, Types: []string{"instant"}}
	gs.Stack = append(gs.Stack, &gameengine.StackItem{Controller: 0, Card: bolt, Kind: "spell"})

	stackBefore := len(gs.Stack)
	gameengine.FireCardTrigger(gs, "instant_or_sorcery_cast", map[string]interface{}{
		"caster_seat": 0,
		"spell_name":  "Lightning Bolt",
		"card":        bolt,
	})
	if len(gs.Stack) != stackBefore+1 {
		t.Fatalf("Kalamax should push a copy onto the stack; stack=%d (was %d)", len(gs.Stack), stackBefore)
	}
	top := gs.Stack[len(gs.Stack)-1]
	if top == nil || top.Card == nil || !top.IsCopy {
		t.Fatalf("top of stack should be a copy; got %+v", top)
	}
	if kalamax.Counters["+1/+1"] != 1 {
		t.Fatalf("Kalamax should gain a +1/+1 counter on copy; got %d", kalamax.Counters["+1/+1"])
	}
}

func TestKalamax_SorceryDoesNotTrigger(t *testing.T) {
	gs := newGame(t, 2)
	kalamax := addPerm(gs, 0, "Kalamax, the Stormsire", "creature")
	kalamax.Tapped = true
	wrath := &gameengine.Card{Name: "Wrath of God", Owner: 0, Types: []string{"sorcery"}}
	gs.Stack = append(gs.Stack, &gameengine.StackItem{Controller: 0, Card: wrath, Kind: "spell"})

	stackBefore := len(gs.Stack)
	gameengine.FireCardTrigger(gs, "instant_or_sorcery_cast", map[string]interface{}{
		"caster_seat": 0,
		"spell_name":  "Wrath of God",
		"card":        wrath,
	})
	if len(gs.Stack) != stackBefore {
		t.Fatalf("sorcery should not be copied; stack=%d", len(gs.Stack))
	}
	if kalamax.Counters["+1/+1"] != 0 {
		t.Fatalf("Kalamax should not gain a counter on sorcery; got %d", kalamax.Counters["+1/+1"])
	}
}

// ---------------------------------------------------------------------
// Breya, Etherium Shaper
// ---------------------------------------------------------------------

func TestBreya_ETBCreatesTwoThopters(t *testing.T) {
	gs := newGame(t, 2)
	beforeLife := gs.Seats[0].Life
	beforeBF := len(gs.Seats[0].Battlefield)
	breya := addPerm(gs, 0, "Breya, Etherium Shaper", "creature", "artifact")

	breyaEtheriumShaperETB(gs, breya)

	thopters := 0
	for _, p := range gs.Seats[0].Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if p.Card.DisplayName() == "Thopter Token" {
			thopters++
		}
	}
	if thopters != 2 {
		t.Fatalf("Breya ETB should make 2 thopters; got %d on bf (size %d→%d)",
			thopters, beforeBF, len(gs.Seats[0].Battlefield))
	}
	if gs.Seats[0].Life != beforeLife {
		t.Fatalf("Breya ETB should NOT gain life; life=%d (was %d)", gs.Seats[0].Life, beforeLife)
	}
}

func TestBreya_ActivatedDamageMode(t *testing.T) {
	gs := newGame(t, 2)
	breya := addPerm(gs, 0, "Breya, Etherium Shaper", "creature", "artifact")
	beforeOppLife := gs.Seats[1].Life

	breyaEtheriumShaperActivate(gs, breya, 0, map[string]interface{}{
		"mode":        0,
		"target_seat": 1,
	})
	if gs.Seats[1].Life != beforeOppLife-3 {
		t.Fatalf("damage mode should deal 3 to opp; before=%d after=%d", beforeOppLife, gs.Seats[1].Life)
	}
}

func TestBreya_ActivatedMinusFourMode(t *testing.T) {
	gs := newGame(t, 2)
	breya := addPerm(gs, 0, "Breya, Etherium Shaper", "creature", "artifact")
	target := addPerm(gs, 1, "Tarmogoyf", "creature")
	target.Card.BasePower = 4
	target.Card.BaseToughness = 5

	breyaEtheriumShaperActivate(gs, breya, 0, map[string]interface{}{
		"mode":        1,
		"target_perm": target,
	})
	mods := target.Modifications
	if len(mods) == 0 || mods[len(mods)-1].Power != -4 || mods[len(mods)-1].Toughness != -4 {
		t.Fatalf("minus-four mode should record -4/-4 mod; got %+v", mods)
	}
}

func TestBreya_ActivatedLifeMode(t *testing.T) {
	gs := newGame(t, 2)
	breya := addPerm(gs, 0, "Breya, Etherium Shaper", "creature", "artifact")
	beforeLife := gs.Seats[0].Life

	breyaEtheriumShaperActivate(gs, breya, 0, map[string]interface{}{"mode": 2})
	if gs.Seats[0].Life != beforeLife+5 {
		t.Fatalf("life mode should gain 5; before=%d after=%d", beforeLife, gs.Seats[0].Life)
	}
}

// ---------------------------------------------------------------------
// Arcades, the Strategist — toughness-as-damage
// ---------------------------------------------------------------------

func TestArcades_DefenderTPSwap(t *testing.T) {
	gs := newGame(t, 2)
	arcades := addPerm(gs, 0, "Arcades, the Strategist", "creature")
	wall := addPerm(gs, 0, "Wall of Roots", "creature")
	wall.Card.BasePower = 0
	wall.Card.BaseToughness = 5
	wall.Flags["kw:defender"] = 1

	arcadesETBRegisterStatics(gs, arcades)
	// Trigger the layer pipeline by reading the layered P/T accessor.
	power := gs.PowerOf(wall)
	if power != 5 {
		t.Fatalf("Arcades should swap wall P/T to 5/0 (toughness-as-damage); PowerOf=%d ToughnessOf=%d", power, gs.ToughnessOf(wall))
	}
	if wall.Flags["can_attack_with_defender"] != 1 {
		t.Fatalf("Arcades should set can_attack_with_defender flag")
	}
}

func TestArcades_NondefenderUnaffected(t *testing.T) {
	gs := newGame(t, 2)
	arcades := addPerm(gs, 0, "Arcades, the Strategist", "creature")
	bear := addPerm(gs, 0, "Grizzly Bears", "creature")
	bear.Card.BasePower = 2
	bear.Card.BaseToughness = 2

	arcadesETBRegisterStatics(gs, arcades)
	if gs.PowerOf(bear) != 2 {
		t.Fatalf("non-defender should not be swapped; power=%d", gs.PowerOf(bear))
	}
	if bear.Flags["can_attack_with_defender"] != 0 {
		t.Fatalf("non-defender should not get can_attack flag")
	}
}

// ---------------------------------------------------------------------
// Baba Lysaga, Night Witch — type-count gate
// ---------------------------------------------------------------------

func TestBaba_PayoffOnThreeDistinctTypes(t *testing.T) {
	gs := newGame(t, 2)
	baba := addPerm(gs, 0, "Baba Lysaga, Night Witch", "creature")
	addLibrary(gs, 0, "A", "B", "C")
	beforeLife := gs.Seats[0].Life
	beforeOppLife := gs.Seats[1].Life

	c := addPerm(gs, 0, "Bear", "creature")
	a := addPerm(gs, 0, "Ring", "artifact")
	e := addPerm(gs, 0, "Aura", "enchantment")

	babaLysagaNightWitchActivate(gs, baba, 0, map[string]interface{}{
		"sacrifice_perms": []*gameengine.Permanent{c, a, e},
	})

	if len(gs.Seats[0].Hand) != 3 {
		t.Fatalf("Baba should draw 3 with 3 distinct types; hand=%d", len(gs.Seats[0].Hand))
	}
	if gs.Seats[0].Life != beforeLife+3 {
		t.Fatalf("Baba should gain 3 life; before=%d after=%d", beforeLife, gs.Seats[0].Life)
	}
	if gs.Seats[1].Life != beforeOppLife-3 {
		t.Fatalf("Baba should drain opp 3; before=%d after=%d", beforeOppLife, gs.Seats[1].Life)
	}
}

func TestBaba_NoPayoffOnTwoDistinctTypes(t *testing.T) {
	gs := newGame(t, 2)
	baba := addPerm(gs, 0, "Baba Lysaga, Night Witch", "creature")
	addLibrary(gs, 0, "A", "B", "C")
	beforeLife := gs.Seats[0].Life

	c1 := addPerm(gs, 0, "Bear", "creature")
	c2 := addPerm(gs, 0, "Wolf", "creature")
	a := addPerm(gs, 0, "Ring", "artifact")

	babaLysagaNightWitchActivate(gs, baba, 0, map[string]interface{}{
		"sacrifice_perms": []*gameengine.Permanent{c1, c2, a},
	})

	if len(gs.Seats[0].Hand) != 0 {
		t.Fatalf("Baba should NOT draw with only 2 distinct types; hand=%d", len(gs.Seats[0].Hand))
	}
	if gs.Seats[0].Life != beforeLife {
		t.Fatalf("Baba should NOT gain life with only 2 distinct types; life=%d", gs.Seats[0].Life)
	}
}

func TestBaba_TruncatesAtThreeSacrifices(t *testing.T) {
	gs := newGame(t, 2)
	baba := addPerm(gs, 0, "Baba Lysaga, Night Witch", "creature")
	addLibrary(gs, 0, "A", "B", "C")

	a := addPerm(gs, 0, "Ring", "artifact")
	c := addPerm(gs, 0, "Bear", "creature")
	e := addPerm(gs, 0, "Aura", "enchantment")
	land := addPerm(gs, 0, "Plains", "land") // 4th — should NOT be sacrificed

	babaLysagaNightWitchActivate(gs, baba, 0, map[string]interface{}{
		"sacrifice_perms": []*gameengine.Permanent{a, c, e, land},
	})

	// land should still be on the battlefield (truncated past 3).
	stillThere := false
	for _, p := range gs.Seats[0].Battlefield {
		if p == land {
			stillThere = true
			break
		}
	}
	if !stillThere {
		t.Fatalf("Baba should sacrifice at most 3 permanents; the 4th (land) was sacrificed")
	}
}

// ---------------------------------------------------------------------
// Skullbriar, the Walking Grave — counter persistence
// ---------------------------------------------------------------------

func TestSkullbriar_CombatDamageAddsCounter(t *testing.T) {
	gs := newGame(t, 2)
	sk := addPerm(gs, 0, "Skullbriar, the Walking Grave", "creature")

	gameengine.FireCardTrigger(gs, "combat_damage_player", map[string]interface{}{
		"source_seat":   0,
		"source_card":   "Skullbriar, the Walking Grave",
		"defender_seat": 1,
		"amount":        1,
	})

	if sk.Counters["+1/+1"] != 1 {
		t.Fatalf("Skullbriar should gain a +1/+1 on combat damage; got %d", sk.Counters["+1/+1"])
	}
}

func TestSkullbriar_CountersPersistThroughGraveyard(t *testing.T) {
	gs := newGame(t, 2)
	sk := addPerm(gs, 0, "Skullbriar, the Walking Grave", "creature")
	sk.Card.Owner = 0
	sk.AddCounter("+1/+1", 3)

	skullbriarOnLeaveBattlefield(gs, sk, map[string]interface{}{
		"perm":    sk,
		"to_zone": "graveyard",
	})

	// The stash key should now hold 3.
	if gs.Flags[skullbriarPlusKey(0)] != 3 {
		t.Fatalf("Skullbriar should stash 3 +1/+1 counters; got %d", gs.Flags[skullbriarPlusKey(0)])
	}

	// Re-ETB a fresh permanent — counters should be restored.
	sk2 := addPerm(gs, 0, "Skullbriar, the Walking Grave", "creature")
	sk2.Card.Owner = 0
	skullbriarOnETB(gs, sk2)
	if sk2.Counters["+1/+1"] != 3 {
		t.Fatalf("Skullbriar should restore 3 counters on re-ETB; got %d", sk2.Counters["+1/+1"])
	}
	// Stash should be consumed.
	if gs.Flags[skullbriarPlusKey(0)] != 0 {
		t.Fatalf("Skullbriar stash should be cleared on restore; got %d", gs.Flags[skullbriarPlusKey(0)])
	}
}

func TestSkullbriar_CountersDoNotPersistThroughHand(t *testing.T) {
	gs := newGame(t, 2)
	sk := addPerm(gs, 0, "Skullbriar, the Walking Grave", "creature")
	sk.Card.Owner = 0
	sk.AddCounter("+1/+1", 5)

	skullbriarOnLeaveBattlefield(gs, sk, map[string]interface{}{
		"perm":    sk,
		"to_zone": "hand",
	})
	if gs.Flags[skullbriarPlusKey(0)] != 0 {
		t.Fatalf("Skullbriar counters should NOT persist through hand; stash=%d", gs.Flags[skullbriarPlusKey(0)])
	}
}

// ---------------------------------------------------------------------
// Karlov of the Ghost Council — counter doubling + exile activated
// ---------------------------------------------------------------------

func TestKarlov_LifegainAddsTwoCounters(t *testing.T) {
	gs := newGame(t, 2)
	karlov := addPerm(gs, 0, "Karlov of the Ghost Council", "creature")
	karlovOfTheGhostCouncilTrigger(gs, karlov, map[string]interface{}{
		"seat":   0,
		"amount": 1,
	})
	if karlov.Counters["+1/+1"] != 2 {
		t.Fatalf("Karlov should gain TWO +1/+1 on lifegain; got %d", karlov.Counters["+1/+1"])
	}
}

func TestKarlov_ExileActivatedRequiresSixCounters(t *testing.T) {
	gs := newGame(t, 2)
	karlov := addPerm(gs, 0, "Karlov of the Ghost Council", "creature")
	karlov.Counters["+1/+1"] = 5 // one short
	gs.Seats[0].ManaPool = 5
	target := addPerm(gs, 1, "Bear", "creature")

	karlovActivate(gs, karlov, 0, map[string]interface{}{
		"target_perm": target,
	})

	// Should have failed; target still on battlefield.
	stillThere := false
	for _, p := range gs.Seats[1].Battlefield {
		if p == target {
			stillThere = true
			break
		}
	}
	if !stillThere {
		t.Fatalf("Karlov should not exile with only 5 counters; target was exiled")
	}
}

func TestKarlov_ExileActivatedSucceeds(t *testing.T) {
	gs := newGame(t, 2)
	karlov := addPerm(gs, 0, "Karlov of the Ghost Council", "creature")
	karlov.Counters["+1/+1"] = 6
	gs.Seats[0].ManaPool = 5
	target := addPerm(gs, 1, "Bear", "creature")

	karlovActivate(gs, karlov, 0, map[string]interface{}{
		"target_perm": target,
	})

	if karlov.Counters["+1/+1"] != 0 {
		t.Fatalf("Karlov should remove 6 counters; got %d", karlov.Counters["+1/+1"])
	}
	for _, p := range gs.Seats[1].Battlefield {
		if p == target {
			t.Fatalf("target should have been exiled")
		}
	}
	if gs.Seats[0].ManaPool != 3 {
		t.Fatalf("Karlov should spend 2 mana ({W}{B}); pool=%d", gs.Seats[0].ManaPool)
	}
}
