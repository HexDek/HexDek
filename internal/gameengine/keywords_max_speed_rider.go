package gameengine

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameast"
)

// keywords_max_speed_rider.go — resolver-side wiring for the §702.178
// "max speed —" rider keyword. Builds on the PlayerSpeed counter system
// (keywords_speed_counter.go, commit 14800d0) which provides the
// per-seat speed (0..MaxSpeedCap) and the MaxSpeedActive predicate.
//
// What a max-speed rider is:
//
//   Some Aetherdrift cards print a rider line like
//     "Max speed — [rider effect]"
//   meaning: when this card resolves, if its controller is at MaxSpeed
//   (Seat.Speed == 4), ALSO apply the rider effect. The rider is part
//   of the spell's one-shot resolution, not a separate triggered
//   ability — it shares the spell's source, controller, and targets.
//
// Surface:
//
//   - HasMaxSpeedRider(card)       — detector. Mirrors HasStrive's two-
//                                    path scheme: AST keyword OR oracle
//                                    text "max speed —" / "max speed -".
//   - ApplyMaxSpeedRider(gs, src)  — executor. Called once per spell
//                                    resolution from resolveSequence
//                                    (guarded by depth counter so
//                                    nested sequences don't re-fire).
//                                    No-op when controller isn't at
//                                    MaxSpeed or the card has no
//                                    rider. When the AST carries a
//                                    tagged "max_speed_rider" ability
//                                    we resolve its Effect; otherwise
//                                    we log a structured
//                                    max_speed_rider_pending event so
//                                    per_card handlers / corpus
//                                    backfill can react.
//   - SpeedDamageReporter(gs, dealerSeat) — thin wrapper around
//                                    AdvanceSpeed kept at the damage
//                                    call sites so observability and
//                                    once-per-turn semantics live in
//                                    one place. Returns whether speed
//                                    actually changed.

// HasMaxSpeedRider reports whether the card has a "max speed —" rider.
// Detection paths (mirrors HasStrive):
//
//  1. cardHasKeywordByName(card, "max speed") — AST keyword tag.
//  2. Oracle text contains "max speed —" or "max speed -" (Scryfall
//     uses em-dash; some older corpus dumps use ASCII hyphen).
//
// Returns false for nil cards / cards without an AST and an oracle text.
// The detector is independent of game state — it answers "does this
// card print a rider", not "is the rider currently active".
func HasMaxSpeedRider(card *Card) bool {
	if card == nil {
		return false
	}
	if cardHasKeywordByName(card, "max speed") {
		return true
	}
	text := OracleTextLower(card)
	if text == "" {
		return false
	}
	return strings.Contains(text, "max speed —") || strings.Contains(text, "max speed -")
}

// findMaxSpeedRiderEffect returns the Effect of the first Ability on
// `card.AST` whose Keyword Name equals "max_speed_rider" (the corpus-
// tagging convention for the rider's payload). Returns nil if the AST
// doesn't carry a tagged rider effect.
//
// We deliberately accept BOTH a Keyword named "max_speed_rider" with
// the payload effect attached to a sibling Activated ability AND a
// direct Activated ability tagged via its Effect. The first match wins
// — corpus producers can choose either shape.
func findMaxSpeedRiderEffect(card *Card) gameast.Effect {
	if card == nil || card.AST == nil {
		return nil
	}
	for _, ab := range card.AST.Abilities {
		switch v := ab.(type) {
		case *gameast.Activated:
			if v == nil {
				continue
			}
			// An Activated ability whose Effect's first item is tagged
			// as the rider payload counts.
			if v.Effect != nil && abilityIsTaggedMaxSpeedRider(v) {
				return v.Effect
			}
		}
	}
	return nil
}

// abilityIsTaggedMaxSpeedRider returns true when an Activated ability
// carries the rider-payload marker. Two recognized shapes:
//
//   1. Cost.Extra contains the literal "max_speed_rider" — the
//      corpus-parser convention for tagging a no-op activation cost
//      that the resolver should never see paid (the ability fires at
//      spell-resolve time, not on activation).
//   2. The ability's Raw text begins with "max speed" (case-
//      insensitive) — a fallback for cards whose corpus dump
//      preserved the printed rider line as an Activated.Raw without
//      a Cost.Extra tag.
//
// Tolerant by design so future tagging schemes extend without
// breaking callers.
func abilityIsTaggedMaxSpeedRider(a *gameast.Activated) bool {
	if a == nil {
		return false
	}
	for _, extra := range a.Cost.Extra {
		if strings.EqualFold(extra, "max_speed_rider") {
			return true
		}
	}
	raw := strings.ToLower(strings.TrimSpace(a.Raw))
	return strings.HasPrefix(raw, "max speed")
}

// ApplyMaxSpeedRider executes the max-speed rider for `src`, if any.
// Returns true if the rider actually fired.
//
// Conditions to fire:
//
//   - src is non-nil and has a Card
//   - HasMaxSpeedRider(src.Card)
//   - MaxSpeedActive(gs, src.Controller) — i.e. controller is at
//     MaxSpeedCap (=4)
//
// On fire we log a max_speed_rider event and, if the AST carries a
// tagged rider effect, ResolveEffect runs it with `src` as the source.
// When no tagged effect is present (most corpus today), we log a
// max_speed_rider_pending event with the source and rule reference so
// per_card handlers + corpus-backfill jobs can find the affected
// spells.
func ApplyMaxSpeedRider(gs *GameState, src *Permanent) bool {
	if gs == nil || src == nil || src.Card == nil {
		return false
	}
	if !HasMaxSpeedRider(src.Card) {
		return false
	}
	if !MaxSpeedActive(gs, src.Controller) {
		return false
	}

	gs.LogEvent(Event{
		Kind:   "max_speed_rider",
		Seat:   src.Controller,
		Source: src.Card.DisplayName(),
		Details: map[string]interface{}{
			"rule":  "702.178",
			"speed": SpeedOf(gs, src.Controller),
		},
	})

	if eff := findMaxSpeedRiderEffect(src.Card); eff != nil {
		ResolveEffect(gs, src, eff)
		return true
	}

	gs.LogEvent(Event{
		Kind:   "max_speed_rider_pending",
		Seat:   src.Controller,
		Source: src.Card.DisplayName(),
		Details: map[string]interface{}{
			"rule":   "702.178",
			"reason": "rider_payload_not_in_ast",
		},
	})
	return true
}

// SpeedDamageReporter is the canonical entry point for damage-step
// hooks into the speed counter. Wraps AdvanceSpeed so all call sites
// (combat damage, noncombat damage, per_card "ping the opponent"
// effects) go through the same once-per-turn gate and emit
// consistently-shaped speed_advance events.
//
// Returns true iff the dealer's speed actually changed. dealerSeat
// values out of range or < 0 are safely ignored.
//
// Round-25 contract: callers do not need to track once-per-turn state
// themselves — pass every qualifying damage event through here and the
// gate enforces the §702.179 limit.
func SpeedDamageReporter(gs *GameState, dealerSeat int) bool {
	return AdvanceSpeed(gs, dealerSeat)
}
