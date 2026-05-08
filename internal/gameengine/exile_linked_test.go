package gameengine

import "testing"

func TestExileLinked_LinksAndReturns(t *testing.T) {
	gs := &GameState{
		Seats: []*Seat{
			{Idx: 0, Life: 40, Hand: []*Card{
				{Name: "Target Creature", Types: []string{"creature"}, Owner: 0},
			}},
		},
		Flags: map[string]int{},
	}

	oRing := &Permanent{
		Card:       &Card{Name: "Oblivion Ring", Types: []string{"enchantment"}, Owner: 0},
		Controller: 0,
		Timestamp:  42,
	}
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, oRing)

	target := gs.Seats[0].Hand[0]
	ExileLinked(gs, oRing, target, 0, "hand")

	if len(gs.Seats[0].Hand) != 0 {
		t.Errorf("expected hand empty after exile, got %d", len(gs.Seats[0].Hand))
	}
	if target.ExiledByTimestamp != 42 {
		t.Errorf("expected ExiledByTimestamp=42, got %d", target.ExiledByTimestamp)
	}
	if len(oRing.LinkedExile) != 1 || oRing.LinkedExile[0] != target {
		t.Errorf("expected oRing.LinkedExile to contain target")
	}

	found := false
	for _, c := range gs.Seats[0].Exile {
		if c == target {
			found = true
		}
	}
	if !found {
		t.Errorf("target should be in exile zone")
	}

	ReturnLinkedExile(gs, oRing, "hand")

	if target.ExiledByTimestamp != 0 {
		t.Errorf("expected ExiledByTimestamp reset to 0, got %d", target.ExiledByTimestamp)
	}
	if len(oRing.LinkedExile) != 0 {
		t.Errorf("expected LinkedExile cleared")
	}
	foundInHand := false
	for _, c := range gs.Seats[0].Hand {
		if c == target {
			foundInHand = true
		}
	}
	if !foundInHand {
		t.Errorf("target should be returned to hand")
	}
}
