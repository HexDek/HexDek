package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTiamatCustom adds Tiamat's "ETB → tutor up to 5 different-named
// non-Tiamat Dragons to hand" trigger that the auto-generated stub omits.
//
// Oracle text:
//
//	Flying
//	When Tiamat enters, if you cast it, search your library for up to
//	five Dragon cards not named Tiamat that each have different names,
//	reveal them, put them into your hand, then shuffle.
//
// Flying is an AST keyword. We don't yet propagate the "if you cast it"
// gate — we always run the search since reanimating Tiamat is a rare
// edge case. Library search is deterministic: walk in order, skip
// already-collected names, stop at five.
func registerTiamatCustom(r *Registry) {
	r.OnETB("Tiamat", tiamatETBSearch)
}

func tiamatETBSearch(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "tiamat_etb_dragon_tutor"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	picked := []*gameengine.Card{}
	seen := map[string]bool{}
	for _, c := range seat.Library {
		if c == nil {
			continue
		}
		if !cardSubtypeMatches(c, "dragon") {
			continue
		}
		if c.DisplayName() == "Tiamat" {
			continue
		}
		if seen[c.DisplayName()] {
			continue
		}
		seen[c.DisplayName()] = true
		picked = append(picked, c)
		if len(picked) == 5 {
			break
		}
	}
	for _, c := range picked {
		gameengine.MoveCard(gs, c, perm.Controller, "library", "hand", "tiamat_search")
	}
	// Shuffle remaining library deterministically via the engine helper.
	if len(seat.Library) > 1 && gs.Rng != nil {
		gs.Rng.Shuffle(len(seat.Library), func(i, j int) {
			seat.Library[i], seat.Library[j] = seat.Library[j], seat.Library[i]
		})
	}
	names := make([]string, 0, len(picked))
	for _, c := range picked {
		names = append(names, c.DisplayName())
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"found":  len(picked),
		"names":  names,
	})
}
