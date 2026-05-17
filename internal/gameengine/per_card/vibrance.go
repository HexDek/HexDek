package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerVibrance wires Vibrance.
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	When this creature enters, if {R}{R} was spent to cast it, this
//	creature deals 3 damage to any target.
//	When this creature enters, if {G}{G} was spent to cast it, search
//	your library for a land card, reveal it, put it into your hand, then
//	shuffle. You gain 2 life.
//	Evoke {R/G}{R/G}
//
// Implementation (Muninn gap #11 — 91K hits):
//   - OnETB fires both color-gated halves. The mana cost is hybrid
//     {3}{R/G}{R/G}, so the actual pip-payment choice belongs to the
//     caster at cast time. The engine doesn't yet track payment of
//     hybrid pips per-cast (resolve.go's "mana_spent" condition defaults
//     true — see internal/gameengine/resolve.go:413), so we follow the
//     engine convention and fire BOTH modes when Vibrance enters via the
//     cast pipeline. We pessimise when entered without being cast (e.g.
//     reanimated, blinked) by firing neither.
//   - Damage mode: 3 damage to the highest-life living opponent
//     (deterministic, matches Aragorn the Uniter's any-target shortcut).
//   - Land mode: take the first land card in the controller's library,
//     move it to hand, shuffle, gain 2 life.
//   - Evoke {R/G}{R/G} is a cast-time alternative cost handled by the
//     cast pipeline (when supported); flagged via emitPartial.
//   - The "mana spent" gating is conservatively approximated; we
//     emitPartial so Muninn can keep tracking the residual gap.
func registerVibrance(r *Registry) {
	r.OnETB("Vibrance", vibranceETB)
}

func vibranceETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "vibrance_color_etb"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil || s.Lost {
		return
	}

	// "If {C}{C} was spent to cast it" — both halves are conditioned on
	// the spell having actually been cast. If Vibrance entered another
	// way (blink, reanimate, Sneak Attack), neither mode fires.
	wasCast := perm.Flags != nil && perm.Flags["was_cast"] != 0

	modes := []string{}

	if wasCast {
		// Red mode: 3 damage to any target (deterministic: highest-life
		// opponent so it pressures the closest threat).
		if target := aragornPickDamageTarget(gs, seat); target >= 0 {
			opp := gs.Seats[target]
			if opp != nil && !opp.Lost {
				gameengine.DealDamage(gs, target, 3, perm.Card.DisplayName())
				modes = append(modes, "red_3_damage")
			}
		}

		// Green mode: land tutor to hand + 2 life.
		var land *gameengine.Card
		for _, c := range s.Library {
			if c == nil {
				continue
			}
			if cardHasType(c, "land") {
				land = c
				break
			}
		}
		if land != nil {
			gameengine.MoveCard(gs, land, seat, "library", "hand", "vibrance_land_search")
			shuffleLibraryPerCard(gs, seat)
			gs.LogEvent(gameengine.Event{
				Kind:   "search_library",
				Seat:   seat,
				Source: perm.Card.DisplayName(),
				Details: map[string]interface{}{
					"found":  []string{land.DisplayName()},
					"reason": "vibrance_green",
				},
			})
			gameengine.GainLife(gs, seat, 2, perm.Card.DisplayName())
			modes = append(modes, "green_land_and_2_life")
		} else {
			// Still shuffle per the cost: "search your library, … then
			// shuffle" runs even on whiff (CR §701.19c).
			shuffleLibraryPerCard(gs, seat)
			gameengine.GainLife(gs, seat, 2, perm.Card.DisplayName())
			modes = append(modes, "green_2_life_no_land")
		}
	}

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     seat,
		"was_cast": wasCast,
		"modes":    modes,
	})

	emitPartial(gs, slug, perm.Card.DisplayName(),
		"hybrid_pip_payment_tracking_unmodeled_both_modes_fire_when_cast")
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"evoke_alt_cost_handled_by_cast_pipeline_only_when_supported")
}
