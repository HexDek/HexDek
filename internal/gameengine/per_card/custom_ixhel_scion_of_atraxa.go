package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerIxhelScionOfAtraxa wires Ixhel, Scion of Atraxa.
//
// Oracle text:
//
//	Flying, vigilance, deathtouch, lifelink, toxic 1
//	At the beginning of your end step, exile the top card of each
//	opponent's library who has three or more poison counters. Until
//	end of turn, you may play those cards, and you may spend mana as
//	though it were mana of any color to cast those spells.
//
// We listen on `end_step` for the controlling seat. For each opponent
// with poison counters >= 3 (read from Seat.Flags["poison_counters"]),
// exile the top library card and stash a "playable until EOT"
// breadcrumb. Actual "play from exile" support is a partial — the
// engine doesn't yet have a "may play exiled cards" hook.
//
// Toxic 1 + the four keyword grants are AST-handled (keywords).
func registerIxhelScionOfAtraxa(r *Registry) {
	r.OnTrigger("Ixhel, Scion of Atraxa", "end_step", ixhelEndStepExile)
}

func ixhelEndStepExile(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ixhel_eos_exile_poisoned_opps"
	if gs == nil || perm == nil {
		return
	}
	// Only on Ixhel's controller's end step.
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		if gs.Active != perm.Controller {
			return
		}
	}
	exiled := 0
	for _, oppSeat := range gs.Opponents(perm.Controller) {
		if oppSeat < 0 || oppSeat >= len(gs.Seats) {
			continue
		}
		opp := gs.Seats[oppSeat]
		if opp == nil || opp.Lost {
			continue
		}
		poison := 0
		if opp.Flags != nil {
			poison = opp.Flags["poison_counters"]
		}
		if poison < 3 {
			continue
		}
		if len(opp.Library) == 0 {
			continue
		}
		top := opp.Library[0]
		gameengine.MoveCard(gs, top, oppSeat, "library", "exile", "ixhel_eos_exile")
		// Tag exiled card so a (future) play-from-exile hook can spot it.
		if top != nil {
			top.Types = append(top.Types, "ixhel_playable_until_eot")
		}
		exiled++
	}
	if exiled == 0 {
		return
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":            perm.Controller,
		"cards_exiled":    exiled,
	})
	emitPartial(gs, "ixhel_play_exiled_cards", perm.Card.DisplayName(),
		"\"may play those cards until end of turn + spend mana of any color\" requires play-from-exile hook; tag set")
}
