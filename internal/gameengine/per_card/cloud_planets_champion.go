package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCloudPlanetsChampion wires Cloud, Planet's Champion.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	During your turn, as long as Cloud is equipped, it has double strike
//	  and indestructible.
//	Equip abilities you activate that target Cloud cost {2} less to
//	  activate.
//
// Both clauses are continuous-effect statics (layers / cost reduction)
// outside the per-card trigger pipeline. We register an ETB partial flag
// so audits can find the gap.
func registerCloudPlanetsChampion(r *Registry) {
	r.OnETB("Cloud, Planet's Champion", cloudPlanetsChampionETB)
}

func cloudPlanetsChampionETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "cloud_planets_champion_static", perm.Card.DisplayName(),
		"equipped_double_strike_indestructible_and_equip_cost_reduction_not_modeled")
}
