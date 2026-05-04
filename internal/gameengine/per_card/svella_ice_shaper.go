package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSvellaIceShaper wires Svella, Ice Shaper.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	{3}, {T}: Create a colorless snow artifact token named Icy
//	Manalith with "{T}: Add one mana of any color."
//	{6}{R}{G}, {T}: Look at the top four cards of your library. You
//	may cast a spell from among them without paying its mana cost.
//	Put the rest on the bottom of your library in a random order.
//
// Implementation:
//   - Activated 0 ({3},{T}): mint Icy Manalith snow artifact token.
//   - Activated 1 ({6}{R}{G},{T}): peek top 4. We pick the highest-CMC
//     non-land we can afford to cast for free; if no spell is castable,
//     bottom all 4. Bottoming order is randomized via the engine's
//     shuffle helper. Free-cast handoff is emitPartial.
func registerSvellaIceShaper(r *Registry) {
	r.OnActivated("Svella, Ice Shaper", svellaActivate)
}

func svellaActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	if gs == nil || src == nil {
		return
	}
	switch abilityIdx {
	case 0:
		svellaMintManalith(gs, src)
	case 1:
		svellaTopFour(gs, src)
	}
}

func svellaMintManalith(gs *gameengine.GameState, src *gameengine.Permanent) {
	const slug = "svella_mint_icy_manalith"
	tok := &gameengine.Card{
		Name:  "Icy Manalith",
		Owner: src.Controller,
		Types: []string{"token", "artifact", "snow"},
	}
	enterBattlefieldWithETB(gs, src.Controller, tok, false)
	src.Tapped = true
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat": src.Controller,
	})
}

func svellaTopFour(gs *gameengine.GameState, src *gameengine.Permanent) {
	const slug = "svella_top_four_free_cast"
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	src.Tapped = true
	if len(seat.Library) == 0 {
		emitFail(gs, slug, src.Card.DisplayName(), "library_empty", nil)
		return
	}
	n := 4
	if n > len(seat.Library) {
		n = len(seat.Library)
	}
	top := seat.Library[:n]
	// Choose best non-land highest-CMC for free cast.
	bestIdx := -1
	bestCMC := -1
	for i, c := range top {
		if c == nil || cardHasType(c, "land") {
			continue
		}
		if cardCMC(c) > bestCMC {
			bestCMC = cardCMC(c)
			bestIdx = i
		}
	}
	chosen := ""
	if bestIdx >= 0 {
		chosen = top[bestIdx].DisplayName()
	}
	// Bottom the rest. Move them out of the top slice in order.
	rest := []*gameengine.Card{}
	for i, c := range top {
		if i == bestIdx {
			continue
		}
		rest = append(rest, c)
	}
	// Remove top n.
	seat.Library = seat.Library[n:]
	// Append rest to bottom.
	seat.Library = append(seat.Library, rest...)
	// If we picked a card, send it to exile to flag the free-cast handoff.
	if bestIdx >= 0 {
		gameengine.MoveCard(gs, top[bestIdx], src.Controller, "library", "exile", "svella_free_cast")
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":     src.Controller,
		"top_size": n,
		"chosen":   chosen,
	})
	if bestIdx >= 0 {
		emitPartial(gs, slug, src.Card.DisplayName(),
			"free_cast_of_chosen_spell_from_exile_not_dispatched")
	}
}
