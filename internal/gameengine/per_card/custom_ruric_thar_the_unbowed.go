package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRuricTharTheUnbowedCustom adds Ruric Thar's noncreature-spell
// punisher trigger that the auto-generated static stub omits.
//
// Oracle text:
//
//	Vigilance, reach
//	Ruric Thar attacks each combat if able.
//	Whenever a player casts a noncreature spell, Ruric Thar deals 6
//	damage to that player.
//
// Vigilance / reach are AST keywords. The must-attack restriction is a
// combat-legality concern (engine territory, partial). The damage rider
// is wired here. Note "a player" — this fires on the controller's own
// noncreature spells too.
func registerRuricTharTheUnbowedCustom(r *Registry) {
	r.OnTrigger("Ruric Thar, the Unbowed", "noncreature_spell_cast", ruricTharBurnNoncreature)
}

func ruricTharBurnNoncreature(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ruric_thar_noncreature_burn"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	caster, ok := ctx["caster_seat"].(int)
	if !ok || caster < 0 || caster >= len(gs.Seats) {
		return
	}
	if isCreature, _ := ctx["is_creature"].(bool); isCreature {
		return
	}
	gameengine.DealDamage(gs, caster, 6, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":        perm.Controller,
		"caster_seat": caster,
		"damage":      6,
	})
}
