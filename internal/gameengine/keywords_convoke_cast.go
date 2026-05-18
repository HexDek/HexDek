package gameengine

// keywords_convoke_cast.go — Convoke cast helper (CR §702.51,
// Ravnica 2005 / Magic 2014 reprint).
//
// CR §702.51a: Convoke is a static ability that functions while the
//              spell is on the stack. "Convoke" means "Your creatures
//              can help cast this spell. Each creature you tap while
//              casting this spell pays for {1} or one mana of that
//              creature's color."
// CR §702.51b: For each creature tapped this way, the controller may
//              choose whether the tap pays for {1} generic mana OR
//              for one mana of that creature's color (the creature
//              must have that color for the colored option to apply).
// CR §702.51c: Summoning sickness does NOT prevent a creature from
//              being tapped for convoke. The "summoning sick" rule
//              (CR §302.1) only restricts attacking and activated
//              abilities with {T}/{Q}; convoke isn't an activated
//              ability, it's an additional-cost mechanic resolved as
//              the spell is cast.
//
// Existing surface (untouched by this file):
//
//   - HasConvoke(card)               — keywords_p0.go
//   - ConvokeCostReduction(gs, seat) — keywords_p0.go (advisory upper-
//                                       bound count for the legacy
//                                       cost_modifiers.go convoke path)
//
// This file adds the explicit cast helper the round-34 task wants:
// CastWithConvoke takes the exact tapped-creature slice the caster
// declared, validates each, taps them, applies the cost reduction,
// pays the remaining mana, and pushes the StackItem with CostMeta
// stamps so downstream resolvers / observers can audit the convoke
// chain.

import (
	"strings"
)

// CastWithConvoke casts `card` from `seatIdx`'s hand using convoke
// (§702.51). The caller declares `tappedCreatures` — the exact
// creatures they want to tap as additional cost; the helper validates
// each, taps them, and applies the cost reduction.
//
// Per-creature validation (each must pass all of):
//
//   - non-nil
//   - is a creature on the battlefield
//   - controlled by `seatIdx`
//   - not currently tapped
//
// Summoning sickness is explicitly allowed (CR §702.51c — convoke is
// not an activated ability and is not gated by §302.1).
//
// Cost computation:
//
//   - Each tapped creature contributes ONE mana point of reduction
//     (either {1} or one of its colors per §702.51b). The caller is
//     not asked to pre-decide which colors satisfy which pips; we
//     simply record each creature's colors in CostMeta so a
//     downstream pip-checker can validate colored requirements were
//     met.
//   - The reduction is capped at the card's CMC (you can't get paid
//     to cast a spell).
//   - The remaining mana (CMC − reduction) must be payable from the
//     seat's existing pool. Insufficient mana = rejection AFTER the
//     creature validation pass (no tapping happens on failure).
//
// On success:
//
//   - Each declared creature is set Tapped=true.
//   - card is removed from hand.
//   - net mana cost is debited from seat.ManaPool.
//   - StackItem is pushed with CostMeta:
//       "alt_cost"               = "convoke"
//       "convoke_creatures_used" = len(tappedCreatures)
//       "convoke_reduction"      = reduction (≤ card.CMC)
//       "convoke_net_cost"       = net (mana actually paid)
//       "convoke_colors"         = []string — one entry per tapped
//                                  creature, joining the creature's
//                                  Colors slice (e.g. "G", "UW",
//                                  "" for colorless); resolver-side
//                                  pip-checkers can iterate.
//
// Returns a CostError with a structured Reason on failure. None of
// the game state is mutated on failure — the validation pass runs
// before any tapping or mana spend.
func CastWithConvoke(gs *GameState, seatIdx int, card *Card, tappedCreatures []*Permanent) (*CostPaymentResult, error) {
	if gs == nil {
		return nil, &CastError{Reason: "nil_game"}
	}
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return nil, &CastError{Reason: "invalid_seat"}
	}
	if card == nil {
		return nil, &CastError{Reason: "nil_card"}
	}
	if !HasConvoke(card) {
		return nil, &CastError{Reason: "no_convoke_keyword"}
	}
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return nil, &CastError{Reason: "nil_seat"}
	}

	// Validate every declared creature BEFORE mutating any state.
	// This is the load-bearing invariant: a single bad creature in
	// the slice fails the whole cast cleanly without partial taps.
	creaturesOnBF := make(map[*Permanent]bool, len(seat.Battlefield))
	for _, p := range seat.Battlefield {
		if p != nil {
			creaturesOnBF[p] = true
		}
	}
	for i, c := range tappedCreatures {
		if c == nil {
			return nil, &CastError{Reason: "nil_convoke_creature"}
		}
		if !c.IsCreature() {
			return nil, &CastError{Reason: "convoke_not_creature"}
		}
		if c.Controller != seatIdx {
			return nil, &CastError{Reason: "convoke_creature_not_controlled"}
		}
		if !creaturesOnBF[c] {
			return nil, &CastError{Reason: "convoke_creature_not_on_battlefield"}
		}
		if c.Tapped {
			return nil, &CastError{Reason: "convoke_creature_already_tapped"}
		}
		// Defensive: detect duplicate creature in the slice. Tapping
		// the same creature twice would pay for {2} from one tap,
		// which §702.51a forbids.
		for j := 0; j < i; j++ {
			if tappedCreatures[j] == c {
				return nil, &CastError{Reason: "convoke_creature_duplicated"}
			}
		}
	}

	// Cost arithmetic.
	normalCost := card.CMC
	if normalCost < 0 {
		normalCost = 0
	}
	reduction := len(tappedCreatures)
	if reduction > normalCost {
		reduction = normalCost
	}
	net := normalCost - reduction
	if seat.ManaPool < net {
		return nil, &CastError{Reason: "insufficient_mana"}
	}

	// Commit: tap creatures, remove card, pay mana, push stack item.
	for _, c := range tappedCreatures {
		c.Tapped = true
		gs.LogEvent(Event{
			Kind:   "tap",
			Seat:   seatIdx,
			Source: c.Card.DisplayName(),
			Details: map[string]interface{}{
				"reason":  "convoke",
				"rule":    "702.51",
				"casting": card.DisplayName(),
			},
		})
	}
	if !removeFromZone(seat, card, ZoneHand) {
		// Restore taps if the card wasn't actually in hand — preserve
		// the no-partial-mutation guarantee on failure.
		for _, c := range tappedCreatures {
			c.Tapped = false
		}
		return nil, &CastError{Reason: "not_in_hand"}
	}
	if net > 0 {
		seat.ManaPool -= net
		SyncManaAfterSpend(seat)
		gs.LogEvent(Event{
			Kind:   "pay_mana",
			Seat:   seatIdx,
			Amount: net,
			Source: card.DisplayName(),
			Details: map[string]interface{}{
				"reason":     "convoke_cast",
				"keyword":    "convoke",
				"rule":       "702.51",
				"normal_cost": normalCost,
				"reduction":  reduction,
			},
		})
	}

	convokeColors := make([]string, 0, len(tappedCreatures))
	for _, c := range tappedCreatures {
		convokeColors = append(convokeColors, joinColors(c.Card.Colors))
	}

	item := &StackItem{
		Card:       card,
		Controller: seatIdx,
		CastZone:   ZoneHand,
		Effect:     collectSpellEffect(card),
		CostMeta: map[string]interface{}{
			"alt_cost":               "convoke",
			"convoke_creatures_used": len(tappedCreatures),
			"convoke_reduction":      reduction,
			"convoke_net_cost":       net,
			"convoke_colors":         convokeColors,
		},
	}
	PushStackItem(gs, item)

	gs.LogEvent(Event{
		Kind:   "convoke_cast",
		Seat:   seatIdx,
		Source: card.DisplayName(),
		Amount: reduction,
		Details: map[string]interface{}{
			"rule":          "702.51",
			"creatures":     len(tappedCreatures),
			"normal_cost":   normalCost,
			"reduction":     reduction,
			"net_paid":      net,
			"colors":        convokeColors,
		},
	})
	return &CostPaymentResult{}, nil
}

// joinColors returns a compact string for the creature's color
// identity ("G", "UW", "" for colorless) for stamping into CostMeta.
// Order is preserved from the source slice — callers that need
// canonical ordering should sort up-stream.
func joinColors(cs []string) string {
	if len(cs) == 0 {
		return ""
	}
	cleaned := make([]string, 0, len(cs))
	for _, c := range cs {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		cleaned = append(cleaned, strings.ToUpper(c))
	}
	return strings.Join(cleaned, "")
}
