package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerObekaBruteChronologistCustom implements Obeka's "end the turn"
// activation. The auto-generated stub is a no-op.
//
// Oracle text:
//
//	{T}: The player whose turn it is may end the turn. (Exile all spells
//	and abilities from the stack. The player whose turn it is discards
//	down to their maximum hand size. Damage wears off, and "this turn"
//	and "until end of turn" effects end.)
//
// Implementation notes:
//   - Tap cost is enforced by the engine's activation dispatcher; this
//     handler defensively sets src.Tapped if it isn't already.
//   - "End the turn" is a complex action (CR 723) that the engine
//     doesn't fully model: we'd need to exile the stack, jump to the
//     cleanup step, and skip every remaining trigger. We approximate by:
//       1. Marking each StackItem as Countered=true so they fizzle when
//          resolution next walks the stack — same shape as the existing
//          reResEndTurn arm in resolve_helpers.go.
//       2. Setting gs.Flags["end_turn_requested"] = 1 so the turn loop
//          can short-circuit if it cooperates (most callers don't —
//          tracked as a partial).
//       3. Logging the end_turn event for downstream observers.
//   - The actual "skip to cleanup, drop UEOT effects, clear damage"
//     phase machinery is engine territory and emitPartial-flagged so
//     the audit can find it.
func registerObekaBruteChronologistCustom(r *Registry) {
	r.OnActivated("Obeka, Brute Chronologist", obekaEndTheTurn)
}

func obekaEndTheTurn(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "obeka_end_the_turn"
	if gs == nil || src == nil || src.Card == nil {
		return
	}
	if src.Tapped {
		emitFail(gs, slug, src.Card.DisplayName(), "already_tapped", nil)
		return
	}
	src.Tapped = true

	// Exile (countersub) all stack items so the stack effectively empties
	// when resolution next walks it.
	exiledFromStack := 0
	for _, item := range gs.Stack {
		if item == nil || item.Countered {
			continue
		}
		item.Countered = true
		exiledFromStack++
	}

	// Mark the request — an honest end-the-turn loop short-circuit is a
	// partial since the engine doesn't yet model "skip to cleanup" cleanly.
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	gs.Flags["end_turn_requested"] = 1

	gs.LogEvent(gameengine.Event{
		Kind:   "end_turn",
		Seat:   src.Controller,
		Source: src.Card.DisplayName(),
		Details: map[string]interface{}{
			"exiled_from_stack": exiledFromStack,
		},
	})

	emitPartial(gs, slug, src.Card.DisplayName(),
		"phase machinery: cleanup-step jump + UEOT-effect expiration not modeled")

	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":              src.Controller,
		"exiled_from_stack": exiledFromStack,
	})
}
