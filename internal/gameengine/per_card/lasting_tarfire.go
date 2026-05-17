package per_card

import (
	"fmt"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerLastingTarfire wires Lasting Tarfire.
//
// Oracle text (Scryfall, verified 2026-05-16):
//
//	At the beginning of each end step, if you put a counter on a
//	creature this turn, this enchantment deals 2 damage to each
//	opponent.
//
// Implementation:
//   - counter_placed listener: when our controller is the source seat
//     AND the target is a creature (any color of counter), set a
//     per-turn flag on the controller.
//   - end_step listener (fires for every player's end step): if the
//     flag is set this turn, deal 2 to each living opponent. Clear the
//     flag at the same time so the next turn re-evaluates from zero.
func registerLastingTarfire(r *Registry) {
	r.OnTrigger("Lasting Tarfire", "counter_placed", lastingTarfireCounter)
	r.OnTrigger("Lasting Tarfire", "end_step", lastingTarfireEndStep)
}

func lastingTarfireCounter(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	source, _ := ctx["source_seat"].(int)
	if source != perm.Controller {
		return
	}
	target, _ := ctx["target_perm"].(*gameengine.Permanent)
	if target == nil || !target.IsCreature() {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if seat.Flags == nil {
		seat.Flags = map[string]int{}
	}
	seat.Flags[lastingTarfireKey(gs.Turn)] = 1
}

func lastingTarfireEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "lasting_tarfire_end_step_drain"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	key := lastingTarfireKey(gs.Turn)
	if seat.Flags == nil || seat.Flags[key] == 0 {
		return
	}
	delete(seat.Flags, key)
	lastingTarfirePruneKeys(seat, gs.Turn)

	dealt := 0
	for i, s := range gs.Seats {
		if s == nil || s.Lost || i == perm.Controller {
			continue
		}
		gameengine.LoseLife(gs, i, 2, perm.Card.DisplayName())
		dealt++
	}
	_ = gs.CheckEnd()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"opps_hit": dealt,
		"damage":   2,
	})
}

func lastingTarfireKey(turn int) string {
	return fmt.Sprintf("tarfire_ctr_t%d", turn+1)
}

func lastingTarfirePruneKeys(seat *gameengine.Seat, currentTurn int) {
	if seat == nil || seat.Flags == nil {
		return
	}
	prefix := "tarfire_ctr_t"
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
