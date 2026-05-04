package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGornogTheRedReaper wires Gornog, the Red Reaper.
//
// Oracle text:
//
//	Haste
//	Cowards can't block Warriors.
//	Whenever one or more Warriors you control attack a player, target
//	creature that player controls becomes a Coward.
//	Attacking Warriors you control get +X/+0, where X is the number
//	of Cowards your opponents control.
//
// Coward type-change and tribal block restriction are static effects
// the per-card surface can't model — emitPartial.
func registerGornogTheRedReaper(r *Registry) {
	r.OnETB("Gornog, the Red Reaper", gornogTheRedReaperETB)
}

func gornogTheRedReaperETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "gornog_red_reaper_partial"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"coward_type_change_warrior_buff_and_block_restriction_unimplemented")
}
