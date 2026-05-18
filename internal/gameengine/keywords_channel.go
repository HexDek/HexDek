package gameengine

// keywords_channel.go — Channel (CR §702.74, Kamigawa: Neon Dynasty 2022)
// as a real hand-only activated ability with discard-as-cost.
//
// CR §702.74a: "Channel — [cost], Discard this card: [effect]" is an
//                activated ability that can be activated only from the
//                player's hand. It uses the stack but is not a spell.
// CR §702.74b: The discard of the card is part of paying the activation
//                cost, not part of the effect. The card hits the
//                graveyard before the effect resolves; cards that count
//                graveyards (delirium, threshold) see the discard
//                before the effect.
// CR §702.74c: Different cards print different Channel effects. The
//                effect itself is per-card and dispatched by the
//                per-card handler registry; this file plumbs the
//                cost-payment + zone-move scaffolding.
//
// Engine surface:
//
//   - HasChannel(card) bool
//       Detects the Channel keyword on the card's AST.
//
//   - ActivateChannel(gs, seat, cardInHand, channelCost) error
//       Verifies the card is in `seat`'s hand and carries the Channel
//       keyword, that the seat can pay `channelCost` mana, then:
//         1. Pays the mana from seat.ManaPool
//         2. Discards the card via DiscardCard (CR §702.74b — discard
//            is part of the cost)
//         3. Pushes a StackItem onto the stack tagged
//            CostMeta["channel_activate"] = true and CastZone "hand"
//            so resolve dispatch + per-card handlers can recognize
//            the activation.
//
// The Channel EFFECT itself is per-card and runs at stack-resolution
// time via the per-card handler registry (e.g. Boseiju's
// destroy-target-permanent). This helper only handles the activation
// scaffolding; the resolution-time effect dispatch happens through the
// same trigger channel any spell uses.

// ---------------------------------------------------------------------------
// HasChannel
// ---------------------------------------------------------------------------

// HasChannel returns true if the card has the channel keyword.
// CR §702.74a — the parser emits "Channel —" preambles as a Keyword
// ability with name "channel" on the card's AST.
func HasChannel(card *Card) bool {
	return cardHasKeywordByName(card, "channel")
}

// ---------------------------------------------------------------------------
// ActivateChannel
// ---------------------------------------------------------------------------

// ActivateChannel runs the §702.74a hand-only Channel activation for a
// specific card in `seatIdx`'s hand.
//
// Validation (atomic — no mutation happens on failure):
//
//   - card must be non-nil and carry the Channel keyword
//   - card must actually be in seat.Hand (CR §702.74a hand-only)
//   - seat must be able to afford channelCost mana
//   - sorcery-speed timing is NOT enforced here — Channel abilities
//     have varied timing windows (Otawara is instant-speed, Boseiju is
//     instant-speed, Cleansing Bolt-style channels could be sorcery-
//     speed). Per-card handlers wrap ActivateChannel and gate timing
//     themselves.
//
// On success:
//
//   - seat.ManaPool decreases by channelCost (logged as a pay_mana event
//     with reason="channel_cost")
//   - The card is moved hand → graveyard via DiscardCard (CR §702.74b
//     — discard is part of the cost; fires the canonical card_discarded
//     trigger so Liliana's Caress, Waste Not, etc. see it)
//   - A StackItem is pushed with CostMeta["channel_activate"] = true,
//     CastZone = "hand", and CostMeta["channel_cost"] = channelCost so
//     resolve.go's ResolveStackTop can dispatch to the appropriate
//     per-card handler and resolution-time observers can filter.
//   - A "channel_activate" log event is emitted.
//   - FireCardTrigger("channel_activated", ctx) fires so per-card
//     handlers receive a structured hook with the card name + cost.
//
// Returns nil on success, a *CastError describing the failure
// otherwise.
func ActivateChannel(gs *GameState, seatIdx int, card *Card, channelCost int) error {
	if gs == nil {
		return &CastError{Reason: "nil_game"}
	}
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return &CastError{Reason: "invalid_seat"}
	}
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return &CastError{Reason: "nil_seat"}
	}
	if card == nil {
		return &CastError{Reason: "nil_card"}
	}
	if !HasChannel(card) {
		return &CastError{Reason: "no_channel_keyword"}
	}
	if channelCost < 0 {
		return &CastError{Reason: "invalid_channel_cost"}
	}
	// CR §702.74a — Channel activates from HAND. Reject if the card
	// isn't actually in this seat's hand (graveyard / battlefield /
	// stack / exile / library all fail).
	inHand := false
	for _, c := range seat.Hand {
		if c == card {
			inHand = true
			break
		}
	}
	if !inHand {
		return &CastError{Reason: "card_not_in_hand"}
	}
	if seat.ManaPool < channelCost {
		return &CastError{Reason: "insufficient_mana_for_channel"}
	}
	// Drannith Magistrate restricts cast-from-non-hand zones — Channel
	// activates from hand, so it's NOT a Drannith violation. Skip the
	// drannithRestrictsZoneCast gate here.

	name := card.DisplayName()

	// 1. Pay mana cost.
	seat.ManaPool -= channelCost
	SyncManaAfterSpend(seat)
	if channelCost > 0 {
		gs.LogEvent(Event{
			Kind:   "pay_mana",
			Seat:   seatIdx,
			Amount: channelCost,
			Source: name,
			Details: map[string]interface{}{
				"reason": "channel_cost",
				"rule":   "702.74a",
			},
		})
	}

	// 2. Discard the card. CR §702.74b — part of the cost, not the
	// effect. DiscardCard fires the canonical card_discarded trigger
	// so Madness, Mayhem, and similar discard-driven mechanics see it.
	DiscardCard(gs, card, seatIdx)

	// 3. Push the activation onto the stack. The resolution-time effect
	// is per-card; the per-card handler reads CostMeta["channel_activate"]
	// when dispatching (or hooks the channel_activated FireCardTrigger
	// fan-out directly).
	item := &StackItem{
		Card:       card,
		Controller: seatIdx,
		Kind:       "activated",
		CastZone:   ZoneHand,
		CostMeta: map[string]interface{}{
			"channel_activate": true,
			"channel_cost":     channelCost,
		},
	}
	PushStackItem(gs, item)

	gs.LogEvent(Event{
		Kind:   "channel_activate",
		Seat:   seatIdx,
		Source: name,
		Amount: channelCost,
		Details: map[string]interface{}{
			"rule": "702.74a",
		},
	})

	FireCardTrigger(gs, "channel_activated", map[string]interface{}{
		"card":            card,
		"card_name":       name,
		"controller_seat": seatIdx,
		"channel_cost":    channelCost,
	})

	return nil
}
