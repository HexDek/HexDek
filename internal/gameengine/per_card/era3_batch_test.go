package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// ---------------------------------------------------------------------------
// Jetmir, Nexus of Revels — keyword tiers at 3 / 6 / 9 creatures
// ---------------------------------------------------------------------------

func TestJetmirEra3_GrantsKeywordsAtThreeCreatures(t *testing.T) {
	gs := newGame(t, 2)
	jetmir := addPerm(gs, 0, "Jetmir, Nexus of Revels", "creature")
	c1 := addPerm(gs, 0, "Grizzly Bears", "creature")
	c2 := addPerm(gs, 0, "Llanowar Elves", "creature")
	// Three creatures total (Jetmir + 2). Should grant vigilance.

	jetmirEra3Apply(gs, jetmir)

	if c1.Flags["kw:vigilance"] != 1 {
		t.Fatalf("expected vigilance on c1 at 3 creatures; flags=%+v", c1.Flags)
	}
	if c2.Flags["kw:vigilance"] != 1 {
		t.Fatalf("expected vigilance on c2 at 3 creatures")
	}
	if c1.Flags["kw:trample"] != 0 {
		t.Fatalf("should not have trample at 3 creatures")
	}
}

func TestJetmirEra3_DoubleStrikeAtNineCreatures(t *testing.T) {
	gs := newGame(t, 2)
	jetmir := addPerm(gs, 0, "Jetmir, Nexus of Revels", "creature")
	for i := 0; i < 8; i++ {
		addPerm(gs, 0, "Grizzly Bears", "creature")
	}
	// 9 creatures total.
	jetmirEra3Apply(gs, jetmir)

	first := gs.Seats[0].Battlefield[1]
	if first.Flags["kw:double strike"] != 1 {
		t.Fatalf("expected double strike at 9 creatures; flags=%+v", first.Flags)
	}
	if first.Flags["kw:trample"] != 1 || first.Flags["kw:vigilance"] != 1 {
		t.Fatalf("expected all three keywords at tier 9")
	}
}

// ---------------------------------------------------------------------------
// Falco Spara, Pactweaver
// ---------------------------------------------------------------------------

func TestFalcoSparaEra3_ETBPlacesThreeCounters(t *testing.T) {
	gs := newGame(t, 2)
	falco := addPerm(gs, 0, "Falco Spara, Pactweaver", "creature")
	falcoSparaEra3ETB(gs, falco)
	if falco.Counters["+1/+1"] != 3 {
		t.Fatalf("expected 3 +1/+1 counters; got %d", falco.Counters["+1/+1"])
	}
}

func TestFalcoSparaEra3_ActivatedRemovesCounterAndPlaysCreature(t *testing.T) {
	gs := newGame(t, 2)
	falco := addPerm(gs, 0, "Falco Spara, Pactweaver", "creature")
	falco.AddCounter("+1/+1", 3)
	gs.Seats[0].Life = 40

	// Top of library is a creature.
	creature := &gameengine.Card{Name: "Llanowar Elves", Owner: 0, Types: []string{"creature"}}
	gs.Seats[0].Library = append(gs.Seats[0].Library, creature)

	falcoSparaEra3Activate(gs, falco, 0, nil)

	if falco.Counters["+1/+1"] != 2 {
		t.Fatalf("expected counter to be removed; got %d", falco.Counters["+1/+1"])
	}
	if gs.Seats[0].Life != 39 {
		t.Fatalf("expected 1 life paid; got %d", gs.Seats[0].Life)
	}
	if len(gs.Seats[0].Hand) != 1 {
		t.Fatalf("expected creature pulled to hand; hand=%d", len(gs.Seats[0].Hand))
	}
}

// ---------------------------------------------------------------------------
// Lord Xander, the Collector
// ---------------------------------------------------------------------------

func TestLordXanderEra3_ETBDiscardsThree(t *testing.T) {
	gs := newGame(t, 2)
	xander := addPerm(gs, 0, "Lord Xander, the Collector", "creature")
	for _, n := range []string{"A", "B", "C", "D", "E"} {
		gs.Seats[1].Hand = append(gs.Seats[1].Hand, &gameengine.Card{Name: n, Owner: 1})
	}

	lordXanderEra3ETB(gs, xander)

	if len(gs.Seats[1].Hand) != 2 {
		t.Fatalf("expected 3 discarded (5→2); hand=%d", len(gs.Seats[1].Hand))
	}
}

func TestLordXanderEra3_AttackMillsHalfDefenderLibrary(t *testing.T) {
	gs := newGame(t, 2)
	xander := addPerm(gs, 0, "Lord Xander, the Collector", "creature")
	for i := 0; i < 10; i++ {
		gs.Seats[1].Library = append(gs.Seats[1].Library,
			&gameengine.Card{Name: "X", Owner: 1})
	}

	gameengine.SetAttackerDefender(xander, 1)
	lordXanderEra3Attack(gs, xander, map[string]interface{}{
		"attacker_perm": xander,
		"attacker_seat": 0,
	})

	if len(gs.Seats[1].Library) != 5 {
		t.Fatalf("expected half-library mill (10→5); got %d", len(gs.Seats[1].Library))
	}
	if len(gs.Seats[1].Graveyard) != 5 {
		t.Fatalf("expected 5 cards in graveyard; got %d", len(gs.Seats[1].Graveyard))
	}
}

// ---------------------------------------------------------------------------
// Hidetsugu and Kairi
// ---------------------------------------------------------------------------

func TestHidetsuguEra3_ETBPutsTwoBackOnTop(t *testing.T) {
	gs := newGame(t, 2)
	hk := addPerm(gs, 0, "Hidetsugu and Kairi", "creature")
	for _, n := range []string{"H1", "H2", "H3", "H4"} {
		gs.Seats[0].Hand = append(gs.Seats[0].Hand, &gameengine.Card{Name: n, Owner: 0})
	}
	libBefore := len(gs.Seats[0].Library)
	handBefore := len(gs.Seats[0].Hand)

	hidetsuguEra3ETB(gs, hk)

	if len(gs.Seats[0].Hand) != handBefore-2 {
		t.Fatalf("expected 2 cards moved out of hand; before=%d after=%d",
			handBefore, len(gs.Seats[0].Hand))
	}
	if len(gs.Seats[0].Library) != libBefore+2 {
		t.Fatalf("expected library to grow by 2; before=%d after=%d",
			libBefore, len(gs.Seats[0].Library))
	}
}

// ---------------------------------------------------------------------------
// Shorikai, Genesis Engine
// ---------------------------------------------------------------------------

func TestShorikaiEra3_ActivateDrawsTwoDiscardOneSpawnsPilot(t *testing.T) {
	gs := newGame(t, 2)
	shorikai := addPerm(gs, 0, "Shorikai, Genesis Engine", "artifact")
	addLibrary(gs, 0, "L1", "L2", "L3", "L4")
	bfBefore := len(gs.Seats[0].Battlefield)

	shorikaiEra3Activate(gs, shorikai, 0, nil)

	// Drew 2, discarded 1 → net +1 in hand.
	if len(gs.Seats[0].Hand) != 1 {
		t.Fatalf("expected 1 net card in hand after loot; hand=%d", len(gs.Seats[0].Hand))
	}
	// Pilot token added to battlefield.
	if len(gs.Seats[0].Battlefield) != bfBefore+1 {
		t.Fatalf("expected pilot token; battlefield grew by %d",
			len(gs.Seats[0].Battlefield)-bfBefore)
	}
}

// ---------------------------------------------------------------------------
// Acererak the Archlich
// ---------------------------------------------------------------------------

func TestAcererakEra3_AttackSpawnsZombiePerOpponent(t *testing.T) {
	gs := newGame(t, 4)
	acer := addPerm(gs, 0, "Acererak the Archlich", "creature")
	bfBefore := len(gs.Seats[0].Battlefield)

	acererakEra3Attack(gs, acer, map[string]interface{}{
		"attacker_perm": acer,
		"attacker_seat": 0,
	})

	// 3 opponents → 3 zombie tokens.
	if len(gs.Seats[0].Battlefield) != bfBefore+3 {
		t.Fatalf("expected 3 zombies; battlefield grew by %d",
			len(gs.Seats[0].Battlefield)-bfBefore)
	}
}

// ---------------------------------------------------------------------------
// Tazri, Beacon of Unity
// ---------------------------------------------------------------------------

func TestTazriEra3_ActivatedPullsPartyMembersFromTopSix(t *testing.T) {
	gs := newGame(t, 2)
	tazri := addPerm(gs, 0, "Tazri, Beacon of Unity", "creature")
	gs.Seats[0].Library = []*gameengine.Card{
		{Name: "Cleric of Life", Owner: 0, Types: []string{"creature", "cleric"}},
		{Name: "Forest", Owner: 0, Types: []string{"land"}},
		{Name: "Rogue Skull", Owner: 0, Types: []string{"creature", "rogue"}},
		{Name: "Sol Ring", Owner: 0, Types: []string{"artifact"}},
		{Name: "Warrior Blade", Owner: 0, Types: []string{"creature", "warrior"}},
		{Name: "Plains", Owner: 0, Types: []string{"land"}},
	}

	tazriEra3Activate(gs, tazri, 0, nil)

	// Should have pulled exactly 2 (max) of the 3 party creatures.
	if len(gs.Seats[0].Hand) != 2 {
		t.Fatalf("expected 2 cards picked; hand=%d", len(gs.Seats[0].Hand))
	}
}

// ---------------------------------------------------------------------------
// Sivriss, Nightmare Speaker
// ---------------------------------------------------------------------------

func TestSivrissEra3_OpponentsPayMillsStayInGraveyard(t *testing.T) {
	gs := newGame(t, 4)
	sivriss := addPerm(gs, 0, "Sivriss, Nightmare Speaker", "creature")
	addLibrary(gs, 0, "S1", "S2", "S3", "S4")
	for i := range gs.Seats {
		gs.Seats[i].Life = 40
	}

	sivrissEra3Activate(gs, sivriss, 0, nil)

	// 3 opps mill 3 cards into our graveyard, each opp pays 3 life.
	if len(gs.Seats[0].Graveyard) != 3 {
		t.Fatalf("expected 3 milled into graveyard; got %d", len(gs.Seats[0].Graveyard))
	}
	if len(gs.Seats[0].Hand) != 0 {
		t.Fatalf("expected nothing returned when opps pay; hand=%d", len(gs.Seats[0].Hand))
	}
	if gs.Seats[1].Life != 37 {
		t.Fatalf("expected opp1 to pay 3; life=%d", gs.Seats[1].Life)
	}
}

func TestSivrissEra3_OpponentsSkipPayMillsReturn(t *testing.T) {
	gs := newGame(t, 2)
	sivriss := addPerm(gs, 0, "Sivriss, Nightmare Speaker", "creature")
	addLibrary(gs, 0, "S1", "S2")
	if gs.Seats[0].Flags == nil {
		gs.Seats[0].Flags = map[string]int{}
	}
	gs.Seats[0].Flags["sivriss_opp_skip_pay"] = 1

	sivrissEra3Activate(gs, sivriss, 0, nil)

	// 1 opp → 1 milled, 1 returned.
	if len(gs.Seats[0].Hand) != 1 {
		t.Fatalf("expected 1 returned to hand; hand=%d", len(gs.Seats[0].Hand))
	}
}

// ---------------------------------------------------------------------------
// Urza, Prince of Kroog
// ---------------------------------------------------------------------------

func TestUrzaPrinceEra3_ActivatedTokenCopyOfArtifact(t *testing.T) {
	gs := newGame(t, 2)
	urza := addPerm(gs, 0, "Urza, Prince of Kroog", "creature")
	urza.Card.BasePower = 2
	urza.Card.BaseToughness = 4
	target := addPerm(gs, 0, "Sol Ring", "artifact")

	urzaPrinceEra3Activate(gs, urza, 0, map[string]interface{}{"target_perm": target})

	// Token should be the most recent battlefield entry on seat 0.
	var tok *gameengine.Permanent
	for _, p := range gs.Seats[0].Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if p.Card.IsCopy {
			tok = p
			break
		}
	}
	if tok == nil {
		t.Fatalf("expected a copy token on the battlefield")
	}
	if !tok.IsCreature() {
		t.Fatalf("token should be a creature")
	}
	if !cardHasType(tok.Card, "soldier") {
		t.Fatalf("token should be a Soldier")
	}
}

// ---------------------------------------------------------------------------
// Plargg and Nassari
// ---------------------------------------------------------------------------

func TestPlarggNassariEra3_UpkeepExilesUntilNonland(t *testing.T) {
	gs := newGame(t, 2)
	plargg := addPerm(gs, 0, "Plargg and Nassari", "creature")
	gs.Seats[0].Library = []*gameengine.Card{
		{Name: "Forest", Owner: 0, Types: []string{"land"}},
		{Name: "Mountain", Owner: 0, Types: []string{"land"}},
		{Name: "Lightning Bolt", Owner: 0, Types: []string{"instant"}},
		{Name: "Fireball", Owner: 0, Types: []string{"sorcery"}},
	}
	gs.Seats[1].Library = []*gameengine.Card{
		{Name: "Plains", Owner: 1, Types: []string{"land"}},
		{Name: "Counterspell", Owner: 1, Types: []string{"instant"}},
	}

	plarggNassariEra3Upkeep(gs, plargg, map[string]interface{}{"active_seat": 0})

	// Seat 0: 2 lands + 1 nonland (Lightning Bolt) exiled (3 total).
	if len(gs.Seats[0].Exile) != 3 {
		t.Fatalf("expected 3 exiled from seat 0; got %d", len(gs.Seats[0].Exile))
	}
	if len(gs.Seats[0].Library) != 1 {
		t.Fatalf("expected 1 card left in seat 0 library (Fireball); got %d",
			len(gs.Seats[0].Library))
	}
	// Seat 1: 1 land + 1 nonland exiled.
	if len(gs.Seats[1].Exile) != 2 {
		t.Fatalf("expected 2 exiled from seat 1; got %d", len(gs.Seats[1].Exile))
	}
}

// ---------------------------------------------------------------------------
// Will, Scion of Peace
// ---------------------------------------------------------------------------

func TestWillScionEra3_StampsDiscountFromLifeGained(t *testing.T) {
	gs := newGame(t, 2)
	will := addPerm(gs, 0, "Will, Scion of Peace", "creature")
	if gs.Seats[0].Flags == nil {
		gs.Seats[0].Flags = map[string]int{}
	}
	gs.Seats[0].Flags["life_gained_this_turn"] = 7

	willScionEra3Activate(gs, will, 0, nil)

	if gs.Seats[0].Flags["will_scion_wu_discount"] != 7 {
		t.Fatalf("expected discount=7; got %d", gs.Seats[0].Flags["will_scion_wu_discount"])
	}
}

// ---------------------------------------------------------------------------
// Felothar the Steadfast: activated handler is owned by
// custom_felothar_steadfast.go (custom wins over era3 stub per QA cleanup).
// The era3 batch retains only the static "ignore defender / damage as
// toughness" markers, exercised by TestEra3Batch_AllRegistered below.
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Registry smoke check
// ---------------------------------------------------------------------------

func TestEra3Batch_AllRegistered(t *testing.T) {
	cards := []string{
		"Jetmir, Nexus of Revels",
		"Falco Spara, Pactweaver",
		"Lord Xander, the Collector",
		"Hidetsugu and Kairi",
		"Shorikai, Genesis Engine",
		"Acererak the Archlich",
		"Tazri, Beacon of Unity",
		"Sivriss, Nightmare Speaker",
		"Urza, Prince of Kroog",
		"Plargg and Nassari",
		"Will, Scion of Peace",
		"Felothar the Steadfast",
	}
	for _, name := range cards {
		hasAny := HasETB(name) || HasResolve(name) || HasActivated(name) || era3HasAnyTrigger(name)
		if !hasAny {
			t.Errorf("%q should have at least one registered handler", name)
		}
	}
}

func era3HasAnyTrigger(name string) bool {
	reg := Global()
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	byEvent, ok := reg.onTrigger[normalizeName(name)]
	if !ok {
		return false
	}
	for _, hs := range byEvent {
		if len(hs) > 0 {
			return true
		}
	}
	return false
}
