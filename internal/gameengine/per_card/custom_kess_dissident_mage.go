package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKessDissidentMageCustom replaces the auto-generated stub with
// a real implementation of Kess's once-per-turn cast-from-graveyard
// privilege. Mirrors the Karador handler shape, gated on instants and
// sorceries instead of creatures.
//
// Oracle text:
//
//	Flying
//	Once during each of your turns, you may cast an instant or
//	sorcery spell from your graveyard. If a spell cast this way
//	would be put into your graveyard, exile it instead.
//
// Implementation:
//   - Tracks usage via perm.Flags["kess_used_this_turn"].
//   - Resets on the controller's untap step via a recurring delayed
//     trigger registered at ETB.
//   - When invoked via the activated hook, picks the best (highest
//     CMC) instant or sorcery from Kess's controller's graveyard,
//     stamps its StackItem with exile_on_resolve so the engine's
//     existing replacement sends the spell to exile after resolution,
//     then resolves it (approximated as direct effect application —
//     the per_card pipeline doesn't have a "cast from non-hand zone"
//     entry point, so we move the card directly into the stack
//     pipeline via a synthetic StackItem).
func registerKessDissidentMageCustom(r *Registry) {
	r.OnETB("Kess, Dissident Mage", kessETB)
	r.OnActivated("Kess, Dissident Mage", kessCastFromGY)
}

func kessETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["kess_used_this_turn"] = 0
	scheduleKessReset(gs, perm)
}

func scheduleKessReset(gs *gameengine.GameState, perm *gameengine.Permanent) {
	gs.RegisterDelayedTrigger(&gameengine.DelayedTrigger{
		TriggerAt:      "your_next_upkeep",
		ControllerSeat: perm.Controller,
		SourceCardName: perm.Card.DisplayName(),
		OneShot:        true,
		EffectFn: func(gs *gameengine.GameState) {
			if perm.Flags != nil {
				perm.Flags["kess_used_this_turn"] = 0
			}
			seat := gs.Seats[perm.Controller]
			if seat == nil {
				return
			}
			for _, p := range seat.Battlefield {
				if p == perm {
					scheduleKessReset(gs, perm)
					return
				}
			}
		},
	})
}

func kessCastFromGY(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "kess_cast_from_graveyard"
	if gs == nil || src == nil {
		return
	}
	if src.Flags == nil {
		src.Flags = map[string]int{}
	}
	if src.Flags["kess_used_this_turn"] > 0 {
		emitFail(gs, slug, src.Card.DisplayName(), "already_used_this_turn", nil)
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	// Caller can pin a specific card via ctx["target_card"]; otherwise
	// pick the highest-CMC instant/sorcery in graveyard.
	var best *gameengine.Card
	if ctx != nil {
		if c, ok := ctx["target_card"].(*gameengine.Card); ok && c != nil {
			best = c
		}
	}
	if best == nil {
		bestCMC := -1
		for _, c := range seat.Graveyard {
			if c == nil {
				continue
			}
			if !cardHasType(c, "instant") && !cardHasType(c, "sorcery") {
				continue
			}
			if cmc := cardCMC(c); cmc > bestCMC {
				best = c
				bestCMC = cmc
			}
		}
	}
	if best == nil {
		emitFail(gs, slug, src.Card.DisplayName(), "no_instant_or_sorcery_in_graveyard", nil)
		return
	}

	// Remove from graveyard and push as a synthetic StackItem stamped
	// for exile-on-resolve so the engine routes it to exile post-
	// resolution per Kess's last clause.
	for i, c := range seat.Graveyard {
		if c == best {
			seat.Graveyard = append(seat.Graveyard[:i], seat.Graveyard[i+1:]...)
			break
		}
	}
	item := &gameengine.StackItem{
		Controller: src.Controller,
		Card:       best,
		Kind:       "spell",
		CostMeta:   map[string]interface{}{"exile_on_resolve": true, "kess_grave_cast": true},
	}
	gameengine.PushStackItem(gs, item)
	src.Flags["kess_used_this_turn"] = 1
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat": src.Controller,
		"cast": best.DisplayName(),
	})
}
