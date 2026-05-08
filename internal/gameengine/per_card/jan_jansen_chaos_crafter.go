package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerJanJansenChaosCrafter wires Jan Jansen, Chaos Crafter.
//
// Oracle text:
//
//	Haste
//	{T}, Sacrifice an artifact creature: Create two Treasure tokens.
//	{T}, Sacrifice a noncreature artifact: Create two 1/1 colorless
//	Construct artifact creature tokens.
//
// Both activated abilities are AI-policy decisions — emitPartial.
func registerJanJansenChaosCrafter(r *Registry) {
	r.OnActivated("Jan Jansen, Chaos Crafter", janJansenActivated)
}

func janJansenActivated(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "jan_jansen_activated"
	if gs == nil || src == nil {
		return
	}
	// Both abilities cost {T}.
	if src.Tapped {
		return
	}
	src.Tapped = true
	emitPartial(gs, slug, src.Card.DisplayName(),
		"sac_artifact_creature_to_treasure_or_construct_unimplemented")
}
