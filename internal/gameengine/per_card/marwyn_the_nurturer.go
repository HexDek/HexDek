package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMarwynTheNurturer wires Marwyn, the Nurturer.
//
// Oracle text:
//
//	Whenever another Elf you control enters, put a +1/+1 counter on
//	Marwyn.
//	{T}: Add an amount of {G} equal to Marwyn's power.
//
// Implementation: ETB listener for elf creatures. Activated mana ability
// is engine-handled via the AST tap-for-mana path; we still install a
// hook so a sim that drives activations directly produces correct mana.
func registerMarwynTheNurturer(r *Registry) {
	r.OnTrigger("Marwyn, the Nurturer", "nonland_permanent_etb", marwynElfETB)
	r.OnActivated("Marwyn, the Nurturer", marwynActivated)
}

func marwynElfETB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "marwyn_elf_counter"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	enteringSeat, _ := ctx["controller_seat"].(int)
	if enteringSeat != perm.Controller {
		return
	}
	enteringCard, _ := ctx["card"].(*gameengine.Card)
	if enteringCard == nil {
		return
	}
	if !cardHasType(enteringCard, "creature") || !cardHasType(enteringCard, "elf") {
		return
	}
	if enteringCard == perm.Card {
		return
	}
	if perm.Counters == nil {
		perm.Counters = map[string]int{}
	}
	perm.Counters["+1/+1"]++
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"elf":      enteringCard.DisplayName(),
		"counters": perm.Counters["+1/+1"],
	})
}

func marwynActivated(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "marwyn_tap_for_green"
	if gs == nil || src == nil || src.Card == nil {
		return
	}
	if src.Tapped {
		emitFail(gs, slug, src.Card.DisplayName(), "already_tapped", nil)
		return
	}
	src.Tapped = true
	power := src.Card.BasePower
	if src.Counters != nil {
		power += src.Counters["+1/+1"]
	}
	if power <= 0 {
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":    src.Controller,
			"added_g": 0,
		})
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	if seat.Mana != nil {
		seat.Mana.G += power
	} else {
		seat.ManaPool += power
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":    src.Controller,
		"added_g": power,
	})
}
