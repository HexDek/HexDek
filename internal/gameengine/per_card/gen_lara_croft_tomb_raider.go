package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerLaraCroftTombRaider wires Lara Croft, Tomb Raider.
//
// Oracle text:
//
//   First strike, reach
//   Whenever Lara Croft attacks, exile up to one target legendary
//   artifact card or legendary land card from a graveyard and put a
//   discovery counter on it. You may play a card from exile with a
//   discovery counter on it this turn.
//   Raid — At end of combat on your turn, if you attacked this turn,
//          create a Treasure token.
//
// R37 port:
//
//   - First strike + reach: AST keyword pipeline.
//   - Attack trigger (exile + discovery counter + play-from-exile
//     permission): NOT ported — would need a discovery-counter +
//     play-from-exile system that doesn't currently exist. Flagged in
//     emitPartial.
//   - Raid: PORTED. end_of_combat_controller trigger fires
//     CreateTreasureToken when the controller attacked this turn.
//     Reads Seat.Turn.Attacked (set by combat.go's DeclareAttackers).
//     The "on your turn" half is enforced by registering against
//     end_of_combat_controller (event_aliases routes the controller
//     variant when gs.Active == perm.Controller).
func registerLaraCroftTombRaider(r *Registry) {
	r.OnETB("Lara Croft, Tomb Raider", laraCroftStaticETB)
	r.OnTrigger("Lara Croft, Tomb Raider", "end_of_combat", laraCroftRaidTreasure)
}

func laraCroftStaticETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "lara_croft_static"
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"attack-trigger discovery-counter + play-from-exile not modeled; Raid Treasure-token handled by end_of_combat hook")
}

// laraCroftRaidTreasure fires at end of combat. Per CR §702.128a
// (Raid), "attacked this turn" is read from Seat.Turn.Attacked. Gated
// on (a) it being our controller's turn — checked via gs.Active —
// and (b) the Seat.Turn.Attacked flag.
func laraCroftRaidTreasure(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "lara_croft_raid_treasure"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) || gs.Seats[seat] == nil {
		return
	}
	// "On your turn" — Raid only fires on the controller's own turn.
	if gs.Active != seat {
		return
	}
	if !gs.Seats[seat].Turn.Attacked {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   seat,
			"reason": "raid_not_satisfied",
		})
		return
	}
	gameengine.CreateTreasureToken(gs, seat)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     seat,
		"treasure": true,
	})
}
