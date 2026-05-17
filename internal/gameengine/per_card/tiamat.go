package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTiamat wires Tiamat (Muninn parser-gap #6, ~127K hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{2}{W}{U}{B}{R}{G}
//	Legendary Creature — Dragon God
//	Flying
//	When Tiamat enters, if you cast it, search your library for up to
//	five Dragon cards not named Tiamat that each have different names,
//	reveal them, put them into your hand, then shuffle.
//
// Implementation:
//   - Flying is AST-side.
//   - Gate on perm.Flags["was_cast"] (cast clause). Non-cast entries
//     (reanimate/blink/Sneak Attack) don't fire.
//   - Walk the controller's library, collect up to 5 Dragon cards with
//     distinct names (case-insensitive), skipping cards named "Tiamat".
//     Hat policy: take the highest-CMC dragons first (best top-end into
//     hand for ramp into).
//   - Move each pick library→hand via MoveCard so any
//     leaves-library/enters-hand replacements fire. Shuffle once after.
//   - "Reveal them" is purely informational here; we emit the names in
//     the per_card_handler log so spectator can replay.
func registerTiamat(r *Registry) {
	r.OnETB("Tiamat", tiamatETB)
}

func tiamatETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "tiamat_etb_tutor_dragons"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil || perm.Flags["was_cast"] != 1 {
		emitFail(gs, slug, perm.Card.DisplayName(), "not_cast", nil)
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}

	// Sort candidates by CMC desc into a working slice. Dedup by lowercase
	// name so "Tiamat that each have different names" holds.
	seen := map[string]bool{"tiamat": true}
	type cand struct {
		card *gameengine.Card
		cmc  int
	}
	var pool []cand
	for _, c := range seat.Library {
		if c == nil {
			continue
		}
		if !cardHasType(c, "dragon") {
			continue
		}
		name := lowercaseName(c)
		if seen[name] {
			continue
		}
		seen[name] = true
		pool = append(pool, cand{card: c, cmc: cardCMC(c)})
	}
	// Selection-sort the top 5 by CMC desc (n is tiny — usually <30).
	picks := []*gameengine.Card{}
	for len(picks) < 5 && len(pool) > 0 {
		bestIdx := 0
		for i := 1; i < len(pool); i++ {
			if pool[i].cmc > pool[bestIdx].cmc {
				bestIdx = i
			}
		}
		picks = append(picks, pool[bestIdx].card)
		pool = append(pool[:bestIdx], pool[bestIdx+1:]...)
	}

	names := make([]string, 0, len(picks))
	for _, c := range picks {
		gameengine.MoveCard(gs, c, perm.Controller, "library", "hand", slug)
		names = append(names, c.DisplayName())
	}
	shuffleLibraryPerCard(gs, perm.Controller)

	gs.LogEvent(gameengine.Event{
		Kind:   "search_library",
		Seat:   perm.Controller,
		Source: perm.Card.DisplayName(),
		Details: map[string]interface{}{
			"found":  names,
			"reason": slug,
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"to_hand":  names,
		"count":    len(names),
	})
}

func lowercaseName(c *gameengine.Card) string {
	if c == nil {
		return ""
	}
	out := []byte(c.DisplayName())
	for i, b := range out {
		if b >= 'A' && b <= 'Z' {
			out[i] = b + 32
		}
	}
	return string(out)
}
