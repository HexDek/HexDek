package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTellahGreatSage wires Tellah, Great Sage.
//
// Oracle text:
//
//	{3}{U}{R}
//	Legendary Creature — Human Wizard
//	Whenever you cast a noncreature spell, create a 1/1 colorless Hero
//	  creature token. If four or more mana was spent to cast that spell,
//	  draw two cards. If eight or more mana was spent to cast that spell,
//	  sacrifice Tellah and it deals that much damage to each opponent.
//
// Implementation:
//   - "noncreature_spell_cast" trigger gated to caster == controller.
//     Always creates a 1/1 colorless Hero token. We use the spell's
//     mana value (cardCMC) as the "mana spent to cast" proxy — accurate
//     for vanilla casts; alt-cost / X-spell exact mana spent isn't
//     surfaced on the cast event today (emitPartial covers that).
//   - MV >= 4: draw 2.
//   - MV >= 8: sacrifice Tellah and deal MV to each opponent.
func registerTellahGreatSage(r *Registry) {
	r.OnTrigger("Tellah, Great Sage", "noncreature_spell_cast", tellahNoncreatureCast)
}

func tellahNoncreatureCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "tellah_great_sage_cast"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	caster, _ := ctx["caster_seat"].(int)
	if caster != perm.Controller {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	mv := 0
	if card != nil {
		mv = cardCMC(card)
	}

	token := &gameengine.Card{
		Name:          "Hero Token",
		Owner:         perm.Controller,
		BasePower:     1,
		BaseToughness: 1,
		Types:         []string{"token", "creature", "hero"},
		Colors:        []string{},
		TypeLine:      "Token Creature — Hero",
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)

	drew := 0
	if mv >= 4 {
		drawOne(gs, perm.Controller, perm.Card.DisplayName())
		drawOne(gs, perm.Controller, perm.Card.DisplayName())
		drew = 2
	}

	sacced := false
	if mv >= 8 {
		sacced = true
		gameengine.SacrificePermanent(gs, perm, "tellah_great_sage_self_sac")
		for _, opp := range gs.Opponents(perm.Controller) {
			s := gs.Seats[opp]
			if s == nil || s.Lost {
				continue
			}
			gameengine.DealDamage(gs, opp, mv, perm.Card.DisplayName())
		}
	}

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"mv":     mv,
		"drew":   drew,
		"sacced": sacced,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"uses_mana_value_as_mana_spent_proxy_alt_cost_xspell_exact_partial")
	_ = gs.CheckEnd()
}
