package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSylvanLibrary wires up Sylvan Library.
//
// Oracle text:
//
//   At the beginning of your draw step, you may draw two additional
//   cards. If you do, choose two cards in your hand drawn this turn.
//   For each of those cards, pay 4 life or put the card on top of
//   your library.
//
// GG enchantment. One of the best card-advantage enchantments in green.
// In cEDH, pilots often pay 8 life to keep both extra draws (especially
// at 40 life in commander).
//
// Per-card resolution:
//   - OnTrigger "draw_step_controller": draw 2 extra cards.
//   - Per-card decision: each drawn card is independently scored. A
//     card is "keep-worthy" if its mana value ≥ 3 OR it's a known
//     value engine / combo piece (heuristic: lands tuck unless we're
//     mana-screwed; cheap cantrips also tuck). For each keep-worthy
//     card we pay 4 life IF current life > 4 + safety_floor (default
//     7 — leaves at least 3 life after paying). The other card goes
//     to the top of the library.
//
// In commander (40 life) Sylvan often pays 8 to keep both because the
// payoff is huge. In limited / mid-life situations the per-card path
// keeps the threats and tucks the lands.
func registerSylvanLibrary(r *Registry) {
	r.OnTrigger("Sylvan Library", "draw_step_controller", sylvanLibraryDraw)
}

// sylvanCardKeepScore returns a rough "should we keep this card" score.
// Higher = better to keep. Lands score low (tuckable for next-turn draws),
// 1-CMC cantrips score low, threats and answers score high.
func sylvanCardKeepScore(c *gameengine.Card) int {
	if c == nil {
		return -1
	}
	if cardHasType(c, "land") {
		return 0
	}
	cmc := cardCMC(c)
	if cmc <= 1 {
		return 1 // cantrips and 1-drops — easy to tuck and re-draw
	}
	if cmc >= 4 {
		return 4 // bombs / engine pieces — definitely keep
	}
	return 2 + cmc/2
}

func sylvanLibraryDraw(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "sylvan_library_draw"
	if gs == nil || perm == nil {
		return
	}
	// Only fires during controller's draw step.
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]

	// Draw two additional cards.
	var drawnCards []*gameengine.Card
	for i := 0; i < 2; i++ {
		c := drawOne(gs, seat, "Sylvan Library")
		if c != nil {
			drawnCards = append(drawnCards, c)
		}
	}
	if len(drawnCards) == 0 {
		emit(gs, slug, "Sylvan Library", map[string]interface{}{
			"seat":  seat,
			"drawn": 0,
		})
		return
	}

	// Per-card decision: rank by sylvanCardKeepScore. The single best
	// card is paid for if life is healthy. The single worst card is
	// almost always tucked. The middle case (one keep-worthy + one
	// tuckable) keeps just the bomb.
	scored := make([]*gameengine.Card, len(drawnCards))
	copy(scored, drawnCards)
	// Sort descending by keep-score.
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if sylvanCardKeepScore(scored[j]) > sylvanCardKeepScore(scored[i]) {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	const safetyFloor = 7 // never drop below this paying for Sylvan
	keep := []*gameengine.Card{}
	tuck := []*gameengine.Card{}
	remainingLife := s.Life
	lifePaid := 0
	for _, c := range scored {
		if sylvanCardKeepScore(c) >= 2 && remainingLife-4 >= safetyFloor {
			keep = append(keep, c)
			remainingLife -= 4
			lifePaid += 4
			continue
		}
		tuck = append(tuck, c)
	}

	if lifePaid > 0 {
		gameengine.LoseLife(gs, seat, lifePaid, "Sylvan Library")
	}
	// Tuck order: lowest-score card goes deepest (top of library is the
	// LAST appended via "library_top" semantics — we MoveCard each one
	// in reverse so the "best of the worst" is what we draw next turn).
	for i := len(tuck) - 1; i >= 0; i-- {
		gameengine.MoveCard(gs, tuck[i], seat, "hand", "library_top", "tuck-top")
	}
	emit(gs, slug, "Sylvan Library", map[string]interface{}{
		"seat":      seat,
		"drawn":     len(drawnCards),
		"kept":      len(keep),
		"returned":  len(tuck),
		"life_paid": lifePaid,
		"life_now":  gs.Seats[seat].Life,
	})
	_ = gs.CheckEnd()
}
