package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerIronSpiderStarkUpgrade wires Iron Spider, Stark Upgrade.
//
// Oracle text:
//
//	Vigilance
//	{T}: Put a +1/+1 counter on each artifact creature and/or
//	Vehicle you control.
//	{2}, Remove two +1/+1 counters from among artifacts you control:
//	Draw a card.
//
// Both activated abilities are AI-policy decisions — emitPartial.
func registerIronSpiderStarkUpgrade(r *Registry) {
	r.OnActivated("Iron Spider, Stark Upgrade", ironSpiderActivated)
}

func ironSpiderActivated(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "iron_spider_activated"
	if gs == nil || src == nil {
		return
	}
	emitPartial(gs, slug, src.Card.DisplayName(),
		"tap_buff_artifact_creatures_and_remove_counters_to_draw_unimplemented")
}
