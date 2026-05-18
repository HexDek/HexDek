package gameengine

// R38 layer-system audit fixes — tests for:
//   (a) Layer 7a CDA sub-layer ordering vs. 7b/7c/7d
//   (b) Replacement ETB ordering: category + APNAP + timestamp
//   (c) Prevention interacting with damage redirection
//
// Rule citations:
//   §613.4a — char-defining P/T applies BEFORE 7b set-P/T
//   §613.4b — set-P/T overrides 7a values
//   §613.4c — modifications / counters layer on top of 7a/7b
//   §613.4d — switch power/toughness applies LAST in sublayer order
//   §616.1  — replacement category order: self < control < copy < back < other
//   §615    — prevention applies to the (post-redirection) target

import (
	"testing"
)

func TestSubLayer_7aCDA_Alone(t *testing.T) {
	gs := newFixtureGame(t)
	goyf := addBattlefield(gs, 0, "Tarmogoyf", 0, 1, "creature")
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard,
		&Card{Name: "Lightning Bolt", Types: []string{"instant"}},
		&Card{Name: "Forest", Types: []string{"land"}},
	)
	gs.Seats[1].Graveyard = append(gs.Seats[1].Graveyard,
		&Card{Name: "Elvish Mystic", Types: []string{"creature"}},
	)
	RegisterTarmogoyf(gs, goyf)
	chars := GetEffectiveCharacteristics(gs, goyf)
	if chars.Power != 3 || chars.Toughness != 4 {
		t.Fatalf("Tarmogoyf CDA: expected 3/4, got %d/%d", chars.Power, chars.Toughness)
	}
}

func TestSubLayer_7a_Then_7b(t *testing.T) {
	gs := newFixtureGame(t)
	goyf := addBattlefield(gs, 0, "Tarmogoyf", 0, 1, "creature")
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard,
		&Card{Name: "Bolt", Types: []string{"instant"}},
		&Card{Name: "Forest", Types: []string{"land"}},
		&Card{Name: "Bear", Types: []string{"creature"}},
	)
	RegisterTarmogoyf(gs, goyf)
	_ = layerAt(gs, 1, "Humility", 0, 0, "enchantment")
	chars := GetEffectiveCharacteristics(gs, goyf)
	if chars.Power != 1 || chars.Toughness != 1 {
		t.Fatalf("7b should override 7a: expected 1/1, got %d/%d", chars.Power, chars.Toughness)
	}
}

func TestSubLayer_7a_Then_7c(t *testing.T) {
	gs := newFixtureGame(t)
	goyf := addBattlefield(gs, 0, "Tarmogoyf", 0, 1, "creature")
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard,
		&Card{Name: "Bolt", Types: []string{"instant"}},
		&Card{Name: "Forest", Types: []string{"land"}},
	)
	RegisterTarmogoyf(gs, goyf)
	anthem := addBattlefield(gs, 0, "Crusade", 0, 0, "enchantment")
	registerAnthemPT(gs, anthem, 1, 1, "test_anthem",
		func(_ *GameState, t *Permanent) bool { return charsHaveType(t.Card.Types, "creature") })
	chars := GetEffectiveCharacteristics(gs, goyf)
	if chars.Power != 3 || chars.Toughness != 4 {
		t.Fatalf("7a+7c: expected 3/4, got %d/%d", chars.Power, chars.Toughness)
	}
}

func TestSubLayer_7a_Then_7b_Then_7c(t *testing.T) {
	gs := newFixtureGame(t)
	goyf := addBattlefield(gs, 0, "Tarmogoyf", 0, 1, "creature")
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard,
		&Card{Name: "Bolt", Types: []string{"instant"}},
		&Card{Name: "Forest", Types: []string{"land"}},
		&Card{Name: "Bear", Types: []string{"creature"}},
		&Card{Name: "Ritual", Types: []string{"sorcery"}},
	)
	RegisterTarmogoyf(gs, goyf)
	_ = layerAt(gs, 1, "Humility", 0, 0, "enchantment")
	anthem := addBattlefield(gs, 0, "Glorious Anthem", 0, 0, "enchantment")
	registerAnthemPT(gs, anthem, 1, 1, "test_anthem",
		func(_ *GameState, t *Permanent) bool { return charsHaveType(t.Card.Types, "creature") })
	chars := GetEffectiveCharacteristics(gs, goyf)
	if chars.Power != 2 || chars.Toughness != 2 {
		t.Fatalf("7a→7b→7c chain: expected 2/2, got %d/%d", chars.Power, chars.Toughness)
	}
}

func TestSubLayer_7a_Then_7d(t *testing.T) {
	gs := newFixtureGame(t)
	goyf := addBattlefield(gs, 0, "Tarmogoyf", 0, 1, "creature")
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard,
		&Card{Name: "Bolt", Types: []string{"instant"}},
		&Card{Name: "Forest", Types: []string{"land"}},
	)
	RegisterTarmogoyf(gs, goyf)
	doran := addBattlefield(gs, 0, "Doran, the Siege Tower", 0, 5, "creature")
	RegisterDoranSiegeTower(gs, doran)
	chars := GetEffectiveCharacteristics(gs, goyf)
	if chars.Power != 3 || chars.Toughness != 2 {
		t.Fatalf("7a→7d switch: expected 3/2, got %d/%d", chars.Power, chars.Toughness)
	}
}

func TestSubLayer_FullChain_7a_7b_7c_7d(t *testing.T) {
	gs := newFixtureGame(t)
	goyf := addBattlefield(gs, 0, "Tarmogoyf", 0, 1, "creature")
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard,
		&Card{Name: "Bolt", Types: []string{"instant"}},
		&Card{Name: "Forest", Types: []string{"land"}},
		&Card{Name: "Bear", Types: []string{"creature"}},
	)
	RegisterTarmogoyf(gs, goyf)
	_ = layerAt(gs, 1, "Humility", 0, 0, "enchantment")
	anthem := addBattlefield(gs, 0, "Bloodlust", 0, 0, "enchantment")
	registerAnthemPT(gs, anthem, 2, 0, "test_anthem_asym",
		func(_ *GameState, t *Permanent) bool { return charsHaveType(t.Card.Types, "creature") })
	doran := addBattlefield(gs, 0, "Doran", 0, 5, "creature")
	RegisterDoranSiegeTower(gs, doran)
	chars := GetEffectiveCharacteristics(gs, goyf)
	if chars.Power != 1 || chars.Toughness != 3 {
		t.Fatalf("full 7a→7b→7c→7d chain: expected 1/3, got %d/%d",
			chars.Power, chars.Toughness)
	}
}

func TestSubLayer_7a_Then_Counter(t *testing.T) {
	gs := newFixtureGame(t)
	goyf := addBattlefield(gs, 0, "Tarmogoyf", 0, 1, "creature")
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard,
		&Card{Name: "Bolt", Types: []string{"instant"}},
	)
	RegisterTarmogoyf(gs, goyf)
	goyf.Counters["+1/+1"] = 2
	chars := GetEffectiveCharacteristics(gs, goyf)
	if chars.Power != 3 || chars.Toughness != 4 {
		t.Fatalf("7a + counters: expected 3/4, got %d/%d", chars.Power, chars.Toughness)
	}
}

func TestReplOrder_SelfReplacementBeforeOther(t *testing.T) {
	gs := newFixtureGame(t)
	src := addBattlefield(gs, 0, "Source", 0, 0, "creature")
	applied := []string{}
	gs.RegisterReplacement(&ReplacementEffect{
		EventType: "would_put_counter", HandlerID: "doubler",
		ControllerSeat: 0, Timestamp: 1, Category: CategoryOther,
		Applies: func(_ *GameState, _ *ReplEvent) bool { return true },
		ApplyFn: func(_ *GameState, ev *ReplEvent) {
			applied = append(applied, "doubler"); ev.SetCount(ev.Count() * 2)
		},
	})
	gs.RegisterReplacement(&ReplacementEffect{
		EventType: "would_put_counter", HandlerID: "self_rep",
		ControllerSeat: 0, Timestamp: 2, Category: CategorySelfReplacement,
		Applies: func(_ *GameState, ev *ReplEvent) bool { return ev.Count() == 0 },
		ApplyFn: func(_ *GameState, ev *ReplEvent) {
			applied = append(applied, "self_rep"); ev.SetCount(1)
		},
	})
	count, _ := FirePutCounterEvent(gs, src, "+1/+1", 0, src)
	if len(applied) != 2 || applied[0] != "self_rep" || applied[1] != "doubler" {
		t.Fatalf("§616.1 order: expected [self_rep, doubler], got %v", applied)
	}
	if count != 2 {
		t.Fatalf("expected count=2, got %d", count)
	}
}

func TestReplOrder_APNAP_WithinCategory(t *testing.T) {
	gs := newFixtureGame(t)
	gs.Active = 0
	src := addBattlefield(gs, 0, "Source", 0, 0, "creature")
	applied := []string{}
	gs.RegisterReplacement(&ReplacementEffect{
		EventType: "would_put_counter", HandlerID: "opp",
		ControllerSeat: 1, Timestamp: 1, Category: CategoryOther,
		Applies: func(_ *GameState, _ *ReplEvent) bool { return true },
		ApplyFn: func(_ *GameState, ev *ReplEvent) {
			applied = append(applied, "opp"); ev.SetCount(ev.Count() + 10)
		},
	})
	gs.RegisterReplacement(&ReplacementEffect{
		EventType: "would_put_counter", HandlerID: "active",
		ControllerSeat: 0, Timestamp: 2, Category: CategoryOther,
		Applies: func(_ *GameState, _ *ReplEvent) bool { return true },
		ApplyFn: func(_ *GameState, ev *ReplEvent) {
			applied = append(applied, "active"); ev.SetCount(ev.Count() * 2)
		},
	})
	_, _ = FirePutCounterEvent(gs, src, "+1/+1", 1, src)
	if len(applied) != 2 || applied[0] != "active" {
		t.Fatalf("APNAP: active should apply first, got %v", applied)
	}
}

func TestReplOrder_ETBDoubler_RunsOncePerEvent(t *testing.T) {
	gs := newFixtureGame(t)
	src := addBattlefield(gs, 0, "Panharmonicon", 0, 0, "artifact")
	RegisterPanharmonicon(gs, src)
	count, _ := FireETBTriggerEvent(gs, src)
	if count != 2 {
		t.Fatalf("Panharmonicon should double to 2, got %d", count)
	}
	count2, _ := FireETBTriggerEvent(gs, src)
	if count2 != 2 {
		t.Fatalf("second event should still double to 2, got %d", count2)
	}
}

func TestRedirect_BasicRedirectsToOtherPlayer(t *testing.T) {
	gs := newFixtureGame(t)
	src := addBattlefield(gs, 0, "Bolt Source", 0, 0, "creature")
	RegisterDamageRedirectPlayer(gs, src, 0, 1, "test_redirect")
	s0, s1 := gs.Seats[0].Life, gs.Seats[1].Life
	applyDamage(gs, src, Target{Kind: TargetKindSeat, Seat: 0}, 3)
	if gs.Seats[0].Life != s0 {
		t.Fatalf("seat 0: expected 0 damage, got %d", s0-gs.Seats[0].Life)
	}
	if gs.Seats[1].Life != s1-3 {
		t.Fatalf("seat 1: expected 3 redirected, got %d", s1-gs.Seats[1].Life)
	}
}

func TestRedirect_PreventionAppliesAtRedirectedTarget(t *testing.T) {
	gs := newFixtureGame(t)
	src := addBattlefield(gs, 0, "Bolt Source", 0, 0, "creature")
	RegisterDamageRedirectPlayer(gs, src, 0, 1, "redirect_to_1")
	AddPreventionShield(gs, PreventionShield{TargetSeat: 1, Amount: 5, OneShot: true, SourceCard: "Healing Salve"})
	s0, s1 := gs.Seats[0].Life, gs.Seats[1].Life
	applyDamage(gs, src, Target{Kind: TargetKindSeat, Seat: 0}, 3)
	if gs.Seats[0].Life != s0 {
		t.Fatalf("seat 0: redirected, expected 0, got %d", s0-gs.Seats[0].Life)
	}
	if gs.Seats[1].Life != s1 {
		t.Fatalf("seat 1: shield should prevent all, got %d", s1-gs.Seats[1].Life)
	}
}

func TestRedirect_PreventionOnOriginalTargetDoesNotApply(t *testing.T) {
	gs := newFixtureGame(t)
	src := addBattlefield(gs, 0, "Bolt Source", 0, 0, "creature")
	RegisterDamageRedirectPlayer(gs, src, 0, 1, "redirect_to_1")
	AddPreventionShield(gs, PreventionShield{TargetSeat: 0, Amount: 5, OneShot: true, SourceCard: "Misplaced Shield"})
	before := len(gs.PreventionShields)
	amtBefore := gs.PreventionShields[0].Amount
	applyDamage(gs, src, Target{Kind: TargetKindSeat, Seat: 0}, 3)
	if len(gs.PreventionShields) != before {
		t.Fatalf("shield on orig target should NOT be consumed; %d → %d", before, len(gs.PreventionShields))
	}
	if gs.PreventionShields[0].Amount != amtBefore {
		t.Fatalf("shield amount unchanged; %d → %d", amtBefore, gs.PreventionShields[0].Amount)
	}
}

func TestRedirect_NoRedirectKeepsPreventionPath(t *testing.T) {
	gs := newFixtureGame(t)
	src := addBattlefield(gs, 0, "Bolt Source", 0, 0, "creature")
	AddPreventionShield(gs, PreventionShield{TargetSeat: 0, Amount: 5, OneShot: true, SourceCard: "Healing Salve"})
	start := gs.Seats[0].Life
	applyDamage(gs, src, Target{Kind: TargetKindSeat, Seat: 0}, 3)
	if gs.Seats[0].Life != start {
		t.Fatalf("shield should prevent all 3, got %d through", start-gs.Seats[0].Life)
	}
}
