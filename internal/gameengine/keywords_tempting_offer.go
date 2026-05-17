package gameengine

// keywords_tempting_offer.go — Tempting offer (CR §702.74 / Commander
// 2013 + Conspiracy political effect).
//
// Printed pattern:
//
//   "Tempting offer — [base effect]. Each other player may [accept
//    effect]; if they do, copy [base effect] for them."
//
// In rules-engine terms, when the spell or ability with Tempting offer
// resolves, its base effect happens for the controller, then each
// other player in turn order (starting with the player to the
// controller's left) chooses whether to accept. For each opponent who
// accepts, the controller copies the base effect — the copy resolves
// for that opponent (targets default to their own resources).
//
// All copies happen during the original spell's resolution; nothing is
// re-pushed onto the stack. This is the same shape as Replicate
// (§702.99) and Demonic Pact-style "each player may"/"if they do"
// riders. By the time ResolveTemptingOffer returns, every base-effect
// copy has fully resolved.
//
// Wiring:
//   - ResolveTemptingOffer is the single entry point. Per-card code
//     that resolves a Tempting offer spell calls it after resolving
//     the non-Tempting body of the card.
//   - acceptCallback wires the per-seat decision in: the AI/Hat or UI
//     prompts each opponent. Passing nil falls back to a deterministic
//     "no one accepts" policy so callers that haven't wired in the
//     decision layer still get correct (if unexciting) behaviour.
//   - The base-effect resolution uses a synthetic *Permanent whose
//     Controller field is the recipient seat. Default-targeted effects
//     in ResolveEffect fall back to src.Controller (see resolveDraw,
//     resolveGainLife, etc.), so the copy automatically targets the
//     accepting opponent's own resources — their library for draws,
//     their seat for life gain, their battlefield for tap effects with
//     no explicit target.

import (
	"github.com/hexdek/hexdek/internal/gameast"
)

// TemptingOfferAccept is the per-seat opt-in decision callback. It is
// called once per living opponent in APNAP order (from controllerSeat's
// left). Return true to accept the offer for that seat.
type TemptingOfferAccept func(seat int) bool

// ResolveTemptingOffer resolves a Tempting offer rider. CR §702.74.
//
// The base effect resolves once for `controllerSeat` (the "original"
// copy), then each living opponent in turn-order from controllerSeat's
// left is polled via acceptCallback; for every accepter the base
// effect is copied and resolved with that opponent as its source
// controller.
//
// Returns the list of recipient seats in resolution order: index 0 is
// always controllerSeat, followed by accepting opponents in APNAP
// order. The caller can use this to drive logging, replay events, or
// post-resolution bookkeeping (e.g. "Selvala, Explorer Returned"-style
// riders that tally how many accepted).
//
// Semantics:
//   - acceptCallback == nil is treated as "no opponent accepts."
//     Conservative default — the rider effect is real but no political
//     payoff lands unless the caller wires in a decision layer.
//   - Eliminated (Lost) seats are skipped silently — §800.4b prevents
//     effects landing on a left player; APNAP order is computed from
//     LivingOpponents.
//   - The function returns early with [controllerSeat] when the
//     controllerSeat itself is invalid or the controller has lost the
//     game, and with nil when the game state or base effect is nil.
//   - The base effect is resolved IN PLACE; no clone is made. This
//     mirrors how Replicate hands the same Effect pointer to each
//     copy on the stack. ResolveEffect treats Effect nodes as
//     read-only.
func ResolveTemptingOffer(
	gs *GameState,
	controllerSeat int,
	baseEffect gameast.Effect,
	acceptCallback TemptingOfferAccept,
) []int {
	if gs == nil || baseEffect == nil {
		return nil
	}
	if controllerSeat < 0 || controllerSeat >= len(gs.Seats) {
		return nil
	}
	ctrl := gs.Seats[controllerSeat]
	if ctrl == nil || ctrl.Lost {
		return nil
	}

	recipients := []int{controllerSeat}

	// Poll each living opponent in APNAP order from controllerSeat's
	// left. CR §101.4a — "starting with the active player and
	// proceeding in turn order"; for choices anchored on a specific
	// player (here the spell's controller), the convention is "starting
	// with the player to their left" which APNAP from controllerSeat
	// gives directly.
	if acceptCallback != nil {
		for _, oppSeat := range gs.LivingOpponents(controllerSeat) {
			if !acceptCallback(oppSeat) {
				gs.LogEvent(Event{
					Kind: "tempting_offer_decline",
					Seat: oppSeat,
					Details: map[string]interface{}{
						"rule":       "702.74",
						"controller": controllerSeat,
					},
				})
				continue
			}
			recipients = append(recipients, oppSeat)
			gs.LogEvent(Event{
				Kind: "tempting_offer_accept",
				Seat: oppSeat,
				Details: map[string]interface{}{
					"rule":       "702.74",
					"controller": controllerSeat,
				},
			})
		}
	}

	gs.LogEvent(Event{
		Kind:   "tempting_offer_resolve",
		Seat:   controllerSeat,
		Amount: len(recipients),
		Details: map[string]interface{}{
			"rule":       "702.74",
			"recipients": append([]int(nil), recipients...),
		},
	})

	// Resolve the base effect once per recipient. The original
	// (controllerSeat) resolves first per the printed wording
	// ("[base effect]. Each other player may [accept]; if they do,
	// copy [base effect] for them.") — the controller's copy is the
	// printed sentence, opponent copies are the rider clause.
	for _, seat := range recipients {
		src := syntheticTemptingOfferSource(seat)
		ResolveEffect(gs, src, baseEffect)
	}

	return recipients
}

// syntheticTemptingOfferSource builds the minimal *Permanent used to
// drive a single copy of a Tempting offer base effect. Controller is
// the recipient seat so default-target effects in ResolveEffect route
// to the recipient's own resources; Owner mirrors Controller to keep
// owner-conditioned logic (e.g. "you may put it on top of YOUR
// library") behaving as the copy implies.
//
// Card is a thin placeholder — it carries the recipient's name for
// log readability and an empty AST so any code path that walks the
// source's abilities short-circuits cleanly. The placeholder is not
// added to any battlefield; it exists only for the duration of the
// ResolveEffect call and is garbage-collected after.
func syntheticTemptingOfferSource(seat int) *Permanent {
	return &Permanent{
		Card: &Card{
			Name:  "Tempting Offer copy",
			Owner: seat,
			AST:   &gameast.CardAST{Name: "Tempting Offer copy"},
		},
		Controller: seat,
		Owner:      seat,
	}
}
