package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerYasharnImplacableEarth wires Yasharn, Implacable Earth.
//
// Oracle text:
//
//   When Yasharn enters, search your library for a basic Forest card and a basic Plains card, reveal those cards, put them into your hand, then shuffle.
//   Players can't pay life or sacrifice nonland permanents to cast spells or activate abilities.
//
// Implementation:
//   - ETB: tutor a basic Forest and a basic Plains into hand, then
//     shuffle.
//   - Static "can't pay life or sacrifice nonland permanents to cast"
//     restriction is a global cost-payment lock and is not modeled here.
func registerYasharnImplacableEarth(r *Registry) {
	r.OnETB("Yasharn, Implacable Earth", yasharnImplacableEarthETB)
}

func yasharnImplacableEarthETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "yasharn_implacable_earth_etb"
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
	found := []string{}
	for _, want := range []string{"forest", "plains"} {
		idx := -1
		for i, c := range s.Library {
			if c == nil {
				continue
			}
			if landMatchesFetchTypes(c, []string{want}) && isBasicLand(c) {
				idx = i
				break
			}
		}
		if idx < 0 {
			continue
		}
		card := s.Library[idx]
		gameengine.MoveCard(gs, card, seat, "library", "hand", "yasharn_tutor")
		found = append(found, card.DisplayName())
	}
	shuffleLibraryPerCard(gs, seat)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     seat,
		"tutored":  found,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"static_no_life_pay_or_nonland_sac_unimplemented")
}
