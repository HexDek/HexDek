package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerK9MarkI wires K-9, Mark I.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Negative — As long as K-9 is untapped, other legendary creatures you
//	  control have ward {1}.
//	Affirmative — {1}{U}, {T}: Target legendary creature can't be
//	  blocked this turn.
//	Doctor's companion (You can have two commanders if the other is the
//	  Doctor.)
//
// Both abilities live in static / ability layers outside the per-card
// trigger pipeline — ward grant is conditional on tapped state, and
// "can't be blocked" is a UEOT effect tracked elsewhere. Register an
// ETB partial flag so audits can find the gap.
func registerK9MarkI(r *Registry) {
	r.OnETB("K-9, Mark I", k9MarkIETB)
	r.OnActivated("K-9, Mark I", k9MarkIActivate)
}

func k9MarkIETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "k9_mark_i_negative_static", perm.Card.DisplayName(),
		"ward_grant_to_legendary_creatures_when_untapped_not_modeled")
}

func k9MarkIActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	if gs == nil || src == nil {
		return
	}
	if src.Tapped {
		return
	}
	src.Tapped = true
	emitPartial(gs, "k9_mark_i_affirmative", src.Card.DisplayName(),
		"cant_be_blocked_ueot_grant_not_modeled")
}
