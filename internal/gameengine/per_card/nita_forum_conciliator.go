package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerNitaForumConciliator wires Nita, Forum Conciliator.
//
// Oracle text:
//
//	Whenever you cast a spell you don't own, put a +1/+1 counter on
//	each creature you control.
//	{2}, Sacrifice another creature: Exile target instant or sorcery
//	card from an opponent's graveyard. You may cast it this turn ...
//	Activate only as a sorcery.
//
// We wire the cast trigger. The activated graveyard-cast is left as a
// parser gap (cross-graveyard cast machinery isn't general here).
func registerNitaForumConciliator(r *Registry) {
	r.OnTrigger("Nita, Forum Conciliator", "spell_cast", nitaSpellCast)
}

func nitaSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "nita_counters_on_foreign_cast"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil {
		return
	}
	if card.Owner == perm.Controller {
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
		p.AddCounter("+1/+1", 1)
		count++
	}
	if count > 0 {
		gs.InvalidateCharacteristicsCache()
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"creatures": count,
	})
}
