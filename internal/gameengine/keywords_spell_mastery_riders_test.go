package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// Round-33 integration tests for SpellMastery + Constellation rider
// hooks. SpellMastery is wired through resolveGatedRider (sibling of
// Threshold/Metalcraft/Hellbent); Constellation is wired through
// FirePermanentETBTriggers via OnConstellationETB (sibling of
// OnEnchantmentETB / Eerie).

// ---------------------------------------------------------------------------
// Test fixtures
// ---------------------------------------------------------------------------

func sm_makeGame(t *testing.T) *GameState {
	t.Helper()
	return NewGameState(2, rand.New(rand.NewSource(33)), nil)
}

// sm_makeRiderCard builds a Card whose AST carries a single Static
// ability with the given raw text. OracleTextLower reconstructs from
// that for the rider detectors.
func sm_makeRiderCard(name, oracleText string) *Card {
	return &Card{
		Name:  name,
		Types: []string{"sorcery"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: oracleText},
			},
		},
	}
}

func sm_makeEnchantmentCard(name, oracleText string) *Card {
	return &Card{
		Name:  name,
		Types: []string{"enchantment"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: oracleText},
			},
		},
	}
}

func sm_makePerm(seat int, card *Card) *Permanent {
	card.Owner = seat
	return &Permanent{
		Card:       card,
		Controller: seat,
		Owner:      seat,
		Flags:      map[string]int{},
	}
}

func sm_makeInstantInGrave(name string) *Card {
	return &Card{Name: name, Types: []string{"instant"}}
}

func sm_makeSorceryInGrave(name string) *Card {
	return &Card{Name: name, Types: []string{"sorcery"}}
}

func sm_makeCreatureInGrave(name string) *Card {
	return &Card{Name: name, Types: []string{"creature"}}
}

func sm_countEvents(gs *GameState, kind string) int {
	n := 0
	for _, ev := range gs.EventLog {
		if ev.Kind == kind {
			n++
		}
	}
	return n
}

// sm_runSpell drives resolveSequence with an empty body — enough to
// fire the resolveGatedRider hook on the outer unwind.
func sm_runSpell(gs *GameState, src *Permanent) {
	resolveSequence(gs, src, &gameast.Sequence{Items: nil})
}

// ---------------------------------------------------------------------------
// (c) Detectors correct.
// ---------------------------------------------------------------------------

func TestHasSpellMastery_DetectorPaths(t *testing.T) {
	// Oracle em-dash.
	em := sm_makeRiderCard("Magmatic Insight",
		"Discard a land card: Draw two cards. Spell mastery — Draw three cards instead.")
	if !HasSpellMastery(em) {
		t.Fatal("HasSpellMastery should detect em-dash rider in oracle text")
	}
	// ASCII hyphen fallback.
	dash := sm_makeRiderCard("Whatever",
		"Spell mastery - Tap target creature.")
	if !HasSpellMastery(dash) {
		t.Fatal("HasSpellMastery should detect ASCII-hyphen rider")
	}
	// AST keyword tagged.
	tagged := &Card{
		Name:  "Tagged SM",
		Types: []string{"instant"},
		AST: &gameast.CardAST{
			Name: "Tagged SM",
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "spell mastery", Raw: "spell mastery"},
			},
		},
	}
	if !HasSpellMastery(tagged) {
		t.Fatal("HasSpellMastery should detect AST keyword tag")
	}
	// Negative.
	plain := sm_makeRiderCard("Plain", "Plain has flying.")
	if HasSpellMastery(plain) {
		t.Fatal("HasSpellMastery must be false on a non-rider card")
	}
	if HasSpellMastery(nil) {
		t.Fatal("HasSpellMastery(nil) should be false")
	}
}

func TestHasConstellationRider_DetectorPaths(t *testing.T) {
	em := sm_makeEnchantmentCard("Eidolon of Blossoms",
		"Constellation — Whenever Eidolon of Blossoms or another enchantment enters under your control, draw a card.")
	if !HasConstellationRider(em) {
		t.Fatal("HasConstellationRider should detect em-dash form")
	}
	dash := sm_makeEnchantmentCard("Whatever Constellation",
		"Constellation - Gain 1 life.")
	if !HasConstellationRider(dash) {
		t.Fatal("HasConstellationRider should detect ASCII-hyphen form")
	}
	tagged := &Card{
		Name:  "Tagged C",
		Types: []string{"enchantment"},
		AST: &gameast.CardAST{
			Name: "Tagged C",
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "constellation", Raw: "constellation"},
			},
		},
	}
	if !HasConstellationRider(tagged) {
		t.Fatal("HasConstellationRider should detect AST keyword tag")
	}
	plain := sm_makeEnchantmentCard("Plain Enchantment", "Plain has no constellation.")
	if HasConstellationRider(plain) {
		t.Fatal("HasConstellationRider must be false on a non-rider card")
	}
	if HasConstellationRider(nil) {
		t.Fatal("HasConstellationRider(nil) should be false")
	}
}

func TestSpellMasteryActive_DelegatesToCheckSpellMastery(t *testing.T) {
	gs := sm_makeGame(t)
	gs.Seats[0].Graveyard = nil
	if SpellMasteryActive(gs, 0) {
		t.Fatal("should be false with empty graveyard")
	}
	gs.Seats[0].Graveyard = []*Card{sm_makeInstantInGrave("a")}
	if SpellMasteryActive(gs, 0) {
		t.Fatal("should be false with 1 instant/sorcery in graveyard")
	}
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, sm_makeSorceryInGrave("b"))
	if !SpellMasteryActive(gs, 0) {
		t.Fatal("should be true with 2 instant/sorcery in graveyard")
	}
}

// ---------------------------------------------------------------------------
// (a) Spell-mastery-active spell resolves with bonus, inactive without.
// ---------------------------------------------------------------------------

func TestSpellMasteryRider_FiresWhenActive(t *testing.T) {
	gs := sm_makeGame(t)
	gs.Seats[0].Graveyard = []*Card{
		sm_makeInstantInGrave("Bolt"),
		sm_makeSorceryInGrave("Day's Undoing"),
	}
	src := sm_makePerm(0, sm_makeRiderCard("Magmatic Insight",
		"Discard a land card: Draw two cards. Spell mastery — Draw three cards instead."))
	sm_runSpell(gs, src)

	if sm_countEvents(gs, "spell_mastery_rider") != 1 {
		t.Fatalf("expected 1 spell_mastery_rider event when active; got %d",
			sm_countEvents(gs, "spell_mastery_rider"))
	}
}

func TestSpellMasteryRider_SilentWhenInactive(t *testing.T) {
	gs := sm_makeGame(t)
	// Only 1 instant in graveyard — below the §702.106 threshold of 2.
	gs.Seats[0].Graveyard = []*Card{sm_makeInstantInGrave("Bolt")}
	src := sm_makePerm(0, sm_makeRiderCard("Magmatic Insight",
		"Spell mastery — Draw three cards instead."))
	sm_runSpell(gs, src)

	if sm_countEvents(gs, "spell_mastery_rider") != 0 {
		t.Fatalf("expected 0 spell_mastery_rider events when inactive; got %d",
			sm_countEvents(gs, "spell_mastery_rider"))
	}
}

func TestSpellMasteryRider_CreaturesInGraveDontCount(t *testing.T) {
	gs := sm_makeGame(t)
	// 3 creatures — not instants/sorceries; spell mastery should NOT fire.
	gs.Seats[0].Graveyard = []*Card{
		sm_makeCreatureInGrave("Bear"),
		sm_makeCreatureInGrave("Goblin"),
		sm_makeCreatureInGrave("Elf"),
	}
	src := sm_makePerm(0, sm_makeRiderCard("Magmatic Insight",
		"Spell mastery — Draw three cards instead."))
	sm_runSpell(gs, src)

	if sm_countEvents(gs, "spell_mastery_rider") != 0 {
		t.Fatal("creatures in graveyard must not satisfy spell mastery")
	}
}

// ---------------------------------------------------------------------------
// (b) Constellation rider fires on enchantment ETB via the ETB hook.
// ---------------------------------------------------------------------------

func TestConstellation_FiresOnEnchantmentETB(t *testing.T) {
	gs := sm_makeGame(t)
	// Pre-existing constellation source on the battlefield. Use the
	// Keyword AST node directly — FireConstellationTriggers reads
	// HasKeyword("constellation"), which matches Keyword nodes (the
	// oracle-text-only path is only for HasConstellationRider).
	carrierCard := &Card{
		Name:  "Eidolon of Blossoms",
		Owner: 0,
		Types: []string{"enchantment", "creature"},
		AST: &gameast.CardAST{
			Name: "Eidolon of Blossoms",
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "constellation", Raw: "constellation"},
			},
		},
	}
	carrier := sm_makePerm(0, carrierCard)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, carrier)

	// Now a new enchantment enters.
	entering := sm_makePerm(0, sm_makeEnchantmentCard("Sigil of the Empty Throne",
		"Whenever you cast an enchantment spell, create a 4/4 Angel."))
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, entering)

	OnConstellationETB(gs, entering)

	if sm_countEvents(gs, "constellation_trigger") < 1 {
		t.Fatalf("expected at least 1 constellation_trigger event; got %d",
			sm_countEvents(gs, "constellation_trigger"))
	}
}

func TestConstellation_NoFireOnNonEnchantmentETB(t *testing.T) {
	gs := sm_makeGame(t)
	carrier := sm_makePerm(0, sm_makeEnchantmentCard("Eidolon of Blossoms",
		"Constellation — Draw a card."))
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, carrier)

	// A creature enters — constellation must NOT fire.
	creatureCard := &Card{
		Name:          "Plain Bear",
		Types:         []string{"creature"},
		BasePower:     2,
		BaseToughness: 2,
		AST:           &gameast.CardAST{Name: "Plain Bear"},
	}
	creaturePerm := sm_makePerm(0, creatureCard)
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, creaturePerm)

	OnConstellationETB(gs, creaturePerm)

	if sm_countEvents(gs, "constellation_trigger") != 0 {
		t.Fatalf("constellation must NOT trigger on creature ETB; got %d",
			sm_countEvents(gs, "constellation_trigger"))
	}
}

// ---------------------------------------------------------------------------
// (d) Per-seat independence.
// ---------------------------------------------------------------------------

func TestSpellMasteryRider_PerSeatIndependent(t *testing.T) {
	gs := sm_makeGame(t)
	// Seat 0 has 2 instants in graveyard — active.
	gs.Seats[0].Graveyard = []*Card{sm_makeInstantInGrave("a"), sm_makeInstantInGrave("b")}
	// Seat 1 has only 1 — inactive.
	gs.Seats[1].Graveyard = []*Card{sm_makeInstantInGrave("c")}

	// Spell controlled by seat 0 — should fire.
	src0 := sm_makePerm(0, sm_makeRiderCard("S0",
		"Spell mastery — Bonus."))
	sm_runSpell(gs, src0)
	if sm_countEvents(gs, "spell_mastery_rider") != 1 {
		t.Fatalf("seat 0's spell-mastery should fire; got %d events",
			sm_countEvents(gs, "spell_mastery_rider"))
	}

	// Spell controlled by seat 1 — should NOT fire even though seat 0's
	// graveyard is loaded.
	src1 := sm_makePerm(1, sm_makeRiderCard("S1",
		"Spell mastery — Bonus."))
	sm_runSpell(gs, src1)
	// Still only the 1 event from before.
	if sm_countEvents(gs, "spell_mastery_rider") != 1 {
		t.Fatalf("seat 1's spell-mastery must not fire (only 1 spell in own grave); got %d events total",
			sm_countEvents(gs, "spell_mastery_rider"))
	}
}

func TestConstellation_PerSeatIndependent(t *testing.T) {
	gs := sm_makeGame(t)
	// Constellation source under seat 0.
	carrier := sm_makePerm(0, sm_makeEnchantmentCard("Eidolon of Blossoms",
		"Constellation — Draw a card."))
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, carrier)

	// Enchantment enters under SEAT 1 — seat 0's constellation must NOT
	// trigger per CR §702.131 ("an enchantment YOU CONTROL").
	enemyEnch := sm_makePerm(1, sm_makeEnchantmentCard("Enemy Aura", "Static text."))
	gs.Seats[1].Battlefield = append(gs.Seats[1].Battlefield, enemyEnch)

	OnConstellationETB(gs, enemyEnch)

	// FireConstellationTriggers scans the ENTERING permanent's
	// controller's battlefield for sources — so seat 1's lack of
	// constellation sources means zero events.
	if sm_countEvents(gs, "constellation_trigger") != 0 {
		t.Fatalf("seat 1's enchantment ETB must not trigger seat 0's constellation; got %d",
			sm_countEvents(gs, "constellation_trigger"))
	}
}

// ---------------------------------------------------------------------------
// Bonus: SpellMastery slots cleanly into resolveGatedRider alongside
// the other three riders without cross-firing.
// ---------------------------------------------------------------------------

func TestSpellMasteryRider_IsolatedFromOtherGatedRiders(t *testing.T) {
	gs := sm_makeGame(t)
	// All four gate states active simultaneously.
	for i := 0; i < 8; i++ {
		gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, sm_makeCreatureInGrave("c"))
	}
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard,
		sm_makeInstantInGrave("i1"), sm_makeSorceryInGrave("s1"))
	for i := 0; i < 3; i++ {
		artifact := &Permanent{
			Card:       &Card{Name: "Mox", Owner: 0, Types: []string{"artifact"}},
			Controller: 0, Owner: 0, Flags: map[string]int{},
		}
		gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, artifact)
	}
	gs.Seats[0].Hand = nil

	src := sm_makePerm(0, sm_makeRiderCard("Plain Spell Mastery Spell",
		"Spell mastery — Draw two cards."))
	sm_runSpell(gs, src)

	if sm_countEvents(gs, "spell_mastery_rider") != 1 {
		t.Fatalf("spell_mastery_rider should fire exactly once; got %d",
			sm_countEvents(gs, "spell_mastery_rider"))
	}
	for _, other := range []string{"threshold_rider", "metalcraft_rider", "hellbent_rider"} {
		if sm_countEvents(gs, other) != 0 {
			t.Fatalf("%s must NOT fire on spell-mastery-only card; got %d",
				other, sm_countEvents(gs, other))
		}
	}
}
