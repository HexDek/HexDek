package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSplinterRadicalRatCustom implements Splinter's "Ninja can't
// be blocked this turn" activation. The auto-generated stub is a no-op.
//
// Oracle text (Scryfall, verified 2026-05-09):
//
//	If a triggered ability of a Ninja creature you control triggers,
//	that ability triggers an additional time.
//	{1}{U}: Target Ninja can't be blocked this turn.
//
// Implementation notes:
//   - The ninja-trigger-doubling clause is a static replacement that
//     belongs in trigger dispatch (engine territory) — emitPartial in
//     the ETB hook flags this so the audit can find it.
//   - The activated ability: pick our best untapped Ninja, set its
//     Flags["unblockable"] = 1 (the engine's runtime "can't be blocked
//     this turn" flag — same one resolve_helpers.go uses). Schedule a
//     delayed trigger to clear the flag at the next end step.
//   - Best Ninja = highest power that isn't already unblockable. If we
//     don't control any Ninja, fail gracefully.
func registerSplinterRadicalRatCustom(r *Registry) {
	r.OnActivated("Splinter, Radical Rat", splinterNinjaUnblockable)
	r.OnETB("Splinter, Radical Rat", func(gs *gameengine.GameState, perm *gameengine.Permanent) {
		emitPartial(gs, "splinter_ninja_trigger_double", perm.Card.DisplayName(),
			"trigger-doubling for Ninja triggered abilities needs trigger-dispatch hook")
	})
}

func splinterNinjaUnblockable(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "splinter_ninja_unblockable"
	if gs == nil || src == nil || src.Card == nil {
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	if seat.ManaPool < 2 {
		emitFail(gs, slug, src.Card.DisplayName(), "insufficient_mana", map[string]interface{}{
			"required":  2,
			"available": seat.ManaPool,
		})
		return
	}

	// Pick our highest-power Ninja (excludes Splinter itself unless he
	// is also a Ninja by subtype — let cardHasSubtype decide).
	var pick *gameengine.Permanent
	bestPower := -1
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil || !p.IsCreature() {
			continue
		}
		if !cardHasSubtype(p.Card, "ninja") {
			continue
		}
		if p.Flags != nil && p.Flags["unblockable"] == 1 {
			continue
		}
		pow := gs.PowerOf(p)
		if pow > bestPower {
			pick = p
			bestPower = pow
		}
	}
	if pick == nil {
		emitFail(gs, slug, src.Card.DisplayName(), "no_eligible_ninja", nil)
		return
	}

	seat.ManaPool -= 2
	gameengine.SyncManaAfterSpend(seat)
	if pick.Flags == nil {
		pick.Flags = map[string]int{}
	}
	pick.Flags["unblockable"] = 1
	captured := pick
	gs.RegisterDelayedTrigger(&gameengine.DelayedTrigger{
		TriggerAt:      "next_end_step",
		ControllerSeat: src.Controller,
		SourceCardName: src.Card.DisplayName(),
		OneShot:        true,
		EffectFn: func(gs *gameengine.GameState) {
			if captured == nil || captured.Flags == nil {
				return
			}
			delete(captured.Flags, "unblockable")
		},
	})

	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":   src.Controller,
		"ninja":  pick.Card.DisplayName(),
		"power":  bestPower,
	})
}
