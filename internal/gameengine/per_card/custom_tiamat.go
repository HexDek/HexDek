package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTiamatCustom adds Tiamat's "ETB → tutor up to 5 different-named
// non-Tiamat Dragons to hand" trigger that the auto-generated stub omits.
//
// Oracle text (Scryfall, verified 2026-05-16):
//
//	Flying
//	When Tiamat enters, if you cast it, search your library for up to
//	five Dragon cards not named Tiamat that each have different names,
//	reveal them, put them into your hand, then shuffle.
//
// Flying is an AST keyword. The "if you cast it" intervening-if (CR
// §603.6c) gates on perm.Flags["was_cast"], which stack.go stamps when a
// permanent resolves through the cast path; blink / reanimate / token
// copy leave it unset, so the tutor silently no-ops on those, matching
// the rules text. Library search is deterministic: walk in order, skip
// already-collected names, stop at five. Picked cards are recorded on
// the emitted event (the "reveal them" clause is observable telemetry
// only — opponents don't see hands in HexDek today).
func registerTiamatCustom(r *Registry) {
	r.OnETB("Tiamat", tiamatETBSearch)
}

func tiamatETBSearch(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "tiamat_etb_dragon_tutor"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil || perm.Flags["was_cast"] != 1 {
		emit(gs, slug, "Tiamat", map[string]interface{}{
			"seat":   perm.Controller,
			"tutor":  "skipped",
			"reason": "not_cast",
		})
		return
	}
	if perm.Controller < 0 || perm.Controller >= len(gs.Seats) {
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
	shuffleLibraryPerCard(gs, perm.Controller)
	names := make([]string, 0, len(picked))
	for _, c := range picked {
		names = append(names, c.DisplayName())
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"found":    len(picked),
		"names":    names,
		"revealed": names,
	})
}
