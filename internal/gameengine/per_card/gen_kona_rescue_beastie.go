package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKonaRescueBeastie wires Kona, Rescue Beastie.
//
// Oracle text (DSK, {3}{G}, 4/3):
//
//	Survival — At the beginning of your second main phase, if Kona is
//	tapped, you may put a permanent card from your hand onto the
//	battlefield.
//
// Implementation:
//   - "postcombat_main_controller" trigger gated on Kona being tapped
//     and the active player owning Kona. Picks the highest-CMC
//     permanent in hand (best cheat target) and drops it onto the
//     battlefield via enterBattlefieldWithETB.
//   - The "you may" choice always opts in: a free permanent on the
//     battlefield is virtually never wrong.
func registerKonaRescueBeastie(r *Registry) {
	r.OnTrigger("Kona, Rescue Beastie", "postcombat_main_controller", konaSurvival)
}

func konaSurvival(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kona_survival_cheat_permanent"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	if !perm.Tapped {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	var pick *gameengine.Card
	bestCMC := -1
	for _, c := range seat.Hand {
		if c == nil {
			continue
		}
		if !cardIsPermanent(c) {
			continue
		}
		cmc := cardCMC(c)
		if cmc > bestCMC {
			bestCMC = cmc
			pick = c
		}
	}
	if pick == nil {
		return
	}
	gameengine.MoveCard(gs, pick, perm.Controller, "hand", "battlefield", "kona_survival")
	enterBattlefieldWithETB(gs, perm.Controller, pick, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"cheated": pick.DisplayName(),
		"cmc":    bestCMC,
	})
}

func cardIsPermanent(c *gameengine.Card) bool {
	if c == nil {
		return false
	}
	for _, t := range c.Types {
		switch t {
		case "creature", "artifact", "enchantment", "land", "planeswalker", "battle":
			return true
		}
	}
	return false
}
