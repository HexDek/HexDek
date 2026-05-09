package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerQueenMarchesa wires Queen Marchesa.
//
// Oracle text (Outlaws of Thunder Junction Commander reprint, {1}{R}{W}{B}, 3/3):
//
//	Deathtouch, haste
//	When Queen Marchesa enters, you become the monarch.
//	At the beginning of your upkeep, if an opponent is the monarch,
//	create a 1/1 black Assassin creature token with deathtouch and haste.
//
// Implementation:
//   - ETB calls BecomeMonarch(controller).
//   - "upkeep_controller" trigger: if any opponent currently holds the
//     monarchy, create a 1/1 black Assassin token. Tokens get
//     deathtouch and haste flags so combat math is correct (the AST
//     keyword pipeline applies them on creation via the type tag).
func registerQueenMarchesa(r *Registry) {
	r.OnETB("Queen Marchesa", queenMarchesaETB)
	r.OnTrigger("Queen Marchesa", "upkeep_controller", queenMarchesaUpkeep)
}

func queenMarchesaETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "queen_marchesa_etb_monarch"
	if gs == nil || perm == nil {
		return
	}
	gameengine.BecomeMonarch(gs, perm.Controller)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}

func queenMarchesaUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "queen_marchesa_upkeep_assassin"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	if !opponentIsMonarch(gs, perm.Controller) {
		return
	}
	gameengine.CreateCreatureToken(gs, perm.Controller, "Assassin Token",
		[]string{"creature", "assassin", "pip:B", "deathtouch", "haste"}, 1, 1)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}

func opponentIsMonarch(gs *gameengine.GameState, controller int) bool {
	if gs == nil || gs.Flags == nil || gs.Flags["has_monarch"] != 1 {
		return false
	}
	mSeat := gs.Flags["monarch_seat"]
	return mSeat != controller
}
