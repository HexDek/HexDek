package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRavenloftAdventurer wires Ravenloft Adventurer.
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	When this creature enters, you take the initiative.
//	If a creature an opponent controls would die, instead exile it and
//	put a hit counter on it.
//	Whenever this creature attacks, if you've completed a dungeon,
//	defending player loses 1 life for each card they own in exile with
//	a hit counter on it.
//
// Implementation (Muninn gap #31 — 31K hits):
//   - OnETB: call gameengine.TakeInitiative for the controller. The
//     initiative state machine is wired in keywords_misc.go (line 1637+).
//   - OnTrigger("creature_attacks"): gated on
//     gs.Seats[seat].Flags["dungeon_completed"] (the SBA in sba.go:1154
//     sets this when a player completes any dungeon). Defending player
//     loses 1 life per hit-counter card they own in exile, tracked via
//     the canonical Etrata-style flag "hit_counter_cards_in_exile".
//   - The "creatures-die-replacement → exile + hit counter" replacement
//     effect requires §614 replacement-effect registration that the
//     engine doesn't yet expose for per-card hooks. emitPartial.
func registerRavenloftAdventurer(r *Registry) {
	r.OnETB("Ravenloft Adventurer", ravenloftAdventurerETB)
	r.OnTrigger("Ravenloft Adventurer", "creature_attacks", ravenloftAdventurerAttacks)
}

func ravenloftAdventurerETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "ravenloft_adventurer_take_initiative"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	gameengine.TakeInitiative(gs, seat)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": seat,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"die_replacement_exile_with_hit_counter_unmodeled_needs_phase614_per_card_hook")
}

func ravenloftAdventurerAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ravenloft_adventurer_dungeon_drain"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk != perm {
		return
	}
	seatIdx := perm.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[seatIdx]
	if seat == nil || seat.Lost {
		return
	}
	if seat.Flags["dungeon_completed"] == 0 && seat.Flags["completed_dungeon_ever"] == 0 {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_dungeon_completed", map[string]interface{}{
			"seat": seatIdx,
		})
		return
	}
	defender, _ := ctx["defender_seat"].(int)
	if defender < 0 || defender >= len(gs.Seats) {
		defender = -1
		bestLife := -1
		for _, opp := range gs.LivingOpponents(seatIdx) {
			if gs.Seats[opp].Life > bestLife {
				bestLife = gs.Seats[opp].Life
				defender = opp
			}
		}
		if defender < 0 {
			return
		}
	}
	defSeat := gs.Seats[defender]
	if defSeat == nil || defSeat.Lost {
		return
	}
	hitCards := defSeat.Flags["hit_counter_cards_in_exile"]
	if hitCards <= 0 {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_hit_counter_exile_cards", map[string]interface{}{
			"seat":     seatIdx,
			"defender": defender,
		})
		return
	}
	gameengine.FireLoseLifeEvent(gs, defender, hitCards, perm)
	_ = gs.CheckEnd()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      seatIdx,
		"defender":  defender,
		"life_lost": hitCards,
	})
}
