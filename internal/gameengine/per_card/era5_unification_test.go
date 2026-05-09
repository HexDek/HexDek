package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// Era 5 unification — focused tests, one (or two) per handler.

// ---------------------------------------------------------------------
// Marchesa, the Black Rose
// ---------------------------------------------------------------------

func TestMarchesa_DyingWithCounterSchedulesReturn(t *testing.T) {
	gs := newGame(t, 2)
	marchesa := addPerm(gs, 0, "Marchesa, the Black Rose", "creature")
	_ = marchesa

	dying := &gameengine.Card{Name: "Reassembling Skeleton", Owner: 0, Types: []string{"creature"}, BasePower: 1}
	dyingPerm := &gameengine.Permanent{
		Card:       dying,
		Controller: 0,
		Owner:      0,
		Counters:   map[string]int{"+1/+1": 1},
		Flags:      map[string]int{},
	}
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, dying)

	delayedBefore := len(gs.DelayedTriggers)
	gameengine.FireCardTrigger(gs, "creature_dies", map[string]interface{}{
		"perm":            dyingPerm,
		"card":            dying,
		"controller_seat": 0,
	})
	if len(gs.DelayedTriggers) <= delayedBefore {
		t.Fatalf("Marchesa should have scheduled a delayed return trigger; before=%d after=%d",
			delayedBefore, len(gs.DelayedTriggers))
	}
}

func TestMarchesa_DyingWithoutCounterDoesNothing(t *testing.T) {
	gs := newGame(t, 2)
	addPerm(gs, 0, "Marchesa, the Black Rose", "creature")

	dying := &gameengine.Card{Name: "Vanilla", Owner: 0, Types: []string{"creature"}, BasePower: 1}
	dyingPerm := &gameengine.Permanent{
		Card: dying, Controller: 0, Owner: 0,
		Counters: map[string]int{}, Flags: map[string]int{},
	}
	delayedBefore := len(gs.DelayedTriggers)
	gameengine.FireCardTrigger(gs, "creature_dies", map[string]interface{}{
		"perm":            dyingPerm,
		"card":            dying,
		"controller_seat": 0,
	})
	if len(gs.DelayedTriggers) != delayedBefore {
		t.Errorf("Marchesa should NOT have scheduled return for counterless creature")
	}
}

// ---------------------------------------------------------------------
// Karador, Ghost Chieftain
// ---------------------------------------------------------------------

func TestKarador_OncePerTurnCreatureFromGraveyard(t *testing.T) {
	gs := newGame(t, 2)
	karador := addPerm(gs, 0, "Karador, Ghost Chieftain", "creature")

	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard,
		&gameengine.Card{Name: "Eternal Witness", Owner: 0, Types: []string{"creature", "cmc:3"}, BasePower: 2},
		&gameengine.Card{Name: "Avenger of Zendikar", Owner: 0, Types: []string{"creature", "cmc:7"}, BasePower: 5},
	)
	bfBefore := len(gs.Seats[0].Battlefield)
	gameengine.InvokeActivatedHook(gs, karador, 0, nil)
	if len(gs.Seats[0].Battlefield) != bfBefore+1 {
		t.Errorf("Karador should have brought back one creature; bf grew by %d",
			len(gs.Seats[0].Battlefield)-bfBefore)
	}

	// Second activation same turn should fail.
	bfBefore = len(gs.Seats[0].Battlefield)
	gameengine.InvokeActivatedHook(gs, karador, 0, nil)
	if len(gs.Seats[0].Battlefield) != bfBefore {
		t.Errorf("Karador should NOT have triggered twice in one turn")
	}
}

// ---------------------------------------------------------------------
// Derevi, Empyrial Tactician
// ---------------------------------------------------------------------

func TestDerevi_ETBTapsOpponentCreature(t *testing.T) {
	gs := newGame(t, 2)
	derevi := addPerm(gs, 0, "Derevi, Empyrial Tactician", "creature")
	opp := addPerm(gs, 1, "Big Stompy", "creature")
	opp.Card.BasePower = 5
	dereviETBTapOrUntap(gs, derevi)
	if !opp.Tapped {
		t.Errorf("Derevi should have tapped the opponent's creature")
	}
}

// ---------------------------------------------------------------------
// Yasharn, Implacable Earth
// ---------------------------------------------------------------------

func TestYasharn_ETBSearchesForestAndPlains(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].Library = []*gameengine.Card{
		{Name: "Forest", Owner: 0, Types: []string{"land", "basic", "forest"}},
		{Name: "Plains", Owner: 0, Types: []string{"land", "basic", "plains"}},
		{Name: "Mountain", Owner: 0, Types: []string{"land", "basic", "mountain"}},
	}
	yasharn := addPerm(gs, 0, "Yasharn, Implacable Earth", "creature")
	yasharnETB(gs, yasharn)
	if len(gs.Seats[0].Hand) != 2 {
		t.Errorf("Yasharn should have fetched 2 basics to hand; got %d", len(gs.Seats[0].Hand))
	}
	if gs.Flags["yasharn_active_seat"] != 1 {
		t.Errorf("yasharn_active_seat flag should be 1; got %d", gs.Flags["yasharn_active_seat"])
	}
}

// ---------------------------------------------------------------------
// Charix, the Raging Isle
// ---------------------------------------------------------------------

func TestCharix_ActivationPumpsByIslandCount(t *testing.T) {
	gs := newGame(t, 2)
	charix := addPerm(gs, 0, "Charix, the Raging Isle", "creature")
	addPerm(gs, 0, "Island", "land", "basic", "island")
	addPerm(gs, 0, "Island", "land", "basic", "island")
	addPerm(gs, 0, "Island", "land", "basic", "island")
	beforePower := charix.Power()
	charixActivatePump(gs, charix, 0, nil)
	if charix.Power() != beforePower+3 {
		t.Errorf("Charix should be +3/-3; got power %d (was %d)", charix.Power(), beforePower)
	}
}

// ---------------------------------------------------------------------
// Kalamax, the Stormsire
// ---------------------------------------------------------------------

func TestKalamax_FirstInstantWhileTappedCopiesAndCounters(t *testing.T) {
	gs := newGame(t, 2)
	kalamax := addPerm(gs, 0, "Kalamax, the Stormsire", "creature")
	kalamax.Tapped = true
	if kalamax.Flags == nil {
		kalamax.Flags = map[string]int{}
	}
	gameengine.FireCardTrigger(gs, "instant_or_sorcery_cast", map[string]interface{}{
		"caster_seat": 0,
		"spell_name":  "Lightning Bolt",
		"is_instant":  true,
	})
	if kalamax.Counters["+1/+1"] != 1 {
		t.Errorf("Kalamax should have a +1/+1 counter after first instant; got %d", kalamax.Counters["+1/+1"])
	}
	if kalamax.Flags["kalamax_first_instant_used"] != 1 {
		t.Errorf("first-instant flag should be set")
	}
}

func TestKalamax_SecondInstantDoesNotCopy(t *testing.T) {
	gs := newGame(t, 2)
	kalamax := addPerm(gs, 0, "Kalamax, the Stormsire", "creature")
	kalamax.Tapped = true
	kalamax.Flags = map[string]int{"kalamax_first_instant_used": 1}
	beforeCounters := kalamax.Counters["+1/+1"]
	gameengine.FireCardTrigger(gs, "instant_or_sorcery_cast", map[string]interface{}{
		"caster_seat": 0,
		"spell_name":  "Counterspell",
		"is_instant":  true,
	})
	if kalamax.Counters["+1/+1"] != beforeCounters {
		t.Errorf("second instant should not add counter; got %d", kalamax.Counters["+1/+1"])
	}
}

// ---------------------------------------------------------------------
// Chainer, Dementia Master
// ---------------------------------------------------------------------

func TestChainer_PaysLifeAndReanimates(t *testing.T) {
	gs := newGame(t, 2)
	chainer := addPerm(gs, 0, "Chainer, Dementia Master", "creature")
	gs.Seats[0].Life = 20
	gs.Seats[1].Graveyard = []*gameengine.Card{
		{Name: "Griselbrand", Owner: 1, Types: []string{"creature", "demon"}, BasePower: 7},
	}
	bfBefore := len(gs.Seats[0].Battlefield)
	chainerReanimate(gs, chainer, 0, nil)
	if gs.Seats[0].Life != 17 {
		t.Errorf("Chainer should pay 3 life; life=%d", gs.Seats[0].Life)
	}
	if len(gs.Seats[0].Battlefield) <= bfBefore {
		t.Errorf("Chainer should have reanimated from opponent's graveyard")
	}
	if len(gs.Seats[1].Graveyard) != 0 {
		t.Errorf("opponent's graveyard should be drained; got %d", len(gs.Seats[1].Graveyard))
	}
}

func TestChainer_RefusesToPayWhenLifeTooLow(t *testing.T) {
	gs := newGame(t, 2)
	chainer := addPerm(gs, 0, "Chainer, Dementia Master", "creature")
	gs.Seats[0].Life = 2
	gs.Seats[0].Graveyard = []*gameengine.Card{
		{Name: "Griselbrand", Owner: 0, Types: []string{"creature"}, BasePower: 7},
	}
	chainerReanimate(gs, chainer, 0, nil)
	if gs.Seats[0].Life != 2 {
		t.Errorf("Chainer should not pay when life too low; got %d", gs.Seats[0].Life)
	}
}

// ---------------------------------------------------------------------
// Ruric Thar, the Unbowed
// ---------------------------------------------------------------------

func TestRuricThar_NoncreatureCastBurnsCaster(t *testing.T) {
	gs := newGame(t, 2)
	addPerm(gs, 0, "Ruric Thar, the Unbowed", "creature")
	gs.Seats[1].Life = 40
	gameengine.FireCardTrigger(gs, "noncreature_spell_cast", map[string]interface{}{
		"caster_seat": 1,
		"spell_name":  "Lightning Bolt",
		"is_creature": false,
	})
	if gs.Seats[1].Life != 34 {
		t.Errorf("Ruric Thar should deal 6 damage to noncreature caster; life=%d", gs.Seats[1].Life)
	}
}

// ---------------------------------------------------------------------
// Selenia, Dark Angel
// ---------------------------------------------------------------------

func TestSelenia_PaysLifeAndBouncesSelf(t *testing.T) {
	gs := newGame(t, 2)
	selenia := addPerm(gs, 0, "Selenia, Dark Angel", "creature")
	gs.Seats[0].Life = 20
	bfBefore := len(gs.Seats[0].Battlefield)
	seleniaPayLifeBounce(gs, selenia, 0, nil)
	if gs.Seats[0].Life != 18 {
		t.Errorf("Selenia should cost 2 life; got %d", gs.Seats[0].Life)
	}
	if len(gs.Seats[0].Battlefield) >= bfBefore {
		t.Errorf("Selenia should be off the battlefield; bf size unchanged")
	}
	inHand := false
	for _, c := range gs.Seats[0].Hand {
		if c == selenia.Card {
			inHand = true
			break
		}
	}
	if !inHand {
		t.Errorf("Selenia should be in owner's hand")
	}
}

// ---------------------------------------------------------------------
// Yurlok of Scorch Thrash
// ---------------------------------------------------------------------

func TestYurlok_GroupRitualAddsManaForAllPlayers(t *testing.T) {
	gs := newGame(t, 3)
	yurlok := addPerm(gs, 0, "Yurlok of Scorch Thrash", "creature")
	for i := range gs.Seats {
		gs.Seats[i].ManaPool = 0
	}
	yurlokGroupMana(gs, yurlok, 0, nil)
	for i, s := range gs.Seats {
		if s.ManaPool != 3 {
			t.Errorf("seat %d should have 3 mana from Yurlok; got %d", i, s.ManaPool)
		}
	}
}

// ---------------------------------------------------------------------
// Sakashima of a Thousand Faces
// ---------------------------------------------------------------------

func TestSakashima_ETBCopiesPowerOfBestCreature(t *testing.T) {
	gs := newGame(t, 2)
	target := addPerm(gs, 0, "Avenger of Zendikar", "creature")
	target.Card.BasePower = 5
	target.Card.BaseToughness = 5
	saka := addPerm(gs, 0, "Sakashima of a Thousand Faces", "creature")
	saka.Card.BasePower = 3
	saka.Card.BaseToughness = 1
	sakashimaCopyETB(gs, saka)
	if saka.Card.BasePower != 5 || saka.Card.BaseToughness != 5 {
		t.Errorf("Sakashima should copy 5/5; got %d/%d", saka.Card.BasePower, saka.Card.BaseToughness)
	}
}

// ---------------------------------------------------------------------
// Araumi of the Dead Tide
// ---------------------------------------------------------------------

func TestAraumi_ExilesAndSpawnsTokenPerOpponent(t *testing.T) {
	gs := newGame(t, 4) // 3 opponents
	araumi := addPerm(gs, 0, "Araumi of the Dead Tide", "creature")
	gs.Seats[0].Graveyard = []*gameengine.Card{
		{Name: "Massacre Wurm", Owner: 0, Types: []string{"creature"}, BasePower: 6},
		{Name: "Filler1", Owner: 0, Types: []string{"creature"}},
		{Name: "Filler2", Owner: 0, Types: []string{"creature"}},
		{Name: "Filler3", Owner: 0, Types: []string{"creature"}},
	}
	bfBefore := len(gs.Seats[0].Battlefield)
	araumiEncore(gs, araumi, 0, nil)
	// 3 opponents → 3 tokens spawned
	if len(gs.Seats[0].Battlefield) != bfBefore+3 {
		t.Errorf("Araumi should spawn 3 token copies; bf grew by %d", len(gs.Seats[0].Battlefield)-bfBefore)
	}
	// Encore target + 3 cost cards → 4 exiled total.
	if len(gs.Seats[0].Exile) != 4 {
		t.Errorf("Araumi should exile 4 cards (target + 3 opponents); exile=%d", len(gs.Seats[0].Exile))
	}
}

func TestAraumi_RefusesIfNotEnoughGraveyardForCost(t *testing.T) {
	gs := newGame(t, 4) // 3 opponents
	araumi := addPerm(gs, 0, "Araumi of the Dead Tide", "creature")
	gs.Seats[0].Graveyard = []*gameengine.Card{
		{Name: "Wurm", Owner: 0, Types: []string{"creature"}, BasePower: 5},
		{Name: "Filler1", Owner: 0, Types: []string{"creature"}},
		// Only 1 cost card available, need 3.
	}
	bfBefore := len(gs.Seats[0].Battlefield)
	araumiEncore(gs, araumi, 0, nil)
	if len(gs.Seats[0].Battlefield) != bfBefore {
		t.Errorf("Araumi should refuse when graveyard insufficient")
	}
}

// ---------------------------------------------------------------------
// Mairsil, the Pretender
// ---------------------------------------------------------------------

func TestMairsil_ETBCagesArtifactFromHand(t *testing.T) {
	gs := newGame(t, 2)
	mairsil := addPerm(gs, 0, "Mairsil, the Pretender", "creature")
	gs.Seats[0].Hand = []*gameengine.Card{
		{Name: "Mox Diamond", Owner: 0, Types: []string{"artifact", "cmc:0"}},
	}
	mairsilETB(gs, mairsil)
	if len(gs.Seats[0].Exile) != 1 {
		t.Errorf("Mairsil should have caged 1 card; exile=%d", len(gs.Seats[0].Exile))
	}
	if len(gs.Seats[0].Hand) != 0 {
		t.Errorf("hand should be empty; got %d", len(gs.Seats[0].Hand))
	}
	caged := gs.Seats[0].Exile[0]
	hasMarker := false
	for _, t := range caged.Types {
		if t == "cage_counter" {
			hasMarker = true
			break
		}
	}
	if !hasMarker {
		t.Errorf("caged card should bear cage_counter marker")
	}
}

// ---------------------------------------------------------------------
// Smoke check — every era 5 unification commander has at least one handler
// ---------------------------------------------------------------------

func TestEra5Unification_AllRegistered(t *testing.T) {
	cards := []string{
		"Marchesa, the Black Rose",
		"Karador, Ghost Chieftain",
		"Derevi, Empyrial Tactician",
		"Yasharn, Implacable Earth",
		"Charix, the Raging Isle",
		"Kalamax, the Stormsire",
		"Chainer, Dementia Master",
		"Ruric Thar, the Unbowed",
		"Selenia, Dark Angel",
		"Yurlok of Scorch Thrash",
		"Sakashima of a Thousand Faces",
		"Araumi of the Dead Tide",
		"Mairsil, the Pretender",
	}
	for _, name := range cards {
		hasAny := HasETB(name) || HasResolve(name) || HasActivated(name) || hasAnyTriggerEra5(name)
		if !hasAny {
			t.Errorf("%q should have at least one registered handler", name)
		}
	}
}

func hasAnyTriggerEra5(name string) bool {
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
