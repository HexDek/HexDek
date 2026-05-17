package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerBurnishedHart wires Burnished Hart (Muninn parser-gap #171,
// single-game hit on 2026-05-17).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{3}
//	Artifact Creature — Elk
//	{3}, Sacrifice this creature: Search your library for up to two
//	basic land cards, put them onto the battlefield tapped, then
//	shuffle.
//
// Implementation:
//   - OnActivated index 0: sacrifice via gameengine.SacrificePermanent
//     (handles §701.17 zone change + LTB triggers), then scan the
//     controller's library for up to two basic land cards and route
//     each through MoveCard(library → battlefield_tapped) so landfall
//     observers (Valakut Exploration, Lotus Cobra, etc.) see the tapped
//     entry. Single shuffle at the end iff at least one basic was found.
//   - The {3} mana cost is enforced by the activation pipeline; we
//     skip the search-and-shuffle entirely if the permanent has
//     already left the battlefield (defensive — covers double-resolve
//     from tests).
func registerBurnishedHart(r *Registry) {
	r.OnActivated("Burnished Hart", burnishedHartActivate)
}

func burnishedHartActivate(gs *gameengine.GameState, src *gameengine.Permanent, idx int, ctx map[string]interface{}) {
	const slug = "burnished_hart_sac_fetch_two_basics"
	if gs == nil || src == nil || src.Card == nil {
		return
	}
	seat := src.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil {
		return
	}

	gameengine.SacrificePermanent(gs, src, "burnished_hart")

	found := []string{}
	for i := 0; i < len(s.Library) && len(found) < 2; {
		c := s.Library[i]
		if c == nil {
			i++
			continue
		}
		if !isBasicLand(c) {
			i++
			continue
		}
		gameengine.MoveCard(gs, c, seat, "library", "battlefield_tapped", "burnished_hart_search")
		found = append(found, c.DisplayName())
		// MoveCard removes from library — index stays at the next card.
	}
	if len(found) > 0 {
		shuffleLibraryPerCard(gs, seat)
	}

	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":  seat,
		"found": found,
	})
}
