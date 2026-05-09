package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// dev/order-replacements — Sylvan Library now scores each drawn card
// independently. Previously the handler binarily kept both / tucked
// both based on a single life threshold; now lands and 1-CMC cantrips
// tuck even at 40 life (saving 4 life for the next bomb).

func TestSylvanLibrary_PerCard_KeepsBombTucksLand(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].Life = 40
	addPerm(gs, 0, "Sylvan Library", "enchantment")
	// Top of library: a 5-CMC bomb and a basic land.
	bomb := &gameengine.Card{Name: "Bomb", Owner: 0, Types: []string{"creature", "cmc:5"}}
	land := &gameengine.Card{Name: "Forest", Owner: 0, Types: []string{"land", "basic"}}
	gs.Seats[0].Library = []*gameengine.Card{bomb, land}

	gameengine.FireCardTrigger(gs, "draw_step_controller", map[string]interface{}{
		"active_seat": 0,
	})

	// Bomb should be in hand, land should be tucked back to library.
	foundBomb := false
	for _, c := range gs.Seats[0].Hand {
		if c == bomb {
			foundBomb = true
		}
		if c == land {
			t.Errorf("land should NOT have been kept; per-card heuristic must tuck it")
		}
	}
	if !foundBomb {
		t.Fatalf("bomb (cmc:5) must be kept; hand=%v", gs.Seats[0].Hand)
	}
	if len(gs.Seats[0].Library) != 1 || gs.Seats[0].Library[0] != land {
		t.Errorf("expected land tucked to library top; got %v", gs.Seats[0].Library)
	}
	// Paid 4 life for the one keep; not 8.
	if gs.Seats[0].Life != 36 {
		t.Errorf("expected life 36 (paid 4 for 1 keep), got %d", gs.Seats[0].Life)
	}
}

func TestSylvanLibrary_PerCard_RespectsSafetyFloor(t *testing.T) {
	gs := newGame(t, 2)
	gs.Seats[0].Life = 9 // life - 4 = 5, safetyFloor = 7 → can't keep
	addPerm(gs, 0, "Sylvan Library", "enchantment")
	bomb1 := &gameengine.Card{Name: "Bomb1", Owner: 0, Types: []string{"creature", "cmc:5"}}
	bomb2 := &gameengine.Card{Name: "Bomb2", Owner: 0, Types: []string{"creature", "cmc:5"}}
	gs.Seats[0].Library = []*gameengine.Card{bomb1, bomb2}

	gameengine.FireCardTrigger(gs, "draw_step_controller", map[string]interface{}{
		"active_seat": 0,
	})

	if gs.Seats[0].Life != 9 {
		t.Errorf("life should be unchanged at 9 (safety floor), got %d", gs.Seats[0].Life)
	}
	if len(gs.Seats[0].Hand) != 0 {
		t.Errorf("nothing should be kept under safety floor; hand=%v", gs.Seats[0].Hand)
	}
	if len(gs.Seats[0].Library) != 2 {
		t.Errorf("both bombs must be tucked; lib size=%d", len(gs.Seats[0].Library))
	}
}

func TestSylvanLibrary_PerCard_MidLifeKeepsOnlyOne(t *testing.T) {
	gs := newGame(t, 2)
	// Life 11 → after 1 keep we're at 7 (safety floor) → can't keep a
	// second.
	gs.Seats[0].Life = 11
	addPerm(gs, 0, "Sylvan Library", "enchantment")
	bomb1 := &gameengine.Card{Name: "Bomb1", Owner: 0, Types: []string{"creature", "cmc:5"}}
	bomb2 := &gameengine.Card{Name: "Bomb2", Owner: 0, Types: []string{"creature", "cmc:5"}}
	gs.Seats[0].Library = []*gameengine.Card{bomb1, bomb2}

	gameengine.FireCardTrigger(gs, "draw_step_controller", map[string]interface{}{
		"active_seat": 0,
	})

	if gs.Seats[0].Life != 7 {
		t.Errorf("expected one keep paid (life 11 → 7); got %d", gs.Seats[0].Life)
	}
	if len(gs.Seats[0].Hand) != 1 {
		t.Errorf("expected exactly 1 card kept; hand=%v", gs.Seats[0].Hand)
	}
	if len(gs.Seats[0].Library) != 1 {
		t.Errorf("expected 1 card tucked; lib size=%d", len(gs.Seats[0].Library))
	}
}
