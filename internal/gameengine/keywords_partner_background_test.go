package gameengine

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// Round-27 tests for IsBackground + HasChooseABackground + the
// Background pairing rule wired through CanBeCommandersTogether.

// ---------------------------------------------------------------------------
// Test card builders
// ---------------------------------------------------------------------------

// kpb_makeBackground builds an enchantment with the Background subtype.
// Mirrors the printed CLB type-line "Legendary Enchantment — Background".
func kpb_makeBackground(name string) *Card {
	return &Card{
		Name:     name,
		Types:    []string{"legendary", "enchantment", "background"},
		TypeLine: "Legendary Enchantment — Background",
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{},
		},
	}
}

// kpb_makeChooserCommander builds a legendary creature carrying the
// "Choose a Background" keyword (the chooser side of the pair).
func kpb_makeChooserCommander(name string) *Card {
	return &Card{
		Name:          name,
		BasePower:     2,
		BaseToughness: 2,
		Types:         []string{"legendary", "creature"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "choose", Raw: "choose a background"},
			},
		},
	}
}

// kpb_makePlainCommander builds a legendary creature with no partner-
// family keyword. Used as the negative-case "regular commander".
func kpb_makePlainCommander(name string) *Card {
	return &Card{
		Name:          name,
		BasePower:     3,
		BaseToughness: 3,
		Types:         []string{"legendary", "creature"},
		AST: &gameast.CardAST{
			Name:      name,
			Abilities: []gameast.Ability{},
		},
	}
}

// ---------------------------------------------------------------------------
// Detector predicates
// ---------------------------------------------------------------------------

func TestIsBackground_Positive(t *testing.T) {
	if !IsBackground(kpb_makeBackground("Cultist of the Absolute")) {
		t.Fatal("IsBackground should be true for a card with the Background subtype")
	}
}

func TestIsBackground_NegativeOnPlainEnchantment(t *testing.T) {
	c := &Card{
		Name:  "Wrath of God",
		Types: []string{"sorcery"},
		AST:   &gameast.CardAST{Name: "Wrath of God"},
	}
	if IsBackground(c) {
		t.Fatal("IsBackground should be false for a non-background card")
	}
	if IsBackground(nil) {
		t.Fatal("IsBackground(nil) should be false")
	}
}

func TestIsBackground_TypeLineFallback(t *testing.T) {
	// Card whose Types slice missed the subtype but TypeLine includes it.
	c := &Card{
		Name:     "Raised by Giants",
		Types:    []string{"legendary", "enchantment"},
		TypeLine: "Legendary Enchantment — Background",
		AST:      &gameast.CardAST{Name: "Raised by Giants"},
	}
	if !IsBackground(c) {
		t.Fatal("IsBackground should fall back to TypeLine when Types omits the subtype")
	}
}

func TestHasChooseABackground_Positive(t *testing.T) {
	if !HasChooseABackground(kpb_makeChooserCommander("Wilson, Refined Grizzly")) {
		t.Fatal("HasChooseABackground should detect the keyword")
	}
}

func TestHasChooseABackground_NegativeOnBackground(t *testing.T) {
	// A Background enchantment does NOT itself have the chooser keyword.
	if HasChooseABackground(kpb_makeBackground("Cultist of the Absolute")) {
		t.Fatal("a Background should not report HasChooseABackground")
	}
	if HasChooseABackground(nil) {
		t.Fatal("HasChooseABackground(nil) should be false")
	}
}

// ---------------------------------------------------------------------------
// (a) ChooseABackground commander + Background enchantment = valid pair.
// ---------------------------------------------------------------------------

func TestCanBeCommandersTogether_BackgroundPairValid(t *testing.T) {
	chooser := kpb_makeChooserCommander("Wilson, Refined Grizzly")
	bg := kpb_makeBackground("Raised by Giants")
	if !CanBeCommandersTogether(chooser, bg) {
		t.Fatal("Choose-a-Background commander + Background should be a legal pair")
	}
}

// ---------------------------------------------------------------------------
// (b) Two ChooseABackground commanders alone = invalid.
// ---------------------------------------------------------------------------

func TestCanBeCommandersTogether_TwoChoosersInvalid(t *testing.T) {
	a := kpb_makeChooserCommander("Wilson, Refined Grizzly")
	b := kpb_makeChooserCommander("Faldorn, Dread Wolf Herald")
	if CanBeCommandersTogether(a, b) {
		t.Fatal("two Choose-a-Background commanders must NOT pair without a Background")
	}
}

// ---------------------------------------------------------------------------
// (c) Regular commander + Background = invalid (no choose).
// ---------------------------------------------------------------------------

func TestCanBeCommandersTogether_PlainCommanderPlusBackgroundInvalid(t *testing.T) {
	plain := kpb_makePlainCommander("Sol'kanar the Tainted")
	bg := kpb_makeBackground("Raised by Giants")
	if CanBeCommandersTogether(plain, bg) {
		t.Fatal("plain commander (no chooser keyword) + Background must NOT pair")
	}
}

// ---------------------------------------------------------------------------
// (d) Two Background enchantments = invalid (no commander spell).
// ---------------------------------------------------------------------------

func TestCanBeCommandersTogether_TwoBackgroundsInvalid(t *testing.T) {
	a := kpb_makeBackground("Raised by Giants")
	b := kpb_makeBackground("Criminal Past")
	if CanBeCommandersTogether(a, b) {
		t.Fatal("two Backgrounds alone must NOT pair (no chooser)")
	}
}

// ---------------------------------------------------------------------------
// (e) Order doesn't matter (a,b or b,a).
// ---------------------------------------------------------------------------

func TestCanBeCommandersTogether_BackgroundPairOrderIndependent(t *testing.T) {
	chooser := kpb_makeChooserCommander("Wilson, Refined Grizzly")
	bg := kpb_makeBackground("Raised by Giants")
	if !CanBeCommandersTogether(chooser, bg) {
		t.Fatal("(chooser, background) should be legal")
	}
	if !CanBeCommandersTogether(bg, chooser) {
		t.Fatal("(background, chooser) should also be legal — order independent")
	}
}

// Bonus: Background should NOT pair with a Friends-Forever, bare-Partner,
// or Partner-with card (cross-category mix forbidden).
func TestCanBeCommandersTogether_BackgroundDoesNotMixWithOtherPartnerFamilies(t *testing.T) {
	bg := kpb_makeBackground("Raised by Giants")

	bare := &Card{
		Name:  "Kraum, Ludevic's Opus",
		Types: []string{"legendary", "creature"},
		AST: &gameast.CardAST{
			Name: "Kraum, Ludevic's Opus",
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "partner", Raw: "partner"},
			},
		},
	}
	if CanBeCommandersTogether(bare, bg) {
		t.Fatal("bare Partner + Background must NOT pair")
	}

	ff := &Card{
		Name:  "Ardenn",
		Types: []string{"legendary", "creature"},
		AST: &gameast.CardAST{
			Name: "Ardenn",
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "friends", Raw: "friends forever"},
			},
		},
	}
	if CanBeCommandersTogether(ff, bg) {
		t.Fatal("Friends Forever + Background must NOT pair")
	}
}
