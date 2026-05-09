package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// Backfill: per-hook behavior tests for custom handlers whose ETB / Activated
// / Trigger surface had at least one hook with no test. The function-level
// "AllRegistered" smoke checks already cover registration; these add the
// missing assertions about side-effects.

// ---------------------------------------------------------------------------
// Chainer, Dementia Master — ETB partials (LTB exile + nightmare anthem)
// ---------------------------------------------------------------------------

func TestChainer_ETBEmitsLTBAndAnthemPartials(t *testing.T) {
	gs := newGame(t, 2)
	chainer := addPerm(gs, 0, "Chainer, Dementia Master", "creature")
	gameengine.InvokeETBHook(gs, chainer)

	if hasEvent(gs, "per_card_partial") < 2 {
		t.Fatalf("expected ≥2 per_card_partial events for Chainer ETB; got %d",
			hasEvent(gs, "per_card_partial"))
	}
}

// ---------------------------------------------------------------------------
// Charix, the Raging Isle — ETB partial (target tax)
// ---------------------------------------------------------------------------

func TestCharix_ETBEmitsTargetTaxPartial(t *testing.T) {
	gs := newGame(t, 2)
	charix := addPerm(gs, 0, "Charix, the Raging Isle", "creature")
	gameengine.InvokeETBHook(gs, charix)

	if hasEvent(gs, "per_card_partial") < 1 {
		t.Fatalf("expected per_card_partial for Charix target-tax ETB; got %d",
			hasEvent(gs, "per_card_partial"))
	}
}

// ---------------------------------------------------------------------------
// Derevi, Empyrial Tactician — combat-damage trigger toggles tap state
// ---------------------------------------------------------------------------

func TestDerevi_CombatDamageTrigger_TapsBiggestOpponentCreature(t *testing.T) {
	gs := newGame(t, 2)
	derevi := addPerm(gs, 0, "Derevi, Empyrial Tactician", "creature")
	attacker := addPerm(gs, 0, "Llanowar Elves", "creature")
	attacker.Card.BasePower = 1

	smallOpp := addPerm(gs, 1, "Squire", "creature")
	smallOpp.Card.BasePower = 1
	bigOpp := addPerm(gs, 1, "Hill Giant", "creature")
	bigOpp.Card.BasePower = 3

	dereviCombatDamageTapOrUntap(gs, derevi, map[string]interface{}{
		"attacker_perm": attacker,
		"defender_seat": 1,
		"amount":        1,
	})

	if !bigOpp.Tapped {
		t.Fatalf("Derevi should tap the highest-power opponent creature on combat trigger")
	}
	if smallOpp.Tapped {
		t.Fatalf("Derevi should not tap the smaller opponent creature when a bigger one is available")
	}
}

func TestDerevi_CombatDamageTrigger_IgnoresOpponentAttacker(t *testing.T) {
	gs := newGame(t, 2)
	derevi := addPerm(gs, 0, "Derevi, Empyrial Tactician", "creature")
	oppAttacker := addPerm(gs, 1, "Hill Giant", "creature")
	oppAttacker.Card.BasePower = 3
	myCreature := addPerm(gs, 0, "Grizzly Bears", "creature")
	myCreature.Card.BasePower = 2

	dereviCombatDamageTapOrUntap(gs, derevi, map[string]interface{}{
		"attacker_perm": oppAttacker,
		"defender_seat": 0,
		"amount":        3,
	})

	// Trigger only fires for our own attackers — nothing should change.
	if myCreature.Tapped {
		t.Fatalf("Derevi should not act on opponent attackers")
	}
}

// ---------------------------------------------------------------------------
// Drivnod, Carnage Dominus — activated indestructible counters
// ---------------------------------------------------------------------------

func TestDrivnod_ActivationExilesTwoAndAddsIndestructible(t *testing.T) {
	gs := newGame(t, 2)
	drivnod := addPerm(gs, 0, "Drivnod, Carnage Dominus", "creature")
	addPerm(gs, 0, "Sac Fodder A", "creature")
	addPerm(gs, 0, "Sac Fodder B", "creature")
	gs.Seats[0].ManaPool = 4

	drivnodIndestructibleActivate(gs, drivnod, 0, nil)

	if drivnod.Counters["indestructible"] != 2 {
		t.Fatalf("expected 2 indestructible counters; got %d", drivnod.Counters["indestructible"])
	}
	if gs.Seats[0].ManaPool != 0 {
		t.Fatalf("expected 4 mana spent; pool=%d", gs.Seats[0].ManaPool)
	}
	if len(gs.Seats[0].Exile) != 2 {
		t.Fatalf("expected 2 cards exiled; got %d", len(gs.Seats[0].Exile))
	}
}

func TestDrivnod_ActivationFailsWithoutTwoOtherCreatures(t *testing.T) {
	gs := newGame(t, 2)
	drivnod := addPerm(gs, 0, "Drivnod, Carnage Dominus", "creature")
	addPerm(gs, 0, "Lone Bear", "creature") // only one fodder
	gs.Seats[0].ManaPool = 4

	drivnodIndestructibleActivate(gs, drivnod, 0, nil)

	if drivnod.Counters["indestructible"] != 0 {
		t.Fatalf("should not gain counters without two fodder creatures")
	}
	if hasEvent(gs, "per_card_failed") < 1 {
		t.Fatalf("expected per_card_failed event")
	}
}

// ---------------------------------------------------------------------------
// Kalamax, the Stormsire — ETB resets first-instant flag
// ---------------------------------------------------------------------------

func TestKalamax_ETBResetsFirstInstantFlag(t *testing.T) {
	gs := newGame(t, 2)
	kalamax := addPerm(gs, 0, "Kalamax, the Stormsire", "creature")
	kalamax.Flags["kalamax_first_instant_used"] = 1 // pretend stale state

	kalamaxETBReset(gs, kalamax)

	if kalamax.Flags["kalamax_first_instant_used"] != 0 {
		t.Fatalf("ETB should clear kalamax_first_instant_used flag; got %d",
			kalamax.Flags["kalamax_first_instant_used"])
	}
	if len(gs.DelayedTriggers) < 1 {
		t.Fatalf("ETB should register the upkeep reset delayed trigger")
	}
}

// ---------------------------------------------------------------------------
// Karador, Ghost Chieftain — ETB initializes once-per-turn flag
// ---------------------------------------------------------------------------

func TestKarador_ETBInitializesUsedFlag(t *testing.T) {
	gs := newGame(t, 2)
	karador := addPerm(gs, 0, "Karador, Ghost Chieftain", "creature")
	karador.Flags["karador_used_this_turn"] = 1 // stale

	karadorETB(gs, karador)

	if karador.Flags["karador_used_this_turn"] != 0 {
		t.Fatalf("ETB should reset karador_used_this_turn flag")
	}
	if len(gs.DelayedTriggers) < 1 {
		t.Fatalf("ETB should register an upkeep reset delayed trigger")
	}
}

// ---------------------------------------------------------------------------
// Kardur, Doomscourge — ETB sets the goad flag
// ---------------------------------------------------------------------------

func TestKardur_ETBSetsGoadFlag(t *testing.T) {
	gs := newGame(t, 2)
	kardur := addPerm(gs, 0, "Kardur, Doomscourge", "creature")

	kardurETBGoadFlag(gs, kardur)

	// Encoded as controller+1 so 0 means "off".
	if gs.Flags["kardur_goad_seat"] != 1 {
		t.Fatalf("expected kardur_goad_seat = controller+1 = 1; got %d",
			gs.Flags["kardur_goad_seat"])
	}
	if hasEvent(gs, "per_card_partial") < 1 {
		t.Fatalf("expected partial breadcrumb for engine-side goad enforcement")
	}
}

// ---------------------------------------------------------------------------
// Solphim, Mayhem Dominus — activated indestructible counters
// ---------------------------------------------------------------------------

func TestSolphim_ActivationExilesTwoAndAddsIndestructible(t *testing.T) {
	gs := newGame(t, 2)
	solphim := addPerm(gs, 0, "Solphim, Mayhem Dominus", "creature")
	addPerm(gs, 0, "Sac Fodder A", "creature")
	addPerm(gs, 0, "Sac Fodder B", "creature")
	gs.Seats[0].ManaPool = 4

	solphimIndestructibleActivate(gs, solphim, 0, nil)

	if solphim.Counters["indestructible"] != 2 {
		t.Fatalf("expected 2 indestructible counters; got %d", solphim.Counters["indestructible"])
	}
	if len(gs.Seats[0].Exile) != 2 {
		t.Fatalf("expected 2 cards exiled; got %d", len(gs.Seats[0].Exile))
	}
}

// ---------------------------------------------------------------------------
// Yurlok of Scorch Thrash — ETB sets the pain-mana static flag
// ---------------------------------------------------------------------------

func TestYurlok_ETBSetsPainManaFlag(t *testing.T) {
	gs := newGame(t, 2)
	yurlok := addPerm(gs, 0, "Yurlok of Scorch Thrash", "creature")

	gameengine.InvokeETBHook(gs, yurlok)

	if gs.Flags["yurlok_pain_mana_active"] != 1 {
		t.Fatalf("expected yurlok_pain_mana_active flag to be set; got %d",
			gs.Flags["yurlok_pain_mana_active"])
	}
	if hasEvent(gs, "per_card_partial") < 1 {
		t.Fatalf("expected partial breadcrumb for engine-side pain-mana hook")
	}
}

// ---------------------------------------------------------------------------
// Zopandrel, Hunger Dominus — activated indestructible counters
// ---------------------------------------------------------------------------

func TestZopandrel_ActivationExilesTwoAndAddsIndestructible(t *testing.T) {
	gs := newGame(t, 2)
	zopandrel := addPerm(gs, 0, "Zopandrel, Hunger Dominus", "creature")
	addPerm(gs, 0, "Sac Fodder A", "creature")
	addPerm(gs, 0, "Sac Fodder B", "creature")
	gs.Seats[0].ManaPool = 4

	zopandrelIndestructibleActivate(gs, zopandrel, 0, nil)

	if zopandrel.Counters["indestructible"] != 2 {
		t.Fatalf("expected 2 indestructible counters; got %d", zopandrel.Counters["indestructible"])
	}
	if len(gs.Seats[0].Exile) != 2 {
		t.Fatalf("expected 2 cards exiled; got %d", len(gs.Seats[0].Exile))
	}
}

// ---------------------------------------------------------------------------
// Felothar the Steadfast (era3) — ETB stamps static markers on creatures
// ---------------------------------------------------------------------------

func TestFelotharEra3_ETBStampsDefenderAndDamageMarkers(t *testing.T) {
	gs := newGame(t, 2)
	felo := addPerm(gs, 0, "Felothar the Steadfast", "creature")
	wall := addPerm(gs, 0, "Wall of Omens", "creature")
	bear := addPerm(gs, 0, "Grizzly Bears", "creature")

	felotharEra3MarkDefenders(gs, felo)

	for _, p := range []*gameengine.Permanent{felo, wall, bear} {
		if p.Flags["felothar_attack_with_defender"] != 1 {
			t.Fatalf("%s missing felothar_attack_with_defender flag", p.Card.DisplayName())
		}
		if p.Flags["felothar_damage_by_toughness"] != 1 {
			t.Fatalf("%s missing felothar_damage_by_toughness flag", p.Card.DisplayName())
		}
	}
}

func TestFelotharEra3_PermanentETBTriggerStampsNewArrival(t *testing.T) {
	gs := newGame(t, 2)
	felo := addPerm(gs, 0, "Felothar the Steadfast", "creature")
	// New creature ETBs after Felothar — trigger fires.
	newcomer := addPerm(gs, 0, "Llanowar Elves", "creature")

	felotharEra3MarkDefendersTrigger(gs, felo, map[string]interface{}{
		"perm": newcomer,
	})

	if newcomer.Flags["felothar_attack_with_defender"] != 1 {
		t.Fatalf("new creature should pick up felothar_attack_with_defender flag")
	}
}
