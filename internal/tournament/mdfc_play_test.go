package tournament

// Regression test for the MDFC back-face-land type leak.
//
// Background: MDFCs (modal double-faced cards) like "Fell the Profane //
// Fell Mire" or "Valakut Awakening // Valakut Stoneforge" appear in the
// AST corpus with a combined Scryfall type_line — e.g. "Instant // Land".
// The deckparser's parseTypes splits on whitespace, producing
// ["instant", "//", "land"] as the runtime Card.Types. That makes
// isLand() return true (correct: the back face is a land), so the
// land-play path picks the card up — but if the runtime card identity
// isn't swapped to the back face before the Permanent is created, the
// permanent on the battlefield retains the front-face "instant" or
// "sorcery" type and Feynman's permanent_types invariant flags it as a
// critical violation.
//
// The fix (gameengine.SwapToBackFace, called from tryPlayLand) replaces
// the runtime Types/Name/TypeLine/CMC with the back-face values when an
// MDFC's BACK face is the land being played.

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
	"github.com/hexdek/hexdek/internal/hat"
)

// fellTheProfaneMDFC builds a synthetic MDFC card mirroring "Fell the
// Profane // Fell Mire" — sorcery front face, land back face. This is
// the canonical shape that triggers the bug: the parser-leaked Types
// contain both "sorcery" (front) and "land" (back).
func fellTheProfaneMDFC() *gameengine.Card {
	return &gameengine.Card{
		Name: "Fell the Profane // Fell Mire",
		// Combined-face Types as parseTypes("Sorcery // Land — Swamp")
		// would produce. The "//" pseudo-token is preserved exactly as
		// the deckparser leaks it to demonstrate that the fix tolerates
		// the leaked input shape.
		Types:    []string{"sorcery", "//", "land", "swamp"},
		TypeLine: "sorcery // land — swamp",
		CMC:      4,
		Colors:   []string{"B"},
		// MDFC back-face data — the deckparser's
		// SupplementWithOracleJSON populates these for layout=modal_dfc
		// rows in oracle-cards.json.
		BackFaceName:     "Fell Mire",
		BackFaceCMC:      0,
		BackFaceTypes:    []string{"land", "swamp"},
		BackFaceTypeLine: "land — swamp",
	}
}

// valakutAwakeningMDFC builds a synthetic version of Valakut Awakening
// // Valakut Stoneforge (instant front, land back). Variant of the test
// shape so the assertion suite covers more than one MDFC class.
func valakutAwakeningMDFC() *gameengine.Card {
	return &gameengine.Card{
		Name:             "Valakut Awakening // Valakut Stoneforge",
		Types:            []string{"instant", "//", "land", "mountain"},
		TypeLine:         "instant // land — mountain",
		CMC:              3,
		Colors:           []string{"R"},
		BackFaceName:     "Valakut Stoneforge",
		BackFaceCMC:      0,
		BackFaceTypes:    []string{"land", "mountain"},
		BackFaceTypeLine: "land — mountain",
	}
}

// playMDFCLandFromHand puts `card` into seat 0's hand, gives that seat a
// GreedyHat, and invokes the land-play path. Returns the Permanent the
// land created, or nil if the play didn't happen.
func playMDFCLandFromHand(t *testing.T, card *gameengine.Card) (*gameengine.GameState, *gameengine.Permanent) {
	t.Helper()
	gs := gameengine.NewGameState(2, rand.New(rand.NewSource(1)), nil)
	seat := gs.Seats[0]
	seat.Hat = &hat.GreedyHat{}
	seat.Hand = append(seat.Hand, card)

	tryPlayLand(gs, 0)

	if len(seat.Battlefield) != 1 {
		t.Fatalf("expected 1 permanent on battlefield, got %d (hand size: %d)",
			len(seat.Battlefield), len(seat.Hand))
	}
	return gs, seat.Battlefield[0]
}

// TestMDFC_FellTheProfane_BackFaceLandHasOnlyLandTypes is the headline
// regression test for the bug. After playing "Fell the Profane //
// Fell Mire" as its land back face, the permanent must carry only the
// back face's land types — no "sorcery", no "//" pseudo-token.
func TestMDFC_FellTheProfane_BackFaceLandHasOnlyLandTypes(t *testing.T) {
	_, perm := playMDFCLandFromHand(t, fellTheProfaneMDFC())

	// 1) The Feynman invariant: no permanent type may include "instant"
	//    or "sorcery". The pre-fix permanent carried "sorcery"; the
	//    post-fix one must not.
	for _, ty := range perm.Card.Types {
		if ty == "sorcery" || ty == "instant" {
			t.Errorf("MDFC back-face land must not carry %q in its Types; got %v",
				ty, perm.Card.Types)
		}
		if ty == "//" {
			t.Errorf("the parser-leaked '//' separator must not survive the back-face swap; got %v",
				perm.Card.Types)
		}
	}

	// 2) The runtime card identity reflects the BACK face, not the
	//    deck-list "Front // Back" name.
	if perm.Card.Name != "Fell Mire" {
		t.Errorf("Permanent name should be back-face 'Fell Mire'; got %q",
			perm.Card.Name)
	}

	// 3) The runtime TypeLine matches the back face.
	if perm.Card.TypeLine != "land — swamp" {
		t.Errorf("Permanent TypeLine should be back-face 'land — swamp'; got %q",
			perm.Card.TypeLine)
	}

	// 4) The permanent IS recognized as a land by the engine.
	if !perm.IsLand() {
		t.Errorf("Permanent must be IsLand()=true after back-face swap")
	}
}

// TestMDFC_ValakutAwakening_BackFaceLandHasOnlyLandTypes — same shape
// as the Fell the Profane test but with an instant-front MDFC. This
// confirms the fix isn't keyed on a single front-face type.
func TestMDFC_ValakutAwakening_BackFaceLandHasOnlyLandTypes(t *testing.T) {
	_, perm := playMDFCLandFromHand(t, valakutAwakeningMDFC())

	for _, ty := range perm.Card.Types {
		if ty == "instant" {
			t.Errorf("MDFC back-face land must not carry 'instant'; got %v",
				perm.Card.Types)
		}
	}
	if perm.Card.Name != "Valakut Stoneforge" {
		t.Errorf("Permanent name should be back-face 'Valakut Stoneforge'; got %q",
			perm.Card.Name)
	}
	if !perm.IsLand() {
		t.Errorf("Permanent must be IsLand()=true after back-face swap")
	}
}

// reverseMDFCLand builds a synthetic reverse-MDFC card mirroring the
// Final Fantasy land cycle (Midgar/Lindblum/Ishgard/Jidoor/Zanarkand):
// front face is a printed land, back face is a sorcery. The combined
// type_line "Land // Sorcery" parses into Types=["land","//","sorcery"]
// — the leaked "sorcery" token is what trips §205.2 / Feynman's
// permanent_types invariant when the land enters the battlefield.
func reverseMDFCLand(frontName, backName string, frontTypes []string, frontTypeLine string) *gameengine.Card {
	combinedTypes := append(append([]string{}, frontTypes...), "//", "sorcery")
	return &gameengine.Card{
		Name:             frontName + " // " + backName,
		Types:            combinedTypes,
		TypeLine:         frontTypeLine + " // sorcery",
		CMC:              0,
		BackFaceName:     backName,
		BackFaceCMC:      4,
		BackFaceTypes:    []string{"sorcery"},
		BackFaceTypeLine: "sorcery",
	}
}

// TestMDFC_ReverseMDFC_FFLandsStripBackFaceSpellTypes covers the Final
// Fantasy reverse-MDFC land cycle (front=land/back=sorcery). After
// playing the land, the permanent must carry only the front-face land
// types — no leaked "sorcery" or "//" pseudo-token. Pre-fix, tryPlayLand
// only swapped forward MDFCs (gated by MDFCBackFaceIsLand) and the
// reverse-MDFC leak survived, producing 339 permanent_types violations
// in the live grinder.
func TestMDFC_ReverseMDFC_FFLandsStripBackFaceSpellTypes(t *testing.T) {
	cases := []struct {
		front, back   string
		frontTypes    []string
		frontTypeLine string
	}{
		{"Midgar, City of Mako", "Reactor Raid",
			[]string{"land"}, "land"},
		{"Lindblum, Home of Theater Ship", "Falcon's Aerial Strike",
			[]string{"land"}, "land"},
		{"Ishgard, Holy See", "Heavensward Pilgrimage",
			[]string{"land"}, "land"},
		{"Jidoor, Opera Capital", "Dancing Diva's Aria",
			[]string{"land"}, "land"},
		{"Zanarkand, Forgotten Ruins", "Final Summoning",
			[]string{"land"}, "land"},
	}
	for _, c := range cases {
		t.Run(c.front, func(t *testing.T) {
			card := reverseMDFCLand(c.front, c.back, c.frontTypes, c.frontTypeLine)
			if !gameengine.IsReverseMDFC(card) {
				t.Fatalf("test setup: %q should be reverse MDFC", c.front)
			}
			_, perm := playMDFCLandFromHand(t, card)

			for _, ty := range perm.Card.Types {
				if ty == "sorcery" || ty == "instant" {
					t.Errorf("reverse MDFC %q must not carry %q on the battlefield; got %v",
						c.front, ty, perm.Card.Types)
				}
				if ty == "//" {
					t.Errorf("reverse MDFC %q must not carry the parser-leaked '//' token; got %v",
						c.front, perm.Card.Types)
				}
			}
			if !perm.IsLand() {
				t.Errorf("reverse MDFC %q must remain IsLand()=true", c.front)
			}
		})
	}
}

// TestMDFC_BackFaceSwap_DoesNotAffectVanillaLand — a sanity check that
// the swap logic only fires for actual MDFCs. A plain basic land in
// hand must enter the battlefield with its front-face identity intact;
// no spurious swap, no attempt to read empty BackFace fields.
func TestMDFC_BackFaceSwap_DoesNotAffectVanillaLand(t *testing.T) {
	plains := &gameengine.Card{
		Name:     "Plains",
		Types:    []string{"basic", "land", "plains"},
		TypeLine: "basic land — plains",
		Colors:   []string{"W"},
	}
	_, perm := playMDFCLandFromHand(t, plains)

	if perm.Card.Name != "Plains" {
		t.Errorf("vanilla Plains must keep its name; got %q", perm.Card.Name)
	}
	if !perm.IsLand() {
		t.Errorf("vanilla Plains must IsLand()=true")
	}
	// Types unchanged.
	for _, want := range []string{"basic", "land", "plains"} {
		found := false
		for _, ty := range perm.Card.Types {
			if ty == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("vanilla Plains must keep type %q; got %v", want, perm.Card.Types)
		}
	}
}
