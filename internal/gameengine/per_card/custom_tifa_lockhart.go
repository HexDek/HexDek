package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTifaLockhartCustom adds Tifa's landfall power-double trigger
// that the auto-generated static stub omits.
//
// Oracle text:
//
//	Trample
//	Landfall — Whenever a land you control enters, double Tifa Lockhart's
//	power until end of turn.
//
// Trample is an AST keyword. Landfall doubles Tifa's power by appending
// a Modification with Power = current power for the rest of the turn —
// so 4 → 8 → 16 → … as more lands enter.
func registerTifaLockhartCustom(r *Registry) {
	r.OnTrigger("Tifa Lockhart", "permanent_etb", tifaLockhartLandfall)
}

func tifaLockhartLandfall(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "tifa_lockhart_landfall"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	entered, _ := ctx["perm"].(*gameengine.Permanent)
	if entered == nil || entered.Card == nil || !entered.IsLand() {
		return
	}
	if entered.Controller != perm.Controller {
		return
	}
	cur := perm.Power()
	if cur <= 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"land_entered":  entered.Card.DisplayName(),
			"current_power": cur,
			"added_power":   0,
		})
		return
	}
	perm.Modifications = append(perm.Modifications, gameengine.Modification{
		Power:     cur,
		Toughness: 0,
		Duration:  "until_end_of_turn",
		Timestamp: gs.NextTimestamp(),
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"land_entered":  entered.Card.DisplayName(),
		"current_power": cur,
		"added_power":   cur,
		"new_power":     perm.Power(),
	})
}
