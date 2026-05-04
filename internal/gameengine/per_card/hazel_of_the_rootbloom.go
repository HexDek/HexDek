package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerHazelOfTheRootbloom wires Hazel of the Rootbloom.
//
// Oracle text:
//
//	{T}, Pay 2 life, Tap X untapped tokens you control: Add X mana
//	in any combination of colors.
//	At the beginning of your end step, create a token that's a copy
//	of target token you control. If that token is a Squirrel, instead
//	create two tokens that are copies of it.
//
// Token-copy creation and the X-tokens-tap activated mana ability are
// non-trivial — emitPartial.
func registerHazelOfTheRootbloom(r *Registry) {
	r.OnETB("Hazel of the Rootbloom", hazelRootbloomETB)
}

func hazelRootbloomETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "hazel_rootbloom_partial"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"end_step_token_copy_and_x_token_mana_ability_unimplemented")
}
