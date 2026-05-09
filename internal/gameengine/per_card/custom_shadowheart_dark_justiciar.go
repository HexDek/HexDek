package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerShadowheartDarkJusticiarCustom implements Shadowheart's
// sac-for-card-draw activation. The auto-generated stub is a no-op.
//
// Oracle text:
//
//	{1}{B}, {T}, Sacrifice another creature: Draw X cards, where X is
//	that creature's power.
//	Choose a Background (You can have a Background as a second commander.)
//
// Implementation notes:
//   - Mana cost {1}{B} = 2 generic; engine activation dispatch handles
//     it normally, but we defensively check seat.ManaPool >= 2 since
//     non-engine callers (test fixtures, replay paths) skip the
//     dispatcher.
//   - Tap cost: defensive check + set Tapped=true.
//   - Sacrifice cost: pick the highest-power non-commander other
//     creature we control. Picking high-power maximizes the X draw,
//     which is the whole point of the activation. Avoid commanders
//     since they'd go to the command zone (tax) instead of giving
//     value.
//   - Effect: draw X = sacrificed creature's power (capped at 0; can't
//     draw negative). Use MoveCard so §614 replacements (e.g. Drannith
//     Magistrate analogues) and counter-trigger work correctly.
//   - Background partner clause is engine-side commander-zone wiring;
//     emitPartial.
func registerShadowheartDarkJusticiarCustom(r *Registry) {
	r.OnActivated("Shadowheart, Dark Justiciar", shadowheartSacDraw)
}

func shadowheartSacDraw(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "shadowheart_sac_draw_x"
	if gs == nil || src == nil || src.Card == nil {
		return
	}
	if src.Tapped {
		emitFail(gs, slug, src.Card.DisplayName(), "already_tapped", nil)
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

	// Pick the best (highest-power) non-commander, non-self creature
	// to sacrifice. Tiebreak: highest Timestamp (newest = most
	// expendable, longest-established stays).
	var victim *gameengine.Permanent
	bestPower := -1
	bestTS := -1
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil || p == src || !p.IsCreature() {
			continue
		}
		if yawgmothIsCommander(gs, p) {
			continue
		}
		pow := gs.PowerOf(p)
		if pow > bestPower || (pow == bestPower && p.Timestamp > bestTS) {
			bestPower = pow
			bestTS = p.Timestamp
			victim = p
		}
	}
	if victim == nil {
		emitFail(gs, slug, src.Card.DisplayName(), "no_creature_to_sacrifice", nil)
		return
	}

	// Pay costs.
	seat.ManaPool -= 2
	gameengine.SyncManaAfterSpend(seat)
	src.Tapped = true
	victimName := victim.Card.DisplayName()
	gameengine.SacrificePermanent(gs, victim, "shadowheart_sac_cost")

	// Effect — draw X cards (X = sacrificed creature's power, ≥0).
	x := bestPower
	if x < 0 {
		x = 0
	}
	drawn := 0
	for i := 0; i < x; i++ {
		if len(seat.Library) == 0 {
			break
		}
		card := seat.Library[0]
		gameengine.MoveCard(gs, card, src.Controller, "library", "hand", "shadowheart_draw")
		drawn++
	}

	emitPartial(gs, slug, src.Card.DisplayName(),
		"Choose a Background partner clause: command-zone selection not modeled at per-card layer")

	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":          src.Controller,
		"sacrificed":    victimName,
		"sacrificed_pw": bestPower,
		"drew":          drawn,
	})
	_ = gs.CheckEnd()
}
