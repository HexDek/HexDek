package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerEzrimAgencyChief wires Ezrim, Agency Chief.
//
// Oracle text (MKM, {1}{W}{W}{U}{U}, 5/5):
//
//	Flying
//	When Ezrim enters, investigate twice.
//	{1}, Sacrifice an artifact: Ezrim gains your choice of vigilance,
//	lifelink, or hexproof until end of turn.
//
// Implementation:
//   - ETB creates two Clue tokens (the standard investigate effect).
//   - The activated grant ({1}, sacrifice an artifact) is left to the
//     AST engine — it's a cost-bearing ability with no auto-activation
//     hook on this card. Cost-unenforced flag retained for the
//     tracking section.
func registerEzrimAgencyChief(r *Registry) {
	r.OnETB("Ezrim, Agency Chief", ezrimETB)
}

func ezrimETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "ezrim_etb_investigate_twice"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	gameengine.CreateClueToken(gs, seat)
	gameengine.CreateClueToken(gs, seat)
	gs.LogEvent(gameengine.Event{
		Kind:   "investigate",
		Seat:   seat,
		Source: perm.Card.DisplayName(),
		Amount: 2,
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  seat,
		"clues": 2,
	})
}
