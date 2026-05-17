package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCrownOfGondor wires Crown of Gondor (Muninn parser-gap #42,
// 20,907 hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{3}
//	Legendary Artifact — Equipment
//	Equipped creature gets +1/+1 for each creature you control.
//	When a legendary creature you control enters, if there is no
//	monarch, you become the monarch.
//	Equip {4}. This ability costs {3} less to activate if you're the
//	monarch.
//
// Implementation:
//   - "When a legendary creature you control enters, if there is no
//     monarch, you become the monarch" — listen on permanent_etb, gate
//     on (entering.Controller == perm.Controller, entering is creature,
//     entering is legendary, !HasMonarch). BecomeMonarch on success.
//   - Equipped-creature anthem (+1/+1 per creature you control) is a
//     static layer-7 effect that the AST keyword pipeline handles via
//     the EquippedCreatureGetsCounters node when present; we don't
//     re-implement it. emitPartial flags the gap so Muninn can confirm
//     the AST layer is producing the buff.
//   - Equip {4} (with {3} discount if you're the monarch) is an
//     activated ability with a conditional cost — the activated-cost
//     conditional-discount layer is not yet wired through per_card;
//     emitPartial.
func registerCrownOfGondor(r *Registry) {
	r.OnTrigger("Crown of Gondor", "permanent_etb", crownOfGondorPermETB)
}

func crownOfGondorPermETB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "crown_of_gondor_monarch"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	entering, _ := ctx["perm"].(*gameengine.Permanent)
	if entering == nil {
		entering, _ = ctx["permanent"].(*gameengine.Permanent)
	}
	if entering == nil || entering == perm || entering.Card == nil {
		return
	}
	if entering.Controller != perm.Controller {
		return
	}
	if !entering.IsCreature() {
		return
	}
	if !cardHasType(entering.Card, "legendary") {
		return
	}
	if gs.Flags != nil && gs.Flags["has_monarch"] == 1 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":       perm.Controller,
			"entered":    entering.Card.DisplayName(),
			"triggered":  false,
			"reason":     "monarch_exists",
		})
		return
	}
	gameengine.BecomeMonarch(gs, perm.Controller)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"entered":   entering.Card.DisplayName(),
		"triggered": true,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"equip_4_monarch_discount_and_per_creature_anthem_via_ast_layer")
}
