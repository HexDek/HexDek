package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMiirymSentinelWyrm wires Miirym, Sentinel Wyrm.
//
// Oracle text:
//
//	Flying, ward {2}
//	Whenever another nontoken Dragon you control enters, create a token
//	that's a copy of it, except the token isn't legendary.
//
// Implementation: ETB observer for nontoken dragons under the same
// controller. Creates a shallow card copy with "token" prepended to types
// and the legendary supertype stripped.
func registerMiirymSentinelWyrm(r *Registry) {
	r.OnTrigger("Miirym, Sentinel Wyrm", "nonland_permanent_etb", miirymDragonETB)
}

func miirymDragonETB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "miirym_dragon_token_copy"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	enteringSeat, _ := ctx["controller_seat"].(int)
	if enteringSeat != perm.Controller {
		return
	}
	enteringCard, _ := ctx["card"].(*gameengine.Card)
	if enteringCard == nil || enteringCard == perm.Card {
		return
	}
	if !cardHasType(enteringCard, "creature") || !cardHasType(enteringCard, "dragon") {
		return
	}
	if cardHasType(enteringCard, "token") {
		return
	}
	// Build a token copy: same characteristics, with "token" type and no
	// "legendary" supertype.
	types := []string{"token"}
	for _, t := range enteringCard.Types {
		if t == "legendary" {
			continue
		}
		types = append(types, t)
	}
	colors := append([]string(nil), enteringCard.Colors...)
	token := &gameengine.Card{
		Name:          enteringCard.Name + " (Miirym Token)",
		Owner:         perm.Controller,
		BasePower:     enteringCard.BasePower,
		BaseToughness: enteringCard.BaseToughness,
		Types:         types,
		Colors:        colors,
		TypeLine:      enteringCard.TypeLine,
		AST:           enteringCard.AST,
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"original": enteringCard.DisplayName(),
	})
}
