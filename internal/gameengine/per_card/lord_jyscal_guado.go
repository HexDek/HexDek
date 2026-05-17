package per_card

import (
	"fmt"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerLordJyscalGuado wires Lord Jyscal Guado (Muninn parser-gap #39, 24,473 hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{1}{W}
//	Legendary Creature — Spirit Cleric
//	Flying
//	At the beginning of each end step, if you put a counter on a
//	creature this turn, investigate.
//
// Pattern mirrors lasting_tarfire.go: track a per-turn "controller put a
// counter on a creature" flag, then create a Clue token at each end step
// when the flag is set. "Each end step" means the listener fires for
// every player's end step; we don't gate on active_seat.
func registerLordJyscalGuado(r *Registry) {
	r.OnTrigger("Lord Jyscal Guado", "counter_placed", lordJyscalGuadoCounter)
	r.OnTrigger("Lord Jyscal Guado", "end_step", lordJyscalGuadoEndStep)
}

func lordJyscalGuadoCounter(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
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
	s := gs.Seats[perm.Controller]
	if s == nil {
		return
	}
	if s.Flags == nil {
		s.Flags = map[string]int{}
	}
	s.Flags[lordJyscalGuadoKey(gs.Turn)] = 1
}

func lordJyscalGuadoEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "lord_jyscal_guado_investigate"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	s := gs.Seats[perm.Controller]
	if s == nil || s.Lost {
		return
	}
	key := lordJyscalGuadoKey(gs.Turn)
	if s.Flags == nil || s.Flags[key] == 0 {
		return
	}
	delete(s.Flags, key)
	lordJyscalGuadoPruneKeys(s, gs.Turn)

	gameengine.CreateClueToken(gs, perm.Controller)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}

func lordJyscalGuadoKey(turn int) string {
	return fmt.Sprintf("jyscal_ctr_t%d", turn+1)
}

func lordJyscalGuadoPruneKeys(seat *gameengine.Seat, currentTurn int) {
	if seat == nil || seat.Flags == nil {
		return
	}
	prefix := "jyscal_ctr_t"
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
