package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerBaylenTheHaymaker wires Baylen, the Haymaker.
//
// Oracle text:
//
//	Tap two untapped tokens you control: Add one mana of any color.
//	Tap three untapped tokens you control: Draw a card.
//	Tap four untapped tokens you control: Put three +1/+1 counters on
//	Baylen. It gains trample until end of turn.
//
// Activated abilities; engine drives via OnActivated. Tap-multiple-tokens
// cost dispatch is non-trivial — emitPartial.
func registerBaylenTheHaymaker(r *Registry) {
	r.OnActivated("Baylen, the Haymaker", baylenActivated)
}

func baylenActivated(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "baylen_token_tap_ability"
	if gs == nil || src == nil {
		return
	}
	emitPartial(gs, slug, src.Card.DisplayName(),
		"tap_multiple_tokens_cost_dispatch_unimplemented")
}
