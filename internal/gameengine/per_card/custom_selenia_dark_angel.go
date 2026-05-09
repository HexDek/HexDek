package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSeleniaDarkAngelCustom implements Selenia's life-pay self-bounce.
// The auto-generated activated stub never enforces the 2-life cost or
// performs the bounce.
//
// Oracle text:
//
//	Flying
//	Pay 2 life: Return Selenia to its owner's hand.
//
// "Pay 2 life" is the ENTIRE cost — there's no mana, no tap. The engine
// won't run AST cost dispatch on it, so the handler enforces the cost.
// We refuse to drop the controller below 1 life (§704.5a SBA) when the
// payment is the only effect — Selenia's controller can't lose to her
// own activation.
func registerSeleniaDarkAngelCustom(r *Registry) {
	r.OnActivated("Selenia, Dark Angel", seleniaPayLifeBounce)
}

func seleniaPayLifeBounce(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "selenia_pay_life_bounce"
	if gs == nil || src == nil {
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	if seat.Life <= 2 {
		emitFail(gs, slug, src.Card.DisplayName(), "insufficient_life", map[string]interface{}{
			"life": seat.Life,
		})
		return
	}
	gameengine.LoseLife(gs, src.Controller, 2, src.Card.DisplayName())

	// Bounce Selenia to her owner's hand. Owner == card.Owner; we honor
	// that for stolen Selenias.
	owner := src.Card.Owner
	if owner < 0 || owner >= len(gs.Seats) {
		owner = src.Controller
	}
	gameengine.MoveCard(gs, src.Card, src.Controller, "battlefield", "hand", "selenia_bounce")
	// MoveCard's battlefield arm is a no-op for the source-side removal
	// in older engine builds; sweep manually so the perm doesn't linger.
	if !removePermanent(gs, src) {
		// Already gone — nothing to do.
	}
	// Ensure the card is in OWNER's hand, not controller's, if owner !=
	// controller. MoveCard above moved to seat=src.Controller's hand.
	if owner != src.Controller {
		ctrlSeat := gs.Seats[src.Controller]
		for i, c := range ctrlSeat.Hand {
			if c == src.Card {
				ctrlSeat.Hand = append(ctrlSeat.Hand[:i], ctrlSeat.Hand[i+1:]...)
				gs.Seats[owner].Hand = append(gs.Seats[owner].Hand, c)
				break
			}
		}
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":      src.Controller,
		"life_paid": 2,
		"owner":     owner,
	})
}
