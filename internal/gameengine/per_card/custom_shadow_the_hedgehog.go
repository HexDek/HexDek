package per_card

import (
	"github.com/hexdek/hexdek/internal/gameast"
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerShadowCustom adds Shadow's "flash/haste creature dies → draw"
// trigger that the auto-generated static stub omits.
//
// Oracle text:
//
//	Haste
//	Whenever Shadow the Hedgehog or another creature you control with
//	flash or haste dies, draw a card.
//	Chaos Control — Each spell you cast has split second if mana from
//	an artifact was spent to cast it.
//
// Haste itself is an AST keyword. The dies-trigger is wired here. The
// Chaos Control split-second rider is engine-side (mana provenance +
// stack interaction) and noted as a partial.
func registerShadowCustom(r *Registry) {
	r.OnTrigger("Shadow the Hedgehog", "creature_dies", shadowOnCreatureDies)
}

func shadowOnCreatureDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "shadow_creature_dies_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	dyingPerm, _ := ctx["perm"].(*gameengine.Permanent)
	dyingCard, _ := ctx["card"].(*gameengine.Card)
	dyingControllerAny := ctx["controller_seat"]
	dyingController, _ := dyingControllerAny.(int)
	if dyingController != perm.Controller {
		return
	}
	isShadow := false
	if dyingPerm != nil && dyingPerm.Card != nil &&
		normalizeName(dyingPerm.Card.DisplayName()) == normalizeName(perm.Card.DisplayName()) {
		isShadow = true
	}
	if !isShadow {
		hasFlashOrHaste := false
		if dyingPerm != nil {
			if dyingPerm.HasKeyword("flash") || dyingPerm.HasKeyword("haste") {
				hasFlashOrHaste = true
			}
		}
		if !hasFlashOrHaste && dyingCard != nil {
			if shadowCardHasKeyword(dyingCard, "flash") || shadowCardHasKeyword(dyingCard, "haste") {
				hasFlashOrHaste = true
			}
		}
		if !hasFlashOrHaste {
			return
		}
	}
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	dyingName := ""
	if dyingCard != nil {
		dyingName = dyingCard.DisplayName()
	}
	reason := "flash_or_haste"
	if isShadow {
		reason = "self"
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"dying":  dyingName,
		"reason": reason,
	})
}

func shadowCardHasKeyword(c *gameengine.Card, name string) bool {
	if c == nil || c.AST == nil {
		return false
	}
	for _, ab := range c.AST.Abilities {
		if kw, ok := ab.(*gameast.Keyword); ok && kw != nil && kw.Name == name {
			return true
		}
	}
	for _, t := range c.Types {
		if t == "kw:"+name {
			return true
		}
	}
	return false
}
