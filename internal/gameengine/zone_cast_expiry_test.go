package gameengine

import "testing"

func TestExpireZoneCastGrants_UntilEndOfTurn(t *testing.T) {
	gs := &GameState{
		Turn:  3,
		Seats: []*Seat{{Idx: 0, Life: 40}},
		Flags: map[string]int{},
		ZoneCastGrants: map[*Card]*ZoneCastPermission{},
	}
	card := &Card{Name: "Exiled Spell", Owner: 0}
	gs.ZoneCastGrants[card] = &ZoneCastPermission{
		Zone:       ZoneExile,
		ManaCost:   -1,
		Duration:   "until_end_of_turn",
		GrantTurn:  3,
	}

	if len(gs.ZoneCastGrants) != 1 {
		t.Fatal("grant should exist before cleanup")
	}

	ExpireZoneCastGrants(gs)

	if len(gs.ZoneCastGrants) != 0 {
		t.Errorf("until_end_of_turn grant should expire at cleanup of grant turn")
	}
}

func TestExpireZoneCastGrants_UntilEndOfNextTurn(t *testing.T) {
	gs := &GameState{
		Turn:  3,
		Seats: []*Seat{{Idx: 0, Life: 40}},
		Flags: map[string]int{},
		ZoneCastGrants: map[*Card]*ZoneCastPermission{},
	}
	card := &Card{Name: "Exiled Spell", Owner: 0}
	gs.ZoneCastGrants[card] = &ZoneCastPermission{
		Zone:       ZoneExile,
		ManaCost:   -1,
		Duration:   "until_end_of_next_turn",
		GrantTurn:  3,
	}

	ExpireZoneCastGrants(gs)
	if len(gs.ZoneCastGrants) != 1 {
		t.Errorf("until_end_of_next_turn should survive cleanup of grant turn")
	}

	gs.Turn = 4
	ExpireZoneCastGrants(gs)
	if len(gs.ZoneCastGrants) != 0 {
		t.Errorf("until_end_of_next_turn should expire at cleanup of next turn")
	}
}

func TestExpireZoneCastGrants_WhileSourceOnBf(t *testing.T) {
	perm := &Permanent{
		Card:      &Card{Name: "Source Perm"},
		Timestamp: 42,
	}
	gs := &GameState{
		Turn:  5,
		Seats: []*Seat{{Idx: 0, Life: 40, Battlefield: []*Permanent{perm}}},
		Flags: map[string]int{},
		ZoneCastGrants: map[*Card]*ZoneCastPermission{},
	}
	card := &Card{Name: "Exiled Card", Owner: 0}
	gs.ZoneCastGrants[card] = &ZoneCastPermission{
		Zone:            ZoneExile,
		ManaCost:        -1,
		Duration:        "while_source_on_bf",
		SourceTimestamp:  42,
	}

	ExpireZoneCastGrants(gs)
	if len(gs.ZoneCastGrants) != 1 {
		t.Errorf("should survive while source is on battlefield")
	}

	gs.Seats[0].Battlefield = nil
	ExpireZoneCastGrants(gs)
	if len(gs.ZoneCastGrants) != 0 {
		t.Errorf("should expire when source leaves battlefield")
	}
}

func TestExpireSourceGrants(t *testing.T) {
	gs := &GameState{
		Turn:  5,
		Seats: []*Seat{{Idx: 0, Life: 40}},
		Flags: map[string]int{},
		ZoneCastGrants: map[*Card]*ZoneCastPermission{},
	}
	card1 := &Card{Name: "Card A", Owner: 0}
	card2 := &Card{Name: "Card B", Owner: 0}
	gs.ZoneCastGrants[card1] = &ZoneCastPermission{
		Zone:           ZoneExile,
		SourceTimestamp: 42,
	}
	gs.ZoneCastGrants[card2] = &ZoneCastPermission{
		Zone:           ZoneExile,
		SourceTimestamp: 99,
	}

	ExpireSourceGrants(gs, 42)

	if len(gs.ZoneCastGrants) != 1 {
		t.Errorf("should only remove grants matching source timestamp 42")
	}
	if gs.ZoneCastGrants[card2] == nil {
		t.Errorf("card2 (timestamp 99) should survive")
	}
}
