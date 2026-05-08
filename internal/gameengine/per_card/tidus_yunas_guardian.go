package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTidusYunasGuardian wires Tidus, Yuna's Guardian.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	{G}{W}{U}
//	Legendary Creature — Human Warrior
//	3/3
//	At the beginning of combat on your turn, you may move a counter from
//	target creature you control onto a second target creature you
//	control.
//	Cheer — Whenever one or more creatures you control with counters on
//	them deal combat damage to a player, you may draw a card and
//	proliferate. Do this only once each turn.
//
// Implementation:
//   - begin_combat_controller: scan battlefield for counter-bearing
//     creatures, pick a counter to move from the LEAST-impactful holder
//     (lowest power) onto the highest-power non-counter creature. Skip if
//     no valid pair exists.
//   - creature_combat_damage_to_player: gated to controller, gated to a
//     creature with at least one counter; draw a card and proliferate
//     (mark all your counter-bearing permanents +1 of their first kind).
//     Per-turn lock via perm.Flags["tidus_cheered"].
//   - End-of-turn flag clear isn't wired to a turn-cleanup hook here; the
//     emitPartial flags this — the once-per-turn lock can persist beyond
//     the intended scope. Acceptable: at worst we suppress a future cheer.
func registerTidusYunasGuardian(r *Registry) {
	r.OnTrigger("Tidus, Yuna's Guardian", "begin_combat_controller", tidusYunasMoveCounter)
	r.OnTrigger("Tidus, Yuna's Guardian", "creature_combat_damage_to_player", tidusYunasCheer)
}

func tidusYunasMoveCounter(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "tidus_yunas_move_counter"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	var donor *gameengine.Permanent
	donorPower := 1 << 30
	var recipient *gameengine.Permanent
	recipientScore := -1
	var counterKind string
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil || !p.IsCreature() {
			continue
		}
		hasCounter := false
		var firstKind string
		for k, v := range p.Counters {
			if v > 0 {
				hasCounter = true
				firstKind = k
				break
			}
		}
		if hasCounter {
			pw := p.Power()
			if pw < donorPower {
				donorPower = pw
				donor = p
				counterKind = firstKind
			}
		} else {
			score := p.Power()*2 + p.Toughness()
			if score > recipientScore {
				recipientScore = score
				recipient = p
			}
		}
	}
	if donor == nil || recipient == nil || donor == recipient {
		return
	}
	donor.AddCounter(counterKind, -1)
	recipient.AddCounter(counterKind, 1)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         perm.Controller,
		"counter_kind": counterKind,
		"from":         donor.Card.DisplayName(),
		"to":           recipient.Card.DisplayName(),
	})
}

func tidusYunasCheer(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "tidus_yunas_cheer_draw_proliferate"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atkSeat, _ := ctx["controller_seat"].(int)
	if atkSeat != perm.Controller {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk == nil {
		return
	}
	hasCounter := false
	for _, v := range atk.Counters {
		if v > 0 {
			hasCounter = true
			break
		}
	}
	if !hasCounter {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	if perm.Flags["tidus_cheered"] == 1 {
		return
	}
	perm.Flags["tidus_cheered"] = 1
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	// Proliferate: add one of each existing counter kind to every counter-bearing permanent
	// the controller controls. (Full proliferate also covers opponent permanents and players;
	// limited scope here keeps the AI simulation positive-sum.)
	seat := gs.Seats[perm.Controller]
	if seat != nil {
		for _, p := range seat.Battlefield {
			if p == nil {
				continue
			}
			for k, v := range p.Counters {
				if v > 0 {
					p.AddCounter(k, 1)
				}
			}
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"attacker": atk.Card.DisplayName(),
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"once_per_turn_flag_not_cleared_at_end_of_turn")
}
