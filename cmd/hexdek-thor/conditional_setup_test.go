package main

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
	"github.com/hexdek/hexdek/internal/gameengine"
)

func TestDetectConditionScaffold(t *testing.T) {
	cases := []struct {
		name     string
		kind     string
		text     string
		wantKind conditionScaffoldKind
		wantSub  string
		wantCnt  int
	}{
		{
			name:     "Land Tax",
			kind:     "intervening_if",
			text:     "an opponent controls more lands than you",
			wantKind: condScaffoldOpponentMoreLands,
		},
		{
			name:     "Knight of the White Orchid",
			kind:     "intervening_if",
			text:     "an opponent controls more lands than you do",
			wantKind: condScaffoldOpponentMoreLands,
		},
		{
			name:     "Ghitu Journeymage",
			kind:     "intervening_if",
			text:     "you control another wizard",
			wantKind: condScaffoldYouControlSubtype,
			wantSub:  "wizard",
		},
		{
			name:     "Compy Swarm",
			kind:     "intervening_if",
			text:     "a creature died this turn",
			wantKind: condScaffoldCreatureDiedThisTurn,
		},
		{
			name:     "Oversold Cemetery",
			kind:     "intervening_if",
			text:     "there are four or more creature cards in your graveyard",
			wantKind: condScaffoldCreatureCardsInGraveyard,
			wantCnt:  4,
		},
		{
			name:     "Ichorid",
			kind:     "intervening_if",
			text:     "a black creature card in your graveyard",
			wantKind: condScaffoldCreatureCardsInGraveyard,
			wantCnt:  4,
		},
		{
			name:     "Lux Artillery",
			kind:     "intervening_if",
			text:     "you have 30 or more energy counters",
			wantKind: condScaffoldEnergyThreshold,
			wantCnt:  30,
		},
		{
			name:     "Generic graveyard target",
			kind:     "intervening_if",
			text:     "a card in your graveyard",
			wantKind: condScaffoldCardInGraveyard,
		},
		{
			name:     "Unknown wraps to none",
			kind:     "intervening_if",
			text:     "the moon is full",
			wantKind: condScaffoldNone,
		},
		{
			name:     "Wrong kind returns none",
			kind:     "fateful_hour",
			text:     "an opponent controls more lands than you",
			wantKind: condScaffoldNone,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cond := &gameast.Condition{Kind: tc.kind, Args: []interface{}{tc.text}}
			got := detectConditionScaffold(cond)
			if got.kind != tc.wantKind {
				t.Errorf("kind: want %v, got %v", tc.wantKind, got.kind)
			}
			if tc.wantSub != "" && got.subtype != tc.wantSub {
				t.Errorf("subtype: want %q, got %q", tc.wantSub, got.subtype)
			}
			if tc.wantCnt != 0 && got.count != tc.wantCnt {
				t.Errorf("count: want %d, got %d", tc.wantCnt, got.count)
			}
		})
	}
}

func TestApplyConditionScaffolding_LandTax(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "intervening_if",
		Args: []interface{}{"an opponent controls more lands than you"},
	}
	cs := applyConditionScaffolding(gs, cond, nil)
	if cs.kind != condScaffoldOpponentMoreLands {
		t.Fatalf("expected OpponentMoreLands, got %v", cs.kind)
	}
	lands := 0
	for _, p := range gs.Seats[1].Battlefield {
		for _, ty := range p.Card.Types {
			if ty == "land" {
				lands++
				break
			}
		}
	}
	if lands < 6 {
		t.Errorf("seat 1 wanted >=6 lands, got %d", lands)
	}
}

func TestApplyConditionScaffolding_OversoldCemetery(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "intervening_if",
		Args: []interface{}{"there are four or more creature cards in your graveyard"},
	}
	cs := applyConditionScaffolding(gs, cond, nil)
	if cs.kind != condScaffoldCreatureCardsInGraveyard {
		t.Fatalf("expected CreatureCardsInGraveyard, got %v", cs.kind)
	}
	creatures := 0
	for _, c := range gs.Seats[0].Graveyard {
		for _, ty := range c.Types {
			if ty == "creature" {
				creatures++
				break
			}
		}
	}
	if creatures < 4 {
		t.Errorf("seat 0 graveyard wanted >=4 creatures, got %d", creatures)
	}
}

func TestApplyConditionScaffolding_CreatureDied(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "intervening_if",
		Args: []interface{}{"a creature died this turn"},
	}
	applyConditionScaffolding(gs, cond, nil)
	if gs.Flags["creature_died_this_turn"] != 1 {
		t.Errorf("creature_died_this_turn flag not set")
	}
}

func TestApplyConditionScaffolding_GhituWizard(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "intervening_if",
		Args: []interface{}{"you control another wizard"},
	}
	applyConditionScaffolding(gs, cond, nil)
	found := false
	for _, p := range gs.Seats[0].Battlefield {
		for _, ty := range p.Card.Types {
			if ty == "wizard" {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("no wizard creature placed on seat 0")
	}
}

func TestPrimeInterveningIf_GainedLife(t *testing.T) {
	gs := newTestGameState(2)
	info := &effectInfo{
		effect: &gameast.ModificationEffect{
			ModKind: "conditional_effect",
			Args:    []interface{}{"if you gained life this turn, draw a card"},
		},
	}
	if !primeInterveningIf(gs, info, nil, nil) {
		t.Fatalf("expected priming to fire for gained life")
	}
	if gs.Seats[0].Flags["life_gained_this_turn"] <= 0 {
		t.Errorf("life_gained_this_turn flag not set: %v", gs.Seats[0].Flags)
	}
	if gs.Seats[0].Life <= 20 {
		t.Errorf("seat0 life should have increased from 20, got %d", gs.Seats[0].Life)
	}
}

func TestPrimeInterveningIf_GainedOrLost(t *testing.T) {
	gs := newTestGameState(2)
	info := &effectInfo{
		effect: &gameast.ModificationEffect{
			ModKind: "conditional_effect",
			Args:    []interface{}{"if you gained or lost life this turn, look at the top four cards"},
		},
	}
	if !primeInterveningIf(gs, info, nil, nil) {
		t.Fatalf("expected priming to fire")
	}
	if gs.Seats[0].Flags["life_gained_this_turn"] <= 0 {
		t.Errorf("gained flag not set")
	}
	if gs.Seats[0].Flags["life_lost_this_turn"] <= 0 {
		t.Errorf("lost flag not set")
	}
}

func TestPrimeInterveningIf_CounterPlaced(t *testing.T) {
	gs := newTestGameState(2)
	info := &effectInfo{
		effect: &gameast.ModificationEffect{
			ModKind: "conditional_effect",
			Args:    []interface{}{"if you put a counter on a creature this turn, investigate"},
		},
	}
	if !primeInterveningIf(gs, info, nil, nil) {
		t.Fatalf("expected priming to fire for counter_placed")
	}
	if gs.Seats[0].Flags["counter_placed_this_turn"] != 1 {
		t.Errorf("counter_placed_this_turn flag not set: %v", gs.Seats[0].Flags)
	}
	found := false
	for _, p := range gs.Seats[0].Battlefield {
		if p.Counters["+1/+1"] >= 1 {
			found = true
		}
	}
	if !found {
		t.Errorf("no creature with +1/+1 counter on seat 0")
	}
}

func TestPrimeInterveningIf_LifeMoreThanStarting(t *testing.T) {
	gs := newTestGameState(2)
	info := &effectInfo{
		effect: &gameast.ModificationEffect{
			ModKind: "conditional_effect",
			Args:    []interface{}{"if you have at least 15 life more than your starting life total, each player loses the game"},
		},
	}
	if !primeInterveningIf(gs, info, nil, nil) {
		t.Fatalf("expected priming to fire")
	}
	if gs.Seats[0].Life < 55 {
		t.Errorf("expected seat0 Life >= 55 (40 starting + 15), got %d", gs.Seats[0].Life)
	}
}

func TestPrimeInterveningIf_Ascend(t *testing.T) {
	gs := newTestGameState(2)
	info := &effectInfo{
		effect: &gameast.ModificationEffect{
			ModKind: "conditional_effect",
			Args:    []interface{}{"if you have the city's blessing, reveal the top card of your library"},
		},
	}
	if !primeInterveningIf(gs, info, nil, nil) {
		t.Fatalf("expected priming to fire")
	}
	if gs.Seats[0].Flags["citys_blessing"] != 1 {
		t.Errorf("citys_blessing flag not set: %v", gs.Seats[0].Flags)
	}
	if len(gs.Seats[0].Battlefield) < 10 {
		t.Errorf("expected at least 10 permanents, got %d", len(gs.Seats[0].Battlefield))
	}
}

func TestPrimeInterveningIf_AnotherKnight(t *testing.T) {
	gs := newTestGameState(2)
	info := &effectInfo{
		effect: &gameast.ModificationEffect{
			ModKind: "conditional_effect",
			Args:    []interface{}{"if you control another knight, look at the top five cards of your library"},
		},
	}
	if !primeInterveningIf(gs, info, nil, nil) {
		t.Fatalf("expected priming to fire")
	}
	found := false
	for _, p := range gs.Seats[0].Battlefield {
		for _, ty := range p.Card.Types {
			if ty == "knight" {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("no knight creature placed on seat 0")
	}
}

func TestPrimeInterveningIf_ExiledWith(t *testing.T) {
	gs := newTestGameState(2)
	src := &gameengine.Permanent{
		Card:       &gameengine.Card{Name: "Smirking Spelljacker", Owner: 0, Types: []string{"creature"}},
		Controller: 0,
		Owner:      0,
		Flags:      map[string]int{},
	}
	info := &effectInfo{
		effect: &gameast.ModificationEffect{
			ModKind: "conditional_effect",
			Args:    []interface{}{"if a card is exiled with it, you may cast the exiled card"},
		},
	}
	if !primeInterveningIf(gs, info, src, nil) {
		t.Fatalf("expected priming to fire")
	}
	if len(gs.Seats[0].Exile) == 0 {
		t.Errorf("expected card in seat 0 exile zone")
	}
	if src.Flags["card_exiled_with"] != 1 {
		t.Errorf("card_exiled_with flag not set on src: %v", src.Flags)
	}
}

func TestPrimeInterveningIf_NoMatch(t *testing.T) {
	gs := newTestGameState(2)
	info := &effectInfo{
		effect: &gameast.ModificationEffect{
			ModKind: "conditional_effect",
			Args:    []interface{}{"if the moon turns blue, win the game"},
		},
	}
	if primeInterveningIf(gs, info, nil, nil) {
		t.Errorf("expected no priming for unrecognised condition")
	}
}

// ---------------------------------------------------------------------------
// Detection tests for new condScaffold kinds.
// ---------------------------------------------------------------------------

func TestDetectConditionScaffold_NewKinds(t *testing.T) {
	cases := []struct {
		name     string
		kind     string
		text     string
		wantKind conditionScaffoldKind
		wantThr  int // threshold
	}{
		{
			name:     "gained life this turn",
			kind:     "intervening_if",
			text:     "you gained life this turn",
			wantKind: condScaffoldGainedLifeThisTurn,
		},
		{
			name:     "gain life this turn variant",
			kind:     "intervening_if",
			text:     "if you gain life this turn, draw a card",
			wantKind: condScaffoldGainedLifeThisTurn,
		},
		{
			name:     "cast a spell this turn",
			kind:     "intervening_if",
			text:     "you cast a spell this turn",
			wantKind: condScaffoldCastSpellThisTurn,
		},
		{
			name:     "cast a noncreature spell this turn",
			kind:     "intervening_if",
			text:     "you cast a noncreature spell this turn",
			wantKind: condScaffoldCastSpellThisTurn,
		},
		{
			name:     "creature entered battlefield this turn",
			kind:     "intervening_if",
			text:     "a creature entered the battlefield under your control this turn",
			wantKind: condScaffoldCreatureETBThisTurn,
		},
		{
			name:     "drew a card this turn",
			kind:     "raw",
			text:     "if you've drawn a card this turn",
			wantKind: condScaffoldDrawnCardThisTurn,
		},
		{
			name:     "attacked this turn",
			kind:     "intervening_if",
			text:     "if you attacked this turn",
			wantKind: condScaffoldAttackedThisTurn,
		},
		{
			name:     "creature attacked this turn",
			kind:     "intervening_if",
			text:     "if a creature attacked this turn",
			wantKind: condScaffoldAttackedThisTurn,
		},
		{
			name:     "sacrificed this turn",
			kind:     "intervening_if",
			text:     "if you sacrificed a creature this turn",
			wantKind: condScaffoldSacrificedThisTurn,
		},
		{
			name:     "combat damage dealt this turn",
			kind:     "intervening_if",
			text:     "a creature dealt combat damage to a player this turn",
			wantKind: condScaffoldCombatDamageDealt,
		},
		{
			name:     "landfall",
			kind:     "intervening_if",
			text:     "landfall — if a land entered the battlefield",
			wantKind: condScaffoldLandfallThisTurn,
		},
		{
			name:     "land entered this turn",
			kind:     "raw",
			text:     "if a land entered the battlefield under your control this turn",
			wantKind: condScaffoldLandfallThisTurn,
		},
		{
			name:     "played a land this turn",
			kind:     "intervening_if",
			text:     "if you played a land this turn",
			wantKind: condScaffoldLandfallThisTurn,
		},
		{
			name:     "discarded this turn",
			kind:     "intervening_if",
			text:     "if you discarded a card this turn",
			wantKind: condScaffoldDiscardedThisTurn,
		},
		{
			name:     "enchanted creature",
			kind:     "raw",
			text:     "enchanted creature has flying",
			wantKind: condScaffoldEnchantedCreature,
		},
		{
			name:     "opponent lost life this turn",
			kind:     "intervening_if",
			text:     "if an opponent lost life this turn",
			wantKind: condScaffoldOpponentLostLife,
		},
		{
			name:     "life above threshold 25",
			kind:     "intervening_if",
			text:     "if you have 25 or more life",
			wantKind: condScaffoldLifeAboveThreshold,
			wantThr:  25,
		},
		{
			name:     "life below threshold 5",
			kind:     "intervening_if",
			text:     "if you have 5 or less life",
			wantKind: condScaffoldLifeBelowThreshold,
			wantThr:  5,
		},
		{
			name:     "life total is 10 or less",
			kind:     "raw",
			text:     "your life total is 10 or less",
			wantKind: condScaffoldLifeBelowThreshold,
			wantThr:  10,
		},
		{
			name:     "upkeep condition",
			kind:     "raw",
			text:     "during your upkeep, you may pay 2",
			wantKind: condScaffoldUpkeepPhase,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cond := &gameast.Condition{Kind: tc.kind, Args: []interface{}{tc.text}}
			got := detectConditionScaffold(cond)
			if got.kind != tc.wantKind {
				t.Errorf("kind: want %v, got %v", tc.wantKind, got.kind)
			}
			if tc.wantThr != 0 && got.threshold != tc.wantThr {
				t.Errorf("threshold: want %d, got %d", tc.wantThr, got.threshold)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Apply tests for new condScaffold kinds.
// ---------------------------------------------------------------------------

func TestApplyConditionScaffolding_GainedLifeThisTurn(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "intervening_if",
		Args: []interface{}{"you gained life this turn"},
	}
	cs := applyConditionScaffolding(gs, cond, nil)
	if cs.kind != condScaffoldGainedLifeThisTurn {
		t.Fatalf("expected GainedLifeThisTurn, got %v", cs.kind)
	}
	if gs.Seats[0].Flags["life_gained_this_turn"] <= 0 {
		t.Errorf("life_gained_this_turn flag not set: %v", gs.Seats[0].Flags)
	}
	if gs.Seats[0].Life <= 20 {
		t.Errorf("seat0 life should have increased, got %d", gs.Seats[0].Life)
	}
}

func TestApplyConditionScaffolding_CastSpellThisTurn(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "intervening_if",
		Args: []interface{}{"you cast a spell this turn"},
	}
	cs := applyConditionScaffolding(gs, cond, nil)
	if cs.kind != condScaffoldCastSpellThisTurn {
		t.Fatalf("expected CastSpellThisTurn, got %v", cs.kind)
	}
	if gs.Seats[0].SpellsCastThisTurn < 1 {
		t.Errorf("SpellsCastThisTurn not incremented: %d", gs.Seats[0].SpellsCastThisTurn)
	}
	if gs.SpellsCastThisTurn < 1 {
		t.Errorf("global SpellsCastThisTurn not incremented: %d", gs.SpellsCastThisTurn)
	}
	if gs.Seats[0].Flags["cast_spell_this_turn"] != 1 {
		t.Errorf("cast_spell_this_turn flag not set")
	}
}

func TestApplyConditionScaffolding_CreatureETBThisTurn(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "intervening_if",
		Args: []interface{}{"a creature entered the battlefield under your control this turn"},
	}
	cs := applyConditionScaffolding(gs, cond, nil)
	if cs.kind != condScaffoldCreatureETBThisTurn {
		t.Fatalf("expected CreatureETBThisTurn, got %v", cs.kind)
	}
	if gs.Flags["creature_etb_this_turn"] != 1 {
		t.Errorf("creature_etb_this_turn flag not set")
	}
	found := false
	for _, p := range gs.Seats[0].Battlefield {
		if p != nil && p.Card != nil && p.Card.Name == "ETB Witness" {
			found = true
		}
	}
	if !found {
		t.Errorf("ETB Witness creature not placed")
	}
}

func TestApplyConditionScaffolding_DrawnCardThisTurn(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "raw",
		Args: []interface{}{"if you've drawn a card this turn"},
	}
	cs := applyConditionScaffolding(gs, cond, nil)
	if cs.kind != condScaffoldDrawnCardThisTurn {
		t.Fatalf("expected DrawnCardThisTurn, got %v", cs.kind)
	}
	if gs.Seats[0].Flags["drawn_card_this_turn"] != 1 {
		t.Errorf("drawn_card_this_turn flag not set")
	}
	if len(gs.Seats[0].Library) < 5 {
		t.Errorf("expected library to have >=5 cards, got %d", len(gs.Seats[0].Library))
	}
}

func TestApplyConditionScaffolding_AttackedThisTurn(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "intervening_if",
		Args: []interface{}{"if you attacked this turn"},
	}
	cs := applyConditionScaffolding(gs, cond, nil)
	if cs.kind != condScaffoldAttackedThisTurn {
		t.Fatalf("expected AttackedThisTurn, got %v", cs.kind)
	}
	if gs.Flags["attacked_this_turn"] != 1 {
		t.Errorf("game attacked_this_turn flag not set")
	}
	if gs.Seats[0].Flags["attacked_this_turn"] != 1 {
		t.Errorf("seat 0 attacked_this_turn flag not set")
	}
}

func TestApplyConditionScaffolding_SacrificedThisTurn(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "intervening_if",
		Args: []interface{}{"if you sacrificed a creature this turn"},
	}
	cs := applyConditionScaffolding(gs, cond, nil)
	if cs.kind != condScaffoldSacrificedThisTurn {
		t.Fatalf("expected SacrificedThisTurn, got %v", cs.kind)
	}
	if gs.Flags["sacrificed_this_turn"] != 1 {
		t.Errorf("sacrificed_this_turn flag not set")
	}
	creatures := 0
	for _, c := range gs.Seats[0].Graveyard {
		for _, ty := range c.Types {
			if ty == "creature" {
				creatures++
				break
			}
		}
	}
	if creatures < 1 {
		t.Errorf("expected creature in graveyard")
	}
}

func TestApplyConditionScaffolding_CombatDamageDealt(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "intervening_if",
		Args: []interface{}{"a creature dealt combat damage to a player this turn"},
	}
	cs := applyConditionScaffolding(gs, cond, nil)
	if cs.kind != condScaffoldCombatDamageDealt {
		t.Fatalf("expected CombatDamageDealt, got %v", cs.kind)
	}
	if gs.Flags["combat_damage_dealt_this_turn"] != 1 {
		t.Errorf("combat_damage_dealt_this_turn flag not set")
	}
}

func TestApplyConditionScaffolding_LandfallThisTurn(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "intervening_if",
		Args: []interface{}{"landfall — if a land entered the battlefield"},
	}
	cs := applyConditionScaffolding(gs, cond, nil)
	if cs.kind != condScaffoldLandfallThisTurn {
		t.Fatalf("expected LandfallThisTurn, got %v", cs.kind)
	}
	if gs.Flags["landfall_this_turn"] != 1 {
		t.Errorf("landfall_this_turn flag not set")
	}
	foundLand := false
	for _, p := range gs.Seats[0].Battlefield {
		if p != nil && p.Card != nil {
			for _, ty := range p.Card.Types {
				if ty == "land" {
					foundLand = true
				}
			}
		}
	}
	if !foundLand {
		t.Errorf("no land placed on seat 0 battlefield")
	}
}

func TestApplyConditionScaffolding_DiscardedThisTurn(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "intervening_if",
		Args: []interface{}{"if you discarded a card this turn"},
	}
	cs := applyConditionScaffolding(gs, cond, nil)
	if cs.kind != condScaffoldDiscardedThisTurn {
		t.Fatalf("expected DiscardedThisTurn, got %v", cs.kind)
	}
	if gs.Seats[0].Flags["discarded_this_turn"] != 1 {
		t.Errorf("discarded_this_turn flag not set")
	}
	if len(gs.Seats[0].Graveyard) < 1 {
		t.Errorf("expected card in graveyard")
	}
	if len(gs.Seats[0].Hand) < 3 {
		t.Errorf("expected hand to have >=3 cards, got %d", len(gs.Seats[0].Hand))
	}
}

func TestApplyConditionScaffolding_EnchantedCreature(t *testing.T) {
	gs := newTestGameState(2)
	src := &gameengine.Permanent{
		Card:       &gameengine.Card{Name: "Ethereal Armor", Owner: 0, Types: []string{"enchantment", "aura"}},
		Controller: 0,
		Owner:      0,
		Flags:      map[string]int{},
	}
	cond := &gameast.Condition{
		Kind: "raw",
		Args: []interface{}{"enchanted creature has flying"},
	}
	cs := applyConditionScaffolding(gs, cond, src)
	if cs.kind != condScaffoldEnchantedCreature {
		t.Fatalf("expected EnchantedCreature, got %v", cs.kind)
	}
	if src.AttachedTo == nil {
		t.Errorf("source permanent should be attached to a creature")
	}
	if src.AttachedTo != nil && !src.AttachedTo.IsCreature() {
		t.Errorf("attached target should be a creature")
	}
}

func TestApplyConditionScaffolding_OpponentLostLife(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "intervening_if",
		Args: []interface{}{"if an opponent lost life this turn"},
	}
	cs := applyConditionScaffolding(gs, cond, nil)
	if cs.kind != condScaffoldOpponentLostLife {
		t.Fatalf("expected OpponentLostLife, got %v", cs.kind)
	}
	if gs.Seats[1].Life >= 20 {
		t.Errorf("opponent life should be reduced, got %d", gs.Seats[1].Life)
	}
	if gs.Seats[1].Flags["life_lost_this_turn"] <= 0 {
		t.Errorf("opponent life_lost_this_turn flag not set")
	}
}

func TestApplyConditionScaffolding_LifeAboveThreshold(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "intervening_if",
		Args: []interface{}{"if you have 25 or more life"},
	}
	cs := applyConditionScaffolding(gs, cond, nil)
	if cs.kind != condScaffoldLifeAboveThreshold {
		t.Fatalf("expected LifeAboveThreshold, got %v", cs.kind)
	}
	if gs.Seats[0].Life < 25 {
		t.Errorf("seat 0 life should be >= 25, got %d", gs.Seats[0].Life)
	}
}

func TestApplyConditionScaffolding_LifeBelowThreshold(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "intervening_if",
		Args: []interface{}{"if you have 5 or less life"},
	}
	cs := applyConditionScaffolding(gs, cond, nil)
	if cs.kind != condScaffoldLifeBelowThreshold {
		t.Fatalf("expected LifeBelowThreshold, got %v", cs.kind)
	}
	if gs.Seats[0].Life > 5 {
		t.Errorf("seat 0 life should be <= 5, got %d", gs.Seats[0].Life)
	}
}

func TestApplyConditionScaffolding_UpkeepPhase(t *testing.T) {
	gs := newTestGameState(2)
	cond := &gameast.Condition{
		Kind: "raw",
		Args: []interface{}{"during your upkeep, you may pay 2"},
	}
	cs := applyConditionScaffolding(gs, cond, nil)
	if cs.kind != condScaffoldUpkeepPhase {
		t.Fatalf("expected UpkeepPhase, got %v", cs.kind)
	}
	if gs.Phase != "beginning" || gs.Step != "upkeep" {
		t.Errorf("expected phase=beginning step=upkeep, got phase=%s step=%s", gs.Phase, gs.Step)
	}
}

// ---------------------------------------------------------------------------
// Verify classifyTrigger returns expected slugs for all known trigger events.
// ---------------------------------------------------------------------------

func TestClassifyTrigger_AllKnownEvents(t *testing.T) {
	cases := []struct {
		name     string
		trigger  *gameast.Trigger
		wantSlug string
	}{
		{
			name:     "creature dies",
			trigger:  &gameast.Trigger{Event: "dies"},
			wantSlug: "creature_dies",
		},
		{
			name:     "creature ETB",
			trigger:  &gameast.Trigger{Event: "etb"},
			wantSlug: "creature_etb",
		},
		{
			name:     "creature enters",
			trigger:  &gameast.Trigger{Event: "enters the battlefield"},
			wantSlug: "creature_etb",
		},
		{
			name:     "attacks",
			trigger:  &gameast.Trigger{Event: "attacks"},
			wantSlug: "attacks",
		},
		{
			name:     "combat damage",
			trigger:  &gameast.Trigger{Event: "deal_combat_damage"},
			wantSlug: "combat_damage",
		},
		{
			name:     "cast spell",
			trigger:  &gameast.Trigger{Event: "cast a spell"},
			wantSlug: "cast_spell",
		},
		{
			name:     "opponent cast",
			trigger:  &gameast.Trigger{Event: "cast a spell", Actor: &gameast.Filter{Base: "an opponent"}},
			wantSlug: "opponent_cast",
		},
		{
			name:     "gain life",
			trigger:  &gameast.Trigger{Event: "gain life"},
			wantSlug: "gain_life",
		},
		{
			name:     "draw card",
			trigger:  &gameast.Trigger{Event: "draw a card"},
			wantSlug: "draw_card",
		},
		{
			name:     "discard",
			trigger:  &gameast.Trigger{Event: "discard a card"},
			wantSlug: "discard",
		},
		{
			name:     "sacrifice",
			trigger:  &gameast.Trigger{Event: "sacrifice a creature"},
			wantSlug: "sacrifice",
		},
		{
			name:     "upkeep",
			trigger:  &gameast.Trigger{Event: "phase", Phase: "upkeep"},
			wantSlug: "upkeep",
		},
		{
			name:     "end step",
			trigger:  &gameast.Trigger{Event: "phase", Phase: "end_step"},
			wantSlug: "end_step",
		},
		{
			name:     "landfall enters",
			trigger:  &gameast.Trigger{Event: "a land enters"},
			wantSlug: "creature_etb",
		},
		{
			name:     "nil trigger",
			trigger:  nil,
			wantSlug: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyTrigger(tc.trigger)
			if got != tc.wantSlug {
				t.Errorf("classifyTrigger: want %q, got %q", tc.wantSlug, got)
			}
		})
	}
}

func newTestGameState(seats int) *gameengine.GameState {
	gs := &gameengine.GameState{
		Turn:   1,
		Active: 0,
		Phase:  "precombat_main",
		Flags:  map[string]int{},
	}
	for i := 0; i < seats; i++ {
		gs.Seats = append(gs.Seats, &gameengine.Seat{
			Life:  20,
			Flags: map[string]int{},
		})
	}
	return gs
}
