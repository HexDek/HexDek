package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGargosViciousWatcher wires Gargos, Vicious Watcher.
//
// Oracle text:
//
//	Vigilance
//	Hydra spells you cast cost {4} less to cast.
//	Whenever a creature you control becomes the target of a spell,
//	Gargos fights up to one target creature you don't control.
//
// Cost reduction is AST. The fight trigger requires the targeting
// observer pipeline; emitPartial.
func registerGargosViciousWatcher(r *Registry) {
	r.OnTrigger("Gargos, Vicious Watcher", "creature_targeted", gargosFight)
}

func gargosFight(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "gargos_targeted_fight"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"target_observer_and_fight_resolution_unimplemented")
}
