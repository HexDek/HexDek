package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerOtharriSunsGlory wires Otharri, Suns' Glory.
//
// Oracle text:
//
//	Flying, lifelink, haste
//	Whenever Otharri attacks, you get an experience counter. Then
//	create a 2/2 red Rebel creature token that's tapped and attacking
//	for each experience counter you have.
//	{2}{R}{W}, Tap an untapped Rebel you control: Return this card
//	from your graveyard to the battlefield tapped.
//
// Experience counters are tracked per-seat. The graveyard recursion is
// left as a parser gap.
func registerOtharriSunsGlory(r *Registry) {
	r.OnTrigger("Otharri, Suns' Glory", "attacks", otharriAttacks)
}

func otharriAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "otharri_attack_rebel_tokens"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	attackerSeat, _ := ctx["seat"].(int)
	if attackerSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if seat.Flags == nil {
		seat.Flags = map[string]int{}
	}
	seat.Flags["experience"]++
	x := seat.Flags["experience"]
	for i := 0; i < x; i++ {
		tok := gameengine.CreateCreatureToken(gs, perm.Controller, "Rebel",
			[]string{"creature", "rebel"}, 2, 2)
		if tok != nil {
			tok.Tapped = true
			if tok.Card != nil {
				tok.Card.Colors = []string{"R"}
			}
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"experience": x,
		"tokens":     x,
	})
}
