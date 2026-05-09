package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerChocoCustom adds Choco's landfall pump that the auto-generated
// static stub omits.
//
// Oracle text:
//
//	Whenever one or more Birds you control attack, look at that many
//	cards from the top of your library. You may put one of them into
//	your hand. Then put any number of land cards from among them onto
//	the battlefield tapped and the rest into your graveyard.
//	Landfall — Whenever a land you control enters, Choco gets +1/+0
//	until end of turn.
//
// The Bird-attack trigger requires looking-at-N + multi-zone choices
// the engine doesn't yet expose to per_card; we stub it with
// emitPartial. The landfall pump is wired in full.
func registerChocoCustom(r *Registry) {
	r.OnTrigger("Choco, Seeker of Paradise", "permanent_etb", chocoLandfall)
	r.OnTrigger("Choco, Seeker of Paradise", "creature_attacks", chocoBirdsAttackPartial)
}

func chocoLandfall(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "choco_landfall_pump"
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
	perm.Modifications = append(perm.Modifications, gameengine.Modification{
		Power:     1,
		Toughness: 0,
		Duration:  "until_end_of_turn",
		Timestamp: gs.NextTimestamp(),
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"land_entered": entered.Card.DisplayName(),
		"new_power":    perm.Power(),
	})
}

func chocoBirdsAttackPartial(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk == nil || atk.Card == nil {
		return
	}
	if atk.Controller != perm.Controller {
		return
	}
	if !cardSubtypeMatches(atk.Card, "bird") {
		return
	}
	emitPartial(gs, "choco_bird_attack_dig", perm.Card.DisplayName(),
		"bird-attack look-at-N + zone-distribution choice not yet exposed to per_card")
}
