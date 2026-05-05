package gameengine

// Odin Golden-Game Oracle — modal/choice spell test suite.
//
// Deterministic, state-based golden assertions for CR §601.2c modal-spell
// resolution. Three canonical shapes are exercised:
//
//   1. Cryptic Command — "Choose two —" with four mutually exclusive modes.
//      Each of the four modes is selected in isolation (Pick=1) and the
//      resolver's side effects are verified against the rules text:
//        a. Counter target spell.
//        b. Return target permanent to its owner's hand.
//        c. Tap all creatures your opponents control.
//        d. Draw a card.
//
//   2. Abzan Charm — "Choose one —" with three modes, one of which is a
//      composite Sequence (Draw + LoseLife):
//        a. Exile target creature with power 3 or greater.
//        b. Put two +1/+1 counters on target creature.
//        c. Draw two cards, then you lose 2 life.
//
//   3. Mystic Confluence — "Choose three. You may choose the same mode more
//      than once." The engine's resolveChoice resolves a single mode pick per
//      Choice node (CR §601.2c is modeled as one ChooseMode call); we model
//      "choose three, may repeat" by wrapping THREE Choice nodes in a
//      Sequence and pinning the test Hat to the same mode index for all
//      three picks. The golden assertion is that the same-mode resolution
//      stacks N times — the Confluence Draw mode chosen 3× draws 3 cards.
//
// All tests are RNG-free and don't run a turn loop. Mode selection is
// pinned via a small `modePickerHat` that returns a fixed index from
// ChooseMode. The Hat's other Hat-interface methods inherit GreedyHatStub's
// no-op defaults (defined in triggers_test.go).

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// modePickerHat is a deterministic test Hat that always returns a fixed
// mode index for ChooseMode calls. Used to pin §601.2c mode selection so
// the test exercises a specific branch of a Choice node.
type modePickerHat struct {
	GreedyHatStub
	pick int
}

func (h *modePickerHat) ChooseMode(_ *GameState, _ int, modes []gameast.Effect) int {
	if h.pick < 0 || h.pick >= len(modes) {
		return 0
	}
	return h.pick
}

// installModeHat sets the mode-pinning Hat on the given seat. Returns the
// hat so the test can introspect it if needed.
func installModeHat(gs *GameState, seat, pick int) *modePickerHat {
	h := &modePickerHat{pick: pick}
	gs.Seats[seat].Hat = h
	return h
}

// crypticCommandModes returns the 4 Cryptic Command mode AST nodes in
// canonical printed order:
//
//	0: Counter target spell.
//	1: Return target permanent to its owner's hand.
//	2: Tap all creatures your opponents control.
//	3: Draw a card.
//
// Constructed once and reused across the 4 Cryptic Command tests.
func crypticCommandModes() []gameast.Effect {
	return []gameast.Effect{
		// Mode 0 — Counter target spell.
		&gameast.CounterSpell{
			Target: gameast.Filter{Base: "spell", Targeted: true},
		},
		// Mode 1 — Return target permanent to its owner's hand.
		&gameast.Bounce{
			Target: gameast.Filter{
				Base: "permanent", OpponentControls: true, Targeted: true,
			},
			To: "owners_hand",
		},
		// Mode 2 — Tap all creatures your opponents control. Uses
		// Quantifier="each" to fan out across all opponent battlefields
		// per CR §608.2 (untargeted "each" effects).
		&gameast.TapEffect{
			Target: gameast.Filter{
				Base:             "creature",
				OpponentControls: true,
				Quantifier:       "each",
			},
		},
		// Mode 3 — Draw a card.
		&gameast.Draw{
			Count:  *gameast.NumInt(1),
			Target: gameast.Filter{Base: "controller"},
		},
	}
}

// -----------------------------------------------------------------------------
// Cryptic Command — Mode 0: Counter target spell.
// -----------------------------------------------------------------------------

func TestOdin_CrypticCommand_Mode0_CounterTargetSpell(t *testing.T) {
	gs := newFixtureGame(t)
	installModeHat(gs, 0, 0) // pin Hat to mode 0 (counter)

	src := addBattlefield(gs, 0, "Cryptic Command", 0, 0, "instant")

	// Opponent puts a creature spell on the stack.
	oppSpell := &StackItem{
		ID:         42,
		Controller: 1,
		Kind:       "spell",
		Card:       &Card{Name: "Phyrexian Dreadnought", Types: []string{"creature"}},
	}
	gs.Stack = append(gs.Stack, oppSpell)

	choice := &gameast.Choice{
		Pick:    *gameast.NumInt(1),
		Options: crypticCommandModes(),
	}
	ResolveEffect(gs, src, choice)

	// Mode 0 fired: target spell countered.
	if !oppSpell.Countered {
		t.Fatalf("mode 0 (counter): opponent spell should be countered")
	}
	if countEvents(gs, "counter_spell") != 1 {
		t.Errorf("expected 1 counter_spell event, got %d", countEvents(gs, "counter_spell"))
	}

	// Negative-space: none of the other three modes should have fired.
	// No bounce → opponent has no card in hand.
	if len(gs.Seats[1].Hand) != 0 {
		t.Errorf("mode 0 should not bounce: opponent hand=%d", len(gs.Seats[1].Hand))
	}
	// No tap → no creatures on board to tap; assert no tap events emitted.
	if countEvents(gs, "tap") != 0 {
		t.Errorf("mode 0 should not tap: tap events=%d", countEvents(gs, "tap"))
	}
	// No draw → controller hand still empty.
	if len(gs.Seats[0].Hand) != 0 {
		t.Errorf("mode 0 should not draw: controller hand=%d", len(gs.Seats[0].Hand))
	}
}

// -----------------------------------------------------------------------------
// Cryptic Command — Mode 1: Return target permanent to owner's hand.
// -----------------------------------------------------------------------------

func TestOdin_CrypticCommand_Mode1_ReturnPermanent(t *testing.T) {
	gs := newFixtureGame(t)
	installModeHat(gs, 0, 1) // pin Hat to mode 1 (bounce)

	src := addBattlefield(gs, 0, "Cryptic Command", 0, 0, "instant")
	target := addBattlefield(gs, 1, "Llanowar Elves", 1, 1, "creature")
	target.Card.Owner = 1

	choice := &gameast.Choice{
		Pick:    *gameast.NumInt(1),
		Options: crypticCommandModes(),
	}
	ResolveEffect(gs, src, choice)

	// Mode 1 fired: target returned to owner's hand.
	if len(gs.Seats[1].Battlefield) != 0 {
		t.Errorf("mode 1: opponent battlefield should be empty, got %d",
			len(gs.Seats[1].Battlefield))
	}
	if len(gs.Seats[1].Hand) != 1 {
		t.Errorf("mode 1: opponent hand should have 1 card, got %d",
			len(gs.Seats[1].Hand))
	}

	// Negative-space.
	if len(gs.Seats[0].Hand) != 0 {
		t.Errorf("mode 1 should not draw: controller hand=%d", len(gs.Seats[0].Hand))
	}
	if countEvents(gs, "counter_spell") != 0 {
		t.Errorf("mode 1 should not counter: counter events=%d",
			countEvents(gs, "counter_spell"))
	}
	// Source itself untapped (mode 2 didn't fire).
	if src.Tapped {
		t.Errorf("mode 1 should not have tapped source")
	}
	if target.Tapped {
		t.Errorf("mode 1 should not tap; bounced target was tapped")
	}
}

// -----------------------------------------------------------------------------
// Cryptic Command — Mode 2: Tap all creatures your opponents control.
// -----------------------------------------------------------------------------

func TestOdin_CrypticCommand_Mode2_TapAllOpponentCreatures(t *testing.T) {
	gs := newFixtureGame(t)
	installModeHat(gs, 0, 2) // pin Hat to mode 2 (tap-all)

	src := addBattlefield(gs, 0, "Cryptic Command", 0, 0, "instant")

	// Three opponent creatures + one of YOUR creatures (must remain untapped).
	oppA := addBattlefield(gs, 1, "Goblin Guide", 2, 2, "creature")
	oppB := addBattlefield(gs, 1, "Lava Spike Goblin", 1, 1, "creature")
	oppC := addBattlefield(gs, 1, "Dragon", 4, 4, "creature")
	mine := addBattlefield(gs, 0, "Llanowar Elves", 1, 1, "creature")

	choice := &gameast.Choice{
		Pick:    *gameast.NumInt(1),
		Options: crypticCommandModes(),
	}
	ResolveEffect(gs, src, choice)

	// Mode 2 fired: all opponent creatures tapped.
	if !oppA.Tapped || !oppB.Tapped || !oppC.Tapped {
		t.Errorf("mode 2: all opponent creatures should be tapped (got A=%v B=%v C=%v)",
			oppA.Tapped, oppB.Tapped, oppC.Tapped)
	}
	// Symmetry check: your own creatures must NOT be tapped (filter is
	// OpponentControls=true).
	if mine.Tapped {
		t.Errorf("mode 2: your own creatures must not be tapped")
	}

	// Negative-space.
	if len(gs.Seats[1].Hand) != 0 {
		t.Errorf("mode 2 should not bounce: opponent hand=%d", len(gs.Seats[1].Hand))
	}
	if len(gs.Seats[0].Hand) != 0 {
		t.Errorf("mode 2 should not draw: controller hand=%d", len(gs.Seats[0].Hand))
	}
}

// -----------------------------------------------------------------------------
// Cryptic Command — Mode 3: Draw a card.
// -----------------------------------------------------------------------------

func TestOdin_CrypticCommand_Mode3_DrawACard(t *testing.T) {
	gs := newFixtureGame(t)
	installModeHat(gs, 0, 3) // pin Hat to mode 3 (draw)

	src := addBattlefield(gs, 0, "Cryptic Command", 0, 0, "instant")
	addLibrary(gs, 0, "Brainstorm")

	choice := &gameast.Choice{
		Pick:    *gameast.NumInt(1),
		Options: crypticCommandModes(),
	}
	ResolveEffect(gs, src, choice)

	// Mode 3 fired: 1 card drawn.
	if len(gs.Seats[0].Hand) != 1 {
		t.Errorf("mode 3: expected 1 card in hand, got %d", len(gs.Seats[0].Hand))
	}
	if got := gs.Seats[0].Hand[0].Name; got != "Brainstorm" {
		t.Errorf("mode 3: wrong card drawn; got %q", got)
	}
	if countEvents(gs, "draw") != 1 {
		t.Errorf("mode 3: expected 1 draw event, got %d", countEvents(gs, "draw"))
	}

	// Negative-space.
	if countEvents(gs, "counter_spell") != 0 {
		t.Errorf("mode 3 should not counter")
	}
	if countEvents(gs, "tap") != 0 {
		t.Errorf("mode 3 should not tap")
	}
}

// -----------------------------------------------------------------------------
// Abzan Charm — three modes, "choose one".
// -----------------------------------------------------------------------------

// abzanCharmModes returns the 3 Abzan Charm mode AST nodes in canonical
// printed order:
//
//	0: Exile target creature with power 3 or greater.
//	1: Put two +1/+1 counters on target creature.
//	2: Draw two cards, then you lose 2 life.
func abzanCharmModes() []gameast.Effect {
	return []gameast.Effect{
		// Mode 0 — Exile target creature with power 3 or greater. The
		// "power >= 3" predicate isn't a first-class field on
		// gameast.Filter; the test sets up exactly one legal target
		// (a 5/5 opponent creature) so PickTarget's MVP heuristic
		// resolves to that creature regardless. Filter.Extra carries
		// the rules-text adjective for forward compatibility.
		&gameast.Exile{
			Target: gameast.Filter{
				Base:             "creature",
				OpponentControls: true,
				Targeted:         true,
				Extra:            []string{"power_ge_3"},
			},
		},
		// Mode 1 — Put two +1/+1 counters on target creature.
		&gameast.CounterMod{
			Op:          "put",
			Count:       *gameast.NumInt(2),
			CounterKind: "+1/+1",
			Target: gameast.Filter{
				Base: "creature", YouControl: true, Targeted: true,
			},
		},
		// Mode 2 — Draw two cards, then you lose 2 life. The "then" makes
		// this a Sequence: Draw 2, then LoseLife 2 (controller).
		&gameast.Sequence{
			Items: []gameast.Effect{
				&gameast.Draw{
					Count:  *gameast.NumInt(2),
					Target: gameast.Filter{Base: "controller"},
				},
				&gameast.LoseLife{
					Amount: *gameast.NumInt(2),
					Target: gameast.Filter{Base: "controller"},
				},
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Abzan Charm — Mode 0: Exile target creature with power 3 or greater.
// -----------------------------------------------------------------------------

func TestOdin_AbzanCharm_Mode0_ExileBigCreature(t *testing.T) {
	gs := newFixtureGame(t)
	installModeHat(gs, 0, 0)

	src := addBattlefield(gs, 0, "Abzan Charm", 0, 0, "instant")
	// Sole legal target — power 5 ≥ 3, so the predicate is satisfied.
	bigGuy := addBattlefield(gs, 1, "Tarmogoyf", 5, 5, "creature")

	// Sanity precondition: target's power is >=3 per the rules text.
	if bigGuy.Power() < 3 {
		t.Fatalf("test setup: target power must be >=3, got %d", bigGuy.Power())
	}

	choice := &gameast.Choice{
		Pick:    *gameast.NumInt(1),
		Options: abzanCharmModes(),
	}
	ResolveEffect(gs, src, choice)

	// Mode 0 fired: target moved from battlefield to exile.
	if len(gs.Seats[1].Battlefield) != 0 {
		t.Errorf("mode 0: opponent battlefield should be empty, got %d",
			len(gs.Seats[1].Battlefield))
	}
	if len(gs.Seats[1].Exile) != 1 {
		t.Errorf("mode 0: opponent exile should have 1 card, got %d",
			len(gs.Seats[1].Exile))
	}
	if len(gs.Seats[1].Graveyard) != 0 {
		t.Errorf("mode 0: exile (not destroy) — graveyard should be empty, got %d",
			len(gs.Seats[1].Graveyard))
	}

	// Negative-space.
	if len(gs.Seats[0].Hand) != 0 {
		t.Errorf("mode 0 should not draw")
	}
	if gs.Seats[0].Life != 20 {
		t.Errorf("mode 0 should not change life; got %d", gs.Seats[0].Life)
	}
}

// -----------------------------------------------------------------------------
// Abzan Charm — Mode 1: Put two +1/+1 counters on target creature.
// -----------------------------------------------------------------------------

func TestOdin_AbzanCharm_Mode1_PutTwoP1P1Counters(t *testing.T) {
	gs := newFixtureGame(t)
	installModeHat(gs, 0, 1)

	src := addBattlefield(gs, 0, "Abzan Charm", 0, 0, "instant")
	target := addBattlefield(gs, 0, "Anafenza, the Foremost", 4, 4, "creature")

	choice := &gameast.Choice{
		Pick:    *gameast.NumInt(1),
		Options: abzanCharmModes(),
	}
	ResolveEffect(gs, src, choice)

	// Mode 1 fired: target now has 2 +1/+1 counters → 6/6.
	if got := target.Counters["+1/+1"]; got != 2 {
		t.Errorf("mode 1: expected 2 +1/+1 counters, got %d", got)
	}
	if target.Power() != 6 || target.Toughness() != 6 {
		t.Errorf("mode 1: expected 6/6 after 2 counters, got %d/%d",
			target.Power(), target.Toughness())
	}

	// Negative-space.
	if len(gs.Seats[0].Hand) != 0 {
		t.Errorf("mode 1 should not draw")
	}
	if gs.Seats[0].Life != 20 {
		t.Errorf("mode 1 should not change life; got %d", gs.Seats[0].Life)
	}
	if len(gs.Seats[1].Exile) != 0 {
		t.Errorf("mode 1 should not exile")
	}
}

// -----------------------------------------------------------------------------
// Abzan Charm — Mode 2: Draw two cards, then you lose 2 life.
// -----------------------------------------------------------------------------

func TestOdin_AbzanCharm_Mode2_DrawTwoLoseTwo(t *testing.T) {
	gs := newFixtureGame(t)
	installModeHat(gs, 0, 2)

	src := addBattlefield(gs, 0, "Abzan Charm", 0, 0, "instant")
	addLibrary(gs, 0, "Mountain", "Forest", "Plains") // 3 cards available

	choice := &gameast.Choice{
		Pick:    *gameast.NumInt(1),
		Options: abzanCharmModes(),
	}
	ResolveEffect(gs, src, choice)

	// Mode 2 fired: 2 cards drawn AND 2 life lost.
	if len(gs.Seats[0].Hand) != 2 {
		t.Errorf("mode 2: expected 2 cards drawn, got %d", len(gs.Seats[0].Hand))
	}
	if len(gs.Seats[0].Library) != 1 {
		t.Errorf("mode 2: expected 1 card left in library, got %d",
			len(gs.Seats[0].Library))
	}
	if gs.Seats[0].Life != 18 {
		t.Errorf("mode 2: expected controller life 18 (20-2), got %d",
			gs.Seats[0].Life)
	}
	// "Draw, THEN lose": Sequence ordering — draw event must precede
	// lose_life event in the log.
	drawIdx, loseIdx := -1, -1
	for i, ev := range gs.EventLog {
		if ev.Kind == "draw" && drawIdx == -1 {
			drawIdx = i
		}
		if ev.Kind == "lose_life" && loseIdx == -1 {
			loseIdx = i
		}
	}
	if drawIdx == -1 {
		t.Fatalf("mode 2: missing draw event")
	}
	if loseIdx == -1 {
		t.Fatalf("mode 2: missing lose_life event")
	}
	if loseIdx <= drawIdx {
		t.Errorf("mode 2: 'then' ordering — lose_life (%d) must come AFTER draw (%d)",
			loseIdx, drawIdx)
	}

	// Negative-space.
	if len(gs.Seats[1].Exile) != 0 {
		t.Errorf("mode 2 should not exile")
	}
}

// -----------------------------------------------------------------------------
// Mystic Confluence — choose three, may choose the same mode more than once.
// -----------------------------------------------------------------------------

// mysticConfluenceModes returns the 3 Mystic Confluence mode AST nodes:
//
//	0: Counter target spell unless its controller pays {1}.
//	1: Return target creature to its owner's hand.
//	2: Draw a card.
func mysticConfluenceModes() []gameast.Effect {
	return []gameast.Effect{
		// Mode 0 — Counter target spell unless its controller pays {1}.
		&gameast.CounterSpell{
			Target: gameast.Filter{Base: "spell", Targeted: true},
			Unless: &gameast.Cost{
				Mana: &gameast.ManaCost{Symbols: []gameast.ManaSymbol{
					{Raw: "{1}", Generic: 1},
				}},
			},
		},
		// Mode 1 — Return target creature to its owner's hand.
		&gameast.Bounce{
			Target: gameast.Filter{
				Base: "creature", OpponentControls: true, Targeted: true,
			},
			To: "owners_hand",
		},
		// Mode 2 — Draw a card.
		&gameast.Draw{
			Count:  *gameast.NumInt(1),
			Target: gameast.Filter{Base: "controller"},
		},
	}
}

// TestOdin_MysticConfluence_DrawModeChosenThrice models "choose three. You
// may choose the same mode more than once" by sequencing three independent
// Choice resolutions (the engine's resolveChoice picks a single mode per
// Choice node — see CR §601.2c modeling note in the file header). Pinning
// the Hat to mode 2 (Draw) for all three picks must yield three separate
// draws, exactly equivalent to printing "draw a card. draw a card. draw a
// card." after mode resolution.
func TestOdin_MysticConfluence_DrawModeChosenThrice(t *testing.T) {
	gs := newFixtureGame(t)
	installModeHat(gs, 0, 2) // mode 2 = Draw

	src := addBattlefield(gs, 0, "Mystic Confluence", 0, 0, "instant")
	addLibrary(gs, 0, "Top1", "Top2", "Top3", "Top4") // 4 cards; need 3

	// Three independent Choice nodes — the same mode (Draw) is chosen for
	// each, modeling Confluence's "may choose the same mode more than once."
	makeChoice := func() *gameast.Choice {
		return &gameast.Choice{
			Pick:    *gameast.NumInt(1),
			Options: mysticConfluenceModes(),
		}
	}
	seq := &gameast.Sequence{
		Items: []gameast.Effect{
			makeChoice(),
			makeChoice(),
			makeChoice(),
		},
	}
	ResolveEffect(gs, src, seq)

	// All three picks selected the Draw mode → 3 cards drawn.
	if got := len(gs.Seats[0].Hand); got != 3 {
		t.Fatalf("expected 3 cards drawn (same mode picked 3×), got %d", got)
	}
	if got := len(gs.Seats[0].Library); got != 1 {
		t.Errorf("expected 1 card left in library, got %d", got)
	}
	// Order preserved: top of library is the first-drawn card.
	wantOrder := []string{"Top1", "Top2", "Top3"}
	for i, want := range wantOrder {
		if got := gs.Seats[0].Hand[i].Name; got != want {
			t.Errorf("hand[%d]: expected %q, got %q", i, want, got)
		}
	}
	// And exactly 3 draw events fired (one per Choice resolution).
	if got := countEvents(gs, "draw"); got != 3 {
		t.Errorf("expected 3 draw events, got %d", got)
	}

	// Negative-space: counter and bounce modes did NOT fire. No spells on
	// the stack to counter, no creatures to bounce — but more importantly,
	// no events of those kinds should exist.
	if got := countEvents(gs, "counter_spell"); got != 0 {
		t.Errorf("mode 0 (counter) should not have fired; got %d events", got)
	}
	// A successful bounce produces a zone_change event with kind "bounce"-
	// adjacent semantics, but the safest signal is the opponent's hand
	// remaining empty.
	if got := len(gs.Seats[1].Hand); got != 0 {
		t.Errorf("mode 1 (bounce) should not have fired; opp hand=%d", got)
	}
}

// TestOdin_MysticConfluence_DistinctModesAcrossPicks is the symmetric
// negative — three Choice nodes, each pinned to a DIFFERENT mode index,
// must each resolve their pinned mode independently. Verifies that
// repeated Choice resolution doesn't share or leak state across picks.
func TestOdin_MysticConfluence_DistinctModesAcrossPicks(t *testing.T) {
	gs := newFixtureGame(t)
	src := addBattlefield(gs, 0, "Mystic Confluence", 0, 0, "instant")
	addLibrary(gs, 0, "Brainstorm")

	// Set up: an opponent spell on the stack (for mode 0) AND an opponent
	// creature on the battlefield (for mode 1).
	oppSpell := &StackItem{
		ID: 7, Controller: 1, Kind: "spell",
		Card: &Card{Name: "Snapcaster Mage", Types: []string{"creature"}},
	}
	gs.Stack = append(gs.Stack, oppSpell)
	oppCreature := addBattlefield(gs, 1, "Llanowar Elves", 1, 1, "creature")
	oppCreature.Card.Owner = 1

	// Resolve three Choices, each with a different pinned mode. We swap
	// the Hat between calls so each pick lands on a fresh mode index.
	picks := []int{0, 1, 2}
	for _, pick := range picks {
		installModeHat(gs, 0, pick)
		ch := &gameast.Choice{
			Pick:    *gameast.NumInt(1),
			Options: mysticConfluenceModes(),
		}
		ResolveEffect(gs, src, ch)
	}

	// Mode 0 fired once → opponent spell countered. (The "unless pay {1}"
	// branch requires the opponent to actively pay; the test setup gives
	// the opponent no untapped mana sources, so the counter resolves.)
	if !oppSpell.Countered {
		t.Errorf("mode 0 should have countered opponent spell")
	}
	// Mode 1 fired once → opponent creature in hand, off the battlefield.
	if len(gs.Seats[1].Battlefield) != 0 {
		t.Errorf("mode 1: opponent battlefield should be empty, got %d",
			len(gs.Seats[1].Battlefield))
	}
	if len(gs.Seats[1].Hand) != 1 {
		t.Errorf("mode 1: opponent hand should have 1 card, got %d",
			len(gs.Seats[1].Hand))
	}
	// Mode 2 fired once → 1 card drawn.
	if len(gs.Seats[0].Hand) != 1 {
		t.Errorf("mode 2: controller hand should have 1 card, got %d",
			len(gs.Seats[0].Hand))
	}
}
