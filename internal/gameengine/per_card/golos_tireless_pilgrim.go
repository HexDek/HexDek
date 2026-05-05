package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGolosTirelessPilgrim wires Golos, Tireless Pilgrim.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	When Golos enters, you may search your library for a land card, put
//	that card onto the battlefield tapped, then shuffle.
//	{2}{W}{U}{B}{R}{G}: Exile the top three cards of your library. You
//	may play them this turn without paying their mana costs.
//
// Implementation:
//   - ETB: search library for the most-color-fixing land available
//     (prefer non-basic duals/shocks/triomes by color count, then basics
//     of the most-needed color), put onto battlefield tapped, shuffle.
//   - Activated 5-color exile-and-cast is mana-prohibitive in simulation
//     and the cast-from-exile machinery would need a free-cast hook —
//     emitPartial flags this; the activated handler is a no-op stub so
//     the registration exists for AI eligibility checks.
func registerGolosTirelessPilgrim(r *Registry) {
	r.OnETB("Golos, Tireless Pilgrim", golosTirelessPilgrimETB)
	r.OnActivated("Golos, Tireless Pilgrim", golosTirelessPilgrimActivate)
}

func golosTirelessPilgrimETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "golos_tireless_pilgrim_land_tutor"
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

	// Pick best land: highest color-source count wins; basic last.
	bestIdx := -1
	bestScore := -1
	for i, c := range s.Library {
		if c == nil || !cardHasType(c, "land") {
			continue
		}
		score := len(c.Colors) * 10
		if !cardHasType(c, "basic") {
			score += 5
		}
		if score == 0 {
			score = 1
		}
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		shuffleLibraryPerCard(gs, seat)
		emitFail(gs, slug, perm.Card.DisplayName(), "no_land_in_library", map[string]interface{}{"seat": seat})
		return
	}
	land := s.Library[bestIdx]
	s.Library = append(s.Library[:bestIdx], s.Library[bestIdx+1:]...)
	shuffleLibraryPerCard(gs, seat)
	enterBattlefieldWithETB(gs, seat, land, true)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     seat,
		"tutored":  land.DisplayName(),
		"entered":  "tapped",
	})
}

func golosTirelessPilgrimActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	if gs == nil || src == nil {
		return
	}
	emitPartial(gs, "golos_tireless_pilgrim_activated", src.Card.DisplayName(),
		"five_color_exile_top_three_play_without_paying_not_implemented")
}
