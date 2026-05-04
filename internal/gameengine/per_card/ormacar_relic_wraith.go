package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerOrmacarRelicWraith wires Ormacar, Relic Wraith.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Vigilance, menace, lifelink
//	Precious (You can have two commanders if the other one is a
//	  legendary noncreature artifact.)
//	As long as you control your Precious, Ormacar gets +X/+X, where X
//	  is the mana value of your Precious.
//
// The +X/+X buff is a continuous-effect characteristic-defining ability
// dependent on the controller's other commander identity. Register the
// trigger only as a partial-flag marker — full Precious wiring lives in
// the layers pipeline.
func registerOrmacarRelicWraith(r *Registry) {
	r.OnETB("Ormacar, Relic Wraith", ormacarETB)
}

func ormacarETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "ormacar_precious_buff", perm.Card.DisplayName(),
		"plus_x_plus_x_from_precious_partner_continuous_static_not_modeled")
}
