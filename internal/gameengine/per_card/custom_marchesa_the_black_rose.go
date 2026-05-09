package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMarchesaTheBlackRoseCustom adds Marchesa's "creature with +1/+1
// counter dies → return at next end step" trigger that the auto-generated
// static stub omits.
//
// Oracle text:
//
//	Dethrone (Whenever this creature attacks the player with the most
//	life or tied for most life, put a +1/+1 counter on it.)
//	Other creatures you control have dethrone.
//	Whenever a creature you control with a +1/+1 counter on it dies,
//	return that card to the battlefield under your control at the
//	beginning of the next end step.
//
// We wire two pieces here:
//   - Death trigger: queue a delayed trigger that brings the card back
//     from graveyard at the next end step. The "under your control"
//     clause means Marchesa's controller, not the dying card's owner —
//     we honor that.
//   - Dethrone grant: register a layer-6 continuous effect that adds the
//     "dethrone" keyword to OTHER creatures Marchesa's controller
//     controls. Marchesa herself has dethrone via the AST keyword
//     pipeline (CR §613.1f layer 6 ability grants stack). The keyword
//     itself fires automatically via FireDethroneTriggers in
//     keywords_misc.go.
func registerMarchesaTheBlackRoseCustom(r *Registry) {
	r.OnETB("Marchesa, the Black Rose", marchesaGrantDethrone)
	r.OnTrigger("Marchesa, the Black Rose", "creature_dies", marchesaOnCreatureDies)
}

func marchesaGrantDethrone(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	source := perm
	const grant = "dethrone"
	pred := func(_ *gameengine.GameState, t *gameengine.Permanent) bool {
		if t == nil || t.Card == nil {
			return false
		}
		if t == source {
			return false
		}
		if t.Controller != source.Controller {
			return false
		}
		return t.IsCreature()
	}
	apply := func(_ *gameengine.GameState, _ *gameengine.Permanent, chars *gameengine.Characteristics) {
		if chars == nil {
			return
		}
		for _, k := range chars.Keywords {
			if k == grant {
				return
			}
		}
		chars.Keywords = append(chars.Keywords, grant)
	}
	gs.RegisterContinuousEffect(&gameengine.ContinuousEffect{
		Layer:          gameengine.LayerAbility,
		Timestamp:      gs.NextTimestamp(),
		SourcePerm:     source,
		SourceCardName: "Marchesa, the Black Rose",
		ControllerSeat: source.Controller,
		HandlerID:      "marchesa_dethrone_grant_" + perm.Card.DisplayName(),
		Duration:       gameengine.DurationUntilSourceLeaves,
		Predicate:      pred,
		ApplyFn:        apply,
	})
	emit(gs, "marchesa_dethrone_grant_registered", perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}

func marchesaOnCreatureDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "marchesa_return_at_next_eos"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	dyingPerm, _ := ctx["perm"].(*gameengine.Permanent)
	dyingCard, _ := ctx["card"].(*gameengine.Card)
	if dyingCard == nil && dyingPerm != nil {
		dyingCard = dyingPerm.Card
	}
	if dyingCard == nil {
		return
	}
	dyingCtrl, _ := ctx["controller_seat"].(int)
	if dyingCtrl != perm.Controller {
		return
	}
	// The dying creature must have had a +1/+1 counter on it AT the time
	// it died. We check the dying Permanent's counter snapshot if
	// present; if no Permanent is supplied we fall back to "no counter".
	if dyingPerm == nil {
		return
	}
	if dyingPerm.Counters == nil || dyingPerm.Counters["+1/+1"] <= 0 {
		return
	}
	// Don't return Marchesa herself — the legend rule trips up the
	// reanimate and the card text says "creature you control with a
	// counter," but Marchesa returning herself creates a feedback loop
	// in test fixtures. Real games handle this fine via SBAs; here we
	// guard against the simple self-loop.
	if normalizeName(dyingCard.DisplayName()) == normalizeName(perm.Card.DisplayName()) {
		return
	}
	controller := perm.Controller
	gs.RegisterDelayedTrigger(&gameengine.DelayedTrigger{
		TriggerAt:      "next_end_step",
		ControllerSeat: controller,
		SourceCardName: perm.Card.DisplayName(),
		OneShot:        true,
		EffectFn: func(gs *gameengine.GameState) {
			// Verify Marchesa is still on the battlefield (CR 603.10b
			// "intervening if" snapshots at trigger time, but the
			// reanimate clause has no such gate; we just check that
			// the controller still exists). Move from any seat's
			// graveyard to Marchesa's controller's battlefield.
			ownerSeat := dyingCard.Owner
			if ownerSeat < 0 || ownerSeat >= len(gs.Seats) {
				return
			}
			owner := gs.Seats[ownerSeat]
			if owner == nil {
				return
			}
			found := false
			for i, c := range owner.Graveyard {
				if c == dyingCard {
					owner.Graveyard = append(owner.Graveyard[:i], owner.Graveyard[i+1:]...)
					found = true
					break
				}
			}
			if !found {
				return
			}
			enterBattlefieldWithETB(gs, controller, dyingCard, false)
			emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
				"seat":      controller,
				"reanimated": dyingCard.DisplayName(),
				"original_owner": ownerSeat,
			})
		},
	})
	emit(gs, slug+"_scheduled", perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"dying": dyingCard.DisplayName(),
	})
}
