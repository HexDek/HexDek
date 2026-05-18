package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// R36 stub-batch ports — five gen_*.go pure-stub handlers ported into
// real per-card behaviour. Each card has at least one happy-path test
// hitting the primary clause; tests follow the muninn_handlers_*_test.go
// style established in this package.

// ---------------------------------------------------------------------------
// Ashling, Flame Dancer — magecraft chain (R, R, then RRRR)
// ---------------------------------------------------------------------------

// addCardToHand mints a card with the given types and CMC and parks it
// in seat's hand. The CMC is encoded as a "cmc:N" Type token because
// the per_card cardCMC() helper reads CMC from that token rather than
// the Card.CMC scalar (see per_card/helpers.go).
func addCardToHand(gs *gameengine.GameState, seat int, name string, cmc int, types ...string) *gameengine.Card {
	ts := append([]string{}, types...)
	ts = append(ts, "cmc:"+itoaR36(cmc))
	c := &gameengine.Card{
		Name:  name,
		Owner: seat,
		CMC:   cmc,
		Types: ts,
	}
	gs.Seats[seat].Hand = append(gs.Seats[seat].Hand, c)
	return c
}

// addGYCard adds a card to seat's graveyard with cmc encoded as a
// Type token (cardCMC convention).
func addGYCard(gs *gameengine.GameState, seat int, name string, cmc int, types ...string) *gameengine.Card {
	ts := append([]string{}, types...)
	ts = append(ts, "cmc:"+itoaR36(cmc))
	c := &gameengine.Card{
		Name:  name,
		Owner: seat,
		CMC:   cmc,
		Types: ts,
	}
	gs.Seats[seat].Graveyard = append(gs.Seats[seat].Graveyard, c)
	return c
}

// itoaR36 — small int-to-string for the cmc token assembly.
func itoaR36(n int) string {
	if n == 0 {
		return "0"
	}
	out := []byte{}
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		out = append([]byte{byte('0' + n%10)}, out...)
		n /= 10
	}
	if neg {
		out = append([]byte{'-'}, out...)
	}
	return string(out)
}

// stampCreaturePT sets BasePower / BaseToughness on a permanent's
// underlying card so SBA doesn't destroy it as a 0/0 mid-test when
// triggers route through the stack-resolution pipeline.
func stampCreaturePT(p *gameengine.Permanent, power, toughness int) *gameengine.Permanent {
	if p == nil || p.Card == nil {
		return p
	}
	p.Card.BasePower = power
	p.Card.BaseToughness = toughness
	return p
}

func TestAshlingFlameDancer_FirstCastDiscardsAndDraws(t *testing.T) {
	gs := newGame(t, 2)
	ashling := stampCreaturePT(addPerm(gs, 0, "Ashling, Flame Dancer", "creature"), 2, 3)
	addCardToHand(gs, 0, "Cheap", 1, "instant")
	addLibrary(gs, 0, "Top", "Top2")

	gameengine.FireCardTrigger(gs, "instant_or_sorcery_cast", map[string]interface{}{
		"caster_seat": 0,
	})

	if len(gs.Seats[0].Hand) != 1 {
		t.Errorf("hand size after 1st magecraft = %d, want 1 (discard then draw)", len(gs.Seats[0].Hand))
	}
	if len(gs.Seats[0].Graveyard) != 1 || gs.Seats[0].Graveyard[0].DisplayName() != "Cheap" {
		t.Errorf("expected discarded card in graveyard; graveyard=%v",
			grave_names(gs.Seats[0].Graveyard))
	}
	if got := ashling.Flags[ashlingFlameDancerTurnKey(gs.Turn)]; got != 1 {
		t.Errorf("resolution counter = %d, want 1", got)
	}
}

func TestAshlingFlameDancer_SecondCastDamagesEachOpponentAndCreatures(t *testing.T) {
	gs := newGame(t, 3) // 3 seats to verify "each opponent"
	ashling := stampCreaturePT(addPerm(gs, 0, "Ashling, Flame Dancer", "creature"), 2, 3)
	_ = ashling
	addCardToHand(gs, 0, "A", 1, "instant")
	addCardToHand(gs, 0, "B", 1, "instant")
	addLibrary(gs, 0, "T1", "T2", "T3")

	// Opponent creatures we'll watch for 2 marked damage.
	opp1Creature := stampCreaturePT(addPerm(gs, 1, "Opp1 Bear", "creature"), 2, 2)
	opp2Creature := stampCreaturePT(addPerm(gs, 2, "Opp2 Wolf", "creature"), 3, 3)

	preLife1 := gs.Seats[1].Life
	preLife2 := gs.Seats[2].Life

	// 1st magecraft fire — no damage.
	gameengine.FireCardTrigger(gs, "instant_or_sorcery_cast", map[string]interface{}{
		"caster_seat": 0,
	})
	if opp1Creature.MarkedDamage != 0 || opp2Creature.MarkedDamage != 0 {
		t.Fatal("1st magecraft fire should not damage anything yet")
	}

	// 2nd magecraft fire — 2 damage to each opponent and each of their creatures.
	gameengine.FireCardTrigger(gs, "instant_or_sorcery_cast", map[string]interface{}{
		"caster_seat": 0,
	})
	if gs.Seats[1].Life != preLife1-2 {
		t.Errorf("seat 1 life = %d, want %d", gs.Seats[1].Life, preLife1-2)
	}
	if gs.Seats[2].Life != preLife2-2 {
		t.Errorf("seat 2 life = %d, want %d", gs.Seats[2].Life, preLife2-2)
	}
	if opp1Creature.MarkedDamage != 2 {
		t.Errorf("opp1 creature marked dmg = %d, want 2", opp1Creature.MarkedDamage)
	}
	if opp2Creature.MarkedDamage != 2 {
		t.Errorf("opp2 creature marked dmg = %d, want 2", opp2Creature.MarkedDamage)
	}
}

func TestAshlingFlameDancer_ThirdCastAddsRRRRMana(t *testing.T) {
	gs := newGame(t, 2)
	stampCreaturePT(addPerm(gs, 0, "Ashling, Flame Dancer", "creature"), 2, 3)
	addCardToHand(gs, 0, "A", 1, "instant")
	addCardToHand(gs, 0, "B", 1, "instant")
	addCardToHand(gs, 0, "C", 1, "instant")
	addLibrary(gs, 0, "T1", "T2", "T3")

	preMana := gs.Seats[0].ManaPool

	// Fire 3 magecraft triggers.
	for i := 0; i < 3; i++ {
		gameengine.FireCardTrigger(gs, "instant_or_sorcery_cast", map[string]interface{}{
			"caster_seat": 0,
		})
	}

	if gs.Seats[0].ManaPool != preMana+4 {
		t.Errorf("mana pool after 3rd magecraft = %d, want %d (+4 from RRRR)",
			gs.Seats[0].ManaPool, preMana+4)
	}
}

func TestAshlingFlameDancer_OpponentCastDoesNotTrigger(t *testing.T) {
	gs := newGame(t, 2)
	ashling := stampCreaturePT(addPerm(gs, 0, "Ashling, Flame Dancer", "creature"), 2, 3)
	addCardToHand(gs, 0, "A", 1, "instant")
	addLibrary(gs, 0, "T1")

	gameengine.FireCardTrigger(gs, "instant_or_sorcery_cast", map[string]interface{}{
		"caster_seat": 1, // opponent's cast
	})

	if got := ashling.Flags[ashlingFlameDancerTurnKey(gs.Turn)]; got != 0 {
		t.Errorf("opponent's cast should NOT bump Ashling's counter; got %d", got)
	}
	if len(gs.Seats[0].Graveyard) != 0 {
		t.Errorf("opponent's cast should not discard; gy=%v", grave_names(gs.Seats[0].Graveyard))
	}
}

// ---------------------------------------------------------------------------
// Tannuk, Memorial Ensign — landfall: 1 dmg each opp; draw on 2nd
// ---------------------------------------------------------------------------

func TestTannukMemorialEnsign_LandfallDamagesEachOpponent(t *testing.T) {
	gs := newGame(t, 3)
	stampCreaturePT(addPerm(gs, 0, "Tannuk, Memorial Ensign", "creature"), 2, 2)
	land := addPerm(gs, 0, "Mountain", "land")

	preLife1 := gs.Seats[1].Life
	preLife2 := gs.Seats[2].Life

	gameengine.FireCardTrigger(gs, "permanent_etb", map[string]interface{}{
		"perm":            land,
		"controller_seat": 0,
	})

	if gs.Seats[1].Life != preLife1-1 {
		t.Errorf("seat 1 life = %d, want %d (1 dmg)", gs.Seats[1].Life, preLife1-1)
	}
	if gs.Seats[2].Life != preLife2-1 {
		t.Errorf("seat 2 life = %d, want %d (1 dmg)", gs.Seats[2].Life, preLife2-1)
	}
}

func TestTannukMemorialEnsign_SecondLandfallDrawsCard(t *testing.T) {
	gs := newGame(t, 2)
	stampCreaturePT(addPerm(gs, 0, "Tannuk, Memorial Ensign", "creature"), 2, 2)
	addLibrary(gs, 0, "Top")
	land1 := addPerm(gs, 0, "Mountain1", "land")
	land2 := addPerm(gs, 0, "Mountain2", "land")

	// 1st landfall — no draw.
	gameengine.FireCardTrigger(gs, "permanent_etb", map[string]interface{}{
		"perm":            land1,
		"controller_seat": 0,
	})
	if len(gs.Seats[0].Hand) != 0 {
		t.Errorf("1st landfall should not draw; hand=%d", len(gs.Seats[0].Hand))
	}

	// 2nd landfall — draws Top.
	gameengine.FireCardTrigger(gs, "permanent_etb", map[string]interface{}{
		"perm":            land2,
		"controller_seat": 0,
	})
	if len(gs.Seats[0].Hand) != 1 || gs.Seats[0].Hand[0].DisplayName() != "Top" {
		t.Errorf("2nd landfall should draw 'Top'; hand=%v", hand_names(gs.Seats[0].Hand))
	}
}

func TestTannukMemorialEnsign_OpponentLandDoesNotTrigger(t *testing.T) {
	gs := newGame(t, 2)
	stampCreaturePT(addPerm(gs, 0, "Tannuk, Memorial Ensign", "creature"), 2, 2)
	land := addPerm(gs, 1, "Opp Mountain", "land") // opponent's land

	preLife1 := gs.Seats[1].Life

	gameengine.FireCardTrigger(gs, "permanent_etb", map[string]interface{}{
		"perm":            land,
		"controller_seat": 1,
	})

	if gs.Seats[1].Life != preLife1 {
		t.Errorf("opponent's land should not trigger landfall; opp life=%d (want %d)",
			gs.Seats[1].Life, preLife1)
	}
}

func TestTannukMemorialEnsign_NonLandETBDoesNotTrigger(t *testing.T) {
	gs := newGame(t, 2)
	stampCreaturePT(addPerm(gs, 0, "Tannuk, Memorial Ensign", "creature"), 2, 2)
	creature := stampCreaturePT(addPerm(gs, 0, "Grizzly Bears", "creature"), 2, 2)

	preLife1 := gs.Seats[1].Life

	gameengine.FireCardTrigger(gs, "permanent_etb", map[string]interface{}{
		"perm":            creature,
		"controller_seat": 0,
	})

	if gs.Seats[1].Life != preLife1 {
		t.Errorf("non-land ETB should not trigger landfall; opp life=%d", gs.Seats[1].Life)
	}
}

// ---------------------------------------------------------------------------
// Toph, Hardheaded Teacher — ETB discard-to-return-instant/sorcery
// ---------------------------------------------------------------------------

func TestTophHardheaded_ETBDiscardsAndReturnsInstantSorcery(t *testing.T) {
	gs := newGame(t, 2)
	toph := stampCreaturePT(addPerm(gs, 0, "Toph, Hardheaded Teacher", "creature"), 3, 3)

	// Hand: a cheap card to discard.
	discardable := addCardToHand(gs, 0, "Cheap Cantrip", 1, "sorcery")
	// Graveyard: an instant + sorcery; the higher-MV should be returned.
	cheap_sorcery := addGYCard(gs, 0, "Cheap GY", 1, "sorcery")
	juicy_instant := addGYCard(gs, 0, "Cyclonic Rift", 7, "instant")

	gameengine.InvokeETBHook(gs, toph)

	// discardable should now be in graveyard.
	inGY := false
	for _, c := range gs.Seats[0].Graveyard {
		if c == discardable {
			inGY = true
		}
	}
	if !inGY {
		t.Error("discardable card should be in graveyard after discard")
	}
	// juicy_instant should be in hand (highest-MV return target).
	inHand := false
	for _, c := range gs.Seats[0].Hand {
		if c == juicy_instant {
			inHand = true
		}
	}
	if !inHand {
		t.Errorf("highest-MV instant/sorcery should have been returned to hand; hand=%v",
			hand_names(gs.Seats[0].Hand))
	}
	// cheap_sorcery should still be in graveyard (not the picked return).
	stillInGY := false
	for _, c := range gs.Seats[0].Graveyard {
		if c == cheap_sorcery {
			stillInGY = true
		}
	}
	if !stillInGY {
		t.Error("cheap graveyard sorcery should still be in graveyard (return picked the bigger one)")
	}
}

func TestTophHardheaded_DeclinesWhenGraveyardEmpty(t *testing.T) {
	gs := newGame(t, 2)
	toph := stampCreaturePT(addPerm(gs, 0, "Toph, Hardheaded Teacher", "creature"), 3, 3)
	addCardToHand(gs, 0, "Hand Card", 1, "sorcery")

	preHand := len(gs.Seats[0].Hand)
	preGY := len(gs.Seats[0].Graveyard)

	gameengine.InvokeETBHook(gs, toph)

	if len(gs.Seats[0].Hand) != preHand {
		t.Errorf("hand size changed when no eligible graveyard target; got %d want %d",
			len(gs.Seats[0].Hand), preHand)
	}
	if len(gs.Seats[0].Graveyard) != preGY {
		t.Errorf("graveyard changed when no eligible target; got %d want %d",
			len(gs.Seats[0].Graveyard), preGY)
	}
}

func TestTophHardheaded_DeclinesWhenHandEmpty(t *testing.T) {
	gs := newGame(t, 2)
	toph := stampCreaturePT(addPerm(gs, 0, "Toph, Hardheaded Teacher", "creature"), 3, 3)
	addGYCard(gs, 0, "Cyclonic Rift", 7, "instant")

	gameengine.InvokeETBHook(gs, toph)

	if len(gs.Seats[0].Graveyard) != 1 {
		t.Errorf("hand-empty decline: graveyard should be untouched; got %d entries",
			len(gs.Seats[0].Graveyard))
	}
	if len(gs.Seats[0].Hand) != 0 {
		t.Error("hand-empty decline: hand should remain empty")
	}
}

// ---------------------------------------------------------------------------
// Magnus the Red — combat damage to player → 3/3 red Spawn token
// ---------------------------------------------------------------------------

func TestMagnusTheRed_CombatDamageMakesSpawnToken(t *testing.T) {
	gs := newGame(t, 2)
	magnus := stampCreaturePT(addPerm(gs, 0, "Magnus the Red", "creature", "legendary"), 6, 6)

	preBattlefield := len(gs.Seats[0].Battlefield)
	gameengine.FireCardTrigger(gs, "combat_damage_to_player", map[string]interface{}{
		"source_perm":  magnus,
		"source_seat":  0,
		"target_seat":  1,
		"amount":       5,
	})

	if len(gs.Seats[0].Battlefield) != preBattlefield+1 {
		t.Fatalf("battlefield should grow by 1 (Spawn token); got %d want %d",
			len(gs.Seats[0].Battlefield), preBattlefield+1)
	}
	newest := gs.Seats[0].Battlefield[len(gs.Seats[0].Battlefield)-1]
	if newest.Card.DisplayName() != "Spawn Token" {
		t.Errorf("new perm = %q, want \"Spawn Token\"", newest.Card.DisplayName())
	}
	if newest.Card.BasePower != 3 || newest.Card.BaseToughness != 3 {
		t.Errorf("Spawn token P/T = %d/%d, want 3/3",
			newest.Card.BasePower, newest.Card.BaseToughness)
	}
}

func TestMagnusTheRed_OtherCreatureCombatDamageDoesNotMakeToken(t *testing.T) {
	gs := newGame(t, 2)
	stampCreaturePT(addPerm(gs, 0, "Magnus the Red", "creature", "legendary"), 6, 6)
	other := stampCreaturePT(addPerm(gs, 0, "Some Other Attacker", "creature"), 3, 3)

	preBattlefield := len(gs.Seats[0].Battlefield)
	// Other creature deals combat damage — Magnus must not fire.
	gameengine.FireCardTrigger(gs, "combat_damage_to_player", map[string]interface{}{
		"source_perm":  other,
		"source_seat":  0,
		"target_seat":  1,
		"amount":       3,
	})

	if len(gs.Seats[0].Battlefield) != preBattlefield {
		t.Errorf("a different creature's combat damage should not spawn a token; battlefield grew %d→%d",
			preBattlefield, len(gs.Seats[0].Battlefield))
	}
}

// ---------------------------------------------------------------------------
// Morlun, Devourer of Spiders — ETB X counters + X damage to opp
// ---------------------------------------------------------------------------

func TestMorlunDevourerOfSpiders_ETBPlacesXCountersAndDealsXDamage(t *testing.T) {
	gs := newGame(t, 3)
	// Stash X=5 via the OnCast capture flag (simulating a cast at X=5).
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	gs.Flags["_morlun_x_0"] = 5

	// Seat 1 has higher life than seat 2 → auto-target should be seat 1.
	gs.Seats[1].Life = 25
	gs.Seats[2].Life = 18

	morlun := stampCreaturePT(addPerm(gs, 0, "Morlun, Devourer of Spiders", "creature", "legendary"), 3, 3)
	gameengine.InvokeETBHook(gs, morlun)

	if morlun.Counters["+1/+1"] != 5 {
		t.Errorf("Morlun +1/+1 counters = %d, want 5", morlun.Counters["+1/+1"])
	}
	if gs.Seats[1].Life != 25-5 {
		t.Errorf("highest-life opponent (seat 1) life = %d, want %d", gs.Seats[1].Life, 25-5)
	}
	if gs.Seats[2].Life != 18 {
		t.Errorf("seat 2 should NOT be damaged; life=%d, want 18", gs.Seats[2].Life)
	}
	// Flag should be consumed.
	if _, exists := gs.Flags["_morlun_x_0"]; exists {
		t.Error("_morlun_x_0 flag should have been consumed by ETB")
	}
}

func TestMorlunDevourerOfSpiders_XZeroNoCountersNoDamage(t *testing.T) {
	gs := newGame(t, 2)
	preLife := gs.Seats[1].Life

	morlun := stampCreaturePT(addPerm(gs, 0, "Morlun, Devourer of Spiders", "creature", "legendary"), 3, 3)
	gameengine.InvokeETBHook(gs, morlun)

	if morlun.Counters["+1/+1"] != 0 {
		t.Errorf("X=0 should mean 0 counters; got %d", morlun.Counters["+1/+1"])
	}
	if gs.Seats[1].Life != preLife {
		t.Errorf("X=0 should deal 0 damage; opp life=%d (want %d)", gs.Seats[1].Life, preLife)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func grave_names(gy []*gameengine.Card) []string {
	out := []string{}
	for _, c := range gy {
		out = append(out, c.DisplayName())
	}
	return out
}

func hand_names(h []*gameengine.Card) []string {
	out := []string{}
	for _, c := range h {
		out = append(out, c.DisplayName())
	}
	return out
}
