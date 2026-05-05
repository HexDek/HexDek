package gameengine

import (
	"testing"
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
