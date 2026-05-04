package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTheWarDoctor wires The War Doctor.
//
// Oracle text:
//
//	{2}{R}{W}
//	Legendary Creature — Time Lord Doctor
//	Whenever one or more other permanents phase out and whenever one or
//	  more other cards are put into exile from anywhere, put a time
//	  counter on The War Doctor.
//	Whenever The War Doctor attacks, it deals damage equal to the number
//	  of time counters on it to any target. If a creature dealt damage
//	  this way would die this turn, exile it instead.
//
// Implementation:
//   - "card_exiled" / "permanent_phased_out" triggers: gated to "another"
//     permanent/card; add a time counter to The War Doctor.
//   - "creature_attacks" trigger gated to The War Doctor: deal damage
//     equal to time counters to the leftmost living opponent (heuristic
//     target). emitPartial for the "if would die, exile instead" rider
//     (replacement effect not surfaced to per_card).
func registerTheWarDoctor(r *Registry) {
	r.OnTrigger("The War Doctor", "card_exiled", theWarDoctorOnExile)
	r.OnTrigger("The War Doctor", "permanent_phased_out", theWarDoctorOnPhaseOut)
	r.OnTrigger("The War Doctor", "creature_attacks", theWarDoctorOnAttack)
}

func theWarDoctorOnExile(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "the_war_doctor_time_counter_exile"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	// Skip self-exile.
	if c, ok := ctx["card"].(*gameengine.Card); ok && c == perm.Card {
		return
	}
	theWarDoctorAddTimeCounter(gs, perm, slug)
}

func theWarDoctorOnPhaseOut(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "the_war_doctor_time_counter_phaseout"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	if p, ok := ctx["perm"].(*gameengine.Permanent); ok && p == perm {
		return
	}
	theWarDoctorAddTimeCounter(gs, perm, slug)
}

func theWarDoctorAddTimeCounter(gs *gameengine.GameState, perm *gameengine.Permanent, slug string) {
	if perm.Counters == nil {
		perm.Counters = map[string]int{}
	}
	perm.Counters["time"]++
	gs.LogEvent(gameengine.Event{
		Kind:   "counter_mod",
		Seat:   perm.Controller,
		Source: perm.Card.DisplayName(),
		Amount: 1,
		Details: map[string]interface{}{
			"counter_kind": "time",
			"op":           "put",
			"reason":       slug,
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"total": perm.Counters["time"],
	})
}

func theWarDoctorOnAttack(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "the_war_doctor_attack_damage"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk == nil || atk != perm {
		return
	}
	dmg := perm.Counters["time"]
	if dmg <= 0 {
		return
	}
	target := -1
	for _, opp := range gs.Opponents(perm.Controller) {
		if gs.Seats[opp] != nil && !gs.Seats[opp].Lost {
			target = opp
			break
		}
	}
	if target < 0 {
		return
	}
	gs.Seats[target].Life -= dmg
	gs.LogEvent(gameengine.Event{
		Kind:   "damage",
		Seat:   perm.Controller,
		Target: target,
		Source: perm.Card.DisplayName(),
		Amount: dmg,
		Details: map[string]interface{}{
			"reason": "the_war_doctor_attack",
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"target": target,
		"damage": dmg,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"creature_dealt_damage_dies_this_turn_exile_replacement_partial")
	_ = gs.CheckEnd()
}
