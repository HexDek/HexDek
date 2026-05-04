package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerOsgirTheReconstructor wires Osgir, the Reconstructor.
//
// Oracle text:
//
//	Vigilance
//	{1}, Sacrifice an artifact: Target creature you control gets +2/+0
//	until end of turn.
//	{X}, {T}, Exile an artifact card with mana value X from your
//	graveyard: Create two tokens that are copies of the exiled card.
//	Activate only as a sorcery.
//
// Activated abilities require player-driven scheduling that the per_card
// dispatch path doesn't currently invoke automatically. Recorded as a
// parser gap so Heimdall reports it correctly.
func registerOsgirTheReconstructor(r *Registry) {
	r.OnETB("Osgir, the Reconstructor", osgirETB)
}

func osgirETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "osgir_etb"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"sac_artifact_pump_and_x_exile_copy_activated_unimplemented")
}
