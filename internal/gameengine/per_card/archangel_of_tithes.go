package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerArchangelOfTithes wires Archangel of Tithes (Muninn parser-gap
// rank ~140, attack-tax static).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{1}{W}{W}{W}
//	Creature — Angel
//	Flying
//	As long as this creature is untapped, creatures can't attack you or
//	planeswalkers you control unless their controller pays {1} for each
//	of those creatures.
//	As long as this creature is attacking, creatures can't block unless
//	their controller pays {1} for each of those creatures.
//
// Implementation:
//   - Flying via AST keyword pipeline.
//   - Both abilities are CR §509-style cost-replacement statics that
//     fire during the §509a declare-attackers / §509b declare-blockers
//     steps. The engine doesn't expose a per-card hook into either
//     step's cost-application phase, so we cannot enforce either tax
//     today.
//   - We DO stamp deterrent flags on permanent_etb so opposing hats
//     can read them (`tax_attack:1`, `tax_block_when_attacking:1`) and
//     factor the cost into their attack/block planning heuristics —
//     this is the same approach used by smothering tithe / archon of
//     emeria style stubs. The flags are cleared at LTB.
//   - emitPartial documents the un-enforced cost replacement.
func registerArchangelOfTithes(r *Registry) {
	r.OnETB("Archangel of Tithes", archangelOfTithesETB)
	r.OnTrigger("Archangel of Tithes", "permanent_ltb", archangelOfTithesLTB)
}

func archangelOfTithesETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "archangel_of_tithes_tax_static"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["tax_attack"] = 1
	perm.Flags["tax_block_when_attacking"] = 1
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"static_attack_and_block_cost_replacement_not_enforced_pending_combat_step_cost_hook")
}

func archangelOfTithesLTB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	leaving, _ := ctx["perm"].(*gameengine.Permanent)
	if leaving != perm || perm.Flags == nil {
		return
	}
	delete(perm.Flags, "tax_attack")
	delete(perm.Flags, "tax_block_when_attacking")
}
