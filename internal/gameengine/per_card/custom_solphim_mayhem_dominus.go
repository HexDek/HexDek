package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSolphimMayhemDominus wires Solphim, Mayhem Dominus.
//
// Oracle text:
//
//	If a source you control would deal noncombat damage to a
//	permanent or player, it deals twice that much damage to that
//	permanent or player instead.
//	{2}{R}{R}, Exile two other creatures you control: Put two
//	indestructible counters on Solphim.
//
// The damage-doubling replacement is engine-layer (CR §614, parallels
// Furnace of Rath). We set a per-seat flag and emit partial so the
// engine's DealDamage path can branch on it. The activated
// indestructible-counter ability is wired in full (mirrors Mondrak).
func registerSolphimMayhemDominus(r *Registry) {
	r.OnETB("Solphim, Mayhem Dominus", solphimSetDamageDoublerFlag)
	r.OnActivated("Solphim, Mayhem Dominus", solphimIndestructibleActivate)
}

func solphimSetDamageDoublerFlag(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "solphim_noncombat_damage_doubler_flag"
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
	seat.Flags["noncombat_damage_doubler_count"]++
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"doublers": seat.Flags["noncombat_damage_doubler_count"],
	})
	emitPartial(gs, "solphim_doubles_noncombat_damage", perm.Card.DisplayName(),
		"noncombat damage doubling requires DealDamage replacement-effect hook; flag set for downstream")
}

func solphimIndestructibleActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "solphim_indestructible_activate"
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
	gameengine.MoveCard(gs, sac1.Card, sac1.Controller, "battlefield", "exile", "solphim_activation_cost")
	gameengine.MoveCard(gs, sac2.Card, sac2.Controller, "battlefield", "exile", "solphim_activation_cost")
	src.AddCounter("indestructible", 2)
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":           src.Controller,
		"exiled":         []string{sac1.Card.DisplayName(), sac2.Card.DisplayName()},
		"indestructible": src.Counters["indestructible"],
	})
}
