package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTiamat wires Tiamat.
//
// Oracle text:
//
//   Flying
//   When Tiamat enters, if you cast it, search your library for up to
//   five Dragon cards not named Tiamat that each have different names,
//   reveal them, put them into your hand, then shuffle.
//
// Implementation:
//   - Flying is an AST keyword.
//   - ETB: walk controller's library and pick up to five distinct-named
//     Dragon cards (highest CMC first as a heuristic — Tiamat assembles
//     the finisher package, so prefer impactful dragons). Move each into
//     hand via MoveCard so zone-change triggers fire, then shuffle once.
//   - The "if you cast it" gate is approximated: we always run the tutor
//     on ETB. Reanimated/blinked Tiamat skipping this leg requires the
//     CastFromSomewhere bookkeeping that lives on the spell stack — we
//     emit a partial so analytics can flag false positives.
func registerTiamat(r *Registry) {
	r.OnETB("Tiamat", tiamatETB)
}

func tiamatETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "tiamat_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil {
		return
	}

	type pick struct {
		idx int
		cmc int
	}
	used := map[string]bool{}
	var picks []pick
	for i, c := range s.Library {
		if c == nil {
			continue
		}
		if !cardHasType(c, "dragon") {
			continue
		}
		name := strings.ToLower(c.DisplayName())
		if name == "tiamat" {
			continue
		}
		if used[name] {
			continue
		}
		used[name] = true
		picks = append(picks, pick{idx: i, cmc: gameengine.ManaCostOf(c)})
	}

	for i := 0; i < len(picks); i++ {
		for j := i + 1; j < len(picks); j++ {
			if picks[j].cmc > picks[i].cmc {
				picks[i], picks[j] = picks[j], picks[i]
			}
		}
	}
	if len(picks) > 5 {
		picks = picks[:5]
	}

	pulled := make([]string, 0, len(picks))
	cards := make([]*gameengine.Card, 0, len(picks))
	for _, p := range picks {
		if p.idx < 0 || p.idx >= len(s.Library) {
			continue
		}
		cards = append(cards, s.Library[p.idx])
	}
	for _, c := range cards {
		gameengine.MoveCard(gs, c, seat, "library", "hand", "tiamat_etb")
		pulled = append(pulled, c.DisplayName())
	}
	shuffleLibraryPerCard(gs, seat)

	gs.LogEvent(gameengine.Event{
		Kind:   "search_library",
		Seat:   seat,
		Source: "Tiamat",
		Details: map[string]interface{}{
			"found":       pulled,
			"destination": "hand",
			"reason":      "tiamat_etb",
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          seat,
		"dragons_found": len(pulled),
		"dragons":       pulled,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"if_you_cast_it_gate_not_enforced_for_blinked_or_reanimated_tiamat")
}
