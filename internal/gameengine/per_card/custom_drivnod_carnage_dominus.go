package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerDrivnodCarnageDominus wires Drivnod, Carnage Dominus.
//
// Oracle text:
//
//	If a creature dying causes a triggered ability of a permanent you
//	control to trigger, that ability triggers an additional time.
//	{2}{B}{B}, Exile two other creatures you control: Put two
//	indestructible counters on Drivnod, Carnage Dominus.
//
// True double-firing of death triggers requires engine support that
// intercepts trigger resolution (parallels Panharmonicon for ETBs).
// We set a per-seat flag and emit a partial so downstream consumers
// (engine death-trigger dispatch) can branch on it. The activated
// indestructible-counter ability is wired in full.
func registerDrivnodCarnageDominus(r *Registry) {
	r.OnETB("Drivnod, Carnage Dominus", drivnodSetDeathDoublerFlag)
	r.OnActivated("Drivnod, Carnage Dominus", drivnodIndestructibleActivate)
}

func drivnodSetDeathDoublerFlag(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "drivnod_death_trigger_doubler_flag"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if seat.Flags == nil {
		seat.Flags = map[string]int{}
	}
	seat.Flags["death_trigger_doubler_count"]++
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"doublers": seat.Flags["death_trigger_doubler_count"],
	})
	emitPartial(gs, "drivnod_doubles_death_triggers", perm.Card.DisplayName(),
		"death-trigger doubling requires resolve-time hook to duplicate stack copies; flag set for downstream")
}

func drivnodIndestructibleActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "drivnod_indestructible_activate"
	if gs == nil || src == nil {
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	if seat.ManaPool < 4 {
		emitFail(gs, slug, src.Card.DisplayName(), "insufficient_mana", map[string]interface{}{
			"mana_pool": seat.ManaPool,
		})
		return
	}
	var sac1, sac2 *gameengine.Permanent
	for _, p := range seat.Battlefield {
		if p == nil || p == src || !p.IsCreature() {
			continue
		}
		if sac1 == nil {
			sac1 = p
			continue
		}
		if sac2 == nil {
			sac2 = p
			break
		}
	}
	if sac1 == nil || sac2 == nil {
		emitFail(gs, slug, src.Card.DisplayName(), "fewer_than_2_other_creatures", nil)
		return
	}
	seat.ManaPool -= 4
	gameengine.MoveCard(gs, sac1.Card, sac1.Controller, "battlefield", "exile", "drivnod_activation_cost")
	gameengine.MoveCard(gs, sac2.Card, sac2.Controller, "battlefield", "exile", "drivnod_activation_cost")
	if src.Counters == nil {
		src.Counters = map[string]int{}
	}
	src.Counters["indestructible"] += 2
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":           src.Controller,
		"exiled":         []string{sac1.Card.DisplayName(), sac2.Card.DisplayName()},
		"indestructible": src.Counters["indestructible"],
	})
}
