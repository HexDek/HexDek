package per_card

import (
	"fmt"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMintharaMercilessSoul wires Minthara, Merciless Soul (Muninn
// parser-gap #85, ~6.5K hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	Minthara has ward {X}, where X is the number of experience counters
//	you have.
//	At the beginning of your end step, if a permanent you controlled
//	left the battlefield this turn, you get an experience counter.
//	Creatures you control get +1/+0 for each experience counter you have.
//
// Implementation:
//   - "permanent_ltb" listener: when any permanent leaves the battlefield
//     whose controller_seat == Minthara's controller, stamp a per-turn flag
//     on the controller's seat. (We listen on Minthara, but the engine fires
//     permanent_ltb for every LTB so the gate is the right scope.)
//   - "end_step" listener gated on active_seat == controller: if the flag is
//     set this turn, bump seat.Flags["experience_counters"].
//   - Scaling ward {X} via experience_counters and the static creature buff
//     are continuous effects — flagged partial. The engine has experience
//     counter scaling support (scaling.go) but the keyword grant ("Minthara
//     has ward {X}") and global anthem effect rely on the Phase 8 layers
//     pass, which is the same gap Phoenix Fleet Airship's static type-change
//     hits.
func registerMintharaMercilessSoul(r *Registry) {
	r.OnTrigger("Minthara, Merciless Soul", "permanent_ltb", mintharaPermLTB)
	r.OnTrigger("Minthara, Merciless Soul", "end_step", mintharaEndStep)
}

func mintharaPermLTB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	controller, _ := ctx["controller_seat"].(int)
	if controller != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if seat.Flags == nil {
		seat.Flags = map[string]int{}
	}
	seat.Flags[mintharaLTBKey(gs.Turn)] = 1
}

func mintharaEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "minthara_end_step_experience"
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
	key := mintharaLTBKey(gs.Turn)
	if seat.Flags == nil || seat.Flags[key] == 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "no_permanent_ltb_this_turn",
		})
		return
	}
	delete(seat.Flags, key)
	mintharaPruneKeys(seat, gs.Turn)
	seat.Flags["experience_counters"]++
	gs.LogEvent(gameengine.Event{
		Kind:   "experience_counter",
		Seat:   perm.Controller,
		Source: perm.Card.DisplayName(),
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"triggered":  true,
		"experience": seat.Flags["experience_counters"],
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"ward_X_and_creature_anthem_need_phase8_layers_overlay")
}

func mintharaLTBKey(turn int) string {
	return fmt.Sprintf("minthara_ltb_t%d", turn+1)
}

func mintharaPruneKeys(seat *gameengine.Seat, currentTurn int) {
	if seat == nil || seat.Flags == nil {
		return
	}
	prefix := "minthara_ltb_t"
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
