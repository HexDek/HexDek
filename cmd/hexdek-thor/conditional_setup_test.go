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
