package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerEdeaPossessedSorceress wires Edea, Possessed Sorceress.
//
// Oracle text:
//
//	Ward {2}
//	At the beginning of combat on your turn, gain control of target
//	creature an opponent controls until end of turn. Untap that
//	creature. It gains haste until end of turn.
//	Whenever a creature you control but don't own dies, return it to
//	the battlefield under its owner's control and you draw a card.
//
// Both halves require control-change machinery beyond the engine's
// per-card hook layer; emitPartial.
func registerEdeaPossessedSorceress(r *Registry) {
	r.OnTrigger("Edea, Possessed Sorceress", "combat_begin", edeaCombat)
	r.OnTrigger("Edea, Possessed Sorceress", "creature_dies", edeaCreatureDies)
}

func edeaCombat(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "edea_threaten_combat"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"temporary_control_change_unimplemented")
}

func edeaCreatureDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "edea_owned_creature_dies_recur"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	deadController, _ := ctx["controller_seat"].(int)
	if deadController != perm.Controller {
		return
	}
	dead, _ := ctx["perm"].(*gameengine.Permanent)
	if dead == nil || dead.Card == nil {
		return
	}
	if dead.Owner == perm.Controller {
		return
	}
	// Return to owner's battlefield + draw.
	card := dead.Card
	gameengine.MoveCard(gs, card, perm.Controller, "graveyard", "battlefield", "edea_recur")
	enterBattlefieldWithETB(gs, dead.Owner, card, false)
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"owner": dead.Owner,
	})
}
