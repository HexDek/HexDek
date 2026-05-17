package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGauFeralYouth wires Gau, Feral Youth (Muninn parser-gap #79,
// ~7.8K hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{1}{R}
//	Legendary Creature — Human Berserker
//	Rage — Whenever Gau attacks, put a +1/+1 counter on it.
//	At the beginning of each end step, if a card left your graveyard
//	this turn, Gau deals damage equal to its power to each opponent.
//
// Implementation:
//   - creature_attacks gated on attacker_perm == self: add a +1/+1
//     counter to Gau (Rage).
//   - end_step: the "card left your graveyard this turn" tracker isn't
//     exposed via TurnCounters today (same engine gap noted on
//     Quintorius and Tormod). emitPartial. The damage half (deals power
//     to each opponent) is conditionally wired — we fire it whenever a
//     proxy heuristic (Sacrificed + ExiledCards + drew-from-yard hits)
//     suggests a graveyard exit may have occurred this turn. This is
//     deliberately over-eager rather than dead — most decks that run
//     Gau pair him with reanimator or flashback shells where the proxy
//     is almost always positive.
func registerGauFeralYouth(r *Registry) {
	r.OnTrigger("Gau, Feral Youth", "creature_attacks", gauFeralYouthOnAttack)
	r.OnTrigger("Gau, Feral Youth", "end_step", gauFeralYouthEndStep)
}

func gauFeralYouthOnAttack(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "gau_feral_youth_rage_counter"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk == nil || atk != perm {
		return
	}
	perm.AddCounter("+1/+1", 1)
	gs.InvalidateCharacteristicsCache()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"counters": perm.Counters["+1/+1"],
	})
}

func gauFeralYouthEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "gau_feral_youth_end_step_damage"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	// Printed "each end step", not "your end step" — no active_seat gate.
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	// Proxy: "a card left your graveyard this turn" ~ any reanimate
	// (PermanentsLeft hits), any exile (ExiledCards), or any flashback
	// (Casts list contains an exile-sourced cast). Cheapest check:
	// nonzero exile or nonzero sacrifices that may have routed through
	// graveyard. We accept false positives — emitPartial flags it.
	left := seat.Turn.ExiledCards + seat.Turn.CastFromExile
	if left == 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "no_graveyard_exit_proxy_hits",
		})
		return
	}
	dmg := perm.Power()
	if dmg <= 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "nonpositive_power",
		})
		return
	}
	hit := 0
	for _, opp := range gs.Opponents(perm.Controller) {
		s := gs.Seats[opp]
		if s == nil || s.Lost {
			continue
		}
		gameengine.DealDamage(gs, opp, dmg, perm.Card.DisplayName())
		hit++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"triggered": true,
		"damage":    dmg,
		"opponents": hit,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"card_left_graveyard_this_turn_tracker_uses_exile_proxy")
}
