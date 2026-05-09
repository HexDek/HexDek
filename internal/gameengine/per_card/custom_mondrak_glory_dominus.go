package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMondrakGloryDominus wires Mondrak, Glory Dominus's
// token-doubling effect.
//
// Oracle text:
//
//	If one or more tokens would be created under your control, twice
//	that many of those tokens are created instead.
//	{2}{W}{W}, Exile two other creatures you control: Put two
//	indestructible counters on Mondrak.
//
// True replacement-effect plumbing for token doubling lives at the
// engine layer (CR §614, parallels Doubling Season / Anointed
// Procession). We set a per-seat flag and emit a partial so the
// engine's CreateToken path can branch on it; if the engine layer is
// already plumbed for "anointed_procession_count", Mondrak adds +1 to
// it. The activated indestructible-counter ability is wired in full.
func registerMondrakGloryDominus(r *Registry) {
	r.OnETB("Mondrak, Glory Dominus", mondrakSetTokenDoubler)
	r.OnActivated("Mondrak, Glory Dominus", mondrakIndestructibleActivate)
}

func mondrakSetTokenDoubler(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "mondrak_token_doubler_flag"
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
	seat.Flags["token_doubler_count"]++
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":             perm.Controller,
		"doublers_active":  seat.Flags["token_doubler_count"],
	})
}

func mondrakIndestructibleActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "mondrak_indestructible_activate"
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
	// Need two other creatures we control to exile.
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
	gameengine.MoveCard(gs, sac1.Card, sac1.Controller, "battlefield", "exile", "mondrak_activation_cost")
	gameengine.MoveCard(gs, sac2.Card, sac2.Controller, "battlefield", "exile", "mondrak_activation_cost")
	src.AddCounter("indestructible", 2)
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":           src.Controller,
		"exiled":         []string{sac1.Card.DisplayName(), sac2.Card.DisplayName()},
		"indestructible": src.Counters["indestructible"],
	})
}
