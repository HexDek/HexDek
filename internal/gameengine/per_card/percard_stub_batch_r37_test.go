package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// R37 stub-batch ports — five gen_*.go pure-stub handlers ported into
// real per-card behaviour. Avoids the r36 set (Ashling, Tannuk, Toph,
// Magnus, Morlun). Picks span ETB token creation (Old One Eye),
// Raid combat trigger (Lara Croft), opponent-targeting state flag
// (Maha), combat-damage trigger calling PerformConnive (Norman
// Osborn), and OnCast stamping CostMeta (Thrun).

// ---------------------------------------------------------------------------
// Old One Eye — ETB creates a 5/5 green Tyranid
// ---------------------------------------------------------------------------

func TestOldOneEye_ETBCreatesGreenTyranid5_5(t *testing.T) {
	gs := newGame(t, 2)
	perm := stampCreaturePT(addPerm(gs, 0, "Old One Eye", "creature"), 5, 5)

	preBF := len(gs.Seats[0].Battlefield)
	oldOneEyeETB(gs, perm)

	// Battlefield should now hold Old One Eye + the tyranid token.
	if got := len(gs.Seats[0].Battlefield); got != preBF+1 {
		t.Fatalf("battlefield size after ETB = %d, want %d (added 1 token)", got, preBF+1)
	}
	// Find the token.
	var tok *gameengine.Permanent
	for _, p := range gs.Seats[0].Battlefield {
		if p == perm {
			continue
		}
		if p.Card != nil && p.Card.Name == "Tyranid Token" {
			tok = p
			break
		}
	}
	if tok == nil {
		t.Fatal("Tyranid Token not found on battlefield")
	}
	if tok.Card.BasePower != 5 || tok.Card.BaseToughness != 5 {
		t.Errorf("Tyranid Token P/T = %d/%d, want 5/5",
			tok.Card.BasePower, tok.Card.BaseToughness)
	}
	hasGreen := false
	for _, c := range tok.Card.Colors {
		if c == "G" {
			hasGreen = true
			break
		}
	}
	if !hasGreen {
		t.Errorf("Tyranid Token colors = %v, want includes G", tok.Card.Colors)
	}
	hasTyranidType := false
	for _, ty := range tok.Card.Types {
		if ty == "tyranid" {
			hasTyranidType = true
			break
		}
	}
	if !hasTyranidType {
		t.Errorf("Tyranid Token types = %v, want includes 'tyranid'", tok.Card.Types)
	}
}

// ---------------------------------------------------------------------------
// Lara Croft, Tomb Raider — Raid: Treasure at end of combat if attacked
// ---------------------------------------------------------------------------

func TestLaraCroft_RaidCreatesTreasureWhenAttacked(t *testing.T) {
	gs := newGame(t, 2)
	lara := stampCreaturePT(addPerm(gs, 0, "Lara Croft, Tomb Raider", "creature"), 2, 4)
	gs.Active = 0 // controller's turn
	gs.Seats[0].Turn.Attacked = true

	laraCroftRaidTreasure(gs, lara, nil)

	// Treasure token should now be on the battlefield.
	found := false
	for _, p := range gs.Seats[0].Battlefield {
		if p.Card != nil && p.Card.Name == "Treasure Token" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Treasure Token not found after Raid trigger")
	}
}

func TestLaraCroft_RaidSkipsWhenNoAttack(t *testing.T) {
	gs := newGame(t, 2)
	lara := stampCreaturePT(addPerm(gs, 0, "Lara Croft, Tomb Raider", "creature"), 2, 4)
	gs.Active = 0
	gs.Seats[0].Turn.Attacked = false

	preBF := len(gs.Seats[0].Battlefield)
	laraCroftRaidTreasure(gs, lara, nil)

	if got := len(gs.Seats[0].Battlefield); got != preBF {
		t.Errorf("Treasure should NOT be created when Raid condition unmet; bf delta=%d", got-preBF)
	}
}

func TestLaraCroft_RaidSkipsOnOpponentTurn(t *testing.T) {
	gs := newGame(t, 2)
	lara := stampCreaturePT(addPerm(gs, 0, "Lara Croft, Tomb Raider", "creature"), 2, 4)
	gs.Active = 1 // opponent's turn
	gs.Seats[0].Turn.Attacked = true

	preBF := len(gs.Seats[0].Battlefield)
	laraCroftRaidTreasure(gs, lara, nil)
	if got := len(gs.Seats[0].Battlefield); got != preBF {
		t.Errorf("Raid only on YOUR turn; opp-turn fire should no-op; bf delta=%d", got-preBF)
	}
}

// ---------------------------------------------------------------------------
// Maha, Its Feathers Night — sets opponent base-toughness-1 flag at ETB
// ---------------------------------------------------------------------------

func TestMaha_ETBSetsBaseToughnessFlag(t *testing.T) {
	gs := newGame(t, 2)
	maha := stampCreaturePT(addPerm(gs, 0, "Maha, Its Feathers Night", "creature"), 6, 6)

	if gs.Flags["maha_base_tough_one_active"] != 0 {
		t.Fatal("flag should be unset before ETB")
	}
	mahaETBSetBaseToughnessFlag(gs, maha)
	if got := gs.Flags["maha_base_tough_one_active"]; got != 1 {
		t.Fatalf("flag = %d, want 1 (seat 0+1 offset)", got)
	}

	// Seat 1 maha should set flag to 2 (seat 1 + 1).
	gs2 := newGame(t, 2)
	mahaSeat1 := stampCreaturePT(addPerm(gs2, 1, "Maha, Its Feathers Night", "creature"), 6, 6)
	mahaETBSetBaseToughnessFlag(gs2, mahaSeat1)
	if got := gs2.Flags["maha_base_tough_one_active"]; got != 2 {
		t.Fatalf("seat 1 maha flag = %d, want 2", got)
	}
}

func TestMaha_LTBClearsFlagOnlyForSelf(t *testing.T) {
	gs := newGame(t, 2)
	maha := stampCreaturePT(addPerm(gs, 0, "Maha, Its Feathers Night", "creature"), 6, 6)
	mahaETBSetBaseToughnessFlag(gs, maha)

	// LTB event for a DIFFERENT permanent should NOT clear.
	other := addPerm(gs, 0, "Other Card", "creature")
	mahaLTBClearBaseToughnessFlag(gs, maha, map[string]interface{}{"perm": other})
	if got := gs.Flags["maha_base_tough_one_active"]; got != 1 {
		t.Errorf("LTB on unrelated perm should NOT clear flag; got %d", got)
	}

	// LTB event for Maha itself should clear.
	mahaLTBClearBaseToughnessFlag(gs, maha, map[string]interface{}{"perm": maha})
	if _, exists := gs.Flags["maha_base_tough_one_active"]; exists {
		t.Errorf("LTB on Maha itself should clear flag; got value %d",
			gs.Flags["maha_base_tough_one_active"])
	}
}

func TestMaha_ClearMahaBaseToughnessReset(t *testing.T) {
	gs := newGame(t, 2)
	maha := stampCreaturePT(addPerm(gs, 0, "Maha, Its Feathers Night", "creature"), 6, 6)
	mahaETBSetBaseToughnessFlag(gs, maha)
	ClearMahaBaseToughness(gs)
	if _, exists := gs.Flags["maha_base_tough_one_active"]; exists {
		t.Errorf("explicit ClearMahaBaseToughness should remove flag")
	}
}

// ---------------------------------------------------------------------------
// Norman Osborn — combat damage to player → connive
// ---------------------------------------------------------------------------

func TestNormanOsborn_ConniveOnOwnCombatDamage(t *testing.T) {
	gs := newGame(t, 2)
	norman := stampCreaturePT(addPerm(gs, 0, "Norman Osborn // Green Goblin", "creature"), 3, 3)

	// Seed library so Connive's drawOne has something + hand for discard.
	addLibrary(gs, 0, "Library Top")
	addCardToHand(gs, 0, "Discard Me", 2, "instant") // nonland → +1/+1 counter on discard

	preCounters := norman.Counters["+1/+1"]
	normanOsbornCombatDamageConnive(gs, norman, map[string]interface{}{
		"source_card":   "Norman Osborn // Green Goblin",
		"source_seat":   0,
		"defender_seat": 1,
		"amount":        3,
	})
	if got := norman.Counters["+1/+1"]; got != preCounters+1 {
		t.Errorf("nonland discard should add +1/+1 counter; got %d, want %d",
			got, preCounters+1)
	}
}

func TestNormanOsborn_IgnoresOtherCreatureCombatDamage(t *testing.T) {
	gs := newGame(t, 2)
	norman := stampCreaturePT(addPerm(gs, 0, "Norman Osborn // Green Goblin", "creature"), 3, 3)
	addLibrary(gs, 0, "TopCard")
	addCardToHand(gs, 0, "InHand", 2, "instant")

	preCounters := norman.Counters["+1/+1"]
	preGY := len(gs.Seats[0].Graveyard)

	normanOsbornCombatDamageConnive(gs, norman, map[string]interface{}{
		"source_card":   "Some Other Creature",
		"source_seat":   0,
		"defender_seat": 1,
		"amount":        2,
	})

	if got := norman.Counters["+1/+1"]; got != preCounters {
		t.Errorf("foreign-source damage should NOT trigger Norman's connive; counters changed by %d",
			got-preCounters)
	}
	if got := len(gs.Seats[0].Graveyard); got != preGY {
		t.Errorf("foreign-source damage should not connive (no discard); gy delta=%d", got-preGY)
	}
}

func TestNormanOsborn_IgnoresEnemyControlledCopy(t *testing.T) {
	gs := newGame(t, 2)
	norman := stampCreaturePT(addPerm(gs, 0, "Norman Osborn // Green Goblin", "creature"), 3, 3)
	addLibrary(gs, 0, "T")
	addCardToHand(gs, 0, "H", 2, "instant")

	preCounters := norman.Counters["+1/+1"]
	// Same card name BUT different controller seat — should not fire.
	normanOsbornCombatDamageConnive(gs, norman, map[string]interface{}{
		"source_card":   "Norman Osborn // Green Goblin",
		"source_seat":   1, // enemy seat
		"defender_seat": 0,
		"amount":        3,
	})

	if got := norman.Counters["+1/+1"]; got != preCounters {
		t.Errorf("enemy-controlled same-name Norman should not trigger ours; counters changed by %d",
			got-preCounters)
	}
}

// ---------------------------------------------------------------------------
// Thrun, Breaker of Silence — OnCast stamps CostMeta uncounterable
// ---------------------------------------------------------------------------

func TestThrun_OnCastStampsCannotBeCountered(t *testing.T) {
	gs := newGame(t, 2)
	thrun := &gameengine.Card{
		Name:  "Thrun, Breaker of Silence",
		Owner: 0,
		Types: []string{"creature", "troll", "shaman", "legendary"},
	}
	item := &gameengine.StackItem{
		Card:       thrun,
		Controller: 0,
	}

	thrunCastUncounterable(gs, item)

	if item.CostMeta == nil {
		t.Fatal("CostMeta should be initialized")
	}
	v, ok := item.CostMeta["cannot_be_countered"]
	if !ok {
		t.Fatal("CostMeta[cannot_be_countered] should be set")
	}
	if b, ok := v.(bool); !ok || !b {
		t.Errorf("CostMeta[cannot_be_countered] = %v, want true (bool)", v)
	}
}

func TestThrun_OnCastNilSafe(t *testing.T) {
	thrunCastUncounterable(nil, nil)
	gs := newGame(t, 2)
	thrunCastUncounterable(gs, nil) // nil item — should no-op
}
