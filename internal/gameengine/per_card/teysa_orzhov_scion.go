package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTeysaOrzhovScion wires Teysa, Orzhov Scion.
//
// Oracle text:
//
//	{1}{W}{B}
//	Legendary Creature — Human Advisor
//	Sacrifice three white creatures: Exile target creature.
//	Whenever another black creature you control dies, create a 1/1 white
//	  Spirit creature token with flying.
//
// Implementation:
//   - "creature_dies" trigger: gated to controller_seat == Teysa's
//     controller, the dying creature is black, and not Teysa herself.
//     Creates a 1/1 white Spirit token with flying.
//   - Activated "sac three white creatures → exile target creature":
//     emitPartial — engine activated-ability pipeline does not yet
//     route sacrificial costs through per_card.
func registerTeysaOrzhovScion(r *Registry) {
	r.OnETB("Teysa, Orzhov Scion", teysaOrzhovScionETB)
	r.OnTrigger("Teysa, Orzhov Scion", "creature_dies", teysaOrzhovScionDies)
}

func teysaOrzhovScionETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "teysa_orzhov_scion_etb", perm.Card.DisplayName(),
		"activated_sac_three_white_creatures_exile_target_creature_partial")
}

func teysaOrzhovScionDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "teysa_orzhov_scion_spirit_token"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	controllerSeat, _ := ctx["controller_seat"].(int)
	if controllerSeat != perm.Controller {
		return
	}
	dyingCard, _ := ctx["card"].(*gameengine.Card)
	if dyingCard == nil {
		return
	}
	if dyingCard == perm.Card {
		return
	}
	black := false
	for _, c := range dyingCard.Colors {
		if c == "B" || c == "b" {
			black = true
			break
		}
	}
	if !black {
		return
	}
	token := &gameengine.Card{
		Name:          "Spirit Token",
		Owner:         perm.Controller,
		BasePower:     1,
		BaseToughness: 1,
		Types:         []string{"token", "creature", "spirit", "kw:flying"},
		Colors:        []string{"W"},
		TypeLine:      "Token Creature — Spirit",
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
		"died": dyingCard.DisplayName(),
	})
}
