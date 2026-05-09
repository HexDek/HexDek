package hat

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
	"github.com/hexdek/hexdek/internal/gameengine"
)

// cardWithKeyword builds a minimal card whose oracle text is a single
// AST keyword (e.g. "flashback"), so OracleTextLower picks it up.
func cardWithKeyword(name string, types []string, cmc int, keyword string) *gameengine.Card {
	ast := &gameast.CardAST{Name: name}
	ast.Abilities = append(ast.Abilities, &gameast.Keyword{Name: keyword})
	return newTestCardMinimal(name, types, cmc, ast)
}

// cardWithStaticText builds a minimal card whose oracle text is the
// supplied static-ability raw text, so OracleTextLower picks it up.
func cardWithStaticText(name string, types []string, cmc int, raw string) *gameengine.Card {
	ast := &gameast.CardAST{Name: name}
	ast.Abilities = append(ast.Abilities, &gameast.Static{Raw: raw})
	return newTestCardMinimal(name, types, cmc, ast)
}

func TestHasGraveyardRecursionValue(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	cases := []struct {
		name    string
		card    *gameengine.Card
		want    bool
	}{
		{"flashback", cardWithKeyword("Lingering Souls", []string{"sorcery"}, 3, "flashback"), true},
		{"unearth", cardWithKeyword("Hellspark Elemental", []string{"creature"}, 1, "unearth"), true},
		{"escape", cardWithKeyword("Underworld Breach", []string{"enchantment"}, 1, "escape"), true},
		{"disturb", cardWithKeyword("Lunarch Veteran", []string{"creature"}, 1, "disturb"), true},
		{"jump-start", cardWithKeyword("Chemister's Insight", []string{"instant"}, 4, "jump-start"), true},
		{"dredge", cardWithKeyword("Stinkweed Imp", []string{"creature"}, 3, "dredge"), true},
		{"vanilla", newTestCardMinimal("Bear", []string{"creature"}, 2, nil), false},
		{"nil", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := h.hasGraveyardRecursionValue(tc.card); got != tc.want {
				t.Fatalf("hasGraveyardRecursionValue(%s) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestHasGraveyardRecursionEnabler(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)

	if h.hasGraveyardRecursionEnabler(gs, 0) {
		t.Fatalf("empty battlefield should not register an enabler")
	}

	// Sun-Titan-style: returns a card from graveyard to the battlefield.
	titan := cardWithStaticText("Sun Titan", []string{"creature"}, 6,
		"When this creature enters or attacks, you may return target permanent card with mana value 3 or less from your graveyard to the battlefield.")
	newTestPermanent(gs.Seats[0], titan, 6, 6)

	if !h.hasGraveyardRecursionEnabler(gs, 0) {
		t.Fatalf("Sun-Titan-style permanent should register as a graveyard recursion enabler")
	}

	// Recursion enabler belongs to seat 0; opponent shouldn't see it as their enabler.
	if h.hasGraveyardRecursionEnabler(gs, 1) {
		t.Fatalf("opponent should not inherit our recursion enabler")
	}
}

// Non-reanimator deck with a flashback card in hand should prefer
// discarding the flashback card over a non-recursion vanilla creature
// of comparable value. This is the deck-agnostic case the task targets.
func TestChooseDiscard_PrefersGraveyardRecursionCards_NonReanimator(t *testing.T) {
	// Midrange strategy — explicitly NOT reanimator.
	sp := &StrategyProfile{Archetype: ArchetypeMidrange}
	h := NewYggdrasilHatWithNoise(sp, 0, 0)
	gs := newTestGame(t, 2)

	flashback := cardWithKeyword("Lingering Souls", []string{"sorcery"}, 3, "flashback")
	vanilla := newTestCardMinimal("Grizzly Bears", []string{"creature"}, 2, nil)

	hand := []*gameengine.Card{flashback, vanilla}
	got := h.ChooseDiscard(gs, 0, hand, 1)
	if len(got) != 1 {
		t.Fatalf("want 1 discard, got %d", len(got))
	}
	if got[0] != flashback {
		t.Fatalf("non-reanimator midrange deck should discard the flashback card first; got %s",
			got[0].DisplayName())
	}
}

// Surveil should send a flashback card to the graveyard for any deck
// (not gated by reanimator archetype).
func TestChooseSurveil_SendsFlashbackToGraveyard_NonReanimator(t *testing.T) {
	sp := &StrategyProfile{Archetype: ArchetypeMidrange}
	h := NewYggdrasilHatWithNoise(sp, 0, 0)
	gs := newTestGame(t, 2)

	flashback := cardWithKeyword("Lingering Souls", []string{"sorcery"}, 3, "flashback")
	keep := newTestCardMinimal("Sol Ring", []string{"artifact"}, 1, nil)
	// Two-card surveil so the empty-top fallback doesn't undo the choice.
	gy, top := h.ChooseSurveil(gs, 0, []*gameengine.Card{flashback, keep})

	foundInGy := false
	for _, c := range gy {
		if c == flashback {
			foundInGy = true
		}
	}
	if !foundInGy {
		t.Fatalf("flashback card should have been surveiled to graveyard; gy=%v top=%v", gy, top)
	}
}
