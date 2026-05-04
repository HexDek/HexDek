package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerUurgSpawnOfTurg wires Uurg, Spawn of Turg.
//
// Oracle text:
//
//	{B}{B}{G}
//	Legendary Creature — Frog Beast
//	Uurg's power is equal to the number of land cards in your graveyard.
//	At the beginning of your upkeep, surveil 1.
//	{B}{G}, Sacrifice a land: You gain 2 life.
//
// Implementation:
//   - Power equals land cards in graveyard: implemented via a "set_power"
//     flag refresh hook that fires on ETB and during upkeep. We update
//     perm.Flags["set_power"] = land_count_in_graveyard so the layers
//     pipeline can consult it (emitPartial because the engine's CDA
//     pipeline does not yet route through per_card cleanly).
//   - Upkeep: surveil 1 via gameengine.Surveil.
//   - Activated {B}{G} sac-land → 2 life: emitPartial.
func registerUurgSpawnOfTurg(r *Registry) {
	r.OnETB("Uurg, Spawn of Turg", uurgETB)
	r.OnTrigger("Uurg, Spawn of Turg", "upkeep_controller", uurgUpkeep)
}

func uurgETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	uurgRefreshPower(gs, perm)
	emitPartial(gs, "uurg_etb", perm.Card.DisplayName(),
		"activated_BG_sac_land_gain_2_partial")
}

func uurgUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "uurg_upkeep_surveil"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	gameengine.Surveil(gs, perm.Controller, 1)
	uurgRefreshPower(gs, perm)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}

func uurgRefreshPower(gs *gameengine.GameState, perm *gameengine.Permanent) {
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	count := 0
	for _, c := range seat.Graveyard {
		if c == nil {
			continue
		}
		if cardHasType(c, "land") {
			count++
		}
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["set_power"] = count
	emit(gs, "uurg_refresh_power", perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"power": count,
	})
}
