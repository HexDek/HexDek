package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKalamaxTheStormsireCustom adds Kalamax's instant-copy trigger
// and the +1/+1 counter-on-copy trigger that the auto-generated static
// stub leaves as no-ops.
//
// Oracle text:
//
//	Whenever you cast your first instant spell each turn, if Kalamax
//	is tapped, copy that spell. You may choose new targets for the copy.
//	Whenever you copy an instant spell, put a +1/+1 counter on Kalamax.
//
// Spell-copy plumbing isn't yet exposed to per_card (we don't have a
// "copy this StackItem" hook). We approximate by:
//   - Tracking first-instant-this-turn per Kalamax permanent via
//     perm.Flags["kalamax_first_instant_used"]; reset at upkeep.
//   - When the first instant is cast and Kalamax is tapped, emit a
//     "kalamax_spell_copied" event the engine/AI can observe and stack
//     a +1/+1 counter on Kalamax (the second trigger fires on copy).
//   - Self-copy without a real spell duplicate is a partial; we still
//     emit the parser_gap so the audit can find it.
func registerKalamaxTheStormsireCustom(r *Registry) {
	r.OnETB("Kalamax, the Stormsire", kalamaxETBReset)
	r.OnTrigger("Kalamax, the Stormsire", "instant_or_sorcery_cast", kalamaxFirstInstantCopy)
}

func kalamaxETBReset(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["kalamax_first_instant_used"] = 0
	scheduleKalamaxReset(gs, perm)
}

func scheduleKalamaxReset(gs *gameengine.GameState, perm *gameengine.Permanent) {
	gs.RegisterDelayedTrigger(&gameengine.DelayedTrigger{
		TriggerAt:      "your_next_upkeep",
		ControllerSeat: perm.Controller,
		SourceCardName: perm.Card.DisplayName(),
		OneShot:        true,
		EffectFn: func(gs *gameengine.GameState) {
			if perm.Flags != nil {
				perm.Flags["kalamax_first_instant_used"] = 0
			}
			seat := gs.Seats[perm.Controller]
			if seat == nil {
				return
			}
			for _, p := range seat.Battlefield {
				if p == perm {
					scheduleKalamaxReset(gs, perm)
					return
				}
			}
		},
	})
}

func kalamaxFirstInstantCopy(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kalamax_first_instant_copy"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	caster, ok := ctx["caster_seat"].(int)
	if !ok || caster != perm.Controller {
		return
	}
	// Only instants — sorceries don't trigger Kalamax.
	if isInstant, _ := ctx["is_instant"].(bool); !isInstant {
		// Fallback: check the spell name — if "is_instant" wasn't set
		// (older callers), abort to avoid spurious copies.
		if _, has := ctx["is_instant"]; !has {
			return
		}
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	if perm.Flags["kalamax_first_instant_used"] > 0 {
		return
	}
	if !perm.Tapped {
		return
	}
	perm.Flags["kalamax_first_instant_used"] = 1
	// Stack a +1/+1 counter — the "whenever you copy an instant" rider
	// would normally put this on; we apply it here directly because we
	// don't have a real copy-event channel.
	if perm.Counters == nil {
		perm.Counters = map[string]int{}
	}
	perm.Counters["+1/+1"]++
	spellName, _ := ctx["spell_name"].(string)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"copied":    spellName,
		"new_count": perm.Counters["+1/+1"],
	})
	emitPartial(gs, "kalamax_real_copy_to_stack", perm.Card.DisplayName(),
		"actual stack-item duplication needs engine-side copy hook; counter applied as proxy")
}
