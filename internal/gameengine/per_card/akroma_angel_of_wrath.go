package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAkromaAngelOfWrath wires Akroma, Angel of Wrath.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	{5}{W}{W}{W}
//	Legendary Creature — Angel
//	6/6
//	Flying, first strike, vigilance, trample, haste, protection from
//	black and from red
//
// Pure keyword-stack vanilla — every ability is parsed via the AST
// keyword pipeline. Register-only stub for eligibility audits.
func registerAkromaAngelOfWrath(r *Registry) {
	r.OnETB("Akroma, Angel of Wrath", akromaAngelOfWrathETB)
}

func akromaAngelOfWrathETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "akroma_angel_of_wrath_keywords_only", perm.Card.DisplayName(),
		"all_abilities_handled_by_ast_keyword_pipeline_register_only_stub")
}
