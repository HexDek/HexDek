package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGiadaCustom adds the Angel-counter-stacking ETB rider that
// Giada's auto-generated activated stub omits.
//
// Oracle text:
//
//	Flying, vigilance
//	Each other Angel you control enters with an additional +1/+1 counter
//	on it for each Angel you already control.
//	{T}: Add {W}. Spend this mana only to cast an Angel spell.
//
// The "enters with additional counters" is technically a replacement
// effect; we approximate by adding the counters immediately after ETB.
// The mana ability is engine-side.
func registerGiadaCustom(r *Registry) {
	r.OnTrigger("Giada, Font of Hope", "permanent_etb", giadaAngelCounters)
}

func giadaAngelCounters(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "giada_angel_counters"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	entered, _ := ctx["perm"].(*gameengine.Permanent)
	if entered == nil || entered == perm || entered.Card == nil {
		return
	}
	if entered.Controller != perm.Controller {
		return
	}
	if !entered.IsCreature() || !cardSubtypeMatches(entered.Card, "angel") {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	// Count angels already in play (excluding the entering one and Giada
	// itself? The text says "each Angel you already control" — Giada IS
	// an Angel and counts; the entering Angel is NOT yet "already
	// controlled" at the moment its ETB replacement applies, so exclude).
	already := 0
	for _, p := range seat.Battlefield {
		if p == nil || p == entered || p.Card == nil {
			continue
		}
		if p.IsCreature() && cardSubtypeMatches(p.Card, "angel") {
			already++
		}
	}
	if already <= 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":     perm.Controller,
			"angel":    entered.Card.DisplayName(),
			"counters": 0,
		})
		return
	}
	entered.AddCounter("+1/+1", already)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"angel":    entered.Card.DisplayName(),
		"counters": already,
	})
}
