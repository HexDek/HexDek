package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerYasharnImplacableEarthCustom implements Yasharn's ETB tutor and
// records the life-pay/sacrifice prohibition as a state flag the cost
// system can branch on.
//
// Oracle text:
//
//	When Yasharn enters, search your library for a basic Forest card
//	and a basic Plains card, reveal those cards, put them into your
//	hand, then shuffle.
//	Players can't pay life or sacrifice nonland permanents to cast
//	spells or activate abilities.
//
// The static prohibition is engine-side; we toggle a global flag so
// cost code can detect Yasharn's presence without a battlefield walk.
func registerYasharnImplacableEarthCustom(r *Registry) {
	r.OnETB("Yasharn, Implacable Earth", yasharnETB)
}

func yasharnETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "yasharn_etb_basic_tutor"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	var forest, plains *gameengine.Card
	for _, c := range seat.Library {
		if c == nil {
			continue
		}
		if forest == nil && cardHasType(c, "basic") && cardHasSubtype(c, "forest") {
			forest = c
			continue
		}
		if plains == nil && cardHasType(c, "basic") && cardHasSubtype(c, "plains") {
			plains = c
		}
	}
	got := []string{}
	if forest != nil {
		gameengine.MoveCard(gs, forest, perm.Controller, "library", "hand", "yasharn_search")
		got = append(got, forest.DisplayName())
	}
	if plains != nil {
		gameengine.MoveCard(gs, plains, perm.Controller, "library", "hand", "yasharn_search")
		got = append(got, plains.DisplayName())
	}
	shuffleLibraryPerCard(gs, perm.Controller)
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	gs.Flags["yasharn_active_seat"] = perm.Controller + 1
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"found":  got,
	})
	emitPartial(gs, "yasharn_pay_life_prohibition", perm.Card.DisplayName(),
		"life-pay/sacrifice cost block needs cost-system branch on yasharn_active_seat flag")
}
