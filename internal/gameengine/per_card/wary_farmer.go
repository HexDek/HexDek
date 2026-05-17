package per_card

import (
	"fmt"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerWaryFarmer wires Wary Farmer (Muninn parser-gap #67, ~9.7K hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{1}{G/W}{G/W}
//	Creature — Kithkin Citizen
//	At the beginning of your end step, if another creature entered the
//	battlefield under your control this turn, surveil 1.
//
// Implementation:
//   - "permanent_etb" listener stamps a per-turn flag on the controller's
//     seat when ANOTHER creature (not Wary Farmer itself) enters under
//     Wary Farmer's controller.
//   - "end_step" listener gated on active_seat == controller: if the flag
//     is set this turn, call gameengine.Surveil(gs, controller, 1).
func registerWaryFarmer(r *Registry) {
	r.OnTrigger("Wary Farmer", "permanent_etb", waryFarmerPermETB)
	r.OnTrigger("Wary Farmer", "end_step", waryFarmerEndStep)
}

func waryFarmerPermETB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	entering, _ := ctx["permanent"].(*gameengine.Permanent)
	if entering == nil {
		entering, _ = ctx["perm"].(*gameengine.Permanent)
	}
	if entering == nil || entering == perm {
		return
	}
	if !entering.IsCreature() {
		return
	}
	if entering.Controller != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if seat.Flags == nil {
		seat.Flags = map[string]int{}
	}
	seat.Flags[waryFarmerEtbKey(gs.Turn)] = 1
}

func waryFarmerEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "wary_farmer_end_step_surveil"
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
	key := waryFarmerEtbKey(gs.Turn)
	if seat.Flags == nil || seat.Flags[key] == 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "no_creature_etb_this_turn",
		})
		return
	}
	delete(seat.Flags, key)
	waryFarmerPruneKeys(seat, gs.Turn)
	gameengine.Surveil(gs, perm.Controller, 1)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"triggered": true,
		"surveil":   1,
	})
}

func waryFarmerEtbKey(turn int) string {
	return fmt.Sprintf("wary_farmer_etb_t%d", turn+1)
}

func waryFarmerPruneKeys(seat *gameengine.Seat, currentTurn int) {
	if seat == nil || seat.Flags == nil {
		return
	}
	prefix := "wary_farmer_etb_t"
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
