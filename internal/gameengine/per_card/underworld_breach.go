package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerUnderworldBreach wires up Underworld Breach.
//
// Oracle text:
//
//	Each nonland card in your graveyard has escape. The escape cost
//	is equal to the card's mana cost plus exile three other cards
//	from your graveyard.
//	At the beginning of the end step, sacrifice Underworld Breach.
//
// Breach is THE combo enabler for storm lines. Combined with Lion's Eye
// Diamond + Brain Freeze, it loops: LED for mana → Brain Freeze from GY
// via escape (exile 3) → mill yourself for more fuel → LED from GY →
// repeat until opponents are milled out.
//
// Implementation:
//   - OnETB: grant escape to every nonland card in the controller's
//     graveyard using RegisterZoneCastGrant + NewBreachEscapePermission.
//     This integrates with the engine's zone_cast.go CastFromZone
//     primitive so the Hat/AI can actually cast spells from the
//     graveyard. Also register a delayed end-step trigger to sacrifice
//     Breach.
//   - Trigger on graveyard changes: when new cards enter the graveyard
//     (mill, discard, etc.), grant them escape too. This is critical for
//     the combo loop where Brain Freeze mills cards that then become
//     escape-castable.
func registerUnderworldBreach(r *Registry) {
	r.OnETB("Underworld Breach", underworldBreachETB)
	// Refresh escape grants when cards enter the controller's graveyard.
	r.OnTrigger("Underworld Breach", "zone_change", underworldBreachRefresh)
	r.OnTrigger("Underworld Breach", "creature_dies", underworldBreachRefresh)
}

func underworldBreachETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "underworld_breach"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}

	// Set the global flag.
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["escape_grants_to_graveyard"] = 1
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	gs.Flags["breach_active_seat_"+intToStr(seat)] = perm.Timestamp

	// Grant escape to every nonland card in the controller's graveyard.
	s := gs.Seats[seat]
	granted := 0
	for _, c := range s.Graveyard {
		if c == nil {
			continue
		}
		if cardHasType(c, "land") {
			continue
		}
		cmc := cardCMC(c)
		escapePerm := gameengine.NewBreachEscapePermission(cmc)
		escapePerm.RequireController = seat
		gameengine.RegisterZoneCastGrant(gs, c, escapePerm)
		granted++
	}

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":            seat,
		"escape_granted":  granted,
		"zone_cast":       "graveyard_escape_with_exile_3",
		"integrated":      true,
	})

	// Register the end-step sacrifice trigger. Delayed triggers fire at
	// phase/step boundaries; "end_of_turn" is the canonical key.
	gs.RegisterDelayedTrigger(&gameengine.DelayedTrigger{
		TriggerAt:      "end_of_turn",
		ControllerSeat: seat,
		SourceCardName: perm.Card.DisplayName(),
		EffectFn: func(gs *gameengine.GameState) {
			// Sacrifice Breach via SacrificePermanent for proper zone-change
			// handling: replacement effects, dies/LTB triggers, commander redirect.
			gameengine.SacrificePermanent(gs, perm, "underworld_breach_end_step")
			// Remove the escape-grant flag.
			delete(gs.Flags, "breach_active_seat_"+intToStr(seat))
			// Revoke all escape grants from the controller's graveyard.
			seatData := gs.Seats[seat]
			if seatData != nil {
				for _, c := range seatData.Graveyard {
					if c != nil {
						gameengine.RemoveZoneCastGrant(gs, c)
					}
				}
			}
		},
	})
}

// underworldBreachRefresh fires when cards enter the graveyard. Any new
// nonland card that arrives while Breach is active gets an escape grant.
func underworldBreachRefresh(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	// Confirm Breach is still active.
	if gs.Flags == nil || gs.Flags["breach_active_seat_"+intToStr(seat)] == 0 {
		return
	}
	// Grant escape to any nonland card in the graveyard that doesn't
	// already have a zone-cast grant.
	s := gs.Seats[seat]
	for _, c := range s.Graveyard {
		if c == nil {
			continue
		}
		if cardHasType(c, "land") {
			continue
		}
		// Check if already granted.
		if gameengine.GetZoneCastGrant(gs, c) != nil {
			continue
		}
		cmc := cardCMC(c)
		escapePerm := gameengine.NewBreachEscapePermission(cmc)
		escapePerm.RequireController = seat
		gameengine.RegisterZoneCastGrant(gs, c, escapePerm)
	}
}
