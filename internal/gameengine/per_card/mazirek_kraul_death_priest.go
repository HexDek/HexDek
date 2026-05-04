package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMazirekKraulDeathPriest wires Mazirek, Kraul Death Priest.
//
// Oracle text:
//
//	Flying
//	Whenever a player sacrifices another permanent, put a +1/+1 counter
//	on each creature you control.
//
// Implementation: permanent_sacrificed observer. The "another" qualifier
// means we don't count Mazirek himself being sacrificed (and his event
// fires only on his own sacrifice). Increments counters on every creature
// the controller controls.
func registerMazirekKraulDeathPriest(r *Registry) {
	r.OnTrigger("Mazirek, Kraul Death Priest", "permanent_sacrificed", mazirekSacrifice)
}

func mazirekSacrifice(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "mazirek_sacrifice_counters"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	// Don't trigger if Mazirek himself was sacrificed.
	sacName, _ := ctx["card"].(string)
	if sacName == "" {
		if c, ok := ctx["card"].(*gameengine.Card); ok && c != nil {
			sacName = c.DisplayName()
		}
	}
	if sacName == perm.Card.DisplayName() {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	count := 0
	for _, p := range seat.Battlefield {
		if p == nil || !p.IsCreature() {
			continue
		}
		if p.Counters == nil {
			p.Counters = map[string]int{}
		}
		p.Counters["+1/+1"]++
		count++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"creatures": count,
	})
}
