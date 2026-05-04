package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRaffWeatherlightStalwart wires Raff, Weatherlight Stalwart.
//
// Oracle text:
//
//	Whenever you cast an instant or sorcery spell, you may tap two
//	untapped creatures you control. If you do, draw a card.
//	{3}{W}{W}: Creatures you control get +1/+1 and gain vigilance
//	until end of turn.
//
// On instant_or_sorcery_cast: try to tap two untapped non-Raff
// creatures and draw a card. Activated anthem is left as parser gap.
func registerRaffWeatherlightStalwart(r *Registry) {
	r.OnTrigger("Raff, Weatherlight Stalwart", "instant_or_sorcery_cast", raffSpellCast)
}

func raffSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "raff_tap_two_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	var first, second *gameengine.Permanent
	for _, p := range seat.Battlefield {
		if p == nil || p == perm || p.Tapped || !p.IsCreature() {
			continue
		}
		if first == nil {
			first = p
		} else if second == nil {
			second = p
			break
		}
	}
	if first == nil || second == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat": perm.Controller,
			"paid": false,
		})
		return
	}
	first.Tapped = true
	second.Tapped = true
	if len(seat.Library) > 0 {
		c := seat.Library[0]
		gameengine.MoveCard(gs, c, perm.Controller, "library", "hand", "draw")
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"paid":   true,
		"tapped": []string{first.Card.DisplayName(), second.Card.DisplayName()},
	})
}
