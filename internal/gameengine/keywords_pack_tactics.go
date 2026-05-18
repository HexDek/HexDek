package gameengine

// keywords_pack_tactics.go — Pack Tactics (CR §702.149, Strixhaven
// Lorehold 2021) as a generic combat-declared trigger.
//
// CR §702.149a: Pack tactics is a triggered ability that functions
//               while the card with pack tactics is on the
//               battlefield. "Pack tactics — [effect]" means
//               "Whenever this creature attacks, if you're attacking
//               with creatures with total power 6 or more, [effect]."
//
// Sibling of Battalion (CR §702.101) but uses a TOTAL-POWER threshold
// (>= 6) instead of a creature-count threshold (>= 3). Implementation
// mirrors keywords_battalion.go's hook shape so the per_card payoff
// surface is consistent: a single fan-out fires
// FireCardTrigger("pack_tactics_triggered", ctx) per qualifying
// source. Per-card handlers implement the actual payoff (typically
// "draw a card" or "creatures you control get +X/+X UEOT").
//
// Wired from declareAttackers in combat.go alongside
// FireBattalionTriggers / FireDethroneTriggers, so attack-time buffs
// (Glorious Anthem and friends) that resolve before the
// declare-attackers trigger window are visible to the power sum via
// each Permanent.Power() lookup.

import (
	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// PackTacticsPowerThreshold
// ---------------------------------------------------------------------------

// PackTacticsPowerThreshold is the printed §702.149 cutoff. A separate
// const so tests + per_card handlers can refer to it symbolically and
// the engine's threshold semantics stay in one place.
const PackTacticsPowerThreshold = 6

// ---------------------------------------------------------------------------
// HasPackTactics
// ---------------------------------------------------------------------------

// HasPackTactics returns true if the card has the pack tactics keyword
// in its AST. Mirrors HasBattalion.
func HasPackTactics(card *Card) bool {
	if card == nil || card.AST == nil {
		return false
	}
	for _, ab := range card.AST.Abilities {
		if kw, ok := ab.(*gameast.Keyword); ok && keywordNameEquals(kw, "pack tactics") {
			return true
		}
	}
	return false
}

// PermanentHasPackTactics is the Permanent-level convenience for the
// declare-attackers hook. Like PermanentHasBattalion it honors runtime
// grants (GrantedAbilities, kw:pack_tactics flag) in addition to AST
// keywords, so future grant-style effects can hand a creature pack
// tactics without round-tripping through corpus.
func PermanentHasPackTactics(p *Permanent) bool {
	if p == nil {
		return false
	}
	if HasPackTactics(p.Card) {
		return true
	}
	for _, g := range p.GrantedAbilities {
		if equalFoldTrimmed(g, "pack tactics") {
			return true
		}
	}
	if p.Flags != nil && p.Flags["kw:pack_tactics"] > 0 {
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// FirePackTacticsTriggers — per-source check
// ---------------------------------------------------------------------------

// FirePackTacticsTriggers evaluates the §702.149a trigger for a single
// attacking source. Sums the current Power() of every attacking
// creature controlled by `attacker.Controller` (so Glorious Anthem and
// other in-play buffs that have already settled count) and, if the
// total is >= PackTacticsPowerThreshold AND the source itself is
// attacking, fires the trigger.
//
// No-op when:
//
//   - gs/attacker/attacker.Card is nil
//   - attacker doesn't have pack tactics
//   - attacker isn't currently an attacking creature
//   - total attacking power for the controller is below the threshold
//
// The trigger context exposes:
//
//   "source"        *Permanent  — the pack-tactics-bearing attacker
//   "controller"    int         — attacker.Controller
//   "defender_seat" int         — the seat the source is attacking
//                                 (forwarded from the caller; -1 when
//                                 unset)
//   "total_power"   int         — sum of attacking-creature powers
//                                 for that controller
//   "attackers"     []*Permanent — the controller's full attacker list
func FirePackTacticsTriggers(gs *GameState, attacker *Permanent, defenderSeat int) bool {
	if gs == nil || attacker == nil || attacker.Card == nil {
		return false
	}
	if !PermanentHasPackTactics(attacker) {
		return false
	}
	// CR §702.149a — "Whenever this creature attacks ..." — source must
	// itself be in the current attack.
	if !attacker.IsAttacking() {
		return false
	}
	seatIdx := attacker.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return false
	}
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return false
	}

	// Sum power of every attacking creature this controller controls.
	// Opponent attackers in extra-combat / control-swap scenarios are
	// excluded by the Controller filter.
	totalPower := 0
	var packAttackers []*Permanent
	for _, p := range seat.Battlefield {
		if p == nil || !p.IsCreature() {
			continue
		}
		if !p.IsAttacking() {
			continue
		}
		totalPower += p.Power()
		packAttackers = append(packAttackers, p)
	}
	if totalPower < PackTacticsPowerThreshold {
		return false
	}

	gs.LogEvent(Event{
		Kind:   "pack_tactics_trigger",
		Seat:   seatIdx,
		Source: attacker.Card.DisplayName(),
		Amount: totalPower,
		Details: map[string]interface{}{
			"total_power":   totalPower,
			"defender_seat": defenderSeat,
			"rule":          "702.149a",
		},
	})
	FireCardTrigger(gs, "pack_tactics_triggered", map[string]interface{}{
		"source":        attacker,
		"controller":    seatIdx,
		"defender_seat": defenderSeat,
		"total_power":   totalPower,
		"attackers":     packAttackers,
	})
	return true
}

// ---------------------------------------------------------------------------
// FirePackTacticsForAttackers — batch declare-attackers hook
// ---------------------------------------------------------------------------
//
// Batched entry point used by combat.go's declareAttackers, mirroring
// FireBattalionTriggers' shape. Iterates every attacker, looks up
// each one's chosen defender (AttackerDefender), and delegates to
// FirePackTacticsTriggers. Kept separate from the per-source helper
// so call sites that want the §702.149a check for a single newly-
// scooped-in attacker (e.g. "enters tapped and attacking" tokens)
// can reach for the per-source variant directly.

// FirePackTacticsForAttackers fans the per-source check across an
// attacker slice. Safe to pass an empty / nil slice.
func FirePackTacticsForAttackers(gs *GameState, attackers []*Permanent) {
	if gs == nil || len(attackers) == 0 {
		return
	}
	for _, p := range attackers {
		if p == nil {
			continue
		}
		defender, _ := AttackerDefender(p)
		FirePackTacticsTriggers(gs, p, defender)
	}
}
