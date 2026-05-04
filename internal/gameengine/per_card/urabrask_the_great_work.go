package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerUrabraskTheGreatWork wires Urabrask // The Great Work (DFC front).
//
// Oracle text (front: Urabrask):
//
//	{2}{R}{R}
//	Legendary Creature — Phyrexian Praetor
//	First strike
//	Whenever you cast an instant or sorcery spell, Urabrask deals 1
//	  damage to target opponent. Add {R}.
//	{R}: Exile Urabrask, then return it to the battlefield transformed
//	  under its owner's control. Activate only as a sorcery and only if
//	  you've cast three or more instant and/or sorcery spells this turn.
//
// Back face (The Great Work) is a Saga handled separately by the engine
// once transformation is wired.
//
// Implementation:
//   - First strike via AST.
//   - "instant_or_sorcery_cast" trigger gated to caster == controller:
//     1 damage to leftmost living opponent + add {R} to controller's
//     mana pool (mana add via a "mana_added" event for AI tracking;
//     emitPartial because the engine's mana pool requires symbol
//     paths not exposed here).
//   - Activated transformation: emitPartial.
func registerUrabraskTheGreatWork(r *Registry) {
	r.OnTrigger("Urabrask", "instant_or_sorcery_cast", urabraskInstantSorceryCast)
	r.OnTrigger("Urabrask // The Great Work", "instant_or_sorcery_cast", urabraskInstantSorceryCast)
	r.OnETB("Urabrask", urabraskETB)
	r.OnETB("Urabrask // The Great Work", urabraskETB)
}

func urabraskETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "urabrask_etb", perm.Card.DisplayName(),
		"activated_R_exile_and_return_transformed_partial")
}

func urabraskInstantSorceryCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "urabrask_is_cast_ping"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	caster, _ := ctx["caster_seat"].(int)
	if caster != perm.Controller {
		return
	}
	target := -1
	for _, opp := range gs.Opponents(perm.Controller) {
		if gs.Seats[opp] != nil && !gs.Seats[opp].Lost {
			target = opp
			break
		}
	}
	if target >= 0 {
		gs.Seats[target].Life -= 1
		gs.LogEvent(gameengine.Event{
			Kind:   "damage",
			Seat:   perm.Controller,
			Target: target,
			Source: perm.Card.DisplayName(),
			Amount: 1,
			Details: map[string]interface{}{
				"reason": "urabrask_is_ping",
			},
		})
	}
	gs.LogEvent(gameengine.Event{
		Kind:   "mana_added",
		Seat:   perm.Controller,
		Source: perm.Card.DisplayName(),
		Amount: 1,
		Details: map[string]interface{}{
			"color":  "R",
			"reason": "urabrask_is_ritual",
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"target": target,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"add_R_mana_pool_symbol_not_routed_through_pool_partial")
	_ = gs.CheckEnd()
}
