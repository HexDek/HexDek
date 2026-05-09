package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerZopandrelHungerDominus wires Zopandrel, Hunger Dominus.
//
// Oracle text:
//
//	Whenever one or more creatures you control deal combat damage to
//	a player, draw a card and you gain that much life.
//	{2}{G}{G}, Exile two other creatures you control: Put two
//	indestructible counters on Zopandrel, Hunger Dominus.
//
// We listen on `combat_damage_to_player`. If the source creature is
// controlled by Zopandrel's controller, we draw 1 and gain `amount`
// life. We DON'T fully model the "one or more" once-per-combat
// dedupe — the engine fires this trigger per damage event, so a
// 3-attacker swing currently draws/lifegains 3 times (an over-count
// that's logged via partial).
func registerZopandrelHungerDominus(r *Registry) {
	r.OnTrigger("Zopandrel, Hunger Dominus", "combat_damage_to_player", zopandrelDrawAndGain)
	r.OnActivated("Zopandrel, Hunger Dominus", zopandrelIndestructibleActivate)
}

func zopandrelDrawAndGain(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "zopandrel_combat_dmg_draw_gain"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	srcSeat, ok := ctx["source_seat"].(int)
	if !ok {
		// Fallback: try source_perm.
		if sp, ok2 := ctx["source_perm"].(*gameengine.Permanent); ok2 && sp != nil {
			srcSeat = sp.Controller
		} else {
			return
		}
	}
	if srcSeat != perm.Controller {
		return
	}
	amount := 0
	if v, ok := ctx["amount"].(int); ok {
		amount = v
	}
	if amount <= 0 {
		return
	}
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	gameengine.GainLife(gs, perm.Controller, amount, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"life":   amount,
	})
	emitPartial(gs, "zopandrel_one_or_more_dedupe", perm.Card.DisplayName(),
		"\"one or more\" combat-damage dedupe (single trigger per swing) not enforced; per-source firing")
}

func zopandrelIndestructibleActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "zopandrel_indestructible_activate"
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
	gameengine.MoveCard(gs, sac1.Card, sac1.Controller, "battlefield", "exile", "zopandrel_activation_cost")
	gameengine.MoveCard(gs, sac2.Card, sac2.Controller, "battlefield", "exile", "zopandrel_activation_cost")
	src.AddCounter("indestructible", 2)
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":           src.Controller,
		"exiled":         []string{sac1.Card.DisplayName(), sac2.Card.DisplayName()},
		"indestructible": src.Counters["indestructible"],
	})
}
