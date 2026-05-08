package gameengine

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// TestRemoveCardFromAllPrivateZones_Sweep verifies the sweep collapses a
// duplicated card pointer down to a single absence — the core fix for the
// reanimate-into-battlefield zone_accounting overcount documented in
// docs/zone-accounting-analysis.md (hypothesis #1).
func TestRemoveCardFromAllPrivateZones_Sweep(t *testing.T) {
	gs := &GameState{Seats: []*Seat{{}, {}}}
	c := &Card{Name: "Reanimate Target", Owner: 0}

	// Simulate the duplication path: card is in graveyard AND has been
	// re-graveyarded by a MoveCard("graveyard"→"battlefield") that fell
	// through moveToZone's default. We model the duplicate by inserting
	// the same pointer twice.
	gs.Seats[0].Graveyard = []*Card{c, c}

	swept := RemoveCardFromAllPrivateZones(gs, 0, c)
	if swept != 2 {
		t.Errorf("swept count: got %d, want 2", swept)
	}
	if len(gs.Seats[0].Graveyard) != 0 {
		t.Errorf("graveyard not emptied: %v", gs.Seats[0].Graveyard)
	}
}

// TestRemoveCardFromAllPrivateZones_AcrossZones confirms the sweep pulls
// the same pointer out of every private zone in a single call.
func TestRemoveCardFromAllPrivateZones_AcrossZones(t *testing.T) {
	gs := &GameState{Seats: []*Seat{{}}}
	c := &Card{Name: "Multi-Zone", Owner: 0}
	other := &Card{Name: "Other", Owner: 0}

	gs.Seats[0].Hand = []*Card{c, other}
	gs.Seats[0].Library = []*Card{other, c, other}
	gs.Seats[0].Graveyard = []*Card{c}
	gs.Seats[0].Exile = []*Card{c, other}
	gs.Seats[0].CommandZone = []*Card{c}

	swept := RemoveCardFromAllPrivateZones(gs, 0, c)
	if swept != 5 {
		t.Errorf("swept: got %d, want 5", swept)
	}
	for name, slice := range map[string][]*Card{
		"hand":     gs.Seats[0].Hand,
		"library":  gs.Seats[0].Library,
		"gy":       gs.Seats[0].Graveyard,
		"exile":    gs.Seats[0].Exile,
		"cmd_zone": gs.Seats[0].CommandZone,
	} {
		for _, x := range slice {
			if x == c {
				t.Errorf("%s still contains the swept card", name)
			}
		}
	}
	// The "other" pointer must survive untouched in each zone where it lived.
	if len(gs.Seats[0].Hand) != 1 || gs.Seats[0].Hand[0] != other {
		t.Errorf("hand: bystander not preserved: %v", gs.Seats[0].Hand)
	}
	if len(gs.Seats[0].Library) != 2 {
		t.Errorf("library: bystanders not preserved: %v", gs.Seats[0].Library)
	}
}

// TestMoveToZone_LibraryArm verifies that a destination of "library" no
// longer falls through to graveyard. Two per-card hooks (nine_fingers_keene,
// runo_stromkirk) emit toZone="library" and the previous default-case
// fallthrough silently sent those cards into the graveyard instead.
func TestMoveToZone_LibraryArm(t *testing.T) {
	gs := &GameState{Seats: []*Seat{{}}}
	c := &Card{Name: "On-A-Trip", Owner: 0}

	gs.moveToZone(0, c, "library")
	if len(gs.Seats[0].Library) != 1 || gs.Seats[0].Library[0] != c {
		t.Errorf("library arm: card not in library: lib=%v gy=%v", gs.Seats[0].Library, gs.Seats[0].Graveyard)
	}
	if len(gs.Seats[0].Graveyard) != 0 {
		t.Errorf("library arm: card leaked to graveyard: %v", gs.Seats[0].Graveyard)
	}
}

// TestMoveToZone_CommandZoneArm — same idea for explicit command_zone moves.
func TestMoveToZone_CommandZoneArm(t *testing.T) {
	gs := &GameState{Seats: []*Seat{{}}}
	c := &Card{Name: "Commander Stuff", Owner: 0}

	gs.moveToZone(0, c, "command_zone")
	if len(gs.Seats[0].CommandZone) != 1 {
		t.Errorf("command_zone arm: card not placed: %v", gs.Seats[0].CommandZone)
	}
	if len(gs.Seats[0].Graveyard) != 0 {
		t.Errorf("command_zone arm: card leaked to graveyard: %v", gs.Seats[0].Graveyard)
	}
}

// TestMoveToZone_BattlefieldArm verifies that a destination of "battlefield"
// creates a Permanent wrapper and appends to Battlefield instead of falling
// through to the graveyard default. This was the root cause of ~80% of
// zone_accounting Feynman violations.
func TestMoveToZone_BattlefieldArm(t *testing.T) {
	gs := &GameState{Seats: []*Seat{{}}}
	c := &Card{Name: "Grizzly Bears", Owner: 0, Types: []string{"creature"}}
	c.AST = &gameast.CardAST{Name: "Grizzly Bears"}

	gs.moveToZone(0, c, "battlefield")
	if len(gs.Seats[0].Battlefield) != 1 {
		t.Fatalf("battlefield arm: card not on battlefield: bf=%d gy=%d",
			len(gs.Seats[0].Battlefield), len(gs.Seats[0].Graveyard))
	}
	if gs.Seats[0].Battlefield[0].Card != c {
		t.Error("battlefield arm: permanent wraps wrong card")
	}
	if len(gs.Seats[0].Graveyard) != 0 {
		t.Errorf("battlefield arm: card leaked to graveyard: %v", gs.Seats[0].Graveyard)
	}
}

// TestMoveToZone_BattlefieldTapped verifies the "battlefield_tapped" variant
// creates a tapped Permanent.
func TestMoveToZone_BattlefieldTapped(t *testing.T) {
	gs := &GameState{Seats: []*Seat{{}}}
	c := &Card{Name: "Temple of Silence", Owner: 0, Types: []string{"land"}}
	c.AST = &gameast.CardAST{Name: "Temple of Silence"}

	gs.moveToZone(0, c, "battlefield_tapped")
	if len(gs.Seats[0].Battlefield) != 1 {
		t.Fatalf("battlefield_tapped: card not on battlefield: bf=%d gy=%d",
			len(gs.Seats[0].Battlefield), len(gs.Seats[0].Graveyard))
	}
	perm := gs.Seats[0].Battlefield[0]
	if !perm.Tapped {
		t.Error("battlefield_tapped: permanent should be tapped")
	}
	if len(gs.Seats[0].Graveyard) != 0 {
		t.Errorf("battlefield_tapped: card leaked to graveyard")
	}
}

// TestMoveToZone_BattlefieldMDFC verifies that an MDFC with a land back face
// gets its types swapped to the land face when entering the battlefield via
// moveToZone. This is the core fix for MDFC permanent_types back-face
// resolution (Fell the Profane // Fell Mire, Valakut Awakening, etc.).
func TestMoveToZone_BattlefieldMDFC(t *testing.T) {
	gs := &GameState{Seats: []*Seat{{}}}
	c := &Card{
		Name:             "Valakut Awakening",
		Owner:            0,
		Types:            []string{"instant"},
		TypeLine:         "Instant // Land",
		BackFaceName:     "Valakut Stoneforge",
		BackFaceTypes:    []string{"land"},
		BackFaceTypeLine: "Land",
	}
	c.AST = &gameast.CardAST{Name: "Valakut Awakening"}

	gs.moveToZone(0, c, "battlefield")
	if len(gs.Seats[0].Battlefield) != 1 {
		t.Fatalf("MDFC battlefield: card not on battlefield: bf=%d gy=%d",
			len(gs.Seats[0].Battlefield), len(gs.Seats[0].Graveyard))
	}
	perm := gs.Seats[0].Battlefield[0]
	// After EnsureBattlefieldFrontFace, the card's types should be the
	// back face's (land), not the front face's (instant).
	hasLand := false
	for _, ty := range perm.Card.Types {
		if ty == "land" {
			hasLand = true
		}
		if ty == "instant" {
			t.Error("MDFC battlefield: front-face 'instant' type leaked onto battlefield")
		}
	}
	if !hasLand {
		t.Errorf("MDFC battlefield: back-face 'land' type not present, got: %v", perm.Card.Types)
	}
	if len(gs.Seats[0].Graveyard) != 0 {
		t.Errorf("MDFC battlefield: card leaked to graveyard")
	}
}

// TestMoveToZone_BattlefieldRejectsNonPermanent verifies that an instant or
// sorcery that somehow reaches moveToZone("battlefield") is redirected to
// graveyard per CR 304.4 / 307.1, NOT silently placed as a Permanent.
func TestMoveToZone_BattlefieldRejectsNonPermanent(t *testing.T) {
	gs := &GameState{Seats: []*Seat{{}}}
	c := &Card{Name: "Lightning Bolt", Owner: 0, Types: []string{"instant"}}
	c.AST = &gameast.CardAST{Name: "Lightning Bolt"}

	gs.moveToZone(0, c, "battlefield")
	if len(gs.Seats[0].Battlefield) != 0 {
		t.Error("non-permanent should not be on battlefield")
	}
	if len(gs.Seats[0].Graveyard) != 1 {
		t.Errorf("non-permanent should be in graveyard, got gy=%d", len(gs.Seats[0].Graveyard))
	}
}

// TestMoveCard_BattlefieldIntegration exercises the full MoveCard path to
// "battlefield" — the actual call pattern used by cheat-into-play effects.
func TestMoveCard_BattlefieldIntegration(t *testing.T) {
	gs := &GameState{Seats: []*Seat{{
		Hand: make([]*Card, 0),
	}}}
	c := &Card{Name: "Emrakul", Owner: 0, Types: []string{"creature"}}
	c.AST = &gameast.CardAST{Name: "Emrakul, the Aeons Torn"}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, c)

	result := MoveCard(gs, c, 0, "hand", "battlefield", "cheat_creature")
	if result.FinalZone != "battlefield" {
		t.Errorf("MoveCard returned dest=%q, want battlefield", result.FinalZone)
	}
	// Card should be on battlefield, not in hand or graveyard.
	if len(gs.Seats[0].Battlefield) != 1 {
		t.Fatalf("MoveCard battlefield: not on battlefield: bf=%d gy=%d hand=%d",
			len(gs.Seats[0].Battlefield), len(gs.Seats[0].Graveyard), len(gs.Seats[0].Hand))
	}
	if gs.Seats[0].Battlefield[0].Card != c {
		t.Error("MoveCard battlefield: wrong card on battlefield")
	}
	if len(gs.Seats[0].Hand) != 0 {
		t.Errorf("MoveCard battlefield: card still in hand")
	}
	if len(gs.Seats[0].Graveyard) != 0 {
		t.Errorf("MoveCard battlefield: card leaked to graveyard")
	}
}
