package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTophGreatestEarthbender wires Toph, Greatest Earthbender.
//
// Oracle text:
//
//   When Toph enters, earthbend X, where X is the amount of mana spent to cast her.
//   Land creatures you control have double strike.
//
// Implementation:
//   - ETB: pick a non-creature land we control, mark it as a 0/0
//     creature with haste, and put X +1/+1 counters on it.
//     X is read from perm.Card.CMC (mana spent to cast). The "amount of
//     mana spent" isn't surfaced cleanly to ETB handlers, so CMC is the
//     best proxy.
//   - The static "Land creatures you control have double strike" is a
//     continuous granted-keyword effect (Layer 6) and is not modeled
//     here.
func registerTophGreatestEarthbender(r *Registry) {
	r.OnETB("Toph, Greatest Earthbender", tophGreatestEarthbenderETB)
}

func tophGreatestEarthbenderETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "toph_greatest_earthbender_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil {
		return
	}
	x := 0
	if perm.Card != nil {
		x = perm.Card.CMC
	}
	if x <= 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   seat,
			"x":      0,
			"reason": "no_mana_spent_recorded",
		})
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"land_creatures_double_strike_unimplemented")
		return
	}
	// Pick a non-creature land we control to earthbend.
	var target *gameengine.Permanent
	for _, p := range s.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if p.IsLand() && !p.IsCreature() {
			target = p
			break
		}
	}
	if target == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   seat,
			"x":      x,
			"reason": "no_land_target",
		})
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"land_creatures_double_strike_unimplemented")
		return
	}
	if target.Flags == nil {
		target.Flags = map[string]int{}
	}
	target.Flags["earthbent"] = 1
	target.Flags["temp_haste"] = 1
	if target.Counters == nil {
		target.Counters = map[string]int{}
	}
	target.Counters["+1/+1"] += x
	target.SummoningSick = false
	gs.InvalidateCharacteristicsCache()
	gs.LogEvent(gameengine.Event{
		Kind: "earthbend", Seat: seat,
		Source: perm.Card.DisplayName(), Amount: x,
		Details: map[string]interface{}{
			"target": target.Card.DisplayName(),
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     seat,
		"x":        x,
		"target":   target.Card.DisplayName(),
		"counters": x,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"land_creatures_double_strike_unimplemented")
}
