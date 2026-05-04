package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerLaughingJasperFlint wires Laughing Jasper Flint.
//
// Oracle text:
//
//	Creatures you control but don't own are Mercenaries in addition to
//	their other types.
//	At the beginning of your upkeep, exile the top X cards of target
//	opponent's library, where X is the number of outlaws you control.
//	Until end of turn, you may cast spells from among those cards, and
//	mana of any type can be spent to cast those spells.
//
// Implementation: upkeep exiles X cards from the opponent with the most
// cards left in library. Free-casting from exile until EOT is non-trivial
// — emitPartial that clause.
func registerLaughingJasperFlint(r *Registry) {
	r.OnTrigger("Laughing Jasper Flint", "upkeep_controller", laughingJasperUpkeep)
}

func laughingJasperOutlawCount(seat *gameengine.Seat) int {
	if seat == nil {
		return 0
	}
	n := 0
	outlawTypes := []string{"assassin", "mercenary", "pirate", "rogue", "warlock"}
	for _, p := range seat.Battlefield {
		if p == nil || !p.IsCreature() || p.Card == nil {
			continue
		}
		for _, t := range outlawTypes {
			if cardHasType(p.Card, t) {
				n++
				break
			}
		}
	}
	return n
}

func laughingJasperUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "laughing_jasper_exile_opponent_library"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	x := laughingJasperOutlawCount(seat)
	if x <= 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":    perm.Controller,
			"outlaws": 0,
		})
		return
	}
	// Pick opponent with most library cards.
	target := -1
	bestLib := -1
	for i, s := range gs.Seats {
		if s == nil || s.Lost || i == perm.Controller {
			continue
		}
		if len(s.Library) > bestLib {
			bestLib = len(s.Library)
			target = i
		}
	}
	if target < 0 {
		return
	}
	exiled := 0
	for i := 0; i < x && len(gs.Seats[target].Library) > 0; i++ {
		top := gs.Seats[target].Library[0]
		moveCardBetweenZones(gs, target, top, "library", "exile", "laughing_jasper_exile")
		exiled++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         perm.Controller,
		"target_seat":  target,
		"outlaws":      x,
		"exiled":       exiled,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"free_cast_from_exiled_opponent_cards_until_eot_unimplemented")
}
