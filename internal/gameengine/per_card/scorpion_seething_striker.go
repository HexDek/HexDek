package per_card

import (
	"fmt"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerScorpionSeethingStriker wires Scorpion, Seething Striker (Muninn
// parser-gap #91, ~5.2K hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{1}{B}
//	Creature — Scorpion
//	Deathtouch
//	At the beginning of your end step, if a creature died this turn,
//	target creature you control connives.
//
// Implementation:
//   - Deathtouch: AST keyword pipeline.
//   - "creature_dies" listener stamps a per-turn flag on the controller's
//     seat. ("A creature died" — any creature, any controller, mirrors
//     Compy Swarm's #69 pattern.)
//   - "end_step" listener gated on active_seat == controller: pick the
//     highest-power creature we control (Scorpion itself is fine if it's
//     our biggest creature, mirroring Raffine's targeting heuristic) and
//     call gameengine.Connive(target, 1).
func registerScorpionSeethingStriker(r *Registry) {
	r.OnTrigger("Scorpion, Seething Striker", "creature_dies", scorpionSeethingDies)
	r.OnTrigger("Scorpion, Seething Striker", "end_step", scorpionSeethingEndStep)
}

func scorpionSeethingDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if seat.Flags == nil {
		seat.Flags = map[string]int{}
	}
	seat.Flags[scorpionSeethingDiedKey(gs.Turn)] = 1
}

func scorpionSeethingEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "scorpion_seething_end_step_connive"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	activeSeat, ok := ctx["active_seat"].(int)
	if !ok || activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	key := scorpionSeethingDiedKey(gs.Turn)
	if seat.Flags == nil || seat.Flags[key] == 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "no_creature_died_this_turn",
		})
		return
	}
	delete(seat.Flags, key)
	scorpionSeethingPruneKeys(seat, gs.Turn)
	var target *gameengine.Permanent
	bestPower := -1
	for _, p := range seat.Battlefield {
		if p == nil || !p.IsCreature() {
			continue
		}
		pw := p.Power()
		if pw > bestPower {
			bestPower = pw
			target = p
		}
	}
	if target == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "no_creature_to_target",
		})
		return
	}
	gameengine.Connive(gs, target, 1)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"target": target.Card.DisplayName(),
	})
}

func scorpionSeethingDiedKey(turn int) string {
	return fmt.Sprintf("scorpion_seething_died_t%d", turn+1)
}

func scorpionSeethingPruneKeys(seat *gameengine.Seat, currentTurn int) {
	if seat == nil || seat.Flags == nil {
		return
	}
	prefix := "scorpion_seething_died_t"
	cutoff := currentTurn + 1
	for k := range seat.Flags {
		if len(k) <= len(prefix) || k[:len(prefix)] != prefix {
			continue
		}
		n := 0
		_, err := fmt.Sscanf(k[len(prefix):], "%d", &n)
		if err != nil {
			continue
		}
		if n < cutoff {
			delete(seat.Flags, k)
		}
	}
}
