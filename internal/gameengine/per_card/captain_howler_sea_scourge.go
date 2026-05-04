package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCaptainHowlerSeaScourge wires Captain Howler, Sea Scourge.
//
// Oracle text:
//
//	Ward—{2}, Pay 2 life.
//	Whenever you discard one or more cards, target creature gets +2/+0
//	until end of turn for each card discarded this way. Whenever that
//	creature deals combat damage to a player this turn, you draw a
//	card.
//
// Cards-discarded count is tracked via the engine's discard event. We
// pump the controller's best creature for each card discarded; the
// "draw on damage" bookkeeping is non-trivial and emitPartial'd.
func registerCaptainHowlerSeaScourge(r *Registry) {
	r.OnTrigger("Captain Howler, Sea Scourge", "card_discarded", captainHowlerDiscard)
}

func captainHowlerDiscard(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "captain_howler_discard_pump"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	discarderSeat, _ := ctx["seat"].(int)
	if discarderSeat != perm.Controller {
		return
	}
	count, _ := ctx["count"].(int)
	if count <= 0 {
		count = 1
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	var best *gameengine.Permanent
	bestPow := -1
	for _, p := range seat.Battlefield {
		if p == nil || !p.IsCreature() {
			continue
		}
		if pow := p.Power(); pow > bestPow {
			best = p
			bestPow = pow
		}
	}
	if best == nil {
		return
	}
	best.Modifications = append(best.Modifications, gameengine.Modification{
		Power:    2 * count,
		Duration: "until_end_of_turn",
		Timestamp: gs.NextTimestamp(),
	})
	gs.InvalidateCharacteristicsCache()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"target": best.Card.DisplayName(),
		"buff":   2 * count,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"draw_on_combat_damage_followup_unimplemented")
}
